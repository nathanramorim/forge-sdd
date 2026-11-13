package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	fn()
	require.NoError(t, w.Close())
	os.Stdout = old
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func TestPrintPipelineHint_SemModoAtivoMostraOPassoAPasso(t *testing.T) {
	dir := t.TempDir()
	out := captureStdout(t, func() { printPipelineHint(dir) })
	assert.Contains(t, out, "forge-sdd pipeline init")
	assert.Contains(t, out, "forge-sdd up --cli claude")
	assert.Contains(t, out, "/implementar <feature>")
	assert.Contains(t, out, ".forge/LEIA-ME.md")
}

func TestPrintPipelineHint_ComModoAtivoSoLembraDeAbrirAsSessoes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
	resetWorkerFlags()
	_, err := runRoot(t, "pipeline", "init", "--dir", dir)
	require.NoError(t, err)
	out := captureStdout(t, func() { printPipelineHint(dir) })
	assert.Contains(t, out, "ATIVO")
	assert.NotContains(t, out, "pipeline init")
}

func TestPipelineInit_MostraProximosPassos(t *testing.T) {
	dir := t.TempDir()
	resetWorkerFlags()
	out, err := runRoot(t, "pipeline", "init", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "Próximos passos")
	assert.Contains(t, out, "forge-sdd up --cli claude")
	assert.Contains(t, out, "/implementar <feature>")

	resetWorkerFlags()
	out, err = runRoot(t, "pipeline", "init", "--dir", t.TempDir(), "--dry-run")
	require.NoError(t, err)
	assert.NotContains(t, out, "Próximos passos", "o dry-run não ativa o modo")
}
