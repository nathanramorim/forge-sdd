package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkTemplates(t *testing.T) {
	paths := Walk()

	require.NotEmpty(t, paths, "Walk() não deve retornar lista vazia")
	assert.GreaterOrEqual(t, len(paths), 25, "deve haver pelo menos 25 templates embutidos")

	// Verifica presença de arquivos essenciais
	essential := []string{
		"templates/.github/copilot-instructions.md.tmpl",
		"templates/.github/chatmodes/orquestrador.chatmode.md.tmpl",
		"templates/.github/chatmodes/archivist.chatmode.md.tmpl",
		"templates/.github/chatmodes/migrator.chatmode.md.tmpl",
		"templates/.github/prompts/proxima-feature.prompt.md.tmpl",
		"templates/.vscode/mcp.json.tmpl",
		"templates/agents/gemini/.gemini/mcp.json.tmpl",
		"templates/sdd/.sdd-version.tmpl",
		"templates/sdd/.sddrc.tmpl",
		"templates/sdd/memory/constitution.md.tmpl",
		"templates/sdd/memory/progress.md.tmpl",
		"templates/sdd/features/index.md.tmpl",
		"templates/sdd/skills/index.md.tmpl",
		"templates/sdd/releases/history.md.tmpl",
		"templates/sdd/.metrics/schema.json.tmpl",
	}

	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}

	for _, e := range essential {
		assert.True(t, pathSet[e], "template essencial ausente: %s", e)
	}

	// Todos os paths devem ser .tmpl
	for _, p := range paths {
		assert.True(t, strings.HasSuffix(p, ".tmpl"), "arquivo não-.tmpl encontrado: %s", p)
	}

	t.Logf("Total de templates embutidos: %d", len(paths))
	for _, p := range paths {
		t.Logf("  %s", p)
	}
}

func TestDryRunNoFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "go", DB: "none",
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.7.0",
		Agents: []string{config.AgentCopilot},
		DryRun: true,
	}

	listed, err := Run(cfg, dir)
	require.NoError(t, err)
	require.NotEmpty(t, listed, "dry-run deve retornar lista de caminhos")

	// nenhum arquivo deve ter sido criado
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "dry-run não deve criar arquivos em disco")

	t.Logf("✓ %d caminhos listados, 0 arquivos criados", len(listed))
}

func TestAgentClaude(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "go", DB: "none",
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.7.0",
		Agents: []string{config.AgentClaude},
	}

	created, err := Run(cfg, dir)
	require.NoError(t, err)
	require.NotEmpty(t, created)

	// deve ter CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	assert.FileExists(t, claudeMD, "CLAUDE.md deve ser criado para agente claude")

	// deve ter commands
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "proxima-feature.prompt.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "nova-feature.prompt.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "status.prompt.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "revisar.prompt.md"))

	// NÃO deve ter copilot-instructions (copilot não foi selecionado)
	assert.NoFileExists(t, filepath.Join(dir, ".github", "copilot-instructions.md"))
	assert.NoFileExists(t, filepath.Join(dir, ".vscode", "mcp.json"))

	t.Logf("✓ agente claude: %d arquivos criados", len(created))
}

func TestAgentGemini(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "node", DB: "postgres",
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.7.0",
		Agents: []string{config.AgentGemini},
	}

	created, err := Run(cfg, dir)
	require.NoError(t, err)
	require.NotEmpty(t, created)

	assert.FileExists(t, filepath.Join(dir, "GEMINI.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "system_instructions.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "skills", "orquestrador.chatmode.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "skills", "c4-architecture.chatmode.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "prompts", "status.prompt.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "prompts", "c4-architecture.prompt.md"))
	assert.FileExists(t, filepath.Join(dir, ".gemini", "mcp.json"))
	assert.NoFileExists(t, filepath.Join(dir, ".github", "copilot-instructions.md"))
	assert.NoFileExists(t, filepath.Join(dir, ".vscode", "mcp.json"))

	t.Logf("✓ agente gemini: %d arquivos criados", len(created))
}

func TestAgentMultiple(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "go", DB: "none",
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.7.0",
		Agents: []string{config.AgentCopilot, config.AgentClaude},
	}

	created, err := Run(cfg, dir)
	require.NoError(t, err)

	// deve ter ambos
	assert.FileExists(t, filepath.Join(dir, ".github", "copilot-instructions.md"))
	assert.FileExists(t, filepath.Join(dir, "CLAUDE.md"))
	assert.NoFileExists(t, filepath.Join(dir, "GEMINI.md"))

	t.Logf("✓ copilot+claude: %d arquivos criados", len(created))
}

func TestUpgradePreservesDomain(t *testing.T) {
	dir := t.TempDir()

	// 1. Primeira execução (simula inicialização em versão antiga v1.5.2)
	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "1.5.2", // mantém a primeira execução na versão antiga para testar o upgrade
		Agents:     []string{config.AgentGemini},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)

	// Modifica os arquivos locais para testar a preservação de conteúdo
	progressFile := filepath.Join(dir, "sdd", "memory", "progress.md")
	versionFile := filepath.Join(dir, "sdd", ".sdd-version")
	geminiFile := filepath.Join(dir, "GEMINI.md")

	err = os.WriteFile(progressFile, []byte("conteudo customizado do progresso"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(geminiFile, []byte("conteudo customizado do agente"), 0644)
	require.NoError(t, err)

	// 2. Segunda execução (simula upgrade para v1.7.0 com novos templates)
	cfgUpgrade := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "1.7.0", // nova versão
		Agents:     []string{config.AgentGemini},
	}
	_, err = Run(cfgUpgrade, dir)
	require.NoError(t, err)

	// 3. Asserções
	// a. progress.md DEVE ter sido preservado (conteúdo do usuário intacto)
	progressData, err := os.ReadFile(progressFile)
	require.NoError(t, err)
	assert.Equal(t, "conteudo customizado do progresso", string(progressData))

	// b. .sdd-version DEVE ter sido atualizado para 1.7.0
	versionData, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Contains(t, string(versionData), "1.7.0")

	// c. GEMINI.md DEVE ter sido atualizado (sobrescrito com o template da nova versão)
	geminiData, err := os.ReadFile(geminiFile)
	require.NoError(t, err)
	assert.NotEqual(t, "conteudo customizado do agente", string(geminiData))
}

