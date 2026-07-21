package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffoldIntegration_Gemini(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project:    "integration-test",
		Stack:      "go",
		DB:         "none",
		Telemetry:  true,
		Lang:       "pt-BR",
		SddVersion: "1.9.1-beta",
		Agents:     []string{config.AgentGemini},
	}

	created, err := Run(cfg, dir)
	require.NoError(t, err)
	require.NotEmpty(t, created)

	// Árvore SDD básica
	assert.FileExists(t, filepath.Join(dir, "sdd", ".sdd-version"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "memory", "progress.md"))
	assert.DirExists(t, filepath.Join(dir, "sdd", "discovery"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "releases", "history.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", ".metrics", "schema.json"))

	// Configuração Gemini
	assert.FileExists(t, filepath.Join(dir, "GEMINI.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "system_instructions.md"))

	// Skills (Chatmodes)
	skills := []string{"orquestrador", "builder", "revisor", "archivist", "specifier", "migrator", "c4-architecture"}
	for _, s := range skills {
		assert.FileExists(t, filepath.Join(dir, ".gemini", "skills", s+".chatmode.md"))
	}

	// Prompts
	prompts := []string{"status", "proxima-feature", "nova-feature", "revisar", "archive", "doctor", "upgrade-sdd", "discovery", "constitution", "c4-architecture"}
	for _, p := range prompts {
		assert.FileExists(t, filepath.Join(dir, ".gemini", "prompts", p+".prompt.md"))
	}

	// Gemini / MCP
	assert.FileExists(t, filepath.Join(dir, ".gemini", "mcp.json"))
	assert.NoFileExists(t, filepath.Join(dir, ".vscode", "mcp.json"))
	
	// Verificar conteúdo do mcp.json do Gemini
	data, err := os.ReadFile(filepath.Join(dir, ".gemini", "mcp.json"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "context7")
	assert.Contains(t, string(data), "git")
}
