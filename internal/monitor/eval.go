package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
)

// Modos do conjunto de avaliação (ADR-0008: sessions convive com chat, ADR-0015).
const (
	EvalModeSessions = "sessions"
	EvalModeChat     = "chat"
)

// EvalMinRefs e EvalMaxRefs delimitam o tamanho do conjunto de referência (feat-104-11).
const (
	EvalMinRefs = 10
	EvalMaxRefs = 20
)

// evalSimFile guarda, por feature de referência, um único desfecho simulado, aplicado igual aos
// dois modos: o simulador testa a apuração, não compara desenhos.
const evalSimFile = "simulacao.json"

// Origens possíveis dos números de eval-<modo>.json.
const (
	EvalOrigemSimulado = "simulado"
	EvalOrigemReal     = "real"
)

// EvalAvisoSimulado precede toda saída com números do simulador.
const EvalAvisoSimulado = "SIMULADO: números de exemplo, não medidos"

// EvalRefsDir devolve sdd/eval/refs.
func EvalRefsDir(dir string) string { return filepath.Join(layout.Root(dir), "eval", "refs") }

// EvalResultPath devolve sdd/.metrics/eval-<modo>.json.
func EvalResultPath(dir, mode string) string {
	return filepath.Join(layout.Metrics(dir), "eval-"+mode+".json")
}

// EvalSim é o desfecho simulado de uma feature de referência, idêntico nos dois modos.
type EvalSim struct {
	Aprovado      bool `json:"aprovado"`
	VoltasRetorno int  `json:"voltas_retorno"`
	Tokens        int  `json:"tokens"`
	Segundos      int  `json:"segundos"`
	Intervencoes  int  `json:"intervencoes"`
}

// EvalRef é uma feature de referência carregada e validada.
type EvalRef struct {
	ID  string
	Sim EvalSim
}

// EvalResult é o conteúdo de sdd/.metrics/eval-<modo>.json.
type EvalResult struct {
	Mode          string  `json:"mode"`
	Origem        string  `json:"origem"`
	Executor      string  `json:"executor"`
	GeneratedAt   string  `json:"generated_at"`
	Features      int     `json:"features"`
	Aprovadas     int     `json:"aprovadas"`
	TaxaAprovacao float64 `json:"taxa_aprovacao"`
	VoltasRetorno int     `json:"voltas_retorno"`
	Tokens        int     `json:"tokens"`
	Segundos      int     `json:"segundos"`
	Intervencoes  int     `json:"intervencoes"`
}

// ValidEvalMode diz se o modo é aceito por forge eval.
func ValidEvalMode(mode string) error {
	if mode != EvalModeSessions && mode != EvalModeChat {
		return fmt.Errorf("modo de avaliação inválido %q: use %q ou %q", mode, EvalModeSessions, EvalModeChat)
	}
	return nil
}

// LoadEvalRefs carrega sdd/eval/refs e reprova o conjunto inteiro se alguma referência tiver
// critério sem teste negativo válido (checagens estáticas do P-25, feat-104-02) ou simulação
// incompleta, ou se o total sair da faixa de EvalMinRefs a EvalMaxRefs.
func LoadEvalRefs(refsDir string) ([]EvalRef, error) {
	entries, err := os.ReadDir(refsDir)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler o conjunto de avaliação %s: %w", refsDir, err)
	}
	var refs []EvalRef
	var problems []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ref, errs := loadEvalRef(filepath.Join(refsDir, e.Name()))
		for _, p := range errs {
			problems = append(problems, e.Name()+": "+p)
		}
		if len(errs) == 0 {
			refs = append(refs, ref)
		}
	}
	total := len(refs) + countRefsWithProblems(problems)
	if total < EvalMinRefs || total > EvalMaxRefs {
		problems = append(problems, fmt.Sprintf("o conjunto tem %d features de referência; o esperado é de %d a %d", total, EvalMinRefs, EvalMaxRefs))
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("conjunto de avaliação reprovado em %s:\n- %s", refsDir, strings.Join(problems, "\n- "))
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })
	return refs, nil
}

func countRefsWithProblems(problems []string) int {
	seen := map[string]bool{}
	for _, p := range problems {
		seen[strings.SplitN(p, ":", 2)[0]] = true
	}
	return len(seen)
}

func loadEvalRef(refDir string) (EvalRef, []string) {
	ref := EvalRef{ID: filepath.Base(refDir)}
	var problems []string
	spec, err := os.ReadFile(filepath.Join(refDir, "spec.md"))
	if err != nil {
		problems = append(problems, fmt.Sprintf("falha ao ler spec.md: %v", err))
	} else if block, ok := gate.ExtractCriteriaBlock(string(spec)); !ok {
		problems = append(problems, fmt.Sprintf("spec.md sem a seção %q", gate.CriteriaHeading))
	} else {
		problems = append(problems, gate.CriteriaProblems(block)...)
	}

	data, err := os.ReadFile(filepath.Join(refDir, evalSimFile))
	if err != nil {
		return ref, append(problems, fmt.Sprintf("falha ao ler %s: %v", evalSimFile, err))
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ref.Sim); err != nil {
		return ref, append(problems, fmt.Sprintf("%s inválido (um único desfecho, igual nos dois modos): %v", evalSimFile, err))
	}
	return ref, problems
}

