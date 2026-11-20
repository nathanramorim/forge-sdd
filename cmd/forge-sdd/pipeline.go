package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

// askCard pergunta se o projeto usa gerenciador de projeto (card obrigatório); variável para os testes.
var askCard = func() (bool, error) {
	return survey.ConfirmYesNo("Usa gerenciador de projeto (ClickUp, Jira, Azure DevOps…)? Se sim, o /implementar exige o card.", false)
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Pipeline multi-sessão (Discovery, Builder e Reviewer): configuração e apoio",
}

var pipelineInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Cria .forge/config/forge.json e notifications.json com os padrões (ativa o modo pipeline)",
	Long: `Grava .forge/config/forge.json e .forge/config/notifications.json a partir dos padrões, sem
sobrescrever arquivos existentes. A existência de forge.json ativa o modo pipeline: /implementar
passa a validar o card e despachar para a sessão Builder. Sem ele, o fluxo da 4.0.0-beta continua.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		return runPipelineInit(cmd, dir, dry)
	},
}

func runPipelineInit(cmd *cobra.Command, dir string, dry bool) error {
	out := cmd.OutOrStdout()
	type item struct {
		path string
		data any
	}
	forgeCfg := config.DefaultForgeConfig() // sem terminal: RequireExistingCard = false
	if _, err := os.Lstat(layout.ForgeConfigFile(dir)); err != nil && !dry && isInteractive() {
		if need, aerr := askCard(); aerr == nil {
			forgeCfg.PMMCP.RequireExistingCard = need
		}
	}
	items := []item{
		{layout.ForgeConfigFile(dir), forgeCfg},
		{layout.NotificationsFile(dir), config.DefaultNotificationsConfig()},
	}
	for _, it := range items {
		rel, _ := filepath.Rel(dir, it.path)
		if _, err := os.Lstat(it.path); err == nil {
			fmt.Fprintf(out, "mantido   %s (já existe)\n", filepath.ToSlash(rel))
			continue
		}
		if dry {
			fmt.Fprintf(out, "[DRY] criaria %s\n", filepath.ToSlash(rel))
			continue
		}
		data, err := json.MarshalIndent(it.data, "", "  ")
		if err != nil {
			return fmt.Errorf("serializar %s: %w", rel, err)
		}
		if err := scaffold.WriteFileSafe(dir, it.path, append(data, '\n')); err != nil {
			return fmt.Errorf("gravar %s: %w", rel, err)
		}
		fmt.Fprintf(out, "criado    %s\n", filepath.ToSlash(rel))
	}

	// O estado de execução nunca é versionado: o diretório ignora tudo menos o próprio .gitignore.
	ignore := filepath.Join(layout.ForgeRuntime(dir), ".gitignore")
	if _, err := os.Lstat(ignore); err != nil && !dry {
		if err := scaffold.WriteFileSafe(dir, ignore, []byte(runtimeGitignore)); err != nil {
			return fmt.Errorf("gravar %s: %w", ignore, err)
		}
	}
	if dry {
		fmt.Fprintln(out, "[DRY-RUN] Nada foi gravado.")
		return nil
	}
	fmt.Fprintln(out, "\nModo pipeline ativo. Próximos passos:")
	fmt.Fprintln(out, "  1. forge                        menu: Pipeline multi-sessão → escolher o agente de cada sessão e abrir")
	fmt.Fprintln(out, "     (ou forge up --discovery claude --builder codex --reviewer gemini)")
	fmt.Fprintln(out, "  2. /implementar <feature>       na sessão forge-discovery: valida o card e despacha ao Builder")
	fmt.Fprintln(out, "  3. forge inbox              mostra o aviso quando o Reviewer aprovar")
	fmt.Fprintln(out, "  Ajuste a CLI e as permissões do agente em .forge/config/forge.json. Guia: .forge/LEIA-ME.md")
	return nil
}

const runtimeGitignore = "# Estado de execução do pipeline (filas, logs, worktrees): não versionar.\n*\n!.gitignore\n"

var pipelineAutonomyCmd = &cobra.Command{
	Use:   "autonomy normal|advanced",
	Short: "Liga ou desliga o modo avançado (experimental): o Reviewer revisa a spec",
	Long: `normal  (padrão) a aprovação da spec é humana.
advanced  o Reviewer revisa a spec no lugar da pessoa (ADR-0013); exige permissions full e a
          confirmação de ambiente isolado (--isolated). Use só em contêiner, VM ou máquina descartável.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		isolated, _ := cmd.Flags().GetBool("isolated")
		cfg, found, err := config.LoadForgeConfig(dir)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("modo pipeline não ativo: rode forge pipeline init")
		}
		switch args[0] {
		case config.AutonomyNormal:
			config.ApplyNormal(&cfg)
		case config.AutonomyAdvanced:
			if !isolated {
				return fmt.Errorf("o modo avançado exige confirmar o ambiente isolado: repita com --isolated")
			}
			config.ApplyAdvanced(&cfg)
		default:
			return fmt.Errorf("modo %q inválido (use normal ou advanced)", args[0])
		}
		if err := config.SaveForgeConfig(dir, cfg); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "autonomy: %s (gravado em .forge/config/forge.json)\n", args[0])
		return nil
	},
}

func init() {
	pipelineAutonomyCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineAutonomyCmd.Flags().Bool("isolated", false, "Confirma que o ambiente é isolado (obrigatório para advanced)")
	pipelineCmd.AddCommand(pipelineAutonomyCmd)
	pipelineInitCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineInitCmd.Flags().Bool("dry-run", false, "Mostra o que seria criado sem gravar")
	pipelineCmd.AddCommand(pipelineInitCmd)
	rootCmd.AddCommand(pipelineCmd)
}
