package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/queue"
)

// Ações possíveis ao aplicar o veredito do Reviewer.
const (
	ActionApproved  = "approved"
	ActionRework    = "rework"
	ActionEscalated = "escalated"
)

const (
	maxFeedbackItems = 20
	maxFindingText   = 1500
)

// ReviewResult resume o que o veredito provocou.
type ReviewResult struct {
	Action          string
	AlreadyApplied  bool
	ReworkJob       *bus.Job
	Notification    *notify.Notification
	Channels        map[string]string
	CardInstruction string
	PRComment       string // comentário sugerido para o PR (quando reviewer.comment_on_pr está ligado)
}

// ApplyVerdict aplica a regra de fluxo ao veredito do modelo (D-3): conforme aprova e notifica;
// violação devolve ao Builder como pr_rework (prioridade alta, tentativa incrementada) até
// max_rework_attempts; no limite, ou em dúvida, escala para a pessoa. Aplicar duas vezes o mesmo
// veredito não repete efeitos.
func ApplyVerdict(dir, jobID, worker string, v queue.Verdict, deps notify.Deps, now time.Time) (*ReviewResult, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotPipelineMode
	}
	ncfg, err := config.LoadNotificationsConfig(dir)
	if err != nil {
		return nil, err
	}
	if err := validateVerdict(v); err != nil {
		return nil, err
	}

	entry, err := bus.Find(dir, layout.RoleReviewer, jobID)
	if err != nil {
		return nil, err
	}
	if entry.Job.Type != bus.TypePRReview {
		return nil, fmt.Errorf("%w: o job %s é do tipo %s, não pr_review", bus.ErrInvalidJob, jobID, entry.Job.Type)
	}
	if entry.State != layout.StateProcessing {
		if entry.State == layout.StateDone {
			return &ReviewResult{AlreadyApplied: true, Action: actionFromNote(entry.Job)}, nil
		}
		return nil, fmt.Errorf("o job %s está em %s; só jobs em processing recebem veredito", jobID, entry.State)
	}
	job := entry.Job
	if job.PR.URL == nil {
		return nil, fmt.Errorf("%w: o job %s não tem PR registrado", bus.ErrInvalidJob, jobID)
	}

	// A posse vem antes de qualquer efeito (notificar, enfileirar): sem ela, cada tentativa repetiria
	// webhook e e-mail antes de falhar em Complete.
	if err := bus.CheckOwner(dir, layout.RoleReviewer, jobID, worker); err != nil {
		return nil, err
	}

	res := &ReviewResult{}
	switch {
	case v.Veredito == "conforme":
		res.Action = ActionApproved
		res.CardInstruction = fmt.Sprintf("Atualize o card %s (%s) para %q via MCP.", job.Card.ID, job.Card.URL, cfg.PMMCP.StatusOnPRApproved)
		n := notify.Notification{
			Type: notify.TypePRApproved, Title: "PR pronto: " + job.Feature,
			Message: "PR revisado, sem inconformidades e pronto para aprovação final.\n" + oneLine(v.Resumo, 600),
			PRURL:   *job.PR.URL, CardURL: job.Card.URL, Feature: job.Feature, JobID: job.ID,
			ID: notificationID(job.ID, "aprovado"), // determinístico: reaplicar não duplica o inbox
		}
		out, err := notify.Dispatch(dir, ncfg, n, deps)
		if err != nil {
			return nil, err
		}
		res.Notification, res.Channels = &n, out.Channels
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, "aprovado: PR sem inconformidades", now); err != nil {
			return nil, err
		}

	case v.Veredito == "violacao" && job.Attempt < job.MaxReworkAttempts:
		res.Action = ActionRework
		feedback := toFeedback(v)
		rework := job
		rework.ID = reworkID(job.ID, job.Attempt+1)
		rework.Type = bus.TypePRRework
		rework.Priority = bus.PriorityHigh
		rework.Attempt = job.Attempt + 1
		rework.ReworkFeedback = feedback
		rework.Lease = nil
		rework.Events = nil
		rework.CreatedAt = time.Time{}
		if _, _, err := bus.Enqueue(dir, layout.RoleBuilder, rework, now); err != nil {
			return nil, fmt.Errorf("devolver ao Builder: %w", err)
		}
		res.ReworkJob = &rework
		if cfg.Reviewer.CommentOnPR {
			res.PRComment = FeedbackMarkdown(job, feedback)
		}
		note := fmt.Sprintf("não conforme: devolvido ao Builder (tentativa %d de %d)", rework.Attempt, job.MaxReworkAttempts)
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, note, now); err != nil {
			return nil, err
		}

	default: // violação no limite de tentativas, ou dúvida: a pessoa decide
		res.Action = ActionEscalated
		title := "Intervenção necessária: " + job.Feature
		msg := fmt.Sprintf("O ciclo automático de correção terminou sem aprovação (tentativa %d de %d, veredito %s). Revise o PR manualmente.\n%s",
			job.Attempt, job.MaxReworkAttempts, v.Veredito, oneLine(v.Resumo, 600))
		if v.Veredito == "duvida" {
			title = "Confirmação humana necessária: " + job.Feature
			msg = "O Reviewer ficou em dúvida e não devolveu o PR automaticamente. Confirme o que precisa ser decidido.\n" + oneLine(v.Resumo, 600)
		}
		n := notify.Notification{ID: notificationID(job.ID, "escalado"), Type: notify.TypeEscalation, Title: title, Message: msg, PRURL: *job.PR.URL, CardURL: job.Card.URL, Feature: job.Feature, JobID: job.ID}
		out, err := notify.Dispatch(dir, ncfg, n, deps)
		if err != nil {
			return nil, err
		}
		res.Notification, res.Channels = &n, out.Channels
		if cfg.Reviewer.CommentOnPR {
			res.PRComment = FeedbackMarkdown(job, toFeedback(v))
		}
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, "escalado para a pessoa: "+v.Veredito, now); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// validateVerdict aplica o contrato único do veredito (queue.ValidateVerdict, DUP-06) e as regras
// extras do pipeline: dúvida e violação exigem achado, e conforme não pode trazer violação.
func validateVerdict(v queue.Verdict) error {
	if err := queue.ValidateVerdict(v); err != nil {
		return err
	}
	violations := 0
	for _, f := range v.Achados {
		if f.Classificacao == queue.VerdictViolacao {
			violations++
		}
	}
	switch v.Veredito {
	case queue.VerdictConforme:
		if violations > 0 {
			return errors.New("veredito conforme não pode trazer achado classificado como violacao")
		}
	case queue.VerdictViolacao:
		if violations == 0 {
			return errors.New("veredito violacao exige ao menos um achado classificado como violacao")
		}
	default:
		if len(v.Achados) == 0 {
			return fmt.Errorf("veredito %s exige ao menos um achado", v.Veredito)
		}
	}
	return nil
}

