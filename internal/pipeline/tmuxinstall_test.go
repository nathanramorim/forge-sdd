package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInstallRunner struct {
	runs   [][]string
	probes [][]string
	runErr error
	prbErr error
}

func (f *fakeInstallRunner) Run(argv []string) error {
	f.runs = append(f.runs, argv)
	return f.runErr
}

func (f *fakeInstallRunner) Probe(argv []string) error {
	f.probes = append(f.probes, argv)
	return f.prbErr
}

func TestTmuxInstall_TabelaPorSOePATH(t *testing.T) {
	mockEnv := func(goos string, root bool, tools ...string) InstallEnv {
		has := map[string]bool{}
		for _, tool := range tools {
			has[tool] = true
		}
		euid := 1000
		if root {
			euid = 0
		}
		return InstallEnv{
			GOOS: goos,
			LookPath: func(name string) (string, error) {
				if has[name] {
					return "/bin/" + name, nil
				}
				return "", errors.New("not found")
			},
			Geteuid: func() int { return euid },
		}
	}

	// 1. macOS com Homebrew (root ou não root nunca usa sudo no brew)
	argv, desc, ok := InstallPlan(mockEnv("darwin", false, "brew"))
	require.True(t, ok)
	assert.Equal(t, []string{"brew", "install", "tmux"}, argv)
	assert.Equal(t, "macOS", desc)

	argv, _, ok = InstallPlan(mockEnv("darwin", true, "brew"))
	require.True(t, ok)
	assert.Equal(t, []string{"brew", "install", "tmux"}, argv)

	// 2. Linux com apt-get + sudo
	argv, desc, ok = InstallPlan(mockEnv("linux", false, "apt-get", "sudo"))
	require.True(t, ok)
	assert.Equal(t, []string{"sudo", "apt-get", "install", "-y", "tmux"}, argv)
	assert.Equal(t, "Debian/Ubuntu", desc)

	// 3. Linux com apt-get sendo root
	argv, _, ok = InstallPlan(mockEnv("linux", true, "apt-get"))
	require.True(t, ok)
	assert.Equal(t, []string{"apt-get", "install", "-y", "tmux"}, argv)

	// 4. Fedora (dnf)
	argv, _, ok = InstallPlan(mockEnv("linux", false, "dnf", "sudo"))
	require.True(t, ok)
	assert.Equal(t, []string{"sudo", "dnf", "install", "-y", "tmux"}, argv)

	// 5. Arch (pacman)
	argv, _, ok = InstallPlan(mockEnv("linux", false, "pacman", "sudo"))
	require.True(t, ok)
	assert.Equal(t, []string{"sudo", "pacman", "-S", "--noconfirm", "tmux"}, argv)

	// 6. Alpine (apk)
	argv, _, ok = InstallPlan(mockEnv("linux", false, "apk", "sudo"))
	require.True(t, ok)
	assert.Equal(t, []string{"sudo", "apk", "add", "tmux"}, argv)

	// 7. openSUSE (zypper)
	argv, _, ok = InstallPlan(mockEnv("linux", false, "zypper", "sudo"))
	require.True(t, ok)
	assert.Equal(t, []string{"sudo", "zypper", "--non-interactive", "install", "tmux"}, argv)

	// 8. Linux sem sudo e não-root
	_, desc, ok = InstallPlan(mockEnv("linux", false, "apt-get"))
	assert.False(t, ok)
	assert.Contains(t, desc, "sudo não está no PATH")

	// 9. Windows nativo
	_, desc, ok = InstallPlan(mockEnv("windows", false, "brew", "apt-get"))
	assert.False(t, ok)
	assert.Contains(t, desc, "Windows nativo")

	// 10. Sem gerenciador conhecido
	_, desc, ok = InstallPlan(mockEnv("linux", false))
	assert.False(t, ok)
	assert.Contains(t, desc, "não achei um gerenciador")
}

func TestTmuxInstall_ExecucaoERecusaDeComandosInseguros(t *testing.T) {
	runner := &fakeInstallRunner{}

	// Comando válido
	err := Install([]string{"brew", "install", "tmux"}, runner)
	require.NoError(t, err)
	require.Len(t, runner.runs, 1)
	require.Len(t, runner.probes, 1)
	assert.Equal(t, []string{"tmux", "-V"}, runner.probes[0])

	// Comando inseguro ou arbitrário é recusado sem executar
	badCommands := [][]string{
		{"rm", "-rf", "/"},
		{"sh", "-c", "echo oi"},
		{"brew", "install", "wget"},
		{"sudo", "apt-get", "remove", "tmux"},
	}
	for _, bad := range badCommands {
		err := Install(bad, runner)
		assert.Error(t, err, "deve recusar comando fora da tabela: %v", bad)
	}
}

