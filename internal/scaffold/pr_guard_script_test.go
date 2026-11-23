package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forge-sdd/cli/internal/config"
)

func cfgClaude(version string) config.Config {
	return config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: version, Agents: []string{config.AgentClaude}}
}

func settingsTemPRGuard(t *testing.T, dir string) bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "pr_guard.py")
}

func linhasIniciais(t *testing.T, path string, n int) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	l := strings.Split(string(data), "\n")
	return l[:n]
}

func TestInitEntregaPRGuardAntesDoHook(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(cfgClaude("5.2.0"), dir)
	require.NoError(t, err)

	script := filepath.Join(dir, "scripts", "pr_guard.py")
	fi, err := os.Stat(script)
	require.NoError(t, err)
	assert.NotZero(t, fi.Mode().Perm()&0o100, "script executável")
	l := linhasIniciais(t, script, 3)
	assert.True(t, strings.HasPrefix(l[0], "#!"))
	assert.Equal(t, "# forge-sdd:generated v5.2.0", l[1], "marcador entre o shebang e o bloco PEP 723")
	assert.Equal(t, "# /// script", l[2])
	assert.True(t, settingsTemPRGuard(t, dir))
}

func TestUpdateEntregaPRGuardAntesDoHook(t *testing.T) {
	dir := t.TempDir()
	changed, warns := SyncClaudeIntegration(dir, "5.2.0", true)
	assert.Empty(t, warns)
	assert.Contains(t, changed, "scripts/pr_guard.py")
	_, err := os.Stat(filepath.Join(dir, "scripts", "pr_guard.py"))
	assert.True(t, os.IsNotExist(err), "dry-run não grava")

	changed, warns = SyncClaudeIntegration(dir, "5.2.0", false)
	assert.Empty(t, warns)
	assert.Contains(t, changed, "scripts/pr_guard.py")
	assert.True(t, IsGeneratedPRGuard(dir))
	assert.True(t, settingsTemPRGuard(t, dir))

	// negativo: segunda rodada não muda nada e não duplica o hook
	changed, _ = SyncClaudeIntegration(dir, "5.2.0", false)
	assert.NotContains(t, changed, "scripts/pr_guard.py")
	assert.NotContains(t, changed, "hook PreToolUse")
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Equal(t, 1, strings.Count(string(data), "pr_guard.py"))
}

func TestFalhaDeGravacaoNaoInstalaOHook(t *testing.T) {
	dir := t.TempDir()
	// `scripts` é um arquivo: a gravação do script falha.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts"), []byte("x"), 0o644))
	_, warns := SyncClaudeIntegration(dir, "5.2.0", false)
	require.NotEmpty(t, warns)
	assert.Contains(t, strings.Join(warns, "\n"), "hook PreToolUse não instalado")
	assert.False(t, settingsTemPRGuard(t, dir), "sem script, sem hook")

	dir2 := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "scripts"), []byte("x"), 0o644))
	_, err := Run(cfgClaude("5.2.0"), dir2)
	require.NoError(t, err)
	assert.False(t, settingsTemPRGuard(t, dir2))
}

func TestPRGuardScriptDoUsuarioFicaIntocado(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "scripts", "pr_guard.py")
	require.NoError(t, os.MkdirAll(filepath.Dir(script), 0o755))
	meu := "#!/usr/bin/env python3\n# meu guard\n"
	require.NoError(t, os.WriteFile(script, []byte(meu), 0o755))

	assert.True(t, shouldPreserve(script, dir))
	assert.True(t, PRGuardUserOwned(dir))
	_, err := Run(cfgClaude("9.0.0"), dir)
	require.NoError(t, err)
	got, _ := os.ReadFile(script)
	assert.Equal(t, meu, string(got))
	changed, _ := SyncClaudeIntegration(dir, "9.0.0", false)
	assert.NotContains(t, changed, "scripts/pr_guard.py")
	got, _ = os.ReadFile(script)
	assert.Equal(t, meu, string(got))
}

func TestPRGuardGeradoAntigoEAtualizado(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(cfgClaude("5.0.0"), dir)
	require.NoError(t, err)
	script := filepath.Join(dir, "scripts", "pr_guard.py")
	assert.Equal(t, "# forge-sdd:generated v5.0.0", linhasIniciais(t, script, 2)[1])
	assert.False(t, shouldPreserve(script, dir), "gerado não é preservado")

	_, err = Run(cfgClaude("5.2.0"), dir)
	require.NoError(t, err)
	assert.Equal(t, "# forge-sdd:generated v5.2.0", linhasIniciais(t, script, 2)[1])

	// negativo: versão menor não rebaixa o script
	_, err = Run(cfgClaude("5.1.0"), dir)
	require.NoError(t, err)
	assert.Equal(t, "# forge-sdd:generated v5.2.0", linhasIniciais(t, script, 2)[1])
}

func TestPRGuardTemplateIgualAoDoRepositorio(t *testing.T) {
	repo, err := os.ReadFile(filepath.Join("..", "..", "scripts", "pr_guard.py"))
	require.NoError(t, err)
	tmpl, err := templatesFS.ReadFile(prGuardTemplate)
	require.NoError(t, err)
	assert.Equal(t, string(repo), string(tmpl), "copie scripts/pr_guard.py para %s", prGuardTemplate)
	assert.NotContains(t, string(tmpl), prGuardMarker, "o marcador é injetado na gravação, não fica no template")
}
