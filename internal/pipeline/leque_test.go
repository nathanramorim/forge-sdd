package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	defer func() { _ = lq.Encerrar() }()
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
	defer func() { _ = lq.Encerrar() }()
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
	defer func() { _ = lq.Encerrar() }()
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
	var eo *ErroOrfao
	require.ErrorAs(t, err, &eo)
	assert.Len(t, eo.Opcoes, 3)
	// O trabalho segue intacto e o lock não ficou preso.
	assert.FileExists(t, filepath.Join(p2, "trabalho.txt"))
	// O órfão vazio da primeira task foi recriado e depois limpo pela falha? Não: só ele é descartável.
	lq3, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	defer func() { _ = lq3.Encerrar() }()
	assert.DirExists(t, p1)
}

func orfaoPreparado(t *testing.T) (dir, wt string) {
	t.Helper()
	dir = lequeProj(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	wt, _ = lq.Caminho("feat-104-01")
	require.NoError(t, lq.Encerrar())
	return dir, wt
}

func TestOrfaoPreservaTrabalho(t *testing.T) {
	br := "forge/" + iniLeque + "-feat-104-01"
	casos := map[string]func(t *testing.T, dir, wt string){
		"head destacado com commit e branch apagada": func(t *testing.T, dir, wt string) {
			testutil.Git(t, wt, "checkout", "-q", "--detach")
			commitNo(t, wt, "d.txt", "d")
			testutil.Git(t, dir, "branch", "-D", br)
		},
		"branch apagada com commit na branch": func(t *testing.T, dir, wt string) {
			commitNo(t, wt, "d.txt", "d")
			testutil.Git(t, wt, "checkout", "-q", "--detach")
			testutil.Git(t, dir, "branch", "-D", br)
		},
		"arquivo não rastreado": func(t *testing.T, dir, wt string) { testutil.WriteFile(t, wt, "solto.txt", "x") },
		"head destacado vazio":  func(t *testing.T, dir, wt string) { testutil.Git(t, wt, "checkout", "-q", "--detach") },
	}
	for nome, prep := range casos {
		t.Run(nome, func(t *testing.T) {
			dir, wt := orfaoPreparado(t)
			prep(t, dir, wt)
			head := testutil.Git(t, wt, "rev-parse", "HEAD")
			_, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
			var eo *ErroOrfao
			require.ErrorAs(t, err, &eo)
			assert.DirExists(t, wt)
			assert.Equal(t, head, testutil.Git(t, wt, "rev-parse", "HEAD"))
		})
	}
}

func TestOrfaoVazioERecriado(t *testing.T) {
	dir, wt := orfaoPreparado(t)
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	defer func() { _ = lq.Encerrar() }()
	assert.DirExists(t, wt)
}

func TestWorktreeSymlinkParaForaRecusado(t *testing.T) {
	dir := lequeProj(t)
	fora := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "runtime", "worktrees"), 0o755))
	require.NoError(t, os.Symlink(fora, filepath.Join(dir, ".forge", "runtime", "worktrees", iniLeque)))
	_, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.Error(t, err)
	ents, _ := os.ReadDir(fora)
	assert.Empty(t, ents)
	// Lock liberado após o erro.
	_, err = harness.AdquirirLock(dirLocks(dir), chaveLockPrincipal, "outro", leaseLeque, time.Now())
	assert.NoError(t, err)

	// Caminho da task em si como link.
	dir2 := lequeProj(t)
	require.NoError(t, os.MkdirAll(filepath.Join(dir2, ".forge", "runtime", "worktrees", iniLeque), 0o755))
	require.NoError(t, os.Symlink(fora, filepath.Join(dir2, ".forge", "runtime", "worktrees", iniLeque, "feat-104-01")))
	_, err = PrepararLeque(dir2, paramsLeque("feat-104-01"))
	require.Error(t, err)
	ents, _ = os.ReadDir(fora)
	assert.Empty(t, ents)
}

