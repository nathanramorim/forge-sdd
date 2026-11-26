package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// commitaTask abre o worktree efêmero da task, commita um arquivo e apaga o worktree.
func commitaTask(t *testing.T, lq *Leque, task, arq, conteudo string) {
	t.Helper()
	p, err := lq.IniciarTask(task)
	require.NoError(t, err)
	commitNo(t, p, arq, conteudo)
	require.NoError(t, lq.EncerrarTask(task))
}

func temBranch(t *testing.T, dir, task string) bool {
	t.Helper()
	return testutil.Git(t, dir, "branch", "--list", branchTask(task)) != ""
}

func TestJuncaoContinua_IntegraAoAprovarEApagaBranch(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01", "feat-104-02")
	commitaTask(t, lq, "feat-104-01", "um.txt", "1")
	assert.NoFileExists(t, filepath.Join(dir, "um.txt"), "só entra quando aprovada")

	ret, err := lq.JuntarAprovada("feat-104-01")
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 1)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado)
	assert.FileExists(t, filepath.Join(dir, "um.txt"))
	assert.False(t, temBranch(t, dir, "feat-104-01"), "integrada, a branch da task é apagada")

	// A task seguinte nasce da base já com a anterior integrada.
	p, err := lq.IniciarTask("feat-104-02")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(p, "um.txt"))
}

func TestJuncaoContinua_PosteriorEsperaAAnterior(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01", "feat-104-02", "feat-104-03")
	commitaTask(t, lq, "feat-104-01", "um.txt", "1")
	commitaTask(t, lq, "feat-104-02", "dois.txt", "2")
	commitaTask(t, lq, "feat-104-03", "tres.txt", "3")

	ret, err := lq.JuntarAprovada("feat-104-03") // aprovada antes das anteriores
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 1)
	assert.Equal(t, JuncaoAguardando, ret.Resultados[0].Estado)
	assert.NoFileExists(t, filepath.Join(dir, "tres.txt"))
	assert.True(t, temBranch(t, dir, "feat-104-03"))

	ret, err = lq.JuntarAprovada("feat-104-01")
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 1, "a 02 ainda não foi aprovada: a 03 segue esperando")
	assert.Equal(t, "feat-104-01", ret.Resultados[0].Task)
	assert.NoFileExists(t, filepath.Join(dir, "tres.txt"))

	ret, err = lq.JuntarAprovada("feat-104-02") // libera a 02 e, em seguida, a 03 já aprovada
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 2)
	assert.Equal(t, "feat-104-02", ret.Resultados[0].Task)
	assert.Equal(t, "feat-104-03", ret.Resultados[1].Task)
	for _, a := range []string{"um.txt", "dois.txt", "tres.txt"} {
		assert.FileExists(t, filepath.Join(dir, a))
	}
	// Ordem do plano no histórico de merges (do mais recente ao mais antigo).
	log := testutil.Git(t, dir, "log", "--merges", "--format=%s")
	assert.Regexp(t, `(?s)feat-104-03.*feat-104-02.*feat-104-01`, log)
}

func TestJuncaoContinua_ConflitoBloqueiaEPreservaBranch(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01", "feat-104-02", "feat-104-03")
	commitaTask(t, lq, "feat-104-01", "a.txt", "versão 1\n")
	commitaTask(t, lq, "feat-104-02", "a.txt", "versão 2\n") // nasceu da base antes da 01 entrar
	commitaTask(t, lq, "feat-104-03", "c.txt", "c")

	ret, err := lq.JuntarAprovada("feat-104-01")
	require.NoError(t, err)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado)

	ret, err = lq.JuntarAprovada("feat-104-02")
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 1)
	assert.Equal(t, JuncaoBloqueada, ret.Resultados[0].Estado)
	assert.Equal(t, []string{"a.txt"}, ret.Resultados[0].Arquivos)
	assert.True(t, ret.Bloqueadas())
	assert.Equal(t, "resolver", ret.Recomendada)
	assert.Len(t, ret.Opcoes, 3)
	assert.True(t, temBranch(t, dir, "feat-104-02"), "branch preservada")
	assert.Empty(t, testutil.Git(t, dir, "status", "--porcelain", "--untracked-files=no"), "merge abortado")
	b, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
	assert.Equal(t, "versão 1\n", string(b))

	// A 03 aprovada não passa por cima da 02 bloqueada.
	ret, err = lq.JuntarAprovada("feat-104-03")
	require.NoError(t, err)
	assert.True(t, ret.Bloqueadas())
	assert.NoFileExists(t, filepath.Join(dir, "c.txt"))
	assert.True(t, temBranch(t, dir, "feat-104-03"))
}

func TestJuncaoContinua_Negativos(t *testing.T) {
	_, lq := lequeEfemero(t, "feat-104-01")
	_, err := lq.JuntarAprovada("feat-104-99")
	require.Error(t, err)
	commitaTask(t, lq, "feat-104-01", "um.txt", "1")
	_, err = lq.JuntarAprovada("feat-104-01")
	require.NoError(t, err)
	_, err = lq.JuntarAprovada("feat-104-01") // já integrada
	require.Error(t, err)
}

func TestJuncaoContinua_RetrabalhoVoltaAEsperarAprovacao(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01")
	commitaTask(t, lq, "feat-104-01", "um.txt", "1")
	p, err := lq.ReabrirTask("feat-104-01")
	require.NoError(t, err)
	commitNo(t, p, "um.txt", "2")
	require.NoError(t, lq.EncerrarTask("feat-104-01"))
	ret, err := lq.JuntarAprovada("feat-104-01")
	require.NoError(t, err)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado)
	b, _ := os.ReadFile(filepath.Join(dir, "um.txt"))
	assert.Equal(t, "2", string(b))
}

func TestJuncaoContinua_JuntarDoFimContinuaFuncionando(t *testing.T) {
	dir, lq := lequeEfemero(t, "feat-104-01", "feat-104-02")
	commitaTask(t, lq, "feat-104-01", "um.txt", "1")
	commitaTask(t, lq, "feat-104-02", "dois.txt", "2")
	ret, err := lq.Juntar()
	require.NoError(t, err)
	require.Len(t, ret.Resultados, 2)
	assert.False(t, ret.Bloqueadas())
	assert.FileExists(t, filepath.Join(dir, "dois.txt"))
	assert.False(t, temBranch(t, dir, "feat-104-02"))
}
