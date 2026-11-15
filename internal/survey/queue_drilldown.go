package survey

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/queue"
)

// QueueDrilldownItem representa uma linha da lista navegável da fila.
type QueueDrilldownItem struct {
	TaskID      string
	Title       string
	Status      string
	Badge       string
	Stage       string
	IsCompleted bool
	TotalTokens int
	Label       string
}

// QueueDrilldownTUI encapsula o estado e a navegação da fila interativa.
type QueueDrilldownTUI struct {
	Dir            string
	Queue          *queue.Queue
	Items          []QueueDrilldownItem
	ActiveItems    []QueueDrilldownItem
	CompletedItems []QueueDrilldownItem
	Height         int
}

// StatusBadge retorna a representação visual humanizada com ícone para cada status da task.
func StatusBadge(status string) string {
	switch strings.ToLower(status) {
	case queue.StatusApproved, "conforme", "done", "completed":
		return "[✓] Concluído"
	case queue.StatusBuilding, "processing", "in_progress":
		return "[▶] Em Andamento"
	case queue.StatusQueued, "pending":
		return "[⌛] Na Fila"
	case queue.StatusChangesRequested, "violacao", "failed", "error":
		return "[✗] Falhou"
	case queue.StatusReviewPending, "review":
		return "[?] Revisão Pendente"
	default:
		if status == "" {
			return "[⌛] Na Fila"
		}
		return fmt.Sprintf("[%s]", status)
	}
}

// PipelineStageBadge indica em qual etapa da pipeline (Discovery -> Builder -> Reviewer) a tarefa está.
func PipelineStageBadge(status string) string {
	switch strings.ToLower(status) {
	case queue.StatusApproved, "conforme", "done", "completed":
		return "Disc→Bld→Rev ✓"
	case queue.StatusReviewPending, "review":
		return "Etapa 3/3: Reviewer"
	case queue.StatusBuilding, "processing", "in_progress":
		return "Etapa 2/3: Builder"
	case queue.StatusChangesRequested, "violacao", "failed", "error":
		return "Etapa 2/3: Rework"
	default:
		return "Etapa 2/3: Fila Bld"
	}
}

// IsTaskCompleted retorna verdadeiro quando a tarefa já concluiu as 3 etapas (Discovery -> Builder -> Reviewer).
func IsTaskCompleted(status string) bool {
	switch strings.ToLower(status) {
	case queue.StatusApproved, "conforme", "done", "completed":
		return true
	default:
		return false
	}
}

// NewQueueDrilldownTUI inicializa a TUI carregando 100% das tarefas da fila sem truncamento,
// separando as tarefas que estão na fila/em andamento das que já foram concluídas.
func NewQueueDrilldownTUI(dir string, q *queue.Queue) *QueueDrilldownTUI {
	if q == nil {
		q = &queue.Queue{Items: []queue.QueueItem{}}
	}

	totalCount := len(q.Items)
	drilldownItems := make([]QueueDrilldownItem, totalCount)
	activeItems := make([]QueueDrilldownItem, 0, totalCount)
	completedItems := make([]QueueDrilldownItem, 0, totalCount)

	for i, it := range q.Items {
		badge := StatusBadge(it.Status)
		stage := PipelineStageBadge(it.Status)
		completed := IsTaskCompleted(it.Status)
		tokMetrics := monitor.CalculateTaskTokenMetrics(dir, it.TaskID, it.SpecPath, it.Status, totalCount)

		title := it.Title
		if title == "" {
			title = it.SpecPath
		}
		groupTag := "[FILA]"
		if completed {
			groupTag = "[CONCLUÍDA]"
		}
		label := fmt.Sprintf("%-11s %-16s %-20s [%-19s] (🪙 %s) %s",
			groupTag,
			it.TaskID,
			badge,
			stage,
			monitor.FormatTokensShort(tokMetrics.TotalTokens),
			title,
		)
		item := QueueDrilldownItem{
			TaskID:      it.TaskID,
			Title:       title,
			Status:      it.Status,
			Badge:       badge,
			Stage:       stage,
			IsCompleted: completed,
			TotalTokens: tokMetrics.TotalTokens,
			Label:       label,
		}
		drilldownItems[i] = item
		if completed {
			completedItems = append(completedItems, item)
		} else {
			activeItems = append(activeItems, item)
		}
	}

	height := len(drilldownItems) + 4
	if height < 10 {
		height = 10
	}
	if height > 24 {
		height = 24
	}

	return &QueueDrilldownTUI{
		Dir:            dir,
		Queue:          q,
		Items:          drilldownItems,
		ActiveItems:    activeItems,
		CompletedItems: completedItems,
		Height:         height,
	}
}

// TotalTasks retorna a contagem total de tarefas mapeadas na TUI.
func (tui *QueueDrilldownTUI) TotalTasks() int {
	return len(tui.Items)
}

// BuildOptions gera as opções de seleção para o componente do huh,
// posicionando primeiro as tarefas Na Fila / Em Andamento e separando abaixo as Concluídas.
func (tui *QueueDrilldownTUI) BuildOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(tui.Items)+1)
	for _, it := range tui.ActiveItems {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}
	for _, it := range tui.CompletedItems {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}
	options = append(options, huh.NewOption("← Voltar ao menu principal / Sair", "__exit__"))
	return options
}

// SelectTask executa o menu interativo com setas e Enter. Devolve o TaskID selecionado ou ErrBack.
func (tui *QueueDrilldownTUI) SelectTask() (string, error) {
	if len(tui.Items) == 0 {
		return "", fmt.Errorf("nenhuma tarefa na fila para exibir")
	}

	var choice string
	sel := huh.NewSelect[string]().
		Title("Esteira SDD: [1. Discovery] ──► [2. Builder] ──► [3. Reviewer]  (↑/↓ navegar • Enter detalhes)").
		Description(fmt.Sprintf("⏳ Na Fila / Em Andamento: %d  │  ✅ Concluídas: %d  │  Total: %d (sem corte) • Esc para sair", len(tui.ActiveItems), len(tui.CompletedItems), len(tui.Items))).
		Options(tui.BuildOptions()...).
		Height(tui.Height).
		Value(&choice)

	err := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(menuKeyMap()).Run()
	if errors.Is(err, huh.ErrUserAborted) || choice == "__exit__" {
		return "", ErrBack
	}
	if err != nil {
		return "", err
	}

	return choice, nil
}
