package pipeline

import (
	"errors"
	"fmt"
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
	name := TmuxSessionName(env.Dir)
	assert.Regexp(t, `^forge-meu-app-[0-9a-f]{6}$`, name)
	assert.Equal(t, name, plan.Session)
	win := "=" + name + ":forge:meu-app"

	c := plan.Create
	require.GreaterOrEqual(t, len(c), 13)
	assert.Equal(t, []string{"tmux", "new-session", "-d", "-s", name, "-n", "forge:meu-app", "-c", "/work/meu-app"}, c[0][:9])
	assert.Contains(t, c[0][9], "claude -n forge-discovery --agent lider", "painel da Discovery")
	assert.Equal(t, []string{"tmux", "set-window-option", "-t", win, "pane-base-index", "0"}, c[3], "painéis endereçáveis mesmo com pane-base-index 1 no tmux.conf")
	assert.Equal(t, []string{"tmux", "split-window", "-h", "-t", win, "-c", "/work/meu-app"}, c[4][:7])
	assert.Contains(t, c[4][7], "worker --role builder --name forge-builder --cli codex", "Builder à direita no alto")
	assert.Equal(t, []string{"tmux", "split-window", "-v", "-t", win + ".1", "-c", "/work/meu-app"}, c[5][:7])
	assert.Contains(t, c[5][7], "worker --role reviewer --name forge-reviewer --cli claude", "Reviewer embaixo")
	assert.Equal(t, []string{"tmux", "set-window-option", "-t", win, "pane-border-status", "top"}, c[6])
	var titles []string
	for _, step := range c {
		if step[1] == "select-pane" && len(step) > 4 && step[4] == "-T" {
			titles = append(titles, step[5])
		}
	}
	assert.Equal(t, []string{"forge-discovery", "forge-builder", "forge-reviewer"}, titles)
	assert.Equal(t, []string{"tmux", "select-pane", "-t", win + ".0"}, c[len(c)-1], "o foco termina na Discovery")
	assert.Equal(t, []string{"tmux", "attach-session", "-t", "=" + name}, plan.Attach, "alvo exato: sem casar por prefixo com outra sessão")

	inside, err := BuildTmuxPlan(env, cfg, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"tmux", "switch-client", "-t", "=" + name}, inside.Attach, "dentro do tmux troca o cliente")
}

func TestTmuxLivePlan(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = "/work/meu-app"
	cfg := config.DefaultForgeConfig()
	cfg.SessionMode = config.SessionModeLive
	cfg.Sessions = map[string]string{
		config.SessionRoleBuilder:  config.CLICodex,
		config.SessionRoleReviewer: config.CLIGemini,
	}

	plan, err := BuildTmuxPlan(env, cfg, false)
	require.NoError(t, err)

	name := TmuxSessionName(env.Dir)
	win := "=" + name + ":forge:meu-app"
	winCtl := "=" + name + ":forge-ctl"

	c := plan.Create
	// Janela principal: Discovery, Builder CLI, Reviewer CLI
	assert.Contains(t, c[0][9], "claude -n forge-discovery --agent lider", "painel 0 da Discovery")
	assert.Contains(t, c[4][7], "codex", "painel 1 do Builder com CLI interativa")
	assert.Contains(t, c[4][7], "--sandbox workspace-write", "permissões project do codex presentes")
	assert.Contains(t, c[5][7], "agy", "painel 2 do Reviewer com CLI interativa")
	assert.NotContains(t, c[5][7], "-i", "sem prompt inicial não passa -i órfão")
	assert.Contains(t, c[5][7], "--dangerously-skip-permissions", "permissões do agy presentes")

	// Janela oculta forge-ctl com os despachantes
	foundNewWin := false
	var ctlCmds []string
	for _, step := range c {
		if step[1] == "new-window" {
			foundNewWin = true
			assert.Equal(t, "forge-ctl", step[4])
			ctlCmds = append(ctlCmds, step[len(step)-1])
		}
		if step[1] == "split-window" && len(step) > 4 && step[4] == winCtl {
			ctlCmds = append(ctlCmds, step[len(step)-1])
		}
	}
	require.True(t, foundNewWin, "deve criar janela forge-ctl")
	require.Len(t, ctlCmds, 2, "deve ter despachante do builder e do reviewer")
	assert.Contains(t, ctlCmds[0], fmt.Sprintf("worker --role builder --name forge-builder --cli codex --deliver %s", shQuote(win+".1")))
	assert.Contains(t, ctlCmds[1], fmt.Sprintf("worker --role reviewer --name forge-reviewer --cli gemini --deliver %s", shQuote(win+".2")))

	// Modo headless permanece idêntico
	cfgHeadless := config.DefaultForgeConfig()
	cfgHeadless.SessionMode = config.SessionModeHeadless
	cfgHeadless.Sessions = cfg.Sessions
	planHeadless, err := BuildTmuxPlan(env, cfgHeadless, false)
	require.NoError(t, err)
	assert.NotContains(t, planHeadless.Create[4][7], "--sandbox", "headless roda worker, não CLI interativa com sandbox no argv externo")
	assert.Contains(t, planHeadless.Create[4][7], "worker --role builder")
}

