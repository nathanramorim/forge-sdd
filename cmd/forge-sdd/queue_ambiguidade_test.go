package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fotoDoDisco devolve caminho -> conteúdo de tudo fora de .git, para provar que nada foi gravado.
func fotoDoDisco(t *testing.T, dir string) map[string]string {
	t.Helper()
	foto := map[string]string{}
	require.NoError(t, filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		foto[p] = string(b)
		return nil
	}))
	return foto
}

func TestQueuePanelEInterativoFalhamEmAmbiguidade(t *testing.T) {
	dir, legado := projetoPorFeature(t, "fix/106-01-algo", "106-a", "106-b")
	antes := fotoDoDisco(t, dir)
	for _, args := range [][]string{
		{"queue", "panel"},
		{"queue", "panel", "--interactive"},
		{"queue", "--interactive"},
	} {
		_, err := runRoot(t, append(args, "--dir", dir)...)
		require.Error(t, err, args)
		assert.ErrorIs(t, err, layout.ErrAmbiguousInitiative, args)
		assert.Contains(t, err.Error(), "106-a", args)
		assert.Contains(t, err.Error(), "106-b", args)
		assert.Contains(t, err.Error(), "--iniciativa", args)
	}
	assert.Equal(t, antes, fotoDoDisco(t, dir), "nada lido ou gravado")
	assert.Contains(t, lerArquivo(t, legado), "legada")
}

// doctor, status e gate: rc != 0, uma mensagem clara com a causa e a saída, nada gravado.
func TestDoctorStatusGateFalhamClaroEmAmbiguidadeEVariavelInvalida(t *testing.T) {
	for _, cmd := range [][]string{
		{"doctor"},
		{"status"},
		{"gate", "--mode", "bloqueio"},
	} {
		dir, _ := projetoPorFeature(t, "fix/106-01-algo", "106-a", "106-b")
		antes := fotoDoDisco(t, dir)
		saida, err := runRoot(t, append(cmd, "--dir", dir)...)
		require.Error(t, err, cmd)
		assert.ErrorIs(t, err, layout.ErrAmbiguousInitiative, cmd)
		for _, quer := range []string{"106-a", "106-b", "--iniciativa", "FORGE_INICIATIVA"} {
			assert.Contains(t, err.Error(), quer, cmd)
		}
		assert.NotContains(t, saida, "Usage:", "sem uso nem trace: %v", cmd)
		assert.NotContains(t, saida, "goroutine", cmd)
		assert.Equal(t, antes, fotoDoDisco(t, dir), "nada gravado: %v", cmd)

		// variável inválida
		t.Setenv("FORGE_INICIATIVA", "999-nada")
		saida, err = runRoot(t, append(cmd, "--dir", dir)...)
		require.Error(t, err, cmd)
		assert.Contains(t, err.Error(), "FORGE_INICIATIVA", cmd)
		assert.Contains(t, err.Error(), "999-nada", cmd)
		assert.False(t, strings.Contains(saida, "Usage:"), cmd)
		assert.Equal(t, antes, fotoDoDisco(t, dir), "nada gravado: %v", cmd)
		t.Setenv("FORGE_INICIATIVA", "")
	}
}

// forge context <task-id> lê a fila da iniciativa ativa: em ambiguidade falha claro, sem ler a legada.
func TestContextFalhaClaroEmAmbiguidade(t *testing.T) {
	dir, _ := projetoPorFeature(t, "fix/106-01-algo", "106-a", "106-b")
	antes := fotoDoDisco(t, dir)
	saida, err := runRoot(t, "context", "t1", "--dir", dir)
	require.Error(t, err)
	assert.ErrorIs(t, err, layout.ErrAmbiguousInitiative)
	for _, quer := range []string{"106-a", "106-b", "--iniciativa", "FORGE_INICIATIVA"} {
		assert.Contains(t, err.Error(), quer)
	}
	assert.NotContains(t, err.Error(), "não encontrada")
	assert.NotContains(t, saida, "Usage:")
	assert.Equal(t, antes, fotoDoDisco(t, dir))
}
