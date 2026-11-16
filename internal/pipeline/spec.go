package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/queue"
)

// Nome do parecer da spec, gravado pelo Go na pasta da iniciativa (o Reviewer é somente leitura).
const specReviewFile = "review-spec.md"

// specJobID deriva o ID do job spec_review a partir do ID do despacho.
func specJobID(id string) string {
	r := id + "-spec"
	if len(r) > 80 {
		r = id[:80-len("-spec")] + "-spec"
	}
	return strings.TrimRight(r, "-.")
}

// nextSpecAttempt devolve a tentativa do próximo spec_review da iniciativa: 1 na primeira, e
// tentativa anterior + 1 quando o último parecer pediu ajuste. Limitada a max+1 (faixa do contrato).
func nextSpecAttempt(dir, initiative string, max int) int {
	entries, err := bus.List(dir, layout.RoleReviewer, layout.StateDone)
	if err != nil {
		return 1
	}
	attempt := 0
	for _, e := range entries {
		j := e.Job
		if j.Type != bus.TypeSpecReview || j.Initiative != initiative || len(j.Events) == 0 {
			continue
		}
		if !strings.HasPrefix(j.Events[len(j.Events)-1].Note, "spec: ajustar") {
			continue
		}
		if j.Attempt+1 > attempt {
			attempt = j.Attempt + 1
		}
	}
	if attempt == 0 {
		return 1
	}
	if attempt > max+1 {
		return max + 1
	}
	return attempt
}

// SpecResult resume o efeito do parecer sobre a spec.
type SpecResult struct {
	Action         string
	AlreadyApplied bool
	BuildJob       *bus.Job
	Notification   *notify.Notification
	Channels       map[string]string
	Warning        string // aviso quando os TaskIDs do job não casam com nenhuma task da fila
}

