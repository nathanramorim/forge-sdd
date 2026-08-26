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
		"templates/.agents/rules/README.md.tmpl",
		"templates/.agents/commands/discovery.md.tmpl",
		"templates/.agents/commands/proxima-feature.md.tmpl",
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
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.9.4",
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
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.9.4",
		Agents: []string{config.AgentClaude},
	}

	created, err := Run(cfg, dir)
	require.NoError(t, err)
	require.NotEmpty(t, created)

	// deve ter CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	assert.FileExists(t, claudeMD, "CLAUDE.md deve ser criado para agente claude")

	// deve ter commands
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "proxima-feature.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "nova-feature.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "status.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "revisar.md"))

	// NÃO deve ter copilot-instructions (copilot não foi selecionado)
	assert.NoFileExists(t, filepath.Join(dir, ".github", "copilot-instructions.md"))
	assert.NoFileExists(t, filepath.Join(dir, ".vscode", "mcp.json"))

	t.Logf("✓ agente claude: %d arquivos criados", len(created))
}

func TestAgentGemini(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "node", DB: "postgres",
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.9.4",
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
		Telemetry: false, Lang: "pt-BR", SddVersion: "1.9.4",
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

	// 2. Segunda execução (simula upgrade para v1.9.4 com novos templates)
	cfgUpgrade := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "1.9.4", // nova versão
		Agents:     []string{config.AgentGemini},
	}
	_, err = Run(cfgUpgrade, dir)
	require.NoError(t, err)

	// 3. Asserções
	// a. progress.md DEVE ter sido preservado (conteúdo do usuário intacto)
	progressData, err := os.ReadFile(progressFile)
	require.NoError(t, err)
	assert.Equal(t, "conteudo customizado do progresso", string(progressData))

	// b. .sdd-version DEVE ter sido atualizado para 1.9.4
	versionData, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Contains(t, string(versionData), "1.9.4")

	// c. GEMINI.md DEVE ter sido atualizado (sobrescrito com o template da nova versão)
	geminiData, err := os.ReadFile(geminiFile)
	require.NoError(t, err)
	assert.NotEqual(t, "conteudo customizado do agente", string(geminiData))
}

func TestUpgradePreservesAgentRulesButRegeneratesCommands(t *testing.T) {
	dir := t.TempDir()

	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "2.0.0-beta",
		Agents:     []string{config.AgentClaude},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)

	rulesReadme := filepath.Join(dir, ".agents", "rules", "README.md")
	commandsDiscovery := filepath.Join(dir, ".agents", "commands", "discovery.md")
	claudeAdapter := filepath.Join(dir, ".claude", "commands", "discovery.md")

	require.FileExists(t, rulesReadme)
	require.FileExists(t, commandsDiscovery)
	require.FileExists(t, claudeAdapter)

	// Usuário cria/edita uma regra própria — deve sobreviver a qualquer update futuro.
	customRule := filepath.Join(dir, ".agents", "rules", "design-system.md")
	require.NoError(t, os.WriteFile(customRule, []byte("regra customizada do usuário"), 0644))
	require.NoError(t, os.WriteFile(rulesReadme, []byte("README customizado pelo usuário"), 0644))

	// Simula edição manual do corpo canônico de um comando — não deve sobreviver ao update
	// (corpo canônico é gerado pelo forge-sdd, não é domínio do usuário).
	require.NoError(t, os.WriteFile(commandsDiscovery, []byte("corpo alterado manualmente"), 0644))

	cfgUpgrade := cfg
	cfgUpgrade.SddVersion = "2.2.0-beta"
	_, err = Run(cfgUpgrade, dir)
	require.NoError(t, err)

	// a. Regra customizada do usuário permanece intacta.
	data, err := os.ReadFile(customRule)
	require.NoError(t, err)
	assert.Equal(t, "regra customizada do usuário", string(data))

	// b. README de .agents/rules/ também é preservado (mesma regra de domínio de sdd/).
	data, err = os.ReadFile(rulesReadme)
	require.NoError(t, err)
	assert.Equal(t, "README customizado pelo usuário", string(data))

	// c. Corpo canônico do comando é regenerado a partir do template (não é domínio do usuário).
	data, err = os.ReadFile(commandsDiscovery)
	require.NoError(t, err)
	assert.NotEqual(t, "corpo alterado manualmente", string(data))

	// d. Idempotência: rodar novamente com a mesma versão não falha nem duplica.
	_, err = Run(cfgUpgrade, dir)
	require.NoError(t, err)
	data, err = os.ReadFile(customRule)
	require.NoError(t, err)
	assert.Equal(t, "regra customizada do usuário", string(data))
}

