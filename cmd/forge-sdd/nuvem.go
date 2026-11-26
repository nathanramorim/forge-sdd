package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/spf13/cobra"
)

var nuvemCmd = &cobra.Command{
	Use:   "nuvem",
	Short: "Roda o forge em sessões do Claude Code na nuvem, sem gastar disco nem memória da sua máquina",
	Long: `Liga o forge nas sessões do Claude Code na nuvem. "forge nuvem ativar" grava em nuvem do
.forge/config/forge.json e escreve no repositório um hook de início de sessão
(.claude/hooks/forge-nuvem.sh). Na nuvem o hook instala o forge e sobe o Builder e o Reviewer em
segundo plano; no seu computador ele não faz nada. Funciona para um repositório só ou para um hub
(--repos lista os repositórios que a sessão precisa ter).`,
}

var nuvemAtivarCmd = &cobra.Command{
	Use:   "ativar",
	Short: "Liga a nuvem e grava o hook no repositório",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		cfg, err := carregarPipeline(dir)
		if err != nil {
			return err
		}
		n := cfg.NuvemOuPadrao()
		n.Ativo = true
		if cmd.Flags().Changed("repos") {
			repos, _ := cmd.Flags().GetString("repos")
			n.Repos = limparLista(strings.Split(repos, ","))
		}
		if cmd.Flags().Changed("instalar") {
			n.Instalar, _ = cmd.Flags().GetString("instalar")
		}
		if cmd.Flags().Changed("ref") {
			n.Ref, _ = cmd.Flags().GetString("ref")
		}
		return gravarNuvem(cmd.OutOrStdout(), dir, cfg, n)
	},
}

var nuvemDesativarCmd = &cobra.Command{
	Use:   "desativar",
	Short: "Desliga a nuvem e remove o hook do repositório",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		cfg, err := carregarPipeline(dir)
		if err != nil {
			return err
		}
		n := cfg.NuvemOuPadrao()
		n.Ativo = false
		return gravarNuvem(cmd.OutOrStdout(), dir, cfg, n)
	},
}

var nuvemStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra a configuração da nuvem e o que falta fazer",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		cfg, err := carregarPipeline(dir)
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), formatNuvemStatus(dir, cfg))
		return nil
	},
}

var nuvemIniciarCmd = &cobra.Command{
	Use:    "iniciar",
	Short:  "Chamado pelo hook na nuvem: sobe o Builder e o Reviewer em segundo plano e resume a esteira",
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		forcar, _ := cmd.Flags().GetBool("forcar")
		return nuvemIniciar(cmd.OutOrStdout(), dir, forcar, os.Getenv)
	},
}

func carregarPipeline(dir string) (config.ForgeConfig, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return cfg, err
	}
	if !found {
		return cfg, fmt.Errorf("%w: rode forge pipeline init antes", pipeline.ErrNotPipelineMode)
	}
	return cfg, nil
}

func limparLista(in []string) []string {
	var out []string
	vistos := map[string]bool{}
	for _, r := range in {
		r = strings.TrimSpace(r)
		if r != "" && !vistos[r] {
			vistos[r] = true
			out = append(out, r)
		}
	}
	return out
}

// gravarNuvem valida, salva em forge.json e aplica o hook e o script no repositório.
func gravarNuvem(out io.Writer, dir string, cfg config.ForgeConfig, n config.NuvemConfig) error {
	if err := n.Validate(); err != nil {
		return err
	}
	cfg.Nuvem = &n
	if err := config.SaveForgeConfig(dir, cfg); err != nil {
		return err
	}
	mudou, err := scaffold.AplicarNuvem(dir, cfg, version, false)
	for _, m := range mudou {
		fmt.Fprintf(out, "✓ %s\n", m)
	}
	if err != nil {
		return err
	}
	if !n.Ativo {
		fmt.Fprintln(out, "Nuvem desligada. Faça commit da remoção para valer nas próximas sessões.")
		return nil
	}
	fmt.Fprintf(out, "✓ nuvem ligada em .forge/config/forge.json (instala por %s, %s)\n\n", n.InstalarEfetivo(), scaffold.NuvemRef(n, version))
	fmt.Fprint(out, pipeline.ChecklistNuvem(n, nomeDoProjeto(dir)))
	return nil
}

func nomeDoProjeto(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return filepath.Base(abs)
}

