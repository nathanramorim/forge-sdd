package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// projetoPorFeature cria um repositório em branch com as iniciativas dadas, cada uma com
// uma fila de uma task chamada <pasta>-t, e uma fila legada com a task "legada".
func projetoPorFeature(t *testing.T, branch string, pastas ...string) (dir, legado string) {
	t.Helper()
	t.Setenv("FORGE_INICIATIVA", "")
	layout.SetOverride("")
	t.Cleanup(func() { layout.SetOverride("") })
	dir = t.TempDir()
	testutil.InitRepo(t, dir, branch)
	for _, p := range pastas {
		testutil.WriteFile(t, dir, "sdd/discovery/"+p+"/queue.json", `{"spec_status":"reviewed","items":[{"task_id":"`+p+`-t","title":"t","status":"queued","spec_path":"x.md"}]}`)
	}
	legado = filepath.Join(dir, "sdd", ".queue", "queue.json")
	testutil.WriteFile(t, dir, "sdd/.queue/queue.json", `{"spec_status":"reviewed","items":[{"task_id":"legada","title":"t","status":"queued","spec_path":"x.md"}]}`)
	return dir, legado
}

func lerArquivo(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

func TestQueueNextEmBranchPorFeature(t *testing.T) {
	dir, legado := projetoPorFeature(t, "feat/106-11-parecer-com-head-sha", "106-pr-draft", "107-outra")
	antesLegado := lerArquivo(t, legado)

	saida := captureStdout(t, func() {
		_, err := runRoot(t, "queue", "list", "--dir", dir)
		require.NoError(t, err)
	})
	assert.Contains(t, saida, "106-pr-draft-t")
	assert.NotContains(t, saida, "legada")
	assert.NotContains(t, saida, "Fila vazia")

	saida = captureStdout(t, func() {
		_, err := runRoot(t, "queue", "next", "--dir", dir)
		require.NoError(t, err)
	})
	assert.Contains(t, saida, "106-pr-draft-t")
	assert.Contains(t, lerArquivo(t, filepath.Join(dir, "sdd", "discovery", "106-pr-draft", "queue.json")), `"building"`)
	assert.Contains(t, lerArquivo(t, filepath.Join(dir, "sdd", "discovery", "107-outra", "queue.json")), `"queued"`)
	assert.Equal(t, antesLegado, lerArquivo(t, legado), "a fila legada não é tocada")
}

func TestQueueAmbiguidadeFalhaSemTocarFilaLegada(t *testing.T) {
	dir, legado := projetoPorFeature(t, "fix/106-01-algo", "106-a", "106-b")
	antes := lerArquivo(t, legado)
	for _, args := range [][]string{
		{"queue", "list"},
		{"queue", "next"},
		{"queue", "complete", "legada"},
		{"queue", "review", "legada", "--verdict-json", "{}"},
		{"queue", "enqueue", "--task", "x.md"},
	} {
		saida := captureStdout(t, func() {
			_, err := runRoot(t, append(args, "--dir", dir)...)
			require.Error(t, err, args)
			assert.ErrorIs(t, err, layout.ErrAmbiguousInitiative, args)
			assert.Contains(t, err.Error(), "106-a", args)
			assert.Contains(t, err.Error(), "106-b", args)
		})
		assert.NotContains(t, saida, "legada", args)
	}
	assert.Equal(t, antes, lerArquivo(t, legado), "sdd/.queue/queue.json inalterado")
	assert.Contains(t, lerArquivo(t, filepath.Join(dir, "sdd", "discovery", "106-a", "queue.json")), `"queued"`)
}

func TestQueueIniciativaFlagEEnv(t *testing.T) {
	dir, legado := projetoPorFeature(t, "fix/106-01-algo", "106-a", "106-b", "108-c")
	antes := lerArquivo(t, legado)

	// a flag resolve a ambiguidade
	saida := captureStdout(t, func() {
		_, err := runRoot(t, "queue", "list", "--dir", dir, "--iniciativa", "106-b")
		require.NoError(t, err)
	})
	assert.Contains(t, saida, "106-b-t")
	assert.NotContains(t, saida, "106-a-t")

	// a variável resolve a ambiguidade
	t.Setenv("FORGE_INICIATIVA", "106-a")
	saida = captureStdout(t, func() {
		_, err := runRoot(t, "queue", "list", "--dir", dir)
		require.NoError(t, err)
	})
	assert.Contains(t, saida, "106-a-t")

	// a flag vence a variável
	saida = captureStdout(t, func() {
		_, err := runRoot(t, "queue", "list", "--dir", dir, "--iniciativa", "108-c")
		require.NoError(t, err)
	})
	assert.Contains(t, saida, "108-c-t")
	assert.NotContains(t, saida, "106-a-t")

	// flag inválida ou inexistente: erro, sem cair na branch nem na fila legada
	for _, bad := range []string{"../x", "999-nada"} {
		saida = captureStdout(t, func() {
			_, err := runRoot(t, "queue", "list", "--dir", dir, "--iniciativa", bad)
			require.Error(t, err, bad)
		})
		assert.NotContains(t, saida, "legada", bad)
	}
	assert.Equal(t, antes, lerArquivo(t, legado))
}
