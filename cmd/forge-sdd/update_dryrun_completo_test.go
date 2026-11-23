package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const constituicaoSemModos = "# Constituição\n\n## Ferramentas e Integrações\n| Campo | Valor |\n|-------|-------|\n| Stack | go |\n\n## Outra\n"

// projetoParaDryRun monta um projeto já instalado (claude e gemini) com uma instrução e uma rule
// editadas pelo usuário, uma skill sdd-* defasada e uma Constituição sem os campos de modo.
func projetoParaDryRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Config{Project: "demo", SddVersion: "5.0.0", Lang: "pt-BR", Agents: []string{config.AgentClaude, config.AgentGemini}, NamingConvention: "sequencial"}
	_, err := scaffold.Run(cfg, dir)
	require.NoError(t, err)
	testutil.WriteFile(t, dir, "GEMINI.md", "# Minhas regras\nsó minhas\n")
	testutil.WriteFile(t, dir, ".agents/rules/sdd-seguranca.md", "regra editada por mim\n")
	testutil.WriteFile(t, dir, ".agents/skills/sdd-build/SKILL.md", "skill defasada\n")
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", constituicaoSemModos)
	require.NoError(t, os.Remove(filepath.Join(dir, ".claude", "settings.json")))
	_ = os.Remove(filepath.Join(dir, ".mcp.json"))
	t.Cleanup(func() { _ = updateCmd.Flags().Set("ci", "false") })
	return dir
}

func dryRun(t *testing.T, dir string, extra ...string) string {
	t.Helper()
	out, err := runRoot(t, append([]string{"update", dir, "--dry-run"}, extra...)...)
	require.NoError(t, err)
	return out
}

func TestUpdateDryRunListaOHookPreToolUse(t *testing.T) {
	dir := projetoParaDryRun(t)
	out := dryRun(t, dir)
	assert.Contains(t, out, "hook PreToolUse seria instalado")
	assert.Contains(t, out, "hook UserPromptSubmit seria instalado")
	assert.Contains(t, out, "hook SessionStart seria instalado")
	assert.Contains(t, out, "MCP forge-sdd seria instalado")
	assert.Contains(t, out, "Versão que seria gravada: ")
}

func TestUpdateDryRunListaArquivosReescritos(t *testing.T) {
	dir := projetoParaDryRun(t)
	out := dryRun(t, dir, "--ci")
	assert.Contains(t, out, "reescrever .agents/skills/sdd-build/SKILL.md")
	assert.Contains(t, out, "criar .github/workflows/forge-pronto.yml")
	assert.Contains(t, out, "scripts/pr_guard.py seria instalado")
}

func TestUpdateDryRunListaArquivosPreservados(t *testing.T) {
	dir := projetoParaDryRun(t)
	out := dryRun(t, dir)
	assert.Contains(t, out, "mantido GEMINI.md")
	assert.Contains(t, out, "mantido .agents/rules/sdd-seguranca.md")
	assert.Contains(t, out, "preservar GEMINI.md")
	assert.Contains(t, out, "preservar .agents/rules/sdd-seguranca.md")
}

func TestUpdateDryRunListaAcrescimosDaConstituicao(t *testing.T) {
	dir := projetoParaDryRun(t)
	out := dryRun(t, dir)
	assert.Contains(t, out, "Serão acrescentados na Constituição (")
	assert.NotContains(t, out, "Serão acrescentados na Constituição (0)")
	assert.Contains(t, out, "  acrescentar ")
	b, err := os.ReadFile(filepath.Join(dir, "sdd/memory/constitution.md"))
	require.NoError(t, err)
	assert.Equal(t, constituicaoSemModos, string(b))
}

func TestUpdateDryRunNaoAlteraNada(t *testing.T) {
	dir := projetoParaDryRun(t)
	antes := snapshotArquivos(t, dir)
	dryRun(t, dir, "--upgrade", "--ci")
	assert.Equal(t, antes, snapshotArquivos(t, dir), "o --dry-run não pode alterar nenhum arquivo")
}

func TestUpdateDryRunCIPreservaWorkflowDoUsuario(t *testing.T) {
	dir := projetoParaDryRun(t)
	testutil.WriteFile(t, dir, ".github/workflows/forge-pronto.yml", "name: meu\n")
	out := dryRun(t, dir, "--ci")
	assert.Contains(t, out, "preservar .github/workflows/forge-pronto.yml")
	assert.NotContains(t, out, "criar .github/workflows/forge-pronto.yml")
	b, err := os.ReadFile(filepath.Join(dir, ".github/workflows/forge-pronto.yml"))
	require.NoError(t, err)
	assert.Equal(t, "name: meu\n", string(b))
}