func TestTmuxPlanProjectIdentity(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = "/work/meu-app"
	cfg := config.DefaultForgeConfig()
	plan, err := BuildTmuxPlan(env, cfg, false)
	require.NoError(t, err)

	name := TmuxSessionName(env.Dir)
	proj := TmuxProjectName(env.Dir)
	assert.Equal(t, "meu-app", proj)

	exact := "=" + name
	winName := "forge:" + proj

	var foundStatusLeft, foundStatusLength, foundBorderFormat, foundWindowName bool
	for _, cmd := range plan.Create {
		assert.NotContains(t, cmd, "-g", "nunca deve usar opções globais (-g)")
		if cmd[1] == "new-session" {
			for i, arg := range cmd {
				if arg == "-n" && i+1 < len(cmd) && cmd[i+1] == winName {
					foundWindowName = true
				}
			}
		}
		if cmd[1] == "set-option" && len(cmd) >= 5 {
			assert.Equal(t, exact, cmd[3], "alvo de set-option deve ser a sessão exata")
			if cmd[4] == "status-left" {
				foundStatusLeft = true
				assert.Contains(t, cmd[5], proj)
				assert.Contains(t, cmd[5], "📦")
			}
			if cmd[4] == "status-left-length" {
				foundStatusLength = true
				assert.Equal(t, "60", cmd[5])
			}
		}
		if cmd[1] == "set-window-option" && len(cmd) >= 5 {
			if cmd[4] == "pane-border-format" {
				foundBorderFormat = true
				assert.Contains(t, cmd[5], proj+" ❯ #{pane_title}")
			}
		}
	}
	assert.True(t, foundWindowName, "nome da janela deve ser forge:<projeto>")
	assert.True(t, foundStatusLeft, "status-left deve conter o nome do projeto")
	assert.True(t, foundStatusLength, "status-left-length deve ser 60")
	assert.True(t, foundBorderFormat, "pane-border-format deve conter <projeto> ❯ #{pane_title}")
}

func TestTmux_NomeDaSessaoSaneadoEUnicoPorCaminho(t *testing.T) {
	for in, base := range map[string]string{
		"/p/meu app":                    "forge-meu-app-",
		"/p/a:b.c":                      "forge-a-b-c-",
		"/p/x;rm -rf ~":                 "forge-x-rm--rf-",
		"/p/" + strings.Repeat("a", 80): "forge-" + strings.Repeat("a", 40) + "-",
		"/":                             "forge-projeto-",
	} {
		got := TmuxSessionName(in)
		assert.True(t, strings.HasPrefix(got, base), "%q -> %q", in, got)
		assert.Regexp(t, `^forge-[A-Za-z0-9_-]+-[0-9a-f]{6}$`, got)
	}
	assert.NotEqual(t, TmuxSessionName("/a/app"), TmuxSessionName("/b/app"), "duas pastas com o mesmo nome não dividem a sessão")
	assert.Equal(t, TmuxSessionName("/a/app"), TmuxSessionName("/a/app/"), "determinístico")
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
	existing  bool
	runs      [][]string
	attached  []string
	failAt    string
	attachErr error
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
func (f *fakeTmux) Attach(argv []string) error {
	f.attached = argv
	return f.attachErr
}

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
	assert.Equal(t, []string{"tmux", "has-session", "-t", "=" + plan.Session}, f.runs[0], "alvo exato")
	assert.Equal(t, plan.Attach, f.attached)
}

func TestTmux_FalhaDepoisDeCriarLimpaSoOQueCriou(t *testing.T) {
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	f := &fakeTmux{failAt: "split-window"}
	_, err = LaunchTmux(plan, f)
	require.Error(t, err)
	assert.Nil(t, f.attached, "não conecta a uma sessão pela metade")
	assert.Equal(t, []string{"tmux", "kill-session", "-t", "=" + plan.Session}, f.runs[len(f.runs)-1], "alvo exato")
}

func TestTmux_FalhaNoPrimeiroPassoNaoMataNada(t *testing.T) {
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	f := &fakeTmux{failAt: "new-session"}
	_, err = LaunchTmux(plan, f)
	require.Error(t, err)
	for _, run := range f.runs {
		assert.NotEqual(t, "kill-session", run[1], "nada foi criado: nada a limpar, e nunca a sessão de outro")
	}
}

type racingTmux struct {
	fakeTmux
	hasCalls int
}

func (r *racingTmux) Run(argv []string) error {
	if argv[1] == "has-session" {
		r.hasCalls++
		r.runs = append(r.runs, argv)
		if r.hasCalls == 1 {
			return errors.New("sem sessão") // ainda não existia
		}
		return nil // outra execução criou no mesmo instante
	}
	if argv[1] == "new-session" {
		r.runs = append(r.runs, argv)
		return errors.New("duplicate session")
	}
	return r.fakeTmux.Run(argv)
}

func TestTmux_CorridaEntreDoisUpsSoConecta(t *testing.T) {
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	r := &racingTmux{}
	created, err := LaunchTmux(plan, r)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, plan.Attach, r.attached)
	for _, run := range r.runs {
		assert.NotEqual(t, "kill-session", run[1], "a sessão que a outra execução criou não é derrubada")
	}
}

func TestTmux_DiretorioRelativoERecusado(t *testing.T) {
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = "."
	_, err := BuildTmuxPlan(env, config.DefaultForgeConfig(), false)
	assert.Error(t, err, "com . o hash seria igual para projetos distintos")
}

func TestTmux_FalhaAoConectarDizOQueFazer(t *testing.T) {
	plan, err := BuildTmuxPlan(lenv("linux", map[string]bool{"tmux": true}, nil), config.DefaultForgeConfig(), false)
	require.NoError(t, err)
	f := &fakeTmux{attachErr: errors.New("exit status 1")}
	created, err := LaunchTmux(plan, f)
	require.Error(t, err)
	assert.True(t, created, "a sessão foi criada mesmo sem conseguir conectar")
	assert.Contains(t, err.Error(), "tmux attach -t =forge-")
	assert.Contains(t, err.Error(), "está pronta")
}
