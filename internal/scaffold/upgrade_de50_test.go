package scaffold

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// versaoAlvo50 é a versão explícita do upgrade nos testes (nunca consulta o NPM).
const versaoAlvo50 = "6.0.0-beta"

// copiaLegado50 copia a fixture congelada (projeto 5.0.0 gerado pelo binário 5.0) para um diretório temporário.
func copiaLegado50(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testutil.CopiaArvore(t, filepath.Join("testdata", "legado-50"), dir)
	return dir
}

// upgradeDe50 reproduz o caminho do `forge update --yes --version X` na camada de arquivos:
// recupera stack/banco, grava a versão explícita e acrescenta os campos de modo.
func upgradeDe50(t *testing.T, dir string) {
	t.Helper()
	rc, err := config.ReadSddrc(dir)
	require.NoError(t, err)
	if rc.Stack == "" || rc.DB == "" {
		s, d := RecoverStack(dir)
		if rc.Stack == "" {
			rc.Stack = s
		}
		if rc.DB == "" {
			rc.DB = d
		}
	}
	cfg := rc.ToConfig()
	if cfg.NamingConvention == "" {
		cfg.NamingConvention = config.DetectNamingConvention(dir)
	}
	cfg.SddVersion = versaoAlvo50
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	_, err = AcrescentaCamposDeModo(dir, io.Discard)
	require.NoError(t, err)
}

func comandosPreToolUse(t *testing.T, dir string) []string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(testutil.ReadFile(t, dir, ".claude/settings.json")), &m))
	var cmds []string
	hooks, _ := m["hooks"].(map[string]any)
	groups, _ := hooks["PreToolUse"].([]any)
	for _, g := range groups {
		for _, h := range g.(map[string]any)["hooks"].([]any) {
			cmds = append(cmds, h.(map[string]any)["command"].(string))
		}
	}
	return cmds
}

