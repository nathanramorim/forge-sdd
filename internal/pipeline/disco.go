package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
)

// Teto de paralelo por máquina (ADR-0029, RN-5 a RN-7 e RN-10): quantos pares Builder e Reviewer
// cabem agora, pelo disco, pela largura da onda e pela janela de uso.

const (
	// LimiteJanelaPadrao é a ocupação da janela de uso acima da qual o teto cai para 1 par (RN-10).
	LimiteJanelaPadrao = 85
	// RemedirACada é o número de tasks entre duas medições do tamanho do worktree (RN-6).
	RemedirACada = 20
	// tamanhoMinimoWorktree evita divisão por zero e vagas infinitas com medição ínfima.
	tamanhoMinimoWorktree = 1 << 20
)

// UsoDisco é a capacidade e o uso de um volume, em bytes.
type UsoDisco struct {
	Total uint64
	Usado uint64
}

// Pct devolve a ocupação do volume, de 0 a 100.
func (u UsoDisco) Pct() float64 {
	if u.Total == 0 {
		return 100
	}
	return float64(u.Usado) * 100 / float64(u.Total)
}

// MedirDisco mede o volume que contém path; se path ainda não existe, mede o primeiro pai que existe.
func MedirDisco(path string) (UsoDisco, error) {
	p := path
	for {
		if _, err := os.Stat(p); err == nil {
			break
		}
		pai := filepath.Dir(p)
		if pai == p {
			break
		}
		p = pai
	}
	return medirVolume(p)
}

// MedirDiscoDosWorktrees mede o volume onde os worktrees vivem (layout.Worktrees).
func MedirDiscoDosWorktrees(dir string) (UsoDisco, error) {
	return MedirDisco(layout.Worktrees(dir))
}

// VagasDeDisco calcula quantos worktrees novos cabem sem passar do limite de ocupação:
// (limite% da capacidade − uso atual) ÷ tamanho do worktree, arredondado para baixo e nunca negativo.
func VagasDeDisco(u UsoDisco, limitePct int, tamanhoWorktree uint64) int {
	if tamanhoWorktree < tamanhoMinimoWorktree {
		tamanhoWorktree = tamanhoMinimoWorktree
	}
	// O limite é rígido: nenhum worktree novo leva o volume acima de limitePct.
	teto := u.Total / 100 * uint64(limitePct)
	if u.Usado >= teto {
		return 0
	}
	return int((teto - u.Usado) / tamanhoWorktree)
}

// ParamsTeto reúne o que o teto considera. JanelaPct negativo significa janela desconhecida;
// MemoriaDisponivelM negativa significa memória desconhecida (a trava de memória não vale).
type ParamsTeto struct {
	Disco              UsoDisco
	LimiteDiscoPct     int    // builder.limite_disco
	TamanhoWorktree    uint64 // bytes
	LarguraOnda        int    // tasks elegíveis agora (PlanWaves + PodeComecar)
	MaxParalelo        int    // builder.max_parallel_tasks; 0 = sem teto fixo, >0 = teto opcional
	JanelaPct          int    // ocupação da janela de uso; negativo = desconhecida
	LimiteJanelaPct    int    // padrão LimiteJanelaPadrao
	MemoriaDisponivelM int    // MB disponíveis agora; negativo = desconhecida
	MemoriaTotalMB     int    // RAM total em MB; 0 ou negativo = desconhecida (só a reserva vale)
	LimiteMemoriaPct   int    // builder.limite_memoria (padrão 80)
	ReservaMemoriaMB   int    // builder.reserva_memoria_mb, piso opcional
	MemoriaPorParMB    int    // estimativa por par; 0 usa MemoriaPorParPadraoMB
}

// MemoriaPorParPadraoMB é a estimativa por par (Builder + Reviewer) enquanto não há medida.
const MemoriaPorParPadraoMB = 1536

const semLimite = 1 << 30

// Teto é o resultado do cálculo, com cada parcela para o status e para o diagnóstico. Parcela sem
// limite vale semLimite.
type Teto struct {
	Vagas        int // pelo disco
	VagasMemoria int // pela memória
	Largura      int // pela onda
	Janela       int // pela janela de uso (1 acima do limite)
	Fixo         int // teto opcional max_parallel_tasks
	Pares        int // o mínimo de todos
	Motivo       string
}

// ReservaDeMemoria é o que fica livre: o maior entre (100 − limite)% da RAM total e o piso em MB.
// Com o total desconhecido vale só o piso.
func ReservaDeMemoria(totalMB, limitePct, pisoMB int) int {
	if limitePct <= 0 {
		limitePct = 80
	}
	r := pisoMB
	if totalMB > 0 {
		if pct := totalMB * (100 - limitePct) / 100; pct > r {
			r = pct
		}
	}
	return r
}

// VagasDeMemoria calcula (disponível − reserva) ÷ memória por par, nunca negativa.
func VagasDeMemoria(disponivelMB, reservaMB, porParMB int) int {
	if porParMB <= 0 {
		porParMB = MemoriaPorParPadraoMB
	}
	livre := disponivelMB - reservaMB
	if livre <= 0 {
		return 0
	}
	return livre / porParMB
}

