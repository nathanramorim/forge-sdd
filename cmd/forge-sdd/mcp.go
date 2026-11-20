package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/doctor"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Servidor MCP do forge-sdd (somente leitura)",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a ferramenta forge_doctor via MCP em stdio",
	Long: `Expõe a ferramenta forge_doctor (diagnóstico do harness, somente leitura) por JSON-RPC em
stdio, para o Claude Code (app desktop incluso), o Codex e outros clientes MCP. O diretório vem
da flag --dir; o cliente não escolhe caminho nem comando.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		diag := func() ([]byte, error) { return mcpDoctorReport(dir, doctor.DefaultDeps()) }
		transport := acp.NewDefaultStdioTransport(acp.NewMCPDispatcher(version, diag))
		defer func() { _ = transport.Close() }()
		if err := transport.Run(ctx); err != nil && err != ctx.Err() {
			return fmt.Errorf("erro no servidor mcp stdio: %w", err)
		}
		return nil
	},
}

// mcpDoctorReport roda o doctor (somente leitura), valida o contrato e serializa em JSON.
func mcpDoctorReport(dir string, deps doctor.Deps) ([]byte, error) {
	rel := doctor.Diagnose(dir, deps)
	if err := doctor.ValidateRelatorio(rel); err != nil {
		return nil, err
	}
	return json.Marshal(rel)
}

func init() {
	mcpServeCmd.Flags().String("dir", ".", "Diretório do projeto Forge-SDD")
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}
