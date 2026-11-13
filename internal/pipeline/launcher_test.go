package pipeline

import (
	"errors"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lenv(goos string, tools map[string]bool, env map[string]string, apps ...string) LaunchEnv {
	return LaunchEnv{
		GOOS: goos, Dir: "/work/meu projeto", Bin: "/usr/local/bin/forge-sdd",
		LookPath: func(n string) (string, error) {
			if tools[n] {
				return "/usr/bin/" + n, nil
			}
			return "", errors.New("não encontrado")
		},
		Getenv: func(k string) string { return env[k] },
		Exists: func(p string) bool {
			for _, a := range apps {
				if a == p {
					return true
				}
			}
			return false
		},
	}
}

func TestLauncher_SessoesPorCLI(t *testing.T) {
	for cli, want := range map[string]string{"claude": "claude -n forge-discovery --agent lider", "agy": "agy --agent lider -i"} {
		ss, err := Sessions(lenv("linux", nil, nil), cli)
		require.NoError(t, err, cli)
		require.Len(t, ss, 3)
		assert.Equal(t, []string{"forge-discovery", "forge-builder", "forge-reviewer"}, []string{ss[0].Title, ss[1].Title, ss[2].Title})
		assert.Contains(t, ss[0].Inner, want, cli)
		assert.Contains(t, ss[1].Inner, "worker --role builder --name forge-builder")
		assert.Contains(t, ss[2].Inner, "worker --role reviewer --name forge-reviewer")
		for _, s := range ss {
			assert.Contains(t, s.Inner, `printf '\033]0;`+s.Title+`\007'`, "título também por escape ANSI")
			assert.Contains(t, s.Inner, `cd '/work/meu projeto'`, "caminho com espaço entre aspas")
		}
		assert.Contains(t, ss[1].Inner, `'/usr/local/bin/forge-sdd' worker`)
	}
	_, err := Sessions(lenv("linux", nil, nil), "gpt")
	assert.Error(t, err)
}

func TestLauncher_MacOSTerminalEiTerm(t *testing.T) {
	plan, err := LaunchPlan(lenv("darwin", nil, nil), "claude")
	require.NoError(t, err)
	require.Len(t, plan, 3)
	assert.Equal(t, "osascript", plan[0].Argv[0])
	assert.Contains(t, plan[0].Argv[2], `tell application "Terminal" to do script`)

	plan, err = LaunchPlan(lenv("darwin", nil, nil, "/Applications/iTerm.app"), "claude")
	require.NoError(t, err)
	assert.Contains(t, plan[1].Argv[2], `tell application "iTerm"`)
	assert.Contains(t, plan[1].Argv[2], "forge-builder")
}

func TestLauncher_LinuxEscolheOTerminalDisponivel(t *testing.T) {
	cases := []struct {
		tools map[string]bool
		first string
		title string
	}{
		{map[string]bool{"gnome-terminal": true, "konsole": true}, "gnome-terminal", "--title=forge-discovery"},
		{map[string]bool{"konsole": true, "x-terminal-emulator": true}, "konsole", "tabtitle=forge-discovery"},
		{map[string]bool{"x-terminal-emulator": true}, "x-terminal-emulator", "forge-discovery"},
	}
	for _, c := range cases {
		plan, err := LaunchPlan(lenv("linux", c.tools, nil), "claude")
		require.NoError(t, err, c.first)
		assert.Equal(t, c.first, plan[0].Argv[0])
		assert.Contains(t, strings.Join(plan[0].Argv, " "), c.title)
		assert.Contains(t, plan[0].Argv[len(plan[0].Argv)-1], "exec bash", "a sessão não fecha ao terminar o comando")
	}
	_, err := LaunchPlan(lenv("linux", nil, nil), "claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nenhum terminal")
}

func TestLauncher_WindowsEWSL(t *testing.T) {
	win := lenv("windows", map[string]bool{"wt.exe": true}, nil)
	win.Dir, win.Bin = `C:\Users\ana\proj`, `C:\Program Files\forge\forge-sdd.exe`
	plan, err := LaunchPlan(win, "agy")
	require.NoError(t, err)
	assert.Equal(t, []string{"wt.exe", "new-tab", "--title", "forge-discovery", "cmd.exe", "/k"}, plan[0].Argv[:6])
	assert.Contains(t, plan[0].Argv[6], "agy --agent lider -i")
	assert.Contains(t, plan[1].Argv[6], `"C:\Program Files\forge\forge-sdd.exe" worker --role builder`)
	assert.Contains(t, plan[1].Argv[6], `cd /d "C:\Users\ana\proj"`)

	win = lenv("windows", nil, nil)
	win.Dir, win.Bin = `C:\p`, `C:\f\forge-sdd.exe`
	plan, err = LaunchPlan(win, "claude")
	require.NoError(t, err)
	assert.Equal(t, []string{"cmd.exe", "/c", "start", `"forge-discovery"`, "cmd.exe", "/k"}, plan[0].Argv[:6], "sem wt.exe cai em cmd /c start")

	wsl := lenv("linux", map[string]bool{"wt.exe": true}, map[string]string{"WSL_DISTRO_NAME": "Ubuntu-22.04"})
	plan, err = LaunchPlan(wsl, "claude")
	require.NoError(t, err)
	assert.Equal(t, "wt.exe", plan[0].Argv[0])
	assert.Contains(t, strings.Join(plan[0].Argv, " "), "wsl.exe -d Ubuntu-22.04 -- bash -lc")

	wsl2 := lenv("linux", nil, map[string]string{"WSL_DISTRO_NAME": "Ubuntu"})
	plan, err = LaunchPlan(wsl2, "claude")
	require.NoError(t, err)
	assert.Equal(t, "cmd.exe", plan[0].Argv[0])
}

func TestLauncher_Violacoes(t *testing.T) {
	bad := lenv("windows", map[string]bool{"wt.exe": true}, nil)
	for _, dir := range []string{`C:\p" & calc & "`, `C:\100%\p`, "C:\\p\nx", `C:\p|x`} {
		bad.Dir, bad.Bin = dir, `C:\f\forge-sdd.exe`
		_, err := LaunchPlan(bad, "claude")
		assert.Error(t, err, "%q", dir)
	}
	bad.Dir, bad.Bin = `C:\p`, `C:\f\forge-sdd.exe" & calc`
	_, err := LaunchPlan(bad, "claude")
	assert.Error(t, err)

	posix := lenv("linux", map[string]bool{"gnome-terminal": true}, nil)
	posix.Dir = "/p\nrm -rf ~"
	_, err = LaunchPlan(posix, "claude")
	assert.Error(t, err)
	posix.Dir = "/p/com 'aspas'"
	plan, err := LaunchPlan(posix, "claude")
	require.NoError(t, err)
	assert.Contains(t, plan[0].Argv[len(plan[0].Argv)-1], `'/p/com '\''aspas'\'''`, "aspas simples escapadas")

	posix.Dir, posix.Bin = "/p", "/bin/forge-sdd; rm -rf ~"
	_, err = LaunchPlan(posix, "claude")
	assert.Error(t, err, "binário com ; é recusado")

	wslBad := lenv("linux", map[string]bool{"wt.exe": true}, map[string]string{"WSL_DISTRO_NAME": "Ubu ntu; calc"})
	_, err = LaunchPlan(wslBad, "claude")
	assert.Error(t, err)
	_, err = LaunchPlan(lenv("plan9", nil, nil), "claude")
	assert.Error(t, err)
	assert.Error(t, Launch([]LaunchCommand{{Title: "x"}}))
}

func TestSessions_AgentePorSessao(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIClaude, config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}
	for _, goos := range []string{"darwin", "linux", "windows"} {
		env := lenv(goos, map[string]bool{"wt.exe": true, "gnome-terminal": true}, nil)
		if goos == "windows" {
			env.Dir, env.Bin = `C:\work\proj`, `C:\f\forge-sdd.exe`
		}
		ss, err := SessionsFor(env, cfg)
		require.NoError(t, err, goos)
		require.Len(t, ss, 3)
		assert.Contains(t, ss[0].Inner, "claude -n forge-discovery --agent lider", goos)
		assert.Contains(t, ss[1].Inner, "worker --role builder --name forge-builder --cli codex", goos)
		assert.Contains(t, ss[2].Inner, "worker --role reviewer --name forge-reviewer --cli gemini", goos)
		plan, err := LaunchPlanFor(env, cfg)
		require.NoError(t, err, goos)
		require.Len(t, plan, 3, goos)
	}
}

