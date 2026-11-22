package main

import (
	"encoding/json"
	"fmt"
	"os"
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
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			return fmt.Errorf("ler flag --dir: %w", err)
		}
		if dir == "" {
			dir = "."
		}

		constitutionPath := layout.Constitution(dir)
		data, err := os.ReadFile(constitutionPath)
		if err != nil {
			return fmt.Errorf("ler sdd/memory/constitution.md: %w", err)
		}

		allIDs := gate.ConstitutionIDLines(string(data))
		asJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return fmt.Errorf("ler flag --json: %w", err)
		}

		ids, ausentes, err := gate.SelecionaIDs(allIDs, args)
		if len(ausentes) > 0 {
			if _, werr := fmt.Fprintf(cmd.ErrOrStderr(), "ID(s) não encontrado(s) na Constituição: %s\n", strings.Join(ausentes, ", ")); werr != nil {
				return fmt.Errorf("escrever aviso de IDs ausentes: %w", werr)
			}
		}
		if err != nil {
			return fmt.Errorf("consultar IDs da Constituição: %w", err)
		}

		if asJSON {
			entries := make([]ruleEntry, 0, len(ids))
			for _, id := range ids {
				entries = append(entries, ruleEntry{ID: id, Description: allIDs[id]})
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(entries)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		for _, id := range ids {
			if _, werr := fmt.Fprintf(w, "%s\t%s\n", id, allIDs[id]); werr != nil {
				return fmt.Errorf("escrever saída de IDs: %w", werr)
			}
		}
		return w.Flush()
	},
}

type ruleEntry struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func init() {
	rulesCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	rulesCmd.Flags().Bool("json", false, "Exibir resultado em formato JSON")
	rootCmd.AddCommand(rulesCmd)
}
