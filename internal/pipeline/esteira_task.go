package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/layout"
)

// Esteira por task entre processos (ADR-0029, feat-122-05/06/09). Sem estado em memória: cada
// worker lê e grava o que precisa em .forge/runtime/tarefas/<iniciativa>/ e serializa as operações
// que mexem no repositório com locks de arquivo (harness.AdquirirLock). Assim Builders e Reviewers em
// processos diferentes, ou depois de uma queda, enxergam o mesmo quadro.

const (
	leaseEsteira  = 2 * time.Minute
	esperaEsteira = 90 * time.Second
	janelaValida  = 15 * time.Minute
)

// dirTarefas devolve .forge/runtime/tarefas/<iniciativa>.
func dirTarefas(dir, ini string) string {
	return filepath.Join(layout.ForgeRuntime(dir), "tarefas", ini)
}

// EstadoNaEsteira é o estado persistido de uma task na esteira.
type EstadoNaEsteira struct {
	Task    string    `json:"task"`
	Estado  string    `json:"estado"`
	Motivo  string    `json:"motivo,omitempty"`
	Commit  string    `json:"commit,omitempty"`
	Gravado time.Time `json:"gravado"`
}

func slugSeguro(s string) bool { return wtIDRe.MatchString(s) }

// gravarJSON grava de forma atômica (arquivo temporário e rename).
func gravarJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SalvarPai guarda o feature_build que nomeia os task_build da iniciativa: o despacho contínuo
// precisa dele para derivar os próximos jobs depois que o dispatch original já terminou.
func SalvarPai(dir string, pai bus.Job) error {
	if !slugSeguro(pai.Initiative) {
		return fmt.Errorf("iniciativa %q inválida", pai.Initiative)
	}
	pai.Lease, pai.Events = nil, nil
	return gravarJSON(filepath.Join(dirTarefas(dir, pai.Initiative), "pai.json"), pai)
}

// LerPai lê o pai salvo por SalvarPai.
func LerPai(dir, ini string) (bus.Job, error) {
	var j bus.Job
	if !slugSeguro(ini) {
		return j, fmt.Errorf("iniciativa %q inválida", ini)
	}
	data, err := os.ReadFile(filepath.Join(dirTarefas(dir, ini), "pai.json"))
	if err != nil {
		return j, err
	}
	err = json.Unmarshal(data, &j)
	return j, err
}

// GravarEstadoTask registra o estado de uma task (planID = "feat-122-01").
func GravarEstadoTask(dir, ini, planID, estado, motivo, commit string, now time.Time) error {
	if !slugSeguro(ini) || !slugSeguro(planID) {
		return fmt.Errorf("iniciativa %q ou task %q inválida", ini, planID)
	}
	return gravarJSON(filepath.Join(dirTarefas(dir, ini), planID+".json"),
		EstadoNaEsteira{Task: planID, Estado: estado, Motivo: higieniza(motivo, 400, false), Commit: commit, Gravado: now.UTC()})
}

// EstadosDaIniciativa monta o quadro do despachante: o estado gravado de cada task e, por cima, as
// tasks com job pendente ou em processamento (Builder ou Reviewer) como em andamento.
func EstadosDaIniciativa(dir, ini string) (map[string]string, error) {
	out := map[string]string{}
	if !slugSeguro(ini) {
		return out, fmt.Errorf("iniciativa %q inválida", ini)
	}
	ents, err := os.ReadDir(dirTarefas(dir, ini))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return out, err
	}
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, "feat-") || !strings.HasSuffix(n, ".json") {
			continue
		}
		data, rerr := os.ReadFile(filepath.Join(dirTarefas(dir, ini), n))
		if rerr != nil {
			continue
		}
		var st EstadoNaEsteira
		if json.Unmarshal(data, &st) == nil && st.Task != "" && st.Estado != "" {
			out[st.Task] = st.Estado
		}
	}
	for _, role := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		for _, state := range []string{layout.StatePending, layout.StateProcessing} {
			entries, lerr := bus.List(dir, role, state)
			if lerr != nil {
				continue
			}
			for _, en := range entries {
				if en.Job.Initiative != ini || en.Job.TaskID == "" {
					continue
				}
				id := PlanTaskID(en.Job.TaskID)
				if out[id] != TaskMesclada {
					out[id] = OndaEmAndamento
				}
			}
		}
	}
	return out, nil
}

