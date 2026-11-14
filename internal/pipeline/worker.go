package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
)

// Resultados de um job para o gancho de telemetria.
const (
	OutcomeDone        = "done"
	OutcomeFailed      = "failed"
	OutcomeDeferred    = "deferred" // limite de uso: o job voltou à fila sem gastar tentativa
	OutcomeUnavailable = "unavailable"
)

// WorkerOptions configura o laço do worker.
type WorkerOptions struct {
	Dir  string
	Role string // builder ou reviewer
	Name string // nome da sessão e do lease (forge-builder)
	// Sleep espera uma duração e devolve false se o contexto acabar (injetável nos testes da pausa).
	Sleep func(context.Context, time.Duration) bool
	// Raw devolve a saída bruta do agente ao terminal (sem a visão limpa).
	Raw bool
	// WatchEvery é o intervalo do acompanhamento por git (padrão 5 s; negativo desliga).
	WatchEvery time.Duration
	// CLI sobrescreve o agente da sessão (forge-sdd up --builder codex sem gravar em forge.json).
	// Vazio: cfg.CLIFor(Role).
	CLI        string
	Once       bool // sai quando a fila esvazia
	Poll       time.Duration
	RenewEvery time.Duration
	Out        io.Writer
	Runner Runner
	// DeliverTarget é o alvo tmux (ex: =sessao:forge.1) para entrega de prompts no modo live (feat 77).
	DeliverTarget string
	// Deliverer executa operações no tmux para o modo live (injetável nos testes).
	Deliverer TmuxDeliverer
	Now       func() time.Time
	// OnJobDone recebe o desfecho de cada job (telemetria); erros dele são ignorados.
	OnJobDone func(job bus.Job, outcome string)
}

// RunWorker é o laço do worker: recupera leases vencidos, reivindica o próximo job, executa o
// agente de forma não interativa e fecha o job conforme o resultado (D-1).
func RunWorker(ctx context.Context, opts WorkerOptions) error {
	if opts.DeliverTarget != "" {
		return RunDeliverWorker(ctx, opts)
	}
	if opts.Role != layout.RoleBuilder && opts.Role != layout.RoleReviewer {
		return fmt.Errorf("papel %q inválido para o worker (use builder ou reviewer)", opts.Role)
	}
	cfg, found, err := config.LoadForgeConfig(opts.Dir)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotPipelineMode
	}
	if opts.Name == "" {
		opts.Name = "forge-" + opts.Role
	}
	if !bus.ValidWorker(opts.Name) {
		return fmt.Errorf("nome do worker %q inválido", opts.Name)
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Sleep == nil {
		opts.Sleep = realSleep
	}
	if opts.Poll <= 0 {
		opts.Poll = 5 * time.Second
	}
	if opts.RenewEvery <= 0 {
		opts.RenewEvery = 5 * time.Minute
	}
	if opts.CLI == "" {
		opts.CLI = cfg.CLIFor(opts.Role)
	}
	if !config.KnownCLI(opts.CLI) {
		return fmt.Errorf("agente %q inválido para o worker %s (use %s)", opts.CLI, opts.Role, strings.Join(config.OfferedCLIs(), ", "))
	}
	profile, ok := cfg.Profile(opts.CLI)
	if !ok {
		return fmt.Errorf("sem perfil de CLI para %q", opts.CLI)
	}
	if opts.Runner == nil {
		opts.Runner = ExecRunner{Profile: profile}
	}
	fmt.Fprintf(opts.Out, "[%s] agente: %s · permissões: %s\n", opts.Name, opts.CLI, config.PermissionSummary(cfg.Permissions, cfg.AllowCommands))

	var wasIdle bool
	for {
		if ctx.Err() != nil {
			return nil
		}
		if n, err := bus.Recover(opts.Dir, opts.Role, opts.Now()); err == nil && n > 0 {
			fmt.Fprintf(opts.Out, "[%s] %d job(s) com lease vencido devolvido(s) à fila\n", opts.Name, n)
		}
		entry, err := bus.Claim(opts.Dir, opts.Role, opts.Name, bus.DefaultLeaseTTL, opts.Now())
		if err == bus.ErrEmpty {
			if opts.Once {
				return nil
			}
			if !wasIdle {
				fmt.Fprintf(opts.Out, "[%s] Aguardando jobs na fila...\n", opts.Name)
				wasIdle = true
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(opts.Poll):
			}
			continue
		}
		if err != nil {
			return err
		}
		wasIdle = false // depois de trabalhar, a próxima ociosidade avisa de novo
		if err := processJob(ctx, opts, cfg, entry); err != nil {
			return err
		}
	}
}

