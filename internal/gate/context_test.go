package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ctxConstitution = `# Constituição

- **ARQ-01:** Fronteiras entre cmd e internal.
- **DUP-02**: Reuse Check obrigatório.
- **SEG-04:** Sem segredos.
- **TST-01:** Critério executável.
- **GOV-04:** Rules sem copiar a Constituição.
`

func ctxProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", ctxConstitution)
	testutil.WriteFile(t, dir, ".agents/rules/sdd-arquitetura.md", "---\ndescription: arq\nglobs: [\"cmd/**\"]\nregras: [ARQ-01, DUP-02]\n---\ncorpo\n")
	testutil.WriteFile(t, dir, ".agents/rules/sdd-seguranca.md", "---\ndescription: seg\nregras: [SEG-04]\n---\ncorpo\n")
	testutil.WriteFile(t, dir, ".agents/skills/sdd-build/SKILL.md", "---\nname: sdd-build\ndescription: d\nstatus: approved\norigin: forge-sdd-core\nscope: core\n---\n# Skill\n")
	testutil.WriteFile(t, dir, ".agents/skills/minha-skill/SKILL.md", "---\nname: minha-skill\ndescription: d\nstatus: approved\norigin: manual\n---\n# Skill\n")
	return dir
}

func TestBuildContext_SoOsIDsCitadosERulesQueOsCobrem(t *testing.T) {
	dir := ctxProject(t)
	testutil.WriteFile(t, dir, "spec.md", "# Spec\n\nCita `ARQ-01` e `TST-01..05`. Usa a skill minha-skill.\n")

	pack, err := BuildContext(dir, "spec.md", nil, "build")
	require.NoError(t, err)

	var ids []string
	for _, id := range pack.IDs {
		ids = append(ids, id.ID)
	}
	assert.Equal(t, []string{"ARQ-01", "TST-01"}, ids, "só os IDs citados, sem o resto da Constitution")

	require.Len(t, pack.Rules, 1)
	assert.Equal(t, ".agents/rules/sdd-arquitetura.md", pack.Rules[0].Path)
	assert.Equal(t, []string{"ARQ-01"}, pack.Rules[0].Covers)

	assert.Equal(t, []string{".agents/skills/sdd-build/SKILL.md", ".agents/skills/minha-skill/SKILL.md"}, pack.Skills)

	out := pack.Render()
	assert.Contains(t, out, "- **ARQ-01:** Fronteiras entre cmd e internal.")
	assert.NotContains(t, out, "SEG-04", "ID não citado não deve aparecer")
	assert.NotContains(t, out, "sdd-seguranca.md")
}

func TestBuildContext_IDsExtrasDaFilaEFormatoAlternativoDeLinha(t *testing.T) {
	dir := ctxProject(t)
	testutil.WriteFile(t, dir, "spec.md", "# Spec sem IDs no texto\n")

	pack, err := BuildContext(dir, "spec.md", []string{"DUP-02", "SEG-04"}, "build")
	require.NoError(t, err)
	require.Len(t, pack.IDs, 2)
	assert.Equal(t, "DUP-02", pack.IDs[0].ID)
	assert.Equal(t, "Reuse Check obrigatório.", pack.IDs[0].Text, "formato **ID**: também é reconhecido")
	assert.Len(t, pack.Rules, 2)
}

func TestBuildContext_RulesCitadasPorNomeEstacaoDiferente(t *testing.T) {
	dir := ctxProject(t)
	testutil.WriteFile(t, dir, ".agents/skills/sdd-governanca/SKILL.md", "---\nname: sdd-governanca\ndescription: d\nstatus: approved\nscope: core\n---\n")
	testutil.WriteFile(t, dir, "spec.md", "# Spec\nCarregue `sdd-seguranca.md`.\n")

	pack, err := BuildContext(dir, "spec.md", nil, "governanca")
	require.NoError(t, err)
	require.Len(t, pack.Rules, 1)
	assert.Equal(t, ".agents/rules/sdd-seguranca.md", pack.Rules[0].Path)
	assert.Empty(t, pack.Rules[0].Covers)
	assert.Equal(t, []string{".agents/skills/sdd-governanca/SKILL.md"}, pack.Skills)
}

func TestBuildContext_Violacoes(t *testing.T) {
	dir := ctxProject(t)
	testutil.WriteFile(t, dir, "spec.md", "# Spec\n`ARQ-01` e `ARQ-99` e `SEG-77`.\n")

	_, err := BuildContext(dir, "spec.md", nil, "build")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ARQ-99")
	assert.Contains(t, err.Error(), "SEG-77")
	assert.NotContains(t, err.Error(), "ARQ-01,", "ID válido não entra na lista de ausentes")

	_, err = BuildContext(dir, "nao-existe.md", nil, "build")
	require.Error(t, err)

	_, err = BuildContext(dir, "spec.md", nil, "../etc")
	require.Error(t, err, "estação fora da lista é recusada")

	require.NoError(t, os.Remove(filepath.Join(dir, "sdd", "memory", "constitution.md")))
	_, err = BuildContext(dir, "spec.md", nil, "build")
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "constitution"))
}

func TestBuildContext_ConfinaOCaminhoDaSpecAoProjeto(t *testing.T) {
	dir := ctxProject(t)
	outside := t.TempDir()
	testutil.WriteFile(t, outside, "segredo.md", "SEG-04 conteudo de fora do projeto")
	testutil.WriteFile(t, dir, "spec.md", "# Spec\n")
	require.NoError(t, os.Symlink(filepath.Join(outside, "segredo.md"), filepath.Join(dir, "link.md")))

	for name, p := range map[string]string{
		"absoluto":       filepath.Join(outside, "segredo.md"),
		"com ..":         "../" + filepath.Base(outside) + "/segredo.md",
		"muitos ..":      "../../../../../../../../etc/hosts",
		"link simbólico": "link.md",
	} {
		_, err := BuildContext(dir, p, nil, "build")
		require.Error(t, err, name)
		assert.NotContains(t, err.Error(), "conteudo de fora", name)
	}

	// Caminho relativo legítimo, mesmo com ".." interno que permanece no projeto.
	testutil.WriteFile(t, dir, "sdd/discovery/70-x/features/a.md", "# Spec\n")
	_, err := BuildContext(dir, "sdd/discovery/70-x/../70-x/features/a.md", nil, "build")
	require.NoError(t, err)
}

func TestBuildContext_SkillDeNomeCurtoNaoCasaPrefixo(t *testing.T) {
	dir := ctxProject(t)
	testutil.WriteFile(t, dir, ".agents/skills/go/SKILL.md", "---\nname: go\ndescription: d\nstatus: approved\norigin: manual\n---\n")
	testutil.WriteFile(t, dir, "spec.md", "# Spec\nUsa golang e a minha-skill.\n")
	pack, err := BuildContext(dir, "spec.md", nil, "build")
	require.NoError(t, err)
	assert.NotContains(t, pack.Skills, ".agents/skills/go/SKILL.md", "'go' não deve casar dentro de 'golang'")
	assert.Contains(t, pack.Skills, ".agents/skills/minha-skill/SKILL.md")
}
