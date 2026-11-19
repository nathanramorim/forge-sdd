package pipeline

import (
	"os"
	"path/filepath"
	"testing"

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
