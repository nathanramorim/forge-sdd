package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules [ID...]",
	Short: "Lista os IDs da Constituição com suas descrições completas",
	Long: `Lista todos os IDs da Constituição (sdd/memory/constitution.md) com suas
descrições completas, sem truncar. Útil para consultar o significado de um ID
mencionado em vereditos do Reviewer ou saídas do forge brief.

Sem argumentos: lista todos os IDs ordenados.
Com argumentos: exibe apenas os IDs solicitados.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}

		constitutionPath := layout.Constitution(dir)
		data, err := os.ReadFile(constitutionPath)
		if err != nil {
			return fmt.Errorf("ler sdd/memory/constitution.md: %w", err)
		}

		allIDs := gate.ConstitutionIDLines(string(data))

		var ids []string
		if len(args) == 0 {
			for id := range allIDs {
				ids = append(ids, id)
			}
			sort.Strings(ids)
		} else {
			var missing []string
			for _, id := range args {
				upper := strings.ToUpper(id)
				if _, ok := allIDs[upper]; !ok {
					missing = append(missing, upper)
				} else {
					ids = append(ids, upper)
				}
			}
			if len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "ID(s) não encontrado(s) na Constituição: %s\n", strings.Join(missing, ", "))
				if len(ids) == 0 {
					return fmt.Errorf("nenhum ID válido encontrado")
				}
			}
			sort.Strings(ids)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		for _, id := range ids {
			fmt.Fprintf(w, "%s\t%s\n", id, allIDs[id])
		}
		return w.Flush()
	},
}

func init() {
	rulesCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	rootCmd.AddCommand(rulesCmd)
}
