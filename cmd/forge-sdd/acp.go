package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/spf13/cobra"
)

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Inicia o servidor ACP (Agent Client Protocol) via transporte stdio",
	Long: `Inicia o servidor do Protocolo ACP (Agent Client Protocol) em modo stdio.
Permite integração bidirecional JSON-RPC 2.0 com IDEs externas (Zed, Cursor, VS Code)
e com o Forge Studio para sincronização de sessões, handoff entre estações e portões.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		dispatcher := acp.NewDispatcherWithVersion(version)
		transport := acp.NewDefaultStdioTransport(dispatcher)
		defer func() {
			_ = transport.Close()
		}()

		if err := transport.Run(ctx); err != nil && err != ctx.Err() {
			return fmt.Errorf("erro no servidor acp stdio: %w", err)
		}
		return nil
	},
}

func init() {
	acpCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
}
