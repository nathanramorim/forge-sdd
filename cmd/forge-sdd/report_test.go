package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportCmd_NoMetrics(t *testing.T) {
	dir := t.TempDir()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"report", "--dir", dir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	assert.Contains(t, buf.String(), "Nenhuma métrica encontrada")
}

func TestReportCmd_AggregatesByTypeAndFeature(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	_, err := WriteSessionMetrics(dir, sessionMetrics{
		Feature: "sdd/discovery/discovery-55-telemetria.md", Outcome: "approved",
		Model: "claude-sonnet-5", TokensInput: 1000, TokensOutput: 500, DurationSeconds: 600,
	}, base)
	require.NoError(t, err)

	_, err = WriteSessionMetrics(dir, sessionMetrics{
		Feature: "sdd/features/feat-55-x/feat-55-01-a.md", Outcome: "approved",
		Model: "claude-sonnet-5", TokensInput: 2000, TokensOutput: 1000, DurationSeconds: 300,
	}, base.Add(time.Hour))
	require.NoError(t, err)

	_, err = WriteSessionMetrics(dir, sessionMetrics{
		Feature: "sdd/features/feat-55-x/feat-55-01-a.md", Outcome: "approved",
		Model: "claude-opus-5", TokensInput: 500, TokensOutput: 500, DurationSeconds: 300,
	}, base.Add(2*time.Hour))
	require.NoError(t, err)

	_, err = WriteSessionMetrics(dir, sessionMetrics{
		Feature: "sdd/features/fix-99-algo.md", Outcome: "blocked",
		Model: "claude-sonnet-5", TokensInput: 100, TokensOutput: 100, DurationSeconds: 60,
	}, base.Add(48*time.Hour))
	require.NoError(t, err)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"report", "--dir", dir})
	err = rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "discovery")
	assert.Contains(t, output, "sdd/discovery/discovery-55-telemetria.md")
	assert.Contains(t, output, "sdd/features/feat-55-x/feat-55-01-a.md")
	assert.Contains(t, output, "claude-opus-5,claude-sonnet-5")
	assert.Contains(t, output, "sdd/features/fix-99-algo.md")
	assert.Contains(t, output, "4000")
	assert.Contains(t, output, "Idade medida do projeto")
	assert.Contains(t, output, "2d")
}
