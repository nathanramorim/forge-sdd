package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker de fila do pipeline: reivindica jobs e executa o agente (builder ou reviewer)",
	Long: `Laço de longa duração para uma sessão do pipeline. Mostra "Aguardando jobs na fila..." quando
ocioso; para cada job reivindicado invoca o agente de forma não interativa com o perfil da CLI de
forge.json (claude, gemini, copilot, codex ou agy), do agente da sessão (sessions) ou de --cli, renova o lease, grava o log em .forge/runtime/logs e fecha o job.

Builder: exige que o agente registre o PR (forge pipeline pr-opened); então enfileira a revisão.
Reviewer: exige que o agente aplique o veredito (forge bus review).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		checkStaleWorktrees(dir, cmd.OutOrStdout())
		role, _ := cmd.Flags().GetString("role")
		name, _ := cmd.Flags().GetString("name")
		once, _ := cmd.Flags().GetBool("once")
		poll, _ := cmd.Flags().GetDuration("poll")
		cli, _ := cmd.Flags().GetString("cli")
		raw, _ := cmd.Flags().GetBool("raw")
		deliver, _ := cmd.Flags().GetString("deliver")

		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return pipeline.RunWorker(ctx, pipeline.WorkerOptions{
			Dir: dir, Role: role, Name: name, CLI: cli, Raw: raw, Once: once, Poll: poll,
			DeliverTarget: deliver,
			Out:           cmd.OutOrStdout(),
			OnJobDone:     recordJobTelemetry(dir, role),
			FeedFormatter: WorkerFeedFormatter(),
		})
	},
}

// WorkerFeedFormatter cria o formatador padrão de feed narrativo para o worker.
func WorkerFeedFormatter() *monitor.ChatFeedFormatter {
	return monitor.NewChatFeedFormatter(false, false)
}

// recordJobTelemetry registra uma sessão por job, chamando o próprio binário com argv explícito.
func recordJobTelemetry(dir, role string) func(job bus.Job, outcome string) {
	return func(job bus.Job, outcome string) {
		rc, err := config.ReadSddrc(dir)
		if err != nil || !rc.Telemetry.Enabled {
			return
		}
		self, err := os.Executable()
		if err != nil {
			return
		}
		if outcome == pipeline.OutcomeDeferred {
			return // o job voltou à fila sem gastar tentativa: não é um resultado da sessão
		}
		u := monitor.ReadJobUsage(dir, role, job.ID, job.SpecPath)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, self, sessionRecordArgs(dir, role, job, taskIDDoJob(job), outcome, u)...).Run()
	}
}

// tipoTaskBuild é o tipo de job por task (bus.TypeTaskBuild, criado pela feat-122-01).
const tipoTaskBuild = "task_build"

// taskIDDoJob devolve o id da task de um job por task; vazio nos demais jobs.
func taskIDDoJob(job bus.Job) string {
	if job.Type == tipoTaskBuild && len(job.TaskIDs) == 1 {
		return job.TaskIDs[0]
	}
	return ""
}

// sessionRecordArgs monta o argv de `forge session record` de um job. Job por task (taskID não vazio)
// gera um registro por task e papel sem campo novo no schema 1.0: `feature` é o caminho da spec da
// task e `phase` é o papel; `medicao` é "real" quando o log tem evento `result` (FLX-04).
func sessionRecordArgs(dir, role string, job bus.Job, taskID, outcome string, u monitor.JobUsage) []string {
	ok := outcome == pipeline.OutcomeDone
	result := "approved"
	if !ok {
		result = "blocked"
	}
	phase := job.ID
	if taskID != "" {
		phase = role
	}
	return []string{"session", "record", "--dir", dir, "--feature", job.SpecPath, "--phase", phase,
		"--agent-path", role, "--outcome", result, fmt.Sprintf("--criterio-atendido=%t", ok),
		"--tokens-input", strconv.Itoa(u.TokensInput),
		"--tokens-output", strconv.Itoa(u.TokensOutput),
		"--tokens-cache-read", strconv.Itoa(u.TokensCacheRead),
		"--tokens-cache-creation", strconv.Itoa(u.TokensCacheCreation),
		"--turns", strconv.Itoa(u.Turns),
		"--duration-seconds", strconv.Itoa(u.DurationSeconds),
		"--model", u.Model,
		"--cost-usd-tabela", strconv.FormatFloat(u.CostUSDTabela, 'f', 4, 64),
		"--medicao", u.Medicao,
	}
}

var pipelinePROpenedCmd = &cobra.Command{
	Use:   "pr-opened",
	Short: "O Builder registra o PR aberto no job em processamento",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		id, _ := cmd.Flags().GetString("id")
		worker := workerFromFlags(cmd)
		number, _ := cmd.Flags().GetInt("number")
		url, _ := cmd.Flags().GetString("url")
		branch, _ := cmd.Flags().GetString("branch")
		entry, err := bus.UpdatePR(dir, layout.RoleBuilder, id, worker, number, url, branch, time.Now())
		if err != nil {
			return err
		}
		// O PR aberto entrega as tasks do job à revisão: a fila e o painel passam a refletir isso.
		if _, err := queue.SetTaskStatuses(dir, entry.Job.Initiative, entry.Job.TaskIDs, queue.StatusReviewPending); err != nil {
			return fmt.Errorf("atualizar a fila da iniciativa: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "PR #%d registrado no job %s\n", number, id)
		return nil
	},
}

var busPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Planeja as ondas de execução paralela das tasks de um job (por **Depende de:**)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		id, _ := cmd.Flags().GetString("id")
		job, _, err := bus.Get(dir, role, id)
		if err != nil {
			return err
		}
		cfg, _, err := config.LoadForgeConfig(dir)
		if err != nil {
			return err
		}
		if job.Initiative == "" {
			return fmt.Errorf("o job %s não tem iniciativa; não há tasks para planejar", id)
		}
		base, err := layout.InitiativeFeatures(dir, job.Initiative)
		if err != nil {
			return err
		}
		batches, err := pipeline.PlanWaves(base, cfg.Builder.LarguraDePlano())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "até %d tarefas em paralelo (use_git_worktrees=%t)\n%s", cfg.Builder.LarguraDePlano(), cfg.Builder.UseGitWorktrees, pipeline.FormatWaves(batches))
		return nil
	},
}

func init() {
	workerCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	workerCmd.Flags().String("role", "", "Papel do worker: builder ou reviewer")
	_ = workerCmd.MarkFlagRequired("role")
	workerCmd.Flags().String("name", "", "Nome da sessão (padrão: forge-<papel>)")
	workerCmd.Flags().String("cli", "", "Agente desta sessão: claude, gemini, copilot ou codex (padrão: sessions de forge.json)")
	workerCmd.Flags().String("deliver", "", "Alvo tmux (=sessao:janela.painel) para entregar prompts no modo live")
	workerCmd.Flags().Bool("raw", false, "Mostra a saída bruta do agente em vez da visão limpa (sem máscara de segredos; o log sempre guarda a bruta)")
	workerCmd.Flags().Bool("once", false, "Sai quando a fila esvazia")
	workerCmd.Flags().Duration("poll", 5*time.Second, "Intervalo de consulta da fila")
	rootCmd.AddCommand(workerCmd)

	pipelinePROpenedCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelinePROpenedCmd.Flags().String("id", "", "ID do job em processamento")
	_ = pipelinePROpenedCmd.MarkFlagRequired("id")
	pipelinePROpenedCmd.Flags().String("worker", "", "Worker dono do lease (padrão: variável FORGE_WORKER, definida pelo worker)")
	pipelinePROpenedCmd.Flags().Int("number", 0, "Número do PR")
	_ = pipelinePROpenedCmd.MarkFlagRequired("number")
	pipelinePROpenedCmd.Flags().String("url", "", "URL do PR")
	_ = pipelinePROpenedCmd.MarkFlagRequired("url")
	pipelinePROpenedCmd.Flags().String("branch", "", "Branch do PR (opcional)")
	pipelineCmd.AddCommand(pipelinePROpenedCmd)

	busFlags(busPlanCmd, true)
	busPlanCmd.Flags().String("id", "", "ID do job")
	_ = busPlanCmd.MarkFlagRequired("id")
	busCmd.AddCommand(busPlanCmd)
}

// workerFromFlags devolve o worker dono do lease: --worker ou, quando o agente roda sob o worker do
// Forge, a variável FORGE_WORKER (SEG-01: o dono do lease é sempre exigido).
func workerFromFlags(cmd *cobra.Command) string {
	if w, _ := cmd.Flags().GetString("worker"); w != "" {
		return w
	}
	return os.Getenv("FORGE_WORKER")
}

var pipelineWorktreeCmd = &cobra.Command{
	Use:   "worktree <add|remove>",
	Short: "Cria ou remove o worktree isolado de uma tarefa do Builder",
	Long: `Cria .forge/runtime/worktrees/<job>/<task> em uma branch forge/<job>-<task> (add) ou remove
