package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
)

// Erros da volta da spec à principal (RN-1, RN-3).
var (
	// ErrSpecSemConferencia1: a spec não pode voltar sem o parecer da Conferência 1.
	ErrSpecSemConferencia1 = errors.New("a spec não volta à principal sem a Conferência 1 aprovada")
	// ErrSpecSemSeguir: no modo com humano, a volta exige o [Seguir] do Gate 1.
	ErrSpecSemSeguir = errors.New("a spec não volta à principal sem o [Seguir] do humano no Gate 1")
	// ErrSpecSemAdaptador: não há adaptador de VCS para fazer a volta.
	ErrSpecSemAdaptador = errors.New("adaptador de VCS ausente")
)

// VoltaDaSpec reúne o estado necessário para decidir se a spec volta à principal.
type VoltaDaSpec struct {
	// Conferencia1Conforme diz se o Reviewer aprovou a spec no ponto `spec`.
	Conferencia1Conforme bool
	// Seguir é a resposta do humano no Gate 1 ([Seguir]). No modo sem humano, o chamador
	// preenche com a decisão automática já registrada.
	Seguir bool
	// ModoRevisao é `com-humano` ou `sem-humano` (campo da Constituição).
	ModoRevisao string
	// Sinal é o sinal gravado no status.md; ele precisa ser o da Conferência 1.
	Sinal string
}

// DecisaoVolta diz se a spec pode voltar e por quê.
type DecisaoVolta struct {
	Pode   bool
	Motivo string
}

// AutorizaVoltaDaSpec aplica RN-1/RN-3: a spec volta à principal só depois da Conferência 1
// conforme e do [Seguir]. No modo sem humano, o [Seguir] vem da decisão automática gravada, mas
// a Conferência 1 continua obrigatória — o modo sem humano dispensa a pergunta, não a revisão.
func AutorizaVoltaDaSpec(v VoltaDaSpec) DecisaoVolta {
	if !v.Conferencia1Conforme {
		return DecisaoVolta{Motivo: ErrSpecSemConferencia1.Error()}
	}
	if v.Sinal != "" && v.Sinal != SinalSpecParaRevisao {
		return DecisaoVolta{Motivo: fmt.Sprintf("o sinal gravado é %q, não %q", v.Sinal, SinalSpecParaRevisao)}
	}
	modo := strings.TrimSpace(v.ModoRevisao)
	if modo == "" {
		modo = "com-humano"
	}
	if modo == "com-humano" && !v.Seguir {
		return DecisaoVolta{Motivo: ErrSpecSemSeguir.Error()}
	}
	if modo != "com-humano" && modo != "sem-humano" {
		return DecisaoVolta{Motivo: fmt.Sprintf("modo de revisão humana %q desconhecido", modo)}
	}
	return DecisaoVolta{Pode: true, Motivo: "Conferência 1 conforme e Gate 1 respondido"}
}

// MesclaDaSpec é o efeito da volta: o que o adaptador fez.
type MesclaDaSpec struct {
	Mesclada bool
	Local    bool   // true quando o VCS é `nenhum` e o merge é local
	Aviso    string // a branch, no caso de `nenhum`
}

// VoltaSpecParaPrincipal faz a spec voltar à principal depois de autorizado. No `github`, mescla
// o PR da spec pelo adaptador; no `nenhum`, só informa que a branch está pronta para o merge
// local (nada é mesclado por conta própria, RN-25).
func VoltaSpecParaPrincipal(dir string, adapt vcs.Adaptador, mescla vcs.Mescla, decisao DecisaoVolta) (MesclaDaSpec, error) {
	if !decisao.Pode {
		if decisao.Motivo == "" {
			return MesclaDaSpec{}, ErrSpecSemConferencia1
		}
		// Preserva a sentinela na cadeia: quem chama testa com errors.Is.
		for _, sentinela := range []error{ErrSpecSemConferencia1, ErrSpecSemSeguir, ErrSpecSemAdaptador} {
			if decisao.Motivo == sentinela.Error() {
				return MesclaDaSpec{}, sentinela
			}
		}
		return MesclaDaSpec{}, errors.New(decisao.Motivo)
	}
	if adapt == nil {
		return MesclaDaSpec{}, ErrSpecSemAdaptador
	}
	if adapt.Nome() == "nenhum" {
		return MesclaDaSpec{
			Mesclada: false,
			Local:    true,
			Aviso:    fmt.Sprintf("VCS nenhum: a branch %q está pronta; o merge da spec é local", mescla.Branch),
		}, nil
	}
	if err := adapt.Mesclar(dir, mescla); err != nil {
		return MesclaDaSpec{}, fmt.Errorf("mesclar a spec no %s: %w", adapt.Nome(), err)
	}
	return MesclaDaSpec{Mesclada: true}, nil
}

// SinalAposVolta devolve o sinal a gravar quando a spec já voltou: nenhum, porque a conferência
// daquele ponto terminou. Gravar de novo faria o Reviewer rodar outra vez sobre spec mesclada.
func SinalAposVolta(em time.Time) Sinal { return Sinal{Ponto: SinalNenhum, Em: em} }
