package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/deprecation"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// avisoDepreciado devolve um PreRun que imprime o aviso de depreciação no
// stderr, registra o uso no CSV (feat-105-07) e deixa o comando seguir. Substituto
// e "desde" vêm do catálogo em internal/deprecation: depreciar outro comando =
// registrar uma entrada lá e ligar este helper no PreRun. Não usa o `Deprecated` do
// cobra porque ele esconde o comando do --help. Falha de registro nunca quebra o comando.
func avisoDepreciado() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		caminho := strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" ")
		ent, ok := deprecation.Buscar(caminho)
		if !ok {
			ent = deprecation.Entrada{Comando: caminho, Substituto: "(ver `forge --help`)", Desde: "ADR-0021"}
		}
		fmt.Fprintf(cmd.ErrOrStderr(),
			"AVISO: o comando %q está depreciado (estações obsoletas, ADR-0021); use %s. Sairá em versão futura, por ADR.\n",
			notify.Sanitize(cmd.CommandPath(), 80, false), notify.Sanitize(ent.Substituto, 120, false))
		if !ok {
			return // fora do catálogo: não grava
		}
		if cmd.Flags().Lookup("dry-run") != nil && cmd.Flags().Changed("dry-run") {
			return // --dry-run não grava (ARQ-05)
		}
		var flags []string
		cmd.Flags().Visit(func(f *pflag.Flag) { flags = append(flags, f.Name) })
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		err := deprecation.Registrar(dir, deprecation.Registro{
			Quando: time.Now(), Comando: ent.Comando,
			Contexto: deprecation.Contexto(cmd.Name(), flags), Agente: deprecation.Agente(os.Getenv),
			Substituto: ent.Substituto, Desde: ent.Desde, VersaoForge: version,
		})
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "AVISO: registro de uso do comando depreciado não gravado: %s\n",
				notify.Sanitize(err.Error(), 200, false))
		}
	}
}
