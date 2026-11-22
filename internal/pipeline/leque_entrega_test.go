package pipeline

import (
	"testing"

	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adaptadorDeLeque simula o VCS para a publicação do leque.
type adaptadorDeLeque struct {
	nome      string
	abertos   map[string]*vcs.PR
	criados   int
	erroEm    string // branch que falha ao procurar
	erroCriar bool
}

func (a *adaptadorDeLeque) Nome() string { return a.nome }
func (a *adaptadorDeLeque) CriarPRDraft(_ string, p vcs.NovoPR) (vcs.PR, error) {
	if a.erroCriar {
		return vcs.PR{}, errDeTeste("falha ao criar")
	}
	a.criados++
	return vcs.PR{Numero: 200 + a.criados, URL: "https://x/pull", Base: p.Base}, nil
}
func (a *adaptadorDeLeque) ProcurarPRAberto(_ string, branch string) (*vcs.PR, error) {
	if branch == a.erroEm {
		return nil, errDeTeste("falha ao procurar")
	}
	return a.abertos[branch], nil
}
func (a *adaptadorDeLeque) CorrigirBase(string, int, string) error { return nil }
func (a *adaptadorDeLeque) MarcarPronto(string, int) error         { return nil }
func (a *adaptadorDeLeque) Mesclar(string, vcs.Mescla) error       { return nil }

type errDeTeste string

func (e errDeTeste) Error() string { return string(e) }

func TestPrepararLequeUsaAPrincipalComoBase(t *testing.T) {
	// No github, a base do leque precisa ser a principal.
	c := ContratoDoLeque{Iniciativa: "106-x", Base: "main", Principal: "main", VCS: "github"}
	require.NoError(t, c.Validar())
	assert.Equal(t, "main", c.BaseEfetiva())

	// A branch da iniciativa não serve de base no github.
	errado := ContratoDoLeque{Iniciativa: "106-x", Base: "feat/106-x", Principal: "main", VCS: "github"}
	require.ErrorIs(t, errado.Validar(), ErrLequeBaseNaoEhPrincipal)

	// No nenhum, a base é livre (não há PR, o merge é local).
	nenhum := ContratoDoLeque{Iniciativa: "106-x", Base: "feat/106-x", Principal: "main", VCS: "nenhum"}
	require.NoError(t, nenhum.Validar())

	// Padrão de principal quando não informada.
	padrao := ContratoDoLeque{Iniciativa: "106-x", Base: "main", VCS: "github"}
	require.NoError(t, padrao.Validar())
	assert.Equal(t, "main", padrao.PrincipalEfetiva())

	// Base vazia recusa; VCS desconhecido recusa.
	require.Error(t, ContratoDoLeque{Iniciativa: "106-x", VCS: "github"}.Validar())
	require.Error(t, ContratoDoLeque{Iniciativa: "106-x", Base: "main", VCS: "svn"}.Validar())
	require.Error(t, ContratoDoLeque{Base: "main", VCS: "github"}.Validar())
}

func TestJuntarSoNoMergeLocalDoNenhum(t *testing.T) {
	t.Run("nenhum permite a juncao local", testJuntarLocal)
}

func testJuntarLocal(t *testing.T) {
	// No nenhum, a junção local é o merge de fato.
	assert.NoError(t, JuntarSoNoMergeLocalDoNenhum("nenhum"))
	assert.NoError(t, JuntarSoNoMergeLocalDoNenhum(""))

	// No github, juntar no worktree principal não é a entrega.
	err := JuntarSoNoMergeLocalDoNenhum("github")
	require.ErrorIs(t, err, ErrJuncaoNaoEhEntrega)
	assert.Contains(t, err.Error(), "não para entregar")

	// VCS desconhecido recusa.
	require.Error(t, JuntarSoNoMergeLocalDoNenhum("azure-devops"))
}

func TestPublicarAbrePRDraftPeloAdaptador(t *testing.T) {
	adapt := &adaptadorDeLeque{nome: "github", abertos: map[string]*vcs.PR{
		"feat/106-x-24b": {Numero: 5, URL: "u", Base: "main"}, // já existe
	}}

	out, err := PublicarAbrePRDraft("/proj", "106-x", []string{"24b", "24c"}, adapt, "main")
	require.NoError(t, err)
	require.Len(t, out, 2)

	// A que já existia é reaproveitada; a outra é criada em draft.
	assert.True(t, out[0].JaExistia)
	assert.Equal(t, 5, out[0].PR.Numero)
	assert.False(t, out[1].JaExistia)
	assert.Equal(t, 1, adapt.criados, "só uma feature precisou de PR novo")

	// Todas as branches são da feature, com base na principal.
	for _, p := range out {
		assert.Contains(t, p.Branch, "106-x-")
		assert.Empty(t, p.Erro)
		if p.PR != nil {
			assert.NotEmpty(t, p.PR.Base)
		}
	}
}

func TestPublicarFalhaParcialNaoImpedeAsOutras(t *testing.T) {
	adapt := &adaptadorDeLeque{nome: "github", abertos: map[string]*vcs.PR{},
		erroEm: "feat/106-x-24b"}
	out, err := PublicarAbrePRDraft("/proj", "106-x", []string{"24b", "24c"}, adapt, "main")
	require.NoError(t, err, "uma falha não derruba a publicação inteira")
	require.Len(t, out, 2)
	assert.NotEmpty(t, out[0].Erro, "a que falhou traz o motivo")
	assert.Empty(t, out[1].Erro, "a outra foi publicada")
	assert.Equal(t, 1, adapt.criados)

	// Todas falhando: devolve erro, com as linhas individuais.
	quebrado := &adaptadorDeLeque{nome: "github", abertos: map[string]*vcs.PR{}, erroCriar: true}
	out2, err2 := PublicarAbrePRDraft("/proj", "106-x", []string{"a", "b"}, quebrado, "main")
	require.ErrorIs(t, err2, ErrFeaturesSemPR)
	assert.Len(t, out2, 2)

	// Lista vazia: nada a publicar é erro.
	_, err = PublicarAbrePRDraft("/proj", "106-x", nil, adapt, "main")
	require.ErrorIs(t, err, ErrFeaturesSemPR)
}

func TestPublicarNoNenhumNaoAbrePR(t *testing.T) {
	adapt := &adaptadorDeLeque{nome: "nenhum"}
	_, err := PublicarAbrePRDraft("/proj", "106-x", []string{"24b"}, adapt, "main")
	require.Error(t, err)
	assert.Equal(t, 0, adapt.criados, "no nenhum nada é publicado como PR")
}

// TestJuntarNaoEhEntregaNoGitHub: no github, a junção local não é a entrega — a entrega é o PR.
func TestJuntarNaoEhEntregaNoGitHub(t *testing.T) {
	err := JuntarSoNoMergeLocalDoNenhum("github")
	require.ErrorIs(t, err, ErrJuncaoNaoEhEntrega)
	assert.Contains(t, err.Error(), "PR em draft")

	// E publicar no github é o caminho da entrega.
	adapt := &adaptadorDeLeque{nome: "github", abertos: map[string]*vcs.PR{}}
	out, err := PublicarAbrePRDraft("/proj", "106-x", []string{"24b"}, adapt, "main")
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.NotNil(t, out[0].PR, "no github a entrega é o PR")
	assert.Equal(t, 1, adapt.criados)
}

func (a *adaptadorDeLeque) CriarIssue(string, vcs.NovaIssue) (vcs.Issue, error) {
	return vcs.Issue{}, vcs.ErrSemPR
}

func (a *adaptadorDeLeque) ProcurarIssueAberta(string, string) (*vcs.Issue, error) {
	return nil, vcs.ErrSemPR
}
