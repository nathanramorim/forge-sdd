package main

import (
	"fmt"
	"os"

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
		fmt.Printf("forge-sdd init — diretório alvo: %s (stub)\n", targetDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
