package scaffold

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractDescription_UsoLine(t *testing.T) {
	content := "# Prompt: archive\n\n**Uso:** Peça \"/archive\" ou \"limpar progresso\"\n\n**Ação:**\n1. Faça algo.\n"
	assert.Equal(t, `Peça "/archive" ou "limpar progresso"`, extractDescription(content))
}

func TestExtractDescription_FrontMatterDescription(t *testing.T) {
	content := "---\ndescription: \"Health check da estrutura SDD de {{.Project}}\"\nagent: agent\n---\n\nConteúdo do prompt.\n"
	got := extractDescription(renderCommandTemplate("t", content, config.Config{Project: "acme"}))
	assert.Equal(t, "Health check da estrutura SDD de acme", got)
}

func TestExtractDescription_FallbackSkipsHeadersAndLabels(t *testing.T) {
	content := "**Ação:**\n1. **PASSO 1 MANDATÓRIO:** Crie a branch `feat/*` localmente.\n"
	got := extractDescription(content)
	assert.Equal(t, "PASSO 1 MANDATÓRIO: Crie a branch `feat/*` localmente.", got)
}

func TestExtractDescription_NoUsableLineReturnsEmpty(t *testing.T) {
	content := "# Título\n\n**Ação:**\n"
	assert.Equal(t, "", extractDescription(content))
}

func TestTruncateDescription_BreaksAtWordBoundary(t *testing.T) {
	long := strings.Repeat("palavra ", 20)
	got := truncateDescription(long)
	assert.True(t, len([]rune(got)) <= descMaxLen+1) // +1 pela reticência "…"
	assert.False(t, strings.HasSuffix(strings.TrimSuffix(got, "…"), " "))
}

func TestCommandCheatSheet_ClaudeHasAllElevenCommandsInOrder(t *testing.T) {
	cfg := config.Config{Project: "acme", Agents: []string{config.AgentClaude}}
	commands := CommandCheatSheet(cfg)

	require.Len(t, commands, len(commandOrder))
	for i, c := range commands {
		assert.Equal(t, "/"+commandOrder[i], c.Command)
		assert.NotEmpty(t, c.Description, "comando %s deveria ter descrição", c.Command)
	}
}

func TestCommandCheatSheet_CopilotOmitsMissingC4Architecture(t *testing.T) {
	cfg := config.Config{Project: "acme", Agents: []string{config.AgentCopilot}}
	commands := CommandCheatSheet(cfg)

	for _, c := range commands {
		assert.NotEqual(t, "/c4-architecture", c.Command, "copilot não possui c4-architecture.prompt.md.tmpl")
	}
	require.Len(t, commands, len(commandOrder)-1)
}

func TestCommandCheatSheet_RendersProjectPlaceholder(t *testing.T) {
	cfg := config.Config{Project: "minhaapp", Agents: []string{config.AgentCopilot}}
	commands := CommandCheatSheet(cfg)

	for _, c := range commands {
		assert.NotContains(t, c.Description, "{{.Project}}", "placeholder de template vazou na descrição de %s", c.Command)
	}
}

func TestCommandCheatSheet_DefaultsToCopilotWhenNoAgents(t *testing.T) {
	commands := CommandCheatSheet(config.Config{Project: "acme"})
	require.Len(t, commands, len(commandOrder)-1)
}
