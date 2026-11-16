package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func advancedCfg() config.ForgeConfig {
	cfg := config.DefaultForgeConfig()
	cfg.Permissions = config.PermissionFull
	cfg.Autonomy = config.AutonomyAdvanced
	cfg.IsolatedConfirmed = true
	return cfg
}

func TestSalvaguardas_TetoPorJob(t *testing.T) {
	dir := t.TempDir()
	cfg := advancedCfg()
	cfg.Safeguards.MaxJobTokens = 100
	logDir := layout.Logs(dir)
	require.NoError(t, os.MkdirAll(logDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "builder-job-1.log"), make([]byte, 4000), 0o600))

	assert.Contains(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-1", time.Minute), "teto de 100 tokens")
	assert.Contains(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-2", 2*time.Hour), "minutos", "teto de tempo")
	assert.Empty(t, JobOverCeiling(dir, config.DefaultForgeConfig(), layout.RoleBuilder, "job-1", 2*time.Hour), "modo normal sem teto")
}

func TestSalvaguardas_TetoDiario(t *testing.T) {
	dir := workerProject(t)
	cfg := advancedCfg()
	cfg.Safeguards.MaxJobsPerDay = 1
	assert.Empty(t, HaltReason(dir, cfg, now))

	seedJob(t, dir, layout.RoleBuilder, "job-a", "feature_build", "normal", now)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	assert.Contains(t, HaltReason(dir, cfg, now), "teto diário")
	assert.Empty(t, HaltReason(dir, config.DefaultForgeConfig(), now), "modo normal sem teto diário")
}

func TestSalvaguardas_Stop(t *testing.T) {
	dir := workerProject(t)
	cfg := config.DefaultForgeConfig()
	assert.Empty(t, HaltReason(dir, cfg, now))
	require.NoError(t, os.MkdirAll(filepath.Dir(StopFile(dir)), 0o755))
	require.NoError(t, os.WriteFile(StopFile(dir), nil, 0o644))
	assert.Contains(t, HaltReason(dir, cfg, now), "STOP")

	seedJob(t, dir, layout.RoleBuilder, "job-a", "feature_build", "normal", now)
	out := runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	assert.Contains(t, out, "esteira parada")
}
