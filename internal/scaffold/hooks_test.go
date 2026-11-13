package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
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
