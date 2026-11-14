package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scripted devolve respostas pré-definidas: Selects por título (prefixo) em fila, Confirms em fila.
type scripted struct {
	selects  map[string][]string // prefixo do título -> valores em ordem
	confirms []bool
	defs     map[string]string // título -> default recebido
	pauses   int
	asked    []string
}

func (s *scripted) Select(title string, opts []menuOption, def string) (string, error) {
	s.asked = append(s.asked, title)
	if s.defs == nil {
		s.defs = map[string]string{}
	}
	s.defs[title] = def
	for prefix, q := range s.selects {
		if strings.HasPrefix(title, prefix) {
			if len(q) == 0 {
				return "", errBack
			}
			v := q[0]
			s.selects[prefix] = q[1:]
			if v == "<esc>" {
				return "", errBack
			}
			return v, nil
		}
	}
	return "", errBack
}

func (s *scripted) Confirm(title string, def bool) (bool, error) {
	s.asked = append(s.asked, title)
	if len(s.confirms) == 0 {
		return false, errBack
	}
	v := s.confirms[0]
	s.confirms = s.confirms[1:]
	return v, nil
}

func (s *scripted) Pause(string) error { s.pauses++; return nil }

type runLog struct{ calls [][]string }

func (r *runLog) run(args ...string) error {
	r.calls = append(r.calls, args)
	return nil
}

func allFound() pipeline.PreflightEnv {
	return pipeline.PreflightEnv{
		LookPath: func(n string) (string, error) { return "/usr/bin/" + n, nil },
		Exists:   func(string) bool { return true },
	}
}

func menuFor(t *testing.T, dir string, p prompter, penv pipeline.PreflightEnv) (*menuEnv, *runLog, *bytes.Buffer) {
	t.Helper()
	rl := &runLog{}
	var out bytes.Buffer
	return &menuEnv{dir: dir, out: &out, p: p, run: rl.run, preflightEnv: penv}, rl, &out
}

func TestMenu_PipelineIniciar(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{
		selects: map[string][]string{
			"Forge SDD":                  {valPipeline, "<esc>"},
			"Pipeline multi-sessão":      {valPipeStart, "<esc>"},
			"Agente da sessão":           {config.CLIClaude, config.CLICodex, config.CLIGemini},
			"Permissões do Builder e do": {config.PermissionProject},
		},
		confirms: []bool{true},
	}
	env, rl, out := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))

	cfg, found, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, config.CLIClaude, cfg.CLIFor(config.SessionRoleDiscovery))
	assert.Equal(t, config.CLICodex, cfg.CLIFor(config.SessionRoleBuilder))
	assert.Equal(t, config.CLIGemini, cfg.CLIFor(config.SessionRoleReviewer))
	assert.Equal(t, [][]string{{"up", "--dir", dir}}, rl.calls, "abre pelo mesmo `up`, com a escolha já gravada")
	assert.Contains(t, out.String(), "Checagem antes de abrir")
}

func TestMenu_EscVolta(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{selects: map[string][]string{"Forge SDD": {valPipeline, valStandard, "<esc>"}, "Pipeline multi-sessão": {"<esc>"}, "Estrutura padrão": {valBack}}}
	env, rl, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	assert.Empty(t, rl.calls, "voltar não executa nada")
	assert.Equal(t, 3, countPrefix(p.asked, "Forge SDD"), "Esc nos submenus volta ao principal; Esc no principal sai")
}

func countPrefix(list []string, prefix string) int {
	n := 0
	for _, s := range list {
		if strings.HasPrefix(s, prefix) {
			n++
		}
	}
	return n
}

func TestMenu_BloqueioImpedeAbrirENaoGrava(t *testing.T) {
	dir := pipelineDir(t, nil)
	before, _ := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	penv := allFound()
	penv.LookPath = func(n string) (string, error) {
		if n == "codex" {
			return "", errors.New("não encontrado")
		}
		return "/usr/bin/" + n, nil
	}
	p := &scripted{selects: map[string][]string{
		"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeStart, "<esc>"},
		"Agente da sessão":           {config.CLIClaude, config.CLICodex, config.CLIClaude},
		"Permissões do Builder e do": {config.PermissionProject},
	}, confirms: []bool{true}}
	env, rl, out := menuFor(t, dir, p, penv)
	require.NoError(t, runMenu(env))
	assert.Empty(t, rl.calls, "bloqueio não abre")
	assert.Contains(t, out.String(), "Há bloqueios")
	after, _ := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	assert.Equal(t, string(before), string(after), "nada é gravado quando há bloqueio")
	assert.Equal(t, 1, len(p.confirms), "nenhuma confirmação foi consumida")
}

func TestMenu_AvisoPedeConfirmacao(t *testing.T) {
	dir := pipelineDir(t, nil)
	penv := allFound()
	penv.Exists = func(p string) bool { return !strings.HasSuffix(p, "AGENTS.md") }
	mk := func(confirms ...bool) (*runLog, *scripted) {
		p := &scripted{selects: map[string][]string{
			"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeStart, "<esc>"},
			"Agente da sessão":           {config.CLICodex, config.CLICodex, config.CLICodex},
			"Permissões do Builder e do": {config.PermissionProject},
		}, confirms: confirms}
		env, rl, _ := menuFor(t, dir, p, penv)
		require.NoError(t, runMenu(env))
		return rl, p
	}
	rl, _ := mk(false) // não abre mesmo assim
	assert.Empty(t, rl.calls)
	rl, p := mk(true, true) // aceita o aviso e confirma
	assert.Equal(t, [][]string{{"up", "--dir", dir}}, rl.calls)
	assert.Contains(t, strings.Join(p.asked, "|"), "Há avisos acima")
}

