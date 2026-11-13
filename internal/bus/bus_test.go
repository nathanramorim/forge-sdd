package bus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 11, 13, 14, 15, 0, 0, time.UTC)

func newJob(id, typ, prio string) Job {
	return Job{
		ID:                id,
		Type:              typ,
		Priority:          prio,
		Feature:           "auth-oauth",
		Title:             "Autenticação OAuth2 com refresh token",
		SpecPath:          "sdd/discovery/70-x/criteria.md",
		ADRRefs:           []string{"sdd/spec/adr/adr-0003.md"},
		RulesRefs:         []string{".agents/rules/sdd-seguranca.md"},
		Card:              Card{Provider: "clickup", ID: "CU-8421", URL: "https://app.clickup.com/t/8421"},
		PR:                PR{Branch: "feat/auth-oauth", BaseBranch: "main"},
		Attempt:           1,
		MaxReworkAttempts: 3,
	}
}

func mustEnqueue(t *testing.T, dir, role string, j Job, at time.Time) string {
	t.Helper()
	p, created, err := Enqueue(dir, role, j, at)
	require.NoError(t, err)
	require.True(t, created)
	return p
}

func TestEnqueue_GravaEmPendingComContratoCompleto(t *testing.T) {
	dir := t.TempDir()
	p := mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)

	assert.Equal(t, filepath.Join(layout.Forge(dir), "runtime", "queues", "builder", "pending"), filepath.Dir(p))
	assert.True(t, strings.HasPrefix(filepath.Base(p), "1-"), "normal usa prefixo 1")
	assert.True(t, strings.HasSuffix(filepath.Base(p), "-job-1.json"))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	var generic map[string]any
	require.NoError(t, json.Unmarshal(raw, &generic))
	for _, k := range []string{"id", "type", "priority", "feature", "title", "spec_path", "adr_refs", "rules_refs", "card", "pr", "attempt", "max_rework_attempts", "rework_feedback", "created_at", "updated_at", "schema_version", "events"} {
		assert.Contains(t, generic, k, "campo %s do contrato", k)
	}
	pr := generic["pr"].(map[string]any)
	assert.Nil(t, pr["number"])
	assert.Nil(t, pr["url"])
	assert.Equal(t, "feat/auth-oauth", pr["branch"])
	assert.NotContains(t, string(raw), "\"lease\"", "sem lease enquanto está em pending")
}

func TestEnqueue_DuplicataEhIdempotente(t *testing.T) {
	dir := t.TempDir()
	first := mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	again, created, err := Enqueue(dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0.Add(time.Hour))
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, first, again)
	list, err := List(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Também depois de concluído: o id não volta para a fila.
	e, err := Claim(dir, layout.RoleBuilder, "w1", time.Minute, t0)
	require.NoError(t, err)
	_, err = Complete(dir, layout.RoleBuilder, e.Job.ID, "w1", "ok", t0)
	require.NoError(t, err)
	_, created, err = Enqueue(dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	require.NoError(t, err)
	assert.False(t, created)
}

func TestClaim_PrioridadeHighAntesDeNormalEFIFO(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("normal-a", TypeFeatureBuild, PriorityNormal), t0)
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("normal-b", TypeFeatureBuild, PriorityNormal), t0.Add(1*time.Second))
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("rework-a", TypePRRework, PriorityHigh), t0.Add(2*time.Second))
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("rework-b", TypePRRework, PriorityHigh), t0.Add(3*time.Second))
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("normal-c", TypeFeatureBuild, PriorityNormal), t0.Add(4*time.Second))

	var order []string
	for {
		e, err := Claim(dir, layout.RoleBuilder, "w1", time.Minute, t0.Add(time.Minute))
		if err == ErrEmpty {
			break
		}
		require.NoError(t, err)
		order = append(order, e.Job.ID)
	}
	assert.Equal(t, []string{"rework-a", "rework-b", "normal-a", "normal-b", "normal-c"}, order,
		"high antes de normal; FIFO dentro da mesma prioridade")
}

