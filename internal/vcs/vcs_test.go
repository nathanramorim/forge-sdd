package vcs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dublê do executor: grava as chamadas e responde por prefixo (TST-04); sem rede, sem gh real.
type dubla struct {
	calls []string
	resp  map[string]string
}

func (d *dubla) run(_ string, name string, args ...string) (string, error) {
	line := name + " " + strings.Join(args, " ")
	d.calls = append(d.calls, line)
	for p, out := range d.resp {
		if strings.HasPrefix(line, p) {
			return out, nil
		}
	}
	return "", nil
}

func constituicao(t *testing.T, valor string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "sdd", "memory")
	require.NoError(t, os.MkdirAll(p, 0o755))
	txt := "## Ferramentas e Integrações\n| Campo | Valor |\n|---|---|\n| VCS / Work Item System | " + valor + " |\n"
	require.NoError(t, os.WriteFile(filepath.Join(p, "constitution.md"), []byte(txt), 0o644))
	return dir
}

func TestGitHubCriaPRComDraft(t *testing.T) {
	d := &dubla{resp: map[string]string{"gh pr create": "https://github.com/o/r/pull/42\n"}}
	a, err := Novo("github", d.run)
	require.NoError(t, err)
	pr, err := a.CriarPRDraft("/proj", NovoPR{Base: "main", Branch: "feat/x", Titulo: "t", CorpoArquivo: "corpo.md"})
	require.NoError(t, err)
	assert.Equal(t, 42, pr.Numero)
	assert.Equal(t, "https://github.com/o/r/pull/42", pr.URL)
	require.Len(t, d.calls, 1)
	assert.Contains(t, d.calls[0], " --draft")
	assert.True(t, strings.HasPrefix(d.calls[0], "gh pr create --base main --head feat/x --title t --body-file corpo.md"))

	// marcar pronto e mesclar usam número fixo, sem shell
	require.NoError(t, a.MarcarPronto("/proj", 42))
	require.NoError(t, a.Mesclar("/proj", Mescla{Numero: 42, Branch: "feat/x", Base: "main"}))
	assert.Contains(t, d.calls, "gh pr ready 42")
	assert.Contains(t, d.calls, "gh pr merge 42")

	// procurar PR aberto lê o JSON do gh
	d.resp["gh pr list"] = `[{"number":7,"url":"https://x/pull/7","baseRefName":"dev"}]`
	got, err := a.ProcurarPRAberto("/proj", "feat/x")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, PR{Numero: 7, URL: "https://x/pull/7", Base: "dev"}, *got)
}

func TestGitHubRecusaSaidaSemNumero(t *testing.T) {
	d := &dubla{resp: map[string]string{"gh pr create": "algo estranho"}}
	a, _ := Novo("github", d.run)
	_, err := a.CriarPRDraft("/p", NovoPR{Base: "main", Branch: "b", Titulo: "t", CorpoArquivo: "c.md"})
	assert.Error(t, err)
}

func TestNenhumDeixaABranchPronta(t *testing.T) {
	d := &dubla{}
	a, err := Novo("nenhum", d.run)
	require.NoError(t, err)
	_, err = a.CriarPRDraft("/p", NovoPR{Base: "main", Branch: "feat/x", Titulo: "t"})
	require.ErrorIs(t, err, ErrSemPR)
	assert.Contains(t, err.Error(), "feat/x")
	assert.True(t, errors.Is(err, ErrSemPR))
}

func TestNenhumNaoExecutaComandoDeVCS(t *testing.T) {
	d := &dubla{}
	a, _ := Novo("nenhum", d.run)
	got, err := a.ProcurarPRAberto("/p", "feat/x")
	require.NoError(t, err)
	assert.Nil(t, got)
	_, _ = a.CriarPRDraft("/p", NovoPR{Branch: "feat/x"})
	assert.ErrorIs(t, a.MarcarPronto("/p", 1), ErrSemPR)
	assert.Empty(t, d.calls, "nenhum não pode rodar gh nem git para PR")

	// o merge local só usa git, na base atual; nunca gh
	d.resp = map[string]string{"git rev-parse --abbrev-ref HEAD": "main"}
	require.NoError(t, a.Mesclar("/p", Mescla{Branch: "feat/x", Base: "main"}))
	for _, c := range d.calls {
		assert.True(t, strings.HasPrefix(c, "git "), c)
	}
	assert.Contains(t, d.calls, "git merge --no-ff --no-edit feat/x")

	// fora da base, recusa sem mesclar
	d.calls, d.resp = nil, map[string]string{"git rev-parse --abbrev-ref HEAD": "outra"}
	assert.Error(t, a.Mesclar("/p", Mescla{Branch: "feat/x", Base: "main"}))
	assert.NotContains(t, strings.Join(d.calls, "\n"), "git merge")
}

func TestLeVCSDaConstituicao(t *testing.T) {
	for _, v := range []string{"github", "nenhum", "azure-devops"} {
		got, err := LeVCS(constituicao(t, v))
		require.NoError(t, err)
		assert.Equal(t, v, got)
	}
	// sem constituição ou sem a linha: padrão github (projeto legado)
	got, err := LeVCS(t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "github", got)
}

func TestVCSDesconhecidoRecusado(t *testing.T) {
	_, err := Novo("gitlab", (&dubla{}).run)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gitlab")
	_, err = Novo("azure-devops", (&dubla{}).run)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "não verificado")
	_, err = Novo("", (&dubla{}).run)
	assert.Error(t, err)
	_, err = DaConstituicao(constituicao(t, "gitlab"), (&dubla{}).run)
	assert.Error(t, err)
}

func TestExecutorRecusaComandoForaDaLista(t *testing.T) {
	for _, nome := range []string{"rm", "az", "sh", "curl"} {
		_, err := RunGitOrGh(t.TempDir(), nome, "x")
		require.Error(t, err, nome)
		assert.Contains(t, err.Error(), "não permitido")
	}
}
