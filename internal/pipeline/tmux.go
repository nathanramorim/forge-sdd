package pipeline

import (
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
	tmuxNameOKRe  = regexp.MustCompile(`^forge-[A-Za-z0-9_-]{1,40}$`)
)

// TmuxSessionName devolve forge-<projeto> saneado: só [A-Za-z0-9_-] (":" e "." têm sentido no
// alvo do tmux) e até 40 caracteres do nome do projeto.
func TmuxSessionName(projectDir string) string {
	base := tmuxNameBadRe.ReplaceAllString(filepath.Base(projectDir), "-")
	base = strings.Trim(base, "-_")
	if len(base) > 40 {
		base = strings.Trim(base[:40], "-_")
	}
	if base == "" {
		base = "projeto"
	}
	return "forge-" + base
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
	sessions, err := SessionsFor(env, cfg)
	if err != nil {
		return nil, err
	}
	name := TmuxSessionName(env.Dir)
	if !tmuxNameOKRe.MatchString(name) {
		return nil, fmt.Errorf("nome de sessão tmux %q inválido", name)
	}
	win := name + ":forge"
	titles := []string{SessionDiscovery, SessionBuilder, SessionReviewer}
	plan := &TmuxPlan{Session: name}
	plan.Create = [][]string{
		{"tmux", "new-session", "-d", "-s", name, "-n", "forge", "-c", env.Dir, sessions[0].Inner},
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
	if insideTmux {
		plan.Attach = []string{"tmux", "switch-client", "-t", name}
	} else {
		plan.Attach = []string{"tmux", "attach-session", "-t", name}
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
// duplicar painéis.
func LaunchTmux(plan *TmuxPlan, r TmuxRunner) (created bool, err error) {
	if r.Run([]string{"tmux", "has-session", "-t", plan.Session}) != nil {
		for _, step := range plan.Create {
			if err := r.Run(step); err != nil {
				_ = r.Run([]string{"tmux", "kill-session", "-t", plan.Session})
				return false, fmt.Errorf("tmux %s: %w", step[1], err)
			}
		}
		created = true
	}
	return created, r.Attach(plan.Attach)
}
