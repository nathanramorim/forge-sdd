package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

var (
	tmuxNameBadRe = regexp.MustCompile(`[^A-Za-z0-9_-]+`)
	tmuxNameOKRe  = regexp.MustCompile(`^forge-[A-Za-z0-9_-]{1,40}-[0-9a-f]{6}$`)
)

// TmuxSessionName devolve forge-<projeto>-<hash>: o nome do projeto saneado (só [A-Za-z0-9_-], porque
// ":" e "." têm sentido no alvo do tmux; até 40 caracteres) e 6 hex do caminho absoluto, para duas
// pastas com o mesmo nome (/a/app e /b/app) não caírem na mesma sessão.
func TmuxSessionName(projectDir string) string {
	base := tmuxNameBadRe.ReplaceAllString(filepath.Base(projectDir), "-")
	base = strings.Trim(base, "-_")
	if len(base) > 40 {
		base = strings.Trim(base[:40], "-_")
	}
	if base == "" {
		base = "projeto"
	}
	sum := sha256.Sum256([]byte(filepath.Clean(projectDir)))
	return "forge-" + base + "-" + hex.EncodeToString(sum[:3])
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
	name := TmuxSessionName(env.Dir)
	if !tmuxNameOKRe.MatchString(name) {
		return nil, fmt.Errorf("nome de sessão tmux %q inválido", name)
	}
	// "=" pede correspondência exata: sem ele o tmux casa por prefixo e pode cair na sessão de outro projeto.
	exact := "=" + name
	win := exact + ":forge"
	titles := []string{SessionDiscovery, SessionBuilder, SessionReviewer}
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
		bldCtlCmd := fmt.Sprintf("%s worker --role builder --name %s --cli %s --deliver %s.1", shQuote(env.Bin), SessionBuilder, cfg.CLIFor(config.SessionRoleBuilder), win)
		bldCtlInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), "forge-builder-ctl", bldCtlCmd)
		revCtlCmd := fmt.Sprintf("%s worker --role reviewer --name %s --cli %s --deliver %s.2", shQuote(env.Bin), SessionReviewer, cfg.CLIFor(config.SessionRoleReviewer), win)
		revCtlInner := fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), "forge-reviewer-ctl", revCtlCmd)

		plan.Create = [][]string{
			{"tmux", "new-session", "-d", "-s", name, "-n", "forge", "-c", env.Dir, discInner},
			{"tmux", "set-window-option", "-t", win, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", win, "-c", env.Dir, bldInner},
			{"tmux", "split-window", "-v", "-t", win + ".1", "-c", env.Dir, revInner},
			{"tmux", "set-window-option", "-t", win, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", win, "pane-border-format", " #{pane_title} "},
			{"tmux", "set-window-option", "-t", win, "automatic-rename", "off"},
			{"tmux", "select-pane", "-t", win + ".0", "-T", SessionDiscovery},
			{"tmux", "select-pane", "-t", win + ".1", "-T", SessionBuilder},
			{"tmux", "select-pane", "-t", win + ".2", "-T", SessionReviewer},
			{"tmux", "select-pane", "-t", win + ".0"},
			{"tmux", "new-window", "-d", "-n", "forge-ctl", "-t", exact, "-c", env.Dir, bldCtlInner},
			{"tmux", "set-window-option", "-t", winCtl, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", winCtl, "-c", env.Dir, revCtlInner},
			{"tmux", "set-window-option", "-t", winCtl, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", winCtl, "pane-border-format", " #{pane_title} "},
			{"tmux", "set-window-option", "-t", winCtl, "automatic-rename", "off"},
			{"tmux", "select-pane", "-t", winCtl + ".0", "-T", "forge-builder-ctl"},
			{"tmux", "select-pane", "-t", winCtl + ".1", "-T", "forge-reviewer-ctl"},
		}
	} else {
		plan.Create = [][]string{
			{"tmux", "new-session", "-d", "-s", name, "-n", "forge", "-c", env.Dir, sessions[0].Inner},
			// Quem usa pane-base-index 1 no tmux.conf teria painéis 1, 2 e 3: fixa 0 antes de endereçá-los.
			{"tmux", "set-window-option", "-t", win, "pane-base-index", "0"},
			{"tmux", "split-window", "-h", "-t", win, "-c", env.Dir, sessions[1].Inner},
			{"tmux", "split-window", "-v", "-t", win + ".1", "-c", env.Dir, sessions[2].Inner},
			{"tmux", "set-window-option", "-t", win, "pane-border-status", "top"},
			{"tmux", "set-window-option", "-t", win, "pane-border-format", " #{pane_title} "},
			{"tmux", "set-window-option", "-t", win, "automatic-rename", "off"},
		}
		for i, t := range titles {
			plan.Create = append(plan.Create, []string{"tmux", "select-pane", "-t", fmt.Sprintf("%s.%d", win, i), "-T", t})
		}
		plan.Create = append(plan.Create, []string{"tmux", "select-pane", "-t", win + ".0"})
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
