// Package doctor diagnostica o harness do pipeline (binário, skills, contratos, worktrees e
// processos). Só lê: nada é corrigido, apagado ou encerrado (RN-7).
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/pipeline"
)

// Status de uma checagem.
const (
	StatusOK    = "ok"
	StatusAviso = "aviso"
	StatusFalha = "falha"
)

// ExitFalha é o código de saída do comando quando alguma checagem falha.
const ExitFalha = 3

const (
	maxDetalhe  = 160
	maxItens    = 5
	comandoInst = "npm install -g @nathanramorim/forge-sdd"
)

// Checagem é o resultado de uma verificação do harness.
type Checagem struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Detalhe  string `json:"detalhe"`
	Sugestao string `json:"sugestao"`
}

// Relatorio é a saída estável de `forge doctor --harness --json`.
type Relatorio struct {
	Versao    int        `json:"versao"`
	Checagens []Checagem `json:"checagens"`
	Falhas    int        `json:"falhas"`
	Avisos    int        `json:"avisos"`
}

// Deps isola o que toca o sistema (processos e execução de binários), para teste.
type Deps struct {
	Runner  monitor.CommandRunner
	Usuario string
	Versao  string
}

// DefaultDeps usa o sistema real e a versão do binário em execução.
func DefaultDeps() Deps {
	return Deps{Runner: monitor.DefaultCommandRunner, Usuario: usuarioAtual(), Versao: config.BinaryVersion()}
}

// Diagnose roda todas as checagens sobre o projeto em dir, sem alterá-lo.
func Diagnose(dir string, d Deps) Relatorio {
	r := Relatorio{Versao: 1, Checagens: []Checagem{
		checaVersao(dir, d.Versao),
		checaConfig(dir),
		checaSkills(dir),
		checaContrato(dir),
		checaPareceres(dir),
		checaWorktrees(dir),
		checaWorkers(d),
	}}
	for _, c := range r.Checagens {
		switch c.Status {
		case StatusFalha:
			r.Falhas++
		case StatusAviso:
			r.Avisos++
		}
	}
	return r
}

// ValidateRelatorio é o contrato da saída: ids únicos, status válido, sugestão em aviso/falha e
// contadores coerentes.
func ValidateRelatorio(r Relatorio) error {
	vistos := map[string]bool{}
	var falhas, avisos int
	for _, c := range r.Checagens {
		if strings.TrimSpace(c.ID) == "" || vistos[c.ID] {
			return fmt.Errorf("relatório do doctor: id vazio ou repetido %q", c.ID)
		}
		vistos[c.ID] = true
		switch c.Status {
		case StatusOK:
		case StatusAviso:
			avisos++
		case StatusFalha:
			falhas++
		default:
			return fmt.Errorf("relatório do doctor: checagem %q com status inválido %q", c.ID, c.Status)
		}
		if c.Status != StatusOK && strings.TrimSpace(c.Sugestao) == "" {
			return fmt.Errorf("relatório do doctor: checagem %q (%s) sem sugestão", c.ID, c.Status)
		}
	}
	if falhas != r.Falhas || avisos != r.Avisos {
		return fmt.Errorf("relatório do doctor: contadores incoerentes (falhas %d/%d, avisos %d/%d)", r.Falhas, falhas, r.Avisos, avisos)
	}
	return nil
}

// ExitCode devolve 3 quando alguma checagem falhou e 0 caso contrário.
func (r Relatorio) ExitCode() int {
	if r.Falhas > 0 {
		return ExitFalha
	}
	return 0
}

func cortar(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxDetalhe {
		return s[:maxDetalhe-3] + "..."
	}
	return s
}

func listar(itens []string) string {
	if len(itens) <= maxItens {
		return strings.Join(itens, ", ")
	}
	return fmt.Sprintf("%s e mais %d", strings.Join(itens[:maxItens], ", "), len(itens)-maxItens)
}

func ok(id, detalhe string) Checagem {
	return Checagem{ID: id, Status: StatusOK, Detalhe: cortar(detalhe)}
}

