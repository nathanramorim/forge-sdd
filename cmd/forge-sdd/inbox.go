package main

import (
	"fmt"

	"github.com/forge-sdd/cli/internal/notify"
	"github.com/spf13/cobra"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Mostra as notificações do pipeline (PR pronto, escalonamentos) da sessão Discovery",
	Long: `Lista as notificações não lidas em .forge/runtime/queues/discovery/inbox, com o link do PR revisado.

--hook imprime só o necessário e marca as notificações como lidas (para o hook da sessão Discovery
não repetir a mesma mensagem a cada prompt); sem notificações não imprime nada.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		out := cmd.OutOrStdout()
		hook, _ := cmd.Flags().GetBool("hook")
		includeRead, _ := cmd.Flags().GetBool("include-read")
		markID, _ := cmd.Flags().GetString("mark-read")
		markAll, _ := cmd.Flags().GetBool("mark-all")
		archive, _ := cmd.Flags().GetBool("archive")

		if markID != "" {
			if err := notify.MarkRead(dir, markID); err != nil {
				return err
			}
			fmt.Fprintf(out, "lida: %s\n", markID)
			return nil
		}
		if markAll {
			n, err := notify.MarkAllRead(dir)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "marcadas como lidas: %d\n", n)
			return nil
		}
		if archive {
			n, err := notify.ArchiveRead(dir)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "arquivadas: %d\n", n)
			return nil
		}

		items, err := notify.ListInbox(dir, !includeRead)
		if err != nil {
			return err
		}
		if hook {
			fmt.Fprint(out, notify.FormatInbox(items, true))
			for _, it := range items {
				if !it.Read {
					_ = notify.MarkRead(dir, it.ID)
				}
			}
			return nil
		}
		if len(items) == 0 {
			fmt.Fprintln(out, "Nenhuma notificação não lida.")
			return nil
		}
		fmt.Fprint(out, notify.FormatInbox(items, false))
		return nil
	},
}

func init() {
	inboxCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	inboxCmd.Flags().Bool("unread", true, "Mostra só as não lidas (padrão; use --include-read para todas)")
	inboxCmd.Flags().Bool("include-read", false, "Inclui as já lidas")
	inboxCmd.Flags().Bool("hook", false, "Formato compacto para o hook da sessão; marca como lidas")
	inboxCmd.Flags().String("mark-read", "", "Marca a notificação como lida")
	inboxCmd.Flags().Bool("mark-all", false, "Marca todas como lidas")
	inboxCmd.Flags().Bool("archive", false, "Move as lidas para o arquivo")
	rootCmd.AddCommand(inboxCmd)
}
