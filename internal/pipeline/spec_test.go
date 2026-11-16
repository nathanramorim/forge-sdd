package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func advancedProject(t *testing.T, advanced bool) string {
	t.Helper()
	dir := project(t, nil)
	if advanced {
		cfg := config.DefaultForgeConfig()
		cfg.Permissions = config.PermissionFull
		cfg.Autonomy = config.AutonomyAdvanced
		cfg.IsolatedConfirmed = true
		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		testutil.WriteFile(t, dir, ".forge/config/forge.json", string(data))
	}
	return dir
}

func reviewerPending(t *testing.T, dir string) []bus.Entry {
	t.Helper()
	list, err := bus.List(dir, layout.RoleReviewer, layout.StatePending)
	require.NoError(t, err)
	return list
}

func TestSpecReview_RoteiaParaReviewer(t *testing.T) {
	role, ok := bus.RoleForType(bus.TypeSpecReview)
	require.True(t, ok)
	assert.Equal(t, layout.RoleReviewer, role)
}

func TestSpecReview_SoEmAvancado(t *testing.T) {
	normal := advancedProject(t, false)
	_, err := Dispatch(normal, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.Len(t, builderPending(t, normal), 1)
	assert.Empty(t, reviewerPending(t, normal), "modo normal segue direto ao Builder")

	adv := advancedProject(t, true)
	res, err := Dispatch(adv, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	assert.Empty(t, builderPending(t, adv))
	jobs := reviewerPending(t, adv)
	require.Len(t, jobs, 1)
	assert.Equal(t, bus.TypeSpecReview, jobs[0].Job.Type)
	assert.Equal(t, 1, jobs[0].Job.Attempt)
}

func TestSpecReview_GateBloqueadoNaoEnfileira(t *testing.T) {
	dir := advancedProject(t, true)
	require.NoError(t, os.Remove(filepath.Join(dir, "sdd/discovery/77-auth-oauth/plan.md")))
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.True(t, res.Blocked())
	assert.Empty(t, reviewerPending(t, dir))
}

func claimSpec(t *testing.T, dir string, at time.Time) bus.Job {
	t.Helper()
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: at})
	require.NoError(t, err)
	e, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, at)
	require.NoError(t, err)
	return e.Job
}

func conforme() queue.Verdict { return queue.Verdict{Veredito: "conforme", Resumo: "Spec conforme."} }

func TestApplySpecVerdict_Aprovado(t *testing.T) {
	dir := advancedProject(t, true)
	job := claimSpec(t, dir, now)
	deps, _ := notifyDeps(nil)
	res, err := ApplySpecVerdict(dir, job.ID, "forge-reviewer", conforme(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)

	pend := builderPending(t, dir)
	require.Len(t, pend, 1)
	assert.Equal(t, bus.TypeFeatureBuild, pend[0].Job.Type)
	assert.Equal(t, 1, pend[0].Job.Attempt)

	data, err := os.ReadFile(filepath.Join(dir, "sdd/discovery/77-auth-oauth", specReviewFile))
	require.NoError(t, err, "review-spec.md é escrito pelo Go")
	assert.Contains(t, string(data), "conforme")

	again, err := ApplySpecVerdict(dir, job.ID, "forge-reviewer", conforme(), deps, now)
	require.NoError(t, err)
	assert.True(t, again.AlreadyApplied)
	assert.Len(t, builderPending(t, dir), 1)
}

func TestApplySpecVerdict_AjustarELimite(t *testing.T) {
	dir := advancedProject(t, true)
	deps, _ := notifyDeps(nil)
	writeQueue(t, dir)

	for attempt := 1; attempt <= 3; attempt++ {
		at := now.Add(time.Duration(attempt) * time.Minute)
		job := claimSpec(t, dir, at)
		assert.Equal(t, attempt, job.Attempt)
		res, err := ApplySpecVerdict(dir, job.ID, "forge-reviewer", violation(), deps, at)
		require.NoError(t, err)
		if attempt < 3 {
			assert.Equal(t, ActionRework, res.Action, "tentativa %d devolve à Discovery", attempt)
			assert.Empty(t, builderPending(t, dir))
			continue
		}
		assert.Equal(t, ActionEscalated, res.Action, "no limite vira blocked")
	}
	data, err := os.ReadFile(filepath.Join(dir, "sdd/discovery/77-auth-oauth/queue.json"))
	require.NoError(t, err)
	var q queue.Queue
	require.NoError(t, json.Unmarshal(data, &q))
	assert.Equal(t, queue.StatusBlocked, q.Items[0].Status)
	assert.NotEmpty(t, q.Items[0].BlockedReason)
}

func TestApplySpecVerdict_ForaDoAvancado(t *testing.T) {
	dir := advancedProject(t, false)
	deps, _ := notifyDeps(nil)
	_, err := ApplySpecVerdict(dir, "x", "forge-reviewer", conforme(), deps, now)
	require.Error(t, err)
}

func writeQueue(t *testing.T, dir string) {
	t.Helper()
	data, err := json.Marshal(queue.Queue{SpecStatus: "reviewed", Items: []queue.QueueItem{{TaskID: "feat-77-01-autenticacao", Status: queue.StatusQueued}}})
	require.NoError(t, err)
	testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/queue.json", string(data))
}
