package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// Modo painel do `forge up` (iniciativa 123, RN-13): nenhuma janela nem tmux. Builder e Reviewer
// rodam em segundo plano, com a saída em arquivo; a sessão interativa do Claude Code com o mod
// forge-painel (ADR-0030) é o único ponto de acompanhamento e lê o barramento e os logs.

// PainelPluginSlug é o mod que o modo painel exige instalado.
const PainelPluginSlug = "forge-painel"

// ErrPainelSemMod é devolvido quando o mod não está instalado no projeto.
var ErrPainelSemMod = errors.New("o modo painel precisa do mod " + PainelPluginSlug)

// PainelWorker é um worker a iniciar em segundo plano: argv explícito (SEG-03), sem shell.
type PainelWorker struct {
	Title   string
	Role    string
	Argv    []string
	LogPath string
	PIDPath string
}

// PainelPlan é o que o `forge up --terminal painel` faz: os workers e o comando que a pessoa
// roda para abrir o Claude Code com o mod. Nenhum dos dois é uma janela.
type PainelPlan struct {
	Workers   []PainelWorker
	Discovery []string
	PluginDir string
}

func painelPluginDir(dir string) string {
	return filepath.Join(dir, ".agents", "plugins", PainelPluginSlug)
}

// PainelModInstalado diz se o mod está em <projeto>/.agents/plugins/forge-painel (marca: plugin.json).
func PainelModInstalado(env LaunchEnv) bool {
	exists := env.Exists
	if exists == nil {
		exists = func(p string) bool { _, err := os.Stat(p); return err == nil }
	}
	return exists(filepath.Join(painelPluginDir(env.Dir), ".claude-plugin", "plugin.json"))
}

// BuildPainelPlan monta o plano sem tocar o disco (ARQ-05): recusa sem o mod instalado, recusa
// quando a sessão Discovery não é o Claude Code (o mod é exclusivo dele, RN-6) e devolve o argv de
// cada worker e o do Claude com `--plugin-dir`.
func BuildPainelPlan(env LaunchEnv, cfg config.ForgeConfig) (*PainelPlan, error) {
	if env.Dir == "" || !filepath.IsAbs(env.Dir) {
		return nil, fmt.Errorf("o diretório do projeto precisa ser absoluto no modo painel (recebeu %q)", env.Dir)
	}
	if env.Bin == "" {
		return nil, errors.New("binário do forge-sdd vazio")
	}
	if !PainelModInstalado(env) {
		return nil, fmt.Errorf("%w: instale com `forge skills instalar plugin %s` (o mod vem no binário) e rode `forge up --terminal painel` de novo (sem o mod não há como acompanhar os workers)", ErrPainelSemMod, PainelPluginSlug)
	}
	discoveryCLI := cfg.CLIFor(config.SessionRoleDiscovery)
	if discoveryCLI != config.CLIClaude {
		return nil, fmt.Errorf("o modo painel exige o claude na sessão discovery (o mod é exclusivo do Claude Code; atual: %q). Use --terminal native ou tmux com outros agentes", discoveryCLI)
	}
	profile, ok := cfg.Profile(discoveryCLI)
	if !ok {
		return nil, fmt.Errorf("cli_profiles sem perfil para %q", discoveryCLI)
	}
	spec := RunSpec{Agent: "lider", Name: SessionDiscovery, Prompt: DiscoveryPrompt}
	spec.Model, spec.Effort = cfg.SessionModelArgs(config.SessionRoleDiscovery)
	args, err := expandArgs(profile.InteractiveArgs, spec)
	if err != nil {
		return nil, err
	}
	plugin := painelPluginDir(env.Dir)
	plan := &PainelPlan{
		PluginDir: plugin,
		Discovery: append(append([]string{profile.Command}, args...), "--plugin-dir", plugin),
	}
	logs := layout.Logs(env.Dir)
	painel := filepath.Join(layout.ForgeRuntime(env.Dir), "painel")
	for _, w := range []struct{ role, name string }{
		{config.SessionRoleBuilder, SessionBuilder},
		{config.SessionRoleReviewer, SessionReviewer},
	} {
		cli := cfg.CLIFor(w.role)
		if !config.KnownCLI(cli) {
			return nil, fmt.Errorf("cli %q inválida para %s (use %s)", cli, w.role, strings.Join(config.OfferedCLIs(), ", "))
		}
		if _, ok := cfg.Profile(cli); !ok {
			return nil, fmt.Errorf("cli_profiles sem perfil para %q", cli)
		}
		plan.Workers = append(plan.Workers, PainelWorker{
			Title:   w.name,
			Role:    w.role,
			Argv:    []string{env.Bin, "worker", "--role", w.role, "--name", w.name, "--cli", cli},
			LogPath: filepath.Join(logs, w.role+"-worker.log"),
			PIDPath: filepath.Join(painel, w.role+".pid"),
		})
	}
	return plan, nil
}

