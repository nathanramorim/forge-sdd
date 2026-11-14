package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// FormatLiveStatus consolida a verificação rápida de status em tempo real (tríplice):
// 1. Ao Vivo no Painel (painel tmux ou últimas linhas de log de runtime)
// 2. Commits Recentes no Git
// 3. Fila de Tarefas
// Retorna a saída canônica formatada (< 200 tokens).
func FormatLiveStatus(dir string) (string, error) {
	return FormatLiveStatusWith(dir, nil, nil)
}

// GitRunnerFn define o contrato de busca de commits para testes e desacoplamento.
type GitRunnerFn func(dir string, max int) ([]string, error)

// FormatLiveStatusWith permite injeção de dependências para testes determinísticos.
func FormatLiveStatusWith(dir string, deliverer TmuxDeliverer, gitFn GitRunnerFn) (string, error) {
	if dir == "" {
		dir = "."
	}

	absDir, err := filepath.Abs(dir)
	proj := TmuxProjectName(dir)
	if err == nil {
		proj = TmuxProjectName(absDir)
	}
	branch := currentBranch(dir)

	role := detectActiveRole(dir)
	liveLines := captureLiveLines(dir, role, deliverer)

	if gitFn == nil {
		gitFn = gitRecentCommits
	}
	commits, err := gitFn(dir, 5)
	if err != nil {
		commits = nil
	}

	q, _ := queue.Load(dir)

	var b strings.Builder
	fmt.Fprintf(&b, "### Status da Esteira — %s (%s)\n\n", proj, branch)

	// 1. Ao Vivo no Painel
	fmt.Fprintf(&b, "1. Ao Vivo no Painel (%s):\n", role)
	if len(liveLines) == 0 {
		b.WriteString("   • Nenhum evento recente no painel ou log de runtime.\n")
		b.WriteString("   • Esteira aguardando novos jobs.\n")
	} else {
		for _, l := range liveLines {
			fmt.Fprintf(&b, "   • %s\n", l)
		}
	}
	b.WriteString("\n")

	// 2. Commits Recentes no Git
	b.WriteString("2. Commits Recentes no Git:\n")
	if len(commits) == 0 {
		b.WriteString("   • (nenhum commit recente)\n")
	} else {
		for _, c := range commits {
			fmt.Fprintf(&b, "   • %s\n", c)
		}
	}
	b.WriteString("\n")

	// 3. Fila de Tarefas
	b.WriteString("3. Fila de Tarefas:\n")
	if q == nil || len(q.Items) == 0 {
		b.WriteString("   • (nenhuma tarefa na fila)\n")
	} else {
		for _, it := range q.Items {
			mark := " "
			if it.Status == queue.StatusApproved || it.Status == "done" {
				mark = "x"
			}
			fmt.Fprintf(&b, "   • [%s] %s: %s\n", mark, it.TaskID, it.Title)
		}
	}

	// Próximo passo
	var nextStep string
	switch {
	case role == "Reviewer":
		nextStep = "Revisão do Reviewer"
	case role == "Builder":
		nextStep = "Acompanhar o Builder"
	case q != nil && len(q.Items) > 0:
		allApproved := true
		for _, it := range q.Items {
			if it.Status != queue.StatusApproved && it.Status != "done" {
				allApproved = false
				break
			}
		}
		if allApproved {
			nextStep = "/archive ou novo /discovery"
		} else {
			nextStep = "Acompanhar o Builder"
		}
	default:
		nextStep = "/discovery"
	}
	fmt.Fprintf(&b, "\nPróximo passo: %s\n", nextStep)

	return b.String(), nil
}

func currentBranch(dir string) string {
	headPath := filepath.Join(dir, ".git", "HEAD")
	if data, err := os.ReadFile(headPath); err == nil {
		s := strings.TrimSpace(string(data))
		if strings.HasPrefix(s, "ref: refs/heads/") {
			return strings.TrimPrefix(s, "ref: refs/heads/")
		}
	}
	br, err := layout.CurrentBranch(dir)
	if err == nil && br != "" {
		return br
	}
	return "desconhecida"
}

func detectActiveRole(dir string) string {
	rev, err := bus.List(dir, layout.RoleReviewer, layout.StateProcessing)
	if err == nil && len(rev) > 0 {
		return "Reviewer"
	}
	bld, err := bus.List(dir, layout.RoleBuilder, layout.StateProcessing)
	if err == nil && len(bld) > 0 {
		return "Builder"
	}
	return "Ocioso"
}

func captureLiveLines(dir string, role string, deliverer TmuxDeliverer) []string {
	proj := TmuxProjectName(dir)
	sess := TmuxSessionName(dir)
	exact := "=" + sess
	win := exact + ":forge:" + proj

	if deliverer == nil {
		deliverer = DefaultTmuxDeliverer()
	}

	targetPane := win + ".1"
	if role == "Reviewer" {
		targetPane = win + ".2"
	}
	out, err := deliverer.CapturePane(targetPane)
	if err == nil {
		lines := extractLastInformativeLines(out, 2)
		if len(lines) > 0 {
			return lines
		}
	}

	return readLastLogLines(dir, role)
}

func readLastLogLines(dir string, role string) []string {
	logsDir := filepath.Join(dir, ".forge", "runtime", "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil || len(entries) == 0 {
		return []string{
			"Nenhum evento recente no painel ou log de runtime.",
			"Esteira aguardando novos jobs.",
		}
	}

	type logFile struct {
		path    string
		modTime time.Time
		role    bool
	}
	var files []logFile
	rolePrefix := strings.ToLower(role)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		isRole := rolePrefix != "" && rolePrefix != "ocioso" && strings.HasPrefix(strings.ToLower(entry.Name()), rolePrefix)
		files = append(files, logFile{
			path:    filepath.Join(logsDir, entry.Name()),
			modTime: info.ModTime(),
			role:    isRole,
		})
	}
	if len(files) == 0 {
		return []string{
			"Nenhum evento recente no painel ou log de runtime.",
			"Esteira aguardando novos jobs.",
		}
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].role != files[j].role {
			return files[i].role
		}
		return files[i].modTime.After(files[j].modTime)
	})

	for _, f := range files {
		data, err := os.ReadFile(f.path)
		if err != nil {
			continue
		}
		lines := extractLastInformativeLines(string(data), 2)
		if len(lines) > 0 {
			return lines
		}
	}

	return []string{
		"Nenhum evento recente no painel ou log de runtime.",
		"Esteira aguardando novos jobs.",
	}
}

func extractLastInformativeLines(content string, count int) []string {
	cleaned := ansiRe.ReplaceAllString(content, "")
	rawLines := strings.Split(cleaned, "\n")
	var nonBlank []string
	for _, line := range rawLines {
		t := strings.TrimSpace(line)
		if t != "" {
			nonBlank = append(nonBlank, t)
		}
	}
	if len(nonBlank) == 0 {
		return nil
	}
	if len(nonBlank) <= count {
		return nonBlank
	}
	return nonBlank[len(nonBlank)-count:]
}

func gitRecentCommits(dir string, max int) ([]string, error) {
	if max <= 0 {
		max = 5
	}
	cmd := exec.Command("git", "log", fmt.Sprintf("-n%d", max), "--oneline")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler git log: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	var lines []string
	for _, l := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines, nil
}
