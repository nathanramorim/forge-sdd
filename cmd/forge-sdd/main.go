package main

import (
	"fmt"
	"os"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

// version é injetada via ldflags: -X main.version=1.4.3
var version = "1.4.3"

var rootCmd = &cobra.Command{
	Use:   "forge-sdd",
	Short: "Scaffolda estruturas Forge-SDD em qualquer projeto",
	Long: `forge-sdd é um CLI que gera a estrutura completa Forge-SDD
(sdd/, .github/, .vscode/) pronta para uso com GitHub Copilot.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do forge-sdd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init [diretório]",
	Short: "Inicializa a estrutura Forge-SDD no diretório alvo",
	Long: `Cria toda a árvore Forge-SDD no diretório especificado (default: diretório atual).
Preenche os templates com as informações do projeto coletadas interativamente ou via flags.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}

		yes, _ := cmd.Flags().GetBool("yes")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		var cfg config.Config
		if yes {
			cfg = config.FromFlags(cmd)
			// validar --agent explicitamente para retornar erro claro
			if f := cmd.Flags().Lookup("agent"); f != nil && f.Changed {
				agents, err := config.ParseAgents(f.Value.String())
				if err != nil {
					return err
				}
				cfg.Agents = agents
			}
			cfg.DryRun = dryRun
		} else {
			var err error
			cfg, err = survey.Run()
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			cfg.DryRun = dryRun
		}

		// usa versão do binário se não sobrescrita pela flag --version
		if cfg.SddVersion == config.Defaults().SddVersion && version != "dev" {
			cfg.SddVersion = version
		}

		created, err := scaffold.Run(cfg, targetDir)
		if err != nil {
			return fmt.Errorf("scaffold falhou: %w", err)
		}

		fmt.Printf("\n✓ Estrutura Forge-SDD criada em %s (%d arquivos)\n\n", targetDir, len(created))
		fmt.Println("Próximos passos:")
		fmt.Println("  1. Abra o projeto no VS Code")
		fmt.Println("  2. Aceite as extensões recomendadas (Copilot, MCP)")
		fmt.Println("  3. Leia sdd/memory/progress.md para começar")
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("yes", false, "Pular prompts e usar flags/defaults")
	initCmd.Flags().Bool("dry-run", false, "Listar arquivos que seriam criados sem gravar no disco")
	initCmd.Flags().String("name", "", "Nome do projeto")
	initCmd.Flags().String("stack", "", "Stack principal: go, node, python, rust, other")
	initCmd.Flags().String("db", "", "Banco de dados: postgres, sqlite, mongo, none")
	initCmd.Flags().String("lang", "", "Idioma: pt-BR, en")
	initCmd.Flags().String("version", "", "Versão Forge-SDD (default: 1.4.3)")
	initCmd.Flags().Bool("no-telemetry", false, "Desabilitar telemetria local")
	initCmd.Flags().String("agent", "", "Agente(s) de IA: copilot, claude, gemini (csv, default: copilot)")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
