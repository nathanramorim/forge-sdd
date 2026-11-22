package pipeline

import (
	"bytes"
	"errors"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestTerminalDecisao(t *testing.T) {
	look := func(tools ...string) func(string) (string, error) {
		return func(n string) (string, error) {
			for _, x := range tools {
				if x == n {
					return "/bin/" + n, nil
				}
			}
			return "", errors.New("not found")
		}
	}
	decide := func(goos, flag, cfgTerm string, tools ...string) (TerminalDecision, string, error) {
		var out bytes.Buffer
		d, err := DecideTerminal(flag, config.ForgeConfig{Terminal: cfgTerm}, TerminalPrompt{
			Env: LaunchEnv{GOOS: goos, LookPath: look(tools...)}, Out: &out,
		})
		return d, out.String(), err
	}

	// padrão: janelas do sistema, mesmo com tmux instalado, sem aviso
	for _, c := range []string{"", config.TerminalAuto, config.TerminalNative} {
		d, out, err := decide("linux", "", c, "tmux")
		require.NoError(t, err)
		assert.Equal(t, config.TerminalNative, d.Terminal)
		assert.Empty(t, out)
	}
	// opção explícita (flag ou config) com tmux presente
	d, _, err := decide("linux", "tmux", "", "tmux")
	require.NoError(t, err)
	assert.Equal(t, config.TerminalTmux, d.Terminal)
	assert.False(t, d.Auto)
	d, _, err = decide("darwin", "", config.TerminalTmux, "tmux")
	require.NoError(t, err)
	assert.Equal(t, config.TerminalTmux, d.Terminal)
	// flag vence a config
	d, _, err = decide("linux", "native", config.TerminalTmux, "tmux")
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, d.Terminal)
	// tmux ausente: avisa com o comando de instalação e cai para native, sem erro
	d, out, err := decide("darwin", "tmux", "", "brew")
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, d.Terminal)
	assert.Contains(t, out, "tmux não encontrado")
	assert.Contains(t, out, "brew install tmux")
	// Windows nativo: avisa e cai para native
	d, out, err = decide("windows", "tmux", "", "tmux")
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, d.Terminal)
	assert.Contains(t, out, "tmux não encontrado")
	// valor inválido é recusado
	_, _, err = decide("linux", "janelas", "", "tmux")
	assert.Error(t, err)
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