o worktree e a branch auxiliar (remove). Usado pelo Builder quando builder.use_git_worktrees está ligado.`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"add", "remove"},
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		id, _ := cmd.Flags().GetString("id")
		task, _ := cmd.Flags().GetString("task")
		base, _ := cmd.Flags().GetString("base")
		switch args[0] {
		case "add":
			if base == "" {
				base = "HEAD"
			}
			path, err := pipeline.CreateWorktree(dir, id, task, base)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		case "remove":
			return pipeline.RemoveWorktree(dir, id, task)
		}
		return fmt.Errorf("ação %q inválida (use add ou remove)", args[0])
	},
}

func init() {
	pipelineWorktreeCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineWorktreeCmd.Flags().String("id", "", "ID do job")
	_ = pipelineWorktreeCmd.MarkFlagRequired("id")
	pipelineWorktreeCmd.Flags().String("task", "", "ID da tarefa")
	_ = pipelineWorktreeCmd.MarkFlagRequired("task")
	pipelineWorktreeCmd.Flags().String("base", "", "Ref base (padrão HEAD)")
	pipelineCmd.AddCommand(pipelineWorktreeCmd)
}

// gcAskConfirm pede confirmação antes de remover worktrees; pode ser substituído em testes.
var gcAskConfirm = func(text string) (bool, error) { return survey.ConfirmYesNo(text, false) }

var pipelineWorktreeGcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Lista e remove worktrees ociosos (jobs concluídos/aprovados ou orphans)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		dry, _ := cmd.Flags().GetBool("dry-run")
		out := cmd.OutOrStdout()

		stale, err := pipeline.ListStaleWorktrees(dir)
		if err != nil {
			return err
		}
		if len(stale) == 0 {
			fmt.Fprintln(out, "nenhum worktree ocioso encontrado.")
			return nil
		}

		fmt.Fprintf(out, "%d worktree(s) encontrado(s):\n", len(stale))
		var safe []pipeline.StaleWorktree
		for _, w := range stale {
			status := "protegido"
			if w.Safe {
				status = "pode remover"
				safe = append(safe, w)
			}
			fmt.Fprintf(out, "  [%s] %s/%s — %s\n", status, w.JobID, w.TaskID, w.Reason)
		}

		if len(safe) == 0 {
			fmt.Fprintln(out, "nenhum worktree seguro para remover.")
			return nil
		}

		if dry {
			fmt.Fprintf(out, "[dry-run] removeria %d worktree(s).\n", len(safe))
			return nil
		}

		ok, err := gcAskConfirm(fmt.Sprintf("Remover %d worktree(s)?", len(safe)))
		if err != nil || !ok {
			fmt.Fprintln(out, "cancelado.")
			return nil
		}

		errs := pipeline.SweepWorktrees(dir, safe)
		fmt.Fprintf(out, "%d worktree(s) removido(s).\n", len(safe)-len(errs))
		for _, e := range errs {
			fmt.Fprintf(out, "  erro: %s\n", e)
		}
		return nil
	},
}

func init() {
	pipelineWorktreeGcCmd.Flags().String("dir", ".", "Diretório alvo")
	pipelineWorktreeGcCmd.Flags().Bool("dry-run", false, "Lista sem remover")
	pipelineWorktreeCmd.AddCommand(pipelineWorktreeGcCmd)
}

var pipelinePRCmd = &cobra.Command{
	Use:   "pr",
	Short: "Abre ou corrige o PR do job: Camada A, push, PR na base da iniciativa e registro (nunca mescla)",
	Long: `Faz pelo Go o passo mecânico do Builder: roda a Camada A (forge gate), faz git push origin
