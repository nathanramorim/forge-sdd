package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/forge-sdd/cli/internal/governance"
	"github.com/spf13/cobra"
)

var governanceCmd = &cobra.Command{
	Use:   "governance",
	Short: "Gerencia blocos e pacotes externos de governança (padrão manifest.yaml)",
}

var governanceApplyBlockCmd = &cobra.Command{
	Use:   "apply-block",
	Short: "Insere ou atualiza de forma idempotente um bloco demarcado em sdd/memory/constitution.md",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		pkg, _ := cmd.Flags().GetString("package")
		ver, _ := cmd.Flags().GetString("version")
		sha, _ := cmd.Flags().GetString("sha")
		blockFile, _ := cmd.Flags().GetString("block-file")

		if pkg == "" || blockFile == "" {
			return fmt.Errorf("--package e --block-file são obrigatórios")
		}

		blockBytes, err := os.ReadFile(blockFile)
		if err != nil {
			return fmt.Errorf("ler block-file: %w", err)
		}

		constPath := filepath.Join(dir, "sdd", "memory", "constitution.md")
		existing, err := os.ReadFile(constPath)
		if err != nil {
			return fmt.Errorf("ler %s: %w", constPath, err)
		}

		startPrefix := fmt.Sprintf("<!-- %s:inicio", pkg)
		endMarker := fmt.Sprintf("<!-- %s:fim -->", pkg)
		updated := governance.ApplyBlock(string(existing), startPrefix, endMarker, string(blockBytes))
		if err := os.WriteFile(constPath, []byte(updated), 0644); err != nil {
			return fmt.Errorf("gravar %s: %w", constPath, err)
		}

		if ver != "" {
			if _, err := governance.WriteVersionRecord(dir, pkg, ver, sha, time.Now().UTC()); err != nil {
				return err
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ Bloco de governança %s aplicado em %s\n", pkg, constPath)
		return nil
	},
}

func init() {
	governanceApplyBlockCmd.Flags().String("dir", ".", "Diretório raiz do projeto")
	governanceApplyBlockCmd.Flags().String("package", "", "Nome do pacote de governança (ex: planning-brain-governanca)")
	governanceApplyBlockCmd.Flags().String("version", "", "Versão do pacote importado (ex: 0.2.1)")
	governanceApplyBlockCmd.Flags().String("sha", "local", "Commit SHA do pacote importado")
	governanceApplyBlockCmd.Flags().String("block-file", "", "Arquivo contendo o bloco demarcado da constitution")
	governanceCmd.AddCommand(governanceApplyBlockCmd)
}
