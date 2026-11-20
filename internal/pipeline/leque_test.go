package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const iniLeque = "104-pipeline-em-subagentes"

func lequeProj(t *testing.T) string {
	t.Helper()
	dir := gitProject(t)
	testutil.Git(t, dir, "checkout", "-q", "-b", "feat/ini")
	return dir
}

func paramsLeque(tasks ...string) ParamsLeque {
	return ParamsLeque{Iniciativa: iniLeque, Dono: "coord-a", Base: "feat/ini", Tasks: tasks}
}

func commitNo(t *testing.T, wt, arq, conteudo string) {
	t.Helper()
	testutil.WriteFile(t, wt, arq, conteudo)
	testutil.Git(t, wt, "add", "-A")
	testutil.Git(t, wt, "commit", "-q", "-m", "x")
}

func TestWorktreePorSubagente(t *testing.T) {
	dir := lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01", "feat-104-02"))
	require.NoError(t, err)
	defer lq.Encerrar()
	p1, _ := lq.Caminho("feat-104-01")
	p2, _ := lq.Caminho("feat-104-02")
	assert.NotEqual(t, p1, p2)
	commitNo(t, p1, "um.txt", "1")
	commitNo(t, p2, "dois.txt", "2")
	assert.NoFileExists(t, filepath.Join(dir, "um.txt"))

	ret, err := lq.Juntar()
	require.NoError(t, err)
	require.False(t, ret.Bloqueadas())
	assert.Equal(t, "feat-104-01", ret.Resultados[0].Task)
	assert.FileExists(t, filepath.Join(dir, "um.txt"))
	assert.FileExists(t, filepath.Join(dir, "dois.txt"))
	assert.Contains(t, testutil.Git(t, dir, "log", "--merges", "--format=%s"), "merge("+iniLeque+"): feat-104-02")
	assert.NoDirExists(t, p1)
}

func TestJuncaoConflitoEscala(t *testing.T) {
	dir := lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01", "feat-104-02", "feat-104-03"))
	require.NoError(t, err)
	defer lq.Encerrar()
	p1, _ := lq.Caminho("feat-104-01")
	p2, _ := lq.Caminho("feat-104-02")
	p3, _ := lq.Caminho("feat-104-03")
	commitNo(t, p1, "a.txt", "versão 1\n")
	commitNo(t, p2, "a.txt", "versão 2\n")
	commitNo(t, p3, "c.txt", "c")

	ret, err := lq.Juntar()
	require.NoError(t, err)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado)
	assert.Equal(t, JuncaoBloqueada, ret.Resultados[1].Estado)
	assert.Equal(t, []string{"a.txt"}, ret.Resultados[1].Arquivos)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[2].Estado)
	assert.Equal(t, "resolver", ret.Recomendada)
	assert.Len(t, ret.Opcoes, 3)
	// Sem merge silencioso nem estado de merge pendente; o trabalho da task bloqueada segue vivo.
	b, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
	assert.Equal(t, "versão 1\n", string(b))
	assert.Empty(t, testutil.Git(t, dir, "status", "--porcelain", "--untracked-files=no"))
	assert.DirExists(t, p2)
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/"+iniLeque+"-feat-104-02"), "feat-104-02")
}

func TestJuncaoTaskSujaBloqueia(t *testing.T) {
	dir := lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	defer lq.Encerrar()
	p1, _ := lq.Caminho("feat-104-01")
	testutil.WriteFile(t, p1, "solto.txt", "x")
	ret, err := lq.Juntar()
	require.NoError(t, err)
	assert.Equal(t, JuncaoBloqueada, ret.Resultados[0].Estado)
	assert.FileExists(t, filepath.Join(p1, "solto.txt"))
}

func TestRecusaWorktreeOcupado(t *testing.T) {
	dir := lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	// Outra sessão (dois leques concorrentes): recusada pelo lock.
	p := paramsLeque("feat-104-02")
	p.Dono = "coord-b"
	_, err = PrepararLeque(dir, p)
	require.ErrorIs(t, err, harness.ErrLockOcupado)
	require.NoError(t, lq.Encerrar())

	// Sujo.
	testutil.WriteFile(t, dir, "a.txt", "alterado")
	_, err = PrepararLeque(dir, p)
	require.ErrorContains(t, err, "não commitadas")
	testutil.Git(t, dir, "checkout", "-q", "--", "a.txt")

	// Branch errada.
	p.Base = "main"
	_, err = PrepararLeque(dir, p)
	require.ErrorContains(t, err, "não em")
	// O lock foi liberado nas recusas.
	p.Base = "feat/ini"
	lq2, err := PrepararLeque(dir, p)
	require.NoError(t, err)
	require.NoError(t, lq2.Encerrar())
}

func TestPrepararLequeEntradaHostil(t *testing.T) {
	dir := lequeProj(t)
	for _, p := range []ParamsLeque{
		{Iniciativa: "../x-y", Dono: "d", Base: "feat/ini", Tasks: []string{"t"}},
		{Iniciativa: iniLeque, Dono: "d", Base: "--upload-pack=x", Tasks: []string{"t"}},
		{Iniciativa: iniLeque, Dono: "d", Base: "a..b", Tasks: []string{"t"}},
		{Iniciativa: iniLeque, Dono: "d", Base: "feat/ini", Tasks: []string{"../t"}},
		{Iniciativa: iniLeque, Dono: "d", Base: "feat/ini", Tasks: []string{"t", "t"}},
		{Iniciativa: iniLeque, Dono: "d", Base: "feat/ini"},
		{Iniciativa: iniLeque, Dono: "", Base: "feat/ini", Tasks: []string{"t"}},
	} {
		_, err := PrepararLeque(dir, p)
		assert.Error(t, err, "%+v", p)
	}
	var lq *Leque
	_, err := lq.Juntar()
	assert.Error(t, err)
}

func TestLequeOrfao(t *testing.T) {
	dir := lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01", "feat-104-02"))
	require.NoError(t, err)
	p1, _ := lq.Caminho("feat-104-01")
	p2, _ := lq.Caminho("feat-104-02")
	commitNo(t, p2, "trabalho.txt", "importante")
	// Sai sem Juntar (processo morto): worktrees ficam órfãos.
	require.NoError(t, lq.Encerrar())

	_, err = PrepararLeque(dir, paramsLeque("feat-104-01", "feat-104-02"))
	require.ErrorContains(t, err, "guarda trabalho")
	// O trabalho segue intacto e o lock não ficou preso.
	assert.FileExists(t, filepath.Join(p2, "trabalho.txt"))
	// O órfão vazio da primeira task foi recriado e depois limpo pela falha? Não: só ele é descartável.
	lq3, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	defer lq3.Encerrar()
	assert.DirExists(t, p1)
}
