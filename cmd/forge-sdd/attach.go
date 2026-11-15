package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [estacao]",
	Short: "Conecta o terminal interativo à sessão de uma estação ACP",
	Long: `Conecta os fluxos de entrada e saída do terminal à sessão pré-aquecida
de uma das 5 estações canônicas do Forge-SDD (lider, pm-architect, planner, builder, governance).
Permite visualização de logs em tempo real e envio de comandos interativos.
A desconexão via Ctrl+C ou comando 'exit' é graciosa e não encerra a sessão subjacente no pool.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requer exatamente 1 argumento: <estacao>. Estações canônicas válidas: %s", strings.Join(queue.CanonicalStations(), ", "))
		}
		station := args[0]
		if !queue.IsValidStation(station) {
			return fmt.Errorf("estação inválida: %q. Estações canônicas válidas: %s", station, strings.Join(queue.CanonicalStations(), ", "))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		station := args[0]
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return runAttach(ctx, station, dir, cmd.InOrStdin(), cmd.OutOrStdout())
	},
}

func runAttach(ctx context.Context, station string, dir string, in io.Reader, out io.Writer) error {
	dispatcher := acp.NewDispatcherWithVersion(version)
	pool := dispatcher.Pool()
	if pool != nil && !pool.IsWarmed() {
		_, _ = pool.Warmup(ctx, dir)
	}

	bridge := dispatcher.AttachBridge()
	if bridge == nil {
		bridge = acp.NewAttachBridge(pool)
	}

	return bridge.Run(ctx, station, in, out)
}

func init() {
	attachCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
}