func processJob(ctx context.Context, opts WorkerOptions, cfg config.ForgeConfig, entry *bus.Entry) error {
	job := entry.Job
	started := opts.Now()

	logPath := filepath.Join(layout.Logs(opts.Dir), opts.Role+"-"+job.ID+".log")
	var logw io.Writer
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err == nil {
		if lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
			defer lf.Close()
			logw = lf
		}
	}
	// O log guarda a saída bruta completa; o terminal recebe a visão limpa (ou a bruta com --raw).
	live := NewLiveView(opts.Out, logw, LiveOptions{CLI: opts.CLI, Raw: opts.Raw, Now: opts.Now})
	var out io.Writer = live
	live.Line(fmt.Sprintf("▶ job %s · %s · tentativa %d · %s · permissões %s", job.Feature, job.Type, job.Attempt, opts.CLI, cfg.Permissions))

	watcher := NewGitWatcher(opts.Dir, opts.Now, live.Activity)
	watchEvery := opts.WatchEvery
	if watchEvery == 0 {
		watchEvery = 5 * time.Second
	}
	watchCtx, stopWatch := context.WithCancel(ctx)
	defer stopWatch()
	watchDone := make(chan struct{})
	if watchEvery > 0 && !opts.Raw {
		go func() {
			defer close(watchDone)
			watcher.Run(watchCtx, watchEvery, live.Line)
		}()
	} else {
		close(watchDone)
	}

	// Renova o lease enquanto o agente trabalha (jobs longos não podem ser recuperados por engano).
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(opts.RenewEvery)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				_ = bus.Renew(opts.Dir, opts.Role, job.ID, opts.Name, bus.DefaultLeaseTTL, opts.Now())
			}
		}
	}()

	rel, _ := filepath.Rel(opts.Dir, entry.Path)
	timeout := time.Duration(cfg.Builder.JobTimeoutMinutes) * time.Minute
	res, runErr := opts.Runner.Run(ctx, RunSpec{
		Agent:       opts.Role,
		Name:        opts.Name,
		Prompt:      workerPrompt(opts.Role, job, filepath.ToSlash(rel), cfg),
		Permissions: cfg.WorkerPermissionArgs(opts.CLI),
		Dir:         opts.Dir,
		Timeout:     timeout,
	}, out)
	close(stop)
	wg.Wait()
	stopWatch()
	<-watchDone // o watcher termina antes do rodapé: nenhuma linha sai depois dele
	live.Flush()

	reason := ""
	switch {
	case runErr != nil:
		reason = "falha ao executar o agente: " + runErr.Error()
	case res.TimedOut:
		reason = fmt.Sprintf("timeout após %d minutos", cfg.Builder.JobTimeoutMinutes)
	case res.ExitCode != 0:
		reason = fmt.Sprintf("o agente saiu com código %d", res.ExitCode)
		// Limite de uso (só com saída diferente de zero): o job volta à fila sem gastar tentativa e o
		// worker espera até o horário informado.
		if ul, ok := DetectUsageLimit(live.Tail(8), opts.Now()); ok && !ul.Capped && ctx.Err() == nil && deferrals(job) < maxUsageLimitDeferrals {
			deferJob(ctx, opts, job, ul, live)
			if opts.OnJobDone != nil {
				opts.OnJobDone(job, OutcomeDeferred)
			}
			return nil
		}
		if au, ok := DetectAgentUnavailable(live.Tail(12), opts.CLI); ok {
			if _, reqErr := bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, "agente indisponível: "+opts.CLI, opts.Now()); reqErr == nil {
				if ncfg, err := config.LoadNotificationsConfig(opts.Dir); err == nil {
					n := notify.Notification{
						Type:    notify.TypeInfo,
						Title:   fmt.Sprintf("Agente %s indisponível no %s", opts.CLI, opts.Role),
						Message: fmt.Sprintf("%s\nAção: %s", au.Message, au.Hint),
						Feature: job.Feature, JobID: job.ID,
						ID: notificationID(job.ID, "agente-indisponivel"),
					}
					_, _ = notify.Dispatch(opts.Dir, ncfg, n, notify.Deps{})
				}
				live.Line(fmt.Sprintf("✗ %s indisponível: %s", opts.CLI, au.Message))
				live.Line(fmt.Sprintf("→ %s", au.Hint))
				if opts.OnJobDone != nil {
					opts.OnJobDone(job, OutcomeUnavailable)
				}
				return fmt.Errorf("%s indisponível: %s", opts.CLI, au.Message)
			}
			// Se o Requeue falhar, cai no bus.Fail de hoje com a mensagem
			reason += ": " + au.Message
		}
		if last := lastUsefulLine(live); last != "" && !strings.Contains(reason, last) {
			reason += ": " + last
		}
		if deferrals(job) >= maxUsageLimitDeferrals {
			reason += fmt.Sprintf(" (já foi adiado %d vezes por limite de uso)", deferrals(job))
		}
	}
	if reason == "" {
		reason = verifyHandoff(opts, entry.Job.ID)
	}

	outcome := OutcomeDone
	if reason != "" {
		outcome = OutcomeFailed
		live.Line(fmt.Sprintf("✗ falhou em %s: %s", FormatElapsed(opts.Now().Sub(started)), reason))
		if _, err := bus.Fail(opts.Dir, opts.Role, job.ID, opts.Name, reason, opts.Now()); err != nil {
			fmt.Fprintf(opts.Out, "[%s] não foi possível marcar %s como falho: %v\n", opts.Name, job.ID, err)
		}
	} else {
		live.Line(fmt.Sprintf("■ concluído em %s · %d arquivos alterados", FormatElapsed(opts.Now().Sub(started)), watcher.Count()))
	}
	if opts.OnJobDone != nil {
		opts.OnJobDone(job, outcome)
	}
	return nil
}

