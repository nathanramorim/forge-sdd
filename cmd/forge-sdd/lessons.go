package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/spf13/cobra"
)

const (
	lessonsHeader      = "# Lições Aprendidas — forge-sdd\n\nPadrões de erro já corrigidos, consultados por Builder/Revisor antes de implementar (lido no READ-MIN, feat-01-04). Entradas mais recentes primeiro; o arquivo é aparado automaticamente para respeitar o orçamento.\n\n"
	lessonsBudgetBytes = 2 * 1024
)

var (
	ruleIDPattern     = regexp.MustCompile(`^[A-Z]{3}-\d{2}$`)
	ruleFinderPattern = regexp.MustCompile(`\b[A-Z]{3}-\d{2}\b`)
)

var lessonsCmd = &cobra.Command{
	Use:   "lessons",
	Short: "Gerenciamento dos artefatos de aprendizado em sdd/memory/licoes/<id>.md e sdd/memory/lessons.md",
}

var lessonsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Registra uma lição aprendida a partir de um fix aprovado, de forma determinística",
	Long: `Acrescenta uma entrada "padrão → correção (referência)" a sdd/memory/licoes/<RULE>.md
(quando --rule é informada) e a sdd/memory/lessons.md, aparando entradas antigas para manter
o orçamento de tamanho (2 KB). Pensado para ser chamado ao final de /revisar ou /novo-fix
quando outcome=approved e a causa raiz é significativa/recorrente — não depende de o LLM
montar o arquivo manualmente.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}
		pattern, _ := cmd.Flags().GetString("pattern")
		fix, _ := cmd.Flags().GetString("fix")
		ref, _ := cmd.Flags().GetString("ref")
		rule, _ := cmd.Flags().GetString("rule")

		if pattern == "" || fix == "" {
			return fmt.Errorf("--pattern e --fix são obrigatórios")
		}

		if strings.TrimSpace(rule) != "" {
			path, err := AppendRuleLesson(targetDir, rule, pattern, fix, ref)
			if err != nil {
				return err
			}
			_, _ = AppendLesson(targetDir, pattern, fix, ref)
			fmt.Printf("✓ Lição registrada em %s\n", path)
			return nil
		}

		path, err := AppendLesson(targetDir, pattern, fix, ref)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Lição registrada em %s\n", path)
		return nil
	},
}

func ruleLessonHeader(ruleID string) string {
	return fmt.Sprintf("# Lições Aprendidas — %s\n\nLições associadas à rule `%s` da Constituição (mais recentes primeiro).\n\n", ruleID, ruleID)
}

func validateRuleID(targetDir, ruleID string) (string, error) {
	rid := strings.TrimSpace(ruleID)
	if strings.Contains(rid, "..") || strings.ContainsAny(rid, `/\`) || !ruleIDPattern.MatchString(rid) {
		return "", fmt.Errorf("rule inválida: %s", rid)
	}
	constPath := layout.Constitution(targetDir)
	if data, err := os.ReadFile(constPath); err == nil {
		found := ruleFinderPattern.FindAllString(string(data), -1)
		valid := false
		for _, id := range found {
			if id == rid {
				valid = true
				break
			}
		}
		if !valid {
			return "", fmt.Errorf("rule %s inexistente na Constituição (%s)", rid, constPath)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("falha ao ler constituição %s: %w", constPath, err)
	}
	return rid, nil
}

// AppendRuleLesson insere uma nova entrada no topo de sdd/memory/licoes/<RULE>.md
// validando o ID da rule e respeitando lessonsBudgetBytes.
func AppendRuleLesson(targetDir, ruleID, pattern, fix, ref string) (string, error) {
	rid, err := validateRuleID(targetDir, ruleID)
	if err != nil {
		return "", err
	}
	licoesDir := filepath.Join(layout.Memory(targetDir), "licoes")
	if err := os.MkdirAll(licoesDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar sdd/memory/licoes: %w", err)
	}
	path := filepath.Join(licoesDir, rid+".md")
	header := ruleLessonHeader(rid)

	return appendLessonToFile(path, header, pattern, fix, ref)
}

// AppendLesson insere uma nova entrada no topo da lista de lições e apara
// entradas antigas para respeitar lessonsBudgetBytes.
func AppendLesson(targetDir, pattern, fix, ref string) (string, error) {
	memDir := layout.Memory(targetDir)
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar sdd/memory: %w", err)
	}
	path := filepath.Join(memDir, "lessons.md")
	return appendLessonToFile(path, lessonsHeader, pattern, fix, ref)
}

func appendLessonToFile(path, header, pattern, fix, ref string) (string, error) {
	existingEntries := []string{}
	if data, err := os.ReadFile(path); err == nil {
		existingEntries = parseLessonEntriesWithHeader(string(data), header)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("falha ao ler %s: %w", path, err)
	}

	entry := formatLessonEntry(pattern, fix, ref)
	entries := append([]string{entry}, existingEntries...)

	content := renderLessonsWithHeader(header, entries)
	for len(content) > lessonsBudgetBytes && len(entries) > 1 {
		entries = entries[:len(entries)-1]
		content = renderLessonsWithHeader(header, entries)
	}
	if len(content) > lessonsBudgetBytes {
		maxEntryLen := lessonsBudgetBytes - len(header) - 1
		entries[0] = truncateLessonEntry(entries[0], maxEntryLen)
		content = renderLessonsWithHeader(header, entries)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
	return path, nil
}

// truncateLessonEntry corta uma entrada que sozinha excede o orçamento do
// arquivo, garantindo que o arquivo nunca ultrapasse lessonsBudgetBytes
// mesmo quando um único pattern/fix/ref é maior que o limite.
func truncateLessonEntry(entry string, maxLen int) string {
	if maxLen <= 1 || len(entry) <= maxLen {
		return entry
	}
	return entry[:maxLen-1] + "…"
}

func formatLessonEntry(pattern, fix, ref string) string {
	if ref != "" {
		return fmt.Sprintf("- %s → %s (%s)", pattern, fix, ref)
	}
	return fmt.Sprintf("- %s → %s", pattern, fix)
}

func parseLessonEntries(content string) []string {
	return parseLessonEntriesWithHeader(content, lessonsHeader)
}

func parseLessonEntriesWithHeader(content, header string) []string {
	idx := strings.Index(content, header)
	body := content
	if idx == 0 {
		body = content[len(header):]
	}
	var entries []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "- ") {
			entries = append(entries, line)
		}
	}
	return entries
}

func renderLessons(entries []string) string {
	return renderLessonsWithHeader(lessonsHeader, entries)
}

func renderLessonsWithHeader(header string, entries []string) string {
	var b strings.Builder
	b.WriteString(header)
	for _, e := range entries {
		b.WriteString(e)
		b.WriteString("\n")
	}
	return b.String()
}

func init() {
	lessonsAddCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	lessonsAddCmd.Flags().String("rule", "", "ID da rule na Constituição (ex: COD-02, DUP-10)")
	lessonsAddCmd.Flags().String("pattern", "", "Padrão de erro identificado (obrigatório)")
	lessonsAddCmd.Flags().String("fix", "", "Correção aplicada (obrigatório)")
	lessonsAddCmd.Flags().String("ref", "", "Referência da feature/fix (ex: sdd/features/fix-50-....md)")

	lessonsCmd.AddCommand(lessonsAddCmd)
}
