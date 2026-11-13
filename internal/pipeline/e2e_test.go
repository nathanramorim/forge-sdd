package pipeline

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_PipelineCompleto percorre o pipeline inteiro com agentes simulados: dispatch com card
// válido, Builder abre o PR, Reviewer reprova uma vez (rework de prioridade alta), o Builder
// corrige, o Reviewer aprova e o inbox da Discovery recebe o link do PR (sem MCP, sem modelo).
func TestE2E_PipelineCompleto(t *testing.T) {
	dir := project(t, nil)
	deps, toasts := notifyDeps(nil)
	clock := func() time.Time { return now }

	// 1. Gate e despacho.
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)

	builderAgent := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) {
		entries, err := bus.List(dir, layout.RoleBuilder, layout.StateProcessing)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		url := "https://github.com/acme/app/pull/12"
		_, err = bus.UpdatePR(dir, layout.RoleBuilder, entries[0].Job.ID, "forge-builder", 12, url, "", now)
		return RunResult{}, err
	}}

	verdicts := []queue.Verdict{
		{Veredito: "violacao", Resumo: "ADR 0003 violada.", Achados: []queue.Finding{{Classificacao: "violacao", Regra: "SEG-01", Arquivo: "internal/auth/handler.go", Linha: 10, Evidencia: "r.URL.Query().Get(\"redirect\")", Explicacao: "Redirect sem validação.", Correcao: "Validar contra lista permitida."}}},
		{Veredito: "conforme", Resumo: "Tudo conforme."},
	}
	var applied []string
	reviewerAgent := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) {
		entries, err := bus.List(dir, layout.RoleReviewer, layout.StateProcessing)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		v := verdicts[0]
		verdicts = verdicts[1:]
		r, err := ApplyVerdict(dir, entries[0].Job.ID, "forge-reviewer", v, deps, now)
		if err == nil {
			applied = append(applied, r.Action)
		}
		return RunResult{}, err
	}}

	runWorker := func(role string, r Runner) {
		require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: role, Once: true, Out: io.Discard, Runner: r, Now: clock}))
	}

	// 2. Builder -> Reviewer (reprova) -> Builder (rework, high) -> Reviewer (aprova).
	runWorker(layout.RoleBuilder, builderAgent)
	runWorker(layout.RoleReviewer, reviewerAgent)
	pend, _ := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	require.Len(t, pend, 1)
	assert.Equal(t, bus.TypePRRework, pend[0].Job.Type)
	assert.Equal(t, bus.PriorityHigh, pend[0].Job.Priority)
	assert.Equal(t, 2, pend[0].Job.Attempt)
	assert.Empty(t, mustInbox(t, dir), "enquanto o ciclo é automático, a pessoa não é incomodada")

	runWorker(layout.RoleBuilder, builderAgent)
	runWorker(layout.RoleReviewer, reviewerAgent)
	assert.Equal(t, []string{ActionRework, ActionApproved}, applied)

	// 3. Resultado: inbox com o link do PR; filas sem pendências; tudo em done.
	items := mustInbox(t, dir)
	require.Len(t, items, 1)
	assert.Equal(t, notify.TypePRApproved, items[0].Type)
	assert.Equal(t, "https://github.com/acme/app/pull/12", items[0].PRURL)
	assert.Contains(t, notify.FormatInbox(items, true), "https://github.com/acme/app/pull/12")
	assert.NotEmpty(t, *toasts)

	for _, role := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		for _, st := range []string{layout.StatePending, layout.StateProcessing, layout.StateFailed} {
			l, err := bus.List(dir, role, st)
			require.NoError(t, err)
			assert.Empty(t, l, "%s/%s deve estar vazio", role, st)
		}
		done, _ := bus.List(dir, role, layout.StateDone)
		assert.Len(t, done, 2, "%s: dois jobs concluídos (original e rework)", role)
	}
}

func mustInbox(t *testing.T, dir string) []notify.InboxItem {
	t.Helper()
	items, err := notify.ListInbox(dir, true)
	require.NoError(t, err)
	return items
}