// TetoDeParalelo calcula pares = min(vagas de disco, vagas de memória, largura da onda, teto da
// janela, teto opcional se > 0), com piso de 1 quando há task elegível e ao menos uma vaga de disco
// e de memória (RN-5, RN-10).
func TetoDeParalelo(p ParamsTeto) Teto {
	t := Teto{
		Vagas:        VagasDeDisco(p.Disco, p.LimiteDiscoPct, p.TamanhoWorktree),
		VagasMemoria: semLimite,
		Largura:      p.LarguraOnda,
		Janela:       semLimite,
		Fixo:         semLimite,
	}
	if p.MemoriaDisponivelM >= 0 {
		t.VagasMemoria = VagasDeMemoria(p.MemoriaDisponivelM, ReservaDeMemoria(p.MemoriaTotalMB, p.LimiteMemoriaPct, p.ReservaMemoriaMB), p.MemoriaPorParMB)
	}
	lim := p.LimiteJanelaPct
	if lim <= 0 {
		lim = LimiteJanelaPadrao
	}
	if p.JanelaPct >= 0 && p.JanelaPct > lim {
		t.Janela = 1
	}
	if p.MaxParalelo > 0 {
		t.Fixo = p.MaxParalelo
	}
	t.Pares, t.Motivo = t.Vagas, "disco"
	for _, c := range []struct {
		v int
		n string
	}{{t.VagasMemoria, "memória"}, {t.Largura, "onda"}, {t.Janela, "janela de uso"}, {t.Fixo, "builder.max_parallel_tasks"}} {
		if c.v < t.Pares {
			t.Pares, t.Motivo = c.v, c.n
		}
	}
	if t.Pares < 1 && t.Largura >= 1 && t.Vagas >= 1 && t.VagasMemoria >= 1 {
		t.Pares = 1
	}
	if t.Pares < 0 {
		t.Pares = 0
	}
	return t
}

// EstadoDisco é o que a máquina guarda em .forge/runtime/disco.json: o tamanho medido de um worktree
// e quantas tasks passaram desde a medição (RN-6). É por máquina e fica fora do git.
type EstadoDisco struct {
	TamanhoWorktree uint64    `json:"tamanho_worktree_bytes"`
	MedidoEm        time.Time `json:"medido_em"`
	TasksDesde      int       `json:"tasks_desde_medicao"`
	// MemoriaParMB é a memória medida de um par Builder e Reviewer (0 = sem medida, usa a estimativa).
	MemoriaParMB int `json:"memoria_par_mb,omitempty"`
}

func caminhoEstadoDisco(dir string) string {
	return filepath.Join(layout.ForgeRuntime(dir), "disco.json")
}

// LerEstadoDisco lê o estado; arquivo ausente ou corrompido devolve ok=false (remedir).
func LerEstadoDisco(dir string) (EstadoDisco, bool) {
	var e EstadoDisco
	data, err := os.ReadFile(caminhoEstadoDisco(dir))
	if err != nil || len(data) > 4096 || json.Unmarshal(data, &e) != nil || e.TamanhoWorktree == 0 {
		return EstadoDisco{}, false
	}
	return e, true
}

// GravarEstadoDisco grava o estado de forma atômica (arquivo temporário e rename).
func GravarEstadoDisco(dir string, e EstadoDisco) error {
	p := caminhoEstadoDisco(dir)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// MedirTamanhoDir soma o tamanho dos arquivos de raiz, sem seguir links e sem entrar em .git nem em
// .forge/runtime (o worktree novo não traz o histórico nem o estado local).
func MedirTamanhoDir(raiz string) (uint64, error) {
	var total uint64
	runtime := layout.ForgeRuntime(raiz)
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				return nil
			}
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || p == runtime) {
			return fs.SkipDir
		}
		if d.Type().IsRegular() {
			if info, ierr := d.Info(); ierr == nil && info.Size() > 0 {
				total += uint64(info.Size())
			}
		}
		return nil
	})
	return total, err
}

// TamanhoDoWorktree devolve o tamanho guardado, medindo o projeto no primeiro uso e de novo a cada
// RemedirACada tasks (RN-6). Com gravar=false (dry-run) a medição não é persistida.
func TamanhoDoWorktree(dir string, gravar bool) (uint64, error) {
	e, ok := LerEstadoDisco(dir)
	if ok && e.TasksDesde < RemedirACada {
		return e.TamanhoWorktree, nil
	}
	tam, err := MedirTamanhoDir(dir)
	if err != nil {
		if ok {
			return e.TamanhoWorktree, nil // mantém a medição antiga se a nova falhar
		}
		return 0, fmt.Errorf("medir o tamanho do worktree: %w", err)
	}
	if gravar {
		_ = GravarEstadoDisco(dir, EstadoDisco{TamanhoWorktree: tam, MedidoEm: time.Now().UTC(), MemoriaParMB: e.MemoriaParMB})
	}
	return tam, nil
}

// RegistrarTaskConcluida soma uma task ao contador que dispara a nova medição.
func RegistrarTaskConcluida(dir string) error {
	e, ok := LerEstadoDisco(dir)
	if !ok {
		return nil // sem medição guardada, a próxima já mede
	}
	e.TasksDesde++
	return GravarEstadoDisco(dir, e)
}

// MemoriaDoPar devolve a memória por par guardada na máquina, ou a estimativa padrão.
func MemoriaDoPar(dir string) int {
	if e, ok := LerEstadoDisco(dir); ok && e.MemoriaParMB > 0 {
		return e.MemoriaParMB
	}
	return MemoriaPorParPadraoMB
}

// RegistrarMemoriaDoPar guarda a memória medida nos processos do primeiro par (por máquina).
func RegistrarMemoriaDoPar(dir string, mb int) error {
	if mb <= 0 {
		return fmt.Errorf("memória por par inválida: %d", mb)
	}
	e, ok := LerEstadoDisco(dir)
	if !ok {
		tam, err := MedirTamanhoDir(dir)
		if err != nil {
			return err
		}
		e = EstadoDisco{TamanhoWorktree: tam, MedidoEm: time.Now().UTC()}
	}
	e.MemoriaParMB = mb
	return GravarEstadoDisco(dir, e)
}
