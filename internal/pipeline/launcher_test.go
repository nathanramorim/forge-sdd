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
	for cli, want := range map[string]string{"claude": "claude -n forge-discovery --agent lider", "agy": "agy -i 'Voce e a sessao forge-discovery"} {
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

func TestLauncherWindows(t *testing.T) {
	win := lenv("windows", map[string]bool{"wt.exe": true}, nil)
	win.Dir, win.Bin = `C:\Users\ana\proj`, `C:\Program Files\forge\forge-sdd.exe`
	plan, err := LaunchPlan(win, "agy")
	require.NoError(t, err)
	require.Len(t, plan, 1, "com wt.exe abre uma única chamada para os 3 painéis")
	argv := plan[0].Argv
	require.GreaterOrEqual(t, len(argv), 25)
	assert.Equal(t, []string{"wt.exe", "-w", "new", "new-tab", "--title", "forge-discovery", "cmd.exe", "/k"}, argv[:8])
	assert.Contains(t, argv[8], "agy -i")
	assert.Equal(t, ";", argv[9])
	assert.Equal(t, []string{"split-pane", "-V", "--title", "forge-builder", "cmd.exe", "/k"}, argv[10:16])
	assert.Contains(t, argv[16], `"C:\Program Files\forge\forge-sdd.exe" worker --role builder`)
	assert.Contains(t, argv[16], `cd /d "C:\Users\ana\proj"`)
	assert.Equal(t, ";", argv[17])
	assert.Equal(t, []string{"split-pane", "-H", "--title", "forge-reviewer", "cmd.exe", "/k"}, argv[18:24])
	assert.Contains(t, argv[24], `"C:\Program Files\forge\forge-sdd.exe" worker --role reviewer`)

	// Sem wt.exe: 3 comandos com cmd.exe /c start
	winNoWT := lenv("windows", nil, nil)
	winNoWT.Dir, winNoWT.Bin = `C:\p`, `C:\f\forge-sdd.exe`
	planNoWT, err := LaunchPlan(winNoWT, "claude")
	require.NoError(t, err)
	require.Len(t, planNoWT, 3)
	assert.Equal(t, []string{"cmd.exe", "/c", "start", `"forge-discovery"`, "cmd.exe", "/k"}, planNoWT[0].Argv[:6], "sem wt.exe cai em cmd /c start")
	assert.Equal(t, []string{"cmd.exe", "/c", "start", `"forge-builder"`, "cmd.exe", "/k"}, planNoWT[1].Argv[:6])
	assert.Equal(t, []string{"cmd.exe", "/c", "start", `"forge-reviewer"`, "cmd.exe", "/k"}, planNoWT[2].Argv[:6])

	// WSL
	wsl := lenv("linux", map[string]bool{"wt.exe": true}, map[string]string{"WSL_DISTRO_NAME": "Ubuntu-22.04"})
	planWSL, err := LaunchPlan(wsl, "claude")
	require.NoError(t, err)
	assert.Equal(t, "wt.exe", planWSL[0].Argv[0])
	assert.Contains(t, strings.Join(planWSL[0].Argv, " "), "wsl.exe -d Ubuntu-22.04 -- bash -lc")

	wsl2 := lenv("linux", nil, map[string]string{"WSL_DISTRO_NAME": "Ubuntu"})
	planWSL2, err := LaunchPlan(wsl2, "claude")
	require.NoError(t, err)
	assert.Equal(t, "cmd.exe", planWSL2[0].Argv[0])
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
		if goos == "windows" {
			require.Len(t, plan, 1, "windows com wt.exe abre uma única janela dividida em 3 painéis")
		} else {
			require.Len(t, plan, 3, goos)
		}
	}
}

func TestSessions_DiscoveryInterativaPorCLI(t *testing.T) {
	want := map[string]string{
		config.CLIClaude:  "claude -n forge-discovery --agent lider",
		config.CLIGemini:  "agy -i 'Voce e a sessao forge-discovery",
		config.CLICopilot: "copilot -i 'Voce e a sessao forge-discovery",
		config.CLICodex:   "codex 'Voce e a sessao forge-discovery",
		config.CLIAgy:     "agy -i 'Voce e a sessao forge-discovery",
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
	assert.Contains(t, ss[0].Inner, `agy -i "Voce e a sessao forge-discovery`)
}

func TestDiscoveryPrompt_ProtocoloBrief(t *testing.T) {
	assert.Contains(t, DiscoveryPrompt, "forge-sdd brief")
	assert.Contains(t, DiscoveryPrompt, "proponha /constitution")
	assert.Contains(t, DiscoveryPrompt, "cite as decisoes do projeto na spec")
	assert.Contains(t, DiscoveryPrompt, "/implementar")
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

func TestSessions_PontoEVirgulaNoWindowsERecusado(t *testing.T) {
	// ; é separador de subcomandos do wt.exe: um argumento de perfil com ; não pode chegar à linha de comando.
	cfg := config.DefaultForgeConfig()
	cfg.CLIProfiles[config.CLIGemini] = config.CLIProfile{Command: "gemini", Args: []string{"-p", "{prompt}"}, InteractiveArgs: []string{"-i", "{prompt}", "a;b"}}
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIGemini}
	env := lenv("windows", map[string]bool{"wt.exe": true}, nil)
	env.Dir, env.Bin = `C:\work\proj`, `C:\f\forge-sdd.exe`
	_, err := SessionsFor(env, cfg)
	assert.Error(t, err)
	// No POSIX o mesmo argumento vai entre aspas simples e é aceito.
	_, err = SessionsFor(lenv("linux", nil, nil), cfg)
	assert.NoError(t, err)
}

func TestSessions_DiscoveryGeminiAbreAgyComPromptI(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIGemini}
	ss, err := SessionsFor(lenv("linux", nil, nil), cfg)
	require.NoError(t, err)
	assert.Contains(t, ss[0].Inner, "agy -i 'Voce e a sessao forge-discovery")
	assert.NotContains(t, ss[0].Inner, "gemini")
}

func TestSessions_LiveModeBuilderReviewerGeminiAbreAgySemFlagI(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.SessionMode = config.SessionModeLive
	cfg.Sessions = map[string]string{
		config.SessionRoleBuilder:  config.CLIGemini,
		config.SessionRoleReviewer: config.CLIGemini,
	}
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), cfg, false)
	require.NoError(t, err)

	// Painel 1 (Builder) e Painel 2 (Reviewer) devem rodar agy sem flag -i
	bldCmd := plan.Create[4][7]
	revCmd := plan.Create[5][7]

	assert.Contains(t, bldCmd, "agy --dangerously-skip-permissions")
	assert.NotContains(t, bldCmd, "-i")
	assert.NotContains(t, bldCmd, "gemini")

	assert.Contains(t, revCmd, "agy --dangerously-skip-permissions")
	assert.NotContains(t, revCmd, "-i")
	assert.NotContains(t, revCmd, "gemini")
}