// PainelStarter inicia um processo destacado e devolve o pid. Injetável nos testes.
type PainelStarter func(w PainelWorker, dir string) (int, error)

// PainelResult diz o que LaunchPainel fez com cada worker.
type PainelResult struct {
	Started  map[string]int // título -> pid
	Already  map[string]int // título -> pid já em execução
	LogPaths map[string]string
}

// LaunchPainel inicia os workers em segundo plano. Worker cujo pid registrado ainda vive não é
// iniciado de novo; reabrir o Claude ou rodar `forge up` outra vez não duplica a esteira.
func LaunchPainel(plan *PainelPlan, dir string, start PainelStarter) (PainelResult, error) {
	res := PainelResult{Started: map[string]int{}, Already: map[string]int{}, LogPaths: map[string]string{}}
	if start == nil {
		start = DefaultPainelStarter
	}
	for _, w := range plan.Workers {
		if len(w.Argv) == 0 {
			return res, fmt.Errorf("comando vazio para %s", w.Title)
		}
		if pid, alive := readLivePID(w.PIDPath); alive {
			res.Already[w.Title] = pid
			res.LogPaths[w.Title] = w.LogPath
			continue
		}
		pid, err := start(w, dir)
		if err != nil {
			return res, fmt.Errorf("iniciar %s em segundo plano: %w", w.Title, err)
		}
		if err := os.MkdirAll(filepath.Dir(w.PIDPath), 0o755); err == nil {
			_ = os.WriteFile(w.PIDPath, []byte(strconv.Itoa(pid)+"\n"), 0o600)
		}
		res.Started[w.Title] = pid
		res.LogPaths[w.Title] = w.LogPath
	}
	return res, nil
}

func readLivePID(path string) (int, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, processAlive(pid)
}

// DefaultPainelStarter abre o log (anexando), inicia o worker destacado do terminal e solta o
// processo: fechar o Claude ou o terminal não o encerra. Argumentos explícitos, sem shell (SEG-03).
func DefaultPainelStarter(w PainelWorker, dir string) (int, error) {
	if err := os.MkdirAll(filepath.Dir(w.LogPath), 0o755); err != nil {
		return 0, err
	}
	log, err := os.OpenFile(w.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, err
	}
	defer log.Close()
	cmd := exec.Command(w.Argv[0], w.Argv[1:]...)
	cmd.Dir = dir
	cmd.Env = EnvComEsteBinario()
	cmd.Stdout = log
	cmd.Stderr = log
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()
	return pid, nil
}

// EnvComEsteBinario devolve o ambiente com a pasta deste binário à frente do PATH: o Claude e os
// workers abertos pelo forge up chamam `forge` e `forge-sdd` pelo nome e precisam achar esta mesma
// versão, não outra instalada antes no PATH.
func EnvComEsteBinario() []string {
	pasta := filepath.Dir(ResolveBin())
	env := os.Environ()
	if !filepath.IsAbs(pasta) {
		return env
	}
	for i, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			env[i] = "PATH=" + pasta + string(os.PathListSeparator) + strings.TrimPrefix(kv, "PATH=")
			return env
		}
	}
	return append(env, "PATH="+pasta)
}
