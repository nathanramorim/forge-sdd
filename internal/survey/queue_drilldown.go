package survey

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/forge-sdd/cli/internal/queue"
)

// QueueDrilldownItem representa uma linha da lista navegável da fila.
type QueueDrilldownItem struct {
	TaskID string
	Title  string
	Status string
	Badge  string
	Label  string
}

// QueueDrilldownTUI encapsula o estado e a navegação da fila interativa.
type QueueDrilldownTUI struct {
	Dir    string
	Queue  *queue.Queue
	Items  []QueueDrilldownItem
	Height int
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

// NewQueueDrilldownTUI inicializa a TUI carregando 100% das tarefas da fila sem truncamento.
func NewQueueDrilldownTUI(dir string, q *queue.Queue) *QueueDrilldownTUI {
	if q == nil {
		q = &queue.Queue{Items: []queue.QueueItem{}}
	}

	drilldownItems := make([]QueueDrilldownItem, len(q.Items))
	for i, it := range q.Items {
		badge := StatusBadge(it.Status)
		title := it.Title
		if title == "" {
			title = it.SpecPath
		}
		label := fmt.Sprintf("%-16s %-20s %s", it.TaskID, badge, title)
		drilldownItems[i] = QueueDrilldownItem{
			TaskID: it.TaskID,
			Title:  title,
			Status: it.Status,
			Badge:  badge,
			Label:  label,
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
		Dir:    dir,
		Queue:  q,
		Items:  drilldownItems,
		Height: height,
	}
}

// TotalTasks retorna a contagem total de tarefas mapeadas na TUI.
func (tui *QueueDrilldownTUI) TotalTasks() int {
	return len(tui.Items)
}

// BuildOptions gera as opções de seleção para o componente do huh.
func (tui *QueueDrilldownTUI) BuildOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(tui.Items)+1)
	for _, it := range tui.Items {
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
		Title("Fila de Tarefas — Use ↑/↓ para navegar e Enter para inspecionar detalhes").
		Description(fmt.Sprintf("Total de tarefas: %d (sem truncamento) • Esc para sair", len(tui.Items))).
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
