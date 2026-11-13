package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra as iniciativas (discovery) e o estado de cada uma",
	Long: `Lista as iniciativas de sdd/discovery com branch, status, próximo passo e a contagem
de tasks da fila por situação. Substitui o index.md e o progress.md globais: o estado
vive na pasta de cada iniciativa e a visão geral é calculada na leitura.

Com --all-branches, lê o status.md e a fila de todas as branches locais e remotas
sem fazer checkout, para enxergar frentes paralelas ainda não mescladas.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		all, _ := cmd.Flags().GetBool("all-branches")

		items, err := layout.ListInitiatives(dir, all)
		if err != nil {
			return err
		}
		return renderStatus(cmd.OutOrStdout(), items)
	},
}

func renderStatus(w io.Writer, items []layout.InitiativeInfo) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "Nenhuma iniciativa encontrada em sdd/discovery.")
		return err
	}
	for _, it := range items {
		marker := " "
		if it.Current {
			marker = "*"
		}
		fmt.Fprintf(w, "%s %s\n", marker, it.ID)
		fmt.Fprintf(w, "    branch:  %s\n", it.Branch)
		fmt.Fprintf(w, "    status:  %s\n", orDash(it.Status))
		fmt.Fprintf(w, "    proximo: %s\n", orDash(it.Next))
		fmt.Fprintf(w, "    fila:    %s\n", formatTasks(it.Tasks))
	}
	return nil
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func formatTasks(tasks map[string]int) string {
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

var nextIDCmd = &cobra.Command{
	Use:   "next-id",
	Short: "Calcula o próximo ID de discovery ou fix sem colidir com outras branches",
	Long: `Devolve o próximo identificador conforme "naming_convention" de sdd/.sddrc
(sequencial, hash ou workitem), considerando a árvore de trabalho e todas as
branches locais e remotas, para duas frentes paralelas não escolherem o mesmo número.
Somente leitura: não cria arquivos nem branches.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		kind, _ := cmd.Flags().GetString("kind")
		if kind != "discovery" && kind != "fix" {
			return fmt.Errorf("--kind inválido %q: use discovery ou fix", kind)
		}
		workitem, _ := cmd.Flags().GetString("workitem")

		convention := "sequencial"
		rc, rcErr := config.ReadSddrc(dir)
		switch {
		case rcErr == nil:
			if rc.NamingConvention != "" {
				convention = rc.NamingConvention
			}
		case errors.Is(rcErr, os.ErrNotExist):
			// sem .sddrc: usa o padrão sequencial
		default:
			return rcErr // .sddrc inválido: não adivinhar a convenção
		}
		id, err := layout.NextID(dir, convention, workitem)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), id)
		return err
	},
}

func init() {
	statusCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	statusCmd.Flags().Bool("all-branches", false, "Inclui iniciativas de todas as branches locais e remotas, sem checkout")
	nextIDCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	nextIDCmd.Flags().String("kind", "discovery", "Tipo de item: discovery ou fix (a numeração é única)")
	nextIDCmd.Flags().String("workitem", "", "ID do workitem (obrigatório quando naming_convention=workitem)")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(nextIDCmd)
}
