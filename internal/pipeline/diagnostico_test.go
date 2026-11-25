package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Conforme: em um bloco só aparecem o worker sem registro, a última falha com motivo e log e o
// pendente velho que o worker pegaria antes dos novos (o caso que levou 6 comandos para achar).
func TestDiagnostico_UmBlocoComFalhaEPendenteVelho(t *testing.T) {
	dir := t.TempDir()
	velho := time.Date(2026, 11, 17, 11, 5, 0, 0, time.UTC)
	now := time.Date(2026, 11, 25, 3, 0, 0, 0, time.UTC)
	j := jobDeTeste("job-20261117-110539-d7e6-a", "d7e6-a", bus.TypeFeatureBuild, "feat/d7e6-a")
	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, velho)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleBuilder, "w", time.Hour, now)
	require.NoError(t, err)
	_, err = bus.Fail(dir, layout.RoleBuilder, j.ID, "w", "o agente saiu com código -1", now)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(layout.Logs(dir), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(layout.Logs(dir), "builder-"+j.ID+".log"), []byte("x"), 0o600))

	p := jobDeTeste("job-20261116-000000-55e0-b", "55e0-b", bus.TypeFeatureBuild, "feat/55e0-b")
	_, _, err = bus.Enqueue(dir, layout.RoleBuilder, p, velho)
	require.NoError(t, err)

	out := FormatDiagnostico(dir, now)
	assert.Contains(t, out, "builder sem registro (rode forge up)")
	assert.Contains(t, out, "fila builder: 1 pendente(s) · 0 em execução")
	assert.Contains(t, out, "última falha: job-20261117-110539-d7e6-a · feature_build")
	assert.Contains(t, out, "o agente saiu com código -1")
	assert.Contains(t, out, "log: .forge/runtime/logs/builder-job-20261117-110539-d7e6-a.log")
	assert.Contains(t, out, "1 pendente(s) com mais de 24h")
}

// Negativo: projeto sem runtime não imprime nada (o status segue igual).
func TestDiagnostico_SemRuntimeVazio(t *testing.T) {
	assert.Empty(t, FormatDiagnostico(t.TempDir(), time.Now()))
}
