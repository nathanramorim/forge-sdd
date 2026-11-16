package pipeline

import (
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

// HaltReason diz por que o despacho deve parar, ou "" se pode seguir. O STOP vale em qualquer
// modo; os tetos de jobs por dia valem só no modo avançado (feat 88-07).
func HaltReason(dir string, cfg config.ForgeConfig, now time.Time) string {
	if _, err := os.Stat(StopFile(dir)); err == nil {
		return "arquivo .forge/runtime/STOP presente"
	}
	if !cfg.IsAdvanced() || cfg.Safeguards.MaxJobsPerDay <= 0 {
		return ""
	}
	if n := jobsStartedToday(dir, now); n >= cfg.Safeguards.MaxJobsPerDay {
		return fmt.Sprintf("teto diário de %d jobs atingido (%d hoje)", cfg.Safeguards.MaxJobsPerDay, n)
	}
	return ""
}

// jobsStartedToday conta as reivindicações (transição para processing) feitas hoje, em UTC.
func jobsStartedToday(dir string, now time.Time) int {
	day := now.UTC().Format("2006-01-02")
	n := 0
	for _, role := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		for _, state := range []string{layout.StateProcessing, layout.StateDone, layout.StateFailed} {
			entries, err := bus.List(dir, role, state)
			if err != nil {
				continue
			}
			for _, e := range entries {
				for _, ev := range e.Job.Events {
					if ev.State == layout.StateProcessing && ev.At.UTC().Format("2006-01-02") == day {
						n++
					}
				}
			}
		}
	}
	return n
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
