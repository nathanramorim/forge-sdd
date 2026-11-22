package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/require"
)

const specCard = "# Feature 900-01 — Titulo\n\n## Descrição de negócio\nEntrega valor.\n"

func cardProjeto(t *testing.T, ligado bool) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "sdd/discovery/900-t/features/feat-900-01-x.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(specCard), 0o644))
	cfg := config.DefaultForgeConfig()
	cfg.PMMCP.CreateOnConception = ligado
	require.NoError(t, config.SaveForgeConfig(dir, cfg))
	return dir
}

func rodaCard(t *testing.T, dir string, run vcs.Runner, args ...string) (string, error) {
	t.Helper()
	queueCardRunner = run
	t.Cleanup(func() {
		queueCardRunner = nil
		for _, n := range []string{"feature", "provider", "id", "url", "dry-run"} {
			f := queueCardCmd.Flags().Lookup(n)
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
	})
	var out bytes.Buffer
	queueCardCmd.SetOut(&out)
	queueCardCmd.SetErr(&out)
	if queueCardCmd.Flags().Lookup("dir") == nil {
		queueCardCmd.Flags().String("dir", "", "")
	}
	_ = queueCardCmd.Flags().Set("dir", dir)
	require.NoError(t, queueCardCmd.ParseFlags(args))
	err := queueCardCmd.RunE(queueCardCmd, nil)
	return out.String(), err
}

func evPath(dir string) string {
	return filepath.Join(dir, "sdd/discovery/900-t", "card-evidence.json")
}

func ghFake(n *int, falha bool) vcs.Runner {
	return func(dir, name string, args ...string) (string, error) {
		*n++
		if falha {
			return "", errors.New("gh caiu")
		}
		if len(args) > 1 && args[1] == "list" {
			return "[]", nil
		}
		return "https://github.com/o/r/issues/5\n", nil
	}
}

func TestQueueCardDesligadoCustoZero(t *testing.T) {
	dir, n := cardProjeto(t, false), 0
	out, err := rodaCard(t, dir, ghFake(&n, false), "--feature", "feat-900-01-x")
	require.NoError(t, err)
	require.Empty(t, out)
	require.Zero(t, n)
	require.NoFileExists(t, evPath(dir))
}

func TestQueueCardGitHubCriaIssueEGravaEvidencia(t *testing.T) {
	dir, n := cardProjeto(t, true), 0
	out, err := rodaCard(t, dir, ghFake(&n, false), "--feature", "feat-900-01-x")
	require.NoError(t, err)
	require.Contains(t, out, "issues/5")
	require.FileExists(t, evPath(dir))
}

func TestQueueCardIdempotenteNaoDuplica(t *testing.T) {
	dir, n := cardProjeto(t, true), 0
	_, err := rodaCard(t, dir, ghFake(&n, false), "--feature", "feat-900-01-x")
	require.NoError(t, err)
	antes := n
	_, err = rodaCard(t, dir, ghFake(&n, false), "--feature", "feat-900-01-x")
	require.NoError(t, err)
	require.Equal(t, antes, n)
}

func TestQueueCardRegistraRetornoMCP(t *testing.T) {
	dir := cardProjeto(t, true)
	_, err := rodaCard(t, dir, nil, "--feature", "feat-900-01-x", "--provider", "clickup", "--id", "abc1", "--url", "https://app.clickup.com/t/abc1")
	require.NoError(t, err)
	b, _ := os.ReadFile(evPath(dir))
	require.Contains(t, string(b), "clickup")
}

func TestQueueCardDryRunNaoCriaNemGrava(t *testing.T) {
	dir, n := cardProjeto(t, true), 0
	out, err := rodaCard(t, dir, ghFake(&n, false), "--feature", "feat-900-01-x", "--dry-run")
	require.NoError(t, err)
	require.Contains(t, out, "dry-run")
	require.Zero(t, n)
	require.NoFileExists(t, evPath(dir))
}

func TestQueueCardFalhaExternaViraAvisoSemErroFatal(t *testing.T) {
	dir, n := cardProjeto(t, true), 0
	out, err := rodaCard(t, dir, ghFake(&n, true), "--feature", "feat-900-01-x")
	require.NoError(t, err)
	require.Contains(t, out, "aviso")
	require.NoFileExists(t, evPath(dir))
}

func TestQueueCardRecusaCamposInvalidos(t *testing.T) {
	for _, a := range [][]string{
		{"--feature", "../x"},
		{"--feature", "feat-900-01-x", "--provider", "clickup", "--id", "1", "--url", "http://x.com/a"},
		{"--feature", "feat-900-01-x", "--provider", "zzz", "--id", "1", "--url", "https://x.com/a"},
		{"--feature", "feat-900-01-x", "--provider", "clickup", "--id", "a\x1b[31m", "--url", "https://x.com/a"},
		{"--feature", "feat-900-01-x", "--provider", "clickup", "--id", "1", "--url", "https://x.com/a\nb"},
	} {
		dir := cardProjeto(t, true)
		_, err := rodaCard(t, dir, nil, a...)
		require.Error(t, err, strings.Join(a, " "))
		require.NoFileExists(t, evPath(dir))
	}
}
