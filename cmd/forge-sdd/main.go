package main

import (
	"fmt"
	"os"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forge-sdd",
	Short: "Scaffolda estruturas Forge-SDD em qualquer projeto",
	Long: `forge-sdd é um CLI que gera a estrutura completa Forge-SDD
(sdd/, .github/, .vscode/) pronta para uso com GitHub Copilot.`,
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

		var cfg config.Config
		if yes {
			cfg = config.FromFlags(cmd)
		} else {
			var err error
			cfg, err = survey.Run()
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
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
	initCmd.Flags().String("name", "", "Nome do projeto")
	initCmd.Flags().String("stack", "", "Stack principal: go, node, python, rust, other")
	initCmd.Flags().String("db", "", "Banco de dados: postgres, sqlite, mongo, none")
	initCmd.Flags().String("lang", "", "Idioma: pt-BR, en")
	initCmd.Flags().String("version", "", "Versão Forge-SDD (default: 1.1.0)")
	initCmd.Flags().Bool("no-telemetry", false, "Desabilitar telemetria local")
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
