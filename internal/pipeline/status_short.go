package pipeline

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

const (
	// ShortStatusMaxBytes é o teto do `forge status` sem terminal (lido por agentes a cada comando).
	ShortStatusMaxBytes = 1500
	shortMaxInitiatives = 5
	shortLineMax        = 160
)

// FormatShortStatus devolve o status curto para agentes: só iniciativas ativas, a fila de cada uma
// e o próximo passo, em até ShortStatusMaxBytes e sem eventos do log.
func FormatShortStatus(dir string) (string, error) {
	items, err := layout.ListInitiatives(dir, false)
	if err != nil {
		return "", err
	}
	abs, aerr := filepath.Abs(dir)
	if aerr != nil {
		abs = dir
	}
	var active []layout.InitiativeInfo
	for _, it := range items {
		if initiativeActive(it) {
			active = append(active, it)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Esteira — %s (%s)\n", TmuxProjectName(abs), currentBranch(dir))
	if len(active) == 0 {
		b.WriteString("Nenhuma iniciativa ativa.\nPróximo passo: /discovery\n")
	} else {
		fmt.Fprintf(&b, "Iniciativas ativas: %d\n", len(active))
		for i, it := range active {
			if i >= shortMaxInitiatives {
				fmt.Fprintf(&b, "(+%d; use --full)\n", len(active)-i)
				break
			}
			marker := " "
			if it.Current {
				marker = "*"
			}
			fmt.Fprintf(&b, "%s %s\n", marker, clip(it.ID, shortLineMax))
			fmt.Fprintf(&b, "    fila: %s\n", clip(shortQueue(it.Tasks), shortLineMax))
			fmt.Fprintf(&b, "    próximo: %s\n", clip(shortNext(it.Next), shortLineMax))
		}
	}
	b.WriteString("Painel completo: forge status --full\n")
	out := b.String()
	if len(out) > ShortStatusMaxBytes {
		out = clipBytes(out, ShortStatusMaxBytes-1) + "\n"
	}
	return out, nil
}

func initiativeActive(it layout.InitiativeInfo) bool {
	if it.Current {
		return true
	}
	for status, n := range it.Tasks {
		if n > 0 && status != queue.StatusApproved && status != "done" {
			return true
		}
	}
	return false
}

func shortQueue(tasks map[string]int) string {
	if len(tasks) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(tasks))
	for k := range tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, tasks[k]))
	}
	return strings.Join(parts, " ")
}

func shortNext(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "…"
}

// clipBytes corta em n bytes sem partir um caractere UTF-8.
func clipBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
