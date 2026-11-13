package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetBusFlags() {
	_ = busListCmd.Flags().Set("state", layout.StatePending)
	_ = busListCmd.Flags().Set("json", "false")
	_ = busCompleteCmd.Flags().Set("note", "")
	_ = busFailCmd.Flags().Set("reason", "")
	_ = busClaimCmd.Flags().Set("ttl", bus.DefaultLeaseTTL.String())
}

func TestBusCLI_CicloCompleto(t *testing.T) {
	dir := t.TempDir()
	resetBusFlags()
	job := bus.Job{
		ID: "job-cli-1", Type: bus.TypeFeatureBuild, Priority: bus.PriorityNormal, Feature: "auth-oauth", Title: "Auth",
		SpecPath: "sdd/discovery/70-x/criteria.md", Card: bus.Card{Provider: "clickup", ID: "CU-1", URL: "https://x/1"},
		PR: bus.PR{Branch: "feat/auth", BaseBranch: "main"}, MaxReworkAttempts: 3,
	}
	data, err := json.Marshal(job)
	require.NoError(t, err)
	jobFile := filepath.Join(t.TempDir(), "job.json")
	require.NoError(t, os.WriteFile(jobFile, data, 0o644))

	out, err := runRoot(t, "bus", "enqueue", "--dir", dir, "--role", "builder", "--job-file", jobFile)
	require.NoError(t, err)
	assert.Contains(t, out, "enfileirado")

	out, err = runRoot(t, "bus", "enqueue", "--dir", dir, "--role", "builder", "--job-file", jobFile)
	require.NoError(t, err)
	assert.Contains(t, out, "idempotente")

	out, err = runRoot(t, "bus", "list", "--dir", dir, "--role", "builder")
	require.NoError(t, err)
	assert.Contains(t, out, "job-cli-1")

	out, err = runRoot(t, "bus", "claim", "--dir", dir, "--role", "builder", "--worker", "w1")
	require.NoError(t, err)
	var claimed bus.Job
	require.NoError(t, json.Unmarshal([]byte(out), &claimed))
	assert.Equal(t, "job-cli-1", claimed.ID)
	require.NotNil(t, claimed.Lease)

	out, err = runRoot(t, "bus", "claim", "--dir", dir, "--role", "builder", "--worker", "w1")
	require.NoError(t, err, "fila vazia não é erro")
	assert.Contains(t, out, "fila vazia")

	_, err = runRoot(t, "bus", "complete", "--dir", dir, "--role", "builder", "--id", "job-cli-1", "--worker", "outro")
	require.Error(t, err)

	out, err = runRoot(t, "bus", "complete", "--dir", dir, "--role", "builder", "--id", "job-cli-1", "--worker", "w1", "--note", "ok")
	require.NoError(t, err)
	assert.Contains(t, out, "concluído")

	resetBusFlags()
	out, err = runRoot(t, "bus", "list", "--dir", dir, "--role", "builder", "--state", "done", "--json")
	require.NoError(t, err)
	var done []bus.Job
	require.NoError(t, json.Unmarshal([]byte(out), &done))
	require.Len(t, done, 1)

	out, err = runRoot(t, "bus", "recover", "--dir", dir, "--role", "builder")
	require.NoError(t, err)
	assert.Contains(t, out, "recuperados: 0")
}

func TestBusCLI_Violacoes(t *testing.T) {
	dir := t.TempDir()
	resetBusFlags()
	_, err := runRoot(t, "bus", "list", "--dir", dir, "--role", "../x")
	require.Error(t, err)
	_, err = runRoot(t, "bus", "list", "--dir", dir, "--role", "discovery")
	require.Error(t, err)
	_, err = runRoot(t, "bus", "list", "--dir", dir, "--role", "builder", "--state", "inbox")
	require.Error(t, err)
	_, err = runRoot(t, "bus", "enqueue", "--dir", dir, "--role", "builder", "--job-file", filepath.Join(dir, "nao-existe.json"))
	require.Error(t, err)

	bad := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte(`{"id":"../x","type":"feature_build"}`), 0o644))
	_, err = runRoot(t, "bus", "enqueue", "--dir", dir, "--role", "builder", "--job-file", bad)
	require.Error(t, err)
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "queues", "builder", "pending"))
}