// verifyHandoff confere o que o agente precisava deixar feito e conclui o job. Devolve o motivo da
// falha, ou vazio se tudo certo.
func verifyHandoff(opts WorkerOptions, jobID string) string {
	return VerifyJobHandoff(opts.Dir, opts.Role, opts.Name, opts.Now, jobID)
}

// VerifyJobHandoff confere o que o agente precisava deixar feito e conclui o job. Devolve o motivo da
// falha, ou vazio se tudo certo.
func VerifyJobHandoff(dir, role, workerName string, now func() time.Time, jobID string) string {
	if now == nil {
		now = time.Now
	}
	e, err := bus.Find(dir, role, jobID)
	if err != nil {
		return "job desapareceu da fila: " + err.Error()
	}
	switch role {
	case layout.RoleBuilder:
		if e.State != layout.StateProcessing {
			return "job fora de processing: " + e.State
		}
		j := e.Job
		if j.PR.URL == nil || j.PR.Number == nil {
			return "o agente terminou sem registrar o PR (forge-sdd pipeline pr-opened)"
		}
		review := j
		review.ID = reviewJobID(j.ID)
		review.Type = bus.TypePRReview
		review.Priority = bus.PriorityNormal
		review.Lease = nil
		review.Events = nil
		review.ReworkFeedback = []bus.Feedback{}
		review.CreatedAt = time.Time{}
		if _, _, err := bus.Enqueue(dir, layout.RoleReviewer, review, now()); err != nil {
			return "não foi possível enfileirar a revisão: " + err.Error()
		}
		note := fmt.Sprintf("PR #%d aberto; revisão enfileirada", *j.PR.Number)
		if _, err := bus.Complete(dir, role, jobID, workerName, note, now()); err != nil {
			return "não foi possível concluir o job: " + err.Error()
		}
	case layout.RoleReviewer:
		// Só o veredito aplicado por `bus review` conta: um `bus complete` solto fecha o job sem
		// notificar nem devolver ao Builder.
		if e.State != layout.StateDone || actionFromNote(e.Job) == "" {
			return "o agente terminou sem aplicar o veredito (forge-sdd bus review)"
		}
	}
	return ""
}

