package gate

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func p24(t *testing.T, dir string) CheckResult {
	t.Helper()
	results, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	r, ok := findCheck(results, "P-24")
	require.True(t, ok, "P-24 deve fazer parte do relatório")
	return r
}

func TestCheckP24_DentroDoOrcamentoPassa(t *testing.T) {
	dir := setupGateDir(t)
	testutil.WriteFile(t, dir, "CLAUDE.md", strings.Repeat("a", MaxInstructionBytes))
	testutil.WriteFile(t, dir, ".agents/skills/sdd-build/SKILL.md", strings.Repeat("a", MaxSkillBytes))
	testutil.WriteFile(t, dir, ".claude/commands/discovery.md", strings.Repeat("a", MaxAdapterBytes))
	testutil.WriteFile(t, dir, ".claude/agents/governance.md", strings.Repeat("a", MaxRoleAdapterBytes))

	r := p24(t, dir)
	assert.True(t, r.Passed, r.Evidence)
	assert.Contains(t, r.Evidence, "dentro do orçamento")
}

func TestCheckP24_ReprovaCadaGrupoAcimaDoLimite(t *testing.T) {
	cases := map[string]int{
		"CLAUDE.md":                         MaxInstructionBytes,
		".github/copilot-instructions.md":   MaxInstructionBytes,
		".agents/skills/sdd-build/SKILL.md": MaxSkillBytes,
		".claude/commands/discovery.md":     MaxAdapterBytes,
		".gemini/prompts/discovery.md":      MaxAdapterBytes,
		".github/prompts/discovery.md":      MaxAdapterBytes,
		".claude/agents/governance.md":      MaxRoleAdapterBytes,
		".github/chatmodes/builder.md":      MaxRoleAdapterBytes,
	}
	for rel, limit := range cases {
		dir := setupGateDir(t)
		testutil.WriteFile(t, dir, rel, strings.Repeat("a", limit+1))
		r := p24(t, dir)
		assert.False(t, r.Passed, rel)
		assert.Contains(t, r.Evidence, rel, "a evidência deve nomear o arquivo")
	}
}

func TestCheckP24_SkillsQueNaoSaoDeEstacaoNaoSaoMedidas(t *testing.T) {
	dir := setupGateDir(t)
	testutil.WriteFile(t, dir, ".agents/skills/minha-skill/SKILL.md", strings.Repeat("a", MaxSkillBytes*3))
	assert.True(t, p24(t, dir).Passed)
}
