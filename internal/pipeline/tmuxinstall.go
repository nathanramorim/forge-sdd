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

// InstallRunner executa o instalador. Run herda stdin, stdout e stderr (a senha do sudo é pedida no
// terminal da pessoa); Probe roda em silêncio, só para saber se o comando funciona.
type InstallRunner interface {
	Run(argv []string) error
	Probe(argv []string) error
}

type execInstall struct{}

func (execInstall) Run(argv []string) error {
	c := exec.Command(argv[0], argv[1:]...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

func (execInstall) Probe(argv []string) error { return exec.Command(argv[0], argv[1:]...).Run() }

// DefaultInstallRunner usa os programas reais.
func DefaultInstallRunner() InstallRunner { return execInstall{} }

// Install roda o argv que InstallPlan devolveu, sem shell, e confere com `tmux -V` se o tmux passou
// a responder. Recusa qualquer argv que não seja uma instalação do tmux pela tabela (SEG-01).
func Install(argv []string, r InstallRunner) error {
	if !isTmuxInstallArgv(argv) {
		return fmt.Errorf("comando de instalação recusado: %q", InstallCommandLine(argv))
	}
	if err := r.Run(argv); err != nil {
		return fmt.Errorf("a instalação falhou: %w", err)
	}
	if err := r.Probe([]string{"tmux", "-V"}); err != nil {
		return fmt.Errorf("instalei, mas o tmux ainda não responde a tmux -V (abra um novo terminal para atualizar o PATH?): %w", err)
	}
	return nil
}

func isTmuxInstallArgv(argv []string) bool {
	if len(argv) > 0 && argv[0] == "sudo" {
		argv = argv[1:]
	}
	if len(argv) == 0 {
		return false
	}
	for _, in := range installTable {
		if argv[0] != in.manager || len(argv) != len(in.args)+1 {
			continue
		}
		same := true
		for i, a := range in.args {
			same = same && argv[i+1] == a
		}
		if same {
			return true
		}
	}
	return false
}

// TerminalPrompt é tudo de que a decisão do terminal precisa; injetável nos testes.
type TerminalPrompt struct {
	Env         LaunchEnv
	Geteuid     func() int
	Interactive bool // há terminal para perguntar; em --dry-run vale false (nada é instalado)
	Confirm     func(title string, def bool) (bool, error)
	Runner      InstallRunner
	Out         io.Writer
	// Remember grava terminal: native em forge.json (relendo o arquivo, nunca a configuração
	// mesclada com as flags da execução).
	Remember func() error
}

// TerminalDecision é onde as sessões abrem e se a escolha veio do auto (só então o tmux pode cair
// para as janelas do sistema).
type TerminalDecision struct {
	Terminal string
	Auto     bool
}

// RememberNative devolve um Remember que grava terminal: native em .forge/config/forge.json de dir.
func RememberNative(dir string) func() error {
	return func() error {
		cfg, found, err := config.LoadForgeConfig(dir)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("sem .forge/config/forge.json para gravar a escolha")
		}
		cfg.Terminal = config.TerminalNative
		return config.SaveForgeConfig(dir, cfg)
	}
}

// DecideTerminal resolve o terminal (ResolveTerminal) e, em `auto` sem tmux, oferece instalá-lo:
// terminal interativo mostra o comando e pergunta; sem terminal só imprime o comando; sem
// instalador ou com falha segue pelas janelas do sistema com uma linha de motivo. `tmux` explícito
// (flag ou forge.json) nunca instala: erra com o comando.
func DecideTerminal(flag string, cfg config.ForgeConfig, p TerminalPrompt) (TerminalDecision, error) {
	out := p.Out
	if out == nil {
		out = io.Discard
	}
	choice := cfg.Terminal
	if flag != "" {
		choice = flag
	}
	auto := choice == "" || choice == config.TerminalAuto
	term, err := ResolveTerminal(flag, cfg, p.Env)
	ienv := installEnvFrom(p.Env, p.Geteuid)
	if err != nil {
		if choice == config.TerminalTmux && p.Env.GOOS != "windows" && !has(p.Env, "tmux") {
			if argv, _, ok := InstallPlan(ienv); ok {
				return TerminalDecision{}, fmt.Errorf("tmux não está no PATH. Instale com: %s (ou use --terminal native)", InstallCommandLine(argv))
			}
		}
		return TerminalDecision{}, err
	}
	d := TerminalDecision{Terminal: term, Auto: auto}
	if !auto || term == config.TerminalTmux || p.Env.GOOS == "windows" {
		return d, nil // nada a oferecer: escolha explícita, tmux presente ou Windows nativo
	}

	// auto sem tmux: o motivo da contingência e, se houver, o comando de instalação.
	argv, desc, ok := InstallPlan(ienv)
	fmt.Fprintln(out, "tmux não encontrado. Com ele as 3 sessões abrem lado a lado numa tela só.")
	if !ok {
		fmt.Fprintf(out, "Não consigo instalar o tmux sozinho (%s); abrindo em janelas do sistema.\n", desc)
		return d, nil
	}
	fmt.Fprintf(out, "Comando para instalar (%s): %s\n", desc, InstallCommandLine(argv))
	if !p.Interactive || p.Confirm == nil {
		fmt.Fprintln(out, "Sem terminal para perguntar: abrindo em janelas do sistema. Instale o tmux e rode de novo para usar os painéis.")
		return d, nil
	}
	if argv[0] == "sudo" {
		fmt.Fprintln(out, "O sudo vai pedir a sua senha neste terminal.")
	}
	yes, err := p.Confirm("Instalar agora?", true)
	if err != nil || !yes { // Esc conta como recusa
		fmt.Fprintln(out, "Sem tmux desta vez: abrindo em janelas do sistema.")
		if err == nil {
			remember, rerr := p.Confirm("Lembrar a escolha e não perguntar de novo? (grava terminal: native em forge.json)", false)
			if rerr == nil && remember && p.Remember != nil {
				if serr := p.Remember(); serr != nil {
					fmt.Fprintf(out, "Não consegui gravar a escolha: %v\n", serr)
				} else {
					fmt.Fprintln(out, "Gravado: terminal: native. Para voltar ao tmux, use --terminal tmux ou mude terminal em forge.json.")
				}
			}
		}
		return d, nil
	}
	runner := p.Runner
	if runner == nil {
		runner = DefaultInstallRunner()
	}
	if ierr := Install(argv, runner); ierr != nil {
		fmt.Fprintf(out, "%v. Abrindo em janelas do sistema.\n", ierr)
		return d, nil
	}
	fmt.Fprintln(out, "tmux instalado.")
	d.Terminal = config.TerminalTmux
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
