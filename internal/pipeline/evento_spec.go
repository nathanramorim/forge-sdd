package pipeline

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
)

// SpecAberta é o payload aditivo do evento `spec.aberta` (feat 113-05). Todos os campos já saem
// higienizados (SEG-08).
type SpecAberta struct {
	Numero string
	URL    string
	Titulo string
}

// EmissorSpecAberta é o gancho opcional do módulo ativo. nil = custo zero (internal/modulo ainda
// não existe na main; a ligação real fica para quando a 109 mesclar).
type EmissorSpecAberta func(SpecAberta) error

// EmiteSpecAberta emite o evento se houver emissor. Sem emissor faz zero chamadas; vcs.ErrSemPR
// significa "sem efeito, sem erro". Devolve se o evento foi entregue.
func EmiteSpecAberta(em EmissorSpecAberta, ev SpecAberta) (bool, error) {
	if em == nil {
		return false, nil
	}
	if err := em(ev); err != nil {
		if errors.Is(err, vcs.ErrSemPR) {
			return false, nil
		}
		return false, fmt.Errorf("evento spec.aberta: %s", higieniza(err.Error(), 200, false))
	}
	return true, nil
}

// cardOpcional lê a evidência quando existe (modo require_existing_card=false), sem bloquear.
// Devolve o card higienizado e um aviso quando a evidência é inválida ou vencida.
func cardOpcional(evPath, feature, initiative string, now time.Time, maxAgeMin int) (CardEvidence, string) {
	ev, ok := leEvidenciaExistente(evPath, feature)
	if !ok {
		ev, ok = leEvidenciaExistente(evPath, initiative)
	}
	if !ok {
		return CardEvidence{}, ""
	}
	if msg := validaEvidenciaParaGravar(ev); msg != "" {
		return CardEvidence{}, "evidência do card ignorada (inválida): o job segue sem card"
	}
	if u, err := url.Parse(ev.URL); err != nil || u.Scheme != "https" {
		return CardEvidence{}, "evidência do card ignorada (url não https): o job segue sem card"
	}
	at, _ := time.Parse(time.RFC3339, ev.CheckedAt)
	if now.Sub(at) > time.Duration(maxAgeMin)*time.Minute || at.Sub(now) > 5*time.Minute {
		return CardEvidence{}, "evidência do card ignorada (vencida): o job segue sem card"
	}
	ev.Provider = higieniza(ev.Provider, 64, false)
	ev.ID = higieniza(ev.ID, 64, false)
	ev.URL = higieniza(ev.URL, 512, false)
	return ev, ""
}
