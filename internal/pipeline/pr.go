package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// PRTemplatePath é o modelo do corpo do PR (FLX-03); suas seções "## " são obrigatórias no corpo.
const PRTemplatePath = ".github/pull_request_template.md"

// CmdRunner roda `name args...` em dir e devolve a saída; só git e gh passam (SEG-03). Injetável.
type CmdRunner func(dir, name string, args ...string) (string, error)

// RunGitOrGh é o CmdRunner real: argv explícito, sem shell.
func RunGitOrGh(dir, name string, args ...string) (string, error) {
	if name != "git" && name != "gh" {
		return "", fmt.Errorf("comando %q não permitido", name)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(stdout.String()), fmt.Errorf("%s %s: %w: %s", name, args[0], err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// PROptions controla `forge pipeline pr`.
type PROptions struct {
	JobID    string
	Worker   string
	BodyFile string // relativo ao projeto
	Run      CmdRunner
	// Gate roda a Camada A; padrão: gate.RunChecks. Devolve as falhas (vazio = passou).
	Gate func(dir, branch, specPath, bodyPath string) ([]string, error)
	Now  time.Time
}

// PRResult descreve o PR aberto ou corrigido.
type PRResult struct {
	Number     int
	URL        string
	Branch     string
	Base       string
	BaseSource string
	Created    bool // false: o PR já existia
	BaseFixed  bool // a base de um PR já aberto foi corrigida
}

var prNumberRe = regexp.MustCompile(`/pull/([0-9]+)\s*$`)

// OpenPR é o passo mecânico do Builder feito pelo Go: Camada A, push com refspec explícito, PR com a
// base da iniciativa (ou correção da base de um PR já aberto), registro no job e tasks em
// review_pending. Nunca mescla nem aprova (ADR-0004).
func OpenPR(dir string, opts PROptions) (*PRResult, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotPipelineMode
	}
	run := opts.Run
	if run == nil {
		run = RunGitOrGh
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	job, _, err := bus.Get(dir, layout.RoleBuilder, opts.JobID)
	if err != nil {
		return nil, err
	}
	branch := job.PR.Branch
	if err := queue.ValidBranchName(branch); err != nil {
		return nil, fmt.Errorf("branch do job: %w", err)
	}
	base, source := prBase(dir, cfg, job, run)
	if branch == "main" || branch == "master" || branch == base {
		return nil, fmt.Errorf("recusado: a branch %q é a base ou o tronco; o PR sai de uma branch de iniciativa (FLX-02)", branch)
	}

	body, err := readPRBody(dir, opts.BodyFile)
	if err != nil {
		return nil, err
	}
	if missing := missingPRSections(dir, body); len(missing) > 0 {
		return nil, fmt.Errorf("o corpo do PR não segue %s: faltam as seções %s (FLX-03)", PRTemplatePath, strings.Join(missing, ", "))
	}
	if cur, err := run(dir, "git", "rev-parse", "--abbrev-ref", "HEAD"); err != nil || cur != branch {
		return nil, fmt.Errorf("a branch atual (%q) não é a do job (%q); faça checkout dela antes", cur, branch)
	}
	if n := commitsAhead(dir, base, branch, run); n == 0 {
		return nil, fmt.Errorf("a branch %q não tem commit à frente de %q; faça o commit antes de abrir o PR", branch, base)
	}

	gateFn := opts.Gate
	if gateFn == nil {
		gateFn = runGateA
	}
	failures, err := gateFn(dir, branch, job.SpecPath, opts.BodyFile)
	if err != nil {
		return nil, err
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("Camada A reprovada; corrija e rode de novo:\n  - %s", strings.Join(failures, "\n  - "))
	}

	if _, err := run(dir, "git", "push", "origin", branch+":"+branch); err != nil {
		return nil, err
	}

	res := &PRResult{Branch: branch, Base: base, BaseSource: source}
	existing, err := findOpenPR(dir, branch, run)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		res.Number, res.URL = existing.Number, existing.URL
		if existing.Base != base {
			if _, err := run(dir, "gh", "pr", "edit", strconv.Itoa(existing.Number), "--base", base); err != nil {
				return nil, err
			}
			res.BaseFixed = true
		}
	} else {
		url, err := run(dir, "gh", "pr", "create", "--base", base, "--head", branch, "--title", prTitle(job), "--body-file", opts.BodyFile)
		if err != nil {
			return nil, err
		}
		m := prNumberRe.FindStringSubmatch(url)
		if m == nil {
			return nil, fmt.Errorf("não consegui ler o número do PR na saída do gh: %q", truncateStr(url, 80))
		}
		res.Number, _ = strconv.Atoi(m[1])
		res.URL, res.Created = url, true
	}

	if _, err := bus.UpdatePR(dir, layout.RoleBuilder, opts.JobID, opts.Worker, res.Number, res.URL, branch, now); err != nil {
		return nil, err
	}
	if _, err := queue.SetTaskStatuses(dir, job.Initiative, job.TaskIDs, queue.StatusReviewPending); err != nil {
		return nil, fmt.Errorf("atualizar a fila da iniciativa: %w", err)
	}
	return res, nil
}

// prBase resolve a base: fix (branch fix/…) usa a base já gravada no job, que é a branch da
// iniciativa; as demais usam o resolvedor único da iniciativa.
func prBase(dir string, cfg config.ForgeConfig, job *bus.Job, run CmdRunner) (string, string) {
	if strings.HasPrefix(job.PR.Branch, "fix/") && job.PR.BaseBranch != "" {
		return job.PR.BaseBranch, "job"
	}
	git := func(d string, args ...string) (string, error) { return run(d, "git", args...) }
	return queue.ResolveBase(dir, job.Initiative, "", cfg.BaseBranch, git)
}

func readPRBody(dir, bodyFile string) (string, error) {
	if strings.TrimSpace(bodyFile) == "" {
		return "", errors.New("--body-file é obrigatório (corpo do PR pelo modelo de .github/pull_request_template.md)")
	}
	path, err := gate.ConfineToDir(dir, bodyFile)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ler o corpo do PR: %w", err)
	}
	return string(data), nil
}

// missingPRSections devolve as seções "## " do modelo que o corpo não tem.
func missingPRSections(dir, body string) []string {
	tpl, err := os.ReadFile(filepath.Join(dir, PRTemplatePath))
	if err != nil {
		return nil // sem modelo no projeto, nada a exigir
	}
	have := map[string]bool{}
	for _, m := range sectionRe.FindAllStringSubmatch(body, -1) {
		have[strings.ToLower(strings.TrimSpace(m[1]))] = true
	}
	var missing []string
	for _, m := range sectionRe.FindAllStringSubmatch(string(tpl), -1) {
		title := strings.TrimSpace(m[1])
		if !have[strings.ToLower(title)] {
			missing = append(missing, fmt.Sprintf("%q", title))
		}
	}
	return missing
}

// commitsAhead conta os commits de branch que a base não tem; tenta a base local e depois origin.
func commitsAhead(dir, base, branch string, run CmdRunner) int {
	for _, b := range []string{base, "origin/" + base} {
		out, err := run(dir, "git", "rev-list", "--count", b+".."+branch)
		if err != nil {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
			return n
		}
	}
	return 0
}

type openPR struct {
	Number int
	URL    string
	Base   string
}

func findOpenPR(dir, branch string, run CmdRunner) (*openPR, error) {
	out, err := run(dir, "gh", "pr", "list", "--head", branch, "--state", "open", "--json", "number,url,baseRefName", "--limit", "1")
	if err != nil {
		return nil, err
	}
	var list []struct {
		Number      int    `json:"number"`
		URL         string `json:"url"`
		BaseRefName string `json:"baseRefName"`
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, fmt.Errorf("saída do gh pr list inválida: %w", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &openPR{Number: list[0].Number, URL: list[0].URL, Base: list[0].BaseRefName}, nil
}

func prTitle(job *bus.Job) string {
	t := strings.Join(strings.Fields(job.Title), " ")
	if t == "" {
		t = job.Feature
	}
	if t == "" {
		t = job.ID
	}
	return truncateStr(t, 100)
}

func truncateStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// runGateA roda a Camada A e devolve as verificações reprovadas.
func runGateA(dir, branch, specPath, bodyPath string) ([]string, error) {
	results, err := gate.RunChecks(dir, gate.Options{Branch: branch, SpecPath: specPath, PRBodyPath: bodyPath, Now: time.Now().UTC()})
	if err != nil {
		return nil, err
	}
	var failed []string
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, fmt.Sprintf("[%s] (%s) %s", r.ID, r.Rule, r.Evidence))
		}
	}
	return failed, nil
}
