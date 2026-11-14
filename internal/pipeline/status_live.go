package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
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

// roleTokens acumula tokens estimados ou medidos em sdd/.metrics/.
type roleTokens struct {
	Discovery int64
	Builder   int64
	Reviewer  int64
	Total     int64
}

func readMetricsTokens(dir string) roleTokens {
	metricsDir := filepath.Join(dir, "sdd", ".metrics")
	entries, err := os.ReadDir(metricsDir)
	if err != nil {
		return roleTokens{}
	}
	var res roleTokens
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(metricsDir, e.Name()))
		if err != nil {
			continue
		}
		var item struct {
			AgentPath    []string `json:"agent_path"`
			Phase        string   `json:"phase"`
			TokensInput  int64    `json:"tokens_input"`
			TokensOutput int64    `json:"tokens_output"`
		}
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}
		toks := item.TokensInput + item.TokensOutput
		res.Total += toks
		pathStr := strings.ToLower(strings.Join(item.AgentPath, " ") + " " + item.Phase)
		switch {
		case strings.Contains(pathStr, "builder"):
			res.Builder += toks
		case strings.Contains(pathStr, "reviewer"):
			res.Reviewer += toks
		default:
			res.Discovery += toks
		}
	}
	return res
}

func formatTokens(tokens int64) string {
	if tokens <= 0 {
		return "—"
	}
	if tokens < 1000 {
		return fmt.Sprintf("%d tok", tokens)
	}
	return fmt.Sprintf("%.1fk tok", float64(tokens)/1000.0)
}

