package survey

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueDrilldown_CardDetailInspection(t *testing.T) {
	items := []queue.QueueItem{
		{
			TaskID:   "task-01",
			Title:    "Formatador Chat Feed",
			SpecPath: "sdd/features/feat-01.md",
			Status:   queue.StatusApproved,
			Rules:    []string{"ARQ-01", "COD-01"},
			GovernanceVerdict: &queue.Verdict{
				Veredito: "conforme",
				Resumo:   "Arquitetura e regras cumpridas.",
				Achados: []queue.Finding{
					{Arquivo: "internal/monitor/chat_feed.go"},
				},
			},
		},
		{
			TaskID:   "task-02",
			Title:    "Falha de Validação",
			SpecPath: "sdd/features/feat-02.md",
			Status:   queue.StatusChangesRequested,
			Rules:    []string{"COD-02"},
			GovernanceVerdict: &queue.Verdict{
				Veredito: "violacao",
				Resumo:   "go vet apontou shadowing.",
				Achados: []queue.Finding{
					{Arquivo: "cmd/forge-sdd/main.go"},
				},
			},
		},
	}

	q := &queue.Queue{Items: items}
	tui := NewQueueDrilldownTUI(".", q)

	// Inspeciona tarefa aprovada
	detail1, err := tui.InspectTaskDetails("task-01")
	require.NoError(t, err)
	assert.Equal(t, "task-01", detail1.TaskID)
	assert.Equal(t, "Formatador Chat Feed", detail1.Title)
	assert.Equal(t, "conforme", detail1.Verdict)
	assert.Equal(t, "Arquitetura e regras cumpridas.", detail1.ReviewerNote)
	assert.Contains(t, detail1.FilesChanged, "internal/monitor/chat_feed.go")
	assert.Equal(t, "1m15s", detail1.Duration)
	assert.Contains(t, detail1.CommitMsg, "feat(task-01): Formatador Chat Feed")

	// Inspeciona tarefa com falha
	detail2, err := tui.InspectTaskDetails("task-02")
	require.NoError(t, err)
	assert.Equal(t, "task-02", detail2.TaskID)
	assert.Equal(t, "violacao", detail2.Verdict)
	assert.Equal(t, "go vet apontou shadowing.", detail2.ReviewerNote)
	assert.Contains(t, detail2.FilesChanged, "cmd/forge-sdd/main.go")
	assert.NotEmpty(t, detail2.Duration)

	// Tarefa inexistente
	_, err = tui.InspectTaskDetails("task-99")
	assert.Error(t, err)
}

func TestQueueDrilldown_InspectDetails(t *testing.T) {
	view := &TaskDetailView{
		TaskID:       "fix-abcc-01",
		Title:        "Telas no frontend-admin",
		Status:       queue.StatusApproved,
		Duration:     "1m15s",
		FilesChanged: []string{"shared/ui/src/components/BottomNav.tsx"},
		CommitMsg:    "fix(fix-abcc-01): corrige navegação mobile",
		Verdict:      "conforme",
		ReviewerNote: "Todos os testes passaram e regras respeitadas.",
		Rules:        []string{"ARQ-01", "COD-01"},
		SpecPath:     "sdd/discovery/fix-abcc-01/criteria.md",
	}

	card := FormatTaskDetailCard(view)
	assert.Contains(t, card, "fix-abcc-01")
	assert.Contains(t, card, "Telas no frontend-admin")
	assert.Contains(t, card, "[✓] Concluído")
	assert.Contains(t, card, "1m15s")
	assert.Contains(t, card, "shared/ui/src/components/BottomNav.tsx")
	assert.Contains(t, card, "fix(fix-abcc-01): corrige navegação mobile")
	assert.Contains(t, card, "conforme")
	assert.Contains(t, card, "Todos os testes passaram e regras respeitadas.")

	// Garante que é uma moldura visual
	assert.True(t, strings.HasPrefix(card, "┌─── 📋 Detalhes da Tarefa"))
	assert.True(t, strings.HasSuffix(card, "┘"))
}
