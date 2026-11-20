package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateClaudeIntegracaoSemMudarVersao(t *testing.T) {
	dir := modelProject(t, nil)
	mcp := filepath.Join(dir, ".mcp.json")
	settings := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.WriteFile(mcp, []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(settings, []byte(`{"permissions":{"allow":["Bash(ls)"]}}`), 0o644))

	// dry-run não grava
	out, err := updateRun(t, dir, true)
	require.NoError(t, err)
	assert.Contains(t, out, "MCP forge-sdd seria instalado")
	assert.Contains(t, out, "hook SessionStart seria instalado")
	got, _ := os.ReadFile(mcp)
	assert.Equal(t, `{}`, string(got))

	out, err = updateRun(t, dir, false)
	require.NoError(t, err)
	assert.Contains(t, out, "MCP forge-sdd instalado")
	assert.Contains(t, out, "hook SessionStart instalado")
	data, _ := os.ReadFile(mcp)
	assert.Contains(t, string(data), "forge-sdd")
	s, _ := os.ReadFile(settings)
	assert.Contains(t, string(s), "forge doctor --harness")
	assert.Contains(t, string(s), "Bash(ls)")

	// negativo: segunda rodada não muda nada
	out, err = updateRun(t, dir, false)
	require.NoError(t, err)
	assert.NotContains(t, out, "instalado")
	after, _ := os.ReadFile(mcp)
	assert.Equal(t, string(data), string(after))
}

func TestUpdateClaudeIntegracaoMCPInvalidoIntacto(t *testing.T) {
	dir := modelProject(t, nil)
	mcp := filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.WriteFile(mcp, []byte(`{"mcpServers":`), 0o644))
	out, err := updateRun(t, dir, false)
	require.NoError(t, err)
	assert.NotContains(t, out, "MCP forge-sdd instalado")
	st, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Contains(t, string(st), "forge doctor --harness") // o resto segue
	got, _ := os.ReadFile(mcp)
	assert.Equal(t, `{"mcpServers":`, string(got))
}
