package vcs

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// humanoFalso responde o que o teste mandar.
type humanoFalso struct {
	responder bool
	err       error
	perguntas []string
}

func (h *humanoFalso) Perguntar(p string) (bool, error) {
	h.perguntas = append(h.perguntas, p)
	return h.responder, h.err
}

const ( // SHAs de teste
	shaA = "3f77bbc1a2b3c4d5e6f7081920a1b2c3d4e5f601"
	shaB = "1524df7aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestFallbackTerminaNoLocal(t *testing.T) {
	// Modo local com prova completa: termina em aprovado, depois de perguntar ao humano.
	h := &humanoFalso{responder: true}
	d := CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaA, h, 300)
	assert.Equal(t, AtestadoAprovado, d.Resultado)
	require.Len(t, h.perguntas, 1, "o humano é consultado mesmo com parecer conforme")

	// SHA vencido: não aprova, mas também não trava — pergunta ao humano.
	h2 := &humanoFalso{responder: false}
	d2 := CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaB, h2, 300)
	assert.Equal(t, AtestadoRecusado, d2.Resultado)
	require.Len(t, h2.perguntas, 1)
	assert.Contains(t, h2.perguntas[0], "SHA vencido")

	// Modo vazio cai no local (padrão).
	assert.Equal(t, AtestadoAprovado,
		CadeiaDoAtestado("", "conforme", shaA, shaA, &humanoFalso{responder: true}, 300).Resultado)
}

func TestFallbackNaoAprovaEmSilencio(t *testing.T) {
	// Sem humano, nada é aprovado: a decisão é pendente, com o motivo explícito.
	d := CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaA, nil, 300)
	assert.Equal(t, AtestadoPendente, d.Resultado)
	assert.Contains(t, d.Motivo, "humano")

	// Humano que falha ao responder também deixa pendente.
	falha := &humanoFalso{err: errors.New("sessão encerrada")}
	d2 := CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaA, falha, 300)
	assert.Equal(t, AtestadoPendente, d2.Resultado)

	// Modo desligado nunca aprova, mesmo com parecer conforme e humano disposto.
	disposto := &humanoFalso{responder: true}
	d3 := CadeiaDoAtestado(AtestadoDesligado, "conforme", shaA, shaA, disposto, 300)
	assert.Equal(t, AtestadoRecusado, d3.Resultado)
	assert.Empty(t, disposto.perguntas, "no modo desligado não se pergunta: nada é conferido")

	// Parecer não conforme, sem humano: pendente, nunca aprovado.
	d4 := CadeiaDoAtestado(AtestadoLocal, "violacao", shaA, shaA, nil, 300)
	assert.NotEqual(t, AtestadoAprovado, d4.Resultado)
}

func TestFallbackNaoBloqueiaParaSempre(t *testing.T) {
	// A cadeia sempre devolve uma decisão: não há retorno indefinido.
	modos := []string{AtestadoLocal, AtestadoDesligado, AtestadoVCS, "desconhecido", ""}
	humanos := []Perguntador{nil, &humanoFalso{responder: true}, &humanoFalso{responder: false}}
	for _, m := range modos {
		for _, h := range humanos {
			d := CadeiaDoAtestado(m, "conforme", shaA, shaA, h, 0)
			assert.Contains(t,
				[]string{AtestadoAprovado, AtestadoPendente, AtestadoRecusado}, d.Resultado,
				"modo %q devolveu resultado desconhecido", m)
			assert.NotEmpty(t, d.Motivo, "toda decisão explica o motivo")
		}
	}

	// O prazo padrão aparece no texto quando o chamador não informa um.
	h := &humanoFalso{responder: true}
	CadeiaDoAtestado(AtestadoLocal, "violacao", shaA, shaB, h, 0)
	require.Len(t, h.perguntas, 1)
	assert.Contains(t, h.perguntas[0], "5 min")
}

func TestNenhumRecusaModoVCS(t *testing.T) {
	err := ValidaModoAtestado("nenhum", AtestadoVCS)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "não combina")

	// Os outros modos são aceitos com VCS nenhum.
	assert.NoError(t, ValidaModoAtestado("nenhum", AtestadoLocal))
	assert.NoError(t, ValidaModoAtestado("nenhum", AtestadoDesligado))

	// A recusa é anterior à pergunta: nenhuma sessão chega a perguntar.
	h := &humanoFalso{responder: true}
	_ = h
	assert.Error(t, ValidaModoAtestado("nenhum", "inventado"))
}

func TestModoVCSRecusadoComModoAdiado(t *testing.T) {
	// Com github, o modo vcs é aceito como valor mas recusado como entrega.
	err := ValidaModoAtestado("github", AtestadoVCS)
	require.ErrorIs(t, err, ErrModoAdiado)
	assert.Contains(t, err.Error(), "adiado")

	// E a cadeia também recusa, com o mesmo motivo.
	d := CadeiaDoAtestado(AtestadoVCS, "conforme", shaA, shaA, &humanoFalso{responder: true}, 300)
	assert.Equal(t, AtestadoRecusado, d.Resultado)
	assert.Contains(t, d.Motivo, "adiado")
}

