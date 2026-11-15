package monitor

import (
	"fmt"
	"strings"
	"time"
)

// EventType categoriza os eventos narrativos do stream dos workers.
type EventType string

const (
	EventStartTask     EventType = "start_task"
	EventReadSpec      EventType = "read_spec"
	EventEditFiles     EventType = "edit_files"
	EventRunValidation EventType = "run_validation"
	EventTaskSuccess   EventType = "task_success"
	EventTaskError     EventType = "task_error"
)

// ChatEvent representa um evento emitido durante o ciclo de vida do worker.
type ChatEvent struct {
	Type      EventType
	Role      string
	TaskID    string
	Title     string
	Step      string
	Files     []string
	Logs      []string
	Duration  time.Duration
	Message   string
	Error     error
	Timestamp time.Time
}

// ChatFeedFormatter formata eventos em balões/cards conversacionais legíveis para humanos.
type ChatFeedFormatter struct {
	NoColor   bool
	ASCIIOnly bool
	TimeFunc  func() time.Time
}

// NewChatFeedFormatter cria uma nova instância do formatador de chat narrativo.
func NewChatFeedFormatter(noColor, asciiOnly bool) *ChatFeedFormatter {
	return &ChatFeedFormatter{
		NoColor:   noColor,
		ASCIIOnly: asciiOnly,
		TimeFunc:  time.Now,
	}
}

// Badges para status humanos com versão Unicode e fallback ASCII seguro.
func (f *ChatFeedFormatter) badge(emoji, ascii string) string {
	if f.ASCIIOnly {
		return ascii
	}
	return emoji
}

// FormatEvent converte um ChatEvent em um bloco visual estruturado e humanizado.
func (f *ChatFeedFormatter) FormatEvent(ev ChatEvent) string {
	switch ev.Type {
	case EventStartTask:
		badge := f.badge("💭", "[?]")
		titlePart := ""
		if ev.Title != "" {
			titlePart = fmt.Sprintf(" (%s)", ev.Title)
		}
		msg := fmt.Sprintf("%s Analisando requisitos e critérios da task %s%s...", badge, ev.TaskID, titlePart)
		return f.wrapBalloon(ev.Role, msg)

	case EventReadSpec:
		badge := f.badge("📋", "[i]")
		detail := ev.Message
		if detail == "" && ev.TaskID != "" {
			detail = fmt.Sprintf("spec da task %s", ev.TaskID)
		}
		msg := fmt.Sprintf("%s Lendo %s...", badge, detail)
		return f.wrapBalloon(ev.Role, msg)

	case EventEditFiles:
		badge := f.badge("🔨", "[*]")
		filesStr := strings.Join(ev.Files, ", ")
		if len(ev.Files) > 3 {
			filesStr = fmt.Sprintf("%s e mais %d arquivos", strings.Join(ev.Files[:3], ", "), len(ev.Files)-3)
		}
		if filesStr == "" {
			filesStr = "arquivos do projeto"
		}
		msg := fmt.Sprintf("%s Editando %s...", badge, filesStr)
		return f.wrapBalloon(ev.Role, msg)

	case EventRunValidation:
		badge := f.badge("🧪", "[>]")
		stepName := ev.Step
		if stepName == "" {
			stepName = "testes automatizados"
		}
		msg := fmt.Sprintf("%s Executando suite de %s...", badge, stepName)
		return f.wrapBalloon(ev.Role, msg)

	case EventTaskSuccess:
		badge := f.badge("✅", "[OK]")
		durStr := ""
		if ev.Duration > 0 {
			durStr = fmt.Sprintf(" em %s", formatShortDuration(ev.Duration))
		}
		msg := fmt.Sprintf("%s Concluído! Task %s finalizada com sucesso%s.", badge, ev.TaskID, durStr)
		if ev.Message != "" {
			msg += "\n" + ev.indent(ev.Message, 4)
		}
		return f.wrapBalloon(ev.Role, msg)

	case EventTaskError:
		return f.FormatErrorCard(ev.Step, ev.Error, ev.Logs)

	default:
		return ev.Message
	}
}

// StartTask gera um balão para o início de análise da tarefa.
func (f *ChatFeedFormatter) StartTask(taskID, title string) string {
	return f.FormatEvent(ChatEvent{
		Type:   EventStartTask,
		TaskID: taskID,
		Title:  title,
	})
}

// ReadSpec gera um balão indicando leitura da spec.
func (f *ChatFeedFormatter) ReadSpec(specPath string) string {
	return f.FormatEvent(ChatEvent{
		Type:    EventReadSpec,
		Message: specPath,
	})
}

// EditFiles gera um balão de modificação de arquivos.
func (f *ChatFeedFormatter) EditFiles(files ...string) string {
	return f.FormatEvent(ChatEvent{
		Type:  EventEditFiles,
		Files: files,
	})
}

// RunValidation gera um balão indicando execução de testes/validações.
func (f *ChatFeedFormatter) RunValidation(step string) string {
	return f.FormatEvent(ChatEvent{
		Type: EventRunValidation,
		Step: step,
	})
}

// CompleteTask gera um balão de conclusão com sucesso.
func (f *ChatFeedFormatter) CompleteTask(taskID string, duration time.Duration, note string) string {
	return f.FormatEvent(ChatEvent{
		Type:     EventTaskSuccess,
		TaskID:   taskID,
		Duration: duration,
		Message:  note,
	})
}

// FormatErrorCard destaca a falha em bloco formatado com os últimos logs relevantes (até 5 linhas).
func (f *ChatFeedFormatter) FormatErrorCard(step string, err error, recentLogs []string) string {
	badge := f.badge("❌", "[FAIL]")
	if step == "" {
		step = "execução"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s Falha na etapa [%s]", badge, step))
	if err != nil {
		sb.WriteString(fmt.Sprintf(": %s", err.Error()))
	}
	sb.WriteString("\n")

	// Captura no máximo os últimos 5 logs relevantes
	filteredLogs := make([]string, 0, 5)
	start := 0
	if len(recentLogs) > 5 {
		start = len(recentLogs) - 5
	}
	for i := start; i < len(recentLogs); i++ {
		line := strings.TrimSpace(recentLogs[i])
		if line != "" {
			filteredLogs = append(filteredLogs, line)
		}
	}

	if len(filteredLogs) > 0 {
		sb.WriteString("    ┌─── Últimos logs ──────────────────────────────────────────\n")
		for _, logLine := range filteredLogs {
			sb.WriteString(fmt.Sprintf("    │ %s\n", logLine))
		}
		sb.WriteString("    └───────────────────────────────────────────────────────────")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func (f *ChatFeedFormatter) wrapBalloon(role, text string) string {
	var prefix string
	if role != "" {
		prefix = fmt.Sprintf("[%s] ", role)
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 1 {
		return prefix + lines[0]
	}

	var sb strings.Builder
	sb.WriteString(prefix + lines[0] + "\n")
	for _, l := range lines[1:] {
		sb.WriteString("  " + l + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (ev ChatEvent) indent(s string, spaces int) string {
	indentStr := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = indentStr + l
	}
	return strings.Join(lines, "\n")
}

func formatShortDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
}
