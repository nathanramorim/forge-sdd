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
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reviewProjectIdsReais cria o projeto de reviewProject com TaskIDs do job e a fila da iniciativa.
func reviewProjectIdsReais(t *testing.T, taskIDs []string, items ...queue.QueueItem) (string, string) {
	t.Helper()
	dir := workerProject(t)
	url := prURL
	n := 12
	j := bus.Job{
		ID: "job-1-review", Type: bus.TypePRReview, Priority: bus.PriorityNormal, Feature: "auth-oauth", Title: "Auth", Initiative: "77-auth-oauth",
		SpecPath: "sdd/discovery/77-auth-oauth/criteria.md", TaskIDs: taskIDs,
		PR:      bus.PR{Number: &n, URL: &url, Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		Attempt: 1, MaxReworkAttempts: 3,
	}
	_, _, err := bus.Enqueue(dir, layout.RoleReviewer, j, now)
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, now)
	require.NoError(t, err)

	data, err := json.Marshal(queue.Queue{SpecStatus: "reviewed", Items: items})
	require.NoError(t, err)
	initDir := filepath.Join(dir, "sdd", "discovery", "77-auth-oauth")
	require.NoError(t, os.MkdirAll(initDir, 0o755))
	testutil.WriteFile(t, initDir, layout.QueueFileName, string(data))
	return dir, filepath.Join(initDir, layout.QueueFileName)
}

func queueStatuses(t *testing.T, path string) map[string]string {
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

func vereditoConforme() queue.Verdict {
	return queue.Verdict{Veredito: "conforme", Resumo: "Tudo conforme com as ADRs e rules."}
}

func TestApplyVerdict_IdsDoJobAprovamTasksDaFilaComPrefixoFeat(t *testing.T) {
	dir, path := reviewProjectIdsReais(t, []string{"77-01", "77-02"},
		queue.QueueItem{TaskID: "feat-77-01-credenciais", Status: queue.StatusQueued},
		queue.QueueItem{TaskID: "feat-77-02", Status: queue.StatusQueued},
		queue.QueueItem{TaskID: "feat-77-10-outra", Status: queue.StatusQueued},
	)
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", vereditoConforme(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)
	assert.Empty(t, res.Warning)
	assert.Equal(t, map[string]string{
		"feat-77-01-credenciais": queue.StatusApproved,
		"feat-77-02":             queue.StatusApproved,
		"feat-77-10-outra":       queue.StatusQueued,
	}, queueStatuses(t, path))
}

func TestApplyVerdict_IdsSemCorrespondenciaGeramAvisoENaoErro(t *testing.T) {
	dir, path := reviewProjectIdsReais(t, []string{"99-01"},
		queue.QueueItem{TaskID: "feat-77-01", Status: queue.StatusQueued},
	)
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", vereditoConforme(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)
	assert.Contains(t, res.Warning, "99-01")
	assert.Equal(t, map[string]string{"feat-77-01": queue.StatusQueued}, queueStatuses(t, path))
}

func TestApplyVerdict_TasksJaAprovadasNaoGeramAviso(t *testing.T) {
	dir, _ := reviewProjectIdsReais(t, []string{"77-01"},
		queue.QueueItem{TaskID: "feat-77-01", Status: queue.StatusApproved},
	)
	deps, _ := notifyDeps(nil)

	res, err := ApplyVerdict(dir, "job-1-review", "forge-reviewer", vereditoConforme(), deps, now)
	require.NoError(t, err)
	assert.Empty(t, res.Warning, "casou, só não havia o que mudar")
}
