package gate

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func briefCompleteProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	const constText = `# Constituição — forge-demo

## Missão
Construir um CLI rápido e determinístico para orquestração de agentes.

## Stack
| Componente | Escolha |
|---|---|
| Go | 1.23 |
| CLI | Cobra |

## Regras
1. Arquitetura limpa entre cmd e internal.
2. Nenhuma escrita fora do workspace.

## IDs da Constituição
- **ARQ-01:** Fronteiras entre cmd e internal (ARQ-01).
- **SEG-01:** Sem comandos inseguros (SEG-01).
- **TST-01:** Critérios executáveis (TST-01).
`
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", constText)
	testutil.WriteFile(t, dir, "sdd/spec/modules.md", "# Módulos\n\n## Pipeline\nOrquestrador.\n\n## Gate\nPortão de governança.\n")
	testutil.WriteFile(t, dir, "sdd/spec/flows.md", "# Fluxos\n\n## Execução\nEntrada do job até o PR.\n\n## Revisão\nAnálise de conformidade.\n")
	testutil.WriteFile(t, dir, "sdd/spec/decisions.md", "# Decisões\n\n## Decisões Resolvidas\n| ID | Decisão | Status |\n|---|---|---|\n| D01 | Usar stdlib Go | resolvido |\n")
	testutil.WriteFile(t, dir, "sdd/spec/adr/adr-0001-armazenamento.md", "# ADR-0001: Armazenamento em arquivos JSON\n\n- **Status:** aceito\n\n## Decisão\nUsar arquivos JSON em disco.\n")
	testutil.WriteFile(t, dir, "sdd/spec/adr/adr-0002-banco-antigo.md", "# ADR-0002: Banco SQLite\n\n- **Status:** substituido\n\n## Decisão\nSubstituído pela ADR-0001.\n")
	testutil.WriteFile(t, dir, ".agents/rules/sdd-arquitetura.md", "---\ndescription: Regras de arquitetura\nregras: [ARQ-01]\n---\nCorpo da rule.\n")

	return dir
}

func TestBrief(t *testing.T) {
	dir := briefCompleteProject(t)

	b, err := BuildBrief(dir, 0)
	require.NoError(t, err)
	assert.Equal(t, "forge-demo", b.Project)
	assert.Equal(t, BriefPronta, b.State)
	assert.Empty(t, b.Missing)

	// Seções na ordem exata
	expectedLabels := []string{
		"Estado", "Missão", "Stack", "Regras", "IDs",
		"ADRs vigentes", "Módulos", "Fluxos", "Decisões", "Segurança", "Rules",
	}
	require.Len(t, b.Sections, len(expectedLabels))
	for i, s := range b.Sections {
		assert.Equal(t, expectedLabels[i], s.Label)
	}

	// ADR substituído não deve constar nas ADRs vigentes
	var adrItems []string
	for _, s := range b.Sections {
		if s.Label == "ADRs vigentes" {
			adrItems = s.Items
		}
	}
	require.Len(t, adrItems, 1)
	assert.Contains(t, adrItems[0], "ADR-0001")
	assert.NotContains(t, strings.Join(adrItems, " "), "ADR-0002")

	// Render fits within budget
	rendered := b.Render()
	assert.LessOrEqual(t, len(rendered), b.MaxBytes)
	assert.Contains(t, rendered, "# Contexto do projeto — forge-demo")
	assert.Contains(t, rendered, "Estado: pronta")

	// Teste com orçamento restrito para exercitar truncamento e "… N itens omitidos"
	smallBudget, err := BuildBrief(dir, 600)
	require.NoError(t, err)
	smallRendered := smallBudget.Render()
	assert.LessOrEqual(t, len(smallRendered), 600)
	assert.Contains(t, smallRendered, "omitidos; leia")
}

func TestBriefEstado(t *testing.T) {
	dir := t.TempDir()

	// Scaffold recém-gerado com placeholders
	const scaffoldConstitution = `# Constituição — projeto-modelo

## Missão
[Uma frase descrevendo o propósito do projeto]

## IDs da Constituição
- **ARQ-01:** Descrição modelo.
`
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", scaffoldConstitution)
	testutil.WriteFile(t, dir, "sdd/spec/modules.md", "# Módulos\n\n## <módulo-principal>\n[descrição]\n")
	testutil.WriteFile(t, dir, "sdd/spec/flows.md", "# Fluxos\n\n## [fluxo]\n[entrada]\n")
	testutil.WriteFile(t, dir, "sdd/spec/decisions.md", "# Decisões\n\n| ID | Decisão | Status |\n|---|---|---|\n")

	b, err := BuildBrief(dir, 0)
	require.NoError(t, err)
	assert.Equal(t, BriefRascunho, b.State)
	assert.Contains(t, b.Missing, "missão (constitution.md)")
	assert.Contains(t, b.Missing, "decisões (sdd/spec/decisions.md)")
	assert.Contains(t, b.Missing, "módulos (sdd/spec/modules.md)")
	assert.Contains(t, b.Missing, "fluxos (sdd/spec/flows.md)")

	// Preenche os campos que faltavam
	const filledConstitution = `# Constituição — projeto-real

## Missão
Entregar software determinístico com portão de governança.

## IDs da Constituição
- **ARQ-01:** Fronteiras.
`
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", filledConstitution)
	testutil.WriteFile(t, dir, "sdd/spec/modules.md", "# Módulos\n\n## Core\nMotor de regras.\n")
	testutil.WriteFile(t, dir, "sdd/spec/flows.md", "# Fluxos\n\n## Pipeline\nExecução de tarefas.\n")
	testutil.WriteFile(t, dir, "sdd/spec/decisions.md", "# Decisões\n\n## Decisões Resolvidas\n| ID | Decisão | Status |\n|---|---|---|\n| D01 | Arquitetura limpa | resolvido |\n")

	b2, err := BuildBrief(dir, 0)
	require.NoError(t, err)
	assert.Equal(t, BriefPronta, b2.State)
	assert.Empty(t, b2.Missing)
}