func TestUpdateMigratesLegacyAgentDirToAgents(t *testing.T) {
	dir := t.TempDir()

	// Simula um projeto escaffoldado antes da renomeação .agent/ -> .agents/.
	legacyRules := filepath.Join(dir, ".agent", "rules")
	legacyCommands := filepath.Join(dir, ".agent", "commands")
	require.NoError(t, os.MkdirAll(legacyRules, 0755))
	require.NoError(t, os.MkdirAll(legacyCommands, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacyRules, "custom.md"), []byte("regra antiga do usuário"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(legacyCommands, "discovery.md"), []byte("corpo antigo"), 0644))

	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "2.2.0-beta",
		Agents:     []string{config.AgentClaude},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)

	// a. A pasta legada não existe mais — foi migrada, não duplicada.
	_, err = os.Stat(filepath.Join(dir, ".agent"))
	assert.True(t, os.IsNotExist(err), "diretório legado .agent/ deveria ter sido removido após a migração")

	// b. A regra customizada do usuário sobreviveu à migração, com o conteúdo original.
	migratedRule := filepath.Join(dir, ".agents", "rules", "custom.md")
	data, err := os.ReadFile(migratedRule)
	require.NoError(t, err)
	assert.Equal(t, "regra antiga do usuário", string(data))

	// c. O corpo do comando foi regenerado a partir do template atual (não é domínio do usuário).
	migratedCommand := filepath.Join(dir, ".agents", "commands", "discovery.md")
	data, err = os.ReadFile(migratedCommand)
	require.NoError(t, err)
	assert.NotEqual(t, "corpo antigo", string(data))

	// d. Idempotência: rodar novamente não falha nem perde a regra migrada.
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	data, err = os.ReadFile(migratedRule)
	require.NoError(t, err)
	assert.Equal(t, "regra antiga do usuário", string(data))
}

// TestUpdateCleansObsoleteClaudePromptSuffix cobre feat-03-02: projetos
// escaffoldados antes da correção de nome do adaptador Claude (feat-03-01)
// tinham .claude/commands/<comando>.prompt.md no disco. Rodar Run/update de
// novo deve criar o arquivo novo (<comando>.md, sempre regenerado — .claude/
// nunca é preservado por shouldPreserve) e remover o antigo, sem exigir
// nenhuma detecção de customização: .claude/commands/ já é 100% regenerado a
// cada init/update hoje, então o conteúdo do arquivo antigo nunca era
// domínio do usuário para começo de conversa.
func TestUpdateCleansObsoleteClaudePromptSuffix(t *testing.T) {
	dir := t.TempDir()

	// Simula um projeto escaffoldado antes da renomeação *.prompt.md -> *.md.
	legacyCommands := filepath.Join(dir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(legacyCommands, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacyCommands, "discovery.prompt.md"), []byte("corpo antigo"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(legacyCommands, "status.prompt.md"), []byte("corpo antigo status"), 0644))

	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Telemetry:  false,
		Lang:       "pt-BR",
		SddVersion: "2.3.0-beta",
		Agents:     []string{config.AgentClaude},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)

	// a. Arquivo novo existe.
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "discovery.md"))
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "status.md"))

	// b. Arquivo antigo (.prompt.md) foi removido — não fica lixo duplicado.
	assert.NoFileExists(t, filepath.Join(dir, ".claude", "commands", "discovery.prompt.md"))
	assert.NoFileExists(t, filepath.Join(dir, ".claude", "commands", "status.prompt.md"))

	// c. Idempotência: rodar de novo não falha nem recria o arquivo antigo.
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(dir, ".claude", "commands", "discovery.prompt.md"))

	// d. --dry-run não escreve nem remove nada.
	dryDir := t.TempDir()
	legacyDry := filepath.Join(dryDir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(legacyDry, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacyDry, "discovery.prompt.md"), []byte("corpo antigo"), 0644))
	dryCfg := cfg
	dryCfg.DryRun = true
	_, err = Run(dryCfg, dryDir)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dryDir, ".claude", "commands", "discovery.prompt.md"))
	assert.NoFileExists(t, filepath.Join(dryDir, ".claude", "commands", "discovery.md"))
}

