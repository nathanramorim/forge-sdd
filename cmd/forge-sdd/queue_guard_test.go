package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func queueProject(t *testing.T, pipeline bool) string {
	t.Helper()
	dir := t.TempDir()
	if pipeline {
		dir = pipelineDir(t, nil)
	}
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	return dir
}

func resetQueueFlags() {
	_ = queueNextCmd.Flags().Set("local", "false")
	_ = queueNextCmd.Flags().Set("json", "false")
	_ = queueCompleteCmd.Flags().Set("local", "false")
}

// liberado: o comando passou pela guarda (o erro, se houver, é o da fila, não o do modo pipeline).
func liberado(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		assert.NotContains(t, err.Error(), "modo pipeline")
		assert.Contains(t, err.Error(), "fila", "o erro vem da fila de tasks, que não existe neste projeto de teste")
	}
}

func TestQueue_GuardaModoPipeline(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := queueProject(t, true)

	for _, args := range [][]string{{"queue", "next", "--dir", dir}, {"queue", "complete", "70-01", "--dir", dir}} {
		resetQueueFlags()
		_, err := runRoot(t, args...)
		require.Error(t, err, args)
		assert.Contains(t, err.Error(), "modo pipeline", args)
		assert.Contains(t, err.Error(), "forge-sdd dispatch", args)
	}

	// --local libera os dois comandos.
	resetQueueFlags()
	_, err := runRoot(t, "queue", "next", "--dir", dir, "--local")
	liberado(t, err)
	resetQueueFlags()
	_, err = runRoot(t, "queue", "complete", "70-01", "--dir", dir, "--local")
	liberado(t, err)

	// Sob um worker (FORGE_WORKER) também liberam.
	t.Setenv("FORGE_WORKER", "forge-builder")
	resetQueueFlags()
	_, err = runRoot(t, "queue", "next", "--dir", dir)
	liberado(t, err)
	resetQueueFlags()
	_, err = runRoot(t, "queue", "complete", "70-01", "--dir", dir)
	liberado(t, err)
}

func TestQueue_ForgeJSONInvalidoNaoDestravaAGuarda(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), []byte("{x"), 0o644))
	resetQueueFlags()
	_, err := runRoot(t, "queue", "next", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modo pipeline")
	assert.Contains(t, err.Error(), "inválido", "diz que o forge.json está quebrado")
}

func TestQueue_SemForgeJSONNadaMuda(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := queueProject(t, false)
	resetQueueFlags()
	_, err := runRoot(t, "queue", "next", "--dir", dir)
	liberado(t, err)
}
