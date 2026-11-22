package doctor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
)

func enfileira(t *testing.T, dir string) {
	t.Helper()
	now := time.Now()
	job := bus.Job{ID: "job-1", Feature: "auth-oauth", Title: "Auth", SpecPath: "sdd/discovery/77-auth-oauth/criteria.md",
		Type: bus.TypeFeatureBuild, Priority: bus.PriorityNormal, PR: bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		Attempt: 1, MaxReworkAttempts: 3, CreatedAt: now}
	if _, _, err := bus.Enqueue(dir, layout.RoleBuilder, job, now); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorFilaSemWorker(t *testing.T) {
	dir := projeto(t, "5.0.0")
	d := deps("5.0.0")
	if c := achar(t, Diagnose(dir, d), "fila-sem-worker"); c.Status != StatusOK {
		t.Fatalf("fila vazia deve ser ok: %+v", c)
	}
	enfileira(t, dir)
	r := Diagnose(dir, d)
	c := achar(t, r, "fila-sem-worker")
	if c.Status != StatusAviso || r.Falhas != 0 {
		t.Fatalf("fila parada sem worker deve avisar sem falhar: %+v", c)
	}
	if err := ValidateRelatorio(r); err != nil {
		t.Fatal(err)
	}
	d.Runner = func(name string, _ ...string) ([]byte, error) {
		if name == "ps" {
			return []byte("PID PPID %CPU RSS ELAPSED COMMAND\n42 1 0.5 1000 01:00 /bin/forge-sdd worker --role builder --dir " + dir + "\n"), nil
		}
		return nil, errNoTmux
	}
	if c := achar(t, Diagnose(dir, d), "fila-sem-worker"); c.Status != StatusOK || !strings.Contains(c.Detalhe, "1 worker") {
		t.Fatalf("com worker ativo não avisa: %+v", c)
	}
}

var errNoTmux = errors.New("sem tmux")
