package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

// ErrLockOcupado indica que outro coordenador já detém o lock da iniciativa.
var ErrLockOcupado = errors.New("iniciativa já tem um coordenador ativo")

// ErrLockPerdido indica que o lock não pertence mais a quem tenta renová-lo ou liberá-lo.
// guardaMax é o prazo para descartar o guarda de recuperação deixado por um processo morto.
const guardaMax = time.Minute

var ErrLockPerdido = errors.New("o lock da iniciativa não pertence mais a este coordenador")

// InfoLock é o conteúdo do arquivo de lock: quem o detém e até quando vale a lease.
type InfoLock struct {
	Dono   string    `json:"dono"`
	PID    int       `json:"pid"`
	Inicio time.Time `json:"inicio"`
	Expira time.Time `json:"expira"`
}

// Lock é a posse do lock de uma iniciativa. Seguro para vários processos: a criação do arquivo é
// atômica (Link sem sobrescrever) e a tomada de um lock expirado passa por os.Rename, que só um processo vence.
//
// Se o processo morrer segurando o lock, o arquivo fica órfão até a lease expirar; depois disso o
// próximo coordenador o recupera sozinho. Nada é apagado antes do prazo: um coordenador lento,
// mas vivo, que renova a lease a tempo (Renovar) nunca perde a posse.
type Lock struct {
	path  string
	dono  string
	lease time.Duration
}

// Caminho devolve o arquivo do lock.
func (l *Lock) Caminho() string { return l.path }

// AdquirirLock tenta obter o lock de iniciativa em dirLocks. Devolve ErrLockOcupado se outro dono
// tem uma lease ainda válida em now. Um lock com lease vencida é tomado.
func AdquirirLock(dirLocks, iniciativa, dono string, lease time.Duration, now time.Time) (*Lock, error) {
	if iniciativa == "" || dono == "" || lease <= 0 {
		return nil, fmt.Errorf("lock: iniciativa, dono e lease positiva são obrigatórios")
	}
	if err := os.MkdirAll(dirLocks, 0o755); err != nil {
		return nil, fmt.Errorf("lock: criar %s: %w", dirLocks, err)
	}
	l := &Lock{path: filepath.Join(dirLocks, iniciativa+".lock"), dono: dono, lease: lease}
	for tentativa := 0; tentativa < 8; tentativa++ {
		ok, err := l.criar(now)
		if err != nil {
			return nil, err
		}
		if ok {
			return l, nil
		}
		atual, err := lerLock(l.path)
		if err != nil {
			// O arquivo sumiu ou está ilegível entre a criação e a leitura: tenta de novo, e um
			// arquivo corrompido é tratado como órfão depois de ser removido.
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if _, rerr := l.tomar(now); rerr != nil {
				return nil, rerr
			}
			continue
		}
		if now.Before(atual.Expira) {
			return nil, fmt.Errorf("%w %q (dono %s, pid %d, lease até %s)", ErrLockOcupado, iniciativa, atual.Dono, atual.PID, atual.Expira.Format(time.RFC3339))
		}
		if _, err := l.tomar(now); err != nil {
			return nil, err
		}
		time.Sleep(time.Duration(tentativa+1) * 5 * time.Millisecond)
	}
	return nil, fmt.Errorf("%w %q: disputa pelo lock órfão", ErrLockOcupado, iniciativa)
}

// criar publica o lock; devolve false se ele já existe.
func (l *Lock) criar(now time.Time) (bool, error) {
	data, err := json.Marshal(InfoLock{Dono: l.dono, PID: os.Getpid(), Inicio: now, Expira: now.Add(l.lease)})
	if err != nil {
		return false, fmt.Errorf("lock: serializar: %w", err)
	}
	// O conteúdo é gravado num temporário e publicado com Link, que é atômico e não sobrescreve:
	// quem lê o lock nunca vê um arquivo pela metade e só um criador vence.
	tmp := l.path + ".novo." + sufixoUnico()
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return false, fmt.Errorf("lock: gravar %s: %w", tmp, err)
	}
	defer func() { _ = os.Remove(tmp) }()
	if err := os.Link(tmp, l.path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, fmt.Errorf("lock: publicar %s: %w", l.path, err)
	}
	return true, nil
}

// tomar remove o lock vencido (ou ilegível). Só quem detém o guarda `.tomada` remove, e só depois
// de reler o lock e confirmar que ainda está vencido: assim ninguém apaga o lock válido que outro
// coordenador acabou de criar. O guarda é um Link atômico; se o processo morrer segurando-o, ele é
// descartado depois de guardaMax pelo mtime. Quem não obtém o guarda apenas devolve false.
func (l *Lock) tomar(now time.Time) (bool, error) {
	guarda := l.path + ".tomada"
	tmp := guarda + "." + sufixoUnico()
	if err := os.WriteFile(tmp, []byte(l.dono), 0o644); err != nil {
		return false, fmt.Errorf("lock: gravar guarda: %w", err)
	}
	defer func() { _ = os.Remove(tmp) }()
	if err := os.Link(tmp, guarda); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return false, fmt.Errorf("lock: obter guarda de recuperação: %w", err)
		}
		if fi, serr := os.Stat(guarda); serr == nil && time.Since(fi.ModTime()) > guardaMax {
			_ = os.Remove(guarda)
		}
		return false, nil
	}
	defer func() { _ = os.Remove(guarda) }()
	if atual, err := lerLock(l.path); err == nil && now.Before(atual.Expira) {
		return false, nil
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("lock: remover lock órfão %s: %w", l.path, err)
	}
	return true, nil
}

func lerLock(path string) (InfoLock, error) {
	var info InfoLock
	data, err := os.ReadFile(path)
	if err != nil {
		return info, err
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return info, fmt.Errorf("lock: %s ilegível: %w", path, err)
	}
	return info, nil
}

// Renovar estende a lease a partir de now. Falha com ErrLockPerdido se o arquivo mudou de dono.
func (l *Lock) Renovar(now time.Time) error {
	atual, err := lerLock(l.path)
	if err != nil || atual.Dono != l.dono || atual.PID != os.Getpid() || !now.Before(atual.Expira) {
		return fmt.Errorf("%w (%s)", ErrLockPerdido, l.path)
	}
	atual.Expira = now.Add(l.lease)
	data, err := json.Marshal(atual)
	if err != nil {
		return fmt.Errorf("lock: serializar: %w", err)
	}
	tmp := l.path + ".tmp." + strconv.Itoa(os.Getpid())
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("lock: gravar renovação: %w", err)
	}
	if err := os.Rename(tmp, l.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("lock: renovar %s: %w", l.path, err)
	}
	return nil
}

// Liberar remove o lock se ainda for deste dono; liberar duas vezes não é erro.
func (l *Lock) Liberar() error {
	atual, err := lerLock(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || atual.Dono != l.dono || atual.PID != os.Getpid() {
		return fmt.Errorf("%w (%s)", ErrLockPerdido, l.path)
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("lock: liberar %s: %w", l.path, err)
	}
	return nil
}

var seqTmp atomic.Uint64

// sufixoUnico distingue arquivos temporários de goroutines e processos (o relógio sozinho repete).
func sufixoUnico() string {
	return strconv.Itoa(os.Getpid()) + "." + strconv.FormatUint(seqTmp.Add(1), 36) + "." + strconv.FormatInt(time.Now().UnixNano(), 36)
}