func reviewJobID(id string) string {
	r := id + "-review"
	if len(r) > 80 {
		r = id[:80-len("-review")] + "-review"
	}
	return strings.TrimRight(r, "-.")
}

func workerPrompt(role string, job bus.Job, rel string, cfg config.ForgeConfig) string {
	p := BuildPrompt(role, job, rel)
	if role == layout.RoleBuilder {
		p += "\n" + BuilderHints(cfg.Builder, job)
	}
	return p
}

// lastUsefulLine devolve a última linha útil da saída do agente (saneada), para o motivo da falha.
func lastUsefulLine(live *LiveView) string {
	tail := live.Tail(5)
	for i := len(tail) - 1; i >= 0; i-- {
		l := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(tail[i]), "✗"))
		if l != "" && !strings.HasPrefix(l, "(hooks globais") && !strings.Contains(l, "skills duplicadas") {
			return l
		}
	}
	return ""
}

// maxUsageLimitDeferrals é o teto de devoluções por limite de uso de um mesmo job: passou disso, o
// job falha com o motivo em vez de pausar para sempre (por exemplo, um limite semanal).
const maxUsageLimitDeferrals = 3

func deferrals(job bus.Job) int {
	n := 0
	for _, e := range job.Events {
		if strings.HasPrefix(e.Note, "adiado:") {
			n++
		}
	}
	return n
}

// deferJob devolve o job à fila, avisa no inbox (uma vez por pausa) e espera até o horário.
func deferJob(ctx context.Context, opts WorkerOptions, job bus.Job, ul UsageLimit, live *LiveView) {
	now := opts.Now()
	ul.ResetAt = ul.ResetAt.In(now.Location()) // o horário mostrado é o seu, não o fuso da mensagem
	reason := "limite de uso do " + opts.CLI
	if _, err := bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, reason, now); err != nil {
		live.Line("✗ não foi possível devolver o job à fila: " + err.Error())
		return
	}
	if ncfg, err := config.LoadNotificationsConfig(opts.Dir); err == nil {
		n := notify.Notification{
			Type:    notify.TypeInfo,
			Title:   fmt.Sprintf("%s pausado: limite de uso do %s", opts.Name, opts.CLI),
			Message: fmt.Sprintf("O job %s voltou à fila sem gastar tentativa. O worker retoma às %s.\n%s", job.Feature, ul.ResetAt.Format("15:04"), ul.Message),
			Feature: job.Feature, JobID: job.ID,
			ID: notificationID(job.ID, fmt.Sprintf("pausa-%d", now.Unix()/60)),
		}
		_, _ = notify.Dispatch(opts.Dir, ncfg, n, notify.Deps{})
	}
	live.Line("⏸ " + ul.Message)
	if !pauseUntil(ctx, ul.ResetAt, opts.CLI, opts.Now, opts.Sleep, live.Line) {
		return
	}
	live.Line("▶ retomando")
}
