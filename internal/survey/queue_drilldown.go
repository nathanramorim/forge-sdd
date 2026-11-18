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
	case queue.StatusBlocked:
		return "[■] Bloqueada"
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

// cleanTaskDisplay normaliza a exibição de TaskID e Title para nunca repetir o mesmo ID duas vezes na linha
// (ex: quando Title é "Feature: <task_id>") nem quebrar linha em terminais de 85%.
func cleanTaskDisplay(taskID, rawTitle string) string {
	t := strings.TrimSpace(rawTitle)
	tLower := strings.ToLower(t)
	idLower := strings.ToLower(strings.TrimSpace(taskID))

	for _, prefix := range []string{"feature:", "feat:", "fix:", "task:"} {
		if strings.HasPrefix(tLower, prefix) {
			trimmed := strings.TrimSpace(t[len(prefix):])
			if strings.EqualFold(trimmed, taskID) {
				t = ""
				break
			}
			t = trimmed
			tLower = strings.ToLower(t)
		}
	}

	if strings.EqualFold(t, taskID) || t == "" {
		return truncateRunes(taskID, 56)
	}
	if strings.HasPrefix(tLower, idLower) {
		rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t[len(taskID):]), "—"))
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "-"))
		if rest == "" {
			return truncateRunes(taskID, 56)
		}
		t = rest
	}
	combined := fmt.Sprintf("%s — %s", taskID, t)
	return truncateRunes(combined, 58)
}

func truncateRunes(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	if maxRunes <= 3 {
		return string(r[:maxRunes])
	}
	return string(r[:maxRunes-3]) + "..."
}

// padRightRunes preenche com espaços à direita até atingir width runas, alinhando colunas sem quebrar.
func padRightRunes(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
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
		display := padRightRunes(cleanTaskDisplay(it.TaskID, title), 50)
		tokStr := fmt.Sprintf("🪙 %4s tok", monitor.FormatTokensShort(tokMetrics.TotalTokens))

		var label string
		if completed {
			idx := len(completedItems) + 1
			label = fmt.Sprintf("✓  %02d. %s  │  %s", idx, display, tokStr)
		} else {
			idx := len(activeItems) + 1
			stageCol := padRightRunes(stage, 19)
			label = fmt.Sprintf("%s  %02d. [%s]  %s  │  %s", badge, idx, stageCol, display, tokStr)
		}
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
// colocando o botão de trocar de aba logo no TOPO da lista para ficar sempre visível sem precisar rolar até o fim.
func (tui *QueueDrilldownTUI) BuildTabOptions(tab string) []huh.Option[string] {
	var current []QueueDrilldownItem
	if tab == TabCompleted {
		current = tui.CompletedItems
	} else {
		current = tui.ActiveItems
	}

	options := make([]huh.Option[string], 0, len(current)+2)
	if tab == TabCompleted {
		options = append(options, huh.NewOption(
			fmt.Sprintf("⇄ [Trocar Aba] Ir para Aba 1: ⏳ Na Fila / Em Andamento (%d tarefas)", len(tui.ActiveItems)),
			"__tab_active__",
		))
	} else {
		options = append(options, huh.NewOption(
			fmt.Sprintf("⇄ [Trocar Aba] Ir para Aba 2: ✅ Concluídas (%d tarefas)", len(tui.CompletedItems)),
			"__tab_completed__",
		))
	}

	for _, it := range current {
		options = append(options, huh.NewOption(it.Label, it.TaskID))
	}

	options = append(options, huh.NewOption("← Fechar painel / Sair (Esc)", "__exit__"))
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

		var titleHeader, descHeader string
		if tab == TabCompleted {
			titleHeader = fmt.Sprintf("ABA ATIVA: ✅ CONCLUÍDAS (%d)   │   ⏳ Na Fila / Em Andamento (%d)", len(tui.CompletedItems), len(tui.ActiveItems))
			descHeader = "Esteira: [1. Discovery] ──► [2. Builder] ──► [3. Reviewer ✓]  •  ↑/↓ navegar • Enter ver detalhes • Esc sair"
		} else {
			titleHeader = fmt.Sprintf("ABA ATIVA: ⏳ NA FILA / EM ANDAMENTO (%d)   │   ✅ Concluídas (%d)", len(tui.ActiveItems), len(tui.CompletedItems))
			descHeader = "Esteira: [1. Discovery] ──► [2. Builder] ──► [3. Reviewer]  •  ↑/↓ navegar • Enter ver detalhes • Esc sair"
		}

		var choice string
		if tab == TabCompleted && len(tui.CompletedItems) > 0 {
			choice = tui.CompletedItems[0].TaskID
		} else if tab == TabActive && len(tui.ActiveItems) > 0 {
			choice = tui.ActiveItems[0].TaskID
		}

		sel := huh.NewSelect[string]().
			Title(titleHeader).
			Description(descHeader).
			Options(tui.BuildTabOptions(tab)...).
			Height(tui.Height).
			Value(&choice)

		err := huh.NewForm(huh.NewGroup(sel)).WithKeyMap(menuKeyMap()).WithAccessible(!isTTY()).Run()
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
