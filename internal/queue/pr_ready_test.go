package queue

import (
	"errors"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shaRevisado = "3f77bbc1a2b3c4d5e6f7081920a1b2c3d4e5f601"
	shaNovo     = "1524df7aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

// adaptadorFalso conta as chamadas e devolve o erro configurado.
type adaptadorFalso struct {
	nome      string
	pronto    int
	mesclado  int
	errPronto error
}

func (a *adaptadorFalso) Nome() string                                     { return a.nome }
func (a *adaptadorFalso) CriarPRDraft(string, vcs.NovoPR) (vcs.PR, error)  { return vcs.PR{}, nil }
func (a *adaptadorFalso) ProcurarPRAberto(string, string) (*vcs.PR, error) { return nil, nil }
func (a *adaptadorFalso) CorrigirBase(string, int, string) error           { return nil }
func (a *adaptadorFalso) MarcarPronto(_ string, n int) error {
	a.pronto++
	return a.errPronto
}
func (a *adaptadorFalso) Mesclar(string, vcs.Mescla) error {
	a.mesclado++
	return nil
}

// itemComParecer monta um item de fila com o parecer informado.
func itemComParecer(veredito, sha string) *QueueItem {
	v := &harness.Parecer{
		Veredito: veredito,
		Resumo:   "camada A e B conformes",
		HeadSHA:  sha,
		Ponto:    harness.PontoEntrega,
		Feature:  "feat-106-21",
	}
	return &QueueItem{TaskID: "feat-106-21", GovernanceVerdict: v}
}

func TestReviewConformeNaoCriaPR(t *testing.T) {
	// O `queue review` de um parecer conforme não abre PR: ele grava o veredito e, no máximo,
	// marca pronto ou mescla, cada um com a sua prova. Abrir PR é do OpenPR, não do Review.
	adapt := &adaptadorFalso{nome: "github"}
	item := itemComParecer(harness.VereditoConforme, shaRevisado)

	// MarcarPronto com o SHA do parecer igual ao HEAD: autoriza e chama o adaptador uma vez.
	require.NoError(t, MarcarPronto("/proj", item, adapt, 42, shaRevisado))
	assert.Equal(t, 1, adapt.pronto)
	assert.Equal(t, 0, adapt.mesclado, "marcar pronto não mescla")

	// MesclarPR com a mesma prova também passa, e é a única forma de mesclar.
	require.NoError(t, MesclarPR("/proj", item, adapt, vcs.Mescla{Numero: 42}, shaRevisado))
	assert.Equal(t, 1, adapt.mesclado)

	// Nenhuma das duas criou PR.
	assert.Equal(t, 1, adapt.pronto)
}

func TestReadySemVereditoDoSHARecusado(t *testing.T) {
	adapt := &adaptadorFalso{nome: "github"}

	casos := map[string]struct {
		item *QueueItem
		head string
		erro error
	}{
		"sem parecer": {
			item: &QueueItem{TaskID: "feat-106-21"},
			head: shaRevisado,
			erro: ErrSemVereditoConforme,
		},
		"parecer de violação": {
			item: itemComParecer(harness.VereditoViolacao, shaRevisado),
			head: shaRevisado,
		},
		"parecer de dúvida": {
			item: itemComParecer(harness.VereditoDuvida, shaRevisado),
			head: shaRevisado,
		},
		"sha vencido": {
			item: itemComParecer(harness.VereditoConforme, shaRevisado),
			head: shaNovo,
			erro: ErrSHAVencido,
		},
		"sem sha no parecer": {
			item: itemComParecer(harness.VereditoConforme, ""),
			head: shaRevisado,
		},
		"head ilegível": {
			item: itemComParecer(harness.VereditoConforme, shaRevisado),
			head: "",
		},
		"item nulo": {
			item: nil,
			head: shaRevisado,
		},
	}

	for nome, c := range casos {
		err := MarcarPronto("/proj", c.item, adapt, 42, c.head)
		require.Error(t, err, nome)
		if c.erro != nil {
			assert.True(t, errors.Is(err, c.erro), "%s: erro %v deveria envolver %v", nome, err, c.erro)
		}
		errMescla := MesclarPR("/proj", c.item, adapt, vcs.Mescla{Numero: 42}, c.head)
		require.Error(t, errMescla, "%s: mesclar também recusa", nome)
	}

	assert.Equal(t, 0, adapt.pronto, "nada foi marcado pronto")
	assert.Equal(t, 0, adapt.mesclado, "nada foi mesclado")
}

func TestMarcarProntoNoVCSSemPR(t *testing.T) {
	adapt := &adaptadorFalso{nome: "nenhum"}
	item := itemComParecer(harness.VereditoConforme, shaRevisado)

	err := MarcarPronto("/proj", item, adapt, 1, shaRevisado)
	require.ErrorIs(t, err, vcs.ErrSemPR)
	assert.Equal(t, 0, adapt.pronto)

	// O contrato do parecer exige o SHA em minúsculas; a comparação com o HEAD é tolerante à
	// caixa, porque a saída do git pode variar.
	gh := &adaptadorFalso{nome: "github"}
	itemGitHub := itemComParecer(harness.VereditoConforme, shaRevisado)
	require.NoError(t, MarcarPronto("/proj", itemGitHub, gh, 7, strings.ToUpper(shaRevisado)))
	assert.Equal(t, 1, gh.pronto)

	// Já um parecer com SHA em maiúsculas é inválido pelo contrato: recusado.
	upper := itemComParecer(harness.VereditoConforme, strings.ToUpper(shaRevisado))
	assert.Error(t, MarcarPronto("/proj", upper, &adaptadorFalso{nome: "github"}, 7, shaRevisado))
}

func TestResumoParaPR(t *testing.T) {
	item := itemComParecer(harness.VereditoConforme, shaRevisado)
	assert.Equal(t, "camada A e B conformes (revisado em 3f77bbc)", ResumoParaPR(item))

	semSHA := itemComParecer(harness.VereditoConforme, "")
	assert.Equal(t, "camada A e B conformes", ResumoParaPR(semSHA))

	assert.Empty(t, ResumoParaPR(nil))
	assert.Empty(t, ResumoParaPR(&QueueItem{TaskID: "x"}))
}
