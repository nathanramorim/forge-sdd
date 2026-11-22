package vcs

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubCriarIssueComDescricaoComercial(t *testing.T) {
	var gotDir, gotName string
	var gotArgs []string
	g := &github{run: func(dir, name string, args ...string) (string, error) {
		gotDir, gotName, gotArgs = dir, name, args
		return "https://github.com/o/r/issues/42\n", nil
	}}
	is, err := g.CriarIssue("/proj", NovaIssue{Titulo: "[113] Feature; rm -rf /", Corpo: "Descrição $(x)", Etiquetas: []string{"spec"}})
	require.NoError(t, err)
	assert.Equal(t, Issue{Numero: 42, URL: "https://github.com/o/r/issues/42", Titulo: "[113] Feature; rm -rf /"}, is)
	assert.Equal(t, "/proj", gotDir)
	assert.Equal(t, "gh", gotName)
	assert.Equal(t, []string{"issue", "create", "--title", "[113] Feature; rm -rf /", "--body", "Descrição $(x)", "--label", "spec"}, gotArgs)
}

func TestGitHubProcurarIssueAbertaIdempotente(t *testing.T) {
	var gotArgs []string
	saida := `[{"number":7,"url":"https://github.com/o/r/issues/7","title":"Outra coisa"},` +
		`{"number":9,"url":"https://github.com/o/r/issues/9","title":"[113] Minha spec"}]`
	g := &github{run: func(_, _ string, args ...string) (string, error) { gotArgs = args; return saida, nil }}
	is, err := g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	require.NotNil(t, is)
	assert.Equal(t, 9, is.Numero)
	assert.Equal(t, []string{"issue", "list", "--state", "open", "--search", "Minha spec", "--limit", "200", "--json", "number,url,title"}, gotArgs)

	g.run = func(_, _ string, _ ...string) (string, error) {
		return `[{"number":3,"url":"u","title":"Minha spec"}]`, nil
	}
	is, err = g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	assert.Equal(t, 3, is.Numero)

	// título só parecido não conta; lista vazia e saída vazia também não
	g.run = func(_, _ string, _ ...string) (string, error) {
		return `[{"number":3,"url":"u","title":"Minha spec maior"}]`, nil
	}
	is, err = g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	assert.Nil(t, is)
	g.run = func(_, _ string, _ ...string) (string, error) { return "", nil }
	is, err = g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	assert.Nil(t, is)
}

func TestProcurarIssueAbertaUsaLimite(t *testing.T) {
	var gotArgs []string
	g := &github{run: func(_, _ string, args ...string) (string, error) { gotArgs = args; return "[]", nil }}
	_, err := g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	i := -1
	for k, a := range gotArgs {
		if a == "--limit" {
			i = k
		}
	}
	require.GreaterOrEqual(t, i, 0)
	require.Less(t, i+1, len(gotArgs))
	assert.Equal(t, "200", gotArgs[i+1])
}

func TestProcurarIssueAbertaAchaAlemDos30(t *testing.T) {
	var itens []string
	for n := 1; n <= 200; n++ {
		titulo := fmt.Sprintf("Outra %d", n)
		if n == 45 {
			titulo = "Minha spec"
		}
		itens = append(itens, fmt.Sprintf(`{"number":%d,"url":"https://github.com/o/r/issues/%d","title":%q}`, n, n, titulo))
	}
	saida := "[" + strings.Join(itens, ",") + "]"
	g := &github{run: func(_, _ string, _ ...string) (string, error) { return saida, nil }}
	is, err := g.ProcurarIssueAberta("/p", "Minha spec")
	require.NoError(t, err)
	require.NotNil(t, is)
	assert.Equal(t, 45, is.Numero)
}

func TestNenhumCriarIssueRetornaSemPR(t *testing.T) {
	n := &nenhum{run: func(_, _ string, _ ...string) (string, error) {
		t.Fatal("o runner não deve ser chamado")
		return "", nil
	}}
	_, err := n.CriarIssue("/p", NovaIssue{Titulo: "x"})
	assert.ErrorIs(t, err, ErrSemPR)
	is, err := n.ProcurarIssueAberta("/p", "x")
	assert.ErrorIs(t, err, ErrSemPR)
	assert.Nil(t, is)
}

func TestGitHubCriarIssueErroDoGh(t *testing.T) {
	boom := errors.New("gh falhou")
	g := &github{run: func(_, _ string, _ ...string) (string, error) { return "", boom }}
	_, err := g.CriarIssue("/p", NovaIssue{Titulo: "x"})
	assert.ErrorIs(t, err, boom)
	_, err = g.ProcurarIssueAberta("/p", "x")
	assert.ErrorIs(t, err, boom)
}

func TestGitHubCriarIssueRecusaTituloVazio(t *testing.T) {
	chamou := false
	g := &github{run: func(_, _ string, _ ...string) (string, error) { chamou = true; return "", nil }}
	_, err := g.CriarIssue("/p", NovaIssue{Titulo: "  "})
	assert.Error(t, err)
	_, err = g.ProcurarIssueAberta("/p", "")
	assert.Error(t, err)
	assert.False(t, chamou)
}

func TestGitHubProcurarIssueSaidaJSONInvalida(t *testing.T) {
	for _, saida := range []string{`nao e json`, `[{"number":1,"url":"u"}]`, `[{"number":1,"url":"u","title":"x","extra":1}]`, `{"number":1}`} {
		g := &github{run: func(_, _ string, _ ...string) (string, error) { return saida, nil }}
		is, err := g.ProcurarIssueAberta("/p", "x")
		assert.Error(t, err, saida)
		assert.Nil(t, is)
	}
	g := &github{run: func(_, _ string, _ ...string) (string, error) { return "sem url de issue", nil }}
	_, err := g.CriarIssue("/p", NovaIssue{Titulo: "x"})
	assert.Error(t, err)
}
