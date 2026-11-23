package main

import (
	"fmt"
	"slices"

	"github.com/forge-sdd/cli/internal/config"
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

// reportClaudeIntegration aplica (ou, com dryRun, lista) o hook SessionStart e o MCP forge-sdd antes
// de qualquer retorno "Nenhuma atualização necessária". Só para projetos com o agente Claude.
func reportClaudeIntegration(cmd *cobra.Command, dir string, agents []string, dryRun bool) {
	if !slices.Contains(agents, config.AgentClaude) {
		return
	}
	changed, warns := scaffold.SyncClaudeIntegration(dir, version, dryRun)
	verb := "instalado"
	if dryRun {
		verb = "seria instalado"
	}
	for _, c := range changed {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", c, verb)
	}
	for _, w := range warns {
		fmt.Fprintf(cmd.ErrOrStderr(), "aviso: %s\n", w)
	}
}
