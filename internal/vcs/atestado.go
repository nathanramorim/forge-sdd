package vcs

import (
	"errors"
	"fmt"
	"strings"
)

// Resultado do atestado de um PR de entrega.
const (
	// AtestadoAprovado: há prova válida; o PR pode sair do draft.
	AtestadoAprovado = "aprovado"
	// AtestadoPendente: falta a resposta do humano; a sessão espera com prazo.
	AtestadoPendente = "pendente"
	// AtestadoRecusado: a prova não vale (modo desligado, SHA vencido, modo adiado).
	AtestadoRecusado = "recusado"
)

// Erros do atestado.
var (
	// ErrModoAdiado: o modo `vcs` está aceito como valor mas não entregue.
	ErrModoAdiado = errors.New(`modo "vcs" adiado: a prova pelo VCS não está disponível nesta entrega`)
	// ErrAtestadoDesligado: o modo `desligado` não confere nada.
	ErrAtestadoDesligado = errors.New(`modo do atestado "desligado": ninguém confere veredito nem aprovação`)
	// ErrSemHumanoNoModoComHumano: a cadeia chegou ao humano e não há humano no modo com-humano.
	ErrSemHumanoNoModoComHumano = errors.New("atestado aguardando o humano; sem resposta, a sessão não avança")
)

// Perguntador é quem faz a pergunta ao humano no fim da cadeia. Devolve true (aprovou),
// false (recusou) e erro quando não há como perguntar (sem humano, sessão encerrada).
type Perguntador interface {
	Perguntar(pergunta string) (bool, error)
}

// Resposta do atestado: o que a cadeia decidiu e por quê.
type Decisao struct {
	Resultado string
	Motivo    string
}

// CadeiaDoAtestado decide o atestado de um PR de entrega, seguindo o modo lido da Constituição
// (ADR-0022, `FLX-06`). A cadeia termina em `local` e nunca aprova em silêncio: quando não há
// prova nem humano, a decisão é `pendente`, não `aprovado`.
//
// Ordem:
//  1. `desligado` recusa com o motivo explícito — nada foi conferido.
//  2. `vcs` é adiado: recusa com ErrModoAdiado.
//  3. `local`: veredito conforme com SHA igual ao HEAD mais a aprovação do humano.
//  4. Em qualquer ponto sem prova, a cadeia pergunta ao humano (prazo declarado pelo chamador);
//     sem humano disponível, fica `pendente` — nunca aprovada.
func CadeiaDoAtestado(modo, veredito, shaParecer, shaHead string, humano Perguntador, prazoSegundos int) Decisao {
	switch strings.TrimSpace(modo) {
	case AtestadoDesligado:
		return Decisao{Resultado: AtestadoRecusado, Motivo: ErrAtestadoDesligado.Error()}
	case AtestadoVCS:
		return Decisao{Resultado: AtestadoRecusado, Motivo: ErrModoAdiado.Error()}
	case AtestadoLocal, "":
		// segue para a cadeia local
	default:
		return Decisao{Resultado: AtestadoRecusado, Motivo: fmt.Sprintf("modo do atestado %q desconhecido", modo)}
	}

	// Prova local: veredito conforme e SHA do parecer igual ao HEAD atual.
	if strings.EqualFold(strings.TrimSpace(veredito), "conforme") &&
		shaParecer != "" && strings.EqualFold(shaParecer, shaHead) {
		// Mesmo conforme, o humano precisa concordar (ADR-0022).
		return pedirAoHumano(humano, prazoSegundos,
			"O Reviewer deu parecer conforme para este commit. Marcar o PR como pronto?")
	}

	motivo := "sem parecer conforme para o HEAD atual"
	if strings.EqualFold(strings.TrimSpace(veredito), "conforme") {
		motivo = "o parecer é de outro commit (SHA vencido)"
	}
	return pedirAoHumano(humano, prazoSegundos,
		fmt.Sprintf("%s (%s). Marcar o PR como pronto mesmo assim?", motivo, prazo(eficaz(prazoSegundos))))
}

// pedirAoHumano fecha a cadeia: pergunta e decide. Sem humano, fica pendente — nunca aprovada.
func pedirAoHumano(humano Perguntador, prazoSegundos int, pergunta string) Decisao {
	if humano == nil {
		return Decisao{Resultado: AtestadoPendente, Motivo: ErrSemHumanoNoModoComHumano.Error()}
	}
	ok, err := humano.Perguntar(pergunta)
	if err != nil {
		return Decisao{Resultado: AtestadoPendente, Motivo: fmt.Sprintf("pergunta ao humano não respondida: %v", err)}
	}
	if !ok {
		return Decisao{Resultado: AtestadoRecusado, Motivo: "o humano recusou marcar o PR como pronto"}
	}
	return Decisao{Resultado: AtestadoAprovado, Motivo: "veredito conforme e aprovação do humano"}
}

// eficaz aplica o prazo padrão quando o chamador não informa um.
func eficaz(segundos int) int {
	if segundos <= 0 {
		return 300
	}
	return segundos
}

// prazo formata o tempo limite para o texto da pergunta.
func prazo(segundos int) string {
	if segundos%60 == 0 {
		return fmt.Sprintf("%d min", segundos/60)
	}
	return fmt.Sprintf("%d s", segundos)
}

// ValidaModoAtestado confere se o modo pode ser usado com o VCS escolhido. É a borda que
// impede uma combinação impossível (`vcs` sem VCS) antes de qualquer pergunta.
func ValidaModoAtestado(nomeVCS, modoAtestado string) error {
	modo := strings.TrimSpace(modoAtestado)
	if modo == "" {
		modo = AtestadoLocal
	}
	switch modo {
	case AtestadoDesligado, AtestadoLocal:
	case AtestadoVCS:
		if strings.TrimSpace(nomeVCS) == "nenhum" {
			return fmt.Errorf("modo do atestado %q não combina com VCS %q", AtestadoVCS, nomeVCS)
		}
		return ErrModoAdiado
	default:
		return fmt.Errorf("modo do atestado %q desconhecido", modo)
	}
	return nil
}