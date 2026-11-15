package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
)

// TmuxProjectName extrai o nome do projeto saneado para identificação visual no tmux.
func TmuxProjectName(projectDir string) string {
	return layout.TmuxProjectName(projectDir)
}

// TmuxSessionName devolve forge-<projeto>-<hash>: o nome do projeto saneado e 6 hex do caminho absoluto.
func TmuxSessionName(projectDir string) string {
	return layout.TmuxSessionName(projectDir)
}

// IsValidTmuxSessionName verifica se o nome da sessão segue o padrão forge-<projeto>-<hash>.
func IsValidTmuxSessionName(name string) bool {
	return layout.IsValidTmuxSessionName(name)
}

// ResolveTerminal escolhe entre tmux e as janelas do sistema. Precedência: flag, forge.json, auto
// (tmux quando está no PATH e o sistema não é Windows nativo).
func ResolveTerminal(flag string, cfg config.ForgeConfig, env LaunchEnv) (string, error) {
	choice := cfg.Terminal
	if flag != "" {
		choice = flag
	}
	if choice == "" {
		choice = config.TerminalAuto
	}
	if !config.ValidTerminal(choice) {
		return "", fmt.Errorf("terminal %q inválido (use auto, tmux ou native)", choice)
	}
	canTmux := env.GOOS != "windows" && has(env, "tmux")
	switch choice {
	case config.TerminalTmux:
		if env.GOOS == "windows" {
			return "", fmt.Errorf("tmux não existe no Windows nativo (use WSL ou --terminal native)")
		}
		if !canTmux {
			return "", fmt.Errorf("tmux não está no PATH (instale com brew install tmux ou apt install tmux, ou use --terminal native)")
		}
		return config.TerminalTmux, nil
	case config.TerminalNative:
		return config.TerminalNative, nil
	}
	if canTmux {
		return config.TerminalTmux, nil
	}
	return config.TerminalNative, nil
}

// TmuxPlan são os comandos que montam a sessão (Create), e o que conecta a ela (Attach). Tudo em
// argv explícito; o tmux roda o comando de cada painel pelo shell, por isso ele já chega quotado.
type TmuxPlan struct {
	Session string
	Create  [][]string
	Attach  []string
}

