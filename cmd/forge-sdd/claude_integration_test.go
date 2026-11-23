package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateClaudeIntegracaoSemMudarVersao(t *testing.T) {
	dir := modelProject(t, nil)
	mcp := filepath.Join(dir, ".mcp.json")
	settings := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.WriteFile(mcp, []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(settings, []byte(`{"permissions":{"allow":["Bash(ls)"]}}`), 0o644))

	// dry-run não grava
	out, err := updateRun(t, dir, true)
	require.NoError(t, err)
	assert.Contains(t, out, "MCP forge-sdd seria instalado")
	assert.Contains(t, out, "hook SessionStart seria instalado")
	got, _ := os.ReadFile(mcp)
	assert.Equal(t, `{}`, string(got))

	out, err = updateRun(t, dir, false)
	require.NoError(t, err)
	assert.Contains(t, out, "MCP forge-sdd instalado")
	assert.Contains(t, out, "hook SessionStart instalado")
	data, _ := os.ReadFile(mcp)
	assert.Contains(t, string(data), "forge-sdd")
	s, _ := os.ReadFile(settings)
	assert.Contains(t, string(s), "forge doctor --harness")
	assert.Contains(t, string(s), "Bash(ls)")

	// negativo: segunda rodada não muda nada
	out, err = updateRun(t, dir, false)
	require.NoError(t, err)
	assert.NotContains(t, out, "instalado")
	after, _ := os.ReadFile(mcp)
	assert.Equal(t, string(data), string(after))
}

func TestUpdateClaudeIntegracaoMCPInvalidoIntacto(t *testing.T) {
	dir := modelProject(t, nil)
	mcp := filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.WriteFile(mcp, []byte(`{"mcpServers":`), 0o644))
	out, err := updateRun(t, dir, false)
	require.Error(t, err) // 118-06c: JSON inválido aborta o update
	assert.NotContains(t, out, "MCP forge-sdd instalado")
	got, _ := os.ReadFile(mcp)
	assert.Equal(t, `{"mcpServers":`, string(got))
}

func TestUpdateClaudeSymlinkMCPNaoSegue(t *testing.T) {
	dir := modelProject(t, nil)
	fora := filepath.Join(t.TempDir(), "alvo.json")
	require.NoError(t, os.WriteFile(fora, []byte(`{"x":1}`), 0o644))
	mcp := filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.Remove(mcp))
	require.NoError(t, os.Symlink(fora, mcp))
	_, err := updateRun(t, dir, false)
	require.NoError(t, err)
	got, _ := os.ReadFile(fora)
	assert.Equal(t, `{"x":1}`, string(got))
	fi, _ := os.Lstat(mcp)
	assert.NotZero(t, fi.Mode()&os.ModeSymlink)
}

func TestUpdateClaudeSymlinkSettingsNaoSegue(t *testing.T) {
	dir := modelProject(t, nil)
	fora := filepath.Join(t.TempDir(), "settings.json")
	require.NoError(t, os.WriteFile(fora, []byte(`{"x":1}`), 0o644))
	st := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.Remove(st))
	require.NoError(t, os.Symlink(fora, st))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{}`), 0o644))
	out, err := updateRun(t, dir, false)
	require.NoError(t, err)
	got, _ := os.ReadFile(fora)
	assert.Equal(t, `{"x":1}`, string(got))
	assert.Contains(t, out, "MCP forge-sdd instalado") // o resto segue
}

func TestUpdateClaudeDirSymlinkParaForaNaoGrava(t *testing.T) {
	dir := modelProject(t, nil)
	fora := filepath.Join(t.TempDir(), "claude")
	require.NoError(t, os.Rename(filepath.Join(dir, ".claude"), fora))
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".claude")))
	before, _ := os.ReadFile(filepath.Join(fora, "settings.json"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{}`), 0o644))
	out, err := updateRun(t, dir, false)
	require.NoError(t, err)
	after, _ := os.ReadFile(filepath.Join(fora, "settings.json"))
	assert.Equal(t, string(before), string(after))
	assert.Contains(t, out, "MCP forge-sdd instalado")
}

func TestUpdateClaudeMCPExistenteDiferenteRespeitadoEModo0600(t *testing.T) {
	dir := modelProject(t, nil)
	mcp := filepath.Join(dir, ".mcp.json")
	orig := `{"mcpServers":{"forge-sdd":{"command":"meu-forge"}}}`
	require.NoError(t, os.WriteFile(mcp, []byte(orig), 0o600))
	_, err := updateRun(t, dir, false)
	require.NoError(t, err)
	got, _ := os.ReadFile(mcp)
	assert.Equal(t, orig, string(got))

	// modo preservado quando a gravação acontece
	require.NoError(t, os.WriteFile(mcp, []byte(`{"outro":1}`), 0o600))
	require.NoError(t, os.Chmod(mcp, 0o600))
	_, err = updateRun(t, dir, false)
	require.NoError(t, err)
	fi, _ := os.Stat(mcp)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	data, _ := os.ReadFile(mcp)
	assert.Contains(t, string(data), "forge-sdd")
	assert.Contains(t, string(data), "outro")
}

