package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const versaoAlvoCmd50 = "5.2.1-beta"

// updateExplicito roda `forge update <dir> --yes --version X` (nunca --upgrade: ele consulta o NPM).
func updateExplicito(t *testing.T, dir string) string {
	t.Helper()
	t.Cleanup(func() {
		_ = updateCmd.Flags().Set("yes", "false")
		_ = updateCmd.Flags().Set("version", "")
	})
	var buf string
	stdout := captureAll(t, func() {
		var err error
		buf, err = runRoot(t, "update", dir, "--yes", "--version", versaoAlvoCmd50)
		require.NoError(t, err)
	})
	return stdout + buf
}

// TestUpgradeDe50 prova, de ponta a ponta pelo comando, o upgrade do projeto 5.0 congelado
// (internal/scaffold/testdata/legado-50) para Claude, Gemini, OpenAI/Codex e Copilot.
func TestUpgradeDe50(t *testing.T) {
	dir := t.TempDir()
	testutil.CopiaArvore(t, filepath.Join("..", "..", "internal", "scaffold", "testdata", "legado-50"), dir)
	ci := testutil.ReadFile(t, dir, ".github/workflows/ci.yml")
	claude := testutil.ReadFile(t, dir, "CLAUDE.md")
	regra := testutil.ReadFile(t, dir, ".agents/rules/sdd-seguranca.md")

	out := updateExplicito(t, dir)
	assert.Contains(t, out, "campo(s) de modo acrescentado(s)")

	rc, err := config.ReadSddrc(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"claude", "gemini", "copilot", "openai"}, rc.Agents)
	assert.Equal(t, "node", rc.Stack)
	assert.Equal(t, "postgres", rc.DB)
	assert.Equal(t, versaoAlvoCmd50, rc.Version)
	assert.Equal(t, versaoAlvoCmd50, strings.TrimSpace(testutil.ReadFile(t, dir, "sdd/.sdd-version")))

	// um caso por agente (TST-03): stack e banco chegam ao adaptador de cada um
	porAgente := map[string]map[string][]string{
		"gemini":  {".gemini/system_instructions.md": {"- **Stack:** node", "- **Banco:** postgres"}, "GEMINI.md": {"stack node."}},
		"openai":  {".openai/system_instructions.md": {"- **Stack:** node", "- **Banco:** postgres"}, "AGENTS.md": {"stack node."}},
		"copilot": {".github/copilot-instructions.md": {"stack node."}},
	}
	for agente, arqs := range porAgente {
		for f, trechos := range arqs {
			for _, tr := range trechos {
				assert.Contains(t, testutil.ReadFile(t, dir, f), tr, agente+" "+f)
			}
		}
	}
	assert.Equal(t, claude, testutil.ReadFile(t, dir, "CLAUDE.md"), "instrução editada intacta")
	assert.FileExists(t, filepath.Join(dir, "CLAUDE.novo.md"))
	assert.Equal(t, regra, testutil.ReadFile(t, dir, ".agents/rules/sdd-seguranca.md"))
	assert.FileExists(t, filepath.Join(dir, ".agents/rules/sdd-seguranca.novo.md"))
	assert.Contains(t, testutil.ReadFile(t, dir, "sdd/memory/constitution.md"), "| Modo do atestado | local |")

	s := testutil.ReadFile(t, dir, ".claude/settings.json")
	assert.Equal(t, 1, strings.Count(s, "forge hook pr-guard"))
	assert.Contains(t, s, "Bash(npm test)")
	assert.Contains(t, s, "npx prettier --write .")
	assert.FileExists(t, filepath.Join(dir, "scripts", "pr_guard.py"), "hook sem script")
	assert.Contains(t, testutil.ReadFile(t, dir, ".mcp.json"), "meu-servidor")

	assert.Equal(t, ci, testutil.ReadFile(t, dir, ".github/workflows/ci.yml"), "ci.yml byte a byte")
	assert.NoFileExists(t, filepath.Join(dir, ".github", "workflows", "forge-pronto.yml"))
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "config", "forge.json"))

	// idempotência: segunda execução não muda nenhum arquivo
	antes := testutil.SnapshotSHA(t, dir)
	updateExplicito(t, dir)
	assert.Equal(t, antes, testutil.SnapshotSHA(t, dir))

	// doctor sem erro e sem escrever no disco
	config.SetBinaryVersion(versaoAlvoCmd50)
	t.Cleanup(func() { config.SetBinaryVersion(version) })
	_ = doctorCmd.Flags().Set("verbose", "false")
	saida := captureAll(t, func() {
		_, err := runRoot(t, "doctor", dir)
		require.NoError(t, err)
	})
	assert.NotContains(t, saida, "✗", saida)
	assert.Equal(t, antes, testutil.SnapshotSHA(t, dir), "doctor não escreve")
}
