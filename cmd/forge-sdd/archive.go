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
		concluidas, _ := cmd.Flags().GetBool("concluidas")
		dry, _ := cmd.Flags().GetBool("dry-run")
		if dry && !concluidas {
			return fmt.Errorf("--dry-run vale só com --concluidas")
		}
		if concluidas {
			return archiveConcluidas(cmd, dir, dry)
		}

		if initID == "" {
			if id, _, ok := layout.ActiveInitiative(dir); ok {
				initID = id
			}
		}

		if initID == "" {
			return fmt.Errorf("iniciativa não identificada; informe via --initiative <id> ou execute na branch da iniciativa")
		}

		res, err := pipeline.AutoArchiveInitiative(dir, initID, time.Now().UTC(), nil)
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

// archiveConcluidas lista (--dry-run) ou arquiva as iniciativas sem task pendente. Cada arquivamento
// segue as mesmas guardas: fora da branch da iniciativa, recusa com o motivo.
func archiveConcluidas(cmd *cobra.Command, dir string, dry bool) error {
	out := cmd.OutOrStdout()
	candidates := pipeline.ListArchiveCandidates(dir)
	if len(candidates) == 0 {
		fmt.Fprintln(out, "Nenhuma iniciativa concluída para arquivar.")
		return nil
	}
	for _, id := range candidates {
		if dry {
			fmt.Fprintf(out, "candidata: %s\n", id)
			continue
		}
		res, err := pipeline.AutoArchiveInitiative(dir, id, time.Now().UTC(), nil)
		if err != nil {
			return err
		}
		if res.Archived {
			fmt.Fprintf(out, "✓ %s arquivada em %s\n", id, res.ArchiveDir)
		} else {
			fmt.Fprintf(out, "- %s não arquivada: %s\n", id, res.Reason)
		}
	}
	return nil
}

func init() {
	archiveCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	archiveCmd.Flags().String("initiative", "", "ID da iniciativa para arquivar (ex: 81-tokens-por-janela-e-reset-suave)")
	archiveCmd.Flags().Bool("concluidas", false, "Trata todas as iniciativas concluídas (sem task pendente)")
	archiveCmd.Flags().Bool("dry-run", false, "Com --concluidas, lista as candidatas sem mover")
	archiveCmd.Flags().Bool("json", false, "Exibir resultado em formato JSON")
	rootCmd.AddCommand(archiveCmd)
}
