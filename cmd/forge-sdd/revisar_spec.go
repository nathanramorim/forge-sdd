package main

import (
	"fmt"
	"time"

	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var revisarSpecCmd = &cobra.Command{
	Use:   "revisar-spec <iniciativa>",
	Short: "Manda a spec da iniciativa (discovery e criteria) ao Reviewer; o veredito volta pelo inbox",
	Long: `Fim do /discovery: enfileira a revisão da spec no Reviewer, em qualquer autonomia. Aprovada, o
inbox avisa e a pessoa decide; o /split-features quebra em tasks e despacha direto ao Builder.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		job, path, err := pipeline.RevisarDiscovery(dir, args[0], time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Spec enviada ao Reviewer: %s (%s)\nO veredito chega no inbox; acompanhe pelo painel.\n", job.ID, path)
		return nil
	},
}

func init() {
	revisarSpecCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	rootCmd.AddCommand(revisarSpecCmd)
}
