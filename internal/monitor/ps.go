package monitor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/forge-sdd/cli/internal/pipeline"
)

// ProcessInfo contém os dados de telemetria de um processo ligado ao projeto.
type ProcessInfo struct {
	PID         int     `json:"pid"`
	ParentPID   int     `json:"parent_pid,omitempty"`
	Role        string  `json:"role"`
	State       string  `json:"state"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryMB    float64 `json:"memory_mb"`
	ElapsedTime string  `json:"elapsed_time"`
	Command     string  `json:"command"`
}

// ProjectProcessReport contém a lista agregada de processos e consumo total do projeto.
type ProjectProcessReport struct {
	Project     string        `json:"project"`
	Session     string        `json:"session"`
	Processes   []ProcessInfo `json:"processes"`
	TotalCPU    float64       `json:"total_cpu"`
	TotalMemory float64       `json:"total_memory_mb"`
}

// CommandRunner abstrai a execução de comandos do sistema operacional (mockável para testes).
type CommandRunner func(name string, args ...string) ([]byte, error)

// DefaultCommandRunner executa os comandos do SO diretamente.
func DefaultCommandRunner(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

type rawProc struct {
	PID     int
	PPID    int
	CPU     float64
	RSSKB   int64
	Elapsed string
	Command string
}

// InspectProjectProcesses inspeciona a sessão tmux e os processos do projeto.
func InspectProjectProcesses(dir string, runner CommandRunner) (*ProjectProcessReport, error) {
	if runner == nil {
		runner = DefaultCommandRunner
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	projectName := filepath.Base(absDir)
	sessionName := pipeline.TmuxSessionName(absDir)

	report := &ProjectProcessReport{
		Project:   projectName,
		Session:   sessionName,
		Processes: []ProcessInfo{},
	}

	// 1. Obter snapshot de todos os processos do sistema
	psOutput, err := runner("ps", "-eo", "pid,ppid,%cpu,rss,etime,command")
	if err != nil {
		return report, fmt.Errorf("falha ao executar ps: %w", err)
	}

	rawMap := parsePSOutput(string(psOutput))

	// 2. Tentar listar painéis do tmux da sessão
	paneTargets := make(map[int]string) // PID -> Role
	tmuxOut, err := runner("tmux", "list-panes", "-s", "-t", "="+sessionName, "-F", "#{pane_pid}|#{pane_title}|#{window_name}")
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(tmuxOut)), "\n") {
			parts := strings.Split(strings.TrimSpace(line), "|")
			if len(parts) >= 3 {
				pid, err := strconv.Atoi(parts[0])
				if err == nil {
					role := parts[1]
					win := parts[2]
					if strings.Contains(win, "ctl") {
						role = "forge-ctl (" + role + ")"
					}
					paneTargets[pid] = role
				}
			}
		}
	}

	// 3. Mapear árvore de processos relacionados à sessão ou ao diretório
	visited := make(map[int]bool)
	var matchedPIDs []int

	for pid := range paneTargets {
		matchedPIDs = append(matchedPIDs, pid)
		visited[pid] = true
	}

	// Se não achou painéis tmux, busca processos que contenham o diretório ou sessionName
	if len(matchedPIDs) == 0 {
		for pid, p := range rawMap {
			if strings.Contains(p.Command, absDir) || strings.Contains(p.Command, sessionName) {
				matchedPIDs = append(matchedPIDs, pid)
				visited[pid] = true
			}
		}
	}

	// Coleta filhos recursivamente
	var collectChildren func(parentPID int)
	collectChildren = func(parentPID int) {
		for pid, p := range rawMap {
			if p.PPID == parentPID && !visited[pid] {
				visited[pid] = true
				matchedPIDs = append(matchedPIDs, pid)
				collectChildren(pid)
			}
		}
	}

	for _, rootPID := range append([]int(nil), matchedPIDs...) {
		collectChildren(rootPID)
	}

	// 4. Montar a lista de ProcessInfo
	for _, pid := range matchedPIDs {
		p, exists := rawMap[pid]
		if !exists {
			continue
		}

		memMB := float64(p.RSSKB) / 1024.0
		role := inferRole(p, paneTargets)
		state := inferState(p)

		info := ProcessInfo{
			PID:         p.PID,
			ParentPID:   p.PPID,
			Role:        role,
			State:       state,
			CPUPercent:  p.CPU,
			MemoryMB:    memMB,
			ElapsedTime: p.Elapsed,
			Command:     p.Command,
		}

		report.Processes = append(report.Processes, info)
		report.TotalCPU += p.CPU
		report.TotalMemory += memMB
	}

	return report, nil
}

func parsePSOutput(output string) map[int]rawProc {
	lines := strings.Split(output, "\n")
	res := make(map[int]rawProc)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, _ := strconv.Atoi(fields[1])
		cpuStr := strings.ReplaceAll(fields[2], ",", ".")
		cpu, _ := strconv.ParseFloat(cpuStr, 64)
		rss, _ := strconv.ParseInt(fields[3], 10, 64)
		elapsed := fields[4]
		cmd := strings.Join(fields[5:], " ")

		res[pid] = rawProc{
			PID:     pid,
			PPID:    ppid,
			CPU:     cpu,
			RSSKB:   rss,
			Elapsed: elapsed,
			Command: cmd,
		}
	}
	return res
}

func inferRole(p rawProc, paneTargets map[int]string) string {
	if r, ok := paneTargets[p.PID]; ok && r != "" {
		return r
	}
	cmdLower := strings.ToLower(p.Command)
	switch {
	case strings.Contains(cmdLower, "tmux new-session") || strings.Contains(cmdLower, "tmux attach"):
		return "tmux session (forge)"
	case strings.Contains(cmdLower, "worker --role builder") || strings.Contains(cmdLower, "forge-builder-ctl"):
		return "forge-builder-ctl (despachante)"
	case strings.Contains(cmdLower, "worker --role reviewer") || strings.Contains(cmdLower, "forge-reviewer-ctl"):
		return "forge-reviewer-ctl (despachante)"
	case strings.Contains(cmdLower, "go test"):
		return "└── subagente tests (go test)"
	case strings.Contains(cmdLower, "git"):
		return "└── git"
	case strings.Contains(cmdLower, "claude"):
		return "claude (agente)"
	case strings.Contains(cmdLower, "agy") || strings.Contains(cmdLower, "gemini"):
		return "gemini (agente)"
	case strings.Contains(cmdLower, "codex"):
		return "codex (agente)"
	default:
		if p.PPID != 1 && p.PPID != 0 {
			cmdBase := filepath.Base(strings.Fields(p.Command)[0])
			return fmt.Sprintf("└── subprocesso (%s)", cmdBase)
		}
		return "processo"
	}
}

func inferState(p rawProc) string {
	if p.CPU >= 1.0 {
		return "executando"
	}
	if p.CPU > 0.0 {
		return "ativo"
	}
	return "ocioso"
}

// FormatProcessTable renderiza a tabela de processos conforme o Protótipo 1.
func FormatProcessTable(r *ProjectProcessReport) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Processos Ativos — %s (Sessão tmux: %s)\n\n", r.Project, r.Session)

	fmt.Fprintf(&b, "%-6s %-30s %-11s %-7s %-11s %-13s %s\n",
		"PID", "PAPEL / COMPONENTE", "ESTADO", "CPU %", "MEM (RSS)", "TEMPO ATIVO", "COMANDO")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 88))

	if len(r.Processes) == 0 {
		fmt.Fprintf(&b, "Nenhum processo ativo encontrado para este projeto.\n")
	} else {
		for _, p := range r.Processes {
			memStr := fmt.Sprintf("%.0f MB", p.MemoryMB)
			cpuStr := fmt.Sprintf("%.1f%%", p.CPUPercent)
			cmdDisplay := p.Command
			if len(cmdDisplay) > 30 {
				cmdDisplay = cmdDisplay[:27] + "..."
			}
			fmt.Fprintf(&b, "%-6d %-30s %-11s %-7s %-11s %-13s %s\n",
				p.PID, p.Role, p.State, cpuStr, memStr, p.ElapsedTime, cmdDisplay)
		}
	}

	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 88))
	fmt.Fprintf(&b, "Total do Projeto: %d processos ativos │ CPU: %.1f%% │ Memória: %.0f MB\n\n",
		len(r.Processes), r.TotalCPU, r.TotalMemory)
	fmt.Fprintf(&b, "[Dica: Use 'forge-sdd ps --kill-zombies' para limpar processos orfãos ou presos]\n")

	return b.String()
}

// Signaler envia um sinal para o processo com determinado PID.
type Signaler func(pid int, sig syscall.Signal) error

// DefaultSignaler usa os.FindProcess e proc.Signal.
func DefaultSignaler(pid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}

// KillZombies encerra processos órfãos ou zumbis candidatos associados ao projeto
// (subprocessos com ParentPID == 1 que não sejam a própria sessão tmux).
func KillZombies(report *ProjectProcessReport) ([]ProcessInfo, error) {
	return KillZombiesWithSignaler(report, DefaultSignaler)
}

// KillZombiesWithSignaler permite mock do envio de sinais em testes.
func KillZombiesWithSignaler(report *ProjectProcessReport, signaler Signaler) ([]ProcessInfo, error) {
	if report == nil {
		return nil, nil
	}
	if signaler == nil {
		signaler = DefaultSignaler
	}
	var killed []ProcessInfo
	for _, p := range report.Processes {
		// Identifica processos candidatos a zumbi (ex: subprocessos com PPID 1 ou estado zumbi)
		if p.ParentPID == 1 && p.Role != "tmux session (forge)" {
			if err := signaler(p.PID, syscall.SIGTERM); err == nil {
				killed = append(killed, p)
			}
		}
	}
	return killed, nil
}
