package pipeline

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func spLoc(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Skip("tzdata indisponível")
	}
	return loc
}

func TestDetectUsageLimit_MensagensReais(t *testing.T) {
	loc := spLoc(t)
	now := time.Date(2026, 11, 14, 22, 30, 0, 0, loc)

	ul, ok := DetectUsageLimit([]string{"  algo", "You've hit your 5-hour limit · resets 10:40pm (America/Sao_Paulo)"}, now)
	require.True(t, ok)
	assert.True(t, ul.Known)
	assert.Equal(t, time.Date(2026, 11, 14, 22, 40, 0, 0, loc), ul.ResetAt)
	assert.False(t, ul.Capped)
	assert.Contains(t, ul.Message, "5-hour limit")

	// Horário já passado hoje vira amanhã (acima do teto de 6 h: falha em vez de pausar).
	ul, ok = DetectUsageLimit([]string{"You've hit your limit · resets 9am (America/Sao_Paulo)"}, now)
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 11, 15, 9, 0, 0, 0, loc), ul.ResetAt)
	assert.True(t, ul.Capped, "10h30 de espera passa do teto de 6 h")

	// Relativo.
	ul, ok = DetectUsageLimit([]string{"Rate limit exceeded. Try again in 30 minutes."}, now)
	require.True(t, ok)
	assert.Equal(t, now.Add(30*time.Minute), ul.ResetAt)
	assert.True(t, ul.Known)

	// Sem horário na mensagem: espera de 30 minutos.
	ul, ok = DetectUsageLimit([]string{"Error 429: too many requests"}, now)
	require.True(t, ok)
	assert.False(t, ul.Known)
	assert.Equal(t, now.Add(UnknownUsageLimitPause), ul.ResetAt)

	for _, msg := range []string{"quota exceeded for project", "usage limit reached", "RESOURCE_EXHAUSTED: quota"} {
		_, ok = DetectUsageLimit([]string{msg}, now)
		assert.True(t, ok, msg)
	}
	_, ok = DetectUsageLimit([]string{"tudo certo", "build ok"}, now)
	assert.False(t, ok)
	_, ok = DetectUsageLimit(nil, now)
	assert.False(t, ok)
}

func TestDetectUsageLimit_HorarioInvalidoNaoQuebra(t *testing.T) {
	now := time.Date(2026, 11, 14, 22, 30, 0, 0, time.UTC)
	ul, ok := DetectUsageLimit([]string{"usage limit reached, resets 99:99pm (Nao/Existe)"}, now)
	require.True(t, ok)
	assert.False(t, ul.Known)
	assert.Equal(t, now.Add(UnknownUsageLimitPause), ul.ResetAt)
}

func TestPauseUntil_ContagemECtrlC(t *testing.T) {
	cur := time.Date(2026, 11, 14, 22, 30, 0, 0, time.UTC)
	until := cur.Add(150 * time.Second)
	var said []string
	ok := pauseUntil(context.Background(), until, "claude", func() time.Time { return cur },
		func(_ context.Context, d time.Duration) bool { cur = cur.Add(d); return true },
		func(s string) { said = append(said, s) })
	assert.True(t, ok)
	require.Len(t, said, 3, "uma linha por minuto: 2m30s, 1m30s e 30s")
	assert.Contains(t, said[0], "retomo às 22:32 (faltam 2m30s)")
	assert.Contains(t, said[2], "faltam 30s")

	// Ctrl+C: o sleep devolve false.
	cur = time.Date(2026, 11, 14, 22, 30, 0, 0, time.UTC)
	ok = pauseUntil(context.Background(), cur.Add(time.Hour), "claude", func() time.Time { return cur },
		func(context.Context, time.Duration) bool { return false }, func(string) {})
	assert.False(t, ok)
}

// ---- worker ----

type scriptRunner struct {
	mu    sync.Mutex
	calls int
	fn    func(call int, out io.Writer) (RunResult, error)
}

func (r *scriptRunner) Run(ctx context.Context, spec RunSpec, out io.Writer) (RunResult, error) {
	r.mu.Lock()
	r.calls++
	n := r.calls
	r.mu.Unlock()
	return r.fn(n, out)
}

func TestWorker_MotivoLegivel(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	r := &scriptRunner{fn: func(int, io.Writer) (RunResult, error) { return RunResult{ExitCode: 1}, nil }}
	r.fn = func(_ int, out io.Writer) (RunResult, error) {
		_, _ = io.WriteString(out, "\x1b[31mErro: módulo x não encontrado\x1b[0m\n")
		_, _ = io.WriteString(out, "SessionEnd hook [http://127.0.0.1:47600/hooks/x] failed: connect ECONNREFUSED\n")
		return RunResult{ExitCode: 1}, nil
	}
	var term bytes.Buffer
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Out: &term, Runner: r, WatchEvery: -1, Now: func() time.Time { return now }}))
	failed, _ := bus.List(dir, layout.RoleBuilder, layout.StateFailed)
	require.Len(t, failed, 1)
	last := failed[0].Job.Events[len(failed[0].Job.Events)-1].Note
	assert.Equal(t, "o agente saiu com código 1: Erro: módulo x não encontrado", last, "a última linha útil, sem ANSI e sem o ruído do hook")
	assert.Contains(t, term.String(), "✗ falhou em 0s: o agente saiu com código 1: Erro: módulo x não encontrado")
}

