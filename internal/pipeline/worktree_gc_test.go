package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreeGC_ListDirAusente verifica que a ausência de .forge/runtime/worktrees retorna
// slice vazio sem erro.
func TestWorktreeGC_ListDirAusente(t *testing.T) {
	dir := t.TempDir()

	result, err := ListStaleWorktrees(dir)

	require.NoError(t, err)
	assert.Empty(t, result)
}

// TestWorktreeGC_SweepSemTargets verifica que SweepWorktrees com slice vazio não retorna erros.
func TestWorktreeGC_SweepSemTargets(t *testing.T) {
	dir := t.TempDir()

	errs := SweepWorktrees(dir, []StaleWorktree{})

	assert.Empty(t, errs)
}

// TestWorktreeGC_OrphanSafe verifica que worktrees sem job correspondente no barramento são
// classificados como seguros para remoção.
func TestWorktreeGC_OrphanSafe(t *testing.T) {
	dir := t.TempDir()
	wtRoot := filepath.Join(dir, ".forge", "runtime", "worktrees")
	jobID := "job-orphan-01"
	taskID := "task-01"
	require.NoError(t, os.MkdirAll(filepath.Join(wtRoot, jobID, taskID), 0o755))

	result, err := ListStaleWorktrees(dir)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.True(t, result[0].Safe, "worktree orphan deve ser marcado como seguro")
	assert.Equal(t, jobID, result[0].JobID)
	assert.Equal(t, taskID, result[0].TaskID)
}

// TestWorktreeGC_SweepNaoRemoveProt verifica que SweepWorktrees não toca worktrees com Safe=false.
func TestWorktreeGC_SweepNaoRemoveProt(t *testing.T) {
	dir := t.TempDir()
	targets := []StaleWorktree{
		{JobID: "job-01", TaskID: "task-01", Safe: false, Reason: "job em execução"},
	}

	errs := SweepWorktrees(dir, targets)

	assert.Empty(t, errs)
}

// TestWorktreeGC_JobEmExecucao verifica que Safe=false quando o job está em StateProcessing.
func TestWorktreeGC_JobEmExecucao(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)

	// Monta worktree falso.
	jobID := "gc-job-01"
	taskID := "97-02"
	wtRoot := filepath.Join(dir, ".forge", "runtime", "worktrees")
	require.NoError(t, os.MkdirAll(filepath.Join(wtRoot, jobID, taskID), 0o755))

	// Cria e reivindica o job no barramento (coloca em processing).
	j := bus.Job{
		ID:                jobID,
		Type:              bus.TypeFeatureBuild,
		Priority:          bus.PriorityNormal,
		Feature:           "97-gc-test",
		Title:             "GC test job processing",
		SpecPath:          "sdd/discovery/97-gc-test/features/feat-97-02-gc.md",
		MaxReworkAttempts: 3,
		PR:                bus.PR{Branch: "feat/97-gc-test", BaseBranch: "main"},
	}
	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, t0)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleBuilder, "test-worker", time.Hour, t0)
	require.NoError(t, err)

	result, err := ListStaleWorktrees(dir)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.False(t, result[0].Safe, "job em execução deve ser protegido")
	assert.Equal(t, layout.StateProcessing, result[0].JobState)
	assert.Equal(t, "job em execução", result[0].Reason)
}

// TestWorktreeGC_TaskAprovada verifica que Safe=true quando o job está em StateDone e a task
// tem StatusApproved na fila da iniciativa. Exercita também NormalizeTaskID via gcMatchesTask
// (item "feat-97-02-gc" casa com worktree "97-02").
func TestWorktreeGC_TaskAprovada(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)

	// Monta worktree falso.
	jobID := "gc-job-02"
	taskID := "97-02"
	initiative := "97-gc-test"
	wtRoot := filepath.Join(dir, ".forge", "runtime", "worktrees")
	require.NoError(t, os.MkdirAll(filepath.Join(wtRoot, jobID, taskID), 0o755))

	// Cria job no barramento, reivindica e conclui (coloca em done).
	j := bus.Job{
		ID:                jobID,
		Type:              bus.TypeFeatureBuild,
		Priority:          bus.PriorityNormal,
		Feature:           "97-gc-test",
		Title:             "GC test job done",
		SpecPath:          "sdd/discovery/97-gc-test/features/feat-97-02-gc.md",
		Initiative:        initiative,
		MaxReworkAttempts: 3,
		PR:                bus.PR{Branch: "feat/97-gc-test", BaseBranch: "main"},
	}
	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, t0)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleBuilder, "test-worker", time.Hour, t0)
	require.NoError(t, err)
	_, err = bus.Complete(dir, layout.RoleBuilder, jobID, "test-worker", "done", t0)
	require.NoError(t, err)

	// Cria queue.json da iniciativa com a task aprovada.
	// item "feat-97-02-gc" casa com worktree taskID "97-02" via gcMatchesTask/NormalizeTaskID.
	queueData := queue.Queue{
		SpecStatus: "reviewed",
		Items: []queue.QueueItem{
			{
				TaskID:   "feat-97-02-gc",
				Title:    "GC test task",
				SpecPath: "sdd/discovery/97-gc-test/features/feat-97-02-gc.md",
				Status:   queue.StatusApproved,
			},
		},
	}
	queuePath := filepath.Join(dir, "sdd", "discovery", initiative, "queue.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(queuePath), 0o755))
	raw, err := json.Marshal(queueData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(queuePath, raw, 0o644))

	result, err := ListStaleWorktrees(dir)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.True(t, result[0].Safe, "job concluído e task aprovada deve ser seguro")
	assert.Equal(t, queue.StatusApproved, result[0].TaskStatus)
}
