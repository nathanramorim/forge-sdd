package pipeline

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sessionProject(t *testing.T, sessions map[string]string, level string) string {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = sessions
	cfg.Permissions = level
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(data))
	return dir
}

func TestWorker_PerfilDaSessao(t *testing.T) {
	dir := sessionProject(t, map[string]string{config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}, config.PermissionProject)
	seedJob(t, dir, layout.RoleBuilder, "job-b", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	seedJob(t, dir, layout.RoleReviewer, "job-r", bus.TypePRReview, bus.PriorityNormal, now)

	for role, want := range map[string]string{
		layout.RoleBuilder:  config.CLICodex,
		layout.RoleReviewer: config.CLIGemini,
	} {
		runner := &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}
		var out strings.Builder
		require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: role, Once: true, Out: &out, Runner: runner, Now: func() time.Time { return now }}))
		assert.Contains(t, out.String(), "agente: "+want, role)
		require.Len(t, runner.specs, 1, role)
		assert.Equal(t, config.PermissionArgs(want, config.PermissionProject, config.DefaultForgeConfig().AllowCommands), runner.specs[0].Permissions, role)
	}
}

func TestWorker_FlagCLISobrescreveASessao(t *testing.T) {
	dir := sessionProject(t, map[string]string{config.SessionRoleBuilder: config.CLICodex}, config.PermissionFull)
	seedJob(t, dir, layout.RoleBuilder, "job-b", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	runner := &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, nil }}
	var out strings.Builder
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, CLI: config.CLICopilot, Once: true, Out: &out, Runner: runner, Now: func() time.Time { return now }}))
	assert.Contains(t, out.String(), "agente: copilot")
	assert.Contains(t, out.String(), "TOTAL", "o nível full é avisado ao iniciar")
	assert.Equal(t, []string{"--allow-all-tools"}, runner.specs[0].Permissions)
}

func TestWorker_AgenteInvalidoNaoInicia(t *testing.T) {
	dir := sessionProject(t, nil, config.PermissionProject)
	err := RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, CLI: "gpt", Once: true})
	assert.Error(t, err)
}
