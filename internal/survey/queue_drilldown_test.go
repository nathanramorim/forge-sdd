package survey

import (
	"fmt"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueDrilldown_ListAllTasksWithoutTruncation(t *testing.T) {
	// Cenário baseado no barbearia-app com 25 tarefas
	const totalTasks = 25
	items := make([]queue.QueueItem, totalTasks)
	for i := 0; i < totalTasks; i++ {
		st := queue.StatusQueued
		switch i % 5 {
		case 0:
			st = queue.StatusApproved
		case 1:
			st = queue.StatusBuilding
		case 2:
			st = queue.StatusReviewPending
		case 3:
			st = queue.StatusChangesRequested
		case 4:
			st = queue.StatusQueued
		}
		items[i] = queue.QueueItem{
			TaskID:   fmt.Sprintf("task-%02d", i+1),
			Title:    fmt.Sprintf("Implementação do componente %02d", i+1),
			SpecPath: fmt.Sprintf("sdd/features/feat-%02d.md", i+1),
			Status:   st,
		}
	}

	q := &queue.Queue{
		FeatureID: "barbearia-app",
		Items:     items,
	}

	tui := NewQueueDrilldownTUI(".", q)
	require.NotNil(t, tui)
	assert.Equal(t, totalTasks, tui.TotalTasks(), "deve carregar exatamente 25 tarefas sem truncamento")

	options := tui.BuildOptions()
	// totalTasks + 1 opção de saída
	assert.Equal(t, totalTasks+1, len(options), "todas as 25 opções mais opção de saída devem estar presentes")

	// Garante que não há indicador de truncamento (... e mais N tarefas)
	for _, opt := range options {
		assert.False(t, strings.Contains(opt.Key, "... e mais"), "nenhuma linha deve conter truncamento")
	}

	// As 20 tarefas Na Fila / Em Andamento vêm primeiro (separadas das 5 concluídas que ficam no final)
	require.Len(t, tui.ActiveItems, 20)
	require.Len(t, tui.CompletedItems, 5)
	assert.Equal(t, "task-02", options[0].Value, "primeira task ativa na fila vem no topo")
	assert.Contains(t, options[0].Key, "[FILA]")
	assert.Equal(t, "task-25", options[19].Value, "última task ativa antes do bloco de concluídas")
	assert.Equal(t, "task-01", options[20].Value, "primeira task concluída fica separada abaixo das ativas")
	assert.Contains(t, options[20].Key, "[CONCLUÍDA]")
	assert.Equal(t, "task-21", options[24].Value, "última task concluída")
}

func TestQueueDrilldown_ListAllTasks(t *testing.T) {
	items := []queue.QueueItem{
		{TaskID: "task-01", Title: "Setup", Status: queue.StatusApproved},
		{TaskID: "task-02", Title: "Build", Status: queue.StatusBuilding},
		{TaskID: "task-03", Title: "Review", Status: queue.StatusReviewPending},
		{TaskID: "task-04", Title: "Fix", Status: queue.StatusChangesRequested},
		{TaskID: "task-05", Title: "Deploy", Status: queue.StatusQueued},
	}
	q := &queue.Queue{Items: items}
	tui := NewQueueDrilldownTUI(".", q)

	require.Len(t, tui.Items, 5)
	assert.Equal(t, "[✓] Concluído", tui.Items[0].Badge)
	assert.Equal(t, "[▶] Em Andamento", tui.Items[1].Badge)
	assert.Equal(t, "[?] Revisão Pendente", tui.Items[2].Badge)
	assert.Equal(t, "[✗] Falhou", tui.Items[3].Badge)
	assert.Equal(t, "[⌛] Na Fila", tui.Items[4].Badge)
}
