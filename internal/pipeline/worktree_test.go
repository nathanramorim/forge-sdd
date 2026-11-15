package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gitProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	testutil.WriteFile(t, dir, "a.txt", "base\n")
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", "base")
	return dir
}

func TestWorktree_CriaIsoladoERemove(t *testing.T) {
	dir := gitProject(t)
	path, err := CreateWorktree(dir, "job-1", "feat-70-04", "main")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(layout.Worktrees(dir), "job-1", "feat-70-04"), path)
	assert.FileExists(t, filepath.Join(path, "a.txt"))

	// Mudança no worktree não afeta o checkout principal.
	testutil.WriteFile(t, path, "novo.txt", "x")
	assert.NoFileExists(t, filepath.Join(dir, "novo.txt"))
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/job-1-feat-70-04"), "forge/job-1-feat-70-04")

	require.NoError(t, RemoveWorktree(dir, "job-1", "feat-70-04"))
	assert.NoDirExists(t, path)
	assert.NotContains(t, testutil.Git(t, dir, "branch", "--list", "forge/*"), "forge/job-1")
	// Remover de novo é idempotente.
	require.NoError(t, RemoveWorktree(dir, "job-1", "feat-70-04"))
}

func TestWorktree_RemocaoMesmoComAlteracoesNaoCommitadas(t *testing.T) {
	dir := gitProject(t)
	path, err := CreateWorktree(dir, "job-2", "feat-70-05", "main")
	require.NoError(t, err)
	testutil.WriteFile(t, path, "sujo.txt", "alteração não commitada após uma falha")
	require.NoError(t, RemoveWorktree(dir, "job-2", "feat-70-05"))
	assert.NoDirExists(t, path)
}

func TestWorktree_Violacoes(t *testing.T) {
	dir := gitProject(t)
	for _, c := range [][3]string{
		{"../fora", "t", "main"}, {"j", "../t", "main"}, {"j/x", "t", "main"}, {"", "t", "main"},
		{"j", "t", "--upload-pack=x"}, {"j", "t", "a..b"}, {"j", "t", "-b"}, {"J", "t", "main"},
	} {
		_, err := CreateWorktree(dir, c[0], c[1], c[2])
		assert.Error(t, err, "%v", c)
	}
	assert.NoDirExists(t, layout.Worktrees(dir), "nada foi criado")
	assert.Error(t, RemoveWorktree(dir, "../x", "t"))

	_, err := CreateWorktree(dir, "j", "t", "ref-que-nao-existe")
	assert.Error(t, err)
}

func TestWorktree_RemocaoRecusaCaminhoForaPorLinkSimbolico(t *testing.T) {
	dir := gitProject(t)
	outside := t.TempDir()
	testutil.WriteFile(t, outside, "importante.txt", "não apague")
	require.NoError(t, os.MkdirAll(filepath.Join(layout.Worktrees(dir), "job-3"), 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(layout.Worktrees(dir), "job-3", "feat-70-04")))

	err := RemoveWorktree(dir, "job-3", "feat-70-04")
	require.Error(t, err)
	assert.FileExists(t, filepath.Join(outside, "importante.txt"))
}
