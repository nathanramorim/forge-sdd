package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readSettings(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	return m
}

func hookCommands(m map[string]any) []string {
	var cmds []string
	hooks, _ := m["hooks"].(map[string]any)
	groups, _ := hooks["UserPromptSubmit"].([]any)
	for _, g := range groups {
		for _, h := range g.(map[string]any)["hooks"].([]any) {
			cmds = append(cmds, h.(map[string]any)["command"].(string))
		}
	}
	return cmds
}

func TestMergeClaudeInboxHook_ArquivoNovo(t *testing.T) {
	dir := t.TempDir()
	changed, err := MergeClaudeInboxHook(dir)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, []string{inboxHookCommand}, hookCommands(readSettings(t, dir)))
}

func TestMergeClaudeInboxHook_PreservaOutrasChavesEEhIdempotente(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	existing := `{"model":"opus","permissions":{"allow":["Bash(go test:*)"]},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo antes"}]}],"UserPromptSubmit":[{"hooks":[{"type":"command","command":"echo meu-hook"}]}]}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(existing), 0o644))

	changed, err := MergeClaudeInboxHook(dir)
	require.NoError(t, err)
	assert.True(t, changed)
	m := readSettings(t, dir)
	assert.Equal(t, "opus", m["model"])
	assert.NotNil(t, m["permissions"])
	assert.NotNil(t, m["hooks"].(map[string]any)["PreToolUse"], "outros hooks preservados")
	assert.Equal(t, []string{"echo meu-hook", inboxHookCommand}, hookCommands(m))

	before, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	changed, err = MergeClaudeInboxHook(dir)
	require.NoError(t, err)
	assert.False(t, changed, "segunda execução não muda nada")
	after, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Equal(t, string(before), string(after))
}

func TestMergeClaudeInboxHook_NaoTocaEmJSONInvalidoOuFormaInesperada(t *testing.T) {
	for name, content := range map[string]string{
		"json inválido":      `{"hooks":`,
		"conteúdo extra":     `{} {}`,
		"não é objeto":       `[1,2]`,
		"hooks não é objeto": `{"hooks":[]}`,
		"submit não é lista": `{"hooks":{"UserPromptSubmit":"x"}}`,
		"null no topo":       `null`,
		"texto qualquer":     `isto nao e json`,
	} {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
		p := filepath.Join(dir, ".claude", "settings.json")
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
		changed, err := MergeClaudeInboxHook(dir)
		require.Error(t, err, name)
		assert.False(t, changed, name)
		data, _ := os.ReadFile(p)
		assert.Equal(t, content, string(data), "%s: o arquivo do usuário fica intacto", name)
		assert.Contains(t, err.Error(), "manualmente", name)
	}
}

func eventoCmds(m map[string]any, ev string) []string {
	var cmds []string
	hooks, _ := m["hooks"].(map[string]any)
	groups, _ := hooks[ev].([]any)
	for _, g := range groups {
		for _, h := range g.(map[string]any)["hooks"].([]any) {
			cmds = append(cmds, h.(map[string]any)["command"].(string))
		}
	}
	return cmds
}

func TestHookSessionStartIdempotente(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	seed := `{"model":"opus","hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"echo meu"}]}]}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(seed), 0o644))

	changed, err := MergeClaudeSessionStartHook(dir)
	require.NoError(t, err)
	assert.True(t, changed)
	// negativo: o segundo update não duplica
	changed, err = MergeClaudeSessionStartHook(dir)
	require.NoError(t, err)
	assert.False(t, changed)
	m := readSettings(t, dir)
	assert.Equal(t, []string{"echo meu", doctorHookCommand}, eventoCmds(m, "SessionStart"))
	assert.Equal(t, "opus", m["model"])

	// JSON inválido nunca é tocado
	bad := []byte(`{"hooks": `)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), bad, 0o644))
	_, err = MergeClaudeSessionStartHook(dir)
	require.Error(t, err)
	got, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Equal(t, bad, got)
}

func TestMCPServerRegistroIdempotente(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.WriteFile(p, []byte(`{"mcpServers":{"outro":{"command":"x"}}}`), 0o644))
	changed, err := MergeClaudeMCPServer(dir)
	require.NoError(t, err)
	assert.True(t, changed)
	changed, err = MergeClaudeMCPServer(dir)
	require.NoError(t, err)
	assert.False(t, changed)
	data, _ := os.ReadFile(p)
	var m map[string]map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Contains(t, m["mcpServers"], "outro")
	assert.Contains(t, m["mcpServers"], "forge-sdd")

	require.NoError(t, os.WriteFile(p, []byte(`[1]`), 0o644))
	_, err = MergeClaudeMCPServer(dir)
	require.Error(t, err)
	got, _ := os.ReadFile(p)
	assert.Equal(t, "[1]", string(got))
}

func TestInstrucaoDoctorNoAgentsMd(t *testing.T) {
	for _, f := range []string{"agents/openai/AGENTS.md.tmpl", "agents/claude/CLAUDE.md.tmpl", "agents/gemini/GEMINI.md.tmpl"} {
		data, err := templatesFS.ReadFile("templates/" + f)
		require.NoError(t, err, f)
		assert.Contains(t, string(data), "forge_doctor", f)
		assert.Contains(t, string(data), "forge doctor --harness", f)
	}
}

