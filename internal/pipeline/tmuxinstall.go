package pipeline

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// InstallEnv é o ambiente do instalador do tmux; tudo injetável para os testes nunca instalarem
// pacote de verdade nem pedirem senha.
type InstallEnv struct {
	GOOS     string
	LookPath func(string) (string, error)
	// Geteuid devolve o usuário efetivo; nil conta como "não é root" (com sudo), para o resultado
	// não depender de quem roda o teste.
	Geteuid func() int
}

// DefaultInstallEnv usa o sistema, o PATH e o usuário reais.
func DefaultInstallEnv() InstallEnv {
	return InstallEnv{GOOS: runtime.GOOS, LookPath: exec.LookPath, Geteuid: os.Geteuid}
}

// installEnvFrom reaproveita o GOOS e o PATH do ambiente de abertura das sessões.
func installEnvFrom(env LaunchEnv, geteuid func() int) InstallEnv {
	return InstallEnv{GOOS: env.GOOS, LookPath: env.LookPath, Geteuid: geteuid}
}

func (e InstallEnv) has(name string) bool {
	if e.LookPath == nil {
		return false
	}
	_, err := e.LookPath(name)
	return err == nil
}

func (e InstallEnv) isRoot() bool { return e.Geteuid != nil && e.Geteuid() == 0 }

// installer é uma linha da tabela de gerenciadores, na ordem em que são tentados.
type installer struct {
	manager   string
	args      []string
	label     string // sistema ou distribuição, para a mensagem
	needsRoot bool   // pede sudo quando a pessoa não é root
}

var installTable = []installer{
	{"brew", []string{"install", "tmux"}, "Homebrew", false}, // o brew recusa rodar como root
	{"apt-get", []string{"install", "-y", "tmux"}, "Debian/Ubuntu", true},
	{"dnf", []string{"install", "-y", "tmux"}, "Fedora/RHEL", true},
	{"pacman", []string{"-S", "--noconfirm", "tmux"}, "Arch", true},
	{"apk", []string{"add", "tmux"}, "Alpine", true},
	{"zypper", []string{"--non-interactive", "install", "tmux"}, "openSUSE", true},
}

// InstallPlan escolhe o comando que instala o tmux pelo SO e pelos programas no PATH, na ordem da
// tabela. Com ok=true, desc nomeia o sistema ("macOS", "Debian/Ubuntu"...) e argv é o comando, sem
// shell, com sudo na frente quando a pessoa não é root e o gerenciador pede. Com ok=false (Windows
// nativo, sem gerenciador conhecido ou sem sudo), desc explica o motivo.
func InstallPlan(env InstallEnv) (argv []string, desc string, ok bool) {
	if env.GOOS == "windows" {
		return nil, "o tmux não existe no Windows nativo (use o WSL ou --terminal native)", false
	}
	for _, in := range installTable {
		if env.GOOS == "darwin" && in.manager != "brew" {
			continue
		}
		if !env.has(in.manager) {
			continue
		}
		cmd := append([]string{in.manager}, in.args...)
		if in.needsRoot && !env.isRoot() {
			if !env.has("sudo") {
				return nil, fmt.Sprintf("instalar com %s exige root, e o sudo não está no PATH", in.manager), false
			}
			cmd = append([]string{"sudo"}, cmd...)
		}
		label := in.label
		if in.manager == "brew" {
			label = "Homebrew"
			if env.GOOS == "darwin" {
				label = "macOS"
			}
		}
		return cmd, label, true
	}
	return nil, "não achei um gerenciador de pacotes conhecido (brew, apt-get, dnf, pacman, apk ou zypper)", false
}

// InstallCommandLine mostra o argv como a pessoa o digitaria (os argumentos da tabela são simples).
func InstallCommandLine(argv []string) string { return strings.Join(argv, " ") }

// TerminalPrompt é o que a decisão do terminal precisa; injetável nos testes.
type TerminalPrompt struct {
	Env     LaunchEnv
	Geteuid func() int
	Out     io.Writer
}

// TerminalDecision é onde as sessões abrem. Auto vale true quando a escolha não foi o tmux
// explícito (então o tmux, se falhar, pode cair para as janelas do sistema).
type TerminalDecision struct {
	Terminal string
	Auto     bool
}

// DecideTerminal resolve o terminal (ResolveTerminal). O tmux é opção (ADR-0024): pedido e ausente,
// avisa (com o comando de instalação, se houver) e segue em janelas do sistema, sem erro. Nunca
// instala nada sozinho.
func DecideTerminal(flag string, cfg config.ForgeConfig, p TerminalPrompt) (TerminalDecision, error) {
	out := p.Out
	if out == nil {
		out = io.Discard
	}
	choice := cfg.Terminal
	if flag != "" {
		choice = flag
	}
	term, err := ResolveTerminal(flag, cfg, p.Env)
	if err != nil {
		return TerminalDecision{}, err
	}
	d := TerminalDecision{Terminal: term, Auto: choice != config.TerminalTmux}
	if choice == config.TerminalTmux && term != config.TerminalTmux {
		fmt.Fprintln(out, "tmux não encontrado; abrindo em janelas do sistema.")
		if argv, _, ok := InstallPlan(installEnvFrom(p.Env, p.Geteuid)); ok {
			fmt.Fprintf(out, "Para usar o tmux, instale: %s\n", InstallCommandLine(argv))
		}
	}
	return d, nil
}

// TmuxAttachError diz que a sessão tmux foi criada e está de pé, mas não deu para conectar a ela.
// Não é motivo de contingência: abrir janelas do sistema duplicaria os agentes.
type TmuxAttachError struct {
	Session string
	Err     error
}

func (e *TmuxAttachError) Error() string {
	return fmt.Sprintf("a sessão %s está pronta, mas não consegui conectar a ela (%v; sem terminal interativo?). Rode: tmux attach -t =%s", e.Session, e.Err, e.Session)
}

func (e *TmuxAttachError) Unwrap() error { return e.Err }

// LaunchTerminalTmux abre o tmux e, em `auto`, cai para a contingência (fallback) quando o tmux não
// consegue montar a sessão: LaunchTmux já desfez o que criou, então o motivo vai para warn e as
// janelas do sistema abrem. `tmux` explícito mantém o erro, e falha só de conexão nunca cai.
func LaunchTerminalTmux(plan *TmuxPlan, r TmuxRunner, auto bool, warn io.Writer, fallback func() error) (created, fellBack bool, err error) {
	created, err = LaunchTmux(plan, r)
	if err == nil {
		return created, false, nil
	}
	var attachErr *TmuxAttachError
	if !auto || fallback == nil || errors.As(err, &attachErr) {
		return created, false, err
	}
	if warn != nil {
		fmt.Fprintf(warn, "O tmux não conseguiu abrir as sessões (%v). Abrindo em janelas do sistema.\n", err)
	}
	return false, true, fallback()
}
