package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pipelineDir(t *testing.T, mut func(*config.ForgeConfig)) string {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultForgeConfig()
	if mut != nil {
		mut(&cfg)
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), data, 0o644))
	return dir
}

func TestUp_FlagsPorSessao(t *testing.T) {
	dir := pipelineDir(t, nil)
	resetWorkerFlags()
	out, err := runRoot(t, "up", "--dir", dir, "--discovery", "claude", "--builder", "codex", "--reviewer", "gemini", "--dry-run", "--terminal", "native")
	require.NoError(t, err)
	assert.Contains(t, out, "claude -n forge-discovery --agent lider")
	assert.Contains(t, out, "--role builder --name forge-builder --cli codex")
	assert.Contains(t, out, "--role reviewer --name forge-reviewer --cli gemini")
	assert.Contains(t, out, "[DRY-RUN]")
}

func TestUp_Precedencia(t *testing.T) {
	// sessions de forge.json < flag da sessão; --cli vale para as sem escolha própria.
	dir := pipelineDir(t, func(c *config.ForgeConfig) {
		c.Sessions = map[string]string{config.SessionRoleBuilder: config.CLICopilot}
	})
	resetWorkerFlags()
	out, err := runRoot(t, "up", "--dir", dir, "--cli", "gemini", "--dry-run", "--terminal", "native")
	require.NoError(t, err)
	assert.Contains(t, out, "--name forge-builder --cli copilot", "sessions de forge.json vence --cli")
	assert.Contains(t, out, "--name forge-reviewer --cli gemini", "--cli vale para a sessão sem escolha")
	assert.Contains(t, out, "gemini -i", "a Discovery usa --cli")

	resetWorkerFlags()
	out, err = runRoot(t, "up", "--dir", dir, "--builder", "codex", "--dry-run", "--terminal", "native")
	require.NoError(t, err)
	assert.Contains(t, out, "--name forge-builder --cli codex", "a flag da sessão vence sessions")
}

func TestUp_ValoresInvalidosERecusados(t *testing.T) {
	dir := pipelineDir(t, nil)
	for _, args := range [][]string{
		{"--builder", "gpt"}, {"--cli", "cursor"}, {"--reviewer", "codex; rm -rf ~"}, {"--discovery", "../x"},
	} {
		resetWorkerFlags()
		_, err := runRoot(t, append([]string{"up", "--dir", dir, "--dry-run"}, args...)...)
		assert.Error(t, err, args)
	}
}

func TestUp_NaoGravaForgeJSON(t *testing.T) {
	dir := pipelineDir(t, nil)
	before, _ := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	resetWorkerFlags()
	_, err := runRoot(t, "up", "--dir", dir, "--builder", "codex", "--dry-run", "--terminal", "native")
	require.NoError(t, err)
	after, _ := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	assert.Equal(t, string(before), string(after), "as flags não alteram forge.json")
}

func TestUp_Terminal(t *testing.T) {
	dir := pipelineDir(t, nil)
	resetWorkerFlags()
	out, err := runRoot(t, "up", "--dir", dir, "--terminal", "native", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "forge-discovery")
	assert.NotContains(t, out, "tmux new-session")

	resetWorkerFlags()
	_, err = runRoot(t, "up", "--dir", dir, "--terminal", "janelas", "--dry-run")
	assert.Error(t, err, "valor fora da lista é recusado")

	if _, lookErr := exec.LookPath("tmux"); lookErr == nil {
		resetWorkerFlags()
		out, err = runRoot(t, "up", "--dir", dir, "--terminal", "tmux", "--dry-run")
		require.NoError(t, err)
		assert.Contains(t, out, "tmux new-session -d -s forge-")
		assert.Contains(t, out, "split-window -h")
		assert.Contains(t, out, "-T forge-reviewer")
		assert.Contains(t, out, "[DRY-RUN] Nada foi aberto (sessão tmux forge-")
	}
}
