package main

import (
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/spf13/cobra"
)

type exitCodeErr struct {
	code int
	msg  string
}

func (e exitCodeErr) Error() string { return e.msg }
func (e exitCodeErr) ExitCode() int { return e.code }

var briefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Imprime o resumo determinístico das decisões do projeto (Missão, Stack, Regras, ADRs)",
	Long: `Monta o resumo curto das decisões do projeto a partir de sdd/memory/constitution.md,
sdd/spec/ e das ADRs vigentes, respeitando o orçamento de contexto. Somente leitura.

Com --check, imprime apenas o estado do contexto (rascunho ou pronta) e o que falta,
saindo com código 3 em rascunho e 0 em pronta.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		maxBytes, _ := cmd.Flags().GetInt("max-bytes")
		check, _ := cmd.Flags().GetBool("check")

		b, err := gate.BuildBrief(dir, maxBytes)
		if err != nil {
			return err
		}

		if check {
			out := b.State
			if len(b.Missing) > 0 {
				out += " (faltam: " + strings.Join(b.Missing, ", ") + ")"
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			if b.State == gate.BriefRascunho {
				return exitCodeErr{code: 3, msg: ""}
			}
			return nil
		}

		fmt.Fprint(cmd.OutOrStdout(), b.Render())
		return nil
	},
}

func init() {
	briefCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	briefCmd.Flags().Int("max-bytes", 0, "Orçamento máximo em bytes (padrão: 6 KB)")
	briefCmd.Flags().Bool("check", false, "Confere se o contexto está pronto, saindo com 3 em rascunho e 0 em pronta")
	rootCmd.AddCommand(briefCmd)
}
