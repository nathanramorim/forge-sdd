package main

import (
	"fmt"
	"time"

	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Conjunto de avaliação: mede o modo sessions contra o modo chat (sdd/eval/refs)",
}

var evalRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Valida sdd/eval/refs e grava sdd/.metrics/eval-<modo>.json com o simulador (sem modelo)",
	Long: `Valida as features de referência (critério com teste negativo, checagens do P-25) e agrega
taxa de aprovação, voltas do Retorno, tokens, tempo e intervenções humanas do modo escolhido.
Usa só o simulador determinístico (simulacao.json de cada referência): não chama modelo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		mode, _ := cmd.Flags().GetString("mode")
		res, path, err := monitor.RunEval(dir, mode, time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "modo %s: %d/%d aprovadas, %d voltas do Retorno, %d tokens, %ds, %d intervenções -> %s\n",
			res.Mode, res.Aprovadas, res.Features, res.VoltasRetorno, res.Tokens, res.Segundos, res.Intervencoes, path)
		return nil
	},
}

var evalCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Imprime a tabela comparativa de sdd/.metrics/eval-sessions.json e eval-chat.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		return monitor.CompareEval(dir, cmd.OutOrStdout())
	},
}

func init() {
	evalCmd.PersistentFlags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	evalRunCmd.Flags().String("mode", "", "Modo avaliado: sessions | chat")
	_ = evalRunCmd.MarkFlagRequired("mode")
	evalCmd.AddCommand(evalRunCmd, evalCompareCmd)
}
