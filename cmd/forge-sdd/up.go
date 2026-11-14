package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

var newLaunchEnv = pipeline.DefaultLaunchEnv

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Abre as 3 sessões do pipeline: forge-discovery, forge-builder e forge-reviewer",
	Long: `Abre três terminais nomeados no sistema operacional atual (macOS, Linux, WSL ou Windows):
  forge-discovery   sessão interativa do agente
  forge-builder     forge-sdd worker --role builder
  forge-reviewer    forge-sdd worker --role reviewer
O agente de cada sessão (claude, gemini, copilot ou codex) vem de --discovery, --builder e
--reviewer, depois de "sessions" em .forge/config/forge.json; --cli vale para as três sessões que
não tiverem escolha própria. --dry-run só imprime os comandos.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}
		cfg, found, err := config.LoadForgeConfig(dir)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("%w (sem .forge/config/forge.json não há pipeline para abrir)", pipeline.ErrNotPipelineMode)
		}
		cfg, err = applySessionFlags(cmd, cfg)
		if err != nil {
			return err
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		abs, err := absDir(dir)
		if err != nil {
			return err
		}
		env := newLaunchEnv(runtime.GOOS, abs, pipeline.ResolveBin())
		termFlag, _ := cmd.Flags().GetString("terminal")
		out := cmd.OutOrStdout()
		prompt := pipeline.TerminalPrompt{
			Env:         env,
			Geteuid:     os.Geteuid,
			Interactive: isInteractive() && !dry,
			Confirm:     survey.ConfirmYesNo,
			Runner:      pipeline.DefaultInstallRunner(),
			Out:         out,
			Remember:    pipeline.RememberNative(abs),
		}
		decision, err := pipeline.DecideTerminal(termFlag, cfg, prompt)
		if err != nil {
			return err
		}
		if decision.Terminal == config.TerminalTmux {
			plan, err := pipeline.BuildTmuxPlan(env, cfg, os.Getenv("TMUX") != "")
			if err != nil {
				return err
			}
			for _, step := range append(append([][]string{}, plan.Create...), plan.Attach) {
				fmt.Fprintln(out, strings.Join(quoteArgs(step), " "))
			}
			if dry {
				mode := cfg.SessionMode
				if mode == "" {
					mode = config.SessionModeHeadless
				}
				fmt.Fprintf(out, "[DRY-RUN] Nada foi aberto (sessão tmux %s, modo %s).\n", plan.Session, mode)
				return nil
			}
			nativeFallback := func() error {
				if cfg.SessionMode == config.SessionModeLive {
					fmt.Fprintln(out, "modo live requer tmux; caindo para headless nas janelas do sistema")
					cfg.SessionMode = config.SessionModeHeadless
				}
				natPlan, err := pipeline.LaunchPlanFor(env, cfg)
				if err != nil {
					return err
				}
				return pipeline.Launch(natPlan)
			}
			created, fellBack, err := pipeline.LaunchTerminalTmux(plan, pipeline.DefaultTmuxRunner(), decision.Auto, out, nativeFallback)
			if err == nil && !created && !fellBack {
				fmt.Fprintf(out, "A sessão %s já existia: reconectado.\n", plan.Session)
			}
			return err
		}
		if cfg.SessionMode == config.SessionModeLive {
			fmt.Fprintln(out, "modo live requer tmux; caindo para headless nas janelas do sistema")
			cfg.SessionMode = config.SessionModeHeadless
		}
		plan, err := pipeline.LaunchPlanFor(env, cfg)
		if err != nil {
			return err
		}
		for _, c := range plan {
			fmt.Fprintf(out, "%-16s %s\n", c.Title, strings.Join(quoteArgs(c.Argv), " "))
		}
		if dry {
			fmt.Fprintln(out, "[DRY-RUN] Nenhuma sessão foi aberta (modo headless).")
			return nil
		}
		return pipeline.Launch(plan)
	},
}

func quoteArgs(argv []string) []string {
	out := make([]string, len(argv))
	for i, a := range argv {
		if strings.ContainsAny(a, " \t\"'$;&|<>()") {
			out[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		} else {
			out[i] = a
		}
	}
	return out
}

func init() {
	upCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	upCmd.Flags().String("cli", "", "Agente das três sessões sem escolha própria (padrão: forge.json)")
	upCmd.Flags().String("discovery", "", "Agente da sessão forge-discovery: claude, gemini, copilot ou codex")
	upCmd.Flags().String("builder", "", "Agente da sessão forge-builder: claude, gemini, copilot ou codex")
	upCmd.Flags().String("reviewer", "", "Agente da sessão forge-reviewer: claude, gemini, copilot ou codex")
	upCmd.Flags().String("terminal", "", "Onde abrir: tmux (três painéis), native (janelas do sistema) ou auto (padrão: tmux se existir)")
	upCmd.Flags().Bool("dry-run", false, "Só imprime os comandos de abertura")
	rootCmd.AddCommand(upCmd)
}

func absDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolver %q: %w", dir, err)
	}
	return abs, nil
}

// applySessionFlags aplica --cli e --discovery/--builder/--reviewer sobre a configuração, sem gravar.
// Precedência: flag da sessão > sessions de forge.json > --cli > cli de forge.json.
func applySessionFlags(cmd *cobra.Command, cfg config.ForgeConfig) (config.ForgeConfig, error) {
	if c, _ := cmd.Flags().GetString("cli"); c != "" {
		if !config.KnownCLI(c) {
			return cfg, fmt.Errorf("--cli %q inválido (use %s)", c, strings.Join(config.OfferedCLIs(), ", "))
		}
		cfg.CLI = c
	}
	sessions := map[string]string{}
	for role, cli := range cfg.Sessions {
		sessions[role] = cli
	}
	for _, role := range config.SessionRoles() {
		if c, _ := cmd.Flags().GetString(role); c != "" {
			sessions[role] = c
		}
	}
	cfg.Sessions = sessions
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}
