package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context <task-id|caminho-da-spec>",
	Short: "Imprime o contexto mínimo de uma task: spec, IDs da Constitution citados, rules e skills",
	Long: `Monta o contexto mínimo para executar uma task sem ler a Constitution inteira:
a spec, só as linhas dos IDs que ela cita, os arquivos de .agents/rules/ que cobrem esses
IDs e as skills aplicáveis (a da estação mais as skills aprovadas citadas na spec).

O argumento é o ID de uma task da fila da iniciativa da branch atual ou o caminho de uma spec relativo ao projeto
(fora de uma branch de iniciativa, informe o caminho da spec).
ID citado e inexistente na Constitution é erro. Somente leitura.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		station, _ := cmd.Flags().GetString("station")

		specPath, extraIDs, err := resolveContextTarget(dir, args[0])
		if err != nil {
			return err
		}
		pack, err := gate.BuildContext(dir, specPath, extraIDs, station)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), pack.Render())
		return err
	},
}

// resolveContextTarget aceita um caminho de spec existente ou um task-id da fila ativa.
func resolveContextTarget(dir, arg string) (specPath string, extraIDs []string, err error) {
	candidate := arg
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(dir, arg)
	}
	if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
		return arg, nil, nil // BuildContext confina o caminho ao projeto
	}

	q, err := queue.Load(dir)
	if err != nil {
		return "", nil, err
	}
	var ids []string
	for _, item := range q.Items {
		ids = append(ids, item.TaskID)
		if item.TaskID == arg {
			return item.SpecPath, item.Rules, nil
		}
	}
	return "", nil, fmt.Errorf("task %q não encontrada na fila ativa (tasks: %s); o argumento também pode ser o caminho de uma spec", arg, strings.Join(ids, ", "))
}

func init() {
	contextCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	contextCmd.Flags().String("station", "build", "Estação cuja skill carregar: concepcao, planejamento, build, governanca ou operacao")
	rootCmd.AddCommand(contextCmd)
}
