package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initAgentes(t *testing.T, agentes string, args ...string) (string, error) {
	t.Helper()
	for k, v := range map[string]string{"yes": "true", "name": "p", "stack": "go", "db": "postgres", "agent": agentes} {
		_ = initCmd.Flags().Set(k, v)
	}
	t.Cleanup(func() { _ = initCmd.Flags().Set("yes", "false"); _ = initCmd.Flags().Set("agent", "") })
	var err error
	out := capturaStderr(t, func() {
		rootCmd.SetArgs(append([]string{"init"}, args...))
		err = rootCmd.Execute()
	})
	return out, err
}

func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// arvore lista caminho, modo e conteúdo de cada item do projeto.
func arvore(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	require.NoError(t, filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == dir {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		fi, _ := d.Info()
		item := fmt.Sprintf("%s %v", rel, fi.Mode())
		if fi.Mode().IsRegular() {
			b, _ := os.ReadFile(p)
			item += fmt.Sprintf(" %d %x", len(b), b)
		}
		out = append(out, item)
		return nil
	}))
	return out
}

func TestInitPontoIdenticoAoCaminhoAbsoluto(t *testing.T) {
	for _, ag := range []string{"claude", "gemini", "copilot", "openai", "claude,gemini,copilot,openai"} {
		t.Run(ag, func(t *testing.T) {
			abs, ponto := t.TempDir(), filepath.Join(t.TempDir(), "com espaço")
			require.NoError(t, os.Mkdir(ponto, 0o755))
			_, err := initAgentes(t, ag, abs)
			require.NoError(t, err)
			chdirTemp(t, ponto)
			out, err := initAgentes(t, ag, ".")
			require.NoError(t, err)
			assert.NotContains(t, out, "aviso")
			esperado, obtido := arvore(t, abs), arvore(t, ponto)
			assert.NotEmpty(t, esperado)
			assert.Equal(t, esperado, obtido)
		})
	}
}

func TestUpdatePontoIdenticoAoCaminhoAbsoluto(t *testing.T) {
	abs, ponto := t.TempDir(), t.TempDir()
	for _, d := range []string{abs, ponto} {
		_, err := initAgentes(t, "claude,gemini", d)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(d, "sdd", ".sdd-version"), []byte("1.0.0"), 0o644))
	}
	_ = updateCmd.Flags().Set("upgrade", "true")
	t.Cleanup(func() { _ = updateCmd.Flags().Set("upgrade", "false") })
	_, err := updateRun(t, abs, false)
	require.NoError(t, err)
	chdirTemp(t, ponto)
	out := capturaStderr(t, func() { _, err = updateRun(t, ".", false) })
	require.NoError(t, err)
	assert.NotContains(t, out, "aviso")
	assert.Equal(t, arvore(t, abs), arvore(t, ponto))
}

func TestInitPontoClaudeSimbolicoRecusaSoOQueDeve(t *testing.T) {
	dir, fora := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".claude")))
	chdirTemp(t, dir)
	out, err := initAgentes(t, "claude", ".")
	require.NoError(t, err)
	vazio(t, fora)
	assert.Contains(t, out, ".claude")
	assert.FileExists(t, filepath.Join(dir, "sdd", ".sddrc"))
	assert.FileExists(t, filepath.Join(dir, ".agents", "skills", "sdd-build", "SKILL.md"))
}

func TestInitTodosRecusadosSaiNaoZero(t *testing.T) {
	dir, fora := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, "sdd")))
	chdirTemp(t, dir)
	_, err := initAgentes(t, "claude", ".")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SEG-02")
	vazio(t, fora)
}

func TestPipelineInitForgeSimbolicoNaoGrava(t *testing.T) {
	dir, fora := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".forge")))
	rootCmd.SetArgs([]string{"pipeline", "init", "--dir", dir})
	err := rootCmd.Execute()
	require.Error(t, err)
	vazio(t, fora)
}

func TestAvisoSanitizaCaminhoComEscEBidi(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "p\x1b[31mX‮Y")
	require.NoError(t, os.Mkdir(dir, 0o755))
	require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(dir, ".claude")))
	out, err := initAgentes(t, "claude", dir)
	require.NoError(t, err)
	require.Contains(t, out, "aviso")
	assert.False(t, strings.ContainsAny(out, "\x1b‮"), "ESC e bidi não podem chegar ao terminal")
}
