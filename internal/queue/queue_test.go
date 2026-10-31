package queue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSpecContent = `# feat-58-04 — Motor da Fila

## Objetivo
Implementar a fila determinística.

## O que já existe (DUP-02)

| Procurei | Onde | Achei | Decisão | Justificativa |
| --- | --- | --- | --- | --- |
| Fila de tasks | internal/ | Não existe | criar | Novo pacote internal/queue |
`

func writeSpecFile(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	full := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	return relPath
}

func TestEnqueueAndLifecycle(t *testing.T) {
	dir := t.TempDir()
	specRel := writeSpecFile(t, dir, "sdd/features/feat-58/task-01.md", validSpecContent)

	err := Enqueue(dir, "feat-58", "feat/58-governanca", "task-01", "Motor da Fila", specRel, []string{"ARQ-03", "DUP-02"})
	require.NoError(t, err)

	q, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, "feat-58", q.FeatureID)
	assert.Equal(t, "feat/58-governanca", q.Branch)
	assert.Equal(t, "reviewed", q.SpecStatus)
	require.Len(t, q.Items, 1)
	assert.Equal(t, StatusQueued, q.Items[0].Status)
	assert.Equal(t, 0, q.Items[0].Attempts)

	// Next muda para building e incrementa Attempts
	item, err := Next(dir)
	require.NoError(t, err)
	assert.Equal(t, "task-01", item.TaskID)
	assert.Equal(t, StatusBuilding, item.Status)
	assert.Equal(t, 1, item.Attempts)

	// Complete muda para review_pending
	require.NoError(t, Complete(dir, "task-01"))
	q, err = Load(dir)
	require.NoError(t, err)
	assert.Equal(t, StatusReviewPending, q.Items[0].Status)
}

func TestEnqueueValidationFailures(t *testing.T) {
	dir := t.TempDir()

	// P-02: arquivo inexistente
	err := Enqueue(dir, "feat-58", "feat/58", "task-missing", "Missing", "sdd/features/nope.md", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "P-02")

	// P-03: sem seção "## O que já existe"
	noSectionSpec := writeSpecFile(t, dir, "sdd/features/no-section.md", "# Spec\n\n## Objetivo\nTeste.\n")
	err = Enqueue(dir, "feat-58", "feat/58", "task-no-sec", "No Section", noSectionSpec, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "P-03")

	// P-03: seção existe mas tabela sem linhas preenchidas
	emptyTableSpec := writeSpecFile(t, dir, "sdd/features/empty-table.md", `# Spec

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
| --- | --- | --- | --- | --- |
`)
	err = Enqueue(dir, "feat-58", "feat/58", "task-empty-table", "Empty Table", emptyTableSpec, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "P-03")
}