func TestJuncaoErroFatalDevolveParciais(t *testing.T) {
	dir := lequeProj(t)
	// Hook (do próprio repositório) suja o principal depois do 1o merge: a 2a task falha de forma fatal.
	hook := filepath.Join(dir, ".git", "hooks", "post-merge")
	require.NoError(t, os.WriteFile(hook, []byte("#!/bin/sh\necho sujo >> a.txt\n"), 0o755))
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01", "feat-104-02"))
	require.NoError(t, err)
	p1, _ := lq.Caminho("feat-104-01")
	p2, _ := lq.Caminho("feat-104-02")
	commitNo(t, p1, "um.txt", "1")
	commitNo(t, p2, "dois.txt", "2")
	ret, err := lq.Juntar()
	require.Error(t, err)
	require.NotNil(t, ret)
	require.Len(t, ret.Resultados, 1)
	assert.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado)
	assert.DirExists(t, p2)
	require.NoError(t, lq.Encerrar())
	require.NoError(t, lq.Encerrar()) // idempotente
}

func semIdentidade(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	cfg := filepath.Join(home, "gitconfig")
	require.NoError(t, os.WriteFile(cfg, []byte("[user]\n\tuseConfigOnly = true\n"), 0o644))
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL", "EMAIL"} {
		t.Setenv(k, "x")
		require.NoError(t, os.Unsetenv(k))
	}
}

func juntarUma(t *testing.T, dir string) *RetornoJuncao {
	t.Helper()
	lq, err := PrepararLeque(dir, paramsLeque("feat-104-01"))
	require.NoError(t, err)
	defer func() { _ = lq.Encerrar() }()
	p1, _ := lq.Caminho("feat-104-01")
	commitNo(t, p1, "um.txt", "1")
	ret, err := lq.Juntar()
	require.NoError(t, err)
	return ret
}

func TestJuncaoSemIdentidadeUsaFallback(t *testing.T) {
	dir := lequeProj(t)
	semIdentidade(t)
	ret := juntarUma(t, dir)
	require.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado, ret.Resultados[0].Motivo)
	assert.Equal(t, "forge-sdd", strings.TrimSpace(testutil.Git(t, dir, "log", "-1", "--format=%cn")))
	cfg, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	require.NoError(t, err)
	assert.NotContains(t, string(cfg), "forge-sdd")
}

func TestJuncaoNaoSobrescreveIdentidadeDoRepositorio(t *testing.T) {
	dir := lequeProj(t)
	semIdentidade(t)
	testutil.Git(t, dir, "config", "user.name", "Pessoa Real")
	testutil.Git(t, dir, "config", "user.email", "real@example.com")
	ret := juntarUma(t, dir)
	require.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado, ret.Resultados[0].Motivo)
	assert.Equal(t, "Pessoa Real", strings.TrimSpace(testutil.Git(t, dir, "log", "-1", "--format=%cn")))
	assert.Equal(t, "real@example.com", strings.TrimSpace(testutil.Git(t, dir, "log", "-1", "--format=%ce")))
}

func TestJuncaoIdentidadeGlobalNaoSobrescrita(t *testing.T) {
	dir := lequeProj(t)
	semIdentidade(t)
	cfg := os.Getenv("GIT_CONFIG_GLOBAL")
	require.NoError(t, os.WriteFile(cfg, []byte("[user]\n\tname = Global Pessoa\n\temail = g@example.com\n"), 0o644))
	ret := juntarUma(t, dir)
	require.Equal(t, JuncaoIntegrada, ret.Resultados[0].Estado, ret.Resultados[0].Motivo)
	assert.Equal(t, "Global Pessoa", strings.TrimSpace(testutil.Git(t, dir, "log", "-1", "--format=%cn")))
}

func TestJuncaoFalhaSemConflitoNaoViraConflito(t *testing.T) {
	dir := lequeProj(t)
	hook := filepath.Join(dir, ".git", "hooks", "pre-merge-commit")
	require.NoError(t, os.WriteFile(hook, []byte("#!/bin/sh\necho recusado >&2\nexit 1\n"), 0o755))
	ret := juntarUma(t, dir)
	r := ret.Resultados[0]
	assert.Equal(t, JuncaoBloqueada, r.Estado)
	assert.Empty(t, r.Arquivos)
	assert.Contains(t, r.Motivo, "não é conflito")
	assert.NotContains(t, r.Motivo, "conflito de junção")
	assert.NotEmpty(t, ret.Opcoes)
	assert.Empty(t, testutil.Git(t, dir, "status", "--porcelain", "--untracked-files=no"))
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/*"), "feat-104-01")
}
