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
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra o status da esteira e iniciativas",
	Long: `Sem flags, imprime o status da esteira em modo snapshot (processamento,
tokens consumidos, progresso DAG, commits recentes e painel ativo).

Com --live (ou -l), abre o painel vivo contínuo da esteira (living dashboard), atualizado
a cada 2 segundos em tempo real.

Com --all-branches, inclui iniciativas de todas as branches locais e remotas.
Com --compact, imprime a versão tríplice enxuta (< 200 tokens) para agentes.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		compact, _ := cmd.Flags().GetBool("compact")
		if compact {
			out, err := pipeline.FormatLiveStatus(dir)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		}

		live, _ := cmd.Flags().GetBool("live")
		if live {
			once, _ := cmd.Flags().GetBool("once")
			if !isatty.IsTerminal(os.Stdout.Fd()) && !cmd.Flags().Changed("once") {
				once = true
			}
			return pipeline.RunLiveDashboard(cmd.Context(), dir, cmd.OutOrStdout(), once)
		}
		all, _ := cmd.Flags().GetBool("all-branches")
		initiatives, _ := cmd.Flags().GetBool("initiatives")
		if all || initiatives {
			items, err := layout.ListInitiatives(dir, all)
			if err != nil {
				return err
			}
			return renderStatus(cmd.OutOrStdout(), items)
		}

		// forge status padrão: snapshot do painel executivo rico da esteira
		dashboard, err := pipeline.FormatRichDashboard(dir, nil, nil, nil)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprint(cmd.OutOrStdout(), dashboard); err != nil {
			return err
		}

		// Lista complementar de iniciativas em sdd/discovery
		items, err := layout.ListInitiatives(dir, false)
		if err == nil {
			if len(items) > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "─── 📁 INICIATIVAS (sdd/discovery) ──────────────────────────────────────────")
			}
			if err := renderStatus(cmd.OutOrStdout(), items); err != nil {
				return err
			}
		}
		return nil
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
	statusCmd.Flags().Bool("initiatives", false, "Lista somente as iniciativas de sdd/discovery")
	statusCmd.Flags().BoolP("live", "l", false, "Exibe o painel vivo contínuo da esteira (living dashboard) em tempo real")
	statusCmd.Flags().Bool("once", false, "Executa uma única iteração do painel vivo e encerra")
	statusCmd.Flags().Bool("compact", false, "Exibe o status tríplice compacto (< 200 tokens) para prompts de IA")
	nextIDCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	nextIDCmd.Flags().String("kind", "discovery", "Tipo de item: discovery ou fix (a numeração é única)")
	nextIDCmd.Flags().String("workitem", "", "ID do workitem (obrigatório quando naming_convention=workitem)")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(nextIDCmd)
}
