package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/forge-sdd/cli/internal/studio"
	"github.com/spf13/cobra"
)

var studioCmd = &cobra.Command{
	Use:   "studio",
	Short: "Inicia o painel visual desacoplado em tempo real (Forge Studio)",
	Long: `Inicia um servidor local HTTP/SSE leve e estritamente Read-Only que projeta em tempo real 
o estado operacional das 5 estações do Forge-SDD, seus personagens compactos, a esteira de handoff e o Inspetor de Governança.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		openBrowser, _ := cmd.Flags().GetBool("open")

		srv, err := studio.NewServer(studio.Config{
			Dir:  dir,
			Host: host,
			Port: port,
		})
		if err != nil {
			return err
		}

		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Start()
		}()

		url := fmt.Sprintf("http://%s:%d", host, port)
		fmt.Printf("\n🎨 Forge-SDD Studio rodando em modo Read-Only!\n")
		fmt.Printf("   URL: %s\n\n", url)
		fmt.Println("Pressione Ctrl+C para encerrar o Studio a qualquer momento.")

		if openBrowser {
			openURL(url)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-errChan:
			return err
		case <-sigChan:
			fmt.Println("\nEncerrando Forge-SDD Studio...")
			return srv.Stop()
		}
	},
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	studioCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	studioCmd.Flags().String("host", "127.0.0.1", "Interface de rede para escuta (padrão local restrito 127.0.0.1)")
	studioCmd.Flags().Int("port", 4321, "Porta HTTP do Studio")
	studioCmd.Flags().Bool("open", true, "Abre o navegador automaticamente na inicialização")

	rootCmd.AddCommand(studioCmd)
}