func TestWorker_LimiteDeUso(t *testing.T) {
	loc := spLoc(t)
	cur := time.Date(2026, 11, 14, 22, 30, 0, 0, loc)
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, cur)
	r := &scriptRunner{}
	r.fn = func(call int, out io.Writer) (RunResult, error) {
		if call == 1 {
			_, _ = io.WriteString(out, "You've hit your 5-hour limit · resets 10:40pm (America/Sao_Paulo)\n")
			return RunResult{ExitCode: 1}, nil
		}
		_, err := bus.UpdatePR(dir, layout.RoleBuilder, "job-1", "forge-builder", 12, "https://github.com/acme/app/pull/12", "feat/77-auth-oauth", cur)
		return RunResult{}, err
	}
	var slept []time.Duration
	var outcomes []string
	var term bytes.Buffer
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{
		Dir: dir, Role: layout.RoleBuilder, Once: true, Out: &term, Runner: r, WatchEvery: -1,
		Now:       func() time.Time { return cur },
		Sleep:     func(_ context.Context, d time.Duration) bool { slept = append(slept, d); cur = cur.Add(d); return true },
		OnJobDone: func(j bus.Job, o string) { outcomes = append(outcomes, o) },
	}))

	assert.Equal(t, 2, r.calls, "o job voltou à fila e rodou de novo depois da pausa")
	assert.Equal(t, []string{OutcomeDeferred, OutcomeDone}, outcomes)
	assert.Len(t, slept, 10, "dez minutos de pausa, um minuto por vez")
	out := term.String()
	assert.Contains(t, out, "⏸ You've hit your 5-hour limit")
	assert.Contains(t, out, "retomo às 22:40 (faltam 10m00s)")
	assert.Contains(t, out, "▶ retomando")

	done, _ := bus.List(dir, layout.RoleBuilder, layout.StateDone)
	require.Len(t, done, 1)
	assert.Equal(t, 1, done[0].Job.Attempt, "o limite de uso não gasta tentativa")
	var notes []string
	for _, e := range done[0].Job.Events {
		notes = append(notes, e.Note)
	}
	assert.Contains(t, strings.Join(notes, "|"), "adiado: limite de uso do claude")

	items, err := notify.ListInbox(dir, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, notify.TypeInfo, items[0].Type)
	assert.Contains(t, items[0].Title, "limite de uso")
	assert.Contains(t, items[0].Message, "22:40")
}

func TestWorker_LimiteDeUsoSemMatchOuComSaidaZeroNaoPausa(t *testing.T) {
	// Texto de "rate limit" numa saída que terminou bem não pausa nada.
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	r := &scriptRunner{fn: func(_ int, out io.Writer) (RunResult, error) {
		_, _ = io.WriteString(out, "implementei o rate limit do endpoint de login\n")
		_, err := bus.UpdatePR(dir, layout.RoleBuilder, "job-1", "forge-builder", 12, "https://github.com/acme/app/pull/12", "feat/77-auth-oauth", now)
		return RunResult{}, err
	}}
	var slept int
	var term bytes.Buffer
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Out: &term, Runner: r, WatchEvery: -1,
		Now: func() time.Time { return now }, Sleep: func(context.Context, time.Duration) bool { slept++; return true }}))
	assert.Equal(t, 0, slept)
	assert.NotContains(t, term.String(), "⏸")

	// Saída diferente de zero sem mensagem de limite: falha como antes.
	dir2 := workerProject(t)
	seedJob(t, dir2, layout.RoleBuilder, "job-2", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	r2 := &scriptRunner{fn: func(_ int, out io.Writer) (RunResult, error) {
		_, _ = io.WriteString(out, "falhou por outro motivo\n")
		return RunResult{ExitCode: 2}, nil
	}}
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir2, Role: layout.RoleBuilder, Once: true, Out: io.Discard, Runner: r2, WatchEvery: -1, Now: func() time.Time { return now }}))
	failed, _ := bus.List(dir2, layout.RoleBuilder, layout.StateFailed)
	assert.Len(t, failed, 1)
}

func TestWorker_LimiteAcimaDoTetoFalhaComOMotivo(t *testing.T) {
	loc := spLoc(t)
	cur := time.Date(2026, 11, 14, 22, 30, 0, 0, loc)
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, cur)
	r := &scriptRunner{fn: func(_ int, out io.Writer) (RunResult, error) {
		_, _ = io.WriteString(out, "You've hit your weekly limit · resets 9am (America/Sao_Paulo)\n")
		return RunResult{ExitCode: 1}, nil
	}}
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Out: io.Discard, Runner: r, WatchEvery: -1,
		Now: func() time.Time { return cur }, Sleep: func(context.Context, time.Duration) bool { t.Fatal("não pode pausar"); return true }}))
	failed, _ := bus.List(dir, layout.RoleBuilder, layout.StateFailed)
	require.Len(t, failed, 1)
	assert.Contains(t, failed[0].Job.Events[len(failed[0].Job.Events)-1].Note, "weekly limit")
}
