package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/spf13/cobra"
)

var busReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Aplica o veredito do Reviewer a um job pr_review (aprova, devolve ao Builder ou escala)",
	Long: `Aplica a regra de fluxo ao veredito JSON do agente reviewer (mesmo contrato de queue review):
  conforme            aprova, grava pr_approved no inbox, aciona os canais e instrui a mover o card
  violacao            devolve ao Builder como pr_rework (prioridade alta, tentativa + 1) com feedback;
                      no limite de max_rework_attempts escala para a pessoa (inbox: escalation)
  duvida              escala para a pessoa, sem rework automático
Aplicar duas vezes o mesmo veredito não repete efeitos.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		id, _ := cmd.Flags().GetString("id")
		worker, _ := cmd.Flags().GetString("worker")
		file, _ := cmd.Flags().GetString("verdict-file")
		raw, err := readLimited(cmd, file)
		if err != nil {
			return err
		}
		var v queue.Verdict
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("veredito inválido: %w", err)
		}
		res, err := pipeline.ApplyVerdict(dir, id, worker, v, notify.Deps{}, time.Now())
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if res.AlreadyApplied {
			fmt.Fprintf(out, "veredito já aplicado ao job %s (%s); nada foi repetido\n", id, res.Action)
			return nil
		}
		switch res.Action {
		case pipeline.ActionApproved:
			fmt.Fprintf(out, "aprovado: PR pronto; notificação gravada no inbox\n")
		case pipeline.ActionRework:
			fmt.Fprintf(out, "não conforme: devolvido ao Builder como %s (tentativa %d de %d, prioridade %s)\n",
				res.ReworkJob.ID, res.ReworkJob.Attempt, res.ReworkJob.MaxReworkAttempts, res.ReworkJob.Priority)
		case pipeline.ActionEscalated:
			fmt.Fprintf(out, "escalado para a pessoa: veja o inbox (forge-sdd inbox)\n")
		}
		if res.CardInstruction != "" {
			fmt.Fprintf(out, "Próximo passo do agente: %s\n", res.CardInstruction)
		}
		names := make([]string, 0, len(res.Channels))
		for k := range res.Channels {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Fprintf(out, "  canal %-17s %s\n", k, res.Channels[k])
		}
		if res.PRComment != "" {
			fmt.Fprintf(out, "Comentário sugerido para o PR (publique com gh pr comment):\n%s", indent(res.PRComment))
		}
		return nil
	},
}

func readLimited(cmd *cobra.Command, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = cmd.InOrStdin()
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("abrir %s: %w", file, err)
		}
		defer f.Close()
		r = f
	}
	data, err := io.ReadAll(io.LimitReader(r, 256*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 256*1024 {
		return nil, fmt.Errorf("veredito acima de 256 KB")
	}
	return data, nil
}

func indent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  | " + l
	}
	return strings.Join(lines, "\n") + "\n"
}

func init() {
	busFlags(busReviewCmd, false)
	busReviewCmd.Flags().String("id", "", "ID do job pr_review em processamento")
	_ = busReviewCmd.MarkFlagRequired("id")
	busReviewCmd.Flags().String("verdict-file", "", "Arquivo JSON do veredito (ou - para stdin)")
	_ = busReviewCmd.MarkFlagRequired("verdict-file")
	busReviewCmd.Flags().String("worker", "", "Worker dono do lease (opcional)")
	busCmd.AddCommand(busReviewCmd)
}