// SimulateEval agrega o desfecho simulado do conjunto num modo (sem modelo, determinístico). O
// resultado é o mesmo nos dois modos, exceto o campo Mode.
func SimulateEval(refs []EvalRef, mode string, at time.Time) (EvalResult, error) {
	if err := ValidEvalMode(mode); err != nil {
		return EvalResult{}, err
	}
	res := EvalResult{Mode: mode, Origem: EvalOrigemSimulado, Executor: "simulador", GeneratedAt: at.UTC().Format(time.RFC3339), Features: len(refs)}
	for _, r := range refs {
		s := r.Sim
		if s.Aprovado {
			res.Aprovadas++
		}
		res.VoltasRetorno += s.VoltasRetorno
		res.Tokens += s.Tokens
		res.Segundos += s.Segundos
		res.Intervencoes += s.Intervencoes
	}
	if res.Features > 0 {
		res.TaxaAprovacao = float64(res.Aprovadas) / float64(res.Features)
	}
	return res, nil
}

// RunEval valida o conjunto, simula o modo e grava sdd/.metrics/eval-<modo>.json.
func RunEval(dir, mode string, at time.Time) (EvalResult, string, error) {
	if err := ValidEvalMode(mode); err != nil {
		return EvalResult{}, "", err
	}
	refs, err := LoadEvalRefs(EvalRefsDir(dir))
	if err != nil {
		return EvalResult{}, "", err
	}
	res, err := SimulateEval(refs, mode, at)
	if err != nil {
		return EvalResult{}, "", err
	}
	path := EvalResultPath(dir, mode)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return EvalResult{}, "", fmt.Errorf("falha ao criar %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return EvalResult{}, "", fmt.Errorf("falha ao serializar o resultado do modo %s: %w", mode, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return EvalResult{}, "", fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
	return res, path, nil
}

// ReadEvalResult lê sdd/.metrics/eval-<modo>.json.
func ReadEvalResult(dir, mode string) (EvalResult, error) {
	path := EvalResultPath(dir, mode)
	data, err := os.ReadFile(path)
	if err != nil {
		return EvalResult{}, fmt.Errorf("falha ao ler %s (rode forge eval run --mode %s): %w", path, mode, err)
	}
	var res EvalResult
	if err := json.Unmarshal(data, &res); err != nil {
		return EvalResult{}, fmt.Errorf("%s inválido: %w", path, err)
	}
	if res.Origem != EvalOrigemSimulado && res.Origem != EvalOrigemReal {
		return EvalResult{}, fmt.Errorf("%s sem origem válida (%q): use %q ou %q", path, res.Origem, EvalOrigemSimulado, EvalOrigemReal)
	}
	if res.Mode != mode {
		return EvalResult{}, fmt.Errorf("%s declara o modo %q, esperado %q", path, res.Mode, mode)
	}
	return res, nil
}

// FormatEvalRun devolve a linha de resumo do forge eval run, avisando quando os números são simulados.
func FormatEvalRun(res EvalResult, path string) string {
	prefix := ""
	if res.Origem != EvalOrigemReal {
		prefix = EvalAvisoSimulado + ". "
	}
	return fmt.Sprintf("%smodo %s (origem %s): %d/%d aprovadas, %d voltas do Retorno, %d tokens, %ds, %d intervenções -> %s\n",
		prefix, res.Mode, res.Origem, res.Aprovadas, res.Features, res.VoltasRetorno, res.Tokens, res.Segundos, res.Intervencoes, path)
}

// CompareEval lê os dois modos e imprime a tabela comparativa. Se algum lado for simulado, recusa
// a comparação, a menos que permitirSimulado seja verdadeiro; mesmo assim imprime o aviso no topo.
func CompareEval(dir string, w io.Writer, permitirSimulado bool) error {
	s, err := ReadEvalResult(dir, EvalModeSessions)
	if err != nil {
		return err
	}
	c, err := ReadEvalResult(dir, EvalModeChat)
	if err != nil {
		return err
	}
	if s.Origem != EvalOrigemReal || c.Origem != EvalOrigemReal {
		if !permitirSimulado {
			return fmt.Errorf("comparação recusada: origem sessions=%s, chat=%s; números simulados não comparam desenhos (use --permitir-simulado só para testar a apuração)", s.Origem, c.Origem)
		}
		if _, err := fmt.Fprintf(w, "%s: esta tabela não compara os desenhos (origem sessions=%s, chat=%s)\n", EvalAvisoSimulado, s.Origem, c.Origem); err != nil {
			return fmt.Errorf("falha ao imprimir a comparação: %w", err)
		}
	}
	rows := [][3]string{
		{"origem", s.Origem, c.Origem},
		{"features", fmt.Sprint(s.Features), fmt.Sprint(c.Features)},
		{"taxa de aprovação", fmt.Sprintf("%.0f%%", s.TaxaAprovacao*100), fmt.Sprintf("%.0f%%", c.TaxaAprovacao*100)},
		{"voltas do Retorno", fmt.Sprint(s.VoltasRetorno), fmt.Sprint(c.VoltasRetorno)},
		{"tokens", fmt.Sprint(s.Tokens), fmt.Sprint(c.Tokens)},
		{"tempo (s)", fmt.Sprint(s.Segundos), fmt.Sprint(c.Segundos)},
		{"intervenções humanas", fmt.Sprint(s.Intervencoes), fmt.Sprint(c.Intervencoes)},
		{"executor", s.Executor, c.Executor},
	}
	if _, err := fmt.Fprintf(w, "%-22s %12s %12s\n", "métrica", EvalModeSessions, EvalModeChat); err != nil {
		return fmt.Errorf("falha ao imprimir a comparação: %w", err)
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-22s %12s %12s\n", r[0], r[1], r[2]); err != nil {
			return fmt.Errorf("falha ao imprimir a comparação: %w", err)
		}
	}
	return nil
}
