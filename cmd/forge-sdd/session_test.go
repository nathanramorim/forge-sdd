package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestSddrc(t *testing.T, dir string, telemetryEnabled bool) {
	t.Helper()
	sddDir := filepath.Join(dir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	rc := `{"project":"teste","version":"1.9.4","lang":"pt-BR","agents":["claude"],"telemetry":{"enabled":` +
		boolStr(telemetryEnabled) + `},"naming_convention":"sequencial"}`
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(rc), 0o644))
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestWriteSessionMetrics(t *testing.T) {
	dir := t.TempDir()
	at := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)

	metrics := sessionMetrics{
		Schema:     "forge-sdd/metrics/1.0",
		Feature:    "sdd/features/feat-01-simplificacao-e-aprendizado-continuo/feat-01-01-telemetria-code-enforced.md",
		Phase:      "01-01",
		AgentPath:  []string{"builder"},
		Outcome:    "approved",
		SddVersion: "1.9.4",
	}

	path, err := WriteSessionMetrics(dir, metrics, at)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "sdd", ".metrics", "session-2026-08-15T120000Z.json"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got sessionMetrics
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, metrics.Feature, got.Feature)
	assert.Equal(t, "approved", got.Outcome)
}

func TestSessionRecordCmd_SkipsWhenTelemetryDisabled(t *testing.T) {
	dir := t.TempDir()
	writeTestSddrc(t, dir, false)

	rootCmd.SetArgs([]string{
		"session", "record",
		"--dir", dir,
		"--feature", "sdd/features/feat-x.md",
		"--outcome", "approved",
	})
	require.NoError(t, rootCmd.Execute())

	entries, err := os.ReadDir(filepath.Join(dir, "sdd", ".metrics"))
	if err == nil {
		assert.Empty(t, entries)
	}
}

func TestSessionRecordCmd_WritesWhenTelemetryEnabled(t *testing.T) {
	dir := t.TempDir()
	writeTestSddrc(t, dir, true)

	rootCmd.SetArgs([]string{
		"session", "record",
		"--dir", dir,
		"--feature", "sdd/features/feat-x.md",
		"--outcome", "blocked",
		"--agent-path", "orquestrador,revisor",
	})
	require.NoError(t, rootCmd.Execute())

	entries, err := os.ReadDir(filepath.Join(dir, "sdd", ".metrics"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestSessionRecordCmd_RequiresValidOutcome(t *testing.T) {
	dir := t.TempDir()
	writeTestSddrc(t, dir, true)

	rootCmd.SetArgs([]string{
		"session", "record",
		"--dir", dir,
		"--feature", "sdd/features/feat-x.md",
		"--outcome", "invalido",
	})
	assert.Error(t, rootCmd.Execute())
}
