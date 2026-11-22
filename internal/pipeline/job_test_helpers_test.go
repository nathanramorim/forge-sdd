package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/require"
)

// setupProjectWithJob monta um projeto de pipeline com um job na fila do papel.
// Vive aqui (antes morava no teste do despachante, que saiu com o tmux na feat-115-02) porque
// os testes de archive e de safeguards dependem dele.
func setupProjectWithJob(t *testing.T, role string) (string, bus.Job) {
	t.Helper()
	dir := project(t, func(d string) {
		cfg := config.DefaultForgeConfig()
		cfg.SessionMode = config.SessionModeLive
		cfg.Context = config.ContextClear
		cfg.Live.StartupWaitSeconds = 0
		cfg.Live.ReadyQuietSeconds = 1
		data, _ := json.Marshal(cfg)
		testutil.WriteFile(t, d, ".forge/config/forge.json", string(data))
	})
	job := bus.Job{
		ID:                "job-test-1",
		Feature:           "auth-oauth",
		Title:             "Auth",
		SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
		Type:              bus.TypeFeatureBuild,
		Priority:          bus.PriorityNormal,
		PR:                bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		Attempt:           1,
		MaxReworkAttempts: 3,
		CreatedAt:         now,
	}
	if role == layout.RoleReviewer {
		prNum := 10
		prURL := "https://github.com/pr/10"
		job.Type = bus.TypePRReview
		job.PR.Number = &prNum
		job.PR.URL = &prURL
	}
	_, _, err := bus.Enqueue(dir, role, job, now)
	require.NoError(t, err)
	return dir, job
}
