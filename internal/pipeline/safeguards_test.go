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
	assert.Contains(t, JobOverCeiling(dir, cfg, layout.RoleBuilder, "job-2", 2*time.Hour), "minutos", "teto de tempo")
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
	assert.Zero(t, config.DefaultForgeConfig().Safeguards.MaxJobsPerDay)
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

// liveStopProject monta um projeto live no modo avançado com teto diário de 1 job já gasto.
func liveStopProject(t *testing.T) (string, config.ForgeConfig) {
	t.Helper()
	dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	cfg.Permissions = config.PermissionFull
	cfg.Autonomy = config.AutonomyAdvanced
	cfg.IsolatedConfirmed = true
	cfg.Safeguards.MaxJobsPerDay = 1
	require.NoError(t, config.SaveForgeConfig(dir, cfg))
	seedJob(t, dir, layout.RoleBuilder, "job-gasto", "feature_build", "normal", now)
	runOnce(t, dir, layout.RoleBuilder, &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}, nil)
	reason, _ := DailyPause(dir, cfg, now)
	require.NotEmpty(t, reason)
	return dir, cfg
}

func TestDeliver_StopEncerraSemGirar(t *testing.T) {
	dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
	require.NoError(t, os.MkdirAll(filepath.Dir(StopFile(dir)), 0o755))
	require.NoError(t, os.WriteFile(StopFile(dir), nil, 0o644))
	var out bytes.Buffer
	err := RunDeliverWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder,
		DeliverTarget: "=t:f.1", Deliverer: &fakeTmuxDeliverer{}, Out: &out, Now: func() time.Time { return now },
		Sleep: func(context.Context, time.Duration) bool { t.Fatal("não deve dormir"); return false }})
	require.NoError(t, err)
	assert.Contains(t, out.String(), "esteira parada")
}

func TestDeliver_PausaEntaoStopEncerra(t *testing.T) {
	dir, _ := liveStopProject(t)
	sleeps := 0
	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- RunDeliverWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder,
			DeliverTarget: "=t:f.1", Deliverer: &fakeTmuxDeliverer{}, Out: &out, Now: func() time.Time { return now },
			Sleep: func(context.Context, time.Duration) bool {
				sleeps++
				if sleeps == 2 {
					_ = os.MkdirAll(filepath.Dir(StopFile(dir)), 0o755)
					_ = os.WriteFile(StopFile(dir), nil, 0o644)
				}
				return sleeps < 50
			}})
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("despachante live não encerrou com o STOP (busy loop)")
	}
	assert.LessOrEqual(t, sleeps, 3)
	assert.Contains(t, out.String(), "esteira em pausa")
	assert.Contains(t, out.String(), "esteira parada")
}
