package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/vcs"
)

// Erros da entrega do leque (RN-22, RN-25).
var (
	// ErrLequeBaseNaoEhPrincipal: o leque parte da principal, não da branch da iniciativa.
	ErrLequeBaseNaoEhPrincipal = errors.New("o leque parte da principal; a branch da iniciativa não serve de base")
	// ErrJuncaoNaoEhEntrega: juntar no worktree principal é integração local, não entrega.
	ErrJuncaoNaoEhEntrega = errors.New("a junção é integração local; a entrega no github é o PR pelo adaptador")
	// ErrFeaturesSemPR: a onda tem features sem PR em draft.
	ErrFeaturesSemPR = errors.New("nenhuma feature da onda tem PR em draft")
)

// ContratoDoLeque diz de onde o leque parte e como cada task é entregue. A base é sempre a
// principal (RN-22): a junção no worktree principal é só integração, não substitui o PR.
type ContratoDoLeque struct {
	// Iniciativa identifica a onda.
	Iniciativa string
	// Base é a branch de onde cada worktree parte. Precisa ser a principal.
	Base string
	// Principal é a branch principal do projeto (`main` por padrão).
	Principal string
	// VCS é o nome do adaptador (`github` ou `nenhum`).
	VCS string
}

// Validar confere o contrato. No `github`, a base tem de ser a principal; no `nenhum`, qualquer
// branch segura serve (não há PR, o merge é local).
func (c ContratoDoLeque) Validar() error {
	if strings.TrimSpace(c.Iniciativa) == "" {
		return fmt.Errorf("iniciativa vazia")
	}
	principal := c.Principal
	if strings.TrimSpace(principal) == "" {
		principal = "main"
	}
	base := strings.TrimSpace(c.Base)
	if base == "" {
		return fmt.Errorf("base vazia: o leque precisa de uma branch de partida")
	}
	if c.VCS == "github" && base != principal {
		return fmt.Errorf("%w: base %q, principal %q", ErrLequeBaseNaoEhPrincipal, base, principal)
	}
	switch c.VCS {
	case "github", "nenhum", "":
	default:
		return fmt.Errorf("VCS %q não suportado pelo leque (use github ou nenhum)", c.VCS)
	}
	return nil
}

// BaseEfetiva devolve a base com o padrão aplicado.
func (c ContratoDoLeque) BaseEfetiva() string { return strings.TrimSpace(c.Base) }

// PrincipalEfetiva devolve a principal com o padrão aplicado.
func (c ContratoDoLeque) PrincipalEfetiva() string {
	if strings.TrimSpace(c.Principal) == "" {
		return "main"
	}
	return c.Principal
}

// JuntarSoNoMergeLocalDoNenhum confere se a junção pode acontecer no worktree principal.
// No `github`, a junção local **não** é a entrega: a entrega é o PR de cada feature. No
// `nenhum`, não há PR e a junção local é o merge de fato.
func JuntarSoNoMergeLocalDoNenhum(vcsNome string) error {
	switch strings.TrimSpace(vcsNome) {
	case "nenhum", "":
		return nil
	case "github":
		return fmt.Errorf("%w: no github, publique cada feature como PR em draft e mescle por lá; "+
			"a junção no worktree principal serve para conferir a integração, não para entregar", ErrJuncaoNaoEhEntrega)
	default:
		return fmt.Errorf("VCS %q desconhecido", vcsNome)
	}
}

// PublicarAbrePRDraft faz a entrega do leque: cada feature da onda recebe o seu PR em draft, pelo
// adaptador, com base na principal. Devolve o PR de cada task e as que falharam.
type PublicacaoDoLeque struct {
	Task      string  `json:"task"`
	Branch    string  `json:"branch"`
	PR        *vcs.PR `json:"pr,omitempty"`
	JaExistia bool    `json:"ja_existia,omitempty"`
	Erro      string  `json:"erro,omitempty"`
}

// PublicarAbrePRDraft abre um PR em draft por feature do leque. Cada PR é independente: um erro
// numa task não impede as outras, e o resultado volta com o motivo de cada falha.
func PublicarAbrePRDraft(dir, iniciativa string, tasks []string, adapt vcs.Adaptador, principal string) ([]PublicacaoDoLeque, error) {
	if adapt == nil {
		return nil, fmt.Errorf("adaptador de VCS ausente")
	}
	// Publicar é abrir PR: só faz sentido no github. No `nenhum` não há VCS remoto e a
	// entrega é o merge local; recusar aqui evita prometer um PR que não existe.
	if adapt.Nome() != "github" {
		return nil, fmt.Errorf("%w: o VCS %q não tem PR; a entrega é o merge local (RN-25)",
			ErrJuncaoNaoEhEntrega, adapt.Nome())
	}
	if strings.TrimSpace(principal) == "" {
		principal = "main"
	}
	if len(tasks) == 0 {
		return nil, ErrFeaturesSemPR
	}

	var out []PublicacaoDoLeque
	var semPR int
	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if task == "" {
			continue
		}
		branch, err := BranchDaFeature(iniciativa, task)
		if err != nil {
			out = append(out, PublicacaoDoLeque{Task: task, Erro: err.Error()})
			semPR++
			continue
		}
		c := ContratoDePRPorFeature{Branch: branch, Base: principal, Principal: principal}
		if err := c.Validar(); err != nil {
			out = append(out, PublicacaoDoLeque{Task: task, Branch: branch, Erro: err.Error()})
			semPR++
			continue
		}
		existente, err := adapt.ProcurarPRAberto(dir, branch)
		if err != nil {
			out = append(out, PublicacaoDoLeque{Task: task, Branch: branch, Erro: err.Error()})
			semPR++
			continue
		}
		if existente != nil {
			out = append(out, PublicacaoDoLeque{Task: task, Branch: branch, PR: existente, JaExistia: true})
			continue
		}
		pr, err := adapt.CriarPRDraft(dir, vcs.NovoPR{Base: principal, Branch: branch, Titulo: "feat(" + iniciativa + "): " + task})
		if err != nil {
			out = append(out, PublicacaoDoLeque{Task: task, Branch: branch, Erro: err.Error()})
			semPR++
			continue
		}
		out = append(out, PublicacaoDoLeque{Task: task, Branch: branch, PR: &pr})
	}
	if semPR == len(out) && semPR > 0 {
		return out, fmt.Errorf("%w: as %d features falharam ao abrir PR", ErrFeaturesSemPR, semPR)
	}
	return out, nil
}
