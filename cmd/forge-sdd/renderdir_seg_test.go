package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturaStderr executa fn e devolve o que foi escrito em os.Stderr.
func capturaStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	done := make(chan string)
	go func() { var b bytes.Buffer; _, _ = io.Copy(&b, r); done <- b.String() }()
	fn()
	_ = w.Close()
	os.Stderr = old
	return <-done
}

func initClaude(t *testing.T, dir string) string {
	t.Helper()
	for k, v := range map[string]string{"yes": "true", "name": "p", "stack": "go", "db": "postgres", "agent": "claude"} {
		_ = initCmd.Flags().Set(k, v)
	}
	t.Cleanup(func() {
		_ = initCmd.Flags().Set("yes", "false")
		_ = initCmd.Flags().Set("agent", "")
	})
	var err error
	out := capturaStderr(t, func() {
		rootCmd.SetArgs([]string{"init", dir})
		err = rootCmd.Execute()
	})
	require.NoError(t, err)
	return out
}

// updateAntigo roda `update` num projeto com versão antiga, o que força a re-renderização.
func updateAntigo(t *testing.T, dir string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sdd-version"), []byte("1.0.0"), 0o644))
	_ = updateCmd.Flags().Set("upgrade", "true")
	t.Cleanup(func() { _ = updateCmd.Flags().Set("upgrade", "false") })
	var err error
	out := capturaStderr(t, func() { _, err = updateRun(t, dir, false) })
	require.NoError(t, err)
	return out
}

func vazio(t *testing.T, dir string) {
	t.Helper()
	es, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, es, "nada pode ser gravado fora do projeto")
}

func TestRenderDirInitClaudeSymlinkParaForaNaoGrava(t *testing.T) {
	dir, fora := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".claude")))
	out := initClaude(t, dir)
	vazio(t, fora)
	assert.Contains(t, out, "aviso")
	assert.FileExists(t, filepath.Join(dir, "sdd", ".sddrc")) // o resto instala
	assert.FileExists(t, filepath.Join(dir, ".agents", "skills", "sdd-build", "SKILL.md"))
}

func TestRenderDirUpdateClaudeSymlinkParaForaNaoGrava(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	fora := t.TempDir()
	require.NoError(t, os.RemoveAll(filepath.Join(dir, ".claude")))
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".claude")))
	out := updateAntigo(t, dir)
	vazio(t, fora)
	assert.Contains(t, out, "aviso")
}

func TestRenderDirPaiDoPaiSymlinkNaoCria(t *testing.T) {
	dir, fora := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".agents"))) // .agents/skills/... ainda não existe
	initClaude(t, dir)
	vazio(t, fora)
}

func TestRenderDirArquivoSymlinkNaoSegue(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	alvo := filepath.Join(dir, ".claude", "commands", "revisar.md")
	require.FileExists(t, alvo)
	fora := filepath.Join(t.TempDir(), "alvo.md")
	require.NoError(t, os.WriteFile(fora, []byte("original"), 0o644))
	require.NoError(t, os.Remove(alvo))
	require.NoError(t, os.Symlink(fora, alvo))
	out := updateAntigo(t, dir)
	got, _ := os.ReadFile(fora)
	assert.Equal(t, "original", string(got))
	assert.Contains(t, out, "aviso")
	assert.FileExists(t, filepath.Join(dir, ".claude", "commands", "implementar.md")) // o resto segue
}

func TestRenderDirSymlinkPendenteNaoCriaAlvo(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	alvo := filepath.Join(dir, ".claude", "commands", "revisar.md")
	pendente := filepath.Join(t.TempDir(), "nao-existe.md")
	require.NoError(t, os.Remove(alvo))
	require.NoError(t, os.Symlink(pendente, alvo))
	updateAntigo(t, dir)
	_, err := os.Lstat(pendente)
	assert.True(t, os.IsNotExist(err), "o alvo do link pendente não pode ser criado")
}

func TestRenderDirDiretorioNoLugarDoArquivoRecusado(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	alvo := filepath.Join(dir, ".claude", "commands", "revisar.md")
	require.NoError(t, os.Remove(alvo))
	require.NoError(t, os.Mkdir(alvo, 0o755))
	out := updateAntigo(t, dir)
	fi, err := os.Lstat(alvo)
	require.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.Contains(t, out, "aviso")
}

func TestRenderDirCasoNormalInstalaTudoEPreservaModo(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	for _, p := range []string{".claude/commands/revisar.md", ".claude/agents", ".agents/skills/sdd-build/SKILL.md", "sdd/.sddrc"} {
		assert.NoError(t, func() error { _, err := os.Lstat(filepath.Join(dir, p)); return err }(), p)
	}
	md := filepath.Join(dir, ".claude", "commands", "revisar.md")
	require.NoError(t, os.WriteFile(md, []byte("editado"), 0o600))
	require.NoError(t, os.Chmod(md, 0o600))
	out := updateAntigo(t, dir)
	assert.NotContains(t, out, "aviso: ")
	fi, _ := os.Stat(md)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	data, _ := os.ReadFile(md)
	assert.NotEqual(t, "editado", string(data)) // comando de agente é regerado
}
