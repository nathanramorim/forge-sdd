package scaffold

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cfgNuvem(n config.NuvemConfig) config.ForgeConfig {
	c := config.DefaultForgeConfig()
	c.Nuvem = &n
	return c
}

func lerSettings(t *testing.T, dir string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

func TestNuvem_AtivarGravaScriptEHookAntesDoDoctor(t *testing.T) {
	dir := t.TempDir()
	_, err := MergeClaudeSessionStartHook(dir) // o doctor já estava lá
	require.NoError(t, err)
	mudou, err := AplicarNuvem(dir, cfgNuvem(config.NuvemConfig{Ativo: true}), "6.0.0-beta", false)
	require.NoError(t, err)
	assert.Len(t, mudou, 2)

	fi, err := os.Stat(filepath.Join(dir, NuvemScriptRel))
	require.NoError(t, err)
	assert.NotZero(t, fi.Mode()&0o100, "script executável")

	grupos := lerSettings(t, dir)["hooks"].(map[string]any)["SessionStart"].([]any)
	require.Len(t, grupos, 2)
	primeiro := grupos[0].(map[string]any)["hooks"].([]any)[0].(map[string]any)
	assert.Contains(t, primeiro["command"], "forge-nuvem.sh", "a nuvem roda antes do doctor, que precisa do forge")

	// idempotente
	mudou, err = AplicarNuvem(dir, cfgNuvem(config.NuvemConfig{Ativo: true}), "6.0.0-beta", false)
	require.NoError(t, err)
	assert.Empty(t, mudou)
}

func TestNuvem_DesativarRemoveSoOQueEraDele(t *testing.T) {
	dir := t.TempDir()
	_, err := MergeClaudeSessionStartHook(dir)
	require.NoError(t, err)
	_, err = AplicarNuvem(dir, cfgNuvem(config.NuvemConfig{Ativo: true}), "6.0.0-beta", false)
	require.NoError(t, err)
	_, err = AplicarNuvem(dir, cfgNuvem(config.NuvemConfig{Ativo: false}), "6.0.0-beta", false)
	require.NoError(t, err)
	grupos := lerSettings(t, dir)["hooks"].(map[string]any)["SessionStart"].([]any)
	require.Len(t, grupos, 1, "o hook do doctor continua")
	assert.NotContains(t, string(mustRead(t, filepath.Join(dir, ".claude", "settings.json"))), "forge-nuvem")
	_, err = os.Stat(filepath.Join(dir, NuvemScriptRel))
	assert.True(t, os.IsNotExist(err))
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return b
}

func TestNuvem_JSONInvalidoNuncaETocado(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	ruim := []byte("{ nao é json")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), ruim, 0o644))
	_, err := AplicarNuvem(dir, cfgNuvem(config.NuvemConfig{Ativo: true}), "6.0.0-beta", false)
	require.Error(t, err)
	assert.Equal(t, ruim, mustRead(t, filepath.Join(dir, ".claude", "settings.json")))
}

func TestNuvem_ScriptNpmEGit(t *testing.T) {
	npm := RenderNuvemScript(config.NuvemConfig{Ativo: true}, "v6.0.0-beta")
	assert.Contains(t, npm, `npm install -g --prefix "$HOME/.local" "@nathanramorim/forge-sdd@6.0.0-beta"`)
	assert.Contains(t, npm, `[ "${CLAUDE_CODE_REMOTE:-}" = "true" ] || exit 0`)
	assert.Contains(t, npm, "exec forge nuvem iniciar")
	git := RenderNuvemScript(config.NuvemConfig{Ativo: true, Instalar: "git", Ref: "claude/project-thread-kv01c3"}, "6.0.0-beta")
	assert.Contains(t, git, `git clone --quiet --depth 1 --branch "claude/project-thread-kv01c3" "https://github.com/nathanramorim/forge-sdd.git"`)
	assert.Contains(t, git, "go build")
	assert.Equal(t, "main", NuvemRef(config.NuvemConfig{Instalar: "git"}, "6.0.0"))
}

// Fora da nuvem o script sai na hora, sem instalar nem tocar em nada.
func TestNuvem_ScriptForaDaNuvemNaoFazNada(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("sem bash")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "s.sh")
	require.NoError(t, os.WriteFile(script, []byte(RenderNuvemScript(config.NuvemConfig{Ativo: true}, "6.0.0")), 0o755))
	c := exec.Command("bash", script)
	c.Env = []string{"HOME=" + dir, "PATH=/usr/bin:/bin", "CLAUDE_PROJECT_DIR=" + dir}
	out, err := c.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(out)))
	_, err = os.Stat(filepath.Join(dir, ".local"))
	assert.True(t, os.IsNotExist(err), "nada instalado fora da nuvem")
}
