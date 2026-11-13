package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Abre as 3 sessões do pipeline: forge-discovery, forge-builder e forge-reviewer",
	Long: `Abre três terminais nomeados no sistema operacional atual (macOS, Linux, WSL ou Windows):
  forge-discovery   sessão interativa do agente (claude ou agy)
  forge-builder     forge-sdd worker --role builder
  forge-reviewer    forge-sdd worker --role reviewer
A CLI vem de --cli ou de .forge/config/forge.json. --dry-run só imprime os comandos.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		cfg, found, err := config.LoadForgeConfig(dir)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("%w (sem .forge/config/forge.json não há pipeline para abrir)", pipeline.ErrNotPipelineMode)
		}
		cli := cfg.CLI
		if c, _ := cmd.Flags().GetString("cli"); c != "" {
			cli = c
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		abs, err := absDir(dir)
		if err != nil {
			return err
		}
		plan, err := pipeline.LaunchPlan(pipeline.DefaultLaunchEnv(runtime.GOOS, abs, pipeline.ResolveBin()), cli)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		for _, c := range plan {
			fmt.Fprintf(out, "%-16s %s\n", c.Title, strings.Join(quoteArgs(c.Argv), " "))
		}
		if dry {
			fmt.Fprintln(out, "[DRY-RUN] Nenhuma sessão foi aberta.")
			return nil
		}
		return pipeline.Launch(plan)
	},
}

func quoteArgs(argv []string) []string {
	out := make([]string, len(argv))
	for i, a := range argv {
		if strings.ContainsAny(a, " \t\"'$;&|<>()") {
			out[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		} else {
			out[i] = a
		}
	}
	return out
}

func init() {
	upCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	upCmd.Flags().String("cli", "", "CLI do agente: claude ou agy (padrão: forge.json)")
	upCmd.Flags().Bool("dry-run", false, "Só imprime os comandos de abertura")
	rootCmd.AddCommand(upCmd)
}

func absDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolver %q: %w", dir, err)
	}
	return abs, nil
}
