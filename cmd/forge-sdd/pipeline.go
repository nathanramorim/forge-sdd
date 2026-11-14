package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/spf13/cobra"
)

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
	items := []item{
		{layout.ForgeConfigFile(dir), config.DefaultForgeConfig()},
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
		if err := os.MkdirAll(filepath.Dir(it.path), 0o755); err != nil {
			return fmt.Errorf("criar %s: %w", filepath.Dir(it.path), err)
		}
		if err := os.WriteFile(it.path, append(data, '\n'), 0o644); err != nil {
			return fmt.Errorf("gravar %s: %w", rel, err)
		}
		fmt.Fprintf(out, "criado    %s\n", filepath.ToSlash(rel))
	}

	// O estado de execução nunca é versionado: o diretório ignora tudo menos o próprio .gitignore.
	ignore := filepath.Join(layout.ForgeRuntime(dir), ".gitignore")
	if _, err := os.Lstat(ignore); err != nil && !dry {
		if err := os.MkdirAll(filepath.Dir(ignore), 0o755); err != nil {
			return fmt.Errorf("criar %s: %w", filepath.Dir(ignore), err)
		}
		if err := os.WriteFile(ignore, []byte(runtimeGitignore), 0o644); err != nil {
			return fmt.Errorf("gravar %s: %w", ignore, err)
		}
	}
	if dry {
		fmt.Fprintln(out, "[DRY-RUN] Nada foi gravado.")
		return nil
	}
	fmt.Fprintln(out, "\nModo pipeline ativo. Próximos passos:")
	fmt.Fprintln(out, "  1. forge                        menu: Pipeline multi-sessão → escolher o agente de cada sessão e abrir")
	fmt.Fprintln(out, "     (ou forge-sdd up --discovery claude --builder codex --reviewer gemini)")
	fmt.Fprintln(out, "  2. /implementar <feature>       na sessão forge-discovery: valida o card e despacha ao Builder")
	fmt.Fprintln(out, "  3. forge-sdd inbox              mostra o aviso quando o Reviewer aprovar")
	fmt.Fprintln(out, "  Ajuste a CLI e as permissões do agente em .forge/config/forge.json. Guia: .forge/LEIA-ME.md")
	return nil
}

const runtimeGitignore = "# Estado de execução do pipeline (filas, logs, worktrees): não versionar.\n*\n!.gitignore\n"

func init() {
	pipelineInitCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	pipelineInitCmd.Flags().Bool("dry-run", false, "Mostra o que seria criado sem gravar")
	pipelineCmd.AddCommand(pipelineInitCmd)
	rootCmd.AddCommand(pipelineCmd)
}
