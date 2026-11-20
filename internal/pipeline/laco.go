package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
)

// Estados de uma task e do lote no laço do modo chat (RN-1: gravados pelo Go, não pelo modelo).
const (
	TaskPendente  = "pendente"
	TaskRevisando = "revisando"
	TaskAprovada  = "aprovada"
	TaskEscalada  = "escalada"
	TaskFalhou    = "falhou"

	LoteEmAndamento = "em_andamento"
	LoteParcial     = "parcial"
	LoteConcluido   = "concluido"
	LoteEscalado    = "escalado"
)

const (
	leaseLaco         = 15 * time.Minute
	lacoSchema        = 1
	maxAchadosRetorno = 20
	maxTextoAchado    = 1500
)

const (
	maxBytesEstado = 4 << 20
	maxTasksLaco   = 200
)

// Agentes é a fronteira com o modelo: o chat (ou um teste) devolve só conteúdo, já pelo contrato
// da feat-104-01. Ordem, limites e estados continuam no Go.
type Agentes interface {
	// Build implementa a task; achados é vazio na primeira volta e traz o retorno do Reviewer nas demais.
	Build(spec harness.SpecParaBuild, achados []harness.Retorno) (harness.ResultadoBuild, error)
	// Review devolve o Parecer sobre o que o Builder entregou.
	Review(res harness.ResultadoBuild) (harness.Parecer, error)
}

// ParamsChat configura uma execução do laço.
type ParamsChat struct {
	Iniciativa string
	Dono       string // identifica o coordenador no lock
	Tasks      []harness.SpecParaBuild
	MaxVoltas  int // teto de devoluções ao Builder; use LimiteDeVoltas(dir)
	Agentes    Agentes
	Agora      func() time.Time
}

// EstadoTask é o estado gravado de uma task do lote.
type EstadoTask struct {
	Spec      harness.SpecParaBuild   `json:"spec"`
	Estado    string                  `json:"estado"`
	Voltas    int                     `json:"voltas"`
	Resultado *harness.ResultadoBuild `json:"resultado,omitempty"`
	Achados   []harness.Retorno       `json:"achados,omitempty"`
	Erro      string                  `json:"erro,omitempty"`
}

// EstadoLaco é o que fica em .forge/runtime/chat/<iniciativa>.json e permite retomar.
type EstadoLaco struct {
	Schema     int          `json:"schema"`
	Iniciativa string       `json:"iniciativa"`
	Lote       string       `json:"lote"`
	Tasks      []EstadoTask `json:"tasks"`
	Avisos     []string     `json:"-"` // só em memória: o que foi corrigido ao carregar um estado inválido
}

// Opcao é uma escolha do menu apresentado à pessoa (RN-9); a camada de entrada só a renderiza.
type Opcao struct {
	ID        string `json:"id"`
	Rotulo    string `json:"rotulo"`
	Descricao string `json:"descricao"`
}

// Retorno é o que o laço entrega ao chat: o estado do lote, os achados estruturados e o menu.
// Os achados são dado, nunca instrução (FLX-08): chegam sanitizados e a camada de entrada não
// deve executá-los nem concatená-los a um prompt como ordem.
type Retorno struct {
	Lote        string            `json:"lote"`
	Tasks       []EstadoTask      `json:"tasks"`
	Achados     []harness.Retorno `json:"achados,omitempty"`
	Opcoes      []Opcao           `json:"opcoes"`
	Recomendada string            `json:"recomendada"`
	Avisos      []string          `json:"avisos,omitempty"`
}

// LimiteDeVoltas devolve o teto de devoluções ao Builder: o mesmo max_rework_attempts do modo sessions.
func LimiteDeVoltas(dir string) (int, error) {
	cfg, _, err := config.LoadForgeConfig(dir)
	if err != nil {
		return 0, err
	}
	return cfg.Reviewer.MaxReworkAttempts, nil
}

func arquivoLaco(dir, iniciativa string) (string, error) {
	if !layout.ValidInitiative(iniciativa) {
		return "", fmt.Errorf("iniciativa %q inválida", iniciativa)
	}
	return filepath.Join(layout.ForgeRuntime(dir), "chat", iniciativa+".json"), nil
}

