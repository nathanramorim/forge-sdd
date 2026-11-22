package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// Erros do split feito pelo Builder (FLX-01, feat-106-23).
var (
	// ErrSplitSemGate1: o Builder só fatia depois do Gate 1 aprovado.
	ErrSplitSemGate1 = errors.New("o Builder só faz o split depois do Gate 1 aprovado")
	// ErrSplitForaDoPlano: a task pedida não está no plan.md aprovado.
	ErrSplitForaDoPlano = errors.New("a task não está no plan.md aprovado")
	// ErrSplitVazio: o plano aprovado não tem tasks para enfileirar.
	ErrSplitVazio = errors.New("o plan.md aprovado não tem tasks")
)

// PlanoAprovado é a lista de tasks que o `plan.md` aprovado autoriza o Builder a enfileirar.
// É a fronteira do split automático: nada fora desta lista entra na fila.
type PlanoAprovado struct {
	Iniciativa string
	Tasks      []string // IDs, na ordem do plano
}

// IdsDasTasksDoPlano extrai os IDs de task do texto do plan.md. Reconhece `feat-<id>-<nn>` e
// `fix-<id>-<nn>` em qualquer lugar da linha (tabela ou lista). Devolve os IDs únicos, na ordem
// em que aparecem.
func IdsDasTasksDoPlano(texto string) []string {
	vistos := map[string]bool{}
	var ids []string
	for _, linha := range strings.Split(texto, "\n") {
		for _, campo := range strings.FieldsFunc(linha, func(r rune) bool {
			return r == '|' || r == '`' || r == '*' || r == ' ' || r == '\t' || r == '(' || r == ')' || r == ':' || r == ','
		}) {
			id := normalizarIdTask(campo)
			if id == "" || vistos[id] {
				continue
			}
			vistos[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// normalizarIdTask aceita `feat-106-23`, `fix-106-01` e variações com sufixo de letra
// (`feat-106-24a`). Devolve vazio para o que não é ID de task.
func normalizarIdTask(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Trim(s, ".:,;()[]<>\"'")
	if !strings.HasPrefix(s, "feat-") && !strings.HasPrefix(s, "fix-") {
		return ""
	}
	partes := strings.Split(s, "-")
	if len(partes) < 3 {
		return ""
	}
	// numero da iniciativa e numero da task precisam ser numéricos (o último pode ter letra).
	if !soDigitos(partes[1]) {
		return ""
	}
	ultimo := partes[len(partes)-1]
	digitos := strings.TrimRight(ultimo, "abcdefghijklmnopqrstuvwxyz")
	if !soDigitos(digitos) || digitos == "" {
		return ""
	}
	return s
}

func soDigitos(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// CarregaPlanoAprovado lê o plan.md da iniciativa e devolve os IDs que ele autoriza. Só vale
// com o Gate 1 aprovado: sem ele, o Builder não fatia nada (FLX-01).
func CarregaPlanoAprovado(dir, iniciativa string) (PlanoAprovado, error) {
	base, err := layout.Initiative(dir, iniciativa)
	if err != nil {
		return PlanoAprovado{}, err
	}
	if !gate1Marked(dir, iniciativa) {
		return PlanoAprovado{}, ErrSplitSemGate1
	}
	caminho := filepath.Join(base, "plan.md")
	data, err := os.ReadFile(caminho)
	if err != nil {
		if os.IsNotExist(err) {
			return PlanoAprovado{}, fmt.Errorf("plan.md ausente em %s", base)
		}
		return PlanoAprovado{}, fmt.Errorf("ler plan.md: %w", err)
	}
	ids := IdsDasTasksDoPlano(string(data))
	if len(ids) == 0 {
		return PlanoAprovado{}, ErrSplitVazio
	}
	return PlanoAprovado{Iniciativa: iniciativa, Tasks: ids}, nil
}

// Contem diz se o ID está no plano aprovado.
func (p PlanoAprovado) Contem(id string) bool {
	alvo := normalizarIdTask(id)
	if alvo == "" {
		// ID que não segue o padrão (usado em teste): compara como veio, sem normalizar.
		alvo = strings.TrimSpace(id)
		if alvo == "" {
			return false
		}
	}
	for _, t := range p.Tasks {
		if t == alvo {
			return true
		}
	}
	return false
}

// OrdemDeEnfileiramento devolve as tasks na ordem do plano, respeitando as dependências que a
// fila já registra. Task pedida fora do plano é recusada antes de qualquer escrita.
func (p PlanoAprovado) OrdemDeEnfileiramento(pedidas []string) ([]string, error) {
	if len(pedidas) == 0 {
		return append([]string(nil), p.Tasks...), nil
	}
	for _, t := range pedidas {
		if !p.Contem(t) {
			return nil, fmt.Errorf("%w: %q (o plano aprovado tem %s)",
				ErrSplitForaDoPlano, t, strings.Join(p.Tasks, ", "))
		}
	}
	// Mantém a ordem do plano entre as pedidas.
	queridas := map[string]bool{}
	for _, t := range pedidas {
		id := normalizarIdTask(t)
		if id == "" {
			id = strings.TrimSpace(t)
		}
		queridas[id] = true
	}
	var out []string
	for _, t := range p.Tasks {
		if queridas[t] {
			out = append(out, t)
		}
	}
	return out, nil
}

// SplitPeloBuilder enfileira as tasks do plano aprovado sem nova aprovação humana — é o que a
// FLX-01 autoriza a partir do Gate 1. Cada task precisa de spec em features/ ou fixes/.
//
// Devolve as tasks enfileiradas e as que já estavam na fila.
func SplitPeloBuilder(dir, iniciativa string, pedidas []string) (enfileiradas, jaExistiam []string, err error) {
	plano, err := CarregaPlanoAprovado(dir, iniciativa)
	if err != nil {
		return nil, nil, err
	}
	ordem, err := plano.OrdemDeEnfileiramento(pedidas)
	if err != nil {
		return nil, nil, err
	}

	q, err := queue.LoadForInitiative(dir, iniciativa)
	if err != nil {
		return nil, nil, fmt.Errorf("ler a fila: %w", err)
	}
	naFila := map[string]bool{}
	for _, it := range q.Items {
		naFila[normalizarIdTask(it.TaskID)] = true
	}

	base, err := layout.Initiative(dir, iniciativa)
	if err != nil {
		return nil, nil, err
	}

	for _, id := range ordem {
		if naFila[id] {
			jaExistiam = append(jaExistiam, id)
			continue
		}
		spec, titulo, err := acharSpecDaTask(dir, base, id)
		if err != nil {
			return enfileiradas, jaExistiam, err
		}
		if err := queue.Enqueue(dir, q.FeatureID, q.Branch, id, titulo, spec, nil); err != nil {
			// O Enqueue grava na fila da iniciativa ativa; se a de destino é outra, ele
			// enxerga uma fila vazia e tenta criar a task de novo. A duplicata é o sinal
			// de que a task já está na fila certa.
			if strings.Contains(err.Error(), "já existe na fila") {
				jaExistiam = append(jaExistiam, id)
				naFila[id] = true
				continue
			}
			return enfileiradas, jaExistiam, fmt.Errorf("enfileirar %s: %w", id, err)
		}
		enfileiradas = append(enfileiradas, id)
		naFila[id] = true
	}
	sort.Strings(jaExistiam)
	return enfileiradas, jaExistiam, nil
}

// acharSpecDaTask procura o arquivo da spec em features/ e fixes/ e lê o título dele.
// Devolve o caminho relativo à raiz do projeto (é o que a fila grava e o gate relê).
func acharSpecDaTask(dir, base, id string) (caminho, titulo string, err error) {
	for _, sub := range []string{"features", "fixes"} {
		dirSub := filepath.Join(base, sub)
		entradas, err := os.ReadDir(dirSub)
		if err != nil {
			continue
		}
		for _, e := range entradas {
			nome := e.Name()
			if e.IsDir() || !strings.HasSuffix(nome, ".md") {
				continue
			}
			semExt := strings.TrimSuffix(nome, ".md")
			if normalizarIdTask(semExt) == id || strings.HasPrefix(semExt, id+"-") {
				abs := filepath.Join(dirSub, nome)
				rel, relErr := filepath.Rel(dir, abs)
				if relErr != nil {
					rel = filepath.Join(sub, nome)
				}
				return filepath.ToSlash(rel), tituloDaSpec(abs, id), nil
			}
		}
	}
	return "", "", fmt.Errorf("%w: %s não tem spec em features/ nem fixes/", ErrSplitForaDoPlano, id)
}

// tituloDaSpec lê o `# <id> — <título>` da primeira linha; sem isso, devolve o ID.
func tituloDaSpec(caminho, id string) string {
	data, err := os.ReadFile(caminho)
	if err != nil {
		return id
	}
	primeira := strings.SplitN(string(data), "\n", 2)[0]
	primeira = strings.TrimSpace(strings.TrimPrefix(primeira, "#"))
	if primeira == "" {
		return id
	}
	if _, resto, ok := strings.Cut(primeira, "—"); ok {
		if t := strings.TrimSpace(resto); t != "" {
			return t
		}
	}
	return primeira
}