func TestReviewAndReEnqueueOnChangesRequested(t *testing.T) {
	dir := t.TempDir()
	specRel := writeSpecFile(t, dir, "sdd/features/feat-58/task-01.md", validSpecContent)

	require.NoError(t, Enqueue(dir, "feat-58", "feat/58", "task-01", "Task 01", specRel, []string{"ARQ-03"}))
	_, err := Next(dir)
	require.NoError(t, err)
	require.NoError(t, Complete(dir, "task-01"))

	// 1. Violação sem evidência deve falhar na validação da Camada B
	invalidViolationJSON := []byte(`{
		"veredito": "violacao",
		"achados": [
			{
				"verificacao": "S-01",
				"regra": "ARQ-03",
				"classificacao": "violacao",
				"severidade": "alta",
				"arquivo": "internal/queue/queue.go",
				"linha": 42,
				"evidencia": "",
				"explicacao": "Faltou evidência",
				"correcao": "Incluir trecho"
			}
		],
		"resumo": "Reprovado sem evidência"
	}`)
	_, err = Review(dir, "task-01", invalidViolationJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "evidencia")

	// 2. Violação válida muda status para changes_requested
	validViolationJSON := []byte(`{
		"veredito": "violacao",
		"achados": [
			{
				"verificacao": "S-01",
				"regra": "ARQ-03",
				"classificacao": "violacao",
				"severidade": "alta",
				"arquivo": "internal/queue/queue.go",
				"linha": 42,
				"evidencia": "db.QueryDirect()",
				"explicacao": "Acesso direto fora da camada",
				"correcao": "Usar repositório"
			}
		],
		"resumo": "1 violação encontrada"
	}`)
	item, err := Review(dir, "task-01", validViolationJSON)
	require.NoError(t, err)
	assert.Equal(t, StatusChangesRequested, item.Status)

	// Re-enfileirar task com changes_requested volta para queued
	require.NoError(t, Enqueue(dir, "feat-58", "feat/58", "task-01", "Task 01", specRel, []string{"ARQ-03"}))
	q, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, StatusQueued, q.Items[0].Status)

	// Next novamente incrementa Attempts para 2
	nextItem, err := Next(dir)
	require.NoError(t, err)
	assert.Equal(t, 2, nextItem.Attempts)
	require.NoError(t, Complete(dir, "task-01"))

	// 3. Veredito "duvida" mantém em review_pending
	duvidaJSON := []byte(`{
		"veredito": "duvida",
		"achados": [
			{
				"verificacao": "S-03",
				"regra": "DUP-02",
				"classificacao": "duvida",
				"severidade": "media",
				"arquivo": "internal/queue/queue.go",
				"linha": 10,
				"evidencia": "",
				"explicacao": "Dúvida sobre helper",
				"correcao": "Confirmar com Líder"
			}
		],
		"resumo": "Aguardando decisão do Líder"
	}`)
	item, err = Review(dir, "task-01", duvidaJSON)
	require.NoError(t, err)
	assert.Equal(t, StatusReviewPending, item.Status)

	// 4. Veredito "conforme" com candidate_skill move para approved
	conformeJSON := []byte(`{
		"veredito": "conforme",
		"achados": [],
		"candidate_skill": {
			"name": "go-json-strict-validator",
			"reason": "Repetido em 3 pacotes de validação",
			"proposed_path": ".agents/skills/go-json-strict-validator/SKILL.md"
		},
		"resumo": "Implementação 100% conforme"
	}`)
	item, err = Review(dir, "task-01", conformeJSON)
	require.NoError(t, err)
	assert.Equal(t, StatusApproved, item.Status)
	require.NotNil(t, item.GovernanceVerdict.CandidateSkill)
	assert.Equal(t, "go-json-strict-validator", item.GovernanceVerdict.CandidateSkill.Name)
}

func TestFormatExecutivePanel(t *testing.T) {
	q := &Queue{
		FeatureID: "feat-58",
		Branch:    "feat/58-governanca",
		Items: []QueueItem{
			{
				TaskID: "feat-58-01",
				Title:  "Constituição por IDs",
				Status: StatusApproved,
				GovernanceVerdict: &Verdict{
					Veredito: VerdictConforme,
					Resumo:   "Aprovado sem ressalvas",
					CandidateSkill: &CandidateSkill{
						Name:         "constitution-id-linter",
						Reason:       "Validação recorrente de IDs",
						ProposedPath: ".agents/skills/constitution-id-linter/SKILL.md",
					},
				},
			},
			{
				TaskID:   "feat-58-04",
				Title:    "Motor da Fila",
				Status:   StatusBuilding,
				Attempts: 1,
			},
			{
				TaskID: "feat-58-05",
				Title:  "Camada B e Painel",
				Status: StatusQueued,
			},
		},
	}

	panel := FormatExecutivePanel(q)
	assert.Contains(t, panel, "Concluído e Aprovado pela Governança (approved)")
	assert.Contains(t, panel, "Em Andamento / Em Revisão (building, review_pending, changes_requested)")
	assert.Contains(t, panel, "O que ainda não foi desenvolvido (queued)")
	assert.Contains(t, panel, "Auto-Skills Propostas (candidate_skill)")
	assert.Contains(t, panel, "feat-58-01")
	assert.Contains(t, panel, "feat-58-04")
	assert.Contains(t, panel, "feat-58-05")
	assert.Contains(t, panel, "constitution-id-linter")
}