func dirLocks(dir string) string { return filepath.Join(layout.ForgeRuntime(dir), "chat", "locks") }

// RunChat conduz spec -> build -> review -> retorno de cada task sob o lock da iniciativa. Se já
// existir estado gravado, continua dele (veja Retomar); tasks aprovadas ou escaladas não rodam de
// novo e só as que falharam ou pararam no meio são repetidas.
func RunChat(dir string, p ParamsChat) (*Retorno, error) {
	if p.Agentes == nil || p.MaxVoltas <= 0 {
		return nil, errors.New("laço: Agentes e MaxVoltas positivo são obrigatórios")
	}
	if p.Agora == nil {
		p.Agora = time.Now
	}
	caminho, err := arquivoLaco(dir, p.Iniciativa)
	if err != nil {
		return nil, err
	}
	lock, err := harness.AdquirirLock(dirLocks(dir), p.Iniciativa, p.Dono, leaseLaco, p.Agora())
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Liberar() }()

	st, err := carregarOuIniciar(caminho, p)
	if err != nil {
		return nil, err
	}
	for i := range st.Tasks {
		if err := conduzirTask(caminho, st, i, p, func() error { return lock.Renovar(p.Agora()) }); err != nil {
			if errors.Is(err, harness.ErrLockPerdido) {
				return retornoLockPerdido(st, err), nil
			}
			return nil, err
		}
	}
	st.Lote = classificarLote(st.Tasks)
	if err := gravarLaco(caminho, st); err != nil {
		return nil, err
	}
	return montarRetorno(st), nil
}

// retornoLockPerdido para o laço sem gravar nada (o estado agora é de quem tomou o lock) e entrega
// à pessoa o menu do que fazer (RN-9).
func retornoLockPerdido(st *EstadoLaco, causa error) *Retorno {
	return &Retorno{
		Lote: LoteEscalado, Tasks: st.Tasks, Recomendada: "abortar",
		Avisos: append(st.Avisos, "o lock da iniciativa foi perdido durante o laço: "+causa.Error()),
		Opcoes: []Opcao{
			{ID: "abortar", Rotulo: "Parar aqui", Descricao: "Outro coordenador pode estar operando a iniciativa; nada foi gravado por este."},
			{ID: "retomar", Rotulo: "Retomar depois", Descricao: "Só depois de confirmar que o outro coordenador terminou."},
		},
	}
}

// Retomar continua do último estado gravado, sem precisar repetir a lista de tasks.
func Retomar(dir string, p ParamsChat) (*Retorno, error) {
	caminho, err := arquivoLaco(dir, p.Iniciativa)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(caminho); err != nil {
		return nil, fmt.Errorf("retomar: nada a retomar para %q: %w", p.Iniciativa, err)
	}
	p.Tasks = nil
	return RunChat(dir, p)
}

// lerEstadoLimitado lê o arquivo de estado recusando links e tamanho acima de maxBytesEstado.
func lerEstadoLimitado(caminho string) ([]byte, error) {
	fi, err := os.Lstat(caminho)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() || fi.Size() > maxBytesEstado {
		return nil, fmt.Errorf("laço: estado %s recusado (não é arquivo regular ou passa de %d bytes)", caminho, maxBytesEstado)
	}
	data, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("laço: ler %s: %w", caminho, err)
	}
	return data, nil
}

