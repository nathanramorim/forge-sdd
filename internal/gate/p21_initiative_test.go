package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitInitiative cria um repositório git na branch indicada e a pasta da iniciativa correspondente.
func gitInitiative(t *testing.T, dir, branch, initiative string) string {
	t.Helper()
	testutil.InitRepo(t, dir, branch)
	base := filepath.Join(dir, "sdd", "discovery", initiative)
	require.NoError(t, os.MkdirAll(base, 0o755))
	return filepath.Join(base, "status.md")
}

func p21(t *testing.T, dir string) CheckResult {
	t.Helper()
	results, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	r, ok := findCheck(results, "P-21")
	require.True(t, ok)
	return r
}

func TestCheckP21_StatusDaIniciativaNoLimitePassa(t *testing.T) {
	dir := setupGateDir(t)
	status := gitInitiative(t, dir, "feat/69-teste-p21", "69-teste-p21")
	require.NoError(t, os.WriteFile(status, []byte(strings.Repeat("a", MaxProgressBytes)), 0o644))

	r := p21(t, dir)
	assert.True(t, r.Passed, r.Evidence)
	assert.Contains(t, r.Evidence, "status.md da iniciativa")
}

func TestCheckP21_StatusAcimaDoLimiteFalha(t *testing.T) {
	dir := setupGateDir(t)
	status := gitInitiative(t, dir, "feat/69-teste-p21", "69-teste-p21")
	require.NoError(t, os.WriteFile(status, []byte(strings.Repeat("a", MaxProgressBytes+1)), 0o644))

	r := p21(t, dir)
	assert.False(t, r.Passed)
	assert.Contains(t, r.Evidence, "status.md da iniciativa")
}

func TestCheckP21_IniciativaSemStatusFalha(t *testing.T) {
	dir := setupGateDir(t)
	gitInitiative(t, dir, "feat/69-teste-p21", "69-teste-p21")

	r := p21(t, dir)
	assert.False(t, r.Passed)
	assert.Contains(t, r.Evidence, "não encontrado")
}

func TestCheckP21_StatusDaIniciativaTemPrecedenciaSobreProgressLegado(t *testing.T) {
	dir := setupGateDir(t)
	status := gitInitiative(t, dir, "feat/69-teste-p21", "69-teste-p21")
	require.NoError(t, os.WriteFile(status, []byte("# ok"), 0o644))
	// progress.md legado estourado não deve reprovar quando há iniciativa ativa.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "progress.md"), []byte(strings.Repeat("a", MaxProgressBytes+50)), 0o644))

	r := p21(t, dir)
	assert.True(t, r.Passed, r.Evidence)
}

func TestCheckP21_SemIniciativaESemProgressPassaComObservacao(t *testing.T) {
	dir := setupGateDir(t)
	require.NoError(t, os.Remove(filepath.Join(dir, "sdd", "memory", "progress.md")))

	r := p21(t, dir)
	assert.True(t, r.Passed)
	assert.Contains(t, r.Evidence, "observação")
}