func TestSessions_DiscoveryInterativaPorCLI(t *testing.T) {
	want := map[string]string{
		config.CLIClaude:  "claude -n forge-discovery --agent lider",
		config.CLIGemini:  "gemini -i 'Voce e a sessao forge-discovery",
		config.CLICopilot: "copilot -i 'Voce e a sessao forge-discovery",
		config.CLICodex:   "codex 'Voce e a sessao forge-discovery",
		config.CLIAgy:     "agy --agent lider -i",
	}
	for cli, w := range want {
		cfg := config.DefaultForgeConfig()
		cfg.Sessions = map[string]string{config.SessionRoleDiscovery: cli}
		ss, err := SessionsFor(lenv("linux", nil, nil), cfg)
		require.NoError(t, err, cli)
		assert.Contains(t, ss[0].Inner, w, cli)
		assert.NotContains(t, ss[0].Inner, "{", "marcadores resolvidos: %s", cli)
	}
	// No Windows o prompt vai entre aspas duplas e o comando não ganha metacaracteres.
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIGemini}
	env := lenv("windows", map[string]bool{"wt.exe": true}, nil)
	env.Dir, env.Bin = `C:\work\proj`, `C:\f\forge-sdd.exe`
	ss, err := SessionsFor(env, cfg)
	require.NoError(t, err)
	assert.Contains(t, ss[0].Inner, `gemini -i "Voce e a sessao forge-discovery`)
}

func TestSessions_AgenteInvalidoERecusado(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleBuilder: "gpt"}
	_, err := SessionsFor(lenv("linux", nil, nil), cfg)
	assert.Error(t, err)
	cfg = config.DefaultForgeConfig()
	delete(cfg.CLIProfiles, config.CLICodex)
	cfg.Sessions = map[string]string{config.SessionRoleReviewer: config.CLICodex}
	_, err = SessionsFor(lenv("linux", nil, nil), cfg)
	assert.Error(t, err, "sem perfil não abre")
}