// ApplySpecVerdict aplica o parecer do Reviewer sobre a spec (ADR-0013). O Go decide e grava, nunca
// o agente (SEG-06): conforme (aprovado) libera a fila enfileirando o feature_build e grava
// review-spec.md; violacao (ajustar) devolve o feedback à Discovery pelo inbox até
// max_rework_attempts; duvida (escalar) ou o limite bloqueiam as tasks e avisam a pessoa.
func ApplySpecVerdict(dir, jobID, worker string, v queue.Verdict, deps notify.Deps, now time.Time) (*SpecResult, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotPipelineMode
	}
	if !cfg.IsAdvanced() {
		return nil, fmt.Errorf("a revisão de spec só vale no modo avançado (autonomy: %s)", config.AutonomyAdvanced)
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
	if entry.Job.Type != bus.TypeSpecReview {
		return nil, fmt.Errorf("%w: o job %s é do tipo %s, não spec_review", bus.ErrInvalidJob, jobID, entry.Job.Type)
	}
	if entry.State != layout.StateProcessing {
		if entry.State == layout.StateDone {
			return &SpecResult{AlreadyApplied: true, Action: actionFromNote(entry.Job)}, nil
		}
		return nil, fmt.Errorf("o job %s está em %s; só jobs em processing recebem parecer", jobID, entry.State)
	}
	job := entry.Job
	if err := bus.CheckOwner(dir, layout.RoleReviewer, jobID, worker); err != nil {
		return nil, err
	}

	persistVerdict(dir, job, v)
	if err := writeSpecReview(dir, job, v, now); err != nil {
		return nil, err
	}

	res := &SpecResult{}
	switch {
	case v.Veredito == queue.VerdictConforme:
		res.Action = ActionApproved
		build := job
		build.ID = strings.TrimSuffix(job.ID, "-spec")
		build.Type = bus.TypeFeatureBuild
		build.Priority = bus.PriorityNormal
		build.Attempt = 1
		build.ReworkFeedback = []bus.Feedback{}
		build.Lease, build.Events, build.CreatedAt = nil, nil, time.Time{}
		if _, _, err := bus.Enqueue(dir, layout.RoleBuilder, build, now); err != nil {
			return nil, fmt.Errorf("liberar a fila ao Builder: %w", err)
		}
		res.BuildJob = &build
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, "spec: aprovada; fila liberada ao Builder", now); err != nil {
			return nil, err
		}

	case v.Veredito == queue.VerdictViolacao && job.Attempt < job.MaxReworkAttempts:
		res.Action = ActionRework
		n := notify.Notification{
			ID: notificationID(job.ID, "ajustar"), Type: notify.TypeInfo,
			Title:   "Ajuste a spec: " + job.Feature,
			Message: fmt.Sprintf("O Reviewer pediu ajustes na spec (tentativa %d de %d). Veja %s e rode forge dispatch de novo.\n%s", job.Attempt, job.MaxReworkAttempts, specReviewRel(job), oneLine(v.Resumo, 600)),
			Feature: job.Feature, JobID: job.ID,
		}
		out, err := notify.Dispatch(dir, ncfg, n, deps)
		if err != nil {
			return nil, err
		}
		res.Notification, res.Channels = &n, out.Channels
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, fmt.Sprintf("spec: ajustar (tentativa %d de %d)", job.Attempt, job.MaxReworkAttempts), now); err != nil {
			return nil, err
		}

	default: // escalar (dúvida) ou limite de tentativas: bloqueia e avisa; as independentes seguem
		res.Action = ActionEscalated
		reason := fmt.Sprintf("spec não aprovada (veredito %s, tentativa %d de %d)", v.Veredito, job.Attempt, job.MaxReworkAttempts)
		if _, err := queue.BlockTasks(dir, job.Initiative, job.TaskIDs, reason); err != nil {
			return nil, fmt.Errorf("bloquear as tasks: %w", err)
		}
		res.Warning = unmatchedWarning(dir, job)
		n := notify.Notification{
			ID: notificationID(job.ID, "escalado"), Type: notify.TypeEscalation,
			Title:   "Spec bloqueada: " + job.Feature,
			Message: reason + ". Decida manualmente; veja " + specReviewRel(job) + ".\n" + oneLine(v.Resumo, 600),
			Feature: job.Feature, JobID: job.ID,
		}
		out, err := notify.Dispatch(dir, ncfg, n, deps)
		if err != nil {
			return nil, err
		}
		res.Notification, res.Channels = &n, out.Channels
		if _, err := bus.Complete(dir, layout.RoleReviewer, jobID, worker, "spec: escalada para a pessoa: "+v.Veredito, now); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func specReviewRel(job bus.Job) string {
	return filepath.ToSlash(filepath.Join("sdd", "discovery", job.Initiative, specReviewFile))
}

// writeSpecReview grava review-spec.md na pasta da iniciativa. Sem iniciativa válida não há onde
// gravar e o fluxo segue; falha real de escrita é erro.
func writeSpecReview(dir string, job bus.Job, v queue.Verdict, now time.Time) error {
	if job.Initiative == "" {
		return nil
	}
	base, err := layout.Initiative(dir, job.Initiative)
	if err != nil {
		return nil
	}
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Parecer da spec — %s\n\n", job.Feature)
	fmt.Fprintf(&b, "- Job: `%s` (tentativa %d de %d)\n- Veredito: `%s`\n- Data: %s\n\n", job.ID, job.Attempt, job.MaxReworkAttempts, v.Veredito, now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "## Resumo\n%s\n", oneLine(v.Resumo, 1500))
	if items := toFeedback(v); len(items) > 0 {
		b.WriteString("\n## Achados\n")
		for i, f := range items {
			fmt.Fprintf(&b, "%d. `%s` em `%s`: %s\n", i+1, f.Rule, f.File, f.Description)
			if f.Fix != "" {
				fmt.Fprintf(&b, "   - Ajuste esperado: %s\n", f.Fix)
			}
		}
	}
	return os.WriteFile(filepath.Join(base, specReviewFile), []byte(b.String()), 0o644)
}
