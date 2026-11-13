package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker de fila do pipeline: reivindica jobs e executa o agente (builder ou reviewer)",
	Long: `Laço de longa duração para uma sessão do pipeline. Mostra "Aguardando jobs na fila..." quando
ocioso; para cada job reivindicado invoca o agente de forma não interativa com o perfil da CLI de
forge.json (claude, gemini, copilot, codex ou agy), do agente da sessão (sessions) ou de --cli, renova o lease, grava o log em .forge/runtime/logs e fecha o job.

Builder: exige que o agente registre o PR (forge-sdd pipeline pr-opened); então enfileira a revisão.
Reviewer: exige que o agente aplique o veredito (forge-sdd bus review).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		role, _ := cmd.Flags().GetString("role")
		name, _ := cmd.Flags().GetString("name")
		once, _ := cmd.Flags().GetBool("once")
		poll, _ := cmd.Flags().GetDuration("poll")
		cli, _ := cmd.Flags().GetString("cli")

		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return pipeline.RunWorker(ctx, pipeline.WorkerOptions{
			Dir: dir, Role: role, Name: name, CLI: cli, Once: once, Poll: poll,
			Out:       cmd.OutOrStdout(),
			OnJobDone: recordJobTelemetry(dir, role),
		})
	},
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
		ok := outcome == pipeline.OutcomeDone
		result := "approved"
		if !ok {
			result = "blocked"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, self, "session", "record", "--dir", dir, "--feature", job.SpecPath, "--phase", job.ID,
			"--agent-path", role, "--outcome", result, fmt.Sprintf("--criterio-atendido=%t", ok)).Run()
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
		if _, err := bus.UpdatePR(dir, layout.RoleBuilder, id, worker, number, url, branch, time.Now()); err != nil {
			return err
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
		batches, err := pipeline.PlanWaves(base, cfg.Builder.MaxParallelTasks)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "até %d tarefas em paralelo (use_git_worktrees=%t)\n%s", cfg.Builder.MaxParallelTasks, cfg.Builder.UseGitWorktrees, pipeline.FormatWaves(batches))
		return nil
	},
}

func init() {
	workerCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	workerCmd.Flags().String("role", "", "Papel do worker: builder ou reviewer")
	_ = workerCmd.MarkFlagRequired("role")
	workerCmd.Flags().String("name", "", "Nome da sessão (padrão: forge-<papel>)")
	workerCmd.Flags().String("cli", "", "Agente desta sessão: claude, gemini, copilot ou codex (padrão: sessions de forge.json)")
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
