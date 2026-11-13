package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/spf13/cobra"
)

var busCmd = &cobra.Command{
	Use:   "bus",
	Short: "Barramento de jobs entre as sessões do pipeline (builder e reviewer)",
	Long: `Opera a fila local em .forge/runtime/queues: um arquivo JSON por job que muda de diretório
(pending, processing, done, failed). A reivindicação é atômica e tem lease; high é consumido antes
de normal e, dentro da mesma prioridade, a ordem é FIFO.`,
}

func busFlags(c *cobra.Command, needRole bool) {
	c.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	if needRole {
		c.Flags().String("role", "", "Fila: builder ou reviewer")
		_ = c.MarkFlagRequired("role")
	}
}

func busArgs(cmd *cobra.Command) (dir, role string) {
	dir, _ = cmd.Flags().GetString("dir")
	if dir == "" {
		dir = "."
	}
	role, _ = cmd.Flags().GetString("role")
	return dir, role
}

var busEnqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "Valida e enfileira um job (JSON de arquivo ou stdin com --job-file -)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		file, _ := cmd.Flags().GetString("job-file")
		var r io.Reader
		if file == "-" {
			r = io.LimitReader(cmd.InOrStdin(), bus.MaxJobBytes+1)
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("abrir %s: %w", file, err)
			}
			defer f.Close()
			r = io.LimitReader(f, bus.MaxJobBytes+1)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if len(data) > bus.MaxJobBytes {
			return fmt.Errorf("%w: acima de %d bytes", bus.ErrTooLarge, bus.MaxJobBytes)
		}
		var job bus.Job
		if err := json.Unmarshal(data, &job); err != nil {
			return fmt.Errorf("%w: %v", bus.ErrInvalidJob, err)
		}
		path, created, err := bus.Enqueue(dir, role, job, time.Now())
		if err != nil {
			return err
		}
		if created {
			fmt.Fprintf(cmd.OutOrStdout(), "enfileirado %s\n", path)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "já existe (idempotente) %s\n", path)
		}
		return nil
	},
}

var busClaimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Reivindica o próximo job (imprime o job em JSON; fila vazia não é erro)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		worker, _ := cmd.Flags().GetString("worker")
		ttl, _ := cmd.Flags().GetDuration("ttl")
		e, err := bus.Claim(dir, role, worker, ttl, time.Now())
		if errors.Is(err, bus.ErrEmpty) {
			fmt.Fprintln(cmd.ErrOrStderr(), "fila vazia")
			return nil
		}
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), e.Job)
	},
}

var busCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Conclui um job reivindicado (processing para done)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		id, _ := cmd.Flags().GetString("id")
		worker, _ := cmd.Flags().GetString("worker")
		note, _ := cmd.Flags().GetString("note")
		e, err := bus.Complete(dir, role, id, worker, note, time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "concluído %s\n", e.Path)
		return nil
	},
}

var busFailCmd = &cobra.Command{
	Use:   "fail",
	Short: "Marca um job reivindicado como falho (processing para failed)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		id, _ := cmd.Flags().GetString("id")
		worker, _ := cmd.Flags().GetString("worker")
		reason, _ := cmd.Flags().GetString("reason")
		e, err := bus.Fail(dir, role, id, worker, reason, time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "falhou %s\n", e.Path)
		return nil
	},
}

var busRecoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Devolve a pending os jobs de processing com lease vencido",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		n, err := bus.Recover(dir, role, time.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "recuperados: %d\n", n)
		return nil
	},
}

var busListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista os jobs de um estado, na ordem de consumo",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		state, _ := cmd.Flags().GetString("state")
		asJSON, _ := cmd.Flags().GetBool("json")
		entries, err := bus.List(dir, role, state)
		if err != nil {
			return err
		}
		if asJSON {
			jobs := make([]bus.Job, 0, len(entries))
			for _, e := range entries {
				jobs = append(jobs, e.Job)
			}
			return printJSON(cmd.OutOrStdout(), jobs)
		}
		if len(entries) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "(vazio) %s/%s\n", role, state)
			return nil
		}
		for _, e := range entries {
			fmt.Fprintf(cmd.OutOrStdout(), "%-8s %-14s tentativa %d  %s  %s\n", e.Job.Priority, e.Job.Type, e.Job.Attempt, e.Job.ID, e.Job.Title)
		}
		return nil
	},
}

var busShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Mostra um job em qualquer estado",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, role := busArgs(cmd)
		id, _ := cmd.Flags().GetString("id")
		e, err := bus.Find(dir, role, id)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "estado: %s\n", e.State)
		return printJSON(cmd.OutOrStdout(), e.Job)
	},
}

func printJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func init() {
	for _, c := range []*cobra.Command{busEnqueueCmd, busClaimCmd, busCompleteCmd, busFailCmd, busRecoverCmd, busListCmd, busShowCmd} {
		busFlags(c, true)
		busCmd.AddCommand(c)
	}
	busEnqueueCmd.Flags().String("job-file", "", "Arquivo JSON do job (ou - para stdin)")
	_ = busEnqueueCmd.MarkFlagRequired("job-file")
	busClaimCmd.Flags().String("worker", "", "Nome do worker (lease)")
	_ = busClaimCmd.MarkFlagRequired("worker")
	busClaimCmd.Flags().Duration("ttl", bus.DefaultLeaseTTL, "Duração do lease")
	for _, c := range []*cobra.Command{busCompleteCmd, busFailCmd, busShowCmd} {
		c.Flags().String("id", "", "ID do job")
		_ = c.MarkFlagRequired("id")
	}
	busCompleteCmd.Flags().String("worker", "", "Worker dono do lease")
	busCompleteCmd.Flags().String("note", "", "Observação registrada no histórico")
	busFailCmd.Flags().String("worker", "", "Worker dono do lease")
	busFailCmd.Flags().String("reason", "", "Motivo da falha")
	busListCmd.Flags().String("state", layout.StatePending, "Estado: pending, processing, done ou failed")
	busListCmd.Flags().Bool("json", false, "Saída em JSON")
	rootCmd.AddCommand(busCmd)
}
