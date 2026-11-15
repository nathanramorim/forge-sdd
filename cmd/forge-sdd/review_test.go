package main

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

func reviewFixture(t *testing.T, attempt int) string {
	t.Helper()
	dir := t.TempDir()
	cfg, _ := json.Marshal(config.DefaultForgeConfig())
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(cfg))
	url, n := "https://github.com/acme/app/pull/12", 12
	j := bus.Job{
		ID: "job-1-review", Type: bus.TypePRReview, Priority: bus.PriorityNormal, Feature: "auth-oauth", Title: "Auth",
		SpecPath: "sdd/discovery/77-auth-oauth/criteria.md", Card: bus.Card{Provider: "clickup", ID: "CU-1", URL: "https://app.clickup.com/t/1"},
		PR: bus.PR{Number: &n, URL: &url, Branch: "feat/77", BaseBranch: "main"}, Attempt: attempt, MaxReworkAttempts: 3,
	}
	_, _, err := bus.Enqueue(dir, layout.RoleReviewer, j, time.Now())
	require.NoError(t, err)
	_, err = bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, time.Now())
	require.NoError(t, err)
	return dir
}

func writeVerdict(t *testing.T, v queue.Verdict) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	p := filepath.Join(t.TempDir(), "verdict.json")
	require.NoError(t, os.WriteFile(p, data, 0o644))
	return p
}

func resetReviewFlags() { _ = busReviewCmd.Flags().Set("worker", "") }

func TestBusReviewCLI_ReworkEIdempotencia(t *testing.T) {
	t.Setenv("FORGE_WORKER", "forge-reviewer")
	dir := reviewFixture(t, 1)
	resetReviewFlags()
	vf := writeVerdict(t, queue.Verdict{Veredito: "violacao", Resumo: "r", Achados: []queue.Finding{
		{Classificacao: "violacao", Regra: "ARQ-01", Arquivo: "a.go", Linha: 3, Evidencia: "x", Explicacao: "e", Correcao: "c"}}})

	out, err := runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", vf)
	require.NoError(t, err)
	assert.Contains(t, out, "devolvido ao Builder como job-1-rework-2 (tentativa 2 de 3, prioridade high)")
	assert.Contains(t, out, "Comentário sugerido para o PR")

	out, err = runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", vf)
	require.NoError(t, err)
	assert.Contains(t, out, "já aplicado")
	pend, _ := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	assert.Len(t, pend, 1)
}

func TestBusReviewCLI_AprovacaoEViolacoes(t *testing.T) {
	t.Setenv("FORGE_WORKER", "forge-reviewer")
	dir := reviewFixture(t, 1)
	resetReviewFlags()
	_, err := runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", filepath.Join(dir, "nao-existe.json"))
	require.Error(t, err)
	bad := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte("{nao"), 0o644))
	_, err = runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", bad)
	require.Error(t, err)
	_, err = runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", writeVerdict(t, queue.Verdict{Veredito: "violacao"}))
	require.Error(t, err, "violação sem achado é recusada")

	out, err := runRoot(t, "bus", "review", "--dir", dir, "--id", "job-1-review", "--verdict-file", writeVerdict(t, queue.Verdict{Veredito: "conforme", Resumo: "tudo certo"}))
	require.NoError(t, err)
	assert.Contains(t, out, "aprovado: PR pronto")
	assert.Contains(t, out, `para "Ready for Review"`)
	assert.Contains(t, out, "canal discovery_session")
}
