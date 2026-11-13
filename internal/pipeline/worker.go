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
)

// Resultados de um job para o gancho de telemetria.
const (
	OutcomeDone   = "done"
	OutcomeFailed = "failed"
)

// WorkerOptions configura o laço do worker.
type WorkerOptions struct {
	Dir        string
	Role       string // builder ou reviewer
	Name       string // nome da sessão e do lease (forge-builder)
	Once       bool   // sai quando a fila esvazia
	Poll       time.Duration
	RenewEvery time.Duration
	Out        io.Writer
	Runner     Runner
	Now        func() time.Time
	// OnJobDone recebe o desfecho de cada job (telemetria); erros dele são ignorados.
	OnJobDone func(job bus.Job, outcome string)
}

// RunWorker é o laço do worker: recupera leases vencidos, reivindica o próximo job, executa o
// agente de forma não interativa e fecha o job conforme o resultado (D-1).
func RunWorker(ctx context.Context, opts WorkerOptions) error {
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
	if opts.Poll <= 0 {
		opts.Poll = 5 * time.Second
	}
	if opts.RenewEvery <= 0 {
		opts.RenewEvery = 5 * time.Minute
	}
	if opts.Runner == nil {
		profile, ok := cfg.CLIProfiles[cfg.CLI]
		if !ok {
			return fmt.Errorf("sem perfil de CLI para %q", cfg.CLI)
		}
		opts.Runner = ExecRunner{Profile: profile}
	}

	var lastIdleMsg time.Time
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
			if lastIdleMsg.IsZero() || opts.Now().Sub(lastIdleMsg) > time.Minute {
				fmt.Fprintf(opts.Out, "[%s] Aguardando jobs na fila...\n", opts.Name)
				lastIdleMsg = opts.Now()
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
		lastIdleMsg = time.Time{} // depois de trabalhar, a próxima ociosidade avisa de novo
		processJob(ctx, opts, cfg, entry)
	}
}

func processJob(ctx context.Context, opts WorkerOptions, cfg config.ForgeConfig, entry *bus.Entry) {
	job := entry.Job
	fmt.Fprintf(opts.Out, "[%s] job %s (%s, tentativa %d) iniciado\n", opts.Name, job.ID, job.Type, job.Attempt)

	logPath := filepath.Join(layout.Logs(opts.Dir), opts.Role+"-"+job.ID+".log")
	var out io.Writer = opts.Out
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err == nil {
		if lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
			defer lf.Close()
			out = io.MultiWriter(opts.Out, lf)
		}
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
		Agent:   opts.Role,
		Name:    opts.Name,
		Prompt:  BuildPrompt(opts.Role, job, filepath.ToSlash(rel)),
		Dir:     opts.Dir,
		Timeout: timeout,
	}, out)
	close(stop)
	wg.Wait()

	reason := ""
	switch {
	case runErr != nil:
		reason = "falha ao executar o agente: " + runErr.Error()
	case res.TimedOut:
		reason = fmt.Sprintf("timeout após %d minutos", cfg.Builder.JobTimeoutMinutes)
	case res.ExitCode != 0:
		reason = fmt.Sprintf("o agente saiu com código %d", res.ExitCode)
	}
	if reason == "" {
		reason = verifyHandoff(opts, entry.Job.ID)
	}

	outcome := OutcomeDone
	if reason != "" {
		outcome = OutcomeFailed
		fmt.Fprintf(opts.Out, "[%s] job %s falhou: %s\n", opts.Name, job.ID, reason)
		if _, err := bus.Fail(opts.Dir, opts.Role, job.ID, opts.Name, reason, opts.Now()); err != nil {
			fmt.Fprintf(opts.Out, "[%s] não foi possível marcar %s como falho: %v\n", opts.Name, job.ID, err)
		}
	} else {
		fmt.Fprintf(opts.Out, "[%s] job %s concluído\n", opts.Name, job.ID)
	}
	if opts.OnJobDone != nil {
		opts.OnJobDone(job, outcome)
	}
}

// verifyHandoff confere o que o agente precisava deixar feito e conclui o job. Devolve o motivo da
// falha, ou vazio se tudo certo.
func verifyHandoff(opts WorkerOptions, jobID string) string {
	e, err := bus.Find(opts.Dir, opts.Role, jobID)
	if err != nil {
		return "job desapareceu da fila: " + err.Error()
	}
	switch opts.Role {
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
		if _, _, err := bus.Enqueue(opts.Dir, layout.RoleReviewer, review, opts.Now()); err != nil {
			return "não foi possível enfileirar a revisão: " + err.Error()
		}
		note := fmt.Sprintf("PR #%d aberto; revisão enfileirada", *j.PR.Number)
		if _, err := bus.Complete(opts.Dir, opts.Role, jobID, opts.Name, note, opts.Now()); err != nil {
			return "não foi possível concluir o job: " + err.Error()
		}
	case layout.RoleReviewer:
		if e.State == layout.StateProcessing {
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
