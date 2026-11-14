package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Monitora os processos, sessões e consumo de recursos do projeto",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		asJSON, _ := cmd.Flags().GetBool("json")
		killZombies, _ := cmd.Flags().GetBool("kill-zombies")

		report, err := monitor.InspectProjectProcesses(dir, monitor.DefaultCommandRunner)
		if err != nil {
			return err
		}

		if killZombies {
			killed := 0
			for _, p := range report.Processes {
				// Identifica processos candidatos a zumbi (ex: subprocessos com PPID 1 ou estado zumbi)
				if p.ParentPID == 1 && p.Role != "tmux session (forge)" {
					if proc, err := os.FindProcess(p.PID); err == nil {
						if err := proc.Signal(syscall.SIGTERM); err == nil {
							fmt.Printf("Encerrado processo órfão PID %d (%s)\n", p.PID, p.Role)
							killed++
						}
					}
				}
			}
			if killed == 0 {
				fmt.Println("Nenhum processo órfão ou zumbi encontrado para encerrar.")
			}
			return nil
		}

		if asJSON {
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Print(monitor.FormatProcessTable(report))
		return nil
	},
}

func init() {
	psCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	psCmd.Flags().Bool("json", false, "Exibir resultado em formato JSON")
	psCmd.Flags().Bool("kill-zombies", false, "Encerrar processos órfãos ou presos ligados ao projeto")
	rootCmd.AddCommand(psCmd)
}