func TestMenu_LembraAEscolha(t *testing.T) {
	dir := pipelineDir(t, func(c *config.ForgeConfig) {
		c.Sessions = map[string]string{config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}
	})
	p := &scripted{selects: map[string][]string{
		"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeStart, "<esc>"},
		"Agente da sessão": {"<esc>"},
	}}
	env, _, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	var builderDef string
	for title, def := range p.defs {
		if strings.Contains(title, "forge-discovery") {
			assert.Equal(t, config.CLIClaude, def, "sem escolha, vale o cli")
		}
		if strings.Contains(title, "forge-builder") {
			builderDef = def
		}
	}
	// Esc na primeira pergunta cancela antes da segunda; a escolha do builder volta marcada na próxima vez.
	p2 := &scripted{selects: map[string][]string{
		"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeStart, "<esc>"},
		"Agente da sessão": {config.CLIClaude, "<esc>"},
	}}
	env2, _, _ := menuFor(t, dir, p2, allFound())
	require.NoError(t, runMenu(env2))
	for title, def := range p2.defs {
		if strings.Contains(title, "forge-builder") {
			builderDef = def
		}
	}
	assert.Equal(t, config.CLICodex, builderDef, "a última escolha vem marcada")
}

func TestMenu_AcoesChamamOsMesmosSubcomandos(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{selects: map[string][]string{
		"Forge SDD":             {valStandard, valPipeline, "<esc>"},
		"Estrutura padrão":      {valStandardStatus, valStandardPanel, valStandardGate, valStandardDoctor, valBack},
		"Pipeline multi-sessão": {valPipeDry, valPipeQueue, valPipeRecover, valBack},
	}}
	env, rl, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	assert.Equal(t, [][]string{
		{"status", "--dir", dir}, {"queue", "panel", "--dir", dir}, {"gate", "--dir", dir}, {"doctor", "--dir", dir},
		{"up", "--dir", dir, "--dry-run"},
		{"bus", "list", "--dir", dir, "--role", "builder"}, {"bus", "list", "--dir", dir, "--role", "reviewer"}, {"inbox", "--dir", dir},
		{"bus", "recover", "--dir", dir, "--role", "builder"}, {"bus", "recover", "--dir", dir, "--role", "reviewer"},
	}, rl.calls)
}

func TestMenu_ProjetoSemForgeOferecaInstalarPrimeiro(t *testing.T) {
	dir := t.TempDir()
	p := &scripted{selects: map[string][]string{"Forge SDD": {valInstall, "<esc>"}}}
	env, rl, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	assert.Equal(t, "init", rl.calls[0][0])
	assert.Equal(t, valInstall, p.defs["Forge SDD "+version+" — o que você quer fazer?"], "instalar é a primeira opção marcada")
}

func TestMenu_PipelineSemForgeJSONOfereceAtivar(t *testing.T) {
	dir := t.TempDir()
	p := &scripted{selects: map[string][]string{"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeStart, "<esc>"}}, confirms: []bool{false}}
	env, rl, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	assert.Empty(t, rl.calls, "recusar ativar não faz nada")
}

func TestMenu_ConfigurarCardEPermissoes(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{
		selects: map[string][]string{
			"Forge SDD": {valPipeline, "<esc>"}, "Pipeline multi-sessão": {valPipeConfig, valBack},
			"Ativar ou reconfigurar":     {"card", "perms", valBack},
			"Permissões do Builder e do": {config.PermissionFull},
		},
		confirms: []bool{false},
	}
	env, _, out := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.False(t, cfg.PMMCP.RequireExistingCard, "projeto sem gerenciador despacha sem card")
	assert.Equal(t, config.PermissionFull, cfg.Permissions)
	assert.Contains(t, out.String(), "Gravado em .forge/config/forge.json")
}

func TestRoot_SemTerminalImprimeAjudaSemPerguntar(t *testing.T) {
	old := isInteractive
	isInteractive = func() bool { return false }
	defer func() { isInteractive = old }()
	rootCmd.SetArgs(nil)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	defer rootCmd.SetOut(nil)
	require.NoError(t, rootCmd.RunE(rootCmd, nil))
	assert.Contains(t, out.String(), "Sem terminal interativo")
	assert.Contains(t, out.String(), "forge-sdd pipeline init")
}

func TestRunSubcommand_ExecutaORunEDoSubcomando(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, runSubcommand("pipeline", "init", "--dir", dir, "--dry-run"))
	_, err := os.Stat(filepath.Join(dir, ".forge", "config", "forge.json"))
	assert.Error(t, err, "o dry-run do subcomando é respeitado")
	assert.Error(t, runSubcommand("naoexiste"))
	// As flags de uma chamada anterior não vazam para a seguinte.
	require.NoError(t, runSubcommand("pipeline", "init", "--dir", dir))
	_, err = os.Stat(filepath.Join(dir, ".forge", "config", "forge.json"))
	assert.NoError(t, err)
}
