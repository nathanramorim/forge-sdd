package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ciCfg() config.Config {
	return config.Config{
		Project: "demo", Stack: "go", DB: "none", Telemetry: true, Lang: "pt-BR", SddVersion: "5.1.0",
		Agents: []string{config.AgentCopilot, config.AgentClaude},
	}
}

func TestSemCINaoCriaWorkflow(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(ciCfg(), dir) // inclui o Copilot, que renderiza templates/.github
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(dir, ".github", "workflows", "forge-pronto.yml"))
	assert.NoDirExists(t, filepath.Join(dir, ".github", "workflows"))
	assert.NoDirExists(t, filepath.Join(dir, "ci"))
}

func TestComCICriaForgePronto(t *testing.T) {
	dir := t.TempDir()
	got, err := RenderCI(ciCfg(), dir)
	require.NoError(t, err)
	dest := filepath.Join(dir, ".github", "workflows", "forge-pronto.yml")
	assert.Equal(t, []string{dest}, got)
	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "# forge-sdd:generated"))
	assert.Contains(t, string(data), "name: pronto?")
	assert.Contains(t, string(data), "gh pr view")
	assert.Contains(t, string(data), "${{ secrets.GITHUB_TOKEN }}")
	assert.NotContains(t, string(data), "echo \"$GH_TOKEN")
	// segunda execução: idempotente
	_, err = RenderCI(ciCfg(), dir)
	require.NoError(t, err)
	again, _ := os.ReadFile(dest)
	assert.Equal(t, string(data), string(again))
}

func TestWorkflowDeclaraReadyForReview(t *testing.T) {
	data, err := templatesFS.ReadFile("templates/ci/forge-pronto.yml.tmpl")
	require.NoError(t, err)
	out, err := renderTemplate("x", data, ciCfg())
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "on:\n  pull_request:\n    types: [ready_for_review]")
	assert.Contains(t, s, "if: github.event_name == 'pull_request'")
	assert.Contains(t, s, "pull-requests: read")
	for _, stack := range []string{"go", "node", "python"} {
		c := ciCfg()
		c.Stack = stack
		o, err := renderTemplate("x", data, c)
		require.NoError(t, err)
		assert.Equal(t, s, string(o), "o workflow não depende da stack")
	}
}

func TestCIPreservaWorkflowDoUsuario(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, ".github", "workflows", "forge-pronto.yml")
	require.NoError(t, os.MkdirAll(filepath.Dir(dest), 0o755))
	meu := "name: meu\non: push\n"
	require.NoError(t, os.WriteFile(dest, []byte(meu), 0o644))
	got, err := RenderCI(ciCfg(), dir)
	require.NoError(t, err)
	assert.Empty(t, got)
	data, _ := os.ReadFile(dest)
	assert.Equal(t, meu, string(data))
}

func TestCINuncaTocaNoCiYml(t *testing.T) {
	dir := t.TempDir()
	ci := filepath.Join(dir, ".github", "workflows", "ci.yml")
	require.NoError(t, os.MkdirAll(filepath.Dir(ci), 0o755))
	require.NoError(t, os.WriteFile(ci, []byte("name: CI\non: push\n"), 0o644))
	old := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
	require.NoError(t, os.Chtimes(ci, old, old))
	_, err := RenderCI(ciCfg(), dir)
	require.NoError(t, err)
	_, err = Run(ciCfg(), dir)
	require.NoError(t, err)
	data, _ := os.ReadFile(ci)
	assert.Equal(t, "name: CI\non: push\n", string(data))
	fi, _ := os.Stat(ci)
	assert.True(t, fi.ModTime().Equal(old), "mtime do ci.yml mudou")
}

func TestUpgradeSddNaoPrometeOQueNaoEntrega(t *testing.T) {
	data, err := templatesFS.ReadFile("templates/.agents/skills/sdd-operacao/references/upgrade-sdd.md.tmpl")
	require.NoError(t, err)
	s := string(data)
	assert.Contains(t, s, "forge update --ci")
	assert.Contains(t, s, "forge-pronto.yml")
	assert.NotContains(t, s, "acrescenta só o job")
	assert.NotContains(t, s, "Scripts de política")
	assert.NotContains(t, s, "seguem cópia manual")
	// ci-politica.py e atestado.py só aparecem na seção que diz que o forge NÃO os entrega.
	i := strings.Index(s, "### O que o forge NÃO faz")
	require.GreaterOrEqual(t, i, 0)
	assert.NotContains(t, s[:i], "ci-politica.py")
	assert.NotContains(t, s[:i], "atestado.py")
}
