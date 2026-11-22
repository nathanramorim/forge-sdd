package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sinalEm = time.Date(2026, 11, 21, 18, 0, 0, 0, time.UTC)

func TestConferencia1AntesDoRetornoDaSpec(t *testing.T) {
	// A spec só volta para a principal depois do sinal da Conferência 1.
	s := SinalDeSpec(sinalEm)
	assert.Equal(t, SinalSpecParaRevisao, s.Ponto)
	assert.Equal(t, "Conferência 1 (spec)", NomeiaCiclo(s.Ponto))

	linha := s.LinhaStatus()
	assert.Equal(t, "Sinal: spec-para-revisao 2026-11-21T18:00:00Z", linha)

	lido, err := ParseLinhaSinal(linha)
	require.NoError(t, err)
	assert.Equal(t, s, lido)

	// Sem sinal, nada dispara: a spec não pode ser devolvida sem a conferência.
	assert.Empty(t, Sinal{}.LinhaStatus())
	assert.Empty(t, NomeiaCiclo(SinalNenhum))
}

func TestConferencia2PorFeature(t *testing.T) {
	s := SinalDeFeature(sinalEm)
	assert.Equal(t, "Conferência 2 (feature)", NomeiaCiclo(s.Ponto))

	lido, err := ParseLinhaSinal(s.LinhaStatus())
	require.NoError(t, err)
	assert.Equal(t, s, lido)
	assert.Equal(t, "pronto-para-revisao", lido.Ponto)
}

func TestConferencia2NaoRodaAntesDoSinal(t *testing.T) {
	// Sem a linha de sinal, o parse falha e a conferência não é disparada.
	for _, linha := range []string{
		"",
		"Sinal:",
		"Sinal: pronto-para-revisao",
		"Pronto: pronto-para-revisao 2026-11-21T18:00:00Z",
	} {
		_, err := ParseLinhaSinal(linha)
		assert.Error(t, err, "linha %q não pode ser aceita", linha)
	}

	// Sinal desconhecido ou data inválida também são recusados (SEG-01).
	_, err := ParseLinhaSinal("Sinal: conferencia-3 2026-11-21T18:00:00Z")
	assert.Error(t, err)
	_, err = ParseLinhaSinal("Sinal: pronto-para-revisao ontem")
	assert.Error(t, err)
	_, err = ParseLinhaSinal("Sinal: pronto-para-revisao 2026-11-21T18:00:00Z extra")
	assert.Error(t, err)

	// Um sinal de spec não serve para a feature: o ponto tem de casar.
	lido, err := ParseLinhaSinal("Sinal: spec-para-revisao 2026-11-21T18:00:00Z")
	require.NoError(t, err)
	assert.NotEqual(t, SinalProntoParaRevisao, lido.Ponto)
}

func TestTerceiraRodadaPerguntaAoHumano(t *testing.T) {
	teto := NovoTetoRodadas(3)

	// Rodadas 1, 2 e 3 rodam.
	for rodada := 1; rodada <= 3; rodada++ {
		assert.False(t, teto.Estourou(rodada), "rodada %d está dentro do teto", rodada)
	}

	// A quarta estoura e vira pergunta ao humano, nunca aprovação silenciosa.
	assert.True(t, teto.Estourou(4))
	pergunta := PerguntaAoTeto("feat-106-21", 4, 3)
	assert.Contains(t, pergunta, "rodada 4")
	assert.Contains(t, pergunta, "teto 3")
	assert.Contains(t, pergunta, "Seguir assim mesmo")

	// Valor inválido de configuração cai no padrão, não solta o laço.
	padrao := NovoTetoRodadas(0)
	assert.Equal(t, MaxRodadasPorFeature, padrao.Max)
	assert.True(t, padrao.Estourou(MaxRodadasPorFeature+1))
	assert.False(t, padrao.Estourou(MaxRodadasPorFeature))

	negativo := NovoTetoRodadas(-5)
	assert.Equal(t, MaxRodadasPorFeature, negativo.Max)

	// A contagem de rodadas só lê o formato gravado, ignorando prosa.
	linhas := []string{
		"# Status",
		LinhaRodada(1, "ajustar"),
		"Rodada: abc — lixo",
		LinhaRodada(2, "conforme"),
	}
	assert.Equal(t, 2, RodadasUsadas(linhas))
	assert.Empty(t, LinhaRodada(0, "x"))
	assert.Equal(t, "Rodada: 1 — sem resultado", LinhaRodada(1, ""))
}

func TestLinhasModoStatus(t *testing.T) {
	linhas := LinhasModoStatus("local", "com-humano")
	require.Len(t, linhas, 2)
	assert.Equal(t, "Atestado: local", linhas[0])
	assert.Equal(t, "Modo: com-humano", linhas[1])

	// Modo não lido não vira valor inventado.
	vazias := LinhasModoStatus("", "  ")
	assert.Equal(t, "Atestado: indefinido", vazias[0])
	assert.Equal(t, "Modo: indefinido", vazias[1])
}
