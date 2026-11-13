package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	// O rootCmd é global: restaura os padrões das flags entre execuções.
	_ = statusCmd.Flags().Set("all-branches", "false")
	_ = nextIDCmd.Flags().Set("kind", "discovery")
	_ = nextIDCmd.Flags().Set("workitem", "")
	_ = updateCmd.Flags().Set("dry-run", "false")
	_ = contextCmd.Flags().Set("station", "build")
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return out.String(), err
}

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	testutil.Git(t, dir, args...)
}

func TestStatusENextID_ComIniciativaEmOutraBranch(t *testing.T) {
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", "10-antiga"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "discovery", "10-antiga", "discovery.md"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(`{"naming_convention":"sequencial"}`), 0o644))
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "base")
	gitT(t, dir, "checkout", "-q", "-b", "feat/70-alfa")
	base := filepath.Join(dir, "sdd", "discovery", "70-alfa")
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "status.md"), []byte("## Status\n```\n70-01 [x] done\n```\n## Proximo passo\nSeguir 70-02.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "queue.json"), []byte(`{"items":[{"task_id":"70-01","status":"approved"},{"task_id":"70-02","status":"queued"}]}`), 0o644))
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "70")
	gitT(t, dir, "checkout", "-q", "main")

	out, err := runRoot(t, "next-id", "--dir", dir)
	require.NoError(t, err)
	assert.Equal(t, "71\n", out)

	out, err = runRoot(t, "status", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "10-antiga")
	assert.NotContains(t, out, "70-alfa")

	out, err = runRoot(t, "status", "--dir", dir, "--all-branches")
	require.NoError(t, err)
	assert.Contains(t, out, "70-alfa")
	assert.Contains(t, out, "feat/70-alfa")
	assert.Contains(t, out, "70-01 [x] done")
	assert.Contains(t, out, "Seguir 70-02.")
	assert.Contains(t, out, "approved=1 queued=1")
}

func TestStatus_SemIniciativas(t *testing.T) {
	out, err := runRoot(t, "status", "--dir", t.TempDir())
	require.NoError(t, err)
	assert.Contains(t, out, "Nenhuma iniciativa")
}

func TestNextID_KindInvalidoEWorkitemSemValor(t *testing.T) {
	_, err := runRoot(t, "next-id", "--dir", t.TempDir(), "--kind", "feature")
	require.Error(t, err)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(`{"naming_convention":"workitem"}`), 0o644))
	_, err = runRoot(t, "next-id", "--dir", dir)
	require.Error(t, err)

	out, err := runRoot(t, "next-id", "--dir", dir, "--workitem", "1024")
	require.NoError(t, err)
	assert.Equal(t, "1024\n", out)
}

func TestNextID_SddrcInvalidoFalhaEmVezDeAdivinhar(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte("{quebrado"), 0o644))
	_, err := runRoot(t, "next-id", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inválido")
}

func TestContext_PorTaskIDEPorCaminho(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "- **ARQ-01:** Fronteiras.\n- **TST-01:** Executável.\n")
	testutil.WriteFile(t, dir, ".agents/skills/sdd-build/SKILL.md", "---\nname: sdd-build\nstatus: approved\nscope: core\n---\n")
	testutil.WriteFile(t, dir, "sdd/discovery/70-x/features/feat-70-01.md", "# Spec\nCita ARQ-01.\n")
	testutil.WriteFile(t, dir, "sdd/.queue/queue.json", `{"items":[{"task_id":"70-01","spec_path":"sdd/discovery/70-x/features/feat-70-01.md","status":"queued","rules":["TST-01"]}]}`)

	out, err := runRoot(t, "context", "70-01", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "- **ARQ-01:** Fronteiras.")
	assert.Contains(t, out, "- **TST-01:** Executável.", "IDs da fila também entram")
	assert.Contains(t, out, ".agents/skills/sdd-build/SKILL.md")

	out, err = runRoot(t, "context", "sdd/discovery/70-x/features/feat-70-01.md", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "ARQ-01")

	_, err = runRoot(t, "context", "99-99", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "70-01", "o erro lista as tasks disponíveis")
}
