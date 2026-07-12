package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeMetric(t *testing.T, dir, name, outcome string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0755))
	content := `{"outcome":"` + outcome + `"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
}

func resetAutopilotFlags(t *testing.T) {
	t.Helper()
	_ = autopilotCmd.Flags().Set("skip-graduation", "false")
	_ = autopilotCmd.Flags().Set("min-cycles", "3")
}

func TestAutopilotCommand_BlockedWithoutEnoughCycles(t *testing.T) {
	tempDir := t.TempDir()
	writeMetric(t, filepath.Join(tempDir, "sdd", ".metrics"), "session-1.json", "approved")
	writeMetric(t, filepath.Join(tempDir, "sdd", ".metrics"), "session-2.json", "blocked")
	defer resetAutopilotFlags(t)

	rootCmd.SetArgs([]string{"autopilot", tempDir})
	err := rootCmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "graduação necessária")
	assert.Contains(t, err.Error(), "1/3")
	assert.NoFileExists(t, filepath.Join(tempDir, "sdd", ".sdd-auto-pilot"))
}

func TestAutopilotCommand_AllowedWithEnoughCycles(t *testing.T) {
	tempDir := t.TempDir()
	metricsDir := filepath.Join(tempDir, "sdd", ".metrics")
	writeMetric(t, metricsDir, "session-1.json", "approved")
	writeMetric(t, metricsDir, "session-2.json", "approved")
	writeMetric(t, metricsDir, "session-3.json", "approved")
	defer resetAutopilotFlags(t)

	rootCmd.SetArgs([]string{"autopilot", tempDir})
	err := rootCmd.Execute()

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "sdd", ".sdd-auto-pilot"))
}

func TestAutopilotCommand_SkipGraduationBypasses(t *testing.T) {
	tempDir := t.TempDir()
	defer resetAutopilotFlags(t)

	_ = autopilotCmd.Flags().Set("skip-graduation", "true")
	rootCmd.SetArgs([]string{"autopilot", tempDir})
	err := rootCmd.Execute()

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "sdd", ".sdd-auto-pilot"))
}

func TestAutopilotCommand_IdempotentWhenAlreadyEnabled(t *testing.T) {
	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sdd-auto-pilot"), []byte{}, 0644))
	defer resetAutopilotFlags(t)

	rootCmd.SetArgs([]string{"autopilot", tempDir})
	err := rootCmd.Execute()

	require.NoError(t, err)
}

func TestAutopilotCommand_CustomMinCycles(t *testing.T) {
	tempDir := t.TempDir()
	metricsDir := filepath.Join(tempDir, "sdd", ".metrics")
	writeMetric(t, metricsDir, "session-1.json", "approved")
	defer resetAutopilotFlags(t)

	_ = autopilotCmd.Flags().Set("min-cycles", "1")
	rootCmd.SetArgs([]string{"autopilot", tempDir})
	err := rootCmd.Execute()

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "sdd", ".sdd-auto-pilot"))
}