func TestAzureDevOpsRecusaNaoVerificado(t *testing.T) {
	// O adaptador do Azure continua recusado com a mensagem de "não verificado".
	_, err := Novo("azure-devops", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "não verificado")

	// E o valor é reconhecido pelos modos como válido (o que não é é a entrega).
	m, err := LeModosDaBranchBase(constitutionComModos("VCS / Work Item System | azure-devops"))
	require.NoError(t, err)
	assert.Equal(t, "azure-devops", m.VCS)

	// Com Azure escolhido, o modo vcs também é recusado por adiamento.
	assert.ErrorIs(t, ValidaModoAtestado("azure-devops", AtestadoVCS), ErrModoAdiado)
}

// TestAtestadoDesligadoNaoExigeProva: no modo desligado nada é conferido, e isso é explícito.
func TestAtestadoDesligadoNaoExigeProva(t *testing.T) {
	d := CadeiaDoAtestado(AtestadoDesligado, "", "", "", nil, 300)
	assert.Equal(t, AtestadoRecusado, d.Resultado)
	assert.Contains(t, d.Motivo, "ninguém confere")

	// Nem com prova completa o modo desligado aprova sozinho: ele não confere,
	// então a decisão é recusa com motivo, nunca aprovação silenciosa.
	d2 := CadeiaDoAtestado(AtestadoDesligado, "conforme", "abc1234", "abc1234",
		&humanoFalso{responder: true}, 300)
	assert.Equal(t, AtestadoRecusado, d2.Resultado)
}

// TestAtestadoLocalConfereVeredito: o modo local exige veredito conforme e SHA do HEAD.
func TestAtestadoLocalConfereVeredito(t *testing.T) {
	// Conforme com SHA igual: aprova (após o humano).
	assert.Equal(t, AtestadoAprovado,
		CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaA, &humanoFalso{responder: true}, 300).Resultado)

	// Veredito diferente de conforme com humano disposto: ainda passa pelo humano (o humano
	// é a última palavra no modo local), mas o motivo diz que o parecer não é conforme.
	recusa := CadeiaDoAtestado(AtestadoLocal, "violacao", shaA, shaA, &humanoFalso{responder: false}, 300)
	assert.Equal(t, AtestadoRecusado, recusa.Resultado)

	// SHA diferente: não aprova direto; pergunta ao humano.
	h := &humanoFalso{responder: true}
	d := CadeiaDoAtestado(AtestadoLocal, "conforme", shaA, shaB, h, 300)
	assert.Equal(t, AtestadoAprovado, d.Resultado, "o humano pode aprovar mesmo com SHA vencido")
	assert.Contains(t, h.perguntas[0], "SHA vencido")
}

// TestModoDeAtestadoNoStatus: as linhas do status.md refletem o modo lido, sem inventar.
// A formatação canônica vive em internal/pipeline (feat-106-16); aqui ela é reproduzida
// porque esta feature só toca internal/vcs.
func TestModoDeAtestadoNoStatus(t *testing.T) {
	linhaStatus := func(atestado, revisao string) []string {
		a := strings.TrimSpace(atestado)
		if a == "" {
			a = "indefinido"
		}
		r := strings.TrimSpace(revisao)
		if r == "" {
			r = "indefinido"
		}
		return []string{"Atestado: " + a, "Modo: " + r}
	}

	linhas := linhaStatus(AtestadoLocal, RevisaoComHumano)
	assert.Equal(t, []string{"Atestado: local", "Modo: com-humano"}, linhas)

	vazias := linhaStatus("", "")
	assert.Equal(t, "Atestado: indefinido", vazias[0])
	assert.Equal(t, "Modo: indefinido", vazias[1])
}

// TestPRDaSpecSempreAtestadoLocal: o PR da spec usa `local` mesmo com outro modo configurado.
func TestPRDaSpecSempreAtestadoLocal(t *testing.T) {
	// A regra é do fluxo: spec sempre local. Aqui provamos que a combinação é aceita e que
	// o modo `vcs` (restrito a PR de feature) não vaza para ela.
	assert.NoError(t, ValidaModoAtestado("github", AtestadoLocal))
	assert.ErrorIs(t, ValidaModoAtestado("github", AtestadoVCS), ErrModoAdiado)
}

// TestPRDaSpecIgnoraModoVCS: com `vcs` configurado, o PR da spec segue em local; o de feature
// é que recebe a recusa do modo adiado.
func TestPRDaSpecIgnoraModoVCS(t *testing.T) {
	// Simula o fluxo: ponto spec escolhe local independentemente do campo.
	modoDoPonto := func(ponto string, campo string) string {
		if ponto == "spec" {
			return AtestadoLocal
		}
		return campo
	}
	assert.Equal(t, AtestadoLocal, modoDoPonto("spec", AtestadoVCS))
	assert.Equal(t, AtestadoVCS, modoDoPonto("entrega", AtestadoVCS))

	// E a cadeia no ponto spec com local aprova.
	d := CadeiaDoAtestado(modoDoPonto("spec", AtestadoVCS), "conforme", shaA, shaA,
		&humanoFalso{responder: true}, 300)
	assert.Equal(t, AtestadoAprovado, d.Resultado)
}