func TestClaim_MoveParaProcessingComLeaseEEvento(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	e, err := Claim(dir, layout.RoleBuilder, "forge-builder", 10*time.Minute, t0)
	require.NoError(t, err)

	assert.Equal(t, layout.StateProcessing, e.State)
	require.NotNil(t, e.Job.Lease)
	assert.Equal(t, "forge-builder", e.Job.Lease.Worker)
	assert.Equal(t, t0.Add(10*time.Minute), e.Job.Lease.ExpiresAt)
	assert.Equal(t, "processing", e.Job.Events[len(e.Job.Events)-1].State)

	pend, _ := List(dir, layout.RoleBuilder, layout.StatePending)
	assert.Empty(t, pend)
	_, err = Claim(dir, layout.RoleBuilder, "w2", time.Minute, t0)
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestClaim_Concorrente(t *testing.T) {
	dir := t.TempDir()
	const jobs = 24
	for i := 0; i < jobs; i++ {
		mustEnqueue(t, dir, layout.RoleBuilder, newJob(fmt.Sprintf("job-%02d", i), TypeFeatureBuild, PriorityNormal), t0.Add(time.Duration(i)*time.Millisecond))
	}

	var mu sync.Mutex
	got := map[string]string{}
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		worker := fmt.Sprintf("w%d", w)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				e, err := Claim(dir, layout.RoleBuilder, worker, time.Minute, t0)
				if err == ErrEmpty {
					return
				}
				if !assert.NoError(t, err) {
					return
				}
				mu.Lock()
				if prev, dup := got[e.Job.ID]; dup {
					t.Errorf("job %s reivindicado por %s e %s", e.Job.ID, prev, worker)
				}
				got[e.Job.ID] = worker
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.Len(t, got, jobs, "todo job é reivindicado exatamente uma vez")

	entries, err := List(dir, layout.RoleBuilder, layout.StateProcessing)
	require.NoError(t, err)
	assert.Len(t, entries, jobs)
	for _, e := range entries {
		assert.Equal(t, got[e.Job.ID], e.Job.Lease.Worker, "o lease pertence a quem venceu")
	}
}

func TestClaim_UmSoJobEOitoWorkers(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleReviewer, func() Job {
		j := newJob("pr-1", TypePRReview, PriorityNormal)
		return j
	}(), t0)
	var wins int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		worker := fmt.Sprintf("w%d", w)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Claim(dir, layout.RoleReviewer, worker, time.Minute, t0); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 1, wins, "exatamente um worker vence a disputa")
}