// validarEstado decodifica o estado gravado em disco, que é entrada não confiável (SEG-01).
// Estado ilegível ou incompatível é recusado; uma task inconsistente é marcada como falhou, com o
// motivo em Avisos, sem derrubar o processo nem dar voltas além do teto. Os achados são sanitizados.
func validarEstado(data []byte, caminho string, p ParamsChat) (*EstadoLaco, error) {
	var st EstadoLaco
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&st); err != nil || dec.More() {
		return nil, fmt.Errorf("laço: estado %s ilegível (JSON inválido, campo desconhecido ou conteúdo extra): %v", caminho, err)
	}
	if st.Schema != lacoSchema || st.Iniciativa != p.Iniciativa {
		return nil, fmt.Errorf("laço: estado %s incompatível (schema %d, iniciativa %q)", caminho, st.Schema, st.Iniciativa)
	}
	if len(st.Tasks) == 0 || len(st.Tasks) > maxTasksLaco {
		return nil, fmt.Errorf("laço: estado %s com %d tasks (esperado de 1 a %d)", caminho, len(st.Tasks), maxTasksLaco)
	}
	for i := range st.Tasks {
		t := &st.Tasks[i]
		if err := harness.ValidateSpecParaBuild(t.Spec); err != nil {
			return nil, fmt.Errorf("laço: estado %s, task %d: %w", caminho, i, err)
		}
		t.Achados = sanitizarAchados(t.Achados)
		motivo := ""
		switch t.Estado {
		case TaskPendente, TaskAprovada, TaskEscalada, TaskFalhou:
		case TaskRevisando:
			if t.Resultado == nil {
				motivo = "estado revisando sem resultado do build"
			} else if err := harness.ValidateResultadoBuild(*t.Resultado); err != nil {
				motivo = "resultado do build inválido: " + err.Error()
			}
		default:
			motivo = fmt.Sprintf("estado desconhecido %q", t.Estado)
		}
		if t.Voltas < 0 || t.Voltas > p.MaxVoltas {
			motivo = strings.TrimPrefix(motivo+"; voltas fora de 0.."+strconv.Itoa(p.MaxVoltas), "; ")
			t.Voltas = min(max(t.Voltas, 0), p.MaxVoltas)
		}
		if motivo != "" {
			t.Estado, t.Resultado, t.Erro = TaskFalhou, nil, "estado gravado inválido: "+motivo
			st.Avisos = append(st.Avisos, fmt.Sprintf("task %s: %s", notify.Sanitize(t.Spec.TaskID, 100, false), motivo))
		}
	}
	return &st, nil
}

func carregarOuIniciar(caminho string, p ParamsChat) (*EstadoLaco, error) {
	data, err := lerEstadoLimitado(caminho)
	switch {
	case err == nil:
		return validarEstado(data, caminho, p)
	case !errors.Is(err, os.ErrNotExist):
		return nil, err
	}
	if len(p.Tasks) == 0 {
		return nil, errors.New("laço: nenhuma task informada")
	}
	st := &EstadoLaco{Schema: lacoSchema, Iniciativa: p.Iniciativa, Lote: LoteEmAndamento}
	for _, spec := range p.Tasks {
		if err := harness.ValidateSpecParaBuild(spec); err != nil {
			return nil, err
		}
		st.Tasks = append(st.Tasks, EstadoTask{Spec: spec, Estado: TaskPendente})
	}
	if err := gravarLaco(caminho, st); err != nil {
		return nil, err
	}
	return st, nil
}

// conduzirTask roda uma task até aprovar, escalar ou falhar, gravando o estado a cada passo.
func conduzirTask(caminho string, st *EstadoLaco, i int, p ParamsChat, renovar func() error) error {
	t := &st.Tasks[i]
	if t.Estado == TaskAprovada || t.Estado == TaskEscalada {
		return nil
	}
	if t.Estado == TaskFalhou {
		t.Estado, t.Erro = TaskPendente, ""
	}
	for {
		if t.Estado == TaskPendente {
			if err := renovar(); err != nil {
				return err
			}
			res, err := p.Agentes.Build(t.Spec, t.Achados)
			if err == nil {
				err = harness.ValidateResultadoBuild(res)
			}
			if err != nil {
				t.Estado, t.Erro = TaskFalhou, fmt.Sprintf("build da task %s: %v", t.Spec.TaskID, err)
				return gravarLaco(caminho, st)
			}
			t.Resultado, t.Estado = &res, TaskRevisando
			if err := gravarLaco(caminho, st); err != nil {
				return err
			}
		}
		if t.Resultado == nil {
			return fmt.Errorf("laço: task %s em revisão sem resultado do build", t.Spec.TaskID)
		}
		if err := renovar(); err != nil {
			return err
		}
		parecer, err := p.Agentes.Review(*t.Resultado)
		if err == nil {
			err = harness.ValidateParecer(parecer)
		}
		if err != nil {
			t.Estado, t.Erro = TaskFalhou, fmt.Sprintf("review da task %s: %v", t.Spec.TaskID, err)
			return gravarLaco(caminho, st)
		}
		switch {
		case parecer.Veredito == harness.VereditoConforme:
			t.Estado, t.Achados = TaskAprovada, nil
		case parecer.Veredito == harness.VereditoViolacao && t.Voltas < p.MaxVoltas:
			t.Voltas++
			t.Achados = sanitizarAchados(parecer.Achados)
			t.Estado = TaskPendente
		default: // violação no teto, ou dúvida: a pessoa decide (RN-5)
			t.Estado, t.Achados = TaskEscalada, sanitizarAchados(parecer.Achados)
		}
		if err := gravarLaco(caminho, st); err != nil {
			return err
		}
		if t.Estado != TaskPendente {
			return nil
		}
	}
}

