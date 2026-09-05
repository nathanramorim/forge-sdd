package main

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// reportItem agrega as métricas de todas as sessões associadas ao mesmo
// caminho de feature/fix/discovery.
type reportItem struct {
	Type            string
	Feature         string
	Sessions        int
	TokensInput     int
	TokensOutput    int
	DurationSeconds int
	Models          map[string]bool
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Relatório determinístico de tokens/modelo/duração por feature, fix e discovery",
	Long: `Agrega sdd/.metrics/session-*.json e imprime, por item (discovery/feature/fix),
os tokens de entrada+saída, os modelos usados, o número de sessões e a duração
total — além da idade medida do projeto (da métrica mais antiga à mais recente).

Determinístico: não depende de o agente de IA montar a tabela manualmente.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}

		records, err := readSessionFiles(targetDir)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			fmt.Println("Nenhuma métrica encontrada em sdd/.metrics — nada para relatar.")
			return nil
		}

		items := map[string]*reportItem{}
		var order []string
		var oldest, newest time.Time

		for _, r := range records {
			m := r.Metrics
			key := m.Feature
			if key == "" {
				key = "(sem feature)"
			}
			item, ok := items[key]
			if !ok {
				item = &reportItem{Type: ClassifySessionType(m.Feature), Feature: key, Models: map[string]bool{}}
				items[key] = item
				order = append(order, key)
			}
			item.Sessions++
			item.TokensInput += m.TokensInput
			item.TokensOutput += m.TokensOutput
			item.DurationSeconds += m.DurationSeconds
			if m.Model != "" {
				item.Models[m.Model] = true
			}

			if !r.Timestamp.IsZero() {
				if oldest.IsZero() || r.Timestamp.Before(oldest) {
					oldest = r.Timestamp
				}
				if newest.IsZero() || r.Timestamp.After(newest) {
					newest = r.Timestamp
				}
			}
		}

		sort.Slice(order, func(i, j int) bool {
			a, b := items[order[i]], items[order[j]]
			if a.Type != b.Type {
				return a.Type < b.Type
			}
			return a.Feature < b.Feature
		})

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
		fmt.Fprintln(w, "TIPO\tITEM\tSESSOES\tTOKENS(IN+OUT)\tMODELOS\tDURACAO")
		for _, key := range order {
			it := items[key]
			models := make([]string, 0, len(it.Models))
			for model := range it.Models {
				models = append(models, model)
			}
			sort.Strings(models)
			modelsStr := strings.Join(models, ",")
			if modelsStr == "" {
				modelsStr = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\n",
				it.Type, it.Feature, it.Sessions, it.TokensInput+it.TokensOutput, modelsStr,
				formatDuration(it.DurationSeconds))
		}
		w.Flush()

		if !oldest.IsZero() && !newest.IsZero() {
			fmt.Printf("\nIdade medida do projeto (1ª → última métrica registrada): %s\n",
				formatDuration(int(newest.Sub(oldest).Seconds())))
		}
		return nil
	},
}

// formatDuration imprime uma duração em segundos como "XdYhZm", omitindo
// unidades zeradas à esquerda (ex: "3h12m", "45m").
func formatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func init() {
	reportCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
}