func TestCompleteEFail_ValidamDonoEMoveParaEstados(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-2", TypeFeatureBuild, PriorityNormal), t0.Add(time.Second))
	_, err := Claim(dir, layout.RoleBuilder, "w1", time.Minute, t0)
	require.NoError(t, err)
	_, err = Claim(dir, layout.RoleBuilder, "w2", time.Minute, t0)
	require.NoError(t, err)

	_, err = Complete(dir, layout.RoleBuilder, "job-1", "w2", "x", t0)
	assert.ErrorIs(t, err, ErrNotOwner, "outro worker não conclui o job alheio")

	done, err := Complete(dir, layout.RoleBuilder, "job-1", "w1", "PR #10 aberto", t0)
	require.NoError(t, err)
	assert.Equal(t, layout.StateDone, done.State)
	assert.Nil(t, done.Job.Lease)

	failed, err := Fail(dir, layout.RoleBuilder, "job-2", "w2", "testes falharam", t0)
	require.NoError(t, err)
	assert.Equal(t, layout.StateFailed, failed.State)
	assert.Equal(t, "testes falharam", failed.Job.Events[len(failed.Job.Events)-1].Note)

	_, err = Complete(dir, layout.RoleBuilder, "job-1", "w1", "", t0)
	assert.ErrorIs(t, err, ErrNotFound, "job já concluído não está mais em processing")
	_, err = Complete(dir, layout.RoleBuilder, "nao-existe", "w1", "", t0)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRecover_LeaseVencido(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-2", TypeFeatureBuild, PriorityNormal), t0.Add(time.Second))
	_, err := Claim(dir, layout.RoleBuilder, "w1", 5*time.Minute, t0)
	require.NoError(t, err)
	_, err = Claim(dir, layout.RoleBuilder, "w2", 60*time.Minute, t0)
	require.NoError(t, err)

	n, err := Recover(dir, layout.RoleBuilder, t0.Add(2*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 0, n, "lease válido não é recuperado")

	n, err = Recover(dir, layout.RoleBuilder, t0.Add(10*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 1, n, "só o lease vencido volta")
	pend, _ := List(dir, layout.RoleBuilder, layout.StatePending)
	require.Len(t, pend, 1)
	assert.Equal(t, "job-1", pend[0].Job.ID)
	assert.Nil(t, pend[0].Job.Lease)
	assert.Contains(t, pend[0].Job.Events[len(pend[0].Job.Events)-1].Note, "lease expirado")

	again, err := Claim(dir, layout.RoleBuilder, "w3", time.Minute, t0.Add(11*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, "job-1", again.Job.ID)
}

func TestRenew_EstendeLeaseEImpedeRecuperacao(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
	_, err := Claim(dir, layout.RoleBuilder, "w1", 5*time.Minute, t0)
	require.NoError(t, err)
	require.NoError(t, Renew(dir, layout.RoleBuilder, "job-1", "w1", 30*time.Minute, t0.Add(4*time.Minute)))
	n, err := Recover(dir, layout.RoleBuilder, t0.Add(20*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, Renew(dir, layout.RoleBuilder, "job-1", "outro", time.Minute, t0), ErrNotOwner)
}

func TestClaim_JSONCorrompidoVaiParaFailedSemTravarAFila(t *testing.T) {
	dir := t.TempDir()
	pd, _ := layout.QueueState(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, os.MkdirAll(pd, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pd, "0-0000000000000000001-quebrado.json"), []byte("{nao-e-json"), 0o644))
	mustEnqueue(t, dir, layout.RoleBuilder, newJob("job-ok", TypeFeatureBuild, PriorityNormal), t0)

	e, err := Claim(dir, layout.RoleBuilder, "w1", time.Minute, t0)
	require.NoError(t, err)
	assert.Equal(t, "job-ok", e.Job.ID, "o job corrompido não bloqueia o próximo")

	fd, _ := layout.QueueState(dir, layout.RoleBuilder, layout.StateFailed)
	assert.FileExists(t, filepath.Join(fd, "0-0000000000000000001-quebrado.json"))
	assert.FileExists(t, filepath.Join(fd, "0-0000000000000000001-quebrado.json.reason.txt"))
}

func TestEnqueue_Violacoes(t *testing.T) {
	dir := t.TempDir()
	mut := map[string]func(j *Job){
		"id com ..":             func(j *Job) { j.ID = "../fora" },
		"id com barra":          func(j *Job) { j.ID = "a/b" },
		"id com espaço":         func(j *Job) { j.ID = "a b" },
		"id vazio":              func(j *Job) { j.ID = "" },
		"id maiúsculo":          func(j *Job) { j.ID = "Job-1" },
		"tipo desconhecido":     func(j *Job) { j.Type = "deploy" },
		"tipo da outra fila":    func(j *Job) { j.Type = TypePRReview },
		"prioridade inválida":   func(j *Job) { j.Priority = "urgente" },
		"rework sem prioridade": func(j *Job) { j.Type = TypePRRework; j.Priority = PriorityNormal },
		"feature inválida":      func(j *Job) { j.Feature = "Auth OAuth" },
		"título com controle":   func(j *Job) { j.Title = "a\x1b[31mb" },
		"spec absoluta":         func(j *Job) { j.SpecPath = "/etc/passwd" },
		"spec com ..":           func(j *Job) { j.SpecPath = "../../x.md" },
		"adr com ..":            func(j *Job) { j.ADRRefs = []string{"a/../../b"} },
		"rule absoluta":         func(j *Job) { j.RulesRefs = []string{"C:\\Windows\\x"} },
		"branch com opção":      func(j *Job) { j.PR.Branch = "--upload-pack=x" },
		"base com ..":           func(j *Job) { j.PR.BaseBranch = "a/../b" },
		"pr number inválido":    func(j *Job) { n := 0; j.PR.Number = &n },
		"max fora da faixa":     func(j *Job) { j.MaxReworkAttempts = 11 },
		"attempt acima do max":  func(j *Job) { j.Attempt = 9 },
		"task id inválido":      func(j *Job) { j.TaskIDs = []string{"../x"} },
		"initiative inválida":   func(j *Job) { j.Initiative = "x/../y" },
		"feedback com controle": func(j *Job) { j.ReworkFeedback = []Feedback{{File: "a.go", Rule: "ARQ-01", Description: "x\x1b[2J"}} },
		"feedback acima do max": func(j *Job) {
			j.ReworkFeedback = []Feedback{{File: "a.go", Rule: "ARQ-01", Description: strings.Repeat("x", 3000)}}
		},
		"feedback com mil itens": func(j *Job) { j.ReworkFeedback = make([]Feedback, 60) },
	}
	for name, m := range mut {
		j := newJob("job-1", TypeFeatureBuild, PriorityNormal)
		m(&j)
		_, _, err := Enqueue(dir, layout.RoleBuilder, j, t0)
		require.Error(t, err, name)
		assert.ErrorIs(t, err, ErrInvalidJob, name)
	}
	pend, err := List(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, err)
	assert.Empty(t, pend, "nada inválido chega à fila")

	// Papel inválido e fila de discovery não aceitam jobs.
	for _, role := range []string{"", "../x", layout.RoleDiscovery, "Builder"} {
		_, _, err := Enqueue(dir, role, newJob("job-1", TypeFeatureBuild, PriorityNormal), t0)
		assert.Error(t, err, role)
	}
	_, err = Claim(dir, layout.RoleBuilder, "worker com espaço", time.Minute, t0)
	assert.Error(t, err)
	_, err = Find(dir, layout.RoleBuilder, "../x")
	assert.Error(t, err)
}

func TestReadJob_RecusaArquivoGrandeELink(t *testing.T) {
	dir := t.TempDir()
	pd, _ := layout.QueueState(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, os.MkdirAll(pd, 0o755))
	big := filepath.Join(pd, "1-0000000000000000001-grande.json")
	require.NoError(t, os.WriteFile(big, []byte(strings.Repeat("a", MaxJobBytes+1)), 0o644))
	_, err := readJob(big)
	assert.ErrorIs(t, err, ErrTooLarge)

	outside := filepath.Join(t.TempDir(), "alvo.json")
	require.NoError(t, os.WriteFile(outside, []byte("{}"), 0o644))
	link := filepath.Join(pd, "1-0000000000000000002-link.json")
	require.NoError(t, os.Symlink(outside, link))
	_, err = readJob(link)
	assert.ErrorIs(t, err, ErrInvalidJob, "link simbólico não é lido como job")
}

func TestEnqueue_JobGrandeEhRecusado(t *testing.T) {
	dir := t.TempDir()
	j := newJob("job-1", TypeFeatureBuild, PriorityNormal)
	for i := 0; i < 50; i++ {
		j.ReworkFeedback = append(j.ReworkFeedback, Feedback{File: "a.go", Rule: "ARQ-01", Description: strings.Repeat("x", 2000), Snippet: strings.Repeat("y", 2000), Fix: strings.Repeat("z", 2000)})
	}
	_, _, err := Enqueue(dir, layout.RoleBuilder, j, t0)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTooLarge)
}

func TestFind_AcheEmQualquerEstado(t *testing.T) {
	dir := t.TempDir()
	mustEnqueue(t, dir, layout.RoleReviewer, newJob("pr-9", TypePRReview, PriorityNormal), t0)
	e, err := Find(dir, layout.RoleReviewer, "pr-9")
	require.NoError(t, err)
	assert.Equal(t, layout.StatePending, e.State)
	_, err = Find(dir, layout.RoleBuilder, "pr-9")
	assert.ErrorIs(t, err, ErrNotFound, "cada papel tem a própria fila")
}

func TestOwnership_WorkerVazioOuLeaseAusenteNaoConcluem(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Enqueue(dir, layout.RoleBuilder, newJob("job-own", TypeFeatureBuild, PriorityNormal), t0)
	require.NoError(t, err)
	_, err = Claim(dir, layout.RoleBuilder, "w1", time.Minute, t0)
	require.NoError(t, err)
	_, err = Complete(dir, layout.RoleBuilder, "job-own", "", "x", t0)
	assert.ErrorIs(t, err, ErrNotOwner, "worker vazio não prova posse")
	_, err = Fail(dir, layout.RoleBuilder, "job-own", "", "x", t0)
	assert.ErrorIs(t, err, ErrNotOwner)
	_, err = UpdatePR(dir, layout.RoleBuilder, "job-own", "", 1, "https://x/1", "", t0)
	assert.ErrorIs(t, err, ErrNotOwner)
	for _, bad := range []string{"https://x/a b", "https://u:p@x/1", "http://", "ftp://x/1", "https://x/\x1b[2J"} {
		_, err = UpdatePR(dir, layout.RoleBuilder, "job-own", "w1", 1, bad, "", t0)
		assert.ErrorIs(t, err, ErrInvalidJob, bad)
	}
}

func TestEnqueue_CreatedAtForaDaFaixaERecusado(t *testing.T) {
	dir := t.TempDir()
	j := newJob("job-data", TypeFeatureBuild, PriorityNormal)
	j.CreatedAt = time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _, err := Enqueue(dir, layout.RoleBuilder, j, t0)
	assert.ErrorIs(t, err, ErrInvalidJob)
	j.CreatedAt = time.Now().Add(10 * 365 * 24 * time.Hour)
	_, _, err = Enqueue(dir, layout.RoleBuilder, j, t0)
	assert.ErrorIs(t, err, ErrInvalidJob)
}
