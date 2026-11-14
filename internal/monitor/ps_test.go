package monitor

import (
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectProcessInspection(t *testing.T) {
	fakePS := `  PID  PPID  %CPU    RSS     ELAPSED COMMAND
48120     1   0.2  18432     2:15:00 tmux new-session -d -s forge-barbearia-app-a1b2c3 -n forge
48125 48120   0.0  83968     2:15:00 agy -i
48130 48120  14.2 148480       18:32 claude -n forge-builder
48201 48130  45.0 112640       00:42 go test -v ./...
48135 48120   0.0  79872     2:15:00 agy -i
48142 48120   0.1  14336     2:15:00 forge-sdd worker --role builder
99999     1   1.0  50000       10:00 unrelated_system_process
`

	fakeTmux := `48125|forge-discovery|forge
48130|forge-builder|forge
48135|forge-reviewer|forge
48142|forge-builder-ctl|forge-ctl
`

	runner := func(name string, args ...string) ([]byte, error) {
		if name == "ps" {
			return []byte(fakePS), nil
		}
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return []byte(fakeTmux), nil
		}
		return nil, nil
	}

	report, err := InspectProjectProcesses("/workspace/barbearia-app", runner)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "barbearia-app", report.Project)
	require.Len(t, report.Processes, 5) // 4 painéis + 1 filho do builder (48201)

	// Verifica o processo filho do builder (go test)
	var foundGoTest bool
	for _, p := range report.Processes {
		if p.PID == 48201 {
			foundGoTest = true
			assert.Equal(t, 48130, p.ParentPID)
			assert.Equal(t, "└── subagente tests (go test)", p.Role)
			assert.Equal(t, "executando", p.State)
			assert.Equal(t, 45.0, p.CPUPercent)
			assert.InDelta(t, 110.0, p.MemoryMB, 5.0)
		}
	}
	assert.True(t, foundGoTest, "deve ter identificado o subprocesso filho go test")

	// Confere totais
	assert.Greater(t, report.TotalCPU, 50.0)
	assert.Greater(t, report.TotalMemory, 400.0)

	// Confere layout da tabela formatada
	table := FormatProcessTable(report)
	assert.Contains(t, table, "Processos Ativos — barbearia-app")
	assert.Contains(t, table, "48125")
	assert.Contains(t, table, "48130")
	assert.Contains(t, table, "48201")
	assert.Contains(t, table, "Total do Projeto:")
	assert.Contains(t, table, "--kill-zombies")
}

func TestProjectProcessInspection_SemTmuxAtivo(t *testing.T) {
	fakePS := `  PID  PPID  %CPU    RSS     ELAPSED COMMAND
12345     1   0.5  20480       05:00 /usr/local/bin/forge-sdd worker --role builder --dir /app/my-project
`
	runner := func(name string, args ...string) ([]byte, error) {
		if name == "ps" {
			return []byte(fakePS), nil
		}
		return nil, assert.AnError
	}

	report, err := InspectProjectProcesses("/app/my-project", runner)
	require.NoError(t, err)
	require.Len(t, report.Processes, 1)
	assert.Equal(t, 12345, report.Processes[0].PID)
	assert.True(t, strings.Contains(report.Processes[0].Role, "builder"))
}

func TestKillZombies(t *testing.T) {
	report := &ProjectProcessReport{
		Project: "test-app",
		Session: "forge-test",
		Processes: []ProcessInfo{
			{PID: 100, ParentPID: 1, Role: "tmux session (forge)"}, // sessão raiz tmux não deve ser morta
			{PID: 101, ParentPID: 1, Role: "processo órfão"},        // órfão deve ser morto
			{PID: 102, ParentPID: 100, Role: "filho do tmux"},      // não órfão
			{PID: 103, ParentPID: 1, Role: "outro órfão"},          // outro órfão
		},
	}

	var signalsSent []int
	signaler := func(pid int, sig syscall.Signal) error {
		assert.Equal(t, syscall.SIGTERM, sig)
		signalsSent = append(signalsSent, pid)
		return nil
	}

	killed, err := KillZombiesWithSignaler(report, signaler)
	require.NoError(t, err)
	require.Len(t, killed, 2)
	assert.Equal(t, []int{101, 103}, signalsSent)
	assert.Equal(t, 101, killed[0].PID)
	assert.Equal(t, 103, killed[1].PID)
}
