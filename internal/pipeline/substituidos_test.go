package pipeline

import (
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jobDeTeste(id, ini, tipo, branch string) bus.Job {
	return bus.Job{
		SchemaVersion: bus.SchemaVersion, ID: id, Type: tipo, Priority: bus.PriorityNormal,
		Feature: ini, Title: "t", Initiative: ini, TaskIDs: []string{"x-01"}, SpecPath: "s.md",
		PR: bus.PR{Branch: branch, BaseBranch: "main"}, Attempt: 1, MaxReworkAttempts: 3, ReworkFeedback: []bus.Feedback{},
	}
}

// Conforme: o despacho novo tira da fila o feature_build pendente antigo da mesma iniciativa (o que o
// worker pegaria antes das tasks) e preserva fixes, outras iniciativas e os jobs do próprio despacho.
func TestDescartarJobsSubstituidos(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 25, 3, 0, 0, 0, time.UTC)
	ini := "d7e6-auditoria-linha-premium"
	for _, j := range []bus.Job{
		jobDeTeste("job-20261117-110539-d7e6-auditoria-linha-premium", ini, bus.TypeFeatureBuild, "feat/"+ini),
		jobDeTeste("job-20261125-031931-d7e6-auditoria-linha-premium", ini, bus.TypeFeatureBuild, "feat/"+ini),
		jobDeTeste("job-20261120-000000-d7e6-fix-01", ini, bus.TypeFeatureBuild, "fix/d7e6-01"),
		jobDeTeste("job-20261117-000000-55e0-outra", "55e0-outra", bus.TypeFeatureBuild, "feat/55e0-outra"),
	} {
		_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, now)
		require.NoError(t, err)
	}
	novo := "job-20261125-031931-d7e6-auditoria-linha-premium"
	fora, err := DescartarJobsSubstituidos(dir, ini, map[string]bool{novo: true}, novo, now)
	require.NoError(t, err)
	assert.Equal(t, []string{"job-20261117-110539-d7e6-auditoria-linha-premium"}, fora)

	pend, err := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, err)
	assert.Len(t, pend, 3)
	falhos, err := bus.List(dir, layout.RoleBuilder, layout.StateFailed)
	require.NoError(t, err)
	require.Len(t, falhos, 1)
}

// Negativo: job em execução nunca é descartado.
func TestDescartar_SoPendente(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 25, 3, 0, 0, 0, time.UTC)
	j := jobDeTeste("job-20261117-110539-d7e6-a", "d7e6-a", bus.TypeFeatureBuild, "feat/d7e6-a")
	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, now)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleBuilder, "w", time.Minute, now)
	require.NoError(t, err)
	_, err = bus.Descartar(dir, layout.RoleBuilder, j.ID, "dispatch", "x", now)
	assert.Error(t, err)
}
