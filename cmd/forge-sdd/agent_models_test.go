package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func modelProject(t *testing.T, models map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	_, err := scaffold.Run(config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: version, Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)
	cfg := config.DefaultForgeConfig()
	cfg.AgentModels = models
	require.NoError(t, config.SaveForgeConfig(dir, cfg))
	return dir
}

func updateRun(t *testing.T, dir string, dry bool) (string, error) {
	t.Helper()
	_ = updateCmd.Flags().Set("yes", "true")
	_ = updateCmd.Flags().Set("dry-run", map[bool]string{true: "true", false: "false"}[dry])
	t.Cleanup(func() {
		_ = updateCmd.Flags().Set("yes", "false")
		_ = updateCmd.Flags().Set("dry-run", "false")
		rootCmd.SetOut(nil)
	})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"update", dir})
	err := rootCmd.Execute()
	return buf.String(), err
}

func agentHeader(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "agents", name+".md"))
	require.NoError(t, err)
	for _, l := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(l, "model:") {
			return strings.TrimSpace(strings.TrimPrefix(l, "model:"))
		}
	}
	return ""
}

func TestUpdateAgentModels_AplicaSemAtualizacaoDeVersao(t *testing.T) {
	dir := modelProject(t, map[string]string{"builder": "sonnet", "reviewer": "opus"})
	out, err := updateRun(t, dir, false)
	require.NoError(t, err)
	assert.Contains(t, out, "model: atualizado")
	assert.Equal(t, "sonnet", agentHeader(t, dir, "builder"))
	assert.Equal(t, "opus", agentHeader(t, dir, "reviewer"))
	assert.Equal(t, "inherit", agentHeader(t, dir, "lider"))

	out, err = updateRun(t, dir, false) // idempotente
	require.NoError(t, err)
	assert.NotContains(t, out, "model:")
	assert.Equal(t, "sonnet", agentHeader(t, dir, "builder"))
}

func TestUpdateAgentModels_DryRunNaoGrava(t *testing.T) {
	dir := modelProject(t, map[string]string{"builder": "sonnet"})
	out, err := updateRun(t, dir, true)
	require.NoError(t, err)
	assert.Contains(t, out, "seria atualizado")
	assert.Equal(t, "inherit", agentHeader(t, dir, "builder"))
}

func TestUpdateAgentModels_ValorInvalidoErraSemGravar(t *testing.T) {
	dir := modelProject(t, map[string]string{"builder": "sonnet"})
	raw := `{"cli":"claude","agent_models":{"builder":"sonnet","reviewer":"ruim\nmodelo"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), []byte(raw), 0o644))
	_, err := updateRun(t, dir, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_models")
	assert.Equal(t, "inherit", agentHeader(t, dir, "builder"), "nada é gravado com valor inválido")
}
