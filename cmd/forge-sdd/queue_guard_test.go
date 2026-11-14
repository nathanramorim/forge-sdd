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

func TestQueue_GuardaModoPipeline(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := queueProject(t, true)

	resetQueueFlags()
	_, err := runRoot(t, "queue", "next", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modo pipeline")
	assert.Contains(t, err.Error(), "forge-sdd dispatch")

	resetQueueFlags()
	_, err = runRoot(t, "queue", "complete", "70-01", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modo pipeline")

	// --local libera (o erro agora é o da fila vazia, não o da guarda).
	resetQueueFlags()
	_, err = runRoot(t, "queue", "next", "--dir", dir, "--local")
	if err != nil {
		assert.NotContains(t, err.Error(), "modo pipeline")
	}

	// Sob um worker (FORGE_WORKER) também libera.
	t.Setenv("FORGE_WORKER", "forge-builder")
	resetQueueFlags()
	_, err = runRoot(t, "queue", "next", "--dir", dir)
	if err != nil {
		assert.NotContains(t, err.Error(), "modo pipeline")
	}
}

func TestQueue_SemForgeJSONNadaMuda(t *testing.T) {
	t.Setenv("FORGE_WORKER", "")
	dir := queueProject(t, false)
	resetQueueFlags()
	_, err := runRoot(t, "queue", "next", "--dir", dir)
	if err != nil {
		assert.NotContains(t, err.Error(), "modo pipeline", "sem forge.json o fluxo da 4.0 segue igual")
	}
}
