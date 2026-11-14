package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Arquiva iniciativas finalizadas e compacta artefatos de progresso",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		initID, _ := cmd.Flags().GetString("initiative")
		asJSON, _ := cmd.Flags().GetBool("json")

		if initID == "" {
			if id, _, ok := layout.ActiveInitiative(dir); ok {
				initID = id
			}
		}

		if initID == "" {
			return fmt.Errorf("iniciativa não identificada; informe via --initiative <id> ou execute na branch da iniciativa")
		}

		res, err := pipeline.AutoArchiveInitiative(dir, initID, time.Now().UTC())
		if err != nil {
			return err
		}

		if asJSON {
			data, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(out, string(data))
			return nil
		}

		if !res.Archived {
			fmt.Fprintf(out, "Iniciativa %s não arquivada: %s\n", initID, res.Reason)
			return nil
		}

		fmt.Fprintf(out, "✓ Iniciativa %s arquivada com sucesso em %s\n", res.Initiative, res.ArchiveDir)
		if len(res.MovedFiles) > 0 {
			fmt.Fprintf(out, "  Artefatos consolidados: %d arquivos\n", len(res.MovedFiles))
		}
		if res.TelemetryPath != "" {
			fmt.Fprintf(out, "  Telemetria gravada em %s\n", res.TelemetryPath)
		}
		return nil
	},
}

func init() {
	archiveCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	archiveCmd.Flags().String("initiative", "", "ID da iniciativa para arquivar (ex: 81-tokens-por-janela-e-reset-suave)")
	archiveCmd.Flags().Bool("json", false, "Exibir resultado em formato JSON")
	rootCmd.AddCommand(archiveCmd)
}
