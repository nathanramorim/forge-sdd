package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// Nomes das três sessões do pipeline.
const (
	SessionDiscovery = "forge-discovery"
	SessionBuilder   = "forge-builder"
	SessionReviewer  = "forge-reviewer"
)

// Session é uma sessão de terminal a abrir: título e o comando interno que ela executa.
type Session struct {
	Title   string
	Inner   string // linha de comando da sessão (POSIX ou cmd.exe, conforme o sistema)
	Windows bool
}

// LaunchCommand é um comando externo que abre uma sessão (argv explícito, sem shell).
type LaunchCommand struct {
	Title string
	Argv  []string
}

// LaunchEnv descreve o ambiente para escolher o terminal; injetável nos testes.
type LaunchEnv struct {
	GOOS     string
	Dir      string // raiz do projeto
	Bin      string // caminho do forge-sdd
	LookPath func(string) (string, error)
	Getenv   func(string) string
	Exists   func(string) bool
}

// DefaultLaunchEnv usa o ambiente real.
func DefaultLaunchEnv(goos, dir, bin string) LaunchEnv {
	return LaunchEnv{
		GOOS: goos, Dir: dir, Bin: bin,
		LookPath: exec.LookPath, Getenv: os.Getenv,
		Exists: func(p string) bool { _, err := os.Stat(p); return err == nil },
	}
}

var (
	safeBinRe   = regexp.MustCompile(`^[A-Za-z0-9._/:\\ ()+-]+$`)
	winUnsafe   = regexp.MustCompile(`["%^&|<>;\r\n\x00]`) // ; é separador de subcomandos do wt.exe
	posixUnsafe = regexp.MustCompile(`[\r\n\x00]`)
)

// shQuote coloca s entre aspas simples para um shell POSIX.
func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// DiscoveryPrompt é o prompt inicial da sessão Discovery nas CLIs que o aceitam. Texto ASCII, sem
// aspas, % nem metacaracteres: precisa sobreviver ao cmd.exe e ao AppleScript sem escapes frágeis.
const DiscoveryPrompt = "Voce e a sessao forge-discovery do Forge SDD. Rode forge-sdd brief no inicio: se estiver em rascunho proponha /constitution antes de qualquer spec. Siga .agents/skills/sdd-concepcao/SKILL.md, cite as decisoes do projeto na spec e use /implementar quando a feature estiver pronta."

var safeArgRe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// quoteArg protege um argumento para a linha de comando de uma sessão: sem aspas quando é seguro
// (comandos legíveis), com aspas simples no POSIX e duplas no cmd.exe quando precisa.
func quoteArg(windows bool, a string) (string, error) {
	if a == "" {
		if windows {
			return `""`, nil
		}
		return "''", nil
	}
	if safeArgRe.MatchString(a) && !(windows && strings.Contains(a, "%")) {
		return a, nil
	}
	if windows {
		if winUnsafe.MatchString(a) {
			return "", fmt.Errorf("argumento com caracteres não suportados no Windows: %q", a)
		}
		return `"` + a + `"`, nil
	}
	if posixUnsafe.MatchString(a) {
		return "", fmt.Errorf("argumento com caracteres de controle: %q", a)
	}
	return shQuote(a), nil
}

func interactiveLine(windows bool, cfg config.ForgeConfig, cli string) (string, error) {
	profile, ok := cfg.Profile(cli)
	if !ok || !config.KnownCLI(cli) {
		return "", fmt.Errorf("cli %q inválida (use %s)", cli, strings.Join(config.OfferedCLIs(), ", "))
	}
	args, err := expandArgs(profile.InteractiveArgs, RunSpec{Agent: "lider", Name: SessionDiscovery, Prompt: DiscoveryPrompt})
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(args)+1)
	for _, a := range append([]string{profile.Command}, args...) {
		q, err := quoteArg(windows, a)
		if err != nil {
			return "", err
		}
		parts = append(parts, q)
	}
	return strings.Join(parts, " "), nil
}

// Sessions mantém a assinatura da 4.1: a mesma CLI nas três sessões.
func Sessions(env LaunchEnv, cli string) ([]Session, error) {
	cfg := config.DefaultForgeConfig()
	cfg.CLI = cli
	cfg.Sessions = nil
	return SessionsFor(env, cfg)
}

