package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adaptadorDeSpec registra o que foi mesclado.
type adaptadorDeSpec struct {
	nome     string
	mesclou  int
	errMerge error
}

func (a *adaptadorDeSpec) Nome() string { return a.nome }
func (a *adaptadorDeSpec) CriarPRDraft(string, vcs.NovoPR) (vcs.PR, error) {
	return vcs.PR{}, nil
}
func (a *adaptadorDeSpec) ProcurarPRAberto(string, string) (*vcs.PR, error) { return nil, nil }
func (a *adaptadorDeSpec) CorrigirBase(string, int, string) error           { return nil }
func (a *adaptadorDeSpec) MarcarPronto(string, int) error                   { return nil }
func (a *adaptadorDeSpec) Mesclar(string, vcs.Mescla) error {
	a.mesclou++
	return a.errMerge
}

func TestSpecMescladaPeloAdaptadorAposSeguir(t *testing.T) {
	decisao := AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true,
		Seguir:               true,
		ModoRevisao:          "com-humano",
		Sinal:                SinalSpecParaRevisao,
	})
	require.True(t, decisao.Pode, decisao.Motivo)

	adapt := &adaptadorDeSpec{nome: "github"}
	res, err := VoltaSpecParaPrincipal("/proj", adapt, vcs.Mescla{Numero: 12, Branch: "feat/1-x", Base: "main"}, decisao)
	require.NoError(t, err)
	assert.True(t, res.Mesclada)
	assert.False(t, res.Local)
	assert.Equal(t, 1, adapt.mesclou)

	// Depois da volta não fica sinal pendente: a conferência daquele ponto terminou.
	assert.Empty(t, SinalAposVolta(time.Now()).LinhaStatus())
}

func TestNenhumMesclaLocal(t *testing.T) {
	decisao := AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true,
		Seguir:               true,
		ModoRevisao:          "com-humano",
	})
	require.True(t, decisao.Pode)

	adapt := &adaptadorDeSpec{nome: "nenhum"}
	res, err := VoltaSpecParaPrincipal("/proj", adapt, vcs.Mescla{Branch: "feat/1-x", Base: "main"}, decisao)
	require.NoError(t, err)
	assert.False(t, res.Mesclada, "no nenhum nada é mesclado por conta própria")
	assert.True(t, res.Local)
	assert.Contains(t, res.Aviso, "feat/1-x")
	assert.Equal(t, 0, adapt.mesclou)

	// Modo sem humano também não têm PR: a decisão automática registrada basta.
	semHumano := AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true,
		ModoRevisao:          "sem-humano",
	})
	assert.True(t, semHumano.Pode, "sem humano não pede [Seguir], mas continua exigindo a Conferência 1")
}

func TestSpecNaoMesclaSemSeguirNoModoComHumano(t *testing.T) {
	// Conferência 1 conforme, mas o humano não disse [Seguir]: a spec não volta.
	decisao := AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true,
		Seguir:               false,
		ModoRevisao:          "com-humano",
	})
	assert.False(t, decisao.Pode)

	adapt := &adaptadorDeSpec{nome: "github"}
	_, err := VoltaSpecParaPrincipal("/proj", adapt, vcs.Mescla{Numero: 12}, decisao)
	require.ErrorIs(t, err, ErrSpecSemSeguir)
	assert.Equal(t, 0, adapt.mesclou, "nada foi mesclado")

	// Modo desconhecido também não passa.
	assert.False(t, AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true, ModoRevisao: "sozinho",
	}).Pode)
}

func TestSpecNaoVoltaSemConferencia1(t *testing.T) {
	// Sem a Conferência 1, nem o [Seguir] libera: a revisão não é dispensável.
	for _, modo := range []string{"com-humano", "sem-humano", ""} {
		decisao := AutorizaVoltaDaSpec(VoltaDaSpec{
			Conferencia1Conforme: false,
			Seguir:               true,
			ModoRevisao:          modo,
		})
		assert.False(t, decisao.Pode, "modo %q não pode voltar sem conferência", modo)
	}

	adapt := &adaptadorDeSpec{nome: "github"}
	_, err := VoltaSpecParaPrincipal("/proj", adapt, vcs.Mescla{Numero: 12},
		DecisaoVolta{Motivo: ErrSpecSemConferencia1.Error()})
	require.ErrorIs(t, err, ErrSpecSemConferencia1)
	assert.Equal(t, 0, adapt.mesclou)

	// Sinal de outro ponto não serve para a Conferência 1.
	assert.False(t, AutorizaVoltaDaSpec(VoltaDaSpec{
		Conferencia1Conforme: true,
		Seguir:               true,
		ModoRevisao:          "com-humano",
		Sinal:                SinalProntoParaRevisao,
	}).Pode)

	// Sem adaptador, a volta autorizada não acontece e diz por quê.
	_, err = VoltaSpecParaPrincipal("/proj", nil, vcs.Mescla{Numero: 1},
		DecisaoVolta{Pode: true})
	require.ErrorIs(t, err, ErrSpecSemAdaptador)
}

func TestVoltaDaSpecPropagaFalhaDoAdaptador(t *testing.T) {
	decisao := DecisaoVolta{Pode: true, Motivo: "ok"}
	adapt := &adaptadorDeSpec{nome: "github", errMerge: errors.New("PR não pronto")}
	_, err := VoltaSpecParaPrincipal("/proj", adapt, vcs.Mescla{Numero: 12}, decisao)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mesclar a spec")
}