// BuildTmuxPlan monta os três painéis: Discovery grande à esquerda, Builder e Reviewer empilhados
// à direita, cada um com o nome da sessão no título.
func BuildTmuxPlan(env LaunchEnv, cfg config.ForgeConfig, insideTmux bool) (*TmuxPlan, error) {
	if env.GOOS == "windows" {
		return nil, fmt.Errorf("tmux não existe no Windows nativo")
	}
	if !filepath.IsAbs(env.Dir) {
		return nil, fmt.Errorf("o diretório do projeto precisa ser absoluto para o tmux (recebeu %q)", env.Dir)
	}
	sessions, err := SessionsFor(env, cfg)
	if err != nil {
		return nil, err
	}
	proj := TmuxProjectName(env.Dir)
	name := TmuxSessionName(env.Dir)
	if !layout.IsValidTmuxSessionName(name) {
		return nil, fmt.Errorf("nome de sessão tmux %q inválido", name)
	}
	// "=" pede correspondência exata: sem ele o tmux casa por prefixo e pode cair na sessão de outro projeto.
	exact := "=" + name
	winName := "forge:" + proj
	win := exact + ":" + winName
	borderFormat := fmt.Sprintf(" [ %s ❯ #{pane_title} ] ", proj)

	tokMetrics, _ := monitor.CalculateTokenMetrics(env.Dir)
	var discTokens, bldTokens, revTokens, totalTokens int
	if tokMetrics != nil {
		discTokens = tokMetrics.Roles[layout.RoleDiscovery].TotalTokens
		bldTokens = tokMetrics.Roles[layout.RoleBuilder].TotalTokens
		revTokens = tokMetrics.Roles[layout.RoleReviewer].TotalTokens
		totalTokens = tokMetrics.TotalTokens
	}

	discTitle := SessionDiscovery
	bldTitle := SessionBuilder
	revTitle := SessionReviewer
	if discTokens > 0 {
		discTitle = fmt.Sprintf("%s │ %s", SessionDiscovery, monitor.FormatTokenBadge(discTokens))
	}
	if bldTokens > 0 {
		bldTitle = fmt.Sprintf("%s │ %s", SessionBuilder, monitor.FormatTokenBadge(bldTokens))
	}
	if revTokens > 0 {
		revTitle = fmt.Sprintf("%s │ %s", SessionReviewer, monitor.FormatTokenBadge(revTokens))
	}
	titles := []string{discTitle, bldTitle, revTitle}

	statusRight := fmt.Sprintf("Tokens Totais: %s (Disc: %s · Bld: %s · Rev: %s)",
		monitor.FormatTokensShort(totalTokens),
		monitor.FormatTokensShort(discTokens),
		monitor.FormatTokensShort(bldTokens),
		monitor.FormatTokensShort(revTokens),
	)

	plan := &TmuxPlan{Session: name}

	if cfg.SessionMode == config.SessionModeLive {
		discCLI, err := interactiveLineForRole(false, cfg, config.SessionRoleDiscovery, SessionDiscovery, "lider")
		if err != nil {
			return nil, err
		}
		builderCLI, err := interactiveLineForRole(false, cfg, config.SessionRoleBuilder, SessionBuilder, "builder")
		if err != nil {
			return nil, err
		}
		revCLI, err := interactiveLineForRole(false, cfg, config.SessionRoleReviewer, SessionReviewer, "reviewer")
		if err != nil {
			return nil, err
		}
		discInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), SessionDiscovery, discCLI)
		bldInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), SessionBuilder, builderCLI)
		revInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), SessionReviewer, revCLI)

		winCtl := exact + ":forge-ctl"
		bldCtlCmd := fmt.Sprintf("%s worker --role builder --name %s --cli %s --deliver %s", shQuote(env.Bin), SessionBuilder, cfg.CLIFor(config.SessionRoleBuilder), shQuote(win+".1"))
		bldCtlInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), "forge-builder-ctl", bldCtlCmd)
		revCtlCmd := fmt.Sprintf("%s worker --role reviewer --name %s --cli %s --deliver %s", shQuote(env.Bin), SessionReviewer, cfg.CLIFor(config.SessionRoleReviewer), shQuote(win+".2"))
		revCtlInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), "forge-reviewer-ctl", revCtlCmd)

		plan.Create = [][]string{
			{"tmux", "new-session", "-d", "-s", name, "-n", winName, "-c", env.Dir, discInner},
			{"tmux", "set-option", "-t", name, "status-left", fmt.Sprintf("#[fg=colour16,bg=colour48,bold] 📦 %s #[default] ", proj)},
			{"tmux", "set-option", "-t", name, "status-left-length", "60"},
			{"tmux", "set-window-option", "-t", win, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", win, "-c", env.Dir, bldInner},
			{"tmux", "split-window", "-v", "-t", win + ".1", "-c", env.Dir, revInner},
			{"tmux", "set-window-option", "-t", win, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", win, "pane-border-format", borderFormat},
			{"tmux", "set-window-option", "-t", win, "automatic-rename", "off"},
			{"tmux", "bind-key", "q", "display-popup", "-w", "85%", "-h", "85%", "-E", fmt.Sprintf("%s queue --interactive", shQuote(env.Bin))},
			{"tmux", "select-pane", "-t", win + ".0", "-T", discTitle},
			{"tmux", "select-pane", "-t", win + ".1", "-T", bldTitle},
			{"tmux", "select-pane", "-t", win + ".2", "-T", revTitle},
			{"tmux", "select-pane", "-t", win + ".0"},
			{"tmux", "new-window", "-d", "-n", "forge-ctl", "-t", exact, "-c", env.Dir, bldCtlInner},
			{"tmux", "set-window-option", "-t", winCtl, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", winCtl, "-c", env.Dir, revCtlInner},
			{"tmux", "set-window-option", "-t", winCtl, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", winCtl, "pane-border-format", borderFormat},
			{"tmux", "set-window-option", "-t", winCtl, "automatic-rename", "off"},
			{"tmux", "select-pane", "-t", winCtl + ".0", "-T", "forge-builder-ctl"},
			{"tmux", "select-pane", "-t", winCtl + ".1", "-T", "forge-reviewer-ctl"},
		}
		if totalTokens > 0 {
			plan.Create = append(plan.Create,
				[]string{"tmux", "set-option", "-t", name, "status-right", statusRight},
				[]string{"tmux", "set-option", "-t", name, "status-right-length", "100"},
			)
		}
	} else {
		plan.Create = [][]string{
			{"tmux", "new-session", "-d", "-s", name, "-n", winName, "-c", env.Dir, sessions[0].Inner},
			{"tmux", "set-option", "-t", name, "status-left", fmt.Sprintf("#[fg=colour16,bg=colour48,bold] 📦 %s #[default] ", proj)},
			{"tmux", "set-option", "-t", name, "status-left-length", "60"},
			// Quem usa pane-base-index 1 no tmux.conf teria painéis 1, 2 e 3: fixa 0 antes de endereçá-los.
			{"tmux", "set-window-option", "-t", win, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", win, "-c", env.Dir, sessions[1].Inner},
			{"tmux", "split-window", "-v", "-t", win + ".1", "-c", env.Dir, sessions[2].Inner},
			{"tmux", "set-window-option", "-t", win, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", win, "pane-border-format", borderFormat},
			{"tmux", "set-window-option", "-t", win, "automatic-rename", "off"},
			{"tmux", "bind-key", "q", "display-popup", "-w", "85%", "-h", "85%", "-E", fmt.Sprintf("%s queue --interactive", shQuote(env.Bin))},
		}
		for i, t := range titles {
			plan.Create = append(plan.Create, []string{"tmux", "select-pane", "-t", fmt.Sprintf("%s.%d", win, i), "-T", t})
		}
		plan.Create = append(plan.Create, []string{"tmux", "select-pane", "-t", win + ".0"})
		if totalTokens > 0 {
			plan.Create = append(plan.Create,
				[]string{"tmux", "set-option", "-t", name, "status-right", statusRight},
				[]string{"tmux", "set-option", "-t", name, "status-right-length", "100"},
			)
		}
	}
	if insideTmux {
		plan.Attach = []string{"tmux", "switch-client", "-t", exact}
	} else {
		plan.Attach = []string{"tmux", "attach-session", "-t", exact}
	}
	return plan, nil
}

// TmuxRunner executa um argv; injetável nos testes. attach liga o terminal da pessoa ao tmux.
type TmuxRunner interface {
	Run(argv []string) error
	Attach(argv []string) error
}

type execTmux struct{}

func (execTmux) Run(argv []string) error { return exec.Command(argv[0], argv[1:]...).Run() }
func (execTmux) Attach(argv []string) error {
	c := exec.Command(argv[0], argv[1:]...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

// DefaultTmuxRunner usa o tmux real.
func DefaultTmuxRunner() TmuxRunner { return execTmux{} }

// LaunchTmux cria a sessão (se ainda não existir) e conecta a ela. Sessão existente só conecta, sem
// duplicar painéis. A limpeza em falha só remove o que esta execução criou: se o primeiro passo
// (new-session) falha porque outra execução criou a sessão no mesmo instante, só conecta.
func LaunchTmux(plan *TmuxPlan, r TmuxRunner) (created bool, err error) {
	exact := "=" + plan.Session
	if r.Run([]string{"tmux", "has-session", "-t", exact}) != nil {
		for i, step := range plan.Create {
			if err := r.Run(step); err != nil {
				if i == 0 {
					if r.Run([]string{"tmux", "has-session", "-t", exact}) == nil {
						return false, r.Attach(plan.Attach) // corrida: outra execução criou a sessão
					}
					return false, fmt.Errorf("tmux %s: %w", step[1], err) // nada foi criado: nada a limpar
				}
				_ = r.Run([]string{"tmux", "kill-session", "-t", exact})
				return false, fmt.Errorf("tmux %s: %w", step[1], err)
			}
		}
		created = true
	}
	if err := r.Attach(plan.Attach); err != nil {
		return created, &TmuxAttachError{Session: plan.Session, Err: err}
	}
	return created, nil
}

// OpenLiveStatusInTmux anexa ou abre o painel vivo em uma janela/painel dedicado do tmux.
func OpenLiveStatusInTmux(dir string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux não encontrado no PATH (instale com brew install tmux ou apt install tmux)")
	}

	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	sess := TmuxSessionName(absDir)
	bin := ResolveBin()
	statusCmd := fmt.Sprintf("%s status --live", shQuote(bin))

	insideTmux := os.Getenv("TMUX") != ""

	// Verifica se a sessão principal forge-<projeto>-<hash> já existe
	hasSess := exec.Command(tmuxPath, "has-session", "-t", "="+sess).Run() == nil

	if hasSess {
		// Sessão existe: verifica se a janela forge-status já existe nela
		checkWin := exec.Command(tmuxPath, "list-windows", "-t", "="+sess, "-F", "#{window_name}")
		outBytes, _ := checkWin.Output()
		hasWin := strings.Contains(string(outBytes), "forge-status")
		if !hasWin {
			// Cria a janela forge-status na sessão existente
			createWin := exec.Command(tmuxPath, "new-window", "-d", "-n", "forge-status", "-t", "="+sess, "-c", absDir, statusCmd)
			_ = createWin.Run()
		}

		if insideTmux {
			// Já estamos no tmux: seleciona a janela forge-status
			selectCmd := exec.Command(tmuxPath, "select-window", "-t", "="+sess+":forge-status")
			return selectCmd.Run()
		}

		// Fora do tmux: anexa na janela forge-status da sessão
		attachCmd := exec.Command(tmuxPath, "attach-session", "-t", "="+sess+":forge-status")
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr
		return attachCmd.Run()
	}

	// Se não existe a sessão do projeto:
	if insideTmux {
		// Abre janela no tmux atual
		winCmd := exec.Command(tmuxPath, "new-window", "-n", "forge-status", "-c", absDir, statusCmd)
		return winCmd.Run()
	}

	// Fora do tmux: cria sessão dedicada forge-status e conecta
	statusSess := sess + "-status"
	createSess := exec.Command(tmuxPath, "new-session", "-s", statusSess, "-n", "forge-status", "-c", absDir, statusCmd)
	createSess.Stdin = os.Stdin
	createSess.Stdout = os.Stdout
	createSess.Stderr = os.Stderr
	return createSess.Run()
}

// TmuxPopupQueueCommand monta o comando tmux para abrir a fila em popup modal flutuante.
func TmuxPopupQueueCommand(binPath string) []string {
	if binPath == "" {
		binPath = ResolveBin()
	}
	cmd := fmt.Sprintf("%s queue --interactive", shQuote(binPath))
	return []string{"tmux", "display-popup", "-w", "85%", "-h", "85%", "-E", cmd}
}

// OpenPopupQueueInTmux executa o popup modal da fila na sessão ativa do tmux.
func OpenPopupQueueInTmux(dir string) error {
	bin := ResolveBin()
	argv := TmuxPopupQueueCommand(bin)
	return exec.Command(argv[0], argv[1:]...).Run()
}
