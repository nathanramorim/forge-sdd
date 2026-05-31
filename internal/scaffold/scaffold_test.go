package scaffold

import (
	"strings"
	"testing"

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
		"templates/sdd/.sdd-version.tmpl",
		"templates/sdd/.sddrc.tmpl",
		"templates/sdd/memory/constitution.md.tmpl",
		"templates/sdd/memory/progress.md.tmpl",
		"templates/sdd/features/index.md.tmpl",
		"templates/sdd/skills/index.md.tmpl",
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