<branch>:<branch>, cria o PR com a base da iniciativa (ou corrige a base de um PR já aberto), registra o
PR no job e marca as tasks como review_pending. Recusa branch main ou igual à base, corpo sem as
seções de .github/pull_request_template.md e branch sem commit à frente da base.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		id, _ := cmd.Flags().GetString("id")
		body, _ := cmd.Flags().GetString("body-file")
		res, err := pipeline.OpenPR(dir, pipeline.PROptions{JobID: id, Worker: workerFromFlags(cmd), BodyFile: body})
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		verb := "aberto"
		switch {
		case res.BaseFixed:
			verb = "já existia; base corrigida"
		case !res.Created:
			verb = "já existia"
		}
		fmt.Fprintf(out, "PR #%d %s: %s\nBase do PR: %s (fonte: %s)\nPR registrado no job %s\n", res.Number, verb, res.URL, res.Base, res.BaseSource, id)
		return nil
	},
}

func init() {
	pipelinePRCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelinePRCmd.Flags().String("id", "", "ID do job em processamento")
	_ = pipelinePRCmd.MarkFlagRequired("id")
	pipelinePRCmd.Flags().String("worker", "", "Worker dono do lease (padrão: variável FORGE_WORKER, definida pelo worker)")
	pipelinePRCmd.Flags().String("body-file", "", "Corpo do PR (relativo ao projeto), pelo modelo de .github/pull_request_template.md")
	_ = pipelinePRCmd.MarkFlagRequired("body-file")
	pipelineCmd.AddCommand(pipelinePRCmd)
}
