// Package bus é o barramento de jobs entre as sessões do pipeline (Discovery, Builder e Reviewer).
//
// Cada job é um arquivo JSON que muda de diretório conforme o estado (pending, processing, done,
// failed). A reivindicação é um os.Rename atômico, de modo que só um worker vence a disputa, e o
// lease permite recuperar jobs de um worker que caiu. O pacote só conhece caminhos do layout e
// não executa agentes nem fala com a rede.
package bus

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
)

// Versão e limites do contrato (SEG-01).
const (
	SchemaVersion    = 1
	MaxJobBytes      = 256 * 1024
	maxListItems     = 50
	maxTextBytes     = 2048
	maxEvents        = 100
	maxTitleLen      = 200
	DefaultLeaseTTL  = 15 * time.Minute
	orphanLeaseGrace = 5 * time.Minute
)

// Tipos de job.
const (
	TypeFeatureBuild = "feature_build"
	TypePRRework     = "pr_rework"
	TypePRReview     = "pr_review"
)

// Prioridades.
const (
	PriorityHigh   = "high"
	PriorityNormal = "normal"
)

// Erros do barramento.
var (
	ErrInvalidJob = errors.New("job inválido")
	ErrEmpty      = errors.New("fila vazia")
	ErrNotFound   = errors.New("job não encontrado")
	ErrNotOwner   = errors.New("o lease do job pertence a outro worker")
	ErrTooLarge   = errors.New("job acima do tamanho máximo")
)

var (
	idRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,80}$`)
	featureRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,80}$`)
	branchRe  = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)
	workerRe  = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
)

// Card liga o job ao item do gerenciador de projeto.
type Card struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
	URL      string `json:"url"`
}

// PR liga o job ao pull request. Number e URL são nulos até o Builder abrir o PR.
type PR struct {
	Number     *int    `json:"number"`
	URL        *string `json:"url"`
	Branch     string  `json:"branch"`
	BaseBranch string  `json:"base_branch"`
}

// Feedback é um item estruturado do rework pedido pelo Reviewer.
type Feedback struct {
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Fix         string `json:"fix,omitempty"`
}

