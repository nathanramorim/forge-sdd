package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projetoGestor(t *testing.T, pm string) string {
	t.Helper()
	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"cli":"claude"`+pm+`}`)
	return dir
}

func mcpsMd(status string) string {
	return "# MCPs\n\n| MCP | Status | Usado por |\n|-----|--------|-----------|\n| clickup | " + status + " | Discovery |\n"
}

func TestResolverGestorConcepcaoDesligadoPorPadrao(t *testing.T) {
	g, err := ResolverGestorConcepcao(t.TempDir(), "github")
	require.NoError(t, err)
	assert.Equal(t, GestorDesligado, g.Modo)
	// VCS=github sozinho não ativa
	g, err = ResolverGestorConcepcao(projetoGestor(t, ""), "github")
	require.NoError(t, err)
	assert.Equal(t, GestorDesligado, g.Modo)
	// project.json sem forge.json também fica desligado
	d := t.TempDir()
	writeForge(t, d, "sdd/modulos/project.json", "{}")
	g, _ = ResolverGestorConcepcao(d, "github")
	assert.Equal(t, GestorDesligado, g.Modo)
}

func TestResolverGestorConcepcaoGitHubQuandoConfigurado(t *testing.T) {
	d := projetoGestor(t, `,"pm_mcp":{"create_on_conception":true}`)
	g, err := ResolverGestorConcepcao(d, "github")
	require.NoError(t, err)
	assert.Equal(t, GestorGitHub, g.Modo)
	g, _ = ResolverGestorConcepcao(d, "gitlab")
	assert.Equal(t, GestorDesligado, g.Modo)

	d2 := projetoGestor(t, "")
	writeForge(t, d2, "sdd/modulos/project.json", "{}")
	g, _ = ResolverGestorConcepcao(d2, "github")
	assert.Equal(t, GestorGitHub, g.Modo)

	// symlink de project.json para fora do projeto não ativa
	fora := filepath.Join(t.TempDir(), "p.json")
	require.NoError(t, os.WriteFile(fora, []byte("{}"), 0o644))
	d3 := projetoGestor(t, "")
	require.NoError(t, os.MkdirAll(filepath.Join(d3, "sdd", "modulos"), 0o755))
	if err := os.Symlink(fora, filepath.Join(d3, "sdd", "modulos", "project.json")); err == nil {
		g, _ = ResolverGestorConcepcao(d3, "github")
		assert.Equal(t, GestorDesligado, g.Modo)
	}
}

func TestResolverGestorConcepcaoMCPAtivo(t *testing.T) {
	d := projetoGestor(t, `,"pm_mcp":{"provider":"clickup","create_on_conception":true}`)
	writeForge(t, d, "sdd/memory/mcps.md", mcpsMd("ativo"))
	g, err := ResolverGestorConcepcao(d, "github")
	require.NoError(t, err)
	assert.Equal(t, GestorConcepcao{Modo: GestorMCP, Provider: "clickup"}, g)
}

func TestResolverGestorConcepcaoMCPIndisponivelDesliga(t *testing.T) {
	for _, st := range []string{"indisponível", "não verificado (sem a ferramenta)", "ativo?", "Ativo"} {
		d := projetoGestor(t, `,"pm_mcp":{"provider":"clickup","create_on_conception":true}`)
		writeForge(t, d, "sdd/memory/mcps.md", mcpsMd(st))
		g, err := ResolverGestorConcepcao(d, "github")
		require.NoError(t, err)
		assert.Equal(t, GestorDesligado, g.Modo, st)
	}
}

func TestResolverGestorConcepcaoProviderDesconhecidoRecusado(t *testing.T) {
	d := projetoGestor(t, `,"pm_mcp":{"provider":"trello","create_on_conception":true}`)
	_, err := ResolverGestorConcepcao(d, "github")
	require.Error(t, err)
	// provider válido em supported, mas sem suporte na concepção
	d = projetoGestor(t, `,"pm_mcp":{"provider":"github-projects","create_on_conception":true}`)
	_, err = ResolverGestorConcepcao(d, "github")
	require.Error(t, err)
}

func TestResolverGestorConcepcaoMcpsMdAusenteOuMalformado(t *testing.T) {
	pm := `,"pm_mcp":{"provider":"jira","create_on_conception":true}`
	d := projetoGestor(t, pm)
	g, err := ResolverGestorConcepcao(d, "github")
	require.NoError(t, err)
	assert.Equal(t, GestorDesligado, g.Modo)
	for _, c := range []string{"", "| jira\n|\n||||", "\x00\xff| jira |", "lixo sem tabela"} {
		writeForge(t, d, "sdd/memory/mcps.md", c)
		g, err = ResolverGestorConcepcao(d, "github")
		require.NoError(t, err)
		assert.Equal(t, GestorDesligado, g.Modo)
	}
	// mcps.md como symlink para fora com "ativo" é ignorado
	fora := filepath.Join(t.TempDir(), "m.md")
	require.NoError(t, os.WriteFile(fora, []byte("| jira | ativo | x |\n"), 0o644))
	require.NoError(t, os.Remove(filepath.Join(d, "sdd/memory/mcps.md")))
	if err := os.Symlink(fora, filepath.Join(d, "sdd/memory/mcps.md")); err == nil {
		g, _ = ResolverGestorConcepcao(d, "github")
		assert.Equal(t, GestorDesligado, g.Modo)
	}
}

func TestLoadForgeConfigSemCreateOnConceptionRetrocompativel(t *testing.T) {
	d := projetoGestor(t, `,"pm_mcp":{"provider":"auto","supported":["clickup"],"status_on_implement":"A","status_on_pr_approved":"B","evidence_max_age_minutes":5}`)
	cfg, found, err := LoadForgeConfig(d)
	require.NoError(t, err)
	assert.True(t, found)
	assert.False(t, cfg.PMMCP.CreateOnConception)
	assert.False(t, DefaultForgeConfig().PMMCP.CreateOnConception)
}
