package pipeline

import (
	"errors"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmux_PlanoExato(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = "/work/meu-app"
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleBuilder: config.CLICodex}
	plan, err := BuildTmuxPlan(env, cfg, false)
	require.NoError(t, err)
	assert.Equal(t, "forge-meu-app", plan.Session)

	c := plan.Create
	require.GreaterOrEqual(t, len(c), 10)
	assert.Equal(t, []string{"tmux", "new-session", "-d", "-s", "forge-meu-app", "-n", "forge", "-c", "/work/meu-app"}, c[0][:9])
	assert.Contains(t, c[0][9], "claude -n forge-discovery --agent lider", "painel da Discovery")
	assert.Equal(t, []string{"tmux", "split-window", "-h", "-t", "forge-meu-app:forge", "-c", "/work/meu-app"}, c[1][:7])
	assert.Contains(t, c[1][7], "worker --role builder --name forge-builder --cli codex", "Builder à direita no alto")
	assert.Equal(t, []string{"tmux", "split-window", "-v", "-t", "forge-meu-app:forge.1", "-c", "/work/meu-app"}, c[2][:7])
	assert.Contains(t, c[2][7], "worker --role reviewer --name forge-reviewer --cli claude", "Reviewer embaixo")
	assert.Equal(t, []string{"tmux", "set-window-option", "-t", "forge-meu-app:forge", "pane-border-status", "top"}, c[3])
	var titles []string
	for _, step := range c {
		if step[1] == "select-pane" && len(step) > 4 && step[4] == "-T" {
			titles = append(titles, step[5])
		}
	}
	assert.Equal(t, []string{"forge-discovery", "forge-builder", "forge-reviewer"}, titles)
	assert.Equal(t, []string{"tmux", "select-pane", "-t", "forge-meu-app:forge.0"}, c[len(c)-1], "o foco termina na Discovery")
	assert.Equal(t, []string{"tmux", "attach-session", "-t", "forge-meu-app"}, plan.Attach)

	inside, err := BuildTmuxPlan(env, cfg, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"tmux", "switch-client", "-t", "forge-meu-app"}, inside.Attach, "dentro do tmux troca o cliente")
}

func TestTmux_NomeDaSessaoSaneado(t *testing.T) {
	for in, want := range map[string]string{
		"/p/meu app":                    "forge-meu-app",
		"/p/a:b.c":                      "forge-a-b-c",
		"/p/x;rm -rf ~":                 "forge-x-rm--rf",
		"/p/":                           "forge-p",
		"/p/" + strings.Repeat("a", 80): "forge-" + strings.Repeat("a", 40),
	} {
		assert.Equal(t, want, TmuxSessionName(in), in)
	}
	assert.Equal(t, "forge-projeto", TmuxSessionName("/"))
}

func TestTmux_InjecaoEmNomeDeProjetoNaoViraComando(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = "/work/x; touch /tmp/pwn #"
	plan, err := BuildTmuxPlan(env, config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	assert.NotContains(t, plan.Session, ";")
	assert.NotContains(t, plan.Session, " ")
	// O diretório vai como argumento próprio de -c e, no comando do painel, entre aspas simples.
	assert.Equal(t, env.Dir, plan.Create[0][8])
	assert.Contains(t, plan.Create[0][9], `cd '/work/x; touch /tmp/pwn #'`)
}

func TestTmux_NaoExisteNoWindowsNativo(t *testing.T) {
	env := lenv("windows", map[string]bool{"tmux": true}, nil)
	env.Dir, env.Bin = `C:\work\proj`, `C:\f\forge-sdd.exe`
	_, err := BuildTmuxPlan(env, config.DefaultForgeConfig(), false)
	assert.Error(t, err)
	term, err := ResolveTerminal("", config.DefaultForgeConfig(), env)
	require.NoError(t, err)
	assert.Equal(t, config.TerminalNative, term, "no Windows o auto fica nas janelas do sistema")
	_, err = ResolveTerminal("tmux", config.DefaultForgeConfig(), env)
	assert.Error(t, err)
}

func TestTmux_ResolverTerminalPorPrecedencia(t *testing.T) {
	with := lenv("darwin", map[string]bool{"tmux": true}, nil)
	without := lenv("darwin", nil, nil)
	cfg := config.DefaultForgeConfig()

	term, _ := ResolveTerminal("", cfg, with)
	assert.Equal(t, config.TerminalTmux, term, "auto com tmux")
	term, _ = ResolveTerminal("", cfg, without)
	assert.Equal(t, config.TerminalNative, term, "auto sem tmux mantém o terminal do sistema")

	cfg.Terminal = config.TerminalNative
	term, _ = ResolveTerminal("", cfg, with)
	assert.Equal(t, config.TerminalNative, term, "forge.json vence o auto")
	term, _ = ResolveTerminal("tmux", cfg, with)
	assert.Equal(t, config.TerminalTmux, term, "a flag vence o forge.json")

	_, err := ResolveTerminal("tmux", config.DefaultForgeConfig(), without)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tmux não está no PATH")
	_, err = ResolveTerminal("janelas", cfg, with)
	assert.Error(t, err)
}

type fakeTmux struct {
	existing bool
	runs     [][]string
	attached []string
	failAt   string
}

func (f *fakeTmux) Run(argv []string) error {
	f.runs = append(f.runs, argv)
	if argv[1] == "has-session" {
		if f.existing {
			return nil
		}
		return errors.New("sem sessão")
	}
	if f.failAt != "" && argv[1] == f.failAt {
		return errors.New("falhou")
	}
	return nil
}
func (f *fakeTmux) Attach(argv []string) error { f.attached = argv; return nil }

func TestTmux_SessaoExistenteSoConecta(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	plan, err := BuildTmuxPlan(env, config.DefaultForgeConfig(), false)
	require.NoError(t, err)

	f := &fakeTmux{existing: true}
	created, err := LaunchTmux(plan, f)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Len(t, f.runs, 1, "só o has-session: nada é recriado")
	assert.Equal(t, plan.Attach, f.attached)

	f = &fakeTmux{}
	created, err = LaunchTmux(plan, f)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, 1+len(plan.Create), len(f.runs))
	assert.Equal(t, plan.Attach, f.attached)
}

func TestTmux_FalhaNaCriacaoLimpaASessao(t *testing.T) {
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	f := &fakeTmux{failAt: "split-window"}
	_, err = LaunchTmux(plan, f)
	require.Error(t, err)
	assert.Nil(t, f.attached, "não conecta a uma sessão pela metade")
	last := f.runs[len(f.runs)-1]
	assert.Equal(t, []string{"tmux", "kill-session", "-t", plan.Session}, last)
}
