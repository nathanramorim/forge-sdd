package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetCIFlags() {
	_ = updateCmd.Flags().Set("ci", "false")
	_ = updateCmd.Flags().Set("yes", "false")
	_ = updateCmd.Flags().Set("upgrade", "false")
	_ = updateCmd.Flags().Set("version", "")
	_ = initCmd.Flags().Set("ci", "false")
	_ = initCmd.Flags().Set("yes", "false")
	_ = initCmd.Flags().Set("dry-run", "false")
	_ = initCmd.Flags().Set("agent", "")
}

func projeto50(t *testing.T) string {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"demo","version":"5.0.0","agents":["copilot","claude"],"lang":"pt-BR","naming_convention":"sequencial"}`)
	testutil.WriteFile(t, dir, "sdd/.sdd-version", "5.0.0")
	testutil.WriteFile(t, dir, ".github/workflows/ci.yml", "name: meu\non: push\n")
	return dir
}

func TestUpdateFlagCIEOptIn(t *testing.T) {
	t.Cleanup(resetCIFlags)
	dir := projeto50(t)
	wf := filepath.Join(dir, ".github/workflows/forge-pronto.yml")
	ci := filepath.Join(dir, ".github/workflows/ci.yml")

	_, err := runRoot(t, "update", dir, "--yes", "--version", "5.0.1")
	require.NoError(t, err)
	assert.NoFileExists(t, wf, "sem --ci nada é criado, nem com o Copilot")
	resetCIFlags()

	_, err = runRoot(t, "update", dir, "--yes", "--ci", "--dry-run")
	require.NoError(t, err)
	assert.NoFileExists(t, wf, "--dry-run só lista")
	resetCIFlags()
	_ = updateCmd.Flags().Set("dry-run", "false")

	_, err = runRoot(t, "update", dir, "--yes", "--ci")
	require.NoError(t, err)
	assert.FileExists(t, wf)
	first, _ := os.ReadFile(wf)
	resetCIFlags()

	_, err = runRoot(t, "update", dir, "--yes", "--ci")
	require.NoError(t, err)
	second, _ := os.ReadFile(wf)
	assert.Equal(t, string(first), string(second))
	meu, _ := os.ReadFile(ci)
	assert.Equal(t, "name: meu\non: push\n", string(meu))
}

func TestInitFlagCIEOptIn(t *testing.T) {
	t.Cleanup(resetCIFlags)
	sem := filepath.Join(t.TempDir(), "sem")
	_, err := runRoot(t, "init", sem, "--yes", "--agent", "copilot,claude")
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(sem, ".github/workflows/forge-pronto.yml"))
	resetCIFlags()

	com := filepath.Join(t.TempDir(), "com")
	_, err = runRoot(t, "init", com, "--yes", "--agent", "copilot,claude", "--ci")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(com, ".github/workflows/forge-pronto.yml"))
}

func TestDoctorNaoPrometeProtecaoDoWorkflow(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, ".github/workflows/forge-pronto.yml")
	require.NoError(t, os.MkdirAll(filepath.Dir(wf), 0o755))

	require.NoError(t, os.WriteFile(wf, []byte("name: meu\non: push\n"), 0o644))
	out := captureStdout(t, func() { reportPRGuardCoverage(dir, []string{"gemini"}) })
	assert.NotContains(t, out, "forge-pronto.yml", "workflow sem marcador não é do forge")
	assert.NotContains(t, out, "vale o job")

	require.NoError(t, os.WriteFile(wf, []byte("# forge-sdd:generated\nname: x\n"), 0o644))
	out = captureStdout(t, func() { reportPRGuardCoverage(dir, []string{"gemini", "copilot"}) })
	assert.Contains(t, out, "só registra a transição")
	assert.Contains(t, out, "não bloqueia")
	assert.NotContains(t, out, "vale o job")
	assert.NotContains(t, out, "segura")
}
