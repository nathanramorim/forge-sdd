package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
)

// Erros do modo de revisão humana (FLX-01, feat-106-17).
var (
	// ErrSemHumanoSemAviso: o modo sem humano só vale depois do aviso de riscos aceito.
	ErrSemHumanoSemAviso = errors.New("modo sem humano exige o aviso de riscos aceito pelo dono")
	// ErrModoRevisaoInvalido: valor fora do enum.
	ErrModoRevisaoInvalido = errors.New("modo de revisão humana inválido")
)

// EscolhaDeModo é o que o dono respondeu no início da sessão.
type EscolhaDeModo struct {
	// Modo é `com-humano` (padrão) ou `sem-humano`.
	Modo string
	// AvisoApresentado diz se o aviso de riscos foi lido para o dono.
	AvisoApresentado bool
	// Aceito diz se o dono aceitou os riscos depois do aviso.
	Aceito bool
	// RegistradoEm carimba a aceitação.
	RegistradoEm time.Time
}

// RiscosDoModoSemHumano são os riscos que precisam ser ditos antes de ligar o modo. O texto
// é fixo porque é ele que precisa ser lido — não pode ser resumido por quem liga o modo.
var RiscosDoModoSemHumano = []string{
	"nenhuma pessoa vê o código antes de o PR sair do draft;",
	"erro do Reviewer passa direto para a principal;",
	"o custo do laço fica sem teto humano.",
}

// AvisoDoModoSemHumano devolve o texto apresentado ao dono. Ele é a fonte única das três
// frases; nenhum comando repete o texto por conta própria.
func AvisoDoModoSemHumano() string {
	var b strings.Builder
	b.WriteString("Ligar o modo sem humano tem estes riscos:\n")
	for i, r := range RiscosDoModoSemHumano {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, r)
	}
	b.WriteString("Só ligue se aceitar os três. O modo volta a com-humano na próxima sessão.")
	return b.String()
}

// DecisaoDeModo é o modo efetivo e a razão.
type DecisaoDeModo struct {
	Modo       string
	Motivo     string
	Registrar  bool // true quando a escolha precisa ser gravada no status.md
	AceitoEm   time.Time
	ComAvisado bool
}

// ResolveModoDeRevisao decide o modo efetivo do início da sessão. `com-humano` é o padrão e
// nunca pede aviso. `sem-humano` só vale com o aviso apresentado e aceito; sem isso, recusa e
// mantém com-humano — o padrão é sempre o seguro.
func ResolveModoDeRevisao(e EscolhaDeModo) (DecisaoDeModo, error) {
	modo := strings.TrimSpace(e.Modo)
	if modo == "" {
		modo = ModoComHumano
	}
	switch modo {
	case ModoComHumano:
		return DecisaoDeModo{Modo: ModoComHumano, Motivo: "modo com humano (padrão)"}, nil
	case ModoSemHumano:
		if !e.AvisoApresentado {
			return DecisaoDeModo{Modo: ModoComHumano}, fmt.Errorf("%w: o aviso de riscos não foi apresentado", ErrSemHumanoSemAviso)
		}
		if !e.Aceito {
			return DecisaoDeModo{Modo: ModoComHumano}, fmt.Errorf("%w: o dono não aceitou os riscos", ErrSemHumanoSemAviso)
		}
		em := e.RegistradoEm
		if em.IsZero() {
			em = time.Now().UTC()
		}
		return DecisaoDeModo{
			Modo:       ModoSemHumano,
			Motivo:     "modo sem humano aceito depois do aviso de riscos",
			Registrar:  true,
			AceitoEm:   em,
			ComAvisado: true,
		}, nil
	default:
		return DecisaoDeModo{Modo: ModoComHumano},
			fmt.Errorf("%w: %q (use com-humano ou sem-humano)", ErrModoRevisaoInvalido, modo)
	}
}

// Constantes de modo, espelhando os valores da Constituição.
const (
	ModoComHumano = "com-humano"
	ModoSemHumano = "sem-humano"
)

// LeModoDaConstituicao lê `Modo de revisão humana` pelos modos já validados. É só um atalho
// com o padrão aplicado, para o fluxo não repetir a regra.
func LeModoDaConstituicao(m vcs.Modos) string {
	modo := strings.TrimSpace(m.RevisaoHumana)
	if modo == "" {
		return ModoComHumano
	}
	return modo
}
