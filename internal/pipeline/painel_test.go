package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// painelEnv monta um LaunchEnv de Linux com o mod instalado (ou não) no diretório do teste.
func painelEnv(t *testing.T, comMod bool) LaunchEnv {
	t.Helper()
	dir := t.TempDir()
	if comMod {
		marca := filepath.Join(dir, ".agents", "plugins", "forge-painel", ".claude-plugin", "plugin.json")
		require.NoError(t, os.MkdirAll(filepath.Dir(marca), 0o755))
		require.NoError(t, os.WriteFile(marca, []byte("{}"), 0o644))
	}
	return LaunchEnv{
		GOOS: "linux", Dir: dir, Bin: "/usr/local/bin/forge-sdd",
		LookPath: func(n string) (string, error) { return "/usr/bin/" + n, nil },
		Getenv:   func(string) string { return "" },
		Exists:   func(p string) bool { _, err := os.Stat(p); return err == nil },
	}
}

func TestPainel_PlanoSemJanelaNemTmux(t *testing.T) {
	env := painelEnv(t, true)
	plan, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)

	require.Len(t, plan.Workers, 2)
	assert.Equal(t, []string{"/usr/local/bin/forge-sdd", "worker", "--role", "builder", "--name", "forge-builder", "--cli", "claude"}, plan.Workers[0].Argv)
	assert.Equal(t, []string{"/usr/local/bin/forge-sdd", "worker", "--role", "reviewer", "--name", "forge-reviewer", "--cli", "claude"}, plan.Workers[1].Argv)
	assert.Equal(t, filepath.Join(env.Dir, ".forge", "runtime", "logs", "builder-worker.log"), plan.Workers[0].LogPath)
	assert.Equal(t, filepath.Join(env.Dir, ".forge", "runtime", "logs", "reviewer-worker.log"), plan.Workers[1].LogPath)

	for _, w := range plan.Workers {
		for _, janela := range []string{"tmux", "osascript", "gnome-terminal", "konsole", "x-terminal-emulator", "wt.exe", "cmd.exe"} {
			assert.NotEqual(t, janela, filepath.Base(w.Argv[0]), "modo painel não abre janela nem tmux")
		}
		assert.NotContains(t, strings.Join(w.Argv, " "), "start", "sem shell: argumentos explícitos")
	}

	// a sessão interativa abre com o mod pelo --plugin-dir (spike da 123-01)
	require.GreaterOrEqual(t, len(plan.Discovery), 3)
	assert.Equal(t, "claude", plan.Discovery[0])
	assert.Equal(t, []string{"--plugin-dir", filepath.Join(env.Dir, ".agents", "plugins", "forge-painel")}, plan.Discovery[len(plan.Discovery)-2:])
}

func TestPainel_PlanoNaoTocaODisco(t *testing.T) {
	env := painelEnv(t, true)
	_, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(env.Dir, ".forge"), "montar o plano (--dry-run) não grava nada")
}

func TestPainel_SemModRecusaComOComandoDeInstalacao(t *testing.T) {
	env := painelEnv(t, false)
	_, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPainelSemMod))
	assert.Contains(t, err.Error(), "forge skills instalar plugin forge-painel")
}

func TestPainel_DiscoveryForaDoClaudeERecusada(t *testing.T) {
	env := painelEnv(t, true)
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIGemini}
	_, err := BuildPainelPlan(env, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exige o claude")
}

func TestPainel_DiretorioRelativoERecusado(t *testing.T) {
	env := painelEnv(t, true)
	env.Dir = "relativo"
	_, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.Error(t, err)
}

func TestPainel_TerminalEResolvidoComEmSemTmux(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	assert.True(t, config.ValidTerminal("painel"))
	assert.False(t, config.ValidTerminal("janelas"))

	env := painelEnv(t, true)
	got, err := ResolveTerminal("painel", cfg, env)
	require.NoError(t, err)
	assert.Equal(t, config.TerminalPainel, got)

	env.LookPath = func(string) (string, error) { return "", errors.New("ausente") }
	got, err = ResolveTerminal("painel", cfg, env)
	require.NoError(t, err)
	assert.Equal(t, config.TerminalPainel, got, "o painel não depende do tmux")

	cfg.Terminal = config.TerminalPainel
	got, err = ResolveTerminal("", cfg, env)
	require.NoError(t, err)
	assert.Equal(t, config.TerminalPainel, got, "vale também vindo de forge.json")

	_, err = ResolveTerminal("janelas", cfg, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "painel")
}

