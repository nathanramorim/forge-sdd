package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lequeEfemero(t *testing.T, tasks ...string) (string, *Leque) {
	t.Helper()
	dir := lequeProj(t)
	p := paramsLeque(tasks...)
	p.Efemero = true
	lq, err := PrepararLeque(dir, p)
	require.NoError(t, err)
	t.Cleanup(func() { _ = lq.Encerrar() })
	return dir, lq
}

func branchTask(task string) string { return "forge/" + iniLeque + "-" + task }

func TestEfemero_NascePorTaskEApagaNoCommit(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01", "feat-104-02")
	// Preparar não cria worktree nenhum.
	_, ok := lq.Caminho("feat-104-01")
	assert.False(t, ok)
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "worktrees", iniLeque, "feat-104-01"))

	p, err := lq.IniciarTask("feat-104-01")
	require.NoError(t, err)
	assert.DirExists(t, p)
	commitNo(t, p, "um.txt", "1")
	require.NoError(t, lq.EncerrarTask("feat-104-01"))

	assert.NoDirExists(t, p)
	// A branch com o commit continua; a outra task não ganhou worktree.
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", branchTask("feat-104-01")), "feat-104-01")
	assert.Equal(t, "1\n", testutil.Git(t, dir, "rev-list", "--count", "feat/ini.."+branchTask("feat-104-01")))
	_, ok = lq.Caminho("feat-104-02")
	assert.False(t, ok)
}

func TestEfemero_EncerrarRecusaTrabalhoNaoCommitado(t *testing.T) {
	_, lq := lequeEfemero(t, "feat-104-01")
	p, err := lq.IniciarTask("feat-104-01")
	require.NoError(t, err)
	testutil.WriteFile(t, p, "solto.txt", "x")
	require.Error(t, lq.EncerrarTask("feat-104-01"))
	assert.FileExists(t, filepath.Join(p, "solto.txt"), "nada se perde")
}

func TestEfemero_RetrabalhoRecriaDaBranchDaTask(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01")
	p, err := lq.IniciarTask("feat-104-01")
	require.NoError(t, err)
	commitNo(t, p, "um.txt", "1")
	require.NoError(t, lq.EncerrarTask("feat-104-01"))

	p2, err := lq.ReabrirTask("feat-104-01")
	require.NoError(t, err)
	b, err := os.ReadFile(filepath.Join(p2, "um.txt"))
	require.NoError(t, err)
	assert.Equal(t, "1", string(b), "o commit anterior é o ponto de partida")
	commitNo(t, p2, "um.txt", "2")
	require.NoError(t, lq.EncerrarTask("feat-104-01"))
	assert.Equal(t, "2\n", testutil.Git(t, dir, "rev-list", "--count", "feat/ini.."+branchTask("feat-104-01")))
}

func TestEfemero_Negativos(t *testing.T) {
	_, lq := lequeEfemero(t, "feat-104-01")
	_, err := lq.ReabrirTask("feat-104-01") // sem branch ainda
	require.Error(t, err)
	_, err = lq.IniciarTask("feat-104-99") // fora do leque
	require.Error(t, err)
	_, err = lq.IniciarTask("../fora")
	require.Error(t, err)
	require.Error(t, lq.EncerrarTask("feat-104-01")) // sem worktree aberto
	_, err = lq.IniciarTask("feat-104-01")
	require.NoError(t, err)
	_, err = lq.IniciarTask("feat-104-01") // duplicado
	require.Error(t, err)
}

func TestEfemero_OrfaoComTrabalhoViraMenu(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01")
	p, err := lq.IniciarTask("feat-104-01")
	require.NoError(t, err)
	commitNo(t, p, "um.txt", "1")
	require.NoError(t, lq.EncerrarTask("feat-104-01"))
	// Um leque novo (execução interrompida) encontra a branch com trabalho: menu, nada removido.
	require.NoError(t, lq.Encerrar())
	p2 := paramsLeque("feat-104-01")
	p2.Efemero = true
	lq2, err := PrepararLeque(dir, p2)
	require.NoError(t, err)
	defer func() { _ = lq2.Encerrar() }()
	_, err = lq2.IniciarTask("feat-104-01")
	var orf *ErroOrfao
	require.ErrorAs(t, err, &orf)
	assert.Len(t, orf.Opcoes, 3)
}

func TestWorktree_ReabrirEMantemBranch(t *testing.T) {
	dir := lequeProj(t)
	p, err := CreateWorktree(dir, "ini-1", "t1", "feat/ini")
	require.NoError(t, err)
	commitNo(t, p, "x.txt", "x")
	require.NoError(t, RemoveWorktreeKeepBranch(dir, "ini-1", "t1"))
	assert.NoDirExists(t, p)
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/ini-1-t1"), "t1")
	p2, err := ReopenWorktree(dir, "ini-1", "t1")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(p2, "x.txt"))
	_, err = ReopenWorktree(dir, "ini-1", "t1") // já ocupado
	require.Error(t, err)
	_, err = ReopenWorktree(dir, "ini-1", "../x")
	require.Error(t, err)
	_, err = ReopenWorktree(dir, "ini-1", "naoexiste")
	require.Error(t, err)
}
