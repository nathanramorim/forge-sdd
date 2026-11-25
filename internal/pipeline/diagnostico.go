package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// FormatDiagnostico responde em um bloco curto o que um agente precisa para saber por que a
// esteira não anda, sem listar pastas: workers vivos, fila por papel, job em execução, última
// falha com motivo e log, pendentes velhos e o modo do Builder (com o teto, no modo por task).
func FormatDiagnostico(dir string, now time.Time) string {
	if _, err := os.Stat(layout.ForgeRuntime(dir)); err != nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("─── 🩺 DIAGNÓSTICO DA ESTEIRA ────────────────────────────────────────────────\n")

	var ws []string
	for _, papel := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		ws = append(ws, papel+" "+estadoDoWorker(dir, papel))
	}
	fmt.Fprintf(&b, "workers: %s\n", strings.Join(ws, " · "))

	var ultimaFalha *bus.Entry
	var quandoFalha time.Time
	var velhos []string
	for _, papel := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		pend, _ := bus.List(dir, papel, layout.StatePending)
		proc, _ := bus.List(dir, papel, layout.StateProcessing)
		fmt.Fprintf(&b, "fila %s: %d pendente(s) · %d em execução\n", papel, len(pend), len(proc))
		for _, e := range proc {
			inicio := ultimoEvento(e.Job, layout.StateProcessing)
			lease := ""
			if e.Job.Lease != nil && e.Job.Lease.ExpiresAt.Before(now) {
				lease = " · lease vencido (worker morreu?)"
			}
			fmt.Fprintf(&b, "  ▶ %s · %s · há %s%s\n", e.Job.ID, rotuloDoJob(e.Job), duracaoCurta(now.Sub(inicio)), lease)
		}
		for _, e := range pend {
			if c := e.Job.CreatedAt; !c.IsZero() && now.Sub(c) > 24*time.Hour {
				velhos = append(velhos, e.Job.ID)
			}
		}
		falhos, _ := bus.List(dir, papel, layout.StateFailed)
		for i := range falhos {
			if t := ultimoEvento(falhos[i].Job, layout.StateFailed); t.After(quandoFalha) {
				quandoFalha, ultimaFalha = t, &falhos[i]
			}
		}
	}
	if ultimaFalha != nil {
		j := ultimaFalha.Job
		papel := filepath.Base(filepath.Dir(filepath.Dir(ultimaFalha.Path)))
		motivo := "sem motivo registrado"
		if n := len(j.Events); n > 0 && j.Events[n-1].Note != "" {
			motivo = firstLine(j.Events[n-1].Note)
		}
		fmt.Fprintf(&b, "última falha: %s · %s · %s · %s\n", j.ID, rotuloDoJob(j), quandoFalha.Local().Format("02/01 15:04"), motivo)
		log := filepath.Join(".forge", "runtime", "logs", papel+"-"+j.ID+".log")
		if _, err := os.Stat(filepath.Join(dir, log)); err == nil {
			fmt.Fprintf(&b, "  log: %s\n", log)
		}
	}
	if len(velhos) > 0 {
		fmt.Fprintf(&b, "⚠ %d pendente(s) com mais de 24h na fila (o worker pega antes dos novos): %s\n", len(velhos), strings.Join(velhos, ", "))
	}

	if r, ok := LerRecursos(dir); ok {
		b.WriteString(FormatRecursos(r, now))
		return b.String()
	}
	cfg, _, err := config.LoadForgeConfig(dir)
	if err == nil {
		if cfg.Builder.PorTask() && !cfg.IsAdvanced() {
			if teto, terr := calculaTetoMem(dir, cfg.Builder, 1<<10, nil, nil, nil, nil, false); terr == nil {
				fmt.Fprintf(&b, "builder: por task · teto agora %d par(es) (limitado por %s)\n", teto.Pares, teto.Motivo)
			} else {
				fmt.Fprintf(&b, "builder: por task · teto indisponível: %s\n", firstLine(terr.Error()))
			}
		} else {
			b.WriteString("builder: por feature (um Builder por feature)\n")
		}
	}
	return b.String()
}

func estadoDoWorker(dir, papel string) string {
	data, err := os.ReadFile(filepath.Join(layout.ForgeRuntime(dir), "painel", papel+".pid"))
	if err != nil {
		return "sem registro (rode forge up)"
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return "pid inválido (rode forge up)"
	}
	if processAlive(pid) {
		return fmt.Sprintf("vivo (pid %d)", pid)
	}
	return fmt.Sprintf("parado (pid %d; rode forge up)", pid)
}

func ultimoEvento(j bus.Job, estado string) time.Time {
	for i := len(j.Events) - 1; i >= 0; i-- {
		if j.Events[i].State == estado {
			return j.Events[i].At
		}
	}
	if !j.UpdatedAt.IsZero() {
		return j.UpdatedAt
	}
	return j.CreatedAt
}

func rotuloDoJob(j bus.Job) string {
	if j.TaskID != "" {
		return j.Type + " " + j.TaskID
	}
	if len(j.TaskIDs) > 1 {
		return fmt.Sprintf("%s (%d tasks)", j.Type, len(j.TaskIDs))
	}
	return j.Type
}

func duracaoCurta(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh%02d", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
