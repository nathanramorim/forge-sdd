package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModoComHumanoEPadrao(t *testing.T) {
	// Sem escolha, o modo é com-humano e nada é registrado.
	d, err := ResolveModoDeRevisao(EscolhaDeModo{})
	require.NoError(t, err)
	assert.Equal(t, ModoComHumano, d.Modo)
	assert.False(t, d.Registrar)

	// com-humano explícito também não pede aviso.
	d2, err := ResolveModoDeRevisao(EscolhaDeModo{Modo: ModoComHumano})
	require.NoError(t, err)
	assert.Equal(t, ModoComHumano, d2.Modo)
	assert.False(t, d2.Registrar)

	// O padrão da Constituição lido pelos modos é com-humano.
	assert.Equal(t, ModoComHumano, LeModoDaConstituicao(vcs.Modos{}))
	assert.Equal(t, ModoComHumano, LeModoDaConstituicao(vcs.Modos{RevisaoHumana: ""}))
	assert.Equal(t, ModoSemHumano, LeModoDaConstituicao(vcs.Modos{RevisaoHumana: "sem-humano"}))
}

func TestModoSemHumanoExigeAvisoDeRiscos(t *testing.T) {
	// Sem o aviso apresentado: recusa e mantém com-humano.
	d, err := ResolveModoDeRevisao(EscolhaDeModo{Modo: ModoSemHumano})
	require.ErrorIs(t, err, ErrSemHumanoSemAviso)
	assert.Equal(t, ModoComHumano, d.Modo, "recusar mantém o modo seguro")
	assert.Contains(t, err.Error(), "aviso de riscos")

	// Aviso apresentado mas riscos não aceitos: recusa também.
	d2, err := ResolveModoDeRevisao(EscolhaDeModo{
		Modo:             ModoSemHumano,
		AvisoApresentado: true,
		Aceito:           false,
	})
	require.ErrorIs(t, err, ErrSemHumanoSemAviso)
	assert.Equal(t, ModoComHumano, d2.Modo)
	assert.Contains(t, err.Error(), "não aceitou")

	// Aviso apresentado e aceito: liga e registra.
	em := time.Date(2026, 11, 21, 21, 0, 0, 0, time.UTC)
	d3, err := ResolveModoDeRevisao(EscolhaDeModo{
		Modo:             ModoSemHumano,
		AvisoApresentado: true,
		Aceito:           true,
		RegistradoEm:     em,
	})
	require.NoError(t, err)
	assert.Equal(t, ModoSemHumano, d3.Modo)
	assert.True(t, d3.Registrar, "a aceitação precisa ser gravada")
	assert.Equal(t, em, d3.AceitoEm)

	// Modo desconhecido recusa e mantém com-humano.
	d4, err := ResolveModoDeRevisao(EscolhaDeModo{Modo: "sozinho"})
	require.ErrorIs(t, err, ErrModoRevisaoInvalido)
	assert.Equal(t, ModoComHumano, d4.Modo)
	assert.True(t, errors.Is(err, ErrModoRevisaoInvalido))

	// O aviso tem os três riscos, escritos: nenhum é resumido.
	aviso := AvisoDoModoSemHumano()
	for _, risco := range RiscosDoModoSemHumano {
		assert.Contains(t, aviso, risco)
	}
	assert.Contains(t, aviso, "nenhuma pessoa vê o código")
	assert.Contains(t, aviso, "sem teto humano")
}
