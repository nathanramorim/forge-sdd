package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	lessonsHeader      = "# Lições Aprendidas — forge-sdd\n\nPadrões de erro já corrigidos, consultados por Builder/Revisor antes de implementar (lido no READ-MIN, feat-01-04). Entradas mais recentes primeiro; o arquivo é aparado automaticamente para respeitar o orçamento.\n\n"
	lessonsBudgetBytes = 2 * 1024
)

var lessonsCmd = &cobra.Command{
	Use:   "lessons",
	Short: "Gerenciamento do artefato de aprendizado sdd/memory/lessons.md",
}

var lessonsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Registra uma lição aprendida a partir de um fix aprovado, de forma determinística",
	Long: `Acrescenta uma entrada "padrão → correção (referência)" a sdd/memory/lessons.md,
aparando entradas antigas para manter o orçamento de tamanho (2 KB). Pensado para
ser chamado ao final de /revisar ou /novo-fix quando outcome=approved e a causa
raiz é significativa/recorrente — não depende de o LLM montar o arquivo manualmente.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}
		pattern, _ := cmd.Flags().GetString("pattern")
		fix, _ := cmd.Flags().GetString("fix")
		ref, _ := cmd.Flags().GetString("ref")

		if pattern == "" || fix == "" {
			return fmt.Errorf("--pattern e --fix são obrigatórios")
		}

		path, err := AppendLesson(targetDir, pattern, fix, ref)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Lição registrada em %s\n", path)
		return nil
	},
}

// AppendLesson insere uma nova entrada no topo da lista de lições e apara
// entradas antigas para respeitar lessonsBudgetBytes.
func AppendLesson(targetDir, pattern, fix, ref string) (string, error) {
	memDir := filepath.Join(targetDir, "sdd", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar sdd/memory: %w", err)
	}
	path := filepath.Join(memDir, "lessons.md")

	existingEntries := []string{}
	if data, err := os.ReadFile(path); err == nil {
		existingEntries = parseLessonEntries(string(data))
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("falha ao ler %s: %w", path, err)
	}

	entry := formatLessonEntry(pattern, fix, ref)
	entries := append([]string{entry}, existingEntries...)

	content := renderLessons(entries)
	for len(content) > lessonsBudgetBytes && len(entries) > 1 {
		entries = entries[:len(entries)-1]
		content = renderLessons(entries)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
	return path, nil
}

func formatLessonEntry(pattern, fix, ref string) string {
	if ref != "" {
		return fmt.Sprintf("- %s → %s (%s)", pattern, fix, ref)
	}
	return fmt.Sprintf("- %s → %s", pattern, fix)
}

func parseLessonEntries(content string) []string {
	idx := strings.Index(content, lessonsHeader)
	body := content
	if idx == 0 {
		body = content[len(lessonsHeader):]
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
	var b strings.Builder
	b.WriteString(lessonsHeader)
	for _, e := range entries {
		b.WriteString(e)
		b.WriteString("\n")
	}
	return b.String()
}

func init() {
	lessonsAddCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	lessonsAddCmd.Flags().String("pattern", "", "Padrão de erro identificado (obrigatório)")
	lessonsAddCmd.Flags().String("fix", "", "Correção aplicada (obrigatório)")
	lessonsAddCmd.Flags().String("ref", "", "Referência da feature/fix (ex: sdd/features/fix-50-....md)")

	lessonsCmd.AddCommand(lessonsAddCmd)
}
