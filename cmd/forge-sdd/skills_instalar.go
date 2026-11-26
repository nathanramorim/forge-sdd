package main

import (
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/spf13/cobra"
)

var skillsDestinosCmd = &cobra.Command{
	Use:   "destinos",
	Short: "Lista onde instalar uma skill ou plugin (projeto aberto e subprojetos com forge)",
	Long: `Imprime os destinos válidos de instalação (ADR-0025). Com --tipo e --slug, informa também
"já disponível em <caminho>" quando o projeto aberto ou o hub acima já tem a skill ou o plugin.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		tipo, _ := cmd.Flags().GetString("tipo")
		slug, _ := cmd.Flags().GetString("slug")
		out := cmd.OutOrStdout()
		if tipo != "" || slug != "" {
			if onde, ok := scaffold.DisponivelViaHub(dir, tipo, slug); ok {
				fmt.Fprintf(out, "já disponível em %s\n", onde)
				return nil
			}
		}
		destinos, err := scaffold.ListarDestinosInstalacao(dir)
		if err != nil {
			return err
		}
		for _, d := range destinos {
			fmt.Fprintln(out, d)
		}
		return nil
	},
}

var skillsInstalarCmd = &cobra.Command{
	Use:   "instalar <skill|plugin> <slug> [--de <diretório-origem>] [--destino <dir>]",
	Short: "Instala uma skill ou plugin no projeto aberto (opcional, sem rede, sem ~/.claude)",
	Long: `Copia a skill ou o plugin de --de para .agents/skills/<slug> ou .agents/plugins/<slug> (ADR-0025).
Plugin que é mod do Claude Code (traz módulo de ganchos ou .claude-plugin/forge.json) só é instalado
com o contrato completo da ADR-0030: manifesto, permissões e versões mínimas. Manifesto incompleto é
recusado antes de gravar qualquer arquivo. Sem --de, instala um mod que viaja no binário
(forge-painel).`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		origem, _ := cmd.Flags().GetString("de")
		destino, _ := cmd.Flags().GetString("destino")
		if destino == "" {
			destino = dir
		}
		var arquivos map[string][]byte
		var err error
		if origem == "" {
			// Sem --de, só um mod que viaja no binário pode ser instalado (ADR-0030).
			embutido, ok := scaffold.ModEmbutido(args[1])
			if args[0] != scaffold.TipoPlugin || !ok {
				return fmt.Errorf("--de <diretório-origem> é obrigatório (sem ele, só os mods embutidos: %s)", strings.Join(scaffold.ModsEmbutidos(), ", "))
			}
			arquivos = embutido
		} else if arquivos, err = scaffold.LerOrigemInstalacao(origem); err != nil {
			return err
		}
		var res scaffold.Resultado
		switch args[0] {
		case scaffold.TipoSkill:
			res, err = scaffold.InstalarSkill(dir, destino, args[1], arquivos)
		case scaffold.TipoPlugin:
			res, err = scaffold.InstalarPlugin(dir, destino, args[1], arquivos)
		default:
			return fmt.Errorf("tipo %q inválido: use skill ou plugin", args[0])
		}
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if res.DisponivelEm != "" {
			fmt.Fprintf(out, "já disponível em %s; nada foi instalado\n", res.DisponivelEm)
			return nil
		}
		fmt.Fprintf(out, "✓ %s %q: %d gravado(s), %d preservado(s)\n", args[0], args[1], len(res.Gravados), len(res.Preservados))
		for _, p := range res.Preservados {
			fmt.Fprintf(out, "  preservado: %s\n", p)
		}
		return nil
	},
}

func init() {
	skillsDestinosCmd.Flags().String("dir", ".", "Projeto aberto")
	skillsDestinosCmd.Flags().String("tipo", "", "skill ou plugin (com --slug, informa se já está disponível)")
	skillsDestinosCmd.Flags().String("slug", "", "Slug a consultar")
	skillsInstalarCmd.Flags().String("dir", ".", "Projeto aberto")
	skillsInstalarCmd.Flags().String("de", "", "Diretório de origem já obtido fora do Go (opcional para os mods embutidos)")
	skillsInstalarCmd.Flags().String("destino", "", "Destino de ListarDestinosInstalacao (padrão: o projeto aberto)")
	skillsCmd.AddCommand(skillsDestinosCmd)
	skillsCmd.AddCommand(skillsInstalarCmd)
}