func aviso(id, detalhe, sugestao string) Checagem {
	return Checagem{ID: id, Status: StatusAviso, Detalhe: cortar(detalhe), Sugestao: cortar(sugestao)}
}

func falha(id, detalhe, sugestao string) Checagem {
	return Checagem{ID: id, Status: StatusFalha, Detalhe: cortar(detalhe), Sugestao: cortar(sugestao)}
}

func checaVersao(dir, binario string) Checagem {
	const id = "versao"
	data, err := os.ReadFile(layout.Sddrc(dir))
	if err != nil {
		return falha(id, fmt.Sprintf("não foi possível ler sdd/.sddrc: %v", err), "rode forge init neste repositório")
	}
	var rc struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &rc); err != nil || strings.TrimSpace(rc.Version) == "" {
		return falha(id, "sdd/.sddrc sem versão legível", "corrija a chave version em sdd/.sddrc")
	}
	proj := rc.Version
	if v, err := config.ReadProjectVersion(dir); err == nil && v != proj {
		return aviso(id, fmt.Sprintf("sdd/.sddrc indica %s e sdd/.sdd-version indica %s", proj, v), "rode forge update")
	}
	if binario == "dev" {
		return ok(id, "binário de desenvolvimento; projeto em "+proj)
	}
	switch {
	case versaoMenor(binario, proj):
		return falha(id, fmt.Sprintf("binário %s mais antigo que o projeto %s", binario, proj), "atualize o binário: "+comandoInst)
	case versaoMenor(proj, binario):
		return aviso(id, fmt.Sprintf("projeto %s mais antigo que o binário %s", proj, binario), "rode forge update")
	}
	return ok(id, fmt.Sprintf("binário e projeto em %s", proj))
}

func checaConfig(dir string) Checagem {
	const id = "forge-json"
	_, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return falha(id, err.Error(), "corrija .forge/config/forge.json (inclui modelos e esforços por papel)")
	}
	if !found {
		return ok(id, "sem .forge/config/forge.json: modo sessions, padrões em uso")
	}
	return ok(id, "forge.json válido, com modelos e esforços")
}

func checaSkills(dir string) Checagem {
	const id = "skills-rules"
	estacoes := []string{"sdd-concepcao", "sdd-planejamento", "sdd-build", "sdd-governanca", "sdd-operacao"}
	var faltam []string
	for _, e := range estacoes {
		if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", e, "SKILL.md")); err != nil {
			faltam = append(faltam, e)
		}
	}
	if ents, err := os.ReadDir(filepath.Join(dir, ".agents", "rules")); err != nil || len(ents) == 0 {
		faltam = append(faltam, ".agents/rules")
	}
	if len(faltam) > 0 {
		return aviso(id, "ausentes: "+listar(faltam), "rode forge update para ressincronizar skills e rules")
	}
	return ok(id, "5 skills de estação e rules presentes")
}

func checaContrato(dir string) Checagem {
	const id = "contrato-skill-binario"
	if err := harness.VerificarContratoParecer(dir); err != nil {
		return falha(id, err.Error(), "alinhe sdd-governanca/references/veredito.md ao contrato Parecer do binário (forge update)")
	}
	return ok(id, "skill de Governança e contrato Parecer coincidem")
}

func checaPareceres(dir string) Checagem {
	const id = "pareceres-gravados"
	arqs, _ := filepath.Glob(filepath.Join(layout.Discovery(dir), "*", "verdicts", "*.json"))
	var ruins []string
	for _, a := range arqs {
		data, err := os.ReadFile(a)
		if err == nil {
			_, err = harness.DecodeParecer(data)
		}
		if err != nil {
			ruins = append(ruins, filepath.Base(a))
		}
	}
	if len(ruins) > 0 {
		return aviso(id, fmt.Sprintf("%d de %d pareceres que o leitor atual não decodifica: %s", len(ruins), len(arqs), listar(ruins)), "são registros antigos: não regrave e arquive a iniciativa antiga se não forem mais úteis")
	}
	return ok(id, fmt.Sprintf("%d pareceres gravados decodificam", len(arqs)))
}

