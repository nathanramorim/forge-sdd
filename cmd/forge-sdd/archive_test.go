package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiveCmd_Execucao(t *testing.T) {
	dir := t.TempDir()
	initDir := filepath.Join(dir, "sdd", "discovery", "79-clareza")
	require.NoError(t, os.MkdirAll(initDir, 0o755))
	testutil.WriteFile(t, initDir, "criteria.md", "# Criteria\n")

	q := queue.Queue{
		FeatureID:  "79",
		SpecStatus: "reviewed",
		Items: []queue.QueueItem{
			{TaskID: "79-01", Status: queue.StatusApproved},
		},
	}
	qData, err := json.Marshal(q)
	require.NoError(t, err)
	testutil.WriteFile(t, initDir, "queue.json", string(qData))

	out, err := runRoot(t, "archive", "--dir", dir, "--initiative", "79-clareza")
	require.NoError(t, err)
	assert.Contains(t, out, "arquivada com sucesso")

	// Verifica se a pasta em sdd/archive/discovery/79-clareza foi criada
	assert.FileExists(t, filepath.Join(dir, "sdd", "archive", "discovery", "79-clareza", "criteria.md"))
}
