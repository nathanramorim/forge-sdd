package queue_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordAndLoadActivity(t *testing.T) {
	tempDir := t.TempDir()

	err := queue.RecordActivity(tempDir, queue.StationLider, "run_command", "git status", "Verificando status", "success")
	require.NoError(t, err)

	err = queue.RecordActivity(tempDir, queue.StationBuilder, "edit_file", "server.go", "Ajustando porta", "running")
	require.NoError(t, err)

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, queue.StationBuilder, items[0].Station)
	assert.Equal(t, "edit_file", items[0].Tool)
	assert.Equal(t, "running", items[0].Status)

	assert.Equal(t, queue.StationLider, items[1].Station)
	assert.Equal(t, "run_command", items[1].Tool)
	assert.Equal(t, "success", items[1].Status)
}

func TestRecordActivityDetailed_ReconcilesPreviousRunning(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Primeira atividade running no builder
	err := queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationBuilder,
		Tool:    "run_command",
		Action:  "go test -race ./...",
		Details: "Executando testes locais",
		Status:  "running",
		Command: "go test -race ./...",
	})
	require.NoError(t, err)

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "running", items[0].Status)

	// 2. Segunda atividade running no mesmo builder -> primeira deve reconciliar para 'success'
	err = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationBuilder,
		Tool:    "edit_file",
		Action:  "pool.go",
		Details: "Editando lógica de pool",
		Status:  "running",
	})
	require.NoError(t, err)

	items, err = queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "running", items[0].Status)
	assert.Equal(t, "pool.go", items[0].Action)

	assert.Equal(t, "success", items[1].Status)
	assert.Equal(t, "go test -race ./...", items[1].Action)
	assert.NotEmpty(t, items[1].CompletedAt)
}

func TestReconcileStationActivities(t *testing.T) {
	tempDir := t.TempDir()

	_ = queue.RecordActivity(tempDir, queue.StationBuilder, "tool1", "act1", "det1", "running")
	_ = queue.RecordActivity(tempDir, queue.StationPMArchitect, "tool2", "act2", "det2", "running")

	// Reconcilia apenas a estação builder
	err := queue.ReconcileStationActivities(tempDir, queue.StationBuilder)
	require.NoError(t, err)

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 2)

	for _, it := range items {
		if it.Station == queue.StationBuilder {
			assert.Equal(t, "success", it.Status)
		} else if it.Station == queue.StationPMArchitect {
			assert.Equal(t, "running", it.Status)
		}
	}
}

func TestSanitizeOrphanActivities(t *testing.T) {
	tempDir := t.TempDir()

	_ = queue.RecordActivity(tempDir, queue.StationBuilder, "tool1", "act1", "det1", "running")
	_ = queue.RecordActivity(tempDir, queue.StationPMArchitect, "tool2", "act2", "det2", "running")
	_ = queue.RecordActivity(tempDir, queue.StationGovernance, "tool3", "act3", "det3", "running")

	err := queue.SanitizeOrphanActivities(tempDir)
	require.NoError(t, err)

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 3)

	for _, it := range items {
		assert.Equal(t, "success", it.Status)
		assert.NotEmpty(t, it.CompletedAt)
	}
}

func TestOutputTruncation(t *testing.T) {
	tempDir := t.TempDir()

	hugeOutput := strings.Repeat("a", 16000)
	err := queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationBuilder,
		Tool:    "run_command",
		Action:  "large output",
		Details: "Teste de truncamento",
		Status:  "success",
		Output:  hugeOutput,
	})
	require.NoError(t, err)

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 1)

	assert.True(t, len(items[0].Output) <= 8192+len("\n... [output truncado em 8KB]"))
	assert.Contains(t, items[0].Output, "[output truncado em 8KB]")
}

func TestConcurrentRecordActivity(t *testing.T) {
	tempDir := t.TempDir()

	var wg sync.WaitGroup
	workers := 10
	iterations := 5

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = queue.RecordActivity(tempDir, queue.StationBuilder, "tool", "action", "detail", "running")
			}
		}(w)
	}

	wg.Wait()

	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	assert.True(t, len(items) > 0)
	assert.True(t, len(items) <= 50)
}