// SessionsFor monta as três sessões com o agente de cada uma (cfg.CLIFor). A Discovery é interativa
// (interactive_args do perfil, agente "lider"); Builder e Reviewer rodam `forge-sdd worker --cli <agente>`. O título é
// fixado também por escape ANSI (`\033]0;<nome>\007`), que cobre CLIs sem flag de título.
func SessionsFor(env LaunchEnv, cfg config.ForgeConfig) ([]Session, error) {
	windows := env.GOOS == "windows"
	for name, v := range map[string]string{"diretório do projeto": env.Dir, "binário do forge-sdd": env.Bin} {
		if v == "" {
			return nil, fmt.Errorf("%s vazio", name)
		}
		if windows && winUnsafe.MatchString(v) {
			return nil, fmt.Errorf("%s contém caracteres não suportados no Windows: %q", name, v)
		}
		if !windows && posixUnsafe.MatchString(v) {
			return nil, fmt.Errorf("%s contém caracteres de controle", name)
		}
	}
	if !safeBinRe.MatchString(env.Bin) {
		return nil, fmt.Errorf("caminho do binário %q contém caracteres não suportados", env.Bin)
	}

	discoveryCLI, err := interactiveLine(windows, cfg, cfg.CLIFor(config.SessionRoleDiscovery))
	if err != nil {
		return nil, err
	}
	workerArgs := func(role, name string) (string, error) {
		cli := cfg.CLIFor(role)
		if !config.KnownCLI(cli) {
			return "", fmt.Errorf("cli %q inválida para %s (use %s)", cli, role, strings.Join(config.OfferedCLIs(), ", "))
		}
		if _, ok := cfg.Profile(cli); !ok {
			return "", fmt.Errorf("cli_profiles sem perfil para %q", cli)
		}
		// --cli sempre explícito: o worker usa exatamente o agente que esta abertura escolheu, mesmo
		// quando a escolha veio de uma flag e não foi gravada em forge.json.
		return "forge-sdd worker --role " + role + " --name " + name + " --cli " + cli, nil
	}
	build := func(title, args string) Session {
		if windows {
			resolved := strings.Replace(args, "forge-sdd ", `"`+env.Bin+`" `, 1)
			return Session{Title: title, Inner: fmt.Sprintf(`title %s && cd /d "%s" && %s`, title, env.Dir, resolved), Windows: true}
		}
		resolved := strings.Replace(args, "forge-sdd ", shQuote(env.Bin)+" ", 1)
		return Session{Title: title, Inner: fmt.Sprintf(`cd %s && printf '\033]0;%s\007' && %s`, shQuote(env.Dir), title, resolved)}
	}
	builder, err := workerArgs(config.SessionRoleBuilder, SessionBuilder)
	if err != nil {
		return nil, err
	}
	reviewer, err := workerArgs(config.SessionRoleReviewer, SessionReviewer)
	if err != nil {
		return nil, err
	}
	return []Session{
		build(SessionDiscovery, discoveryCLI),
		build(SessionBuilder, builder),
		build(SessionReviewer, reviewer),
	}, nil
}

// LaunchPlan devolve os comandos que abrem as três sessões no terminal do sistema.
func LaunchPlan(env LaunchEnv, cli string) ([]LaunchCommand, error) {
	cfg := config.DefaultForgeConfig()
	cfg.CLI = cli
	return LaunchPlanFor(env, cfg)
}

// LaunchPlanFor devolve os comandos que abrem as três sessões com o agente de cada uma.
func LaunchPlanFor(env LaunchEnv, cfg config.ForgeConfig) ([]LaunchCommand, error) {
	sessions, err := SessionsFor(env, cfg)
	if err != nil {
		return nil, err
	}
	var out []LaunchCommand
	for _, s := range sessions {
		argv, err := terminalArgv(env, s)
		if err != nil {
			return nil, err
		}
		out = append(out, LaunchCommand{Title: s.Title, Argv: argv})
	}
	return out, nil
}

func has(env LaunchEnv, name string) bool {
	_, err := env.LookPath(name)
	return err == nil
}

func terminalArgv(env LaunchEnv, s Session) ([]string, error) {
	isWSL := env.GOOS == "linux" && env.Getenv("WSL_DISTRO_NAME") != ""
	switch {
	case env.GOOS == "darwin":
		if env.Exists("/Applications/iTerm.app") {
			script := fmt.Sprintf(`tell application "iTerm" to tell (create window with default profile) to tell current session to write text "%s"`, appleEscape(s.Inner))
			return []string{"osascript", "-e", script}, nil
		}
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s"`, appleEscape(s.Inner))
		return []string{"osascript", "-e", script}, nil

	case env.GOOS == "windows":
		if has(env, "wt.exe") || has(env, "wt") {
			return []string{"wt.exe", "new-tab", "--title", s.Title, "cmd.exe", "/k", s.Inner}, nil
		}
		return []string{"cmd.exe", "/c", "start", `"` + s.Title + `"`, "cmd.exe", "/k", s.Inner}, nil

	case isWSL:
		distro := env.Getenv("WSL_DISTRO_NAME")
		if !regexp.MustCompile(`^[A-Za-z0-9._-]+$`).MatchString(distro) {
			return nil, fmt.Errorf("WSL_DISTRO_NAME %q inválido", distro)
		}
		inner := s.Inner + "; exec bash"
		if has(env, "wt.exe") {
			return []string{"wt.exe", "new-tab", "--title", s.Title, "wsl.exe", "-d", distro, "--", "bash", "-lc", inner}, nil
		}
		return []string{"cmd.exe", "/c", "start", `"` + s.Title + `"`, "wsl.exe", "-d", distro, "--", "bash", "-lc", inner}, nil

	case env.GOOS == "linux":
		inner := s.Inner + "; exec bash"
		switch {
		case has(env, "gnome-terminal"):
			return []string{"gnome-terminal", "--title=" + s.Title, "--", "bash", "-lc", inner}, nil
		case has(env, "konsole"):
			return []string{"konsole", "-p", "tabtitle=" + s.Title, "-e", "bash", "-lc", inner}, nil
		case has(env, "x-terminal-emulator"):
			return []string{"x-terminal-emulator", "-T", s.Title, "-e", "bash", "-lc", inner}, nil
		}
		return nil, errors.New("nenhum terminal encontrado (gnome-terminal, konsole ou x-terminal-emulator)")
	}
	return nil, fmt.Errorf("sistema %q não suportado pelo launcher", env.GOOS)
}

func appleEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

// Launch executa os comandos de abertura sem esperar o término (cada terminal fica independente).
func Launch(cmds []LaunchCommand) error {
	for _, c := range cmds {
		if len(c.Argv) == 0 {
			return fmt.Errorf("comando vazio para %s", c.Title)
		}
		if err := exec.Command(c.Argv[0], c.Argv[1:]...).Start(); err != nil {
			return fmt.Errorf("abrir a sessão %s: %w", c.Title, err)
		}
	}
	return nil
}

// ResolveBin devolve o caminho absoluto deste binário.
func ResolveBin() string {
	if p, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
	}
	return "forge-sdd"
}
