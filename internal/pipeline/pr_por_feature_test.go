package pipeline

import (
	"errors"
	"testing"

	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adaptadorDeFeature registra chamadas e simula PR aberto.
type adaptadorDeFeature struct {
	nome        string
	existente   *vcs.PR
	criados     int
	baseCorrig  int
	basesCorre  []string
	erroProcura error
}

func (a *adaptadorDeFeature) Nome() string { return a.nome }
func (a *adaptadorDeFeature) CriarPRDraft(_ string, p vcs.NovoPR) (vcs.PR, error) {
	a.criados++
	return vcs.PR{Numero: 100 + a.criados, URL: "https://x/pull/" + p.Branch, Base: p.Base}, nil
}
func (a *adaptadorDeFeature) ProcurarPRAberto(string, string) (*vcs.PR, error) {
	return a.existente, a.erroProcura
}
func (a *adaptadorDeFeature) CorrigirBase(_ string, _ int, base string) error {
	a.baseCorrig++
	a.basesCorre = append(a.basesCorre, base)
	return nil
}
func (a *adaptadorDeFeature) MarcarPronto(string, int) error   { return nil }
func (a *adaptadorDeFeature) Mesclar(string, vcs.Mescla) error { return nil }

func TestUmaBranchEUmPRDraftPorFeature(t *testing.T) {
	feixes, err := FeixesParalelos("106-pr-draft-revisao-e-ci",
		[]string{"24a", "24b", "24c"}, "main")
	require.NoError(t, err)
	require.Len(t, feixes, 3, "uma branch e um PR por feature")

	branches := map[string]bool{}
	for _, f := range feixes {
		assert.Equal(t, "main", f.Base, "todas as bases são a principal")
		assert.Equal(t, "main", f.Principal)
		assert.False(t, branches[f.Branch], "branch repetida: %s", f.Branch)
		branches[f.Branch] = true
	}
	assert.True(t, branches["feat/106-pr-draft-revisao-e-ci-24a"])
	assert.True(t, branches["feat/106-pr-draft-revisao-e-ci-24b"])

	// PR novo é criado em draft, com a base na principal.
	adapt := &adaptadorDeFeature{nome: "github"}
	pr, err := AbrirPRDaFeature("/proj", adapt, feixes[0], vcs.NovoPR{Titulo: "feat 24a"})
	require.NoError(t, err)
	require.NotNil(t, pr)
	assert.Equal(t, 1, adapt.criados)
	assert.Equal(t, "main", pr.Base)

	// Com PR já aberto, reaproveita: nenhum PR novo.
	adapt2 := &adaptadorDeFeature{nome: "github", existente: &vcs.PR{Numero: 7, URL: "u", Base: "main"}}
	pr2, err := AbrirPRDaFeature("/proj", adapt2, feixes[0], vcs.NovoPR{})
	require.NoError(t, err)
	assert.Equal(t, 7, pr2.Numero)
	assert.Equal(t, 0, adapt2.criados, "não abre um segundo PR para a mesma feature")
}

func TestFeatureSemPRDraftNaoAvanca(t *testing.T) {
	// Base na branch da iniciativa: recusa, porque é exatamente o que a RN-22 proíbe.
	errado := ContratoDePRPorFeature{
		Branch:    "feat/106-x-24a",
		Base:      "feat/106-pr-draft-revisao-e-ci",
		Principal: "main",
	}
	require.ErrorIs(t, errado.Validar(), ErrBaseDaFeatureInvalida)

	adapt := &adaptadorDeFeature{nome: "github"}
	_, err := AbrirPRDaFeature("/proj", adapt, errado, vcs.NovoPR{})
	require.ErrorIs(t, err, ErrBaseDaFeatureInvalida)
	assert.Equal(t, 0, adapt.criados, "nada foi aberto com base errada")

	// Branch vazia: sem PR, a feature não avança.
	vazia := ContratoDePRPorFeature{Base: "main", Principal: "main"}
	require.ErrorIs(t, vazia.Validar(), ErrFeatureSemPRDraft)

	// Base vazia: também recusa.
	semBase := ContratoDePRPorFeature{Branch: "feat/106-x-24a", Principal: "main"}
	require.ErrorIs(t, semBase.Validar(), ErrBaseDaFeatureInvalida)

	// VCS nenhum: a branch fica pronta, com aviso, não erro de contrato.
	feixes, err := FeixesParalelos("106-x", []string{"24a"}, "main")
	require.NoError(t, err)
	nenhum := &adaptadorDeFeature{nome: "nenhum"}
	_, err = AbrirPRDaFeature("/proj", nenhum, feixes[0], vcs.NovoPR{})
	require.ErrorIs(t, err, vcs.ErrSemPR)

	// Erro do adaptador ao procurar PR também impede o avanço.
	quebrado := &adaptadorDeFeature{nome: "github", erroProcura: errors.New("rede")}
	_, err = AbrirPRDaFeature("/proj", quebrado, feixes[0], vcs.NovoPR{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "procurar PR aberto")
}

func TestPRComBaseErradaEhCorrigido(t *testing.T) {
	feixes, err := FeixesParalelos("106-x", []string{"24a"}, "main")
	require.NoError(t, err)

	adapt := &adaptadorDeFeature{nome: "github", existente: &vcs.PR{
		Numero: 9, URL: "u", Base: "feat/106-x",
	}}
	pr, err := AbrirPRDaFeature("/proj", adapt, feixes[0], vcs.NovoPR{})
	require.NoError(t, err)
	assert.Equal(t, 1, adapt.baseCorrig, "a base errada é corrigida para a principal")
	assert.Equal(t, "main", pr.Base)
}

func TestFeixesParalelosIgnoraRepetidas(t *testing.T) {
	feixes, err := FeixesParalelos("106-x", []string{"a", "a", "b", ""}, "main")
	require.NoError(t, err)
	assert.Len(t, feixes, 2)

	_, err = FeixesParalelos("106-x", nil, "main")
	require.Error(t, err)

	_, err = FeixesParalelos("", []string{"a"}, "main")
	require.Error(t, err)
}

func (a *adaptadorDeFeature) CriarIssue(string, vcs.NovaIssue) (vcs.Issue, error) {
	return vcs.Issue{}, vcs.ErrSemPR
}

func (a *adaptadorDeFeature) ProcurarIssueAberta(string, string) (*vcs.Issue, error) {
	return nil, vcs.ErrSemPR
}
