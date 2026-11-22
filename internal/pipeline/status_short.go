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
	return FormatShortStatusAll(dir, false)
}

// FormatShortStatusAll é o FormatShortStatus com a opção de incluir as iniciativas concluídas
// (`forge status --all`), que aparecem marcadas com "[concluída]".
func FormatShortStatusAll(dir string, all bool) (string, error) {
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
		if all || initiativeActive(it) {
			active = append(active, it)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Esteira — %s (%s)\n", nomeDoProjeto(abs), currentBranch(dir))
	if len(active) == 0 {
		b.WriteString("Nenhuma iniciativa ativa.\nPróximo passo: /discovery\n")
	} else {
		label := "Iniciativas ativas"
		if all {
			label = "Iniciativas"
		}
		fmt.Fprintf(&b, "%s: %d\n", label, len(active))
		for i, it := range active {
			if i >= shortMaxInitiatives {
				fmt.Fprintf(&b, "(+%d; use --full)\n", len(active)-i)
				break
			}
			marker := " "
			if it.Current {
				marker = "*"
			}
			tag := ""
			if IsConcluded(it.Tasks) {
				tag = " [concluída]"
			}
			fmt.Fprintf(&b, "%s %s%s\n", marker, clip(it.ID, shortLineMax), tag)
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

// initiativeActive: a iniciativa da branch atual sempre aparece; as demais somem quando a fila
// está toda concluída. Fila ausente ou vazia não é "concluída" e continua visível.
func initiativeActive(it layout.InitiativeInfo) bool {
	if it.Current {
		return true
	}
	return !IsConcluded(it.Tasks)
}

// IsConcluded é a regra pura: há itens e todos estão concluídos (completed, approved ou done).
func IsConcluded(tasks map[string]int) bool {
	total := 0
	for status, n := range tasks {
		if n <= 0 {
			continue
		}
		if status != "completed" && status != queue.StatusApproved && status != "done" {
			return false
		}
		total += n
	}
	return total > 0
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