// toFeedback converte os achados em feedback estruturado, sanitizado e limitado.
func toFeedback(v queue.Verdict) []bus.Feedback {
	var out []bus.Feedback
	for _, f := range v.Achados {
		if v.Veredito == "violacao" && f.Classificacao != "violacao" {
			continue
		}
		out = append(out, bus.Feedback{
			File:        notify.Sanitize(f.Arquivo, 400, false),
			Line:        max0(f.Linha),
			Snippet:     notify.Sanitize(f.Evidencia, 400, false),
			Rule:        notify.Sanitize(f.Regra, 100, false),
			Description: notify.Sanitize(f.Explicacao, maxFindingText, true),
			Fix:         notify.Sanitize(f.Correcao, maxFindingText, true),
		})
		if len(out) >= maxFeedbackItems {
			break
		}
	}
	return out
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func oneLine(s string, max int) string { return notify.Sanitize(s, max, false) }

func reworkID(jobID string, attempt int) string {
	base := strings.TrimSuffix(jobID, "-review")
	suffix := fmt.Sprintf("-rework-%d", attempt)
	if len(base)+len(suffix) > 80 {
		base = base[:80-len(suffix)]
	}
	return strings.TrimRight(base, "-.") + suffix
}

func actionFromNote(j bus.Job) string {
	if len(j.Events) == 0 {
		return ""
	}
	note := j.Events[len(j.Events)-1].Note
	switch {
	case strings.HasPrefix(note, "aprovado"):
		return ActionApproved
	case strings.HasPrefix(note, "não conforme"):
		return ActionRework
	case strings.HasPrefix(note, "escalado"):
		return ActionEscalated
	}
	return ""
}

// FeedbackMarkdown formata o feedback para um comentário no PR (publicado pelo agente com `gh`).
func FeedbackMarkdown(job bus.Job, items []bus.Feedback) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**Forge SDD — revisão automática** (tentativa %d de %d)\n\n", job.Attempt, job.MaxReworkAttempts)
	if len(items) == 0 {
		b.WriteString("Sem itens estruturados; veja o resumo da revisão.\n")
		return b.String()
	}
	for i, f := range items {
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		fmt.Fprintf(&b, "%d. `%s` em `%s`: %s\n", i+1, f.Rule, loc, f.Description)
		if f.Fix != "" {
			fmt.Fprintf(&b, "   - Ajuste esperado: %s\n", f.Fix)
		}
	}
	return b.String()
}

// notificationID deriva um ID determinístico e válido para o inbox (até 80 caracteres): job IDs
// longos são truncados e recebem um hash curto para não colidirem.
func notificationID(jobID, suffix string) string {
	id := jobID + "-" + suffix
	if len(id) <= 80 {
		return id
	}
	sum := sha256.Sum256([]byte(jobID))
	return jobID[:60] + "-" + hex.EncodeToString(sum[:4]) + "-" + suffix
}
