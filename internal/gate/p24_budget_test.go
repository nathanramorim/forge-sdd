package gate

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestCheckP24_ReprovaArquivosGeradosAcimaDoLimite(t *testing.T) {
	cases := map[string]int{
		".agents/skills/sdd-build/SKILL.md": MaxSkillBytes,
		".claude/commands/discovery.md":     MaxAdapterBytes,
		".gemini/prompts/discovery.md":      MaxAdapterBytes,
		".github/prompts/implementar.md":    MaxAdapterBytes,
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

// Arquivos que pertencem ao projeto do usuário nunca reprovam o P-24 (nem no modo bloqueio).
func TestCheckP24_ArquivosDoProjetoApenasAvisam(t *testing.T) {
	dir := setupGateDir(t)
	testutil.WriteFile(t, dir, "CLAUDE.md", strings.Repeat("a", MaxInstructionBytes*3))
	testutil.WriteFile(t, dir, ".github/copilot-instructions.md", strings.Repeat("a", MaxInstructionBytes+1))
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", ctxConstitution+strings.Repeat("a", MaxConstitutionSize))
	testutil.WriteFile(t, dir, ".claude/commands/meu-comando.md", strings.Repeat("a", MaxAdapterBytes*10))
	testutil.WriteFile(t, dir, ".claude/agents/meu-agente.md", strings.Repeat("a", MaxRoleAdapterBytes*10))
	testutil.WriteFile(t, dir, ".github/prompts/meu-prompt.prompt.md", strings.Repeat("a", MaxAdapterBytes*10))

	r := p24(t, dir)
	assert.True(t, r.Passed, r.Evidence)
	assert.Contains(t, r.Evidence, "aviso")
	assert.Contains(t, r.Evidence, "CLAUDE.md")
	assert.NotContains(t, r.Evidence, "meu-comando", "comando do próprio projeto fica fora do orçamento")
	assert.NotContains(t, r.Evidence, "meu-agente")
	assert.NotContains(t, r.Evidence, "meu-prompt")
}

func TestCheckP24_SkillsQueNaoSaoDeEstacaoNaoSaoMedidas(t *testing.T) {
	dir := setupGateDir(t)
	testutil.WriteFile(t, dir, ".agents/skills/minha-skill/SKILL.md", strings.Repeat("a", MaxSkillBytes*3))
	assert.True(t, p24(t, dir).Passed)
}

// O próprio repositório do forge-sdd precisa passar no portão que ele entrega.
func TestCheckP24_RepositorioDoForgeSddPassaNoProprioPortao(t *testing.T) {
	repo := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(repo, ".agents", "skills", "sdd-build", "SKILL.md")); err != nil {
		t.Skip("execução fora do repositório do forge-sdd")
	}
	r := p24(t, repo)
	assert.True(t, r.Passed, r.Evidence)
	assert.NotContains(t, r.Evidence, "aviso", "o repositório do forge-sdd não pode ter arquivo próprio acima do orçamento")
}

// Os 45 IDs canônicos não podem se perder em reescritas da Constitution.
func TestConstitutionDoRepositorioMantemTodosOsIDs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "sdd", "memory", "constitution.md"))
	if err != nil {
		t.Skip("execução fora do repositório do forge-sdd")
	}
	lines := ConstitutionIDLines(string(data))
	var want []string
	for _, p := range []struct {
		prefix string
		ids    []int
	}{
		{"ARQ", []int{1, 2, 3, 4, 5, 6, 10, 11}},
		{"SEG", []int{1, 2, 3, 4, 5, 6, 7}},
		{"DUP", []int{1, 2, 3, 4, 5, 6, 10}},
		{"COD", []int{1, 2, 3, 4, 5, 6}},
		{"TST", []int{1, 2, 3, 4, 5}},
		{"FLX", []int{1, 2, 3, 4, 5, 6, 7, 8}},
		{"GOV", []int{1, 2, 3, 4}},
	} {
		for _, n := range p.ids {
			want = append(want, fmt.Sprintf("%s-%02d", p.prefix, n))
		}
	}
	assert.Len(t, want, 45)
	for _, id := range want {
		assert.NotEmpty(t, lines[id], "ID %s ausente ou vazio na Constitution", id)
	}
	assert.Len(t, lines, len(want), "a Constitution não deve ganhar nem perder IDs sem ADR (GOV-01)")
	assert.LessOrEqual(t, len(data), MaxConstitutionSize)
}
