package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/notify"
)

// TmuxDeliverer opera buffers, teclas e captura de tela no tmux para o despachante (feat 77).
type TmuxDeliverer interface {
	CapturePane(target string) (string, error)
	LoadBuffer(name string, data []byte) error
	PasteBuffer(name, target string) error
	SendKeys(target string, keys ...string) error
}

type execTmuxDeliverer struct{}

func (execTmuxDeliverer) CapturePane(target string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", target)
	out, err := cmd.Output()
	return string(out), err
}

func (execTmuxDeliverer) LoadBuffer(name string, data []byte) error {
	cmd := exec.Command("tmux", "load-buffer", "-b", name, "-")
	cmd.Stdin = bytes.NewReader(data)
	return cmd.Run()
}

func (execTmuxDeliverer) PasteBuffer(name, target string) error {
	return exec.Command("tmux", "paste-buffer", "-p", "-d", "-b", name, "-t", target).Run()
}

func (execTmuxDeliverer) SendKeys(target string, keys ...string) error {
	args := append([]string{"send-keys", "-t", target}, keys...)
	return exec.Command("tmux", args...).Run()
}

// DefaultTmuxDeliverer devolve o executor de entrega real do tmux.
func DefaultTmuxDeliverer() TmuxDeliverer {
	return execTmuxDeliverer{}
}