// travar adquire um lock de arquivo esperando até esperaEsteira; devolve a função de liberação.
func travar(dir, chave, dono string) (func(), error) {
	if len(chave) > 100 {
		chave = strings.TrimRight(chave[:100], "-")
	}
	if strings.TrimSpace(dono) == "" {
		dono = "forge-sdd"
	}
	limite := time.Now().Add(esperaEsteira)
	for {
		l, err := harness.AdquirirLock(dirLocks(dir), chave, dono, leaseEsteira, time.Now())
		if err == nil {
			return func() { _ = l.Liberar() }, nil
		}
		if !errors.Is(err, harness.ErrLockOcupado) || time.Now().After(limite) {
			return nil, err
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// GarantirBranchDaFeature cria feat/<iniciativa> a partir de base quando ela ainda não existe.
func GarantirBranchDaFeature(dir, ini, base string) error {
	feat := layout.InitiativeBranch(ini)
	if !refSegura(feat) {
		return fmt.Errorf("branch %q inválida", feat)
	}
	if _, err := layout.Git(dir, "rev-parse", "--git-dir"); err != nil {
		return err // fora de repositório: layout.ErrNotRepo
	}
	if _, err := gitLimpo(dir, "rev-parse", "--verify", "-q", "refs/heads/"+feat); err == nil {
		return nil
	}
	if base == "" || base == feat {
		base = "HEAD"
	}
	if base != "HEAD" && !refSegura(base) {
		return fmt.Errorf("base %q inválida", base)
	}
	if _, err := gitLimpo(dir, "branch", feat, base); err != nil {
		return fmt.Errorf("criar %s a partir de %s: %w", feat, base, err)
	}
	return nil
}

// WorktreeDaTask é o worktree efêmero preparado para um task_build.
type WorktreeDaTask struct {
	Path     string
	Retomada []string // arquivos sujos herdados de uma execução interrompida (RN-16)
	Reaberto bool     // recriado da branch da task (retrabalho ou retomada sem worktree)
}

// PrepararWorktreeDaTask garante o worktree da task antes de o Builder rodar: reaproveita o que
// sobrou de uma queda (sujo vira retomada), reabre da branch da task quando ela já existe
// (retrabalho) ou cria da branch da feature. Serializado entre processos (git worktree não é).
func PrepararWorktreeDaTask(dir string, job bus.Job, dono string) (*WorktreeDaTask, error) {
	if job.Type != bus.TypeTaskBuild || job.TaskID == "" {
		return nil, fmt.Errorf("job %s não é task_build", job.ID)
	}
	ini, planID := job.Initiative, PlanTaskID(job.TaskID)
	path, err := WorktreePath(dir, ini, planID)
	if err != nil {
		return nil, err
	}
	solta, err := travar(dir, "worktrees-admin", dono)
	if err != nil {
		return nil, err
	}
	defer solta()
	if st, serr := os.Stat(path); serr == nil && st.IsDir() {
		if _, gerr := gitLimpo(path, "rev-parse", "--git-dir"); gerr == nil {
			out, _ := gitLimpo(path, "status", "--porcelain", "--untracked-files=all", "--", ".", ":(exclude).forge")
			return &WorktreeDaTask{Path: path, Retomada: arquivosDoStatus(out)}, nil
		}
		// pasta sem worktree válido: deixa o git limpar o registro e recria
		_, _ = gitLimpo(dir, "worktree", "prune")
		if rerr := os.RemoveAll(path); rerr != nil {
			return nil, rerr
		}
	}
	_, _ = gitLimpo(dir, "worktree", "prune")
	if _, gerr := gitLimpo(dir, "rev-parse", "--verify", "-q", "refs/heads/"+TaskBranch(ini, job.TaskID)); gerr == nil {
		p, rerr := ReopenWorktree(dir, ini, planID)
		if rerr != nil {
			return nil, rerr
		}
		return &WorktreeDaTask{Path: p, Reaberto: true}, nil
	}
	if err := GarantirBranchDaFeature(dir, ini, ""); err != nil {
		return nil, err
	}
	p, err := CreateWorktree(dir, ini, planID, layout.InitiativeBranch(ini))
	if err != nil {
		return nil, err
	}
	return &WorktreeDaTask{Path: p}, nil
}

func arquivosDoStatus(porcelain string) []string {
	var out []string
	for _, l := range strings.Split(porcelain, "\n") {
		if len(l) > 3 {
			out = append(out, strings.TrimSpace(l[3:]))
		}
	}
	return out
}

// CopiarRetorno traz o retorno JSON gravado pelo Builder dentro do worktree para o runtime do
// repositório principal, onde a verificação o lê. Retorno ausente não é erro aqui.
func CopiarRetorno(dir, wt, jobID string) error {
	if wt == "" || wt == dir {
		return nil
	}
	data, err := os.ReadFile(RetornoPath(wt, jobID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	dst := RetornoPath(dir, jobID)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

// EncerrarWorktreeDaTask apaga o worktree da task quando ele está limpo (o trabalho está nos commits
// da branch da task, que fica). Sujo, ele fica para a retomada e devolve false.
func EncerrarWorktreeDaTask(dir string, job bus.Job, dono string) (bool, error) {
	ini, planID := job.Initiative, PlanTaskID(job.TaskID)
	path, err := WorktreePath(dir, ini, planID)
	if err != nil {
		return false, err
	}
	if _, serr := os.Stat(path); serr != nil {
		return true, nil
	}
	out, err := gitLimpo(path, "status", "--porcelain", "--untracked-files=all", "--", ".", ":(exclude).forge")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(out) != "" {
		return false, nil
	}
	solta, err := travar(dir, "worktrees-admin", dono)
	if err != nil {
		return false, err
	}
	defer solta()
	if err := RemoveWorktreeKeepBranch(dir, ini, planID); err != nil {
		return false, err
	}
	return true, nil
}

// ResultadoJuncaoTask é o desfecho da junção de uma task aprovada na branch da feature.
type ResultadoJuncaoTask struct {
	Estado     string // JuncaoIntegrada ou JuncaoBloqueada
	Commit     string
	Conflitos  []string
	Motivo     string
	FeatPronta bool // todas as tasks do plano estão mescladas
}

// gitSaida roda git com argumentos explícitos e devolve stdout e o código de saída.
func gitSaida(dir string, args ...string) (string, int, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANGUAGE=C")
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return so.String(), ee.ExitCode(), nil
	}
	if err != nil {
		return "", -1, err
	}
	return so.String(), 0, nil
}

// JuntarTaskNaFeature faz o squash da branch da task na branch da feature sem tocar em nenhum
// worktree (merge-tree e commit-tree): conflito bloqueia a task e preserva a branch; sucesso grava
// "mesclada" e apaga a branch da task. Em série por iniciativa (lock de arquivo).
func JuntarTaskNaFeature(dir string, job bus.Job, dono string, now time.Time) (*ResultadoJuncaoTask, error) {
	if job.TaskID == "" || !slugSeguro(job.Initiative) {
		return nil, fmt.Errorf("job %s sem task ou iniciativa válida", job.ID)
	}
	ini, planID := job.Initiative, PlanTaskID(job.TaskID)
	feat, branch := layout.InitiativeBranch(ini), TaskBranch(ini, job.TaskID)
	if !refSegura(feat) || !refSegura(branch) {
		return nil, fmt.Errorf("branch inválida: %q ou %q", feat, branch)
	}
	solta, err := travar(dir, "juncao-"+ini, dono)
	if err != nil {
		return nil, err
	}
	defer solta()
	if err := GarantirBranchDaFeature(dir, ini, ""); err != nil {
		return nil, err
	}
	featSHA, err := gitLimpo(dir, "rev-parse", "--verify", "refs/heads/"+feat)
	if err != nil {
		return nil, err
	}
	if _, err := gitLimpo(dir, "rev-parse", "--verify", "-q", "refs/heads/"+branch); err != nil {
		// Branch já apagada: a task foi mesclada antes (reaplicação idempotente).
		_ = GravarEstadoTask(dir, ini, planID, TaskMesclada, "branch da task já integrada", featSHA, now)
		return &ResultadoJuncaoTask{Estado: JuncaoIntegrada, Commit: featSHA, FeatPronta: featPronta(dir, ini)}, nil
	}
	out, code, err := gitSaida(dir, "merge-tree", "--write-tree", "--name-only", "--no-messages", feat, branch)
	if err != nil {
		return nil, err
	}
	linhas := strings.Split(strings.TrimSpace(out), "\n")
	if code == 1 {
		res := &ResultadoJuncaoTask{Estado: JuncaoBloqueada, Motivo: "conflito com a branch da feature"}
		for _, l := range linhas[1:] {
			if l = strings.TrimSpace(l); l != "" {
				res.Conflitos = append(res.Conflitos, l)
			}
		}
		_ = GravarEstadoTask(dir, ini, planID, OndaBloqueada, res.Motivo+": "+strings.Join(res.Conflitos, ", "), "", now)
		return res, nil
	}
	if code != 0 || len(linhas) == 0 {
		return nil, fmt.Errorf("git merge-tree saiu com código %d", code)
	}
	tree := strings.TrimSpace(linhas[0])
	featTree, _ := gitLimpo(dir, "rev-parse", featSHA+"^{tree}")
	novo := featSHA
	if tree != featTree {
		msg := fmt.Sprintf("feat(%s): %s", job.TaskID, higieniza(job.Title, 120, false))
		args := []string{}
		if !temIdentidadeGit(dir) {
			args = append(args, "-c", "user.name=forge-sdd", "-c", "user.email=forge-sdd@localhost")
		}
		args = append(args, "commit-tree", tree, "-p", featSHA, "-m", msg)
		if novo, err = gitLimpo(dir, args...); err != nil {
			return nil, fmt.Errorf("commit da junção: %w", err)
		}
		if err := avancarFeature(dir, feat, novo, featSHA); err != nil {
			_ = GravarEstadoTask(dir, ini, planID, OndaBloqueada, err.Error(), "", now)
			return &ResultadoJuncaoTask{Estado: JuncaoBloqueada, Motivo: err.Error()}, nil
		}
	}
	_, _ = gitLimpo(dir, "branch", "-D", branch)
	if err := GravarEstadoTask(dir, ini, planID, TaskMesclada, "", novo, now); err != nil {
		return nil, err
	}
	return &ResultadoJuncaoTask{Estado: JuncaoIntegrada, Commit: novo, FeatPronta: featPronta(dir, ini)}, nil
}

func temIdentidadeGit(dir string) bool {
	n, e1 := gitLimpo(dir, "config", "--get", "user.name")
	m, e2 := gitLimpo(dir, "config", "--get", "user.email")
	return e1 == nil && e2 == nil && n != "" && m != ""
}

// avancarFeature move feat para novo. Se o worktree principal está nessa branch, avança por
// fast forward só quando ele está limpo; sujo, recusa (nunca sobrescreve trabalho da pessoa).
func avancarFeature(dir, feat, novo, antigo string) error {
	atual, _ := gitLimpo(dir, "symbolic-ref", "-q", "--short", "HEAD")
	if atual == feat {
		st, err := gitLimpo(dir, "status", "--porcelain", "--untracked-files=no")
		if err != nil {
			return err
		}
		if st != "" {
			return fmt.Errorf("o worktree principal está em %s com mudanças: junção adiada", feat)
		}
		_, err = gitLimpo(dir, "merge", "--ff-only", "-q", novo)
		return err
	}
	_, err := gitLimpo(dir, "update-ref", "refs/heads/"+feat, novo, antigo)
	return err
}

// featPronta diz se todas as tasks do plano da iniciativa estão mescladas.
func featPronta(dir, ini string) bool {
	plano, err := planoDeTasks(dir, ini, 64)
	if err != nil || len(plano) == 0 {
		return false
	}
	est, _ := EstadosDaIniciativa(dir, ini)
	for _, b := range plano {
		for _, t := range b {
			if est[t.ID] != TaskMesclada {
				return false
			}
		}
	}
	return true
}

// LerJanelaDoPainel lê o percentual da janela de 5 horas gravado pelo mod forge-painel em
// .forge/runtime/painel/janela.json. Só vale se recente (15 min); sem ele, a janela é desconhecida.
func LerJanelaDoPainel(dir string, now time.Time) (int, bool) {
	data, err := os.ReadFile(filepath.Join(layout.ForgeRuntime(dir), "painel", "janela.json"))
	if err != nil {
		return 0, false
	}
	var j struct {
		CincoHoras *float64 `json:"five_hour_pct"`
		SeteDias   *float64 `json:"seven_day_pct"`
		Em         string   `json:"at"`
	}
	if json.Unmarshal(data, &j) != nil {
		return 0, false
	}
	em, err := time.Parse(time.RFC3339, j.Em)
	if err != nil || now.Sub(em) > janelaValida || em.Sub(now) > time.Minute {
		return 0, false
	}
	pct := -1.0
	for _, v := range []*float64{j.CincoHoras, j.SeteDias} {
		if v != nil && *v > pct {
			pct = *v
		}
	}
	if pct < 0 {
		return 0, false
	}
	if pct > 100 {
		pct = 100
	}
	return int(pct), true
}

// DespacharIniciativa roda uma rodada do despachante para a iniciativa (RN-7): chamada a cada worktree
// apagado ou task mesclada. Sem pai salvo (modo feature ou dispatch antigo) não faz nada.
func DespacharIniciativa(dir, ini string, now time.Time) (*ResultadoDespacho, error) {
	pai, err := LerPai(dir, ini)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ResultadoDespacho{}, nil
		}
		return nil, err
	}
	cfg, _, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, err
	}
	plano, err := planoDeTasks(dir, ini, cfg.Builder.LarguraDePlano())
	if err != nil {
		return nil, err
	}
	estados, err := EstadosDaIniciativa(dir, ini)
	if err != nil {
		return nil, err
	}
	return Despachar(EntradaDespacho{
		Dir: dir, Pai: pai, Plano: plano, Estados: estados, Builder: cfg.Builder,
		Janela: func() (int, bool) { return LerJanelaDoPainel(dir, now) },
		Now:    now,
	})
}
