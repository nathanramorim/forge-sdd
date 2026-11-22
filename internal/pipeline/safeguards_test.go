package pipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
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
	assert.Empty(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-2", 2*time.Hour), "sem max_job_minutes não há teto de tempo")
	cfg.Safeguards.MaxJobMinutes = 30
	assert.Contains(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-2", 2*time.Hour), "30 minutos", "teto de tempo")
	assert.Empty(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-2", 10*time.Minute), "dentro do teto")
	assert.Empty(t, JobOverCeiling(dir, config.DefaultForgeConfig(), layout.RoleBuilder, "job-1", 2*time.Hour), "modo normal sem teto")
}

func TestSalvaguardas_TetoDiario(t *testing.T) {
	dir := workerProject(t)
	cfg := advancedCfg()
	cfg.Safeguards.MaxJobsPerDay = 1
	reason, _ := DailyPause(dir, cfg, now)
	assert.Empty(t, reason)

	seedJob(t, dir, layout.RoleBuilder, "job-a", "feature_build", "normal", now)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	reason, until := DailyPause(dir, cfg, now)
	assert.Contains(t, reason, "teto diário")
	assert.True(t, until.After(now))
	assert.Equal(t, 0, until.Hour())
	assert.Empty(t, HaltReason(dir, cfg, now), "o teto pausa, não encerra")
	reason, _ = DailyPause(dir, config.DefaultForgeConfig(), now)
	assert.Empty(t, reason, "modo normal sem teto diário")
}

func TestSalvaguardas_ContaJobsDistintos(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-a", "feature_build", "normal", now)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	assert.Equal(t, 1, jobsStartedToday(dir, now))
}

func TestDefaultSafeguards_SemTetoDiario(t *testing.T) {
	assert.Equal(t, config.SafeguardsConfig{}, config.DefaultForgeConfig().Safeguards, "nenhum teto ligado por padrão")
}

func TestJobOverCeiling_PadraoNaoParaJobLongo(t *testing.T) {
	dir := t.TempDir()
	cfg := advancedCfg()
	logDir := layout.Logs(dir)
	require.NoError(t, os.MkdirAll(logDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "builder-job-1.log"), make([]byte, 40_000_000), 0o600))
	assert.Empty(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-1", 10*time.Hour))
}

func TestRun_PausaNoTeto(t *testing.T) {
	dir := workerProject(t)
	cfg := advancedCfg()
	cfg.Safeguards.MaxJobsPerDay = 1
	reason, _ := DailyPause(dir, cfg, now)
	require.Empty(t, reason)
	seedJob(t, dir, layout.RoleBuilder, "job-a", "feature_build", "normal", now)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	reason, until := DailyPause(dir, cfg, now)
	require.NotEmpty(t, reason)

	// no dia seguinte a pausa termina sozinha
	assert.Empty(t, firstReason(DailyPause(dir, cfg, until.Add(time.Minute))))

	// a espera não encerra o worker: só sai pelo contexto
	ctx, cancel := context.WithCancel(context.Background())
	sleeps := 0
	var out bytes.Buffer
	opts := WorkerOptions{Dir: dir, Out: &out, Now: func() time.Time { return now }, Poll: time.Second,
		Sleep: func(context.Context, time.Duration) bool {
			sleeps++
			if sleeps == 3 {
				cancel()
				return false
			}
			return true
		}}
	assert.False(t, waitDailyPause(ctx, opts, cfg, "w"))
	assert.Equal(t, 3, sleeps)
	assert.Equal(t, 1, strings.Count(out.String(), "esteira em pausa"), "avisa uma vez")
}

func firstReason(r string, _ time.Time) string { return r }

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

