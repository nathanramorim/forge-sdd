package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/vcs"
)

// Erros da abertura de PR por feature (RN-22, RN-7).
var (
	// ErrFeatureSemPRDraft: a feature não avança sem o PR dela em draft.
	ErrFeatureSemPRDraft = errors.New("a feature não avança sem o PR em draft da própria branch")
	// ErrBaseDaFeatureInvalida: a branch da feature precisa nascer da principal.
	ErrBaseDaFeatureInvalida = errors.New("a base do PR da feature precisa ser a principal, nunca a branch da iniciativa")
	// ErrFeatureJaTemPR: já existe PR aberto para a branch; o caso normal é reaproveitar.
	ErrFeatureJaTemPR = errors.New("a branch da feature já tem PR aberto")
)

// BranchDaFeature devolve o nome da branch de uma feature do split: `<prefixo>/<id>-<nome>-<sid>`
// (RN-22). O prefixo vem do tipo da task (`feat` ou `fix`), para a branch continuar legível.
func BranchDaFeature(iniciativa, sid string) (string, error) {
	ini := strings.TrimSpace(iniciativa)
	sid = strings.TrimSpace(sid)
	if ini == "" || sid == "" {
		return "", fmt.Errorf("iniciativa e sid são obrigatórios para nomear a branch da feature")
	}
	// A iniciativa já carrega o prefixo (`106-pr-draft-revisao-e-ci`); o sid completa o nome.
	return "feat/" + ini + "-" + sid, nil
}

// ContratoDePRPorFeature descreve o que uma feature precisa ter para avançar: uma branch própria,
// um PR em draft e a base na principal. É o contrato que o Builder cumpre ao abrir a entrega.
type ContratoDePRPorFeature struct {
	// Branch é a branch da feature (`feat/<id>-<nome>-<sid>`).
	Branch string
	// Base é a branch de destino. Sempre a principal.
	Base string
	// Principal é a branch principal do projeto (`main` por padrão).
	Principal string
}

// Validar confere o contrato antes de abrir o PR. Uma base diferente da principal é recusada:
// PR de feature com base na branch da iniciativa é exatamente o que a RN-22 proíbe.
func (c ContratoDePRPorFeature) Validar() error {
	if strings.TrimSpace(c.Branch) == "" {
		return fmt.Errorf("%w: branch vazia", ErrFeatureSemPRDraft)
	}
	if strings.TrimSpace(c.Base) == "" {
		return fmt.Errorf("%w: base vazia", ErrBaseDaFeatureInvalida)
	}
	principal := c.Principal
	if strings.TrimSpace(principal) == "" {
		principal = "main"
	}
	if c.Base != principal {
		return fmt.Errorf("%w: base %q, principal %q", ErrBaseDaFeatureInvalida, c.Base, principal)
	}
	// A branch da feature não pode ser a branch da iniciativa: ela precisa do próprio nome.
	if !strings.Contains(strings.TrimPrefix(strings.TrimPrefix(c.Branch, "feat/"), "fix/"), "-") {
		return fmt.Errorf("%w: %q não parece uma branch de feature", ErrFeatureSemPRDraft, c.Branch)
	}
	return nil
}

// AbrirPRDaFeature garante que a feature tem o PR dela em draft. Se já existe, devolve o PR sem
// criar outro (o caso normal, RN-22). Se o adaptador não tem PR (`nenhum`), devolve o aviso em vez
// de erro: a branch fica pronta.
func AbrirPRDaFeature(dir string, adapt vcs.Adaptador, c ContratoDePRPorFeature, p vcs.NovoPR) (*vcs.PR, error) {
	if err := c.Validar(); err != nil {
		return nil, err
	}
	if adapt == nil {
		return nil, fmt.Errorf("adaptador de VCS ausente")
	}
	if adapt.Nome() == "nenhum" {
		return nil, fmt.Errorf("%w: a branch %q fica pronta para o merge local", vcs.ErrSemPR, c.Branch)
	}
	existente, err := adapt.ProcurarPRAberto(dir, c.Branch)
	if err != nil {
		return nil, fmt.Errorf("procurar PR aberto de %s: %w", c.Branch, err)
	}
	if existente != nil {
		// Já existe: o PR é reaproveitado e a base é corrigida se estiver errada.
		if existente.Base != "" && existente.Base != c.Principal {
			if err := adapt.CorrigirBase(dir, existente.Numero, c.Principal); err != nil {
				return nil, fmt.Errorf("corrigir a base do PR #%d: %w", existente.Numero, err)
			}
			existente.Base = c.Principal
		}
		return existente, nil
	}
	p.Base = c.Principal
	p.Branch = c.Branch
	pr, err := adapt.CriarPRDraft(dir, p)
	if err != nil {
		return nil, fmt.Errorf("abrir o PR em draft de %s: %w", c.Branch, err)
	}
	return &pr, nil
}

// FeixesParalelos organiza as features de uma onda em feixes de PR independentes: cada feature tem
// a sua branch e o seu PR, todos com base na principal. Nenhum PR tem base em outro.
func FeixesParalelos(iniciativa string, sids []string, principal string) ([]ContratoDePRPorFeature, error) {
	if strings.TrimSpace(principal) == "" {
		principal = "main"
	}
	vistos := map[string]bool{}
	var out []ContratoDePRPorFeature
	for _, sid := range sids {
		sid = strings.TrimSpace(sid)
		if sid == "" || vistos[sid] {
			continue
		}
		vistos[sid] = true
		branch, err := BranchDaFeature(iniciativa, sid)
		if err != nil {
			return nil, err
		}
		c := ContratoDePRPorFeature{Branch: branch, Base: principal, Principal: principal}
		if err := c.Validar(); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, errors.New("nenhuma feature para organizar em feixes")
	}
	return out, nil
}