func formatNuvemStatus(dir string, cfg config.ForgeConfig) string {
	var b strings.Builder
	n := cfg.NuvemOuPadrao()
	if !n.Ativo {
		b.WriteString("Nuvem: desligada. Ligue com forge nuvem ativar (ou pelo menu › Configurações).\n")
		return b.String()
	}
	fmt.Fprintf(&b, "Nuvem: ligada · instala por %s (%s)\n", n.InstalarEfetivo(), scaffold.NuvemRef(n, version))
	if len(n.Repos) > 0 {
		fmt.Fprintf(&b, "Hub: %s\n", strings.Join(n.Repos, ", "))
	}
	hook := "✗ falta (rode forge nuvem ativar)"
	if scaffold.NuvemHookInstalado(dir) {
		if _, err := os.Stat(filepath.Join(dir, scaffold.NuvemScriptRel)); err == nil {
			hook = "✓ instalado"
		}
	}
	fmt.Fprintf(&b, "Hook no repositório: %s\n", hook)
	if pipeline.NaNuvem(nil) {
		b.WriteString("Esta sessão: na nuvem\n")
		if falta := pipeline.ReposFaltando(dir, n.Repos); len(falta) > 0 {
			fmt.Fprintf(&b, "✗ repositórios do hub ausentes nesta sessão: %s\n", strings.Join(falta, ", "))
		}
	} else {
		b.WriteString("Esta sessão: local (o hook não faz nada aqui)\n")
	}
	b.WriteString("\n" + pipeline.ChecklistNuvem(n, nomeDoProjeto(dir)))
	return b.String()
}

// nuvemIniciar é o que o hook roda depois de instalar o forge. A saída vira contexto da sessão,
// então é curta. Nunca falha a sessão: problemas viram aviso.
func nuvemIniciar(out io.Writer, dir string, forcar bool, getenv func(string) string) error {
	if !forcar && !pipeline.NaNuvem(getenv) {
		fmt.Fprintln(out, "forge nuvem: fora da nuvem, nada a fazer.")
		return nil
	}
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		fmt.Fprintf(out, "forge nuvem: %v\n", err)
		return nil
	}
	if !found || !cfg.NuvemAtiva() {
		return nil
	}
	abs, err := absDir(dir)
	if err != nil {
		return err
	}
	if falta := pipeline.ReposFaltando(abs, cfg.NuvemOuPadrao().Repos); len(falta) > 0 {
		fmt.Fprintf(out, "⚠ forge nuvem: faltam os repositórios %s nesta sessão; adicione-os ao ambiente no claude.ai.\n", strings.Join(falta, ", "))
	}
	if _, err := scaffold.SincronizarModEmbutido(abs, pipeline.PainelPluginSlug); err != nil {
		fmt.Fprintf(out, "⚠ forge nuvem: mod %s não instalado: %v\n", pipeline.PainelPluginSlug, err)
	}
	env := newLaunchEnv(runtime.GOOS, abs, pipeline.ResolveBin())
	plan, err := pipeline.BuildPainelPlan(env, cfg)
	if err != nil {
		if errors.Is(err, pipeline.ErrPainelSemMod) {
			err = fmt.Errorf("mod %s ausente", pipeline.PainelPluginSlug)
		}
		fmt.Fprintf(out, "⚠ forge nuvem: esteira não iniciada: %v\n", err)
		return nil
	}
	res, err := pipeline.LaunchPainel(plan, abs, newPainelStarter)
	var subiu []string
	for _, w := range plan.Workers {
		if pid, ok := res.Started[w.Title]; ok {
			subiu = append(subiu, fmt.Sprintf("%s (pid %d)", w.Title, pid))
		} else if pid, ok := res.Already[w.Title]; ok {
			subiu = append(subiu, fmt.Sprintf("%s já rodando (pid %d)", w.Title, pid))
		}
	}
	if err != nil {
		fmt.Fprintf(out, "⚠ forge nuvem: %v\n", err)
	}
	fmt.Fprintf(out, "Forge na nuvem: %s em segundo plano. Cada task aprovada vai para o GitHub pela branch da entrega; a máquina da nuvem é descartável.\n", strings.Join(subiu, ", "))
	fmt.Fprint(out, pipeline.FormatDiagnostico(abs, time.Now()))
	return nil
}

func init() {
	for _, c := range []*cobra.Command{nuvemAtivarCmd, nuvemDesativarCmd, nuvemStatusCmd, nuvemIniciarCmd} {
		c.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
		nuvemCmd.AddCommand(c)
	}
	nuvemAtivarCmd.Flags().String("repos", "", "Repositórios do hub que a sessão precisa ter (nome ou dono/nome, separados por vírgula)")
	nuvemAtivarCmd.Flags().String("instalar", "", "Como instalar o forge na nuvem: npm (padrão) ou git (compila da branch)")
	nuvemAtivarCmd.Flags().String("ref", "", "Versão npm ou branch git (padrão: a versão deste forge, ou main no git)")
	nuvemIniciarCmd.Flags().Bool("forcar", false, "Roda mesmo fora da nuvem (testes)")
	rootCmd.AddCommand(nuvemCmd)
}
