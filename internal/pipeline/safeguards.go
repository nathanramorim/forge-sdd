package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/notify"
)

// StopFileName é o arquivo que para a esteira de forma limpa: o worker termina o job em curso e
// não reivindica outro enquanto o arquivo existir.
const StopFileName = "STOP"

// StopFile devolve .forge/runtime/STOP.
func StopFile(dir string) string { return filepath.Join(layout.ForgeRuntime(dir), StopFileName) }

// HaltReason diz por que o despacho deve encerrar, ou "" se pode seguir. Só o STOP encerra, em
// qualquer modo; o teto diário não encerra, pausa (DailyPause).
func HaltReason(dir string, _ config.ForgeConfig, _ time.Time) string {
	if _, err := os.Stat(StopFile(dir)); err == nil {
		return "arquivo .forge/runtime/STOP presente"
	}
	return ""
}

// DailyPause diz se o teto de jobs por dia (só no modo avançado) foi atingido. Devolve o motivo e o
// início do próximo dia no fuso de now; sem teto atingido, o motivo vem vazio.
func DailyPause(dir string, cfg config.ForgeConfig, now time.Time) (string, time.Time) {
	limit := cfg.Safeguards.MaxJobsPerDay
	if !cfg.IsAdvanced() || limit <= 0 {
		return "", time.Time{}
	}
	n := jobsStartedToday(dir, now)
	if n < limit {
		return "", time.Time{}
	}
	y, m, d := now.Date()
	return fmt.Sprintf("teto diário de %d jobs atingido (%d hoje)", limit, n), time.Date(y, m, d+1, 0, 0, 0, 0, now.Location())
}

// jobsStartedToday conta os jobs distintos reivindicados (transição para processing) no dia de now,
// no fuso de now; reclaim, rework e revisão do mesmo job contam uma vez.
func jobsStartedToday(dir string, now time.Time) int {
	loc := now.Location()
	day := now.Format("2006-01-02")
	seen := map[string]bool{}
	for _, role := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		for _, state := range []string{layout.StateProcessing, layout.StateDone, layout.StateFailed} {
			entries, err := bus.List(dir, role, state)
			if err != nil {
				continue
			}
			for _, e := range entries {
				for _, ev := range e.Job.Events {
					if ev.State == layout.StateProcessing && ev.At.In(loc).Format("2006-01-02") == day {
						seen[e.Job.ID] = true
					}
				}
			}
		}
	}
	return len(seen)
}

// notifyPause avisa uma vez por dia e motivo que a esteira está em pausa até o próximo dia.
func notifyPause(dir, reason string, until, now time.Time) {
	ncfg, err := config.LoadNotificationsConfig(dir)
	if err != nil {
		return
	}
	n := notify.Notification{
		ID:      "esteira-pausa-" + now.Format("20060102"),
		Type:    notify.TypeEscalation,
		Title:   "Esteira em pausa",
		Message: reason + ". Em pausa até " + until.Format("02/01/2006 15:04") + "; retoma sozinha.",
	}
	_, _ = notify.Dispatch(dir, ncfg, n, notify.Deps{})
}

// waitDailyPause bloqueia enquanto o teto diário estiver atingido. Devolve false se o contexto
// terminou ou o STOP apareceu (o chamador reavalia); true quando pode seguir.
func waitDailyPause(ctx context.Context, opts WorkerOptions, cfg config.ForgeConfig, name string) bool {
	notified := false
	for {
		reason, until := DailyPause(opts.Dir, cfg, opts.Now())
		if reason == "" {
			return true
		}
		if !notified {
			fmt.Fprintf(opts.Out, "[%s] esteira em pausa: %s; retoma em %s\n", name, reason, until.Format("02/01 15:04"))
			notifyPause(opts.Dir, reason, until, opts.Now())
			notified = true
		}
		if opts.Once || HaltReason(opts.Dir, cfg, opts.Now()) != "" || !opts.Sleep(ctx, opts.Poll) {
			return false
		}
	}
}

// JobOverCeiling diz se o job passou do teto de tokens (estimado pelo tamanho do log) ou de tempo
// (builder.job_timeout_minutes). Só vale no modo avançado.
func JobOverCeiling(dir string, cfg config.ForgeConfig, role, jobID string, elapsed time.Duration) string {
	if !cfg.IsAdvanced() {
		return ""
	}
	if limit := cfg.Safeguards.MaxJobTokens; limit > 0 {
		if info, err := os.Stat(filepath.Join(layout.Logs(dir), role+"-"+jobID+".log")); err == nil {
			if tok := monitor.EstimateTokensFromBytes(int(info.Size())); tok > limit {
				return fmt.Sprintf("job %s passou do teto de %d tokens (~%d)", jobID, limit, tok)
			}
		}
	}
	if max := time.Duration(cfg.Builder.JobTimeoutMinutes) * time.Minute; max > 0 && elapsed > max {
		return fmt.Sprintf("job %s passou do teto de %d minutos", jobID, cfg.Builder.JobTimeoutMinutes)
	}
	return ""
}

// notifyHalt grava o motivo da parada no inbox (idempotente por dia e motivo).
func notifyHalt(dir, reason string, now time.Time) {
	ncfg, err := config.LoadNotificationsConfig(dir)
	if err != nil {
		return
	}
	n := notify.Notification{
		ID:      haltNotificationID(now, reason),
		Type:    notify.TypeEscalation,
		Title:   "Esteira parada",
		Message: reason + ". Remova a causa (ou o arquivo STOP) e rode forge up de novo.",
	}
	_, _ = notify.Dispatch(dir, ncfg, n, notify.Deps{})
}

func haltNotificationID(now time.Time, reason string) string {
	sum := sha256.Sum256([]byte(reason))
	return "esteira-parada-" + now.UTC().Format("20060102") + "-" + hex.EncodeToString(sum[:4])
}