func checaWorktrees(dir string) Checagem {
	const id = "worktrees"
	var orfaos []string
	stale, err := pipeline.ListStaleWorktrees(dir)
	if err != nil {
		return aviso(id, fmt.Sprintf("não foi possível listar worktrees do pipeline: %v", err), "confira .forge/runtime/worktrees")
	}
	for _, s := range stale {
		if s.Safe {
			orfaos = append(orfaos, filepath.ToSlash(filepath.Join(".forge/runtime/worktrees", s.JobID, s.TaskID)))
		}
	}
	if ents, err := os.ReadDir(filepath.Join(dir, ".claude", "worktrees")); err == nil {
		for _, e := range ents {
			if e.IsDir() {
				orfaos = append(orfaos, ".claude/worktrees/"+e.Name())
			}
		}
	}
	if len(orfaos) > 0 {
		return aviso(id, fmt.Sprintf("%d worktree(s) para revisar: %s", len(orfaos), listar(orfaos)), "o doctor não remove nada: confira com git worktree list e remova à mão")
	}
	return ok(id, "nenhum worktree órfão")
}

var caminhoWorker = regexp.MustCompile(`^/(?:[^/\s]+/)*forge-sdd/(\d+\.\d+\.\d+(?:-[0-9A-Za-z.]+)?)/forge-sdd$`)

// versaoDoCaminho extrai a versão do caminho do cache do wrapper npm
// (<cache>/forge-sdd/<versão>/forge-sdd). Nada é executado (SEG-01, RN-7).
func versaoDoCaminho(exe string) (string, bool) {
	m := caminhoWorker.FindStringSubmatch(exe)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// limpa troca caracteres de controle e limita o tamanho do que veio do ps antes de imprimir.
func limpa(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '?'
		}
		return r
	}, s)
	if len(s) > 60 {
		s = s[:57] + "..."
	}
	return s
}

func checaWorkers(d Deps) Checagem {
	const id = "workers"
	if d.Runner == nil {
		return ok(id, "listagem de processos indisponível")
	}
	out, err := d.Runner("ps", "-eo", "user=,pid=,command=")
	if err != nil {
		return aviso(id, fmt.Sprintf("não foi possível listar processos: %v", err), "rode forge ps")
	}
	var velhos []string
	for _, linha := range strings.Split(string(out), "\n") {
		campos := strings.Fields(linha)
		if len(campos) < 4 || campos[0] != d.Usuario || campos[3] != "worker" || !filepath.IsAbs(campos[2]) || !strings.HasPrefix(filepath.Base(campos[2]), "forge") {
			continue
		}
		v, achou := versaoDoCaminho(campos[2])
		switch {
		case !achou:
			velhos = append(velhos, fmt.Sprintf("pid %s (versão desconhecida: %s)", limpa(campos[1]), limpa(campos[2])))
		case versaoMenor(v, d.Versao) || versaoMenor(d.Versao, v):
			velhos = append(velhos, fmt.Sprintf("pid %s (%s)", limpa(campos[1]), limpa(v)))
		}
	}
	if len(velhos) > 0 {
		return aviso(id, fmt.Sprintf("worker(s) de versão diferente de %s: %s", d.Versao, listar(velhos)), "encerre-os à mão e suba de novo com forge up; o doctor não encerra processos nem executa binários")
	}
	return ok(id, "nenhum worker de outra versão")
}

func usuarioAtual() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

// versaoMenor diz se a < b, tratando um sufixo de pré-lançamento (-beta) como anterior à versão
// final; config.CompareVersions ignora o sufixo e não serve aqui.
func versaoMenor(a, b string) bool {
	if c := config.CompareVersions(a, b); c != 0 {
		return c < 0
	}
	preA := strings.Contains(a, "-")
	preB := strings.Contains(b, "-")
	if preA != preB {
		return preA
	}
	return preA && a < b
}