// waitPaneIdle espera o painel ficar estável (conteúdo inalterado por quietDuration).
func waitPaneIdle(ctx context.Context, d TmuxDeliverer, target string, quietDuration, timeout time.Duration, sleep func(context.Context, time.Duration) bool, now func() time.Time) (bool, error) {
	started := now()
	lastText := ""
	lastChange := now()

	for {
		txt, err := d.CapturePane(target)
		if err != nil {
			return false, err
		}
		cur := now()
		if txt != lastText {
			lastText = txt
			lastChange = cur
		} else {
			if cur.Sub(lastChange) >= quietDuration {
				return true, nil
			}
		}
		if timeout > 0 && cur.Sub(started) >= timeout {
			return false, nil
		}
		if sleep != nil {
			if !sleep(ctx, 200*time.Millisecond) {
				return false, ctx.Err()
			}
		} else {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
	}
}

// deliverBuffer cola os dados em target via load-buffer + paste-buffer -p + Enter.
func deliverBuffer(d TmuxDeliverer, target string, bufName string, data []byte) error {
	if err := d.LoadBuffer(bufName, data); err != nil {
		return fmt.Errorf("load-buffer: %w", err)
	}
	if err := d.PasteBuffer(bufName, target); err != nil {
		return fmt.Errorf("paste-buffer: %w", err)
	}
	if err := d.SendKeys(target, "Enter"); err != nil {
		return fmt.Errorf("send-keys Enter: %w", err)
	}
	return nil
}

// deliverJobPrompt cola o prompt no painel e aguarda o agente aceitar (mudança no conteúdo).
func deliverJobPrompt(ctx context.Context, d TmuxDeliverer, target string, jobID string, prompt string, sleep func(context.Context, time.Duration) bool, now func() time.Time, timeoutAccept time.Duration) (bool, error) {
	bufName := fmt.Sprintf("forge-%s-%d", jobID, now().UnixNano())
	initialText, _ := d.CapturePane(target)

	if err := deliverBuffer(d, target, bufName, []byte(prompt)); err != nil {
		return false, err
	}

	if timeoutAccept <= 0 {
		timeoutAccept = 30 * time.Second
	}

	started := now()
	retried := false

	for {
		if sleep != nil {
			if !sleep(ctx, 200*time.Millisecond) {
				return false, ctx.Err()
			}
		} else {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
		txt, _ := d.CapturePane(target)
		if txt != initialText {
			return true, nil
		}
		if now().Sub(started) >= timeoutAccept {
			if !retried {
				retried = true
				started = now()
				_ = d.SendKeys(target, "Enter")
				continue
			}
			return false, nil
		}
	}
}

// deliverContextReset envia o comando de limpeza de contexto e aguarda o painel ficar ocioso.
func deliverContextReset(ctx context.Context, d TmuxDeliverer, target string, jobID string, resetCmd string, quietDuration time.Duration, sleep func(context.Context, time.Duration) bool, now func() time.Time) error {
	bufName := fmt.Sprintf("forge-reset-%s-%d", jobID, now().UnixNano())
	if err := deliverBuffer(d, target, bufName, []byte(resetCmd)); err != nil {
		return err
	}
	_, err := waitPaneIdle(ctx, d, target, quietDuration, 10*time.Second, sleep, now)
	return err
}

// RunDeliverWorker opera o laço do despachante no modo live (feat 77).
func RunDeliverWorker(ctx context.Context, opts WorkerOptions) error {
	cfg, _, err := config.LoadForgeConfig(opts.Dir)
	if err != nil {
		return err
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
	d := opts.Deliverer
	if d == nil {
		d = DefaultTmuxDeliverer()
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

	profile, ok := cfg.Profile(opts.CLI)
	if !ok {
		return fmt.Errorf("sem perfil de CLI para %q", opts.CLI)
	}

	target := opts.DeliverTarget
	quietSec := cfg.Live.ReadyQuietSeconds
	if quietSec <= 0 {
		quietSec = 3
	}
	quietDuration := time.Duration(quietSec) * time.Second

	startupSec := cfg.Live.StartupWaitSeconds
	if startupSec < 0 {
		startupSec = 0
	}
	startupDuration := time.Duration(startupSec) * time.Second

	firstJob := true
	warnedNoReset := false

	fmt.Fprintf(opts.Out, "[%s] despachante live para %s · agente: %s\n", opts.Name, target, opts.CLI)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if _, err := bus.Recover(opts.Dir, opts.Role, opts.Now()); err != nil {
			fmt.Fprintf(opts.Out, "[%s] aviso ao recuperar leases: %v\n", opts.Name, err)
		}

		entry, err := bus.Claim(opts.Dir, opts.Role, opts.Name, bus.DefaultLeaseTTL, opts.Now())
		if err == bus.ErrEmpty {
			if opts.Once {
				return nil
			}
			if !opts.Sleep(ctx, opts.Poll) {
				return nil
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("claim: %w", err)
		}

		job := entry.Job
		fmt.Fprintf(opts.Out, "[%s] aguardando painel pronto (%d s parado) …\n", opts.Name, quietSec)

		if firstJob {
			if startupDuration > 0 {
				if !opts.Sleep(ctx, startupDuration) {
					_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, "cancelado durante inicialização", opts.Now())
					return nil
				}
			}
			firstJob = false
		}

		idle, err := waitPaneIdle(ctx, d, target, quietDuration, 30*time.Second, opts.Sleep, opts.Now)
		if err != nil || !idle {
			msg := "painel ocupado ou não respondeu"
			if err != nil {
				msg = err.Error()
			}
			fmt.Fprintf(opts.Out, "[%s] ✗ %s\n", opts.Name, msg)
			_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, msg, opts.Now())
			if opts.Once {
				return nil
			}
			continue
		}

		// Limpeza de contexto antes do job
		contextMode := cfg.Context
		if contextMode == "" {
			contextMode = config.ContextClear
		}
		if contextMode == config.ContextClear {
			if profile.ResetCommand != "" {
				if err := deliverContextReset(ctx, d, target, job.ID, profile.ResetCommand, quietDuration, opts.Sleep, opts.Now); err != nil {
					fmt.Fprintf(opts.Out, "[%s] aviso ao resetar contexto: %v\n", opts.Name, err)
				}
			} else if !warnedNoReset {
				fmt.Fprintf(opts.Out, "[%s] perfil %s sem reset_command: mantendo contexto contínuo\n", opts.Name, opts.CLI)
				warnedNoReset = true
			}
		}

		rel, _ := filepath.Rel(opts.Dir, entry.Path)
		prompt := workerPrompt(opts.Role, job, filepath.ToSlash(rel), cfg)

		accepted, err := deliverJobPrompt(ctx, d, target, job.ID, prompt, opts.Sleep, opts.Now, 30*time.Second)
		if err != nil || !accepted {
			msg := "o agente não aceitou o prompt"
			if err != nil {
				msg = err.Error()
			}
			fmt.Fprintf(opts.Out, "[%s] ✗ %s\n", opts.Name, msg)
			_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, msg, opts.Now())
			if opts.Once {
				return nil
			}
			continue
		}

		fmt.Fprintf(opts.Out, "▶ job %s · entregue ao painel %s · %s\n", job.ID, target, opts.CLI)

		// Monitora conclusão por estado
		started := opts.Now()
		lastRenew := started
		timeoutMin := cfg.Builder.JobTimeoutMinutes
		if timeoutMin <= 0 {
			timeoutMin = 60
		}
		timeoutDuration := time.Duration(timeoutMin) * time.Minute

		jobDone := false
		for !jobDone {
			if !opts.Sleep(ctx, 500*time.Millisecond) {
				return nil
			}

			// Verifica se terminou por estado
			reason := VerifyJobHandoff(opts.Dir, opts.Role, opts.Name, opts.Now, job.ID)
			if reason == "" {
				jobDone = true
				fmt.Fprintf(opts.Out, "■ job %s concluído · contexto limpo · painel livre\n", job.ID)
				if opts.OnJobDone != nil {
					opts.OnJobDone(job, OutcomeDone)
				}
				break
			}

			// Renova lease
			if opts.Now().Sub(lastRenew) >= opts.RenewEvery {
				if err := bus.Renew(opts.Dir, opts.Role, job.ID, opts.Name, bus.DefaultLeaseTTL, opts.Now()); err != nil {
					fmt.Fprintf(opts.Out, "[%s] aviso ao renovar lease: %v\n", opts.Name, err)
				}
				lastRenew = opts.Now()
			}

			// Inspeciona tela do painel (feat 77-04)
			paneText, captureErr := d.CapturePane(target)
			if captureErr != nil {
				failMsg := "painel fechado: " + captureErr.Error()
				fmt.Fprintf(opts.Out, "[%s] ✗ %s\n", opts.Name, failMsg)
				_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, failMsg, opts.Now())
				if ncfg, err := config.LoadNotificationsConfig(opts.Dir); err == nil {
					n := notify.Notification{
						Type:    notify.TypeInfo,
						Title:   fmt.Sprintf("Painel %s fechado no %s", target, opts.Role),
						Message: fmt.Sprintf("O painel da CLI foi fechado inesperadamente. O job %s voltou à fila.\n%s", job.Feature, captureErr),
						Feature: job.Feature, JobID: job.ID,
						ID: notificationID(job.ID, "painel-fechado"),
					}
					_, _ = notify.Dispatch(opts.Dir, ncfg, n, notify.Deps{})
				}
				return fmt.Errorf("painel fechado: %w", captureErr)
			}

			lines := extractLastLines(paneText, 20)

			// 1. Limite de uso
			if ul, ok := DetectUsageLimit(lines, opts.Now()); ok && !ul.Capped {
				now := opts.Now()
				ul.ResetAt = ul.ResetAt.In(now.Location())
				reason := "limite de uso do " + opts.CLI
				_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, reason, now)
				if ncfg, err := config.LoadNotificationsConfig(opts.Dir); err == nil {
					n := notify.Notification{
						Type:    notify.TypeInfo,
						Title:   fmt.Sprintf("%s pausado: limite de uso do %s", opts.Name, opts.CLI),
						Message: fmt.Sprintf("O job %s voltou à fila sem gastar tentativa. O despachante retoma às %s.\n%s", job.Feature, ul.ResetAt.Format("15:04"), ul.Message),
						Feature: job.Feature, JobID: job.ID,
						ID: notificationID(job.ID, fmt.Sprintf("pausa-%d", now.Unix()/60)),
					}
					_, _ = notify.Dispatch(opts.Dir, ncfg, n, notify.Deps{})
				}
				fmt.Fprintf(opts.Out, "[%s] ⏸ %s\n", opts.Name, ul.Message)
				if opts.OnJobDone != nil {
					opts.OnJobDone(job, OutcomeDeferred)
				}
				if !pauseUntil(ctx, ul.ResetAt, opts.CLI, opts.Now, opts.Sleep, func(s string) {
					fmt.Fprintf(opts.Out, "[%s] %s\n", opts.Name, s)
				}) {
					return nil
				}
				fmt.Fprintf(opts.Out, "[%s] ▶ retomando\n", opts.Name)
				break
			}

			// 2. Agente indisponível
			if au, ok := DetectAgentUnavailable(lines, opts.CLI); ok {
				_, _ = bus.Requeue(opts.Dir, opts.Role, job.ID, opts.Name, "agente indisponível: "+opts.CLI, opts.Now())
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
				fmt.Fprintf(opts.Out, "[%s] ✗ %s indisponível: %s\n", opts.Name, opts.CLI, au.Message)
				fmt.Fprintf(opts.Out, "[%s] → %s\n", opts.Name, au.Hint)
				if opts.OnJobDone != nil {
					opts.OnJobDone(job, OutcomeUnavailable)
				}
				return fmt.Errorf("%s indisponível: %s", opts.CLI, au.Message)
			}

			// Verifica timeout
			if opts.Now().Sub(started) >= timeoutDuration {
				failReason := "tempo limite excedido"
				fmt.Fprintf(opts.Out, "[%s] ✗ job %s falhou: %s\n", opts.Name, job.ID, failReason)
				_, _ = bus.Fail(opts.Dir, opts.Role, job.ID, opts.Name, failReason, opts.Now())
				if opts.OnJobDone != nil {
					opts.OnJobDone(job, OutcomeFailed)
				}
				break
			}
		}

		if opts.Once {
			return nil
		}
	}
}

func extractLastLines(s string, n int) []string {
	raw := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	var lines []string
	for _, l := range raw {
		c := clean(l)
		if strings.TrimSpace(c) != "" {
			lines = append(lines, strings.TrimSpace(c))
		}
	}
	if len(lines) > n {
		return lines[len(lines)-n:]
	}
	return lines
}
