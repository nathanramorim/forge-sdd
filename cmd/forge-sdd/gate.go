package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/spf13/cobra"
)

var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Executa a Camada A Determinística do Portão de Governança (checks P-xx)",
	Long: `Valida determinísticamente as regras da Constituição e do fluxo SDD (P-01..P-23).
Em modo 'observacao', imprime o relatório e retorna exit 0.
Em modo 'bloqueio', retorna erro (exit != 0) caso alguma verificação falhe.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		specPath, _ := cmd.Flags().GetString("spec")
		branch, _ := cmd.Flags().GetString("branch")
		prBody, _ := cmd.Flags().GetString("pr-body")
		goModChanged, _ := cmd.Flags().GetBool("go-mod-changed")
		mode, _ := cmd.Flags().GetString("mode")
		asJSON, _ := cmd.Flags().GetBool("json")

		if mode == "" {
			mode = gate.ModeObservacao
		}
		if mode != gate.ModeObservacao && mode != gate.ModeBloqueio {
			return fmt.Errorf("--mode inválido %q: use %q ou %q", mode, gate.ModeObservacao, gate.ModeBloqueio)
		}

		results, err := gate.RunChecks(dir, gate.Options{
			Branch:       branch,
			SpecPath:     specPath,
			GoModChanged: goModChanged,
			PRBodyPath:   prBody,
			Now:          time.Now().UTC(),
		})
		if err != nil {
			return err
		}

		hasFailures := gate.HasFailures(results)

		if asJSON {
			payload := struct {
				Mode    string             `json:"mode"`
				Passed  bool               `json:"passed"`
				Results []gate.CheckResult `json:"results"`
			}{
				Mode:    mode,
				Passed:  !hasFailures,
				Results: results,
			}
			data, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Printf("🛡️  Portão de Governança (Camada A) — modo: %s\n\n", mode)
			for _, r := range results {
				icon := "✓"
				if !r.Passed {
					icon = "✗"
				}
				fmt.Printf("  %s [%s] (%s): %s\n", icon, r.ID, r.Rule, r.Evidence)
				if !r.Passed && r.Remediation != "" {
					fmt.Printf("      ↳ Correção: %s\n", r.Remediation)
				}
			}
			fmt.Println()
		}

		if hasFailures && mode == gate.ModeBloqueio {
			return fmt.Errorf("portão Camada A reprovado em modo %q", gate.ModeBloqueio)
		}
		return nil
	},
}

func init() {
	gateCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	gateCmd.Flags().String("spec", "", "Caminho para a Spec da feature/task para validar P-02/P-03/P-14")
	gateCmd.Flags().String("branch", "", "Nome da branch atual para validar P-01")
	gateCmd.Flags().String("pr-body", "", "Caminho para arquivo com o corpo do PR para validar P-20 (constitution-exception)")
	gateCmd.Flags().Bool("go-mod-changed", false, "Indica se go.mod foi alterado na branch (habilita P-14)")
	gateCmd.Flags().String("mode", gate.ModeObservacao, "Modo de operação do portão: observacao | bloqueio")
	gateCmd.Flags().Bool("json", false, "Exibir relatório em formato JSON")
}