func TestMergePreToolUseHook(t *testing.T) {
	t.Run("arquivo novo grava com matcher Bash", func(t *testing.T) {
		dir := t.TempDir()
		changed, err := MergeClaudePreToolUseHRook(dir)
		require.NoError(t, err)
		assert.True(t, changed)

		m := readSettings(t, dir)
		hooks, _ := m["hooks"].(map[string]any)
		groups, _ := hooks["PreToolUse"].([]any)
		require.Len(t, groups, 1)
		entrada := groups[0].(map[string]any)
		assert.Equal(t, "Bash", entrada["matcher"], "o hook só roda em comandos de shell")

		inner := entrada["hooks"].([]any)
		require.Len(t, inner, 1)
		assert.Equal(t, prGuardCommand, inner[0].(map[string]any)["command"])
	})

	t.Run("preserva o hook do usuário e é idempotente", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
		existing := `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo antes"}]}]}}`
		p := filepath.Join(dir, ".claude", "settings.json")
		require.NoError(t, os.WriteFile(p, []byte(existing), 0o644))

		changed, err := MergeClaudePreToolUseHRook(dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, []string{"echo antes", prGuardCommand}, eventoCmds(readSettings(t, dir), "PreToolUse"))

		before, _ := os.ReadFile(p)
		changed, err = MergeClaudePreToolUseHRook(dir)
		require.NoError(t, err)
		assert.False(t, changed, "segunda execução não muda nada")
		after, _ := os.ReadFile(p)
		assert.Equal(t, string(before), string(after))
	})
}

func TestMergePreToolUseHookJSONInvalidoNaoETocado(t *testing.T) {
	for name, content := range map[string]string{
		"json inválido":          `{"hooks":`,
		"conteúdo extra":         `{} {}`,
		"não é objeto":           `[1,2]`,
		"hooks não é objeto":     `{"hooks":[]}`,
		"pretooluse não é lista": `{"hooks":{"PreToolUse":"x"}}`,
		"null no topo":           `null`,
		"texto qualquer":         `isto nao e json`,
	} {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
		p := filepath.Join(dir, ".claude", "settings.json")
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))

		changed, err := MergeClaudePreToolUseHRook(dir)
		require.Error(t, err, name)
		assert.False(t, changed, name)
		data, _ := os.ReadFile(p)
		assert.Equal(t, content, string(data), "%s: o arquivo do usuário fica intacto", name)
		assert.Contains(t, err.Error(), "manualmente", name)
	}
}

func TestHookPRGuardMigraComandoAntigo(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	antigo := `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"uv run scripts/pr_guard.py --hook"}]}]}}`
	p := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.WriteFile(p, []byte(antigo), 0o644))

	changed, err := MergeClaudePreToolUseHRook(dir)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, []string{"forge hook pr-guard"}, eventoCmds(readSettings(t, dir), "PreToolUse"))

	changed, err = MergeClaudePreToolUseHRook(dir)
	require.NoError(t, err)
	assert.False(t, changed, "já migrado: nada muda")
}

func TestHookPRGuardNaoDuplicaNemTocaHooksDoUsuario(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	cfg := `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo meu"},{"type":"command","command":"uv run scripts/pr_guard.py --hook"}]},{"matcher":"Edit","hooks":[{"type":"command","command":"uv run scripts/pr_guard.py --hook --extra"}]}]}}`
	p := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.WriteFile(p, []byte(cfg), 0o644))

	_, err := MergeClaudePreToolUseHRook(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"echo meu", "forge hook pr-guard", "uv run scripts/pr_guard.py --hook --extra"},
		eventoCmds(readSettings(t, dir), "PreToolUse"), "troca só o comando exato, sem duplicar nem tocar o resto")

	invalido := `{"hooks":`
	require.NoError(t, os.WriteFile(p, []byte(invalido), 0o644))
	_, err = MergeClaudePreToolUseHRook(dir)
	require.Error(t, err)
	got, _ := os.ReadFile(p)
	assert.Equal(t, invalido, string(got), "JSON inválido não é tocado")
}

func TestJSONInvalidoNoSettingsRetornaErroTipado(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(`{ "hooks": {bad`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{"mcpServers": `), 0o644))

	_, err := MergeClaudeSessionStartHook(dir)
	var ij *InvalidJSONError
	require.ErrorAs(t, err, &ij)
	assert.Equal(t, filepath.Join(dir, ".claude", "settings.json"), ij.Path)

	bad := ClaudeJSONInvalidos(dir)
	require.Len(t, bad, 2)
	assert.Contains(t, bad[1].Path, ".mcp.json")

	// caso conforme: JSON válido não gera erro
	ok := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(ok, ".mcp.json"), []byte(`{}`), 0o644))
	assert.Empty(t, ClaudeJSONInvalidos(ok))
}

func TestJSONInvalidoNaoReescreveOArquivo(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	settings := filepath.Join(dir, ".claude", "settings.json")
	mcp := filepath.Join(dir, ".mcp.json")
	require.NoError(t, os.WriteFile(settings, []byte(`{ "hooks": {bad`), 0o644))
	require.NoError(t, os.WriteFile(mcp, []byte(`{"mcpServers": `), 0o644))

	SyncClaudeIntegration(dir, "5.0.0", false)
	_, _ = MergeClaudeMCPServer(dir)
	_, _ = MergeClaudeInboxHook(dir)

	got, _ := os.ReadFile(settings)
	assert.Equal(t, `{ "hooks": {bad`, string(got))
	got, _ = os.ReadFile(mcp)
	assert.Equal(t, `{"mcpServers": `, string(got))
}

func TestAvisoDoMCPSaiUmaVez(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{"mcpServers": `), 0o644))
	_, warns := SyncClaudeIntegration(dir, "5.0.0", false)
	n := 0
	for _, w := range warns {
		if strings.Contains(w, ".mcp.json") {
			n++
		}
	}
	assert.Equal(t, 1, n, "%v", warns)
	assert.Len(t, ClaudeJSONInvalidos(dir), 1)
}
