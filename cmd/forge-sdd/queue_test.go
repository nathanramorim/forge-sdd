package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_QueueInteractive(t *testing.T) {
	dir := t.TempDir()

	// 1. Testa invocação em fila vazia
	out, err := runRoot(t, "queue", "--dir", dir, "--interactive")
	require.NoError(t, err)
	assert.Contains(t, out, "Nenhuma tarefa na fila ativa.")

	// 2. Testa subcomando queue panel --interactive
	outPanel, err := runRoot(t, "queue", "panel", "--dir", dir, "-i")
	require.NoError(t, err)
	assert.Contains(t, outPanel, "Nenhuma tarefa na fila ativa.")

	// 3. Testa fila com tarefas sem estourar panics
	qLegacyDir := filepath.Join(dir, "sdd", ".queue")
	require.NoError(t, os.MkdirAll(qLegacyDir, 0o755))
	qData, err := json.Marshal(&queue.Queue{
		FeatureID: "84-painel",
		Items: []queue.QueueItem{
			{TaskID: "84-01", Title: "Formatador Chat Feed", Status: queue.StatusApproved},
			{TaskID: "84-02", Title: "Integração Worker", Status: queue.StatusBuilding},
		},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(qLegacyDir, "queue.json"), qData, 0o644))

	// Carregamento da fila e verificação das estruturas
	loadedQ, err := queue.Load(dir)
	require.NoError(t, err)
	assert.Len(t, loadedQ.Items, 2)
}