// Lease marca quem reivindicou o job e até quando.
type Lease struct {
	Worker    string    `json:"worker"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Event é uma transição registrada no histórico do job.
type Event struct {
	At    time.Time `json:"at"`
	State string    `json:"state"`
	By    string    `json:"by,omitempty"`
	Note  string    `json:"note,omitempty"`
}

// Job é o contrato de trabalho entre as sessões (D-5).
type Job struct {
	SchemaVersion     int        `json:"schema_version"`
	ID                string     `json:"id"`
	Type              string     `json:"type"`
	Priority          string     `json:"priority"`
	Feature           string     `json:"feature"`
	Title             string     `json:"title"`
	Initiative        string     `json:"initiative,omitempty"`
	TaskIDs           []string   `json:"task_ids,omitempty"`
	SpecPath          string     `json:"spec_path"`
	ADRRefs           []string   `json:"adr_refs"`
	RulesRefs         []string   `json:"rules_refs"`
	Card              Card       `json:"card"`
	PR                PR         `json:"pr"`
	Attempt           int        `json:"attempt"`
	MaxReworkAttempts int        `json:"max_rework_attempts"`
	ReworkFeedback    []Feedback `json:"rework_feedback"`
	Lease             *Lease     `json:"lease,omitempty"`
	Events            []Event    `json:"events,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// RoleForType devolve o papel cuja fila recebe o tipo de job.
func RoleForType(t string) (string, bool) {
	switch t {
	case TypeFeatureBuild, TypePRRework:
		return layout.RoleBuilder, true
	case TypePRReview:
		return layout.RoleReviewer, true
	}
	return "", false
}

// Validate confere o job na borda, antes de gravar ou processar (SEG-01). role pode ser vazio
// para validar apenas o contrato.
func (j *Job) Validate(role string) error {
	bad := func(format string, a ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalidJob, fmt.Sprintf(format, a...))
	}
	if j.SchemaVersion != SchemaVersion {
		return bad("schema_version %d não suportada (esperado %d)", j.SchemaVersion, SchemaVersion)
	}
	if !idRe.MatchString(j.ID) {
		return bad("id %q fora do padrão ^[a-z0-9][a-z0-9._-]{0,80}$", j.ID)
	}
	want, ok := RoleForType(j.Type)
	if !ok {
		return bad("type %q desconhecido", j.Type)
	}
	if role != "" && role != want {
		return bad("type %q não pertence à fila %q (pertence a %q)", j.Type, role, want)
	}
	if j.Priority != PriorityHigh && j.Priority != PriorityNormal {
		return bad("priority %q inválida (use high ou normal)", j.Priority)
	}
	if j.Type == TypePRRework && j.Priority != PriorityHigh {
		return bad("pr_rework exige priority high")
	}
	if !featureRe.MatchString(j.Feature) {
		return bad("feature %q inválida", j.Feature)
	}
	if err := checkText("title", j.Title, maxTitleLen, false); err != nil {
		return bad("%v", err)
	}
	if j.Initiative != "" && !layout.ValidInitiative(j.Initiative) {
		return bad("initiative %q inválida", j.Initiative)
	}
	if err := checkRelPath("spec_path", j.SpecPath); err != nil {
		return bad("%v", err)
	}
	for field, list := range map[string][]string{"adr_refs": j.ADRRefs, "rules_refs": j.RulesRefs} {
		if len(list) > maxListItems {
			return bad("%s tem mais de %d itens", field, maxListItems)
		}
		for _, p := range list {
			if err := checkRelPath(field, p); err != nil {
				return bad("%v", err)
			}
		}
	}
	if len(j.TaskIDs) > maxListItems {
		return bad("task_ids tem mais de %d itens", maxListItems)
	}
	for _, id := range j.TaskIDs {
		if !idRe.MatchString(id) {
			return bad("task_ids: %q inválido", id)
		}
	}
	for field, v := range map[string]string{"card.provider": j.Card.Provider, "card.id": j.Card.ID, "card.url": j.Card.URL} {
		if err := checkText(field, v, 512, false); err != nil {
			return bad("%v", err)
		}
	}
	if !branchRe.MatchString(j.PR.Branch) || strings.HasPrefix(j.PR.Branch, "-") || strings.Contains(j.PR.Branch, "..") {
		return bad("pr.branch %q inválida", j.PR.Branch)
	}
	if !branchRe.MatchString(j.PR.BaseBranch) || strings.HasPrefix(j.PR.BaseBranch, "-") || strings.Contains(j.PR.BaseBranch, "..") {
		return bad("pr.base_branch %q inválida", j.PR.BaseBranch)
	}
	if j.PR.URL != nil {
		if err := checkText("pr.url", *j.PR.URL, 512, false); err != nil {
			return bad("%v", err)
		}
	}
	if j.PR.Number != nil && *j.PR.Number < 1 {
		return bad("pr.number %d inválido", *j.PR.Number)
	}
	if j.MaxReworkAttempts < 1 || j.MaxReworkAttempts > 10 {
		return bad("max_rework_attempts %d fora da faixa 1 a 10", j.MaxReworkAttempts)
	}
	if j.Attempt < 1 || j.Attempt > j.MaxReworkAttempts+1 {
		return bad("attempt %d fora da faixa 1 a %d", j.Attempt, j.MaxReworkAttempts+1)
	}
	if len(j.ReworkFeedback) > maxListItems {
		return bad("rework_feedback tem mais de %d itens", maxListItems)
	}
	for i, f := range j.ReworkFeedback {
		for field, v := range map[string]string{"file": f.File, "snippet": f.Snippet, "rule": f.Rule, "description": f.Description, "fix": f.Fix} {
			if err := checkText(fmt.Sprintf("rework_feedback[%d].%s", i, field), v, maxTextBytes, true); err != nil {
				return bad("%v", err)
			}
		}
		if f.Line < 0 {
			return bad("rework_feedback[%d].line negativa", i)
		}
	}
	return nil
}

// checkText recusa caracteres de controle e textos acima do limite; multiline libera \n e \t.
func checkText(field, v string, max int, multiline bool) error {
	if len(v) > max {
		return fmt.Errorf("%s acima de %d bytes", field, max)
	}
	for _, r := range v {
		if r == '\n' || r == '\t' {
			if multiline {
				continue
			}
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return fmt.Errorf("%s contém caractere de controle", field)
		}
	}
	return nil
}

// checkRelPath aceita só caminho relativo ao projeto, sem ".." nem controle (SEG-01).
func checkRelPath(field, p string) error {
	if p == "" {
		return fmt.Errorf("%s vazio", field)
	}
	if err := checkText(field, p, 512, false); err != nil {
		return err
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") || (len(p) > 1 && p[1] == ':') {
		return fmt.Errorf("%s %q deve ser relativo ao projeto", field, p)
	}
	for _, part := range strings.Split(strings.ReplaceAll(p, "\\", "/"), "/") {
		if part == ".." {
			return fmt.Errorf("%s %q sai do projeto", field, p)
		}
	}
	return nil
}

// ValidWorker informa se o nome do worker é seguro para uso em lease e em mensagens.
func ValidWorker(w string) bool { return workerRe.MatchString(w) }
