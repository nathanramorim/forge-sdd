package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetWorkerFlags() {
	_ = pipelinePROpenedCmd.Flags().Set("worker", "")
	_ = pipelinePROpenedCmd.Flags().Set("branch", "")
	_ = workerCmd.Flags().Set("once", "false")
	_ = workerCmd.Flags().Set("name", "")
	_ = workerCmd.Flags().Set("cli", "")
	for _, f := range []string{"cli", "discovery", "builder", "reviewer", "terminal"} {
		_ = upCmd.Flags().Set(f, "")
	}
	_ = upCmd.Flags().Set("dry-run", "false")
}

func claimedJob(t *testing.T, dir string) {
	t.Helper()
	j := bus.Job{
		ID: "job-1", Type: bus.TypeFeatureBuild, Priority: bus.PriorityNormal, Feature: "auth-oauth", Title: "Auth", Initiative: "77-auth-oauth",
		SpecPath: "sdd/discovery/77-auth-oauth/criteria.md", Card: bus.Card{Provider: "clickup", ID: "CU-1", URL: "https://x/1"},
		PR: bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"}, MaxReworkAttempts: 3,
	}
	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, j, time.Now())
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleBuilder, "forge-builder", time.Minute, time.Now())
	require.NoError(t, err)
}

func TestPipelinePROpened_RegistraEValida(t *testing.T) {
	t.Setenv("FORGE_WORKER", "forge-builder")
	dir := t.TempDir()
	resetWorkerFlags()
	claimedJob(t, dir)

	out, err := runRoot(t, "pipeline", "pr-opened", "--dir", dir, "--id", "job-1", "--number", "12", "--url", "https://github.com/acme/app/pull/12")
	require.NoError(t, err)
	assert.Contains(t, out, "PR #12 registrado")
	job, state, err := bus.Get(dir, layout.RoleBuilder, "job-1")
	require.NoError(t, err)
	assert.Equal(t, layout.StateProcessing, state)
	require.NotNil(t, job.PR.URL)
	assert.Equal(t, "https://github.com/acme/app/pull/12", *job.PR.URL)

	for name, args := range map[string][]string{
		"url não http":     {"--number", "1", "--url", "javascript:alert(1)"},
		"url com controle": {"--number", "1", "--url", "https://x/\x1b[2J"},
		"número inválido":  {"--number", "0", "--url", "https://x/1"},
		"branch com opção": {"--number", "1", "--url", "https://x/1", "--branch", "--upload-pack=x"},
		"worker alheio":    {"--number", "1", "--url", "https://x/1", "--worker", "outro"},
		"url com espaço":   {"--number", "1", "--url", "https://x/a b"},
		"url com userinfo": {"--number", "1", "--url", "https://u:p@x/1"},
	} {
		resetWorkerFlags()
		_, err := runRoot(t, append([]string{"pipeline", "pr-opened", "--dir", dir, "--id", "job-1"}, args...)...)
		assert.Error(t, err, name)
	}
	_, err = runRoot(t, "pipeline", "pr-opened", "--dir", dir, "--id", "nao-existe", "--number", "1", "--url", "https://x/1")
	assert.Error(t, err)
}

func TestBusPlan_MostraOndasPorDependencia(t *testing.T) {
	dir := t.TempDir()
	cfg, _ := json.Marshal(config.DefaultForgeConfig())
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(cfg))
	base := "sdd/discovery/77-auth-oauth/features/"
	testutil.WriteFile(t, dir, base+"feat-77-01-a.md", "# a\n**Depende de:** —\n")
	testutil.WriteFile(t, dir, base+"feat-77-02-b.md", "# b\n**Depende de:** feat-77-01\n")
	testutil.WriteFile(t, dir, base+"feat-77-03-c.md", "# c\n**Depende de:** feat-77-01\n")
	claimedJob(t, dir)

	out, err := runRoot(t, "bus", "plan", "--dir", dir, "--role", "builder", "--id", "job-1")
	require.NoError(t, err)
	assert.Contains(t, out, "até 3 tarefas em paralelo")
	assert.Contains(t, out, "lote 1 (onda 0, sozinha): feat-77-01")
	assert.Contains(t, out, "em paralelo): feat-77-02, feat-77-03")

	_, err = runRoot(t, "bus", "plan", "--dir", dir, "--role", "builder", "--id", "nao-existe")
	require.Error(t, err)
	_, err = runRoot(t, "bus", "plan", "--dir", dir, "--role", "reviewer", "--id", "job-1")
	require.Error(t, err)
}

func TestWorkerCLI_OnceSemJobsSaiELimpo(t *testing.T) {
	dir := t.TempDir()
	resetWorkerFlags()
	_, err := runRoot(t, "worker", "--dir", dir, "--role", "builder", "--once")
	require.Error(t, err, "sem forge.json o worker não roda")
	assert.Contains(t, fmt.Sprint(err), "pipeline init")

	cfg, _ := json.Marshal(config.DefaultForgeConfig())
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(cfg))
	resetWorkerFlags()
	_, err = runRoot(t, "worker", "--dir", dir, "--role", "builder", "--once")
	require.NoError(t, err)

	resetWorkerFlags()
	_, err = runRoot(t, "worker", "--dir", dir, "--role", "discovery", "--once")
	require.Error(t, err)
}

func TestPipelinePROpened_SemWorkerERecusado(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := t.TempDir()
	resetWorkerFlags()
	claimedJob(t, dir)
	_, err := runRoot(t, "pipeline", "pr-opened", "--dir", dir, "--id", "job-1", "--number", "1", "--url", "https://x/1")
	require.Error(t, err, "sem worker nem FORGE_WORKER o dono do lease não é provado")
}
