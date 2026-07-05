package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [diretório]",
	Short: "Remove completamente a estrutura do Forge-SDD e configurações de agentes",
	Long: `Deleta a pasta sdd/ (features, progresso, configuração) e todos os arquivos de integração
dos agentes de IA (CLAUDE.md, GEMINI.md, .claude/, .gemini/, etc.) instalados no repositório.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}

		destroyYes, _ := cmd.Flags().GetBool("yes")
		destroyDryRun, _ := cmd.Flags().GetBool("dry-run")

		sddrcPath := filepath.Join(targetDir, "sdd", ".sddrc")
		if _, err := os.Stat(sddrcPath); os.IsNotExist(err) {
			fmt.Printf("ℹ️  Nenhuma estrutura Forge-SDD activa encontrada em '%s'.\n", targetDir)
			return nil
		}

		// Se não foi passada a flag --yes, pede confirmação interativa
		if !destroyYes && !destroyDryRun {
			confirm, err := survey.RunConfirm(
				"⚠️ ATENÇÃO: Tem certeza que deseja apagar o Forge-SDD?",
				"Esta ação apagará permanentemente a pasta sdd/ (histórico, progresso, especificações) e todas as configurações dos agentes no diretório alvo.",
				false,
			)
			if err != nil {
				return fmt.Errorf("ação cancelada: %w", err)
			}
			if !confirm {
				fmt.Println("Operação abortada pelo usuário.")
				return nil
			}
		}

		pathsToRemove := []string{
			filepath.Join(targetDir, "sdd"),
			filepath.Join(targetDir, "CLAUDE.md"),
			filepath.Join(targetDir, ".claude"),
			filepath.Join(targetDir, "GEMINI.md"),
			filepath.Join(targetDir, ".gemini"),
			filepath.Join(targetDir, "OPENAI.md"),
			filepath.Join(targetDir, ".openai"),
			filepath.Join(targetDir, ".github", "copilot-instructions.md"),
			filepath.Join(targetDir, ".vscode", "mcp.json"),
		}

		fmt.Println("🧹 Iniciando limpeza da metodologia...")

		removedCount := 0
		for _, path := range pathsToRemove {
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					continue
				}
			}

			if destroyDryRun {
				fmt.Printf("  [DRY-RUN] Seria removido: %s\n", path)
				removedCount++
				continue
			}

			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("  ✗ Falha ao remover %s: %v\n", path, err)
			} else {
				fmt.Printf("  ✓ Removido com sucesso: %s\n", path)
				removedCount++
			}
		}

		if destroyDryRun {
			fmt.Printf("\n[DRY-RUN] Fim da simulação. %d item(ns) seriam removidos.\n", removedCount)
		} else {
			fmt.Printf("\n🔥 Estrutura Forge-SDD removida com sucesso (%d itens limpos).\n", removedCount)
		}

		return nil
	},
}

func init() {
	destroyCmd.Flags().BoolP("yes", "y", false, "Ignorar prompts de confirmação interativos")
	destroyCmd.Flags().Bool("dry-run", false, "Simular a remoção listando os arquivos afetados")
}
