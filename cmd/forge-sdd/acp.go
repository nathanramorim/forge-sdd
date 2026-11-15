package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/spf13/cobra"
)

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Inicia o servidor ACP (Agent Client Protocol) via transporte stdio",
	Long: `Inicia o servidor do Protocolo ACP (Agent Client Protocol) em modo stdio.
Permite integração bidirecional JSON-RPC 2.0 com IDEs externas (Zed, Cursor, VS Code)
para sincronização de sessões, handoff entre estações e portões. O servidor também
mantém o heartbeat do pool de sessões enquanto estiver ativo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}

		dispatcher := acp.NewDispatcherWithVersion(version)
		stopBackground := startACPBackground(ctx, dir, dispatcher.Pool())
		defer stopBackground()

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

// startACPBackground higieniza atividades órfãs do workspace e inicia o heartbeat
// do pool enquanto o servidor ACP estiver ativo. A função devolvida encerra o heartbeat.
func startACPBackground(ctx context.Context, dir string, pool *acp.SessionPool) func() {
	_, _ = queue.SanitizeOrphanActivities(dir)
	if pool == nil {
		return func() {}
	}

	pool.AddStateListener(func(station string, state acp.SessionState, _ time.Time) {
		if state == acp.SessionStateStandby {
			_ = queue.ReconcileStationActivities(dir, station)
		}
	})
	pool.StartHeartbeat(ctx)

	return func() { _ = pool.Close() }
}

func init() {
	acpCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
}