func TestDoctorAcusaHookSemScript(t *testing.T) {
	dir := modelProject(t, nil)
	var err error
	out := captureStdout(t, func() { _, err = runRoot(t, "doctor", dir) })
	require.NoError(t, err)
	assert.NotContains(t, out, "Erro de configuração")

	require.NoError(t, os.Remove(filepath.Join(dir, "scripts", "pr_guard.py")))
	// O comando novo tolera o script ausente; só o comando antigo bloquearia o Bash.
	out = captureStdout(t, func() { _, err = runRoot(t, "doctor", dir) })
	require.NoError(t, err)
	assert.NotContains(t, out, "Erro de configuração")
	sp := filepath.Join(dir, ".claude", "settings.json")
	cur, _ := os.ReadFile(sp)
	require.NoError(t, os.WriteFile(sp, []byte(strings.ReplaceAll(string(cur), "forge hook pr-guard", "uv run scripts/pr_guard.py --hook")), 0o644))
	out = captureStdout(t, func() { _, err = runRoot(t, "doctor", dir) })
	require.Error(t, err)
	assert.Contains(t, out, "Erro de configuração")
	assert.Contains(t, out, "forge update")
}

func TestUpdateReparaHookSemScript(t *testing.T) {
	dir := modelProject(t, nil)
	script := filepath.Join(dir, "scripts", "pr_guard.py")
	require.NoError(t, os.Remove(script))

	out, err := updateRun(t, dir, false)
	require.NoError(t, err)
	assert.Contains(t, out, "scripts/pr_guard.py instalado")
	_, statErr := os.Stat(script)
	require.NoError(t, statErr)
	s, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Equal(t, 1, strings.Count(string(s), "hook pr-guard"), "não duplica o hook")
}

// jsonInvalidoProject prepara um projeto 5.0 com settings.json e .mcp.json inválidos.
func jsonInvalidoProject(t *testing.T) (dir, settings, mcp string) {
	t.Helper()
	dir = modelProject(t, nil)
	settings = filepath.Join(dir, ".claude", "settings.json")
	mcp = filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.WriteFile(settings, []byte(`{ "hooks": {bad`), 0o644))
	require.NoError(t, os.WriteFile(mcp, []byte(`{"mcpServers": `), 0o644))
	return dir, settings, mcp
}

func TestUpdateJSONInvalidoSaiComErroENaoGravaVersao(t *testing.T) {
	dir, settings, mcp := jsonInvalidoProject(t)
	versaoAntes, _ := os.ReadFile(filepath.Join(dir, "sdd", ".sdd-version"))
	rcAntes, _ := os.ReadFile(filepath.Join(dir, "sdd", ".sddrc"))
	_ = updateCmd.Flags().Set("version", "9.9.9")
	t.Cleanup(func() { _ = updateCmd.Flags().Set("version", "") })

	var errBuf bytes.Buffer
	rootCmd.SetErr(&errBuf)
	t.Cleanup(func() { rootCmd.SetErr(nil) })
	_, err := updateRun(t, dir, false)
	require.Error(t, err)

	got, _ := os.ReadFile(settings)
	assert.Equal(t, `{ "hooks": {bad`, string(got))
	got, _ = os.ReadFile(mcp)
	assert.Equal(t, `{"mcpServers": `, string(got))
	versaoDepois, _ := os.ReadFile(filepath.Join(dir, "sdd", ".sdd-version"))
	rcDepois, _ := os.ReadFile(filepath.Join(dir, "sdd", ".sddrc"))
	assert.Equal(t, string(versaoAntes), string(versaoDepois))
	assert.Equal(t, string(rcAntes), string(rcDepois))
	// a orientação sai uma vez só por arquivo
	assert.Equal(t, 1, strings.Count(errBuf.String(), ".mcp.json"), errBuf.String())
	assert.NotContains(t, errBuf.String(), "servidor MCP não registrado")
	assert.NotContains(t, errBuf.String(), "MCP forge-sdd não registrado")
}

func TestUpdateSemTTYOrientaYes(t *testing.T) {
	dir := modelProject(t, nil)
	old := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdinIsTerminal = old })
	rootCmd.SetArgs([]string{"update", dir})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--yes")
	assert.NotContains(t, err.Error(), "TTY")
}

func TestUpdateDryRunRelataJSONInvalidoSemFalhar(t *testing.T) {
	dir, settings, mcp := jsonInvalidoProject(t)
	var errBuf bytes.Buffer
	rootCmd.SetErr(&errBuf)
	t.Cleanup(func() { rootCmd.SetErr(nil) })
	_, err := updateRun(t, dir, true)
	require.NoError(t, err)
	assert.Contains(t, errBuf.String(), ".mcp.json")
	got, _ := os.ReadFile(settings)
	assert.Equal(t, `{ "hooks": {bad`, string(got))
	got, _ = os.ReadFile(mcp)
	assert.Equal(t, `{"mcpServers": `, string(got))
}