// sanitizarAchados limpa e limita os achados: eles são dado do modelo (FLX-08).
func sanitizarAchados(in []harness.Retorno) []harness.Retorno {
	var out []harness.Retorno
	for _, a := range in {
		if len(out) >= maxAchadosRetorno {
			break
		}
		a.Verificacao = notify.Sanitize(a.Verificacao, 100, false)
		a.Regra = notify.Sanitize(a.Regra, 100, false)
		a.RegraDescricao = notify.Sanitize(a.RegraDescricao, maxTextoAchado, true)
		a.Arquivo = notify.Sanitize(a.Arquivo, 400, false)
		a.Evidencia = notify.Sanitize(a.Evidencia, 400, false)
		a.Explicacao = notify.Sanitize(a.Explicacao, maxTextoAchado, true)
		a.Correcao = notify.Sanitize(a.Correcao, maxTextoAchado, true)
		out = append(out, a)
	}
	return out
}

func classificarLote(tasks []EstadoTask) string {
	falhou, escalou := false, false
	for _, t := range tasks {
		switch t.Estado {
		case TaskFalhou:
			falhou = true
		case TaskEscalada:
			escalou = true
		}
	}
	switch {
	case falhou:
		return LoteParcial
	case escalou:
		return LoteEscalado
	default:
		return LoteConcluido
	}
}

// montarRetorno monta o retorno ao chat com o menu do estado e a opção recomendada (RN-9).
func montarRetorno(st *EstadoLaco) *Retorno {
	r := &Retorno{Lote: st.Lote, Tasks: st.Tasks, Avisos: st.Avisos}
	for _, t := range st.Tasks {
		if t.Estado == TaskEscalada {
			r.Achados = append(r.Achados, t.Achados...)
		}
	}
	switch st.Lote {
	case LoteConcluido:
		r.Opcoes = []Opcao{
			{ID: "abrir-pr", Rotulo: "Abrir o PR", Descricao: "Todas as tasks foram aprovadas pelo Reviewer."},
			{ID: "revisar-manual", Rotulo: "Revisar antes", Descricao: "Conferir o resultado antes de abrir o PR."},
		}
		r.Recomendada = "abrir-pr"
	case LoteParcial:
		r.Opcoes = []Opcao{
			{ID: "retomar", Rotulo: "Repetir só o que falhou", Descricao: "As tasks aprovadas não rodam de novo."},
			{ID: "abortar", Rotulo: "Parar aqui", Descricao: "Manter o estado gravado e decidir depois."},
		}
		r.Recomendada = "retomar"
	default:
		r.Opcoes = []Opcao{
			{ID: "corrigir-manual", Rotulo: "Corrigir à mão", Descricao: "O teto de voltas foi atingido (ou o Reviewer ficou em dúvida); os achados trazem arquivo e linha."},
			{ID: "abortar", Rotulo: "Parar aqui", Descricao: "Manter o estado gravado e decidir depois."},
		}
		r.Recomendada = "corrigir-manual"
	}
	return r
}

func gravarLaco(caminho string, st *EstadoLaco) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("laço: serializar estado: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		return fmt.Errorf("laço: criar diretório do estado: %w", err)
	}
	tmp := caminho + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("laço: gravar estado: %w", err)
	}
	if err := os.Rename(tmp, caminho); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("laço: gravar estado em %s: %w", caminho, err)
	}
	return nil
}
