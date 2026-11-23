package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateNaoPerdeDadoDoUsuario(t *testing.T) {
	t.Run("stack_e_banco", func(t *testing.T) {
		dir := t.TempDir()
		cfg := config.Defaults()
		cfg.Project = "proj"
		cfg.Stack, cfg.DB = "node", "postgres"
		cfg.Agents = []string{"claude", "gemini", "openai", "copilot"}
		_, err := Run(cfg, dir)
		require.NoError(t, err)
		// simula 5.0: .sddrc sem stack/db
		cfg5 := cfg
		cfg5.Stack, cfg5.DB = "", ""
		require.NoError(t, UpdateSddrc(cfg5, dir))
		rc, err := config.ReadSddrc(dir)
		require.NoError(t, err)
		require.Empty(t, rc.Stack)

		stack, db := RecoverStack(dir)
		rc.Stack, rc.DB = stack, db
		_, err = Run(rc.ToConfig(), dir)
		require.NoError(t, err)

		for _, f := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md", ".github/copilot-instructions.md"} {
			b, _ := os.ReadFile(filepath.Join(dir, f))
			assert.Contains(t, string(b), "stack node.", f)
		}
		for _, f := range []string{".gemini/system_instructions.md", ".openai/system_instructions.md"} {
			b, _ := os.ReadFile(filepath.Join(dir, f))
			assert.Contains(t, string(b), "- **Stack:** node", f)
			assert.Contains(t, string(b), "- **Banco:** postgres", f)
		}
		rc2, err := config.ReadSddrc(dir)
		require.NoError(t, err)
		assert.Equal(t, "node", rc2.Stack)
		assert.Equal(t, "postgres", rc2.DB)
	})
}

func TestUpdateRepopulaStackNosAdaptadores(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "| Runtime | python | |\n| DB | sqlite | |\n")
	cfg := config.Defaults()
	cfg.Project = "p"
	cfg.Agents = []string{"gemini"}
	cfg.Stack, cfg.DB = RecoverStack(dir)
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	b, _ := os.ReadFile(filepath.Join(dir, ".gemini/system_instructions.md"))
	assert.Contains(t, string(b), "- **Stack:** python")
	assert.Contains(t, string(b), "- **Banco:** sqlite")
	rc, err := config.ReadSddrc(dir)
	require.NoError(t, err)
	assert.Equal(t, "python", rc.Stack)
}

func TestRecuperaStackDoAdaptadorRecusaValorInvalido(t *testing.T) {
	for nome, v := range map[string]string{
		"aspas":  `no"de`,
		"chaves": "{{.X}}",
		"barra":  `a\b`,
		"longo":  strings.Repeat("a", 41),
		"espaco": "a b",
		"crase":  "`x`",
		"vazio":  "",
	} {
		t.Run(nome, func(t *testing.T) {
			dir := t.TempDir()
			testutil.WriteFile(t, dir, ".gemini/system_instructions.md", "- **Stack:** "+v+"\n- **Banco:** "+v+"\n")
			testutil.WriteFile(t, dir, "CLAUDE.md", "p: stack "+v+". Fluxo.\n")
			s, d := RecoverStack(dir)
			assert.Empty(t, s)
			assert.Empty(t, d)
		})
	}
}

func TestRecuperaStackNaOrdemDasFontes(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "| Runtime | rust | |\n| DB | none | |\n")
	testutil.WriteFile(t, dir, "sdd/spec/stack.md", "| Runtime | go | | |\n| DB | mongo | | |\n")
	testutil.WriteFile(t, dir, ".gemini/system_instructions.md", "- **Stack:** node\n- **Banco:** postgres\n")
	testutil.WriteFile(t, dir, "CLAUDE.md", "p: stack python. Fluxo.\n")
	s, d := RecoverStack(dir)
	assert.Equal(t, "rust", s)
	assert.Equal(t, "none", d)

	// constitution sem DB: cai no stack.md; sem Runtime válido: stack.md
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "| Runtime |  | |\n")
	s, d = RecoverStack(dir)
	assert.Equal(t, "go", s)
	assert.Equal(t, "mongo", d)

	// sem spec: gemini
	require.NoError(t, os.Remove(filepath.Join(dir, "sdd/spec/stack.md")))
	s, d = RecoverStack(dir)
	assert.Equal(t, "node", s)
	assert.Equal(t, "postgres", d)

	// só prosa: copilot quando CLAUDE/GEMINI/AGENTS não têm
	require.NoError(t, os.Remove(filepath.Join(dir, ".gemini/system_instructions.md")))
	require.NoError(t, os.Remove(filepath.Join(dir, "CLAUDE.md")))
	testutil.WriteFile(t, dir, ".github/copilot-instructions.md", "p: stack zig. Fluxo.\n")
	s, d = RecoverStack(dir)
	assert.Equal(t, "zig", s)
	assert.Empty(t, d)
}

func TestRecuperaStackProjetoSoClaude(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "CLAUDE.md", "proj: stack node. Fluxo por feature em `sdd/FLOW.md`.\n")
	s, d := RecoverStack(dir)
	assert.Equal(t, "node", s)
	assert.Empty(t, d, "banco não aparece na prosa: fica vazio, nunca none")

	empty := t.TempDir()
	s, d = RecoverStack(empty)
	assert.Empty(t, s)
	assert.Empty(t, d)
}
