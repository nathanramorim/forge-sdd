package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reviewProjectComFila cria um job pr_review da iniciativa 77-auth-oauth cobrindo taskIDs,
// com o queue.json dessa iniciativa em disco.
func reviewProjectComFila(t *testing.T, taskIDs []string, statuses map[string]string) (dir, queuePath string) {
	t.Helper()
	dir = workerProject(t)
	url, n := prURL, 12
	j := bus.Job{
		ID: "job-1-review", Type: bus.TypePRReview, Priority: bus.PriorityNormal, Feature: "auth-oauth", Title: "Auth", Initiative: "77-auth-oauth",
		TaskIDs:           taskIDs,
		SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
		PR:                bus.PR{Number: &n, URL: &url, Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		MaxReworkAttempts: 3,
	}
	_, _, err := bus.Enqueue(dir, layout.RoleReviewer, j, now)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, now)
	require.NoError(t, err)

	base, err := layout.Initiative(dir, "77-auth-oauth")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(base, 0o755))
	var items []queue.QueueItem
	for _, id := range []string{"01", "02", "03"} {
		items = append(items, queue.QueueItem{TaskID: id, Status: statuses[id]})
	}
	data, err := json.Marshal(queue.Queue{SpecStatus: "reviewed", Items: items})
	require.NoError(t, err)
	queuePath = filepath.Join(base, layout.QueueFileName)
	require.NoError(t, os.WriteFile(queuePath, data, 0o644))
	return dir, queuePath
}

func statusesDe(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var q queue.Queue
	require.NoError(t, json.Unmarshal(data, &q))
	out := map[string]string{}
	for _, it := range q.Items {
		out[it.TaskID] = it.Status
	}
	return out
}

func TestReview_ConformePropagaApprovedParaAsTasksDoJob(t *testing.T) {
	dir, qp := reviewProjectComFila(t, []string{"01", "02"}, map[string]string{
		"01": queue.StatusQueued, "02": queue.StatusChangesRequested, "03": queue.StatusQueued,
	})
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "Tudo conforme."}, deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)
	assert.Equal(t, map[string]string{"01": queue.StatusApproved, "02": queue.StatusApproved, "03": queue.StatusQueued}, statusesDe(t, qp))

	// reaplicar o mesmo veredito é idempotente e não mexe na fila de novo
	res, err = ApplyVerdict(dir, "job-1-review", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "Tudo conforme."}, deps, now)
	require.NoError(t, err)
	assert.True(t, res.AlreadyApplied)
	assert.Equal(t, map[string]string{"01": queue.StatusApproved, "02": queue.StatusApproved, "03": queue.StatusQueued}, statusesDe(t, qp))
}

func TestReview_ConformeSemTaskIDsMantemAFila(t *testing.T) {
	dir, qp := reviewProjectComFila(t, nil, map[string]string{"01": queue.StatusQueued, "02": queue.StatusQueued, "03": queue.StatusQueued})
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "ok"}, deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)
	assert.Equal(t, map[string]string{"01": queue.StatusQueued, "02": queue.StatusQueued, "03": queue.StatusQueued}, statusesDe(t, qp))
}

func TestReview_ViolacaoNaoAprovaAsTasks(t *testing.T) {
	dir, qp := reviewProjectComFila(t, []string{"01"}, map[string]string{"01": queue.StatusReviewPending})
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", violation(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionRework, res.Action)
	assert.Equal(t, queue.StatusReviewPending, statusesDe(t, qp)["01"])
}

func TestReview_GuardaCopiaDoVeredito(t *testing.T) {
	dir, _ := reviewProjectComFila(t, []string{"01"}, map[string]string{"01": queue.StatusReviewPending})
	deps, _ := notifyDeps(nil)

	_, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "Tudo conforme."}, deps, now)
	require.NoError(t, err)

	base, err := layout.Initiative(dir, "77-auth-oauth")
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(base, "verdicts", "job-1-review.json"))
	require.NoError(t, err)
	var got queue.Verdict
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "conforme", got.Veredito)
}

func TestReview_SemPastaDaIniciativaNaoCriaCopiaNemFalha(t *testing.T) {
	dir := reviewProject(t, 1) // job de 77-auth-oauth, sem pasta da iniciativa em disco
	deps, _ := notifyDeps(nil)

	_, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "ok"}, deps, now)
	require.NoError(t, err)
	base, err := layout.Initiative(dir, "77-auth-oauth")
	require.NoError(t, err)
	_, statErr := os.Stat(base)
	assert.True(t, os.IsNotExist(statErr), "não deve criar a pasta da iniciativa")
}
