package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var dispatchCmd = &cobra.Command{
	Use:   "dispatch <feature>",
	Short: "Gate pré-build e despacho da feature para a sessão Builder (modo pipeline)",
	Long: `Roda o gate pré-build e, se passar, enfileira o job em .forge/runtime/queues/builder/pending.

O gate verifica (todas as pendências de uma vez): discovery, criteria e plan; ao menos uma task com
as cinco seções obrigatórias; IDs da Constituição, rules e ADRs citados existentes; e a evidência do
card (card-evidence.json na pasta da iniciativa, gravada pelo agente depois de consultar o MCP),
com provider suportado, feature correta e dentro do prazo. Sem forge.json (modo pipeline inativo)
o comando orienta forge-sdd pipeline init.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		evidence, _ := cmd.Flags().GetString("card-evidence")
		dry, _ := cmd.Flags().GetBool("dry-run")
		out := cmd.OutOrStdout()

		res, err := pipeline.Dispatch(dir, args[0], pipeline.DispatchOptions{EvidencePath: evidence, DryRun: dry})
		if err != nil {
			if errors.Is(err, pipeline.ErrNotPipelineMode) {
				return fmt.Errorf("%w (sem .forge/config/forge.json o /implementar segue o fluxo da 4.0.0-beta)", err)
			}
			return err
		}
		if res.Blocked() {
			fmt.Fprintf(out, "Despacho bloqueado: %d pendência(s)\n", len(res.Pending))
			for _, p := range res.Pending {
				fmt.Fprintf(out, "  - %s\n", p)
			}
			return fmt.Errorf("gate pré-build reprovado para %q; nada foi enfileirado", args[0])
		}
		if dry {
			fmt.Fprintf(out, "[DRY-RUN] gate aprovado para %s; criaria o job %s (nada foi gravado)\n", res.Initiative, res.Job.ID)
		} else if res.Created {
			fmt.Fprintf(out, "Job %s enfileirado em builder/pending (%s)\n", res.Job.ID, strings.TrimPrefix(res.JobPath, dir))
		} else {
			fmt.Fprintf(out, "Job %s já estava na fila (idempotente)\n", res.Job.ID)
		}
		if res.CardInstruction != "" {
			fmt.Fprintf(out, "Próximo passo do agente: %s\n", res.CardInstruction)
		} else {
			fmt.Fprintln(out, "Despacho sem card (pm_mcp.require_existing_card: false). O Builder já pode pegar o job; acompanhe pelo painel dele.")
		}
		return nil
	},
}

func init() {
	dispatchCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	dispatchCmd.Flags().String("card-evidence", "", "Caminho (relativo ao projeto) da evidência do card; padrão: <iniciativa>/card-evidence.json")
	dispatchCmd.Flags().Bool("dry-run", false, "Roda o gate e mostra o job sem gravar")
	rootCmd.AddCommand(dispatchCmd)
}
