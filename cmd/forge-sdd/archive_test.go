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

	gitT(t, dir, "init", "-q", "-b", "feat/79-clareza")

	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "base")

	out, err := runRoot(t, "archive", "--dir", dir, "--initiative", "79-clareza")
	require.NoError(t, err)
	assert.Contains(t, out, "arquivada com sucesso")

	// Verifica se a pasta em sdd/archive/discovery/79-clareza foi criada
	assert.FileExists(t, filepath.Join(dir, "sdd", "archive", "discovery", "79-clareza", "criteria.md"))
}

func TestArchiveCmd_Concluidas(t *testing.T) {
	dir := t.TempDir()
	for id, status := range map[string]string{"79-pronta": "approved", "80-andando": "building"} {
		base := filepath.Join(dir, "sdd", "discovery", id)
		require.NoError(t, os.MkdirAll(base, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "queue.json"), []byte(`{"items":[{"task_id":"01","status":"`+status+`"}]}`), 0o644))
	}
	_ = archiveCmd.Flags().Set("concluidas", "false")
	_ = archiveCmd.Flags().Set("dry-run", "false")
	t.Cleanup(func() {
		_ = archiveCmd.Flags().Set("concluidas", "false")
		_ = archiveCmd.Flags().Set("dry-run", "false")
	})

	out, err := runRoot(t, "archive", "--dir", dir, "--concluidas", "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, out, "79-pronta")
	assert.NotContains(t, out, "80-andando")
	assert.DirExists(t, filepath.Join(dir, "sdd", "discovery", "79-pronta"), "dry-run não move")
}
