package main

import (
	"fmt"

	"github.com/forge-sdd/cli/internal/notify"
	"github.com/spf13/cobra"
)

// avisoDepreciado devolve um PreRun que imprime o aviso de depreciação no
// stderr e deixa o comando seguir normalmente (feat-105-03, ADR-0021). Não usa
// o `Deprecated` do cobra porque ele esconde o comando do `--help`, e a
// decisão é só avisar: o comando continua listado e funcionando.
func avisoDepreciado(usar string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"AVISO: o comando %q está depreciado (estações obsoletas, ADR-0021); use %s. Sairá em versão futura, por ADR.\n",
			notify.Sanitize(cmd.CommandPath(), 80, false), notify.Sanitize(usar, 120, false))
	}
}
