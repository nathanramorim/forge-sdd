package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Teste de integração com um tmux de verdade. Só roda com FORGE_TMUX_REAL=1 e tmux no PATH; usa um
// servidor isolado (-L) e nunca encosta nas sessões da pessoa.
//
//	FORGE_TMUX_REAL=1 go test ./internal/pipeline -run TestTmuxReal -v
type isolatedTmux struct {
	socket string
	conf   string
}

func (r isolatedTmux) args(argv []string) []string {
	out := []string{"tmux", "-L", r.socket}
	if r.conf != "" {
		out = append(out, "-f", r.conf)
	}
	return append(out, argv[1:]...)
}

func (r isolatedTmux) Run(argv []string) error {
	a := r.args(argv)
	cmd := exec.Command(a[0], a[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s (%w)", strings.Join(argv, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r isolatedTmux) Attach([]string) error { return nil } // sem terminal: não anexa

func (r isolatedTmux) out(t *testing.T, args ...string) string {
	t.Helper()
	a := r.args(append([]string{"tmux"}, args...))
	b, err := exec.Command(a[0], a[1:]...).CombinedOutput()
	require.NoError(t, err, "tmux %v: %s", args, b)
	return strings.TrimSpace(string(b))
}

func realTmux(t *testing.T, conf string) isolatedTmux {
	t.Helper()
	if os.Getenv("FORGE_TMUX_REAL") != "1" {
		t.Skip("defina FORGE_TMUX_REAL=1 para rodar contra um tmux real")
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux não está no PATH")
	}
	r := isolatedTmux{socket: "forge-test-" + strconv.Itoa(os.Getpid()) + "-" + strconv.FormatInt(time.Now().UnixNano()%100000, 10), conf: conf}
	t.Cleanup(func() { _ = r.Run([]string{"tmux", "kill-server"}) })
	return r
}

func realPlan(t *testing.T) (*TmuxPlan, config.ForgeConfig) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultForgeConfig()
	// Nada de agente de verdade: a Discovery vira um `sleep`; o Builder e o Reviewer rodam o worker.
	cfg.CLIProfiles[config.CLIClaude] = config.CLIProfile{Command: "sleep", Args: []string{"1"}, InteractiveArgs: []string{"300"}}
	// O "forge-sdd" dos painéis do Builder e do Reviewer é um script que só dorme: assim os painéis
	// ficam vivos durante o teste (um painel cujo comando sai desaparece e muda a numeração).
	bin := filepath.Join(t.TempDir(), "fake-forge")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\nexec sleep 300\n"), 0o755))
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = dir
	env.Bin = bin
	plan, err := BuildTmuxPlan(env, cfg, false)
	require.NoError(t, err)
	return plan, cfg
}

func panesOf(t *testing.T, r isolatedTmux, session string) [][]string {
	t.Helper()
	raw := r.out(t, "list-panes", "-t", "="+session+":0", "-F", "#{pane_index}|#{pane_title}|#{pane_left}|#{pane_top}|#{pane_width}|#{pane_height}")
	var rows [][]string
	for _, l := range strings.Split(raw, "\n") {
		rows = append(rows, strings.Split(l, "|"))
	}
	return rows
}

func checkLayout(t *testing.T, rows [][]string) {
	t.Helper()
	require.Len(t, rows, 3, "três painéis")
	for i, want := range []string{SessionDiscovery, SessionBuilder, SessionReviewer} {
		assert.Equal(t, strconv.Itoa(i), rows[i][0], "índice do painel %d (pane-base-index fixado em 0)", i)
		assert.Equal(t, want, rows[i][1], "título do painel %d", i)
	}
	left := func(r []string) int { n, _ := strconv.Atoi(r[2]); return n }
	top := func(r []string) int { n, _ := strconv.Atoi(r[3]); return n }
	assert.Equal(t, 0, left(rows[0]), "Discovery à esquerda")
	assert.Greater(t, left(rows[1]), 0, "Builder à direita")
	assert.Equal(t, left(rows[1]), left(rows[2]), "Builder e Reviewer na mesma coluna")
	assert.Less(t, top(rows[1]), top(rows[2]), "Reviewer embaixo do Builder")
}

func TestTmuxReal_PlanoMontaOsTresPaineisNomeados(t *testing.T) {
	r := realTmux(t, "")
	plan, _ := realPlan(t)
	created, err := LaunchTmux(plan, r)
	require.NoError(t, err)
	assert.True(t, created)
	checkLayout(t, panesOf(t, r, plan.Session))
	winName := r.out(t, "list-windows", "-t", "="+plan.Session, "-F", "#{window_name}")
	assert.True(t, strings.HasPrefix(winName, "forge:"), "nome da janela deve começar com forge:")
	assert.Equal(t, "top", r.out(t, "show-window-options", "-t", "="+plan.Session+":0", "-v", "pane-border-status"))
}

func TestTmuxReal_PaneBaseIndex1NoTmuxConfNaoQuebra(t *testing.T) {
	conf := filepath.Join(t.TempDir(), "tmux.conf")
	require.NoError(t, os.WriteFile(conf, []byte("set -g base-index 1\nset -g pane-base-index 1\n"), 0o644))
	r := realTmux(t, conf)
	plan, _ := realPlan(t)
	created, err := LaunchTmux(plan, r)
	require.NoError(t, err, "com pane-base-index 1 os painéis ainda são endereçados como 0, 1 e 2")
	assert.True(t, created)
	checkLayout(t, panesOf(t, r, plan.Session))
}

func TestTmuxReal_SessaoExistenteSoConectaSemDuplicarPaineis(t *testing.T) {
	r := realTmux(t, "")
	plan, _ := realPlan(t)
	_, err := LaunchTmux(plan, r)
	require.NoError(t, err)
	created, err := LaunchTmux(plan, r)
	require.NoError(t, err)
	assert.False(t, created, "a segunda execução só conecta")
	assert.Len(t, panesOf(t, r, plan.Session), 3, "nada foi duplicado")
}

func TestTmuxReal_AlvoExatoNaoCasaPorPrefixo(t *testing.T) {
	r := realTmux(t, "")
	require.NoError(t, r.Run([]string{"tmux", "new-session", "-d", "-s", "forge-app-2abcdef", "sleep", "60"}))
	// Sem "=" o tmux casa "forge-app-2" por prefixo com a sessão de outro projeto.
	assert.NoError(t, r.Run([]string{"tmux", "has-session", "-t", "forge-app-2"}), "sem = casa por prefixo (o problema)")
	assert.Error(t, r.Run([]string{"tmux", "has-session", "-t", "=forge-app-2"}), "com = só casa o nome exato")
	assert.NoError(t, r.Run([]string{"tmux", "has-session", "-t", "=forge-app-2abcdef"}))
}

func TestTmuxReal_FalhaNoMeioLimpaSoASessaoCriada(t *testing.T) {
	r := realTmux(t, "")
	// Uma sessão alheia com o prefixo do nome não pode ser derrubada.
	plan, _ := realPlan(t)
	other := plan.Session + "-outra"
	require.NoError(t, r.Run([]string{"tmux", "new-session", "-d", "-s", other, "sleep", "60"}))
	// Quebra o passo de split do reviewer apontando para um painel que não existe.
	plan.Create[5][5] = "=" + plan.Session + ":forge.9"
	_, err := LaunchTmux(plan, r)
	require.Error(t, err)
	assert.Error(t, r.Run([]string{"tmux", "has-session", "-t", "=" + plan.Session}), "a sessão pela metade foi removida")
	assert.NoError(t, r.Run([]string{"tmux", "has-session", "-t", "=" + other}), "a sessão alheia com o prefixo continua viva")
}

func TestTmuxReal_LivePlanCriaJanelaForgeEControle(t *testing.T) {
	r := realTmux(t, "")
	dir := t.TempDir()
	cfg := config.DefaultForgeConfig()
	cfg.SessionMode = config.SessionModeLive

	fakeCLI := filepath.Join(t.TempDir(), "fake-cli")
	require.NoError(t, os.WriteFile(fakeCLI, []byte("#!/bin/sh\nexec sleep 300\n"), 0o755))
	cfg.CLIProfiles[config.CLIClaude] = config.CLIProfile{Command: fakeCLI, Args: []string{"1"}, InteractiveArgs: []string{}}

	bin := filepath.Join(t.TempDir(), "fake-forge")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\nexec sleep 300\n"), 0o755))
	env := lenv("linux", map[string]bool{"tmux": true}, nil)
	env.Dir = dir
	env.Bin = bin

	plan, err := BuildTmuxPlan(env, cfg, false)
	require.NoError(t, err)

	created, err := LaunchTmux(plan, r)
	require.NoError(t, err)
	assert.True(t, created)

	// Janela principal tem 3 painéis
	checkLayout(t, panesOf(t, r, plan.Session))

	// Confere se as janelas 'forge:<projeto>' e 'forge-ctl' foram criadas
	rawWins := r.out(t, "list-windows", "-t", "="+plan.Session, "-F", "#{window_name}")
	wins := strings.Split(rawWins, "\n")
	assert.Contains(t, wins, "forge:"+TmuxProjectName(dir))
	assert.Contains(t, wins, "forge-ctl")

	// Janela forge-ctl tem 2 painéis (despachantes builder e reviewer)
	ctlPanes := r.out(t, "list-panes", "-t", "="+plan.Session+":forge-ctl", "-F", "#{pane_index}")
	assert.Len(t, strings.Split(ctlPanes, "\n"), 2)
}
