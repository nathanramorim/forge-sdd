package main

import (
	"fmt"

	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/spf13/cobra"
)

// reportAgentModels aplica agent_models do forge.json aos agentes (feat 104-05) e informa o resultado.
// Valor inválido devolve erro sem gravar; com dryRun só mostra o que seria gravado.
func reportAgentModels(cmd *cobra.Command, dir string, dryRun bool) error {
	changed, err := scaffold.SyncAgentModels(dir, dryRun)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	verb := "atualizado"
	if dryRun {
		verb = "seria atualizado"
	}
	for _, f := range changed {
		fmt.Fprintf(out, "model: %s (%s)\n", verb, f)
	}
	return nil
}
