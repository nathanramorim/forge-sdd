package main

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gestureScript(gesture string, confirms ...bool) *scripted {
	return &scripted{
		selects: map[string][]string{
			"Forge SDD":                  {valUp, "<esc>"},
			"Agente da sessão":           {config.CLIClaude, config.CLIClaude, config.CLIClaude},
			"Permissões do Builder e do": {config.PermissionProject},
		},
		gestures: []string{gesture},
		confirms: confirms,
	}
}

func TestMenu_UpGesto_DireitaLiga(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := gestureScript(survey.GestureRight, true, true) // ambiente isolado, abrir agora
	env, rl, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))

	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.IsAdvanced())
	assert.Equal(t, config.PermissionFull, cfg.Permissions)
	assert.Equal(t, [][]string{{"up", "--dir", dir}}, rl.calls)
	var abrir string
	for _, a := range p.asked {
		if strings.HasPrefix(a, "Abrir agora?") {
			abrir = a
		}
		if strings.HasPrefix(a, "Forge SDD") {
			assert.NotContains(t, a, "avançado", "o menu não exibe o modo antes de ligado")
		}
	}
	assert.Contains(t, abrir, "avançado", "a linha de confirmação mostra o modo")
}

func TestMenu_UpGesto_EsquerdaDesliga(t *testing.T) {
	dir := pipelineDir(t, func(c *config.ForgeConfig) {
		c.Autonomy, c.IsolatedConfirmed, c.Permissions = config.AutonomyAdvanced, true, config.PermissionFull
	})
	p := gestureScript(survey.GestureLeft, true)
	p.selects["Permissões do Builder e do"] = []string{config.PermissionFull}
	env, _, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))

	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.False(t, cfg.IsAdvanced())
	assert.False(t, cfg.IsolatedConfirmed)
}

func TestMenu_Autonomy_SemConfirmacaoNaoGrava(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := gestureScript(survey.GestureRight, false)
	env, rl, out := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))

	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.False(t, cfg.IsAdvanced())
	assert.Empty(t, rl.calls)
	assert.Contains(t, out.String(), "nada foi gravado")
}

func TestMenu_Autonomy_Subcomando(t *testing.T) {
	dir := pipelineDir(t, nil)
	require.Error(t, runSubcommandIn(t, dir, "advanced"), "sem --isolated recusa")
	require.NoError(t, runSubcommandIn(t, dir, "advanced", "--isolated"))
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.IsAdvanced())
	require.NoError(t, runSubcommandIn(t, dir, "normal"))
	cfg, _, _ = config.LoadForgeConfig(dir)
	assert.False(t, cfg.IsAdvanced())
	require.Error(t, runSubcommandIn(t, dir, "turbo"))
}

func runSubcommandIn(t *testing.T, dir string, args ...string) error {
	t.Helper()
	return runSubcommand(append([]string{"pipeline", "autonomy", "--dir", dir}, args...)...)
}

func TestMenu_Settings_AlternaCard(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{selects: map[string][]string{
		"Forge SDD": {valSettings, "<esc>"},
		"Settings":  {"card", valBack},
	}}
	env, _, out := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.PMMCP.RequireExistingCard, "alterna de desativado para ativado")
	assert.Contains(t, out.String(), "ativado")
}

func TestMenu_Settings_TetoDiario(t *testing.T) {
	dir := pipelineDir(t, nil)
	p := &scripted{selects: map[string][]string{
		"Forge SDD": {valSettings, "<esc>"},
		"Settings":  {"teto-diario", "teto-diario", valBack},
	}}
	env, _, _ := menuFor(t, dir, p, allFound())
	require.NoError(t, runMenu(env))
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.Zero(t, cfg.Safeguards.MaxJobsPerDay, "padrão desligado; duas alternâncias voltam a desligado")
}

func TestPipelineInit_PerguntaCard(t *testing.T) {
	oldI, oldA := isInteractive, askCard
	defer func() { isInteractive, askCard = oldI, oldA }()
	isInteractive = func() bool { return true }
	askCard = func() (bool, error) { return true, nil }

	dir := t.TempDir()
	require.NoError(t, runSubcommand("pipeline", "init", "--dir", dir))
	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.PMMCP.RequireExistingCard, "grava a resposta")

	isInteractive = func() bool { return false }
	dir2 := t.TempDir()
	require.NoError(t, runSubcommand("pipeline", "init", "--dir", dir2))
	cfg, _, _ = config.LoadForgeConfig(dir2)
	assert.False(t, cfg.PMMCP.RequireExistingCard, "sem terminal grava false")
}
