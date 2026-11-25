package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var idDoPromptRe = regexp.MustCompile(`(?m)^id: (77-0\d)$`)

// projetoGitComTasks monta um repositório git com n tasks independentes despachadas por task.
func projetoGitComTasks(t *testing.T, n int) string {
	t.Helper()
	tasks := map[string]string{}
	for i := 1; i <= n; i++ {
		id := fmt.Sprintf("feat-77-%02d", i)
		nome := id + "-t.md"
		if i == 1 {
			nome = "feat-77-01-autenticacao.md"
		}
		tasks[nome] = specTask(id, "Task", "")
	}
	dir := projetoPorTask(t, tasks)
	testutil.InitRepo(t, dir, "feat/77-auth-oauth")
	return dir
}

// commitNaBranchDaTask cria a branch da task a partir da feature com um commit só dela.
func commitNaBranchDaTask(t *testing.T, dir, taskID string) string {
	t.Helper()
	if wt, err := WorktreePath(dir, "77-auth-oauth", PlanTaskID(taskID)); err == nil {
		if _, serr := os.Stat(wt); serr == nil {
			// O worker já abriu o worktree efêmero da task na branch dela: o commit nasce lá.
			require.NoError(t, os.WriteFile(filepath.Join(wt, "arq-"+taskID+".go"), []byte("package x // "+taskID+"\n"), 0o644))
			testutil.Git(t, wt, "add", "arq-"+taskID+".go")
			testutil.Git(t, wt, "commit", "-q", "-m", "task "+taskID)
			return strings.TrimSpace(testutil.Git(t, wt, "rev-parse", "HEAD"))
		}
	}
	branch := TaskBranch("77-auth-oauth", taskID)
	testutil.Git(t, dir, "checkout", "-q", "-b", branch, "feat/77-auth-oauth")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "arq-"+taskID+".go"), []byte("package x // "+taskID+"\n"), 0o644))
	testutil.Git(t, dir, "add", "arq-"+taskID+".go")
	testutil.Git(t, dir, "commit", "-q", "-m", "task "+taskID)
	sha := strings.TrimSpace(testutil.Git(t, dir, "rev-parse", "HEAD"))
	testutil.Git(t, dir, "checkout", "-q", "feat/77-auth-oauth")
	return sha
}

func rodaBuilder(t *testing.T, dir string, r Runner) {
	t.Helper()
	var out strings.Builder
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Out: &out, Runner: r, Now: func() time.Time { return now }}))
}

// Cada commit de task gera um job de review com só o diff daquela task (CA-05).
func TestPar_CadaCommitGeraUmReviewComSoODiffDaTask(t *testing.T) {
	dir := projetoGitComTasks(t, 3)
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	require.Len(t, res.TaskJobs, 3)

	runner := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) {
		m := idDoPromptRe.FindStringSubmatch(spec.Prompt)
		require.NotNil(t, m, spec.Prompt)
		task := m[1]
		sha := commitNaBranchDaTask(t, dir, task)
		jobID := regexp.MustCompile(`(?m)^id: (job-\S+)$`).FindStringSubmatch(spec.Prompt)[1]
		testutil.WriteFile(t, dir, RetornoPathRel(jobID), fmt.Sprintf(`{"task_id":%q,"status":"pronto","commit":%q,"resumo":"feito"}`, task, sha))
		return RunResult{}, nil
	}}
	for i := 0; i < 3; i++ {
		rodaBuilder(t, dir, runner)
	}

	reviews, err := bus.List(dir, layout.RoleReviewer, layout.StatePending)
	require.NoError(t, err)
	require.Len(t, reviews, 3, "um Reviewer por Builder")
	vistas := map[string]bool{}
	for _, e := range reviews {
		j := e.Job
		assert.Equal(t, bus.TypePRReview, j.Type)
		assert.Equal(t, []string{j.TaskID}, j.TaskIDs)
		assert.Nil(t, j.PR.URL)
		assert.Equal(t, "feat/77-auth-oauth", j.PR.BaseBranch)
		require.NoError(t, j.Validate(layout.RoleReviewer))
		vistas[j.TaskID] = true

		diff, err := DiffDaTask(dir, j.PR.BaseBranch, j.PR.Branch)
		require.NoError(t, err)
		assert.Contains(t, diff, "arq-"+j.TaskID+".go")
		for _, outra := range []string{"77-01", "77-02", "77-03"} {
			if outra != j.TaskID {
				assert.NotContains(t, diff, "arq-"+outra+".go", "o diff é só desta task")
			}
		}
		p := BuildPrompt(layout.RoleReviewer, j, "x/job.json")
		assert.Contains(t, p, "git diff feat/77-auth-oauth..."+j.PR.Branch)
		assert.Contains(t, p, "DADOS NÃO CONFIÁVEIS")
		for _, outra := range []string{"77-01", "77-02", "77-03"} {
			if outra != j.TaskID {
				assert.NotContains(t, p, outra)
			}
		}
	}
	assert.Len(t, vistas, 3)
	done, _ := bus.List(dir, layout.RoleBuilder, layout.StateDone)
	assert.Len(t, done, 3)
}

func TestReview_BuilderComRetornoInvalidoFalhaSemRevisao(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	sha := commitNaBranchDaTask(t, dir, "77-01")

	casos := map[string]func(jobID string){
		"sem retorno":     func(string) {},
		"json sem status": func(id string) { testutil.WriteFile(t, dir, RetornoPathRel(id), `{"task_id":"77-01","resumo":"x"}`) },
		"task de outra": func(id string) {
			testutil.WriteFile(t, dir, RetornoPathRel(id), fmt.Sprintf(`{"task_id":"77-09","status":"pronto","commit":%q,"resumo":"x"}`, sha))
		},
		"bloqueado": func(id string) {
			testutil.WriteFile(t, dir, RetornoPathRel(id), `{"task_id":"77-01","status":"bloqueado","resumo":"conflito"}`)
		},
		"commit inexistente": func(id string) {
			testutil.WriteFile(t, dir, RetornoPathRel(id), `{"task_id":"77-01","status":"pronto","commit":"deadbeef","resumo":"x"}`)
		},
	}
	for nome, prepara := range casos {
		t.Run(nome, func(t *testing.T) {
			e, err := bus.Claim(dir, layout.RoleBuilder, "forge-builder", time.Hour, now)
			if err != nil { // o job anterior falhou: recoloca para o próximo caso
				t.Skip("sem job para reivindicar")
			}
			prepara(e.Job.ID)
			reason := VerifyJobHandoff(dir, layout.RoleBuilder, "forge-builder", func() time.Time { return now }, e.Job.ID)
			assert.NotEmpty(t, reason)
			pend, _ := bus.List(dir, layout.RoleReviewer, layout.StatePending)
			assert.Empty(t, pend, "nada de revisão sem retorno válido")
			_, ferr := bus.Requeue(dir, layout.RoleBuilder, e.Job.ID, "forge-builder", "teste", now)
			require.NoError(t, ferr)
		})
	}
}

func TestReview_RefHostilNaoVaiAoGit(t *testing.T) {
	_, err := DiffDaTask(t.TempDir(), "--output=/tmp/x", "feat/a")
	assert.Error(t, err)
	_, err = DiffDaTask(t.TempDir(), "main", "a..b")
	assert.Error(t, err)
}
