package pipeline

import (
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/bus"
)

// HandoffCardParams contém as informações para gerar o Card de Handoff Compacto.
type HandoffCardParams struct {
	Initiative     string
	Branch         string
	PreviousTaskID string
	PreviousStatus string
	NextTaskID     string
	NextTaskTitle  string
	Scope          string
	ContextRules   []string
	TargetRole     string
}

// BuildHandoffCard gera o card compacto de handoff (< 800 tokens de entrada) conforme Protótipo 2 e CA-05.
func BuildHandoffCard(p HandoffCardParams) string {
	var b strings.Builder
	b.WriteString("────────────────────────────────────────────────────────────────────────\n")
	b.WriteString("▶ CONTINUAÇÃO AUTOMÁTICA DA ESTEIRA — HANDOFF DE CONTEXTO\n")
	b.WriteString("────────────────────────────────────────────────────────────────────────\n")

	iniciativa := p.Initiative
	if iniciativa == "" {
		iniciativa = "iniciativa-ativa"
	}
	fmt.Fprintf(&b, "• Iniciativa: %s\n", iniciativa)

	if p.Branch != "" {
		fmt.Fprintf(&b, "• Branch:     %s\n", p.Branch)
	}

	if p.PreviousTaskID != "" {
		prevStatus := p.PreviousStatus
		if prevStatus == "" {
			prevStatus = "concluída e arquivada com sucesso"
		}
		fmt.Fprintf(&b, "• Status:     Task %s %s.\n", p.PreviousTaskID, prevStatus)
	}

	if p.NextTaskID != "" {
		taskDesc := p.NextTaskID
		if p.NextTaskTitle != "" {
			taskDesc = fmt.Sprintf("Task %s (%s)", p.NextTaskID, p.NextTaskTitle)
		}
		fmt.Fprintf(&b, "• Próxima:    %s\n", taskDesc)
	}

	if p.Scope != "" {
		fmt.Fprintf(&b, "• Escopo:     %s\n", p.Scope)
	}

	if len(p.ContextRules) > 0 {
		fmt.Fprintf(&b, "• Contexto:   Carregados IDs %s\n", strings.Join(p.ContextRules, ", "))
	}

	target := p.TargetRole
	if target == "" {
		target = "Builder"
	}
	target = strings.ToUpper(target[:1]) + target[1:]

	nextID := p.NextTaskID
	if nextID == "" {
		nextID = "próxima task"
	} else {
		nextID = "a task " + nextID
	}
	fmt.Fprintf(&b, "\n%s, continue a implementação d%s a partir deste ponto.\n", target, nextID)

	return b.String()
}

// HandoffCardFromJob monta o Card de Handoff a partir de um bus.Job.
func HandoffCardFromJob(job bus.Job, prevTaskID, prevStatus string) string {
	rules := append([]string(nil), job.RulesRefs...)
	scope := job.Title
	if scope == "" {
		scope = job.Feature
	}
	return BuildHandoffCard(HandoffCardParams{
		Initiative:     job.Feature,
		Branch:         job.PR.Branch,
		PreviousTaskID: prevTaskID,
		PreviousStatus: prevStatus,
		NextTaskID:     job.ID,
		NextTaskTitle:  job.Title,
		Scope:          scope,
		ContextRules:   rules,
		TargetRole:     "Builder",
	})
}

// EstimateTokens estima o número de tokens do texto (aproximação 1 token ~= 4 caracteres).
func EstimateTokens(text string) int {
	words := len(strings.Fields(text))
	runes := len([]rune(text))
	tokChars := (runes + 3) / 4
	if tokChars > words {
		return tokChars
	}
	return words
}
