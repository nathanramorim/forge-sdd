package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

// askConsent pergunta à pessoa; variável para os testes. Só roda em terminal interativo.
var askConsent = func(text string) (bool, error) { return survey.ConfirmYesNo(text, false) }

// workerRunner executa ps/tmux na detecção de workers; variável para os testes (nil usa o sistema).
var workerRunner monitor.CommandRunner

// semWorkerAviso é o aviso de fila sem consumidor; só informa, nunca falha.
const semWorkerAviso = "nenhum worker ativo: as etapas ficam na fila; para executar neste chat, oriente os subagentes (veja a skill) ou inicie os workers"

var pipelineTocaAiCmd = &cobra.Command{
	Use:   "toca-ai <iniciativa>",
	Short: "Entrega à esteira as etapas planejadas pelo /toca-ai (um job por etapa, na ordem do DAG)",
	Long: `Confere o pré-voo (requisitos.json), o portão da spec e enfileira um job por etapa, em ordem de
dependência. Repetir o comando não duplica jobs. Com --preflight só lista o que falta da pessoa, de
uma vez, sem enfileirar nada. Com --block registra um imprevisto: bloqueia a etapa com o item exato.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		preflight, _ := cmd.Flags().GetBool("preflight")
		dry, _ := cmd.Flags().GetBool("dry-run")
		block, _ := cmd.Flags().GetString("block")
		reason, _ := cmd.Flags().GetString("reason")
		out := cmd.OutOrStdout()

		if block != "" {
			err := pipeline.ReportHumanPending(dir, args[0], []pipeline.HumanPending{{TaskID: block, Item: reason}}, notify.DefaultDeps(), time.Now())
			if err == nil {
				fmt.Fprintf(out, "etapa %s bloqueada: %s (lista em %s)\n", block, reason, pipeline.HumanPendingFile)
			}
			return err
		}
		res, err := pipeline.EnqueueGoalSteps(dir, args[0], pipeline.GoalOptions{DryRun: dry || preflight})
		if err != nil {
			if errors.Is(err, pipeline.ErrNotPipelineMode) {
				return fmt.Errorf("%w (o /toca-ai precisa do pipeline ativo)", err)
			}
			return err
		}
		if len(res.Pending) > 0 {
			fmt.Fprintf(out, "Faltam %d item(ns) da sua parte; nada foi enfileirado:\n", len(res.Pending))
			for _, p := range res.Pending {
				fmt.Fprintf(out, "  - %s\n", p)
			}
			return fmt.Errorf("pré-voo com pendências para %q", args[0])
		}
		if preflight {
			fmt.Fprintf(out, "Pré-voo limpo: %d etapa(s) prontas (%s). Nada foi enfileirado.\n", len(res.Steps), strings.Join(res.Steps, ", "))
			return nil
		}
		if res.Awaiting != "" {
			fmt.Fprintf(out, "%s\n", res.Awaiting)
			return nil
		}
		verb := "enfileiradas"
		if dry {
			verb = "enfileiraria"
		}
		fmt.Fprintf(out, "Base do PR: %s (fonte: %s)\n", res.BaseBranch, res.BaseSource)
		fmt.Fprintf(out, "Etapas %s: %d novas, %d já estavam na fila (ordem: %s)\n", verb, len(res.Created), len(res.Existing), strings.Join(res.Steps, " → "))
		if !dry {
			if n, err := monitor.ActiveWorkers(dir, workerRunner); err != nil || n == 0 {
				fmt.Fprintln(out, semWorkerAviso)
			}
		}
		return nil
	},
}

var pipelineConsentCmd = &cobra.Command{
	Use:   "consent <id>",
	Short: "Grava a sua resposta a um consentimento do requisitos.json (só a pessoa responde)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		initiative, _ := cmd.Flags().GetString("iniciativa")
		if !isInteractive() {
			return fmt.Errorf("o consentimento só vale com a resposta de uma pessoa em terminal interativo; um agente não pode concedê-lo")
		}
		req, err := pipeline.LoadRequirements(dir, initiative)
		if err != nil {
			return err
		}
		if req != nil {
			for _, c := range req.Consentimentos {
				if c.ID != args[0] {
					continue
				}
				ok, err := askConsent(c.Texto)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("consentimento %q não concedido; nada foi gravado", c.ID)
				}
				if err := pipeline.RecordConsent(dir, c.ID, c.Texto, time.Now()); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "consentimento %q gravado\n", c.ID)
				return nil
			}
		}
		return fmt.Errorf("consentimento %q não consta em requisitos.json da iniciativa %q", args[0], initiative)
	},
}

func init() {
	pipelineTocaAiCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineTocaAiCmd.Flags().Bool("preflight", false, "Só lista o que falta da pessoa, sem enfileirar")
	pipelineTocaAiCmd.Flags().Bool("dry-run", false, "Mostra o que faria sem gravar")
	pipelineTocaAiCmd.Flags().String("block", "", "ID da etapa a bloquear por imprevisto (use com --reason)")
	pipelineTocaAiCmd.Flags().String("reason", "", "Item exato que falta para a etapa bloqueada")
	pipelineConsentCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineConsentCmd.Flags().String("iniciativa", "", "Iniciativa do requisitos.json")
	_ = pipelineConsentCmd.MarkFlagRequired("iniciativa")
	pipelineCmd.AddCommand(pipelineTocaAiCmd, pipelineConsentCmd)
}
