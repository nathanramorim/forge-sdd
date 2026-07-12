package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const defaultAutopilotMinCycles = 3

var autopilotCmd = &cobra.Command{
	Use:   "autopilot [diretório]",
	Short: "Ativa o modo autopilot (sdd/.sdd-auto-pilot) após graduação por ciclos completos",
	Long: `Cria o arquivo sdd/.sdd-auto-pilot, sinalizando aos agentes de IA que podem pular a
confirmação humana entre as fases do ciclo SDD. Por padrão, essa ativação é bloqueada até que
o projeto acumule um número mínimo de ciclos completos (outcome "approved") registrados em
sdd/.metrics/, preservando a função didática da metodologia para quem ainda está aprendendo.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}

		skipGraduation, _ := cmd.Flags().GetBool("skip-graduation")
		minCycles, _ := cmd.Flags().GetInt("min-cycles")
		if minCycles <= 0 {
			minCycles = defaultAutopilotMinCycles
		}

		flagPath := filepath.Join(targetDir, "sdd", ".sdd-auto-pilot")
		if fileExists(flagPath) {
			fmt.Println("✓ Modo autopilot já está ativado (sdd/.sdd-auto-pilot presente).")
			return nil
		}

		approved, err := countApprovedCycles(filepath.Join(targetDir, "sdd", ".metrics"))
		if err != nil {
			return fmt.Errorf("falha ao ler sdd/.metrics: %w", err)
		}

		if !skipGraduation && approved < minCycles {
			missing := minCycles - approved
			return fmt.Errorf(
				"graduação necessária para o autopilot: %d/%d ciclos completos (outcome \"approved\") registrados em sdd/.metrics/. Faltam %d ciclo(s). Use --skip-graduation para um bypass consciente",
				approved, minCycles, missing,
			)
		}

		if err := os.MkdirAll(filepath.Dir(flagPath), 0755); err != nil {
			return fmt.Errorf("falha ao criar %s: %w", filepath.Dir(flagPath), err)
		}
		if err := os.WriteFile(flagPath, []byte{}, 0644); err != nil {
			return fmt.Errorf("falha ao criar %s: %w", flagPath, err)
		}

		if skipGraduation && approved < minCycles {
			fmt.Printf("⚠️  Modo autopilot ativado via bypass consciente (--skip-graduation), com apenas %d/%d ciclos completos registrados.\n", approved, minCycles)
		} else {
			fmt.Printf("✓ Modo autopilot ativado (%d/%d ciclos completos registrados). Criado sdd/.sdd-auto-pilot.\n", approved, minCycles)
		}
		return nil
	},
}

// countApprovedCycles lê os arquivos sdd/.metrics/session-*.json e conta quantos
// têm "outcome": "approved", indicando um ciclo SDD completo bem-sucedido.
func countApprovedCycles(metricsDir string) (int, error) {
	entries, err := os.ReadDir(metricsDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var metric struct {
		Outcome string `json:"outcome"`
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(metricsDir, e.Name()))
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &metric); err != nil {
			continue
		}
		if metric.Outcome == "approved" {
			count++
		}
	}
	return count, nil
}

func init() {
	autopilotCmd.Flags().Bool("skip-graduation", false, "Bypass consciente: ativa o autopilot mesmo sem os ciclos mínimos completos")
	autopilotCmd.Flags().Int("min-cycles", defaultAutopilotMinCycles, "Número mínimo de ciclos completos (outcome approved) exigidos para graduação")
	rootCmd.AddCommand(autopilotCmd)
}
