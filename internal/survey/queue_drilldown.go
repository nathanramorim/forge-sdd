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

const (
	TabActive    = "active"
	TabCompleted = "completed"
)

// QueueDrilldownTUI encapsula o estado e a navegação da fila interativa separada em 2 abas.
type QueueDrilldownTUI struct {
	Dir            string
	Queue          *queue.Queue
	Items          []QueueDrilldownItem
	ActiveItems    []QueueDrilldownItem
	CompletedItems []QueueDrilldownItem
	ActiveTab      string
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
		label := fmt.Sprintf("%s  %s — %s  •  %s  •  🪙 %s",
			badge,
			it.TaskID,
			title,
			stage,
			monitor.FormatTokensShort(tokMetrics.TotalTokens),
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

	initialTab := TabActive
	if len(activeItems) == 0 && len(completedItems) > 0 {
		initialTab = TabCompleted
	}

	return &QueueDrilldownTUI{
		Dir:            dir,
		Queue:          q,
		Items:          drilldownItems,
		ActiveItems:    activeItems,
		CompletedItems: completedItems,
		ActiveTab:      initialTab,
		Height:         height,
	}
}

// TotalTasks retorna a contagem total de tarefas mapeadas na TUI.
func (tui *QueueDrilldownTUI) TotalTasks() int {
	return len(tui.Items)
}

// BuildOptions gera todas as opções (ativas primeiro, concluídas depois) sem truncamento.
func (tui *QueueDrilldownTUI) BuildOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(tui.Items)+1)
	for _, it := range tui.ActiveItems {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}
	for _, it := range tui.CompletedItems {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}
	options = append(options, huh.NewOption("← Fechar painel / Sair", "__exit__"))
	return options
}

// BuildTabOptions gera as opções apenas da aba selecionada (1. Na Fila / Em Andamento ou 2. Concluídas),
// incluindo a opção de alternar entre as 2 abas.
func (tui *QueueDrilldownTUI) BuildTabOptions(tab string) []huh.Option[string] {
	var current []QueueDrilldownItem
	if tab == TabCompleted {
		current = tui.CompletedItems
	} else {
		current = tui.ActiveItems
	}

	options := make([]huh.Option[string], 0, len(current)+2)
	for _, it := range current {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}

	if tab == TabCompleted {
		options = append(options, huh.NewOption(
			fmt.Sprintf("⇄ Ver Aba 1: ⏳ Na Fila / Em Andamento (%d tarefas)", len(tui.ActiveItems)),
			"__tab_active__",
		))
	} else {
		options = append(options, huh.NewOption(
			fmt.Sprintf("⇄ Ver Aba 2: ✅ Concluídas (%d tarefas)", len(tui.CompletedItems)),
			"__tab_completed__",
		))
	}

	options = append(options, huh.NewOption("← Fechar painel / Sair", "__exit__"))
	return options
}

// SelectTask executa o menu interativo separado nas 2 abas (Na Fila vs Concluídas). Devolve o TaskID selecionado ou ErrBack.
func (tui *QueueDrilldownTUI) SelectTask() (string, error) {
	if len(tui.Items) == 0 {
		return "", fmt.Errorf("nenhuma tarefa na fila para exibir")
	}

	for {
		tab := tui.ActiveTab
		if tab == "" {
			tab = TabActive
		}

		tabHeader := fmt.Sprintf("▸ [Aba 1: ⏳ Na Fila / Em Andamento (%d)]   │    Aba 2: ✅ Concluídas (%d)", len(tui.ActiveItems), len(tui.CompletedItems))
		if tab == TabCompleted {
			tabHeader = fmt.Sprintf("   Aba 1: ⏳ Na Fila / Em Andamento (%d)    │  ▸ [Aba 2: ✅ Concluídas (%d)]", len(tui.ActiveItems), len(tui.CompletedItems))
		}

		var choice string
		sel := huh.NewSelect[string]().
			Title("Esteira SDD: [1. Discovery] ──► [2. Builder] ──► [3. Reviewer]").
			Description(tabHeader + "  •  ↑/↓ navegar • Enter abrir • Esc sair").
			Options(tui.BuildTabOptions(tab)...).
			Height(tui.Height).
			Value(&choice)

		err := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(menuKeyMap()).Run()
		if errors.Is(err, huh.ErrUserAborted) || choice == "__exit__" {
			return "", ErrBack
		}
		if err != nil {
			return "", err
		}
		if choice == "__tab_completed__" {
			tui.ActiveTab = TabCompleted
			continue
		}
		if choice == "__tab_active__" {
			tui.ActiveTab = TabActive
			continue
		}

		return choice, nil
	}
}