func formatProgressBar(completed, total int, width int) string {
	if total <= 0 {
		return "[────────────────────] 0%"
	}
	if width <= 0 {
		width = 20
	}
	ratio := float64(completed) / float64(total)
	if ratio > 1.0 {
		ratio = 1.0
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	pct := int(ratio * 100)
	return fmt.Sprintf("[%s%s] %d%% (%d/%d tasks)",
		strings.Repeat("█", filled),
		strings.Repeat("░", empty),
		pct, completed, total)
}

// FormatRichDashboard gera a interface visual completa do painel vivo da esteira.
func FormatRichDashboard(dir string, deliverer TmuxDeliverer, gitFn GitRunnerFn, report *monitor.ProjectProcessReport) (string, error) {
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
	commits, _ := gitFn(dir, 3)

	if report == nil {
		report, _ = monitor.InspectProjectProcesses(dir, monitor.DefaultCommandRunner)
	}

	toks := readMetricsTokens(dir)
	q, _ := queue.Load(dir)

	var b bytes.Buffer

	// 1. Cabeçalho em Caixa
	roleBadge := "OCIOSO (AGUARDANDO JOBS)"
	if role == "Builder" {
		roleBadge = "⚡ BUILDER EM ATIVIDADE"
	} else if role == "Reviewer" {
		roleBadge = "🔍 REVIEWER EM REVISÃO"
	}

	fmt.Fprintf(&b, "╭────────────────────────────────────────────────────────────────────────────╮\n")
	fmt.Fprintf(&b, "│ 📦 %-24s │ Branch: %-37s │\n", proj, branch)
	fmt.Fprintf(&b, "│ %-74s │\n", roleBadge)
	fmt.Fprintf(&b, "╰────────────────────────────────────────────────────────────────────────────╯\n\n")

	// 2. Processamento e Consumo por Sessão
	fmt.Fprintf(&b, "─── 🖥️  PROCESSAMENTO & CONSUMO POR SESSÃO ──────────────────────────────────\n")
	fmt.Fprintf(&b, "%-22s %-12s %-8s %-12s %s\n", "PAPEL", "ESTADO", "CPU %", "MEM (RAM)", "TOKENS ACUMULADOS")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 76))

	if report != nil && len(report.Processes) > 0 {
		for _, p := range report.Processes {
			memStr := fmt.Sprintf("%.0f MB", p.MemoryMB)
			cpuStr := fmt.Sprintf("%.1f%%", p.CPUPercent)
			tokStr := "—"
			roleLower := strings.ToLower(p.Role)
			switch {
			case strings.Contains(roleLower, "discovery"):
				tokStr = formatTokens(toks.Discovery)
			case strings.Contains(roleLower, "builder"):
				tokStr = formatTokens(toks.Builder)
			case strings.Contains(roleLower, "reviewer"):
				tokStr = formatTokens(toks.Reviewer)
			}
			fmt.Fprintf(&b, "%-22s %-12s %-8s %-12s %s\n", p.Role, p.State, cpuStr, memStr, tokStr)
		}
		fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 76))
		fmt.Fprintf(&b, "Total do Projeto: %d processos │ CPU: %.1f%% │ Memória: %.0f MB │ Tokens: %s\n\n",
			len(report.Processes), report.TotalCPU, report.TotalMemory, formatTokens(toks.Total))
	} else {
		fmt.Fprintf(&b, "Nenhum processo ativo no momento para esta sessão.\n\n")
	}

	// 3. Progresso da Esteira (DAG)
	fmt.Fprintf(&b, "─── 📊 PROGRESSO DA ESTEIRA (DAG) ──────────────────────────────────────────\n")
	if q == nil || len(q.Items) == 0 {
		fmt.Fprintf(&b, "(nenhuma tarefa enfileirada na iniciativa atual)\n\n")
	} else {
		completed := 0
		for _, it := range q.Items {
			if it.Status == queue.StatusApproved || it.Status == "done" || it.Status == "completed" {
				completed++
			}
		}
		fmt.Fprintf(&b, "Progresso: %s\n", formatProgressBar(completed, len(q.Items), 20))
		for _, it := range q.Items {
			mark := "[ ]"
			switch it.Status {
			case queue.StatusApproved, "done", "completed":
				mark = "[✓]"
			case "in_progress", "processing":
				mark = "[▶]"
			case "review_pending":
				mark = "[🔍]"
			}
			fmt.Fprintf(&b, "%s %s: %s\n", mark, it.TaskID, it.Title)
		}
		b.WriteString("\n")
	}

	// 4. Ao Vivo no Painel Ativo
	fmt.Fprintf(&b, "─── 📡 AO VIVO NO PAINEL ATIVO (%s) ─────────────────────────────\n", role)
	if len(liveLines) == 0 {
		b.WriteString("• Nenhum evento recente no painel ou log de runtime.\n")
		b.WriteString("• Esteira em espera.\n")
	} else {
		for _, l := range liveLines {
			fmt.Fprintf(&b, "• %s\n", l)
		}
	}
	b.WriteString("\n")

	// 5. Commits Recentes
	fmt.Fprintf(&b, "─── 🌿 COMMITS RECENTES NO GIT ─────────────────────────────────────────────\n")
	if len(commits) == 0 {
		b.WriteString("• (nenhum commit recente)\n")
	} else {
		for _, c := range commits {
			fmt.Fprintf(&b, "• %s\n", c)
		}
	}
	b.WriteString("\n")

	// 6. Próximo Passo
	var nextStep string
	switch {
	case role == "Reviewer":
		nextStep = "Revisão do Reviewer"
	case role == "Builder":
		nextStep = "Acompanhar o Builder"
	case q != nil && len(q.Items) > 0:
		allApproved := true
		for _, it := range q.Items {
			if it.Status != queue.StatusApproved && it.Status != "done" && it.Status != "completed" {
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
	fmt.Fprintf(&b, "Próximo passo: %s\n", nextStep)

	return b.String(), nil
}

// RunLiveDashboard opera o painel vivo contínuo ou em modo snapshot (once).
func RunLiveDashboard(ctx context.Context, dir string, out io.Writer, once bool) error {
	if out == nil {
		out = os.Stdout
	}
	if ctx == nil {
		ctx = context.Background()
	}

	render := func() error {
		content, err := FormatRichDashboard(dir, nil, nil, nil)
		if err != nil {
			return err
		}
		if !once {
			// Limpa a tela e reseta cursor no topo
			fmt.Fprint(out, "\033[H\033[2J")
		}
		fmt.Fprint(out, content)
		if !once {
			fmt.Fprint(out, "\n[Atualizando a cada 2s • Pressione Ctrl+C para sair do painel vivo]\n")
		}
		return nil
	}

	if once {
		return render()
	}

	// Primeira renderização imediata
	if err := render(); err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(out, "\nPainel vivo encerrado.")
			return nil
		case <-ticker.C:
			if err := render(); err != nil {
				return err
			}
		}
	}
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