func TestTerminalDecisao(t *testing.T) {
	lenvMock := func(hasTmux bool, tools ...string) LaunchEnv {
		has := map[string]bool{"tmux": hasTmux}
		for _, tool := range tools {
			has[tool] = true
		}
		return LaunchEnv{
			GOOS: "darwin",
			LookPath: func(name string) (string, error) {
				if has[name] {
					return "/bin/" + name, nil
				}
				return "", errors.New("not found")
			},
		}
	}

	cfg := config.DefaultForgeConfig()
	cfg.Terminal = config.TerminalAuto

	// 1. auto sem tmux, interativo e resposta SIM instala e usa tmux
	var out bytes.Buffer
	runner := &fakeInstallRunner{}
	d, err := DecideTerminal("", cfg, TerminalPrompt{
		Env:         lenvMock(false, "brew"),
		Interactive: true,
		Out:         &out,
		Confirm:     func(string, bool) (bool, error) { return true, nil },
		Runner:      runner,
	})
	require.NoError(t, err)
	assert.Equal(t, config.TerminalTmux, d.Terminal)
	require.Len(t, runner.runs, 1)
	assert.Contains(t, out.String(), "tmux instalado")

	// 2. auto sem tmux, interativo e resposta NÃO lembra escolha gravando native
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
	rawCfg, _ := json.Marshal(cfg)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), rawCfg, 0o644))

	out.Reset()
	confirms := []bool{false, true} // 1: Instalar? Não. 2: Lembrar? Sim.
	cIdx := 0
	d, err = DecideTerminal("", cfg, TerminalPrompt{
		Env:         lenvMock(false, "brew"),
		Interactive: true,
		Out:         &out,
		Confirm: func(string, bool) (bool, error) {
			ans := confirms[cIdx]
			cIdx++
			return ans, nil
		},
		Remember: RememberNative(dir),
	})
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, d.Terminal)
	assert.Contains(t, out.String(), "Gravado: terminal: native")

	saved, _, _ := config.LoadForgeConfig(dir)
	assert.Equal(t, config.TerminalNative, saved.Terminal)

	// 3. auto sem tmux e não interativo: só imprime aviso e segue native
	out.Reset()
	d, err = DecideTerminal("", cfg, TerminalPrompt{
		Env:         lenvMock(false, "brew"),
		Interactive: false,
		Out:         &out,
	})
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, d.Terminal)
	assert.Contains(t, out.String(), "Sem terminal para perguntar: abrindo em janelas do sistema")

	// 4. --terminal tmux explícito: nunca instala e falha com comando sugerido
	_, err = DecideTerminal(config.TerminalTmux, cfg, TerminalPrompt{
		Env: lenvMock(false, "brew"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tmux não está no PATH. Instale com: brew install tmux")
}

type fakeTmuxRunner struct {
	hasSession bool
	errOn      map[string]error
	runs       [][]string
}

func (f *fakeTmuxRunner) Run(argv []string) error {
	f.runs = append(f.runs, argv)
	if len(argv) > 1 {
		if argv[1] == "has-session" {
			if f.hasSession {
				return nil
			}
			return errors.New("sem sessao")
		}
		if f.errOn != nil {
			if err, ok := f.errOn[argv[1]]; ok {
				return err
			}
		}
	}
	return nil
}

func (f *fakeTmuxRunner) Attach(argv []string) error {
	f.runs = append(f.runs, argv)
	if len(argv) > 1 && f.errOn != nil {
		if err, ok := f.errOn[argv[1]]; ok {
			return err
		}
	}
	return nil
}

func TestTmuxFallback(t *testing.T) {
	plan := &TmuxPlan{
		Session: "test-sess",
		Create:  [][]string{{"tmux", "new-session", "-d", "-s", "test-sess"}},
		Attach:  []string{"tmux", "attach-session", "-t", "=test-sess"},
	}

	// 1. Em auto=true, erro no new-session desfaz e aciona fallback
	var warn bytes.Buffer
	fallbackCalled := false
	runnerErr := &fakeTmuxRunner{errOn: map[string]error{"new-session": errors.New("server exited unexpectedly")}}

	created, fellBack, err := LaunchTerminalTmux(plan, runnerErr, true, &warn, func() error {
		fallbackCalled = true
		return nil
	})
	require.NoError(t, err)
	assert.False(t, created)
	assert.True(t, fellBack)
	assert.True(t, fallbackCalled)
	assert.Contains(t, warn.String(), "Abrindo em janelas do sistema")

	// 2. Em auto=false (explícito), erro não chama fallback e retorna erro
	warn.Reset()
	fallbackCalled = false
	_, fellBack, err = LaunchTerminalTmux(plan, runnerErr, false, &warn, func() error {
		fallbackCalled = true
		return nil
	})
	require.Error(t, err)
	assert.False(t, fellBack)
	assert.False(t, fallbackCalled)

	// 3. TmuxAttachError (sessão criada, attach falhou) nunca chama fallback para não duplicar agentes
	runnerAttachErr := &fakeTmuxRunner{errOn: map[string]error{"attach-session": errors.New("no tty")}}
	fallbackCalled = false
	_, fellBack, err = LaunchTerminalTmux(plan, runnerAttachErr, true, nil, func() error {
		fallbackCalled = true
		return nil
	})
	require.Error(t, err)
	var target *TmuxAttachError
	assert.True(t, errors.As(err, &target))
	assert.False(t, fellBack)
	assert.False(t, fallbackCalled)
}