func TestUpgradeDe50(t *testing.T) {
	dir := copiaLegado50(t)
	antes := testutil.SnapshotSHA(t, dir)
	ci := testutil.ReadFile(t, dir, ".github/workflows/ci.yml")
	claudeEditado := testutil.ReadFile(t, dir, "CLAUDE.md")
	regraEditada := testutil.ReadFile(t, dir, ".agents/rules/sdd-seguranca.md")
	require.Contains(t, claudeEditado, "Minhas regras")
	require.NotContains(t, testutil.ReadFile(t, dir, "sdd/memory/constitution.md"), "Modo do atestado")

	upgradeDe50(t, dir)

	// stack e banco preservados nos adaptadores (e gravados no .sddrc)
	rc, err := config.ReadSddrc(dir)
	require.NoError(t, err)
	assert.Equal(t, "node", rc.Stack)
	assert.Equal(t, "postgres", rc.DB)
	for _, f := range []string{"GEMINI.md", "AGENTS.md", ".github/copilot-instructions.md"} {
		assert.Contains(t, testutil.ReadFile(t, dir, f), "stack node.", f)
	}
	for _, f := range []string{".gemini/system_instructions.md", ".openai/system_instructions.md"} {
		c := testutil.ReadFile(t, dir, f)
		assert.Contains(t, c, "- **Stack:** node", f)
		assert.Contains(t, c, "- **Banco:** postgres", f)
	}

	// instrução editada intacta, com .novo.md ao lado; não editadas adotam o template novo
	assert.Equal(t, claudeEditado, testutil.ReadFile(t, dir, "CLAUDE.md"))
	assert.Contains(t, testutil.ReadFile(t, dir, "CLAUDE.novo.md"), generatedMarker)
	assert.Contains(t, testutil.ReadFile(t, dir, "GEMINI.md"), generatedMarker)
	assert.NoFileExists(t, filepath.Join(dir, "GEMINI.novo.md"))

	// rules: editada preservada, não editadas atualizadas
	assert.Equal(t, regraEditada, testutil.ReadFile(t, dir, ".agents/rules/sdd-seguranca.md"))
	assert.NotEqual(t, antes[".agents/rules/sdd-fluxo-e-portao.md"], testutil.SnapshotSHA(t, dir)[".agents/rules/sdd-fluxo-e-portao.md"], "rule não editada é atualizada")
	assert.NoFileExists(t, filepath.Join(dir, ".agents/rules/sdd-fluxo-e-portao.novo.md"))
	assert.Contains(t, testutil.ReadFile(t, dir, ".agents/rules/sdd-seguranca.novo.md"), "Guia Prático — Segurança")

	// campos de modo acrescentados sem alterar linhas existentes
	const_ := testutil.ReadFile(t, dir, "sdd/memory/constitution.md")
	for _, campo := range []string{"| Modo do atestado | local |", "| Modo de revisão humana | com-humano |", "| WIP humano | 3 |", "| Módulos permitidos | nenhum |", "| Início automático do Builder | desligado |"} {
		assert.Contains(t, const_, campo)
	}
	assert.Contains(t, const_, "| VCS / Work Item System | github |", "linha existente intacta")
	assert.Contains(t, const_, "| Runtime | node | |", "stack do usuário intacta")

	// versão: .sddrc e .sdd-version iguais
	assert.Equal(t, versaoAlvo50, rc.Version)
	assert.Equal(t, versaoAlvo50, strings.TrimSpace(testutil.ReadFile(t, dir, "sdd/.sdd-version")))

	// hook PreToolUse só com script presente
	assert.Equal(t, []string{"forge hook pr-guard"}, comandosPreToolUse(t, dir))
	assert.True(t, IsGeneratedPRGuard(dir))

	// ci.yml do usuário intacto; sem forge-pronto.yml sem --ci
	assert.Equal(t, ci, testutil.ReadFile(t, dir, ".github/workflows/ci.yml"))
	assert.NoFileExists(t, filepath.Join(dir, ".github", "workflows", "forge-pronto.yml"))

	// forge.json não é criado/sobrescrito pelo update
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "config", "forge.json"))

	// JSON do usuário preservado
	s := testutil.ReadFile(t, dir, ".claude/settings.json")
	assert.Contains(t, s, "Bash(npm test)")
	assert.Contains(t, s, "npx prettier --write .")
	assert.Contains(t, s, "forge inbox --unread --hook")
	assert.Contains(t, testutil.ReadFile(t, dir, ".mcp.json"), "meu-servidor")
	assert.Contains(t, testutil.ReadFile(t, dir, ".mcp.json"), "forge-sdd")
	assert.Empty(t, ClaudeJSONInvalidos(dir))

	// layout: nenhuma migração necessária
	plano, err := PlanLayoutMigration(dir)
	require.NoError(t, err)
	assert.True(t, plano.Empty())
}

func TestUpgradeDe50Idempotente(t *testing.T) {
	dir := copiaLegado50(t)
	upgradeDe50(t, dir)
	primeira := testutil.SnapshotSHA(t, dir)
	upgradeDe50(t, dir)
	assert.Equal(t, primeira, testutil.SnapshotSHA(t, dir), "segunda execução não muda nenhum arquivo")
}

func TestUpgradeDe50NaoTocaNoCiYml(t *testing.T) {
	dir := copiaLegado50(t)
	antes, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "ci.yml"))
	require.NoError(t, err)
	upgradeDe50(t, dir)
	depois, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "ci.yml"))
	require.NoError(t, err)
	assert.Equal(t, antes, depois, "byte a byte")
}

// TestUpgradeDe50NenhumHookSemScript: todo hook do forge no settings.json tem o script que ele executa.
func TestUpgradeDe50NenhumHookSemScript(t *testing.T) {
	dir := copiaLegado50(t)
	upgradeDe50(t, dir)
	for _, c := range comandosPreToolUse(t, dir) {
		if strings.Contains(c, "pr-guard") {
			assert.FileExists(t, PRGuardScriptPath(dir), "hook %q sem scripts/pr_guard.py", c)
		}
	}
	assert.NotEmpty(t, comandosPreToolUse(t, dir))
}
