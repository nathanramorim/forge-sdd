package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func constPath(dir string) string { return filepath.Join(dir, "sdd", "memory", "constitution.md") }

func escreveConst(t *testing.T, texto string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Dir(constPath(dir)), 0o755))
	require.NoError(t, os.WriteFile(constPath(dir), []byte(texto), 0o644))
	return dir
}

func leConst(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(constPath(dir))
	require.NoError(t, err)
	return string(b)
}

const constAntiga = "# C\n\n## Ferramentas e Integrações\n| Campo | Valor |\n|---|---|\n| VCS / Work Item System | github |\n\nTexto do usuário.\n\n## Regras\n1. x\n"

func TestInitConstituicaoTemCamposDeModo(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{Project: "demo", Stack: "go", DB: "postgres", Lang: "pt-BR", SddVersion: "5.0.0"}, dir)
	require.NoError(t, err)
	c := leConst(t, dir)
	for _, l := range []string{
		"## Ferramentas e Integrações",
		"| Modo do atestado | local |",
		"| Modo de revisão humana | com-humano |",
		"| WIP humano | 3 |",
		"| Módulos permitidos | nenhum |",
		"| Início automático do Builder | desligado |",
	} {
		assert.Contains(t, c, l)
	}
}

func TestUpdateAcrescentaSoOsCamposQueFaltam(t *testing.T) {
	dir := escreveConst(t, strings.Replace(constAntiga, "|\n\nTexto", "|\n| WIP humano | 5 |\n\nTexto", 1))
	var w bytes.Buffer
	n, err := AcrescentaCamposDeModo(dir, &w)
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	c := leConst(t, dir)
	assert.Equal(t, 1, strings.Count(c, "| WIP humano |"))
	assert.Contains(t, c, "| WIP humano | 5 |")
	assert.Contains(t, c, "| Modo do atestado | local |")
	assert.Contains(t, c, "| Início automático do Builder | desligado |")
	assert.Contains(t, c, "\n\nTexto do usuário.\n\n## Regras\n1. x\n")
	assert.Empty(t, w.String())
}

func TestUpdateNaoAlteraValorExistenteDeModo(t *testing.T) {
	dir := escreveConst(t, strings.Replace(constAntiga, "|\n\nTexto", "|\n| Modo do atestado | typo-invalido |\n| Módulos permitidos | a, b |\n\nTexto", 1))
	_, err := AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	c := leConst(t, dir)
	assert.Contains(t, c, "| Modo do atestado | typo-invalido |")
	assert.Contains(t, c, "| Módulos permitidos | a, b |")
	assert.NotContains(t, c, "| Modo do atestado | local |")
}

func TestUpdateConstituicaoIlegivelNaoETocada(t *testing.T) {
	// Sem arquivo: nada a fazer e nada criado.
	dir := t.TempDir()
	n, err := AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	assert.Zero(t, n)
	_, err = os.Stat(constPath(dir))
	assert.True(t, os.IsNotExist(err))

	// Constituição ilegível (é um diretório): erro e nada é gravado.
	dir2 := t.TempDir()
	require.NoError(t, os.MkdirAll(constPath(dir2), 0o755))
	_, err = AcrescentaCamposDeModo(dir2, &bytes.Buffer{})
	assert.Error(t, err)
	fi, statErr := os.Stat(constPath(dir2))
	require.NoError(t, statErr)
	assert.True(t, fi.IsDir())
}

func TestAcrescimoDeModoEIdempotente(t *testing.T) {
	dir := escreveConst(t, constAntiga)
	n, err := AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	antes := leConst(t, dir)
	passado := time.Now().Add(-time.Hour).Truncate(time.Second)
	require.NoError(t, os.Chtimes(constPath(dir), passado, passado))

	n, err = AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Equal(t, antes, leConst(t, dir))
	fi, err := os.Stat(constPath(dir))
	require.NoError(t, err)
	assert.True(t, fi.ModTime().Equal(passado), "segunda execução não pode regravar")
}

func TestAcrescimoDeModoSecaoSemTabelaNaoGravaEAvisa(t *testing.T) {
	orig := "# C\n\n## Ferramentas e Integrações\nSó texto, sem tabela.\n\n## Regras\n1. x\n"
	dir := escreveConst(t, orig)
	var w bytes.Buffer
	n, err := AcrescentaCamposDeModo(dir, &w)
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Equal(t, orig, leConst(t, dir))
	assert.Contains(t, w.String(), "aviso")

	// Seção ausente: criada ao fim, com o texto anterior intacto.
	dir2 := escreveConst(t, "# C\n\nTexto.\n")
	n, err = AcrescentaCamposDeModo(dir2, &bytes.Buffer{})
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	c := leConst(t, dir2)
	assert.True(t, strings.HasPrefix(c, "# C\n\nTexto.\n"))
	assert.Contains(t, c, "## Ferramentas e Integrações\n| Campo | Valor |")
	assert.Contains(t, c, "| WIP humano | 3 |")
}

func TestShouldPreserveConstituicaoSoAceitaAcrescimo(t *testing.T) {
	dir := escreveConst(t, constAntiga)
	assert.True(t, shouldPreserve(constPath(dir), dir))

	passado := time.Now().Add(-time.Hour).Truncate(time.Second)
	require.NoError(t, os.Chtimes(constPath(dir), passado, passado))
	_, err := Run(config.Config{Project: "demo", Stack: "go", DB: "postgres", Lang: "pt-BR", SddVersion: "5.0.0"}, dir)
	require.NoError(t, err)
	assert.Equal(t, constAntiga, leConst(t, dir), "Run não toca a Constituição")
	fi, err := os.Stat(constPath(dir))
	require.NoError(t, err)
	assert.True(t, fi.ModTime().Equal(passado))

	_, err = AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	assert.NotEqual(t, constAntiga, leConst(t, dir), "só o passo de acréscimo grava")
}