func TestPainel_LaunchIniciaCadaWorkerEGravaOPID(t *testing.T) {
	env := painelEnv(t, true)
	plan, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)

	var iniciados []string
	pid := 1000
	res, err := LaunchPainel(plan, env.Dir, func(w PainelWorker, dir string) (int, error) {
		iniciados = append(iniciados, w.Title)
		assert.Equal(t, env.Dir, dir)
		pid++
		return pid, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"forge-builder", "forge-reviewer"}, iniciados)
	assert.Equal(t, map[string]int{"forge-builder": 1001, "forge-reviewer": 1002}, res.Started)
	b, err := os.ReadFile(plan.Workers[0].PIDPath)
	require.NoError(t, err)
	assert.Equal(t, "1001\n", string(b))
	assert.Equal(t, filepath.Join(env.Dir, ".forge", "runtime", "painel", "builder.pid"), plan.Workers[0].PIDPath)
}

func TestPainel_LaunchNaoDuplicaWorkerEmExecucao(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sem sonda de pid no Windows")
	}
	env := painelEnv(t, true)
	plan, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)
	// o builder já tem um pid vivo (este processo de teste); o reviewer não tem
	require.NoError(t, os.MkdirAll(filepath.Dir(plan.Workers[0].PIDPath), 0o755))
	require.NoError(t, os.WriteFile(plan.Workers[0].PIDPath, []byte(strconv.Itoa(os.Getpid())+"\n"), 0o600))

	var iniciados []string
	res, err := LaunchPainel(plan, env.Dir, func(w PainelWorker, _ string) (int, error) {
		iniciados = append(iniciados, w.Title)
		return 4242, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"forge-reviewer"}, iniciados)
	assert.Equal(t, os.Getpid(), res.Already["forge-builder"])
}

func TestPainel_LaunchIgnoraPIDMortoOuLixo(t *testing.T) {
	env := painelEnv(t, true)
	plan, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(plan.Workers[0].PIDPath), 0o755))
	require.NoError(t, os.WriteFile(plan.Workers[0].PIDPath, []byte("não é pid\n"), 0o600))
	require.NoError(t, os.WriteFile(plan.Workers[1].PIDPath, []byte("-5\n"), 0o600))

	n := 0
	_, err = LaunchPainel(plan, env.Dir, func(PainelWorker, string) (int, error) { n++; return 7, nil })
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestPainel_FalhaAoIniciarNaoGravaPIDNemSegue(t *testing.T) {
	env := painelEnv(t, true)
	plan, err := BuildPainelPlan(env, config.DefaultForgeConfig())
	require.NoError(t, err)
	_, err = LaunchPainel(plan, env.Dir, func(PainelWorker, string) (int, error) { return 0, errors.New("sem binário") })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forge-builder")
	assert.NoFileExists(t, plan.Workers[0].PIDPath)
	assert.NoFileExists(t, plan.Workers[1].PIDPath)
}

// TestPainelHelperProcess é o "worker" do teste do iniciador real: escreve e sai.
func TestPainelHelperProcess(t *testing.T) {
	if os.Getenv("FORGE_TEST_PAINEL_HELPER") != "1" {
		return
	}
	os.Stdout.WriteString("olá do worker em segundo plano\n")
	os.Stderr.WriteString("erro do worker\n")
	os.Exit(0)
}

func TestPainel_IniciadorRealGravaAsSaidasNoLog(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o processo auxiliar do teste usa a convenção de processo do Unix")
	}
	t.Setenv("FORGE_TEST_PAINEL_HELPER", "1")
	dir := t.TempDir()
	w := PainelWorker{
		Title:   "forge-builder",
		Argv:    []string{os.Args[0], "-test.run=^TestPainelHelperProcess$"},
		LogPath: filepath.Join(dir, ".forge", "runtime", "logs", "builder-worker.log"),
	}
	pid, err := DefaultPainelStarter(w, dir)
	require.NoError(t, err)
	assert.Greater(t, pid, 0)

	require.Eventually(t, func() bool {
		b, err := os.ReadFile(w.LogPath)
		return err == nil && strings.Contains(string(b), "olá do worker") && strings.Contains(string(b), "erro do worker")
	}, 10*time.Second, 50*time.Millisecond)
	fi, err := os.Stat(w.LogPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(), "o log nasce privado")
}

func TestPainel_IniciadorRealComBinarioInexistenteErra(t *testing.T) {
	dir := t.TempDir()
	_, err := DefaultPainelStarter(PainelWorker{Argv: []string{filepath.Join(dir, "nao-existe")}, LogPath: filepath.Join(dir, "x.log")}, dir)
	require.Error(t, err)
}
