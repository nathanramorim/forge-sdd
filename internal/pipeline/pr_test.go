package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const prTemplate = "## Spec\n\n## O que muda\n\n## Rollback\n"
const prBody = "## Spec\nx\n## O que muda\ny\n## Rollback\nz\n"

type fakeRun struct {
	calls     []string
	branch    string
	ahead     string
	openPRs   string
	createOut string
	reflog    string
	failOn    string
}

func (f *fakeRun) run(_ string, name string, args ...string) (string, error) {
	line := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, line)
	if f.failOn != "" && strings.HasPrefix(line, f.failOn) {
		return "", errors.New("falhou")
	}
	switch {
	case strings.HasPrefix(line, "git rev-parse --abbrev-ref HEAD"):
		return f.branch, nil
	case strings.HasPrefix(line, "git reflog"):
		return f.reflog, nil
	case strings.HasPrefix(line, "git rev-list --count"):
		return f.ahead, nil
	case strings.HasPrefix(line, "gh pr list"):
		return f.openPRs, nil
	case strings.HasPrefix(line, "gh pr create"):
		return f.createOut, nil
	}
	return "", nil
}

func (f *fakeRun) called(prefix string) bool {
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

// prProject monta um job do Builder em processamento na iniciativa 77 com a fila já em building.
func prProject(t *testing.T) (dir, jobID string) {
	t.Helper()
	dir = project(t, nil)
	testutil.WriteFile(t, dir, ".github/pull_request_template.md", prTemplate)
	testutil.WriteFile(t, dir, "corpo.md", prBody)
	testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/queue.json", `{"base_branch":"feat/88-modo","items":[{"task_id":"77-01","status":"building"}]}`)
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	e, err := bus.Claim(dir, layout.RoleBuilder, "forge-builder", 3600e9, now)
	require.NoError(t, err)
	return dir, e.Job.ID
}

func okGate(string, string, string, string) ([]string, error) { return nil, nil }

func TestPipelinePR_AbreComABaseDaIniciativaERegistra(t *testing.T) {
	dir, id := prProject(t)
	f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "3", createOut: "https://github.com/o/r/pull/42"}

	res, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate, Now: now})

	require.NoError(t, err)
	assert.Equal(t, 42, res.Number)
	assert.True(t, res.Created)
	assert.Equal(t, "feat/88-modo", res.Base)
	assert.Equal(t, queue.BaseFromQueue, res.BaseSource)
	assert.True(t, f.called("git push origin feat/77-auth-oauth:feat/77-auth-oauth"))
	assert.True(t, f.called("gh pr create --base feat/88-modo --head feat/77-auth-oauth"))
	assert.False(t, f.called("gh pr merge"))
	j, _, err := bus.Get(dir, layout.RoleBuilder, id)
	require.NoError(t, err)
	require.NotNil(t, j.PR.Number)
	assert.Equal(t, 42, *j.PR.Number)
	q, err := queue.LoadInitiative(dir, "77-auth-oauth")
	require.NoError(t, err)
	assert.Equal(t, queue.StatusReviewPending, q.Items[0].Status)
}

func TestPipelinePR_CorrigeABaseDePRExistente(t *testing.T) {
	dir, id := prProject(t)
	f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "1",
		openPRs: `[{"number":7,"url":"https://github.com/o/r/pull/7","baseRefName":"main"}]`}

	res, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate, Now: now})

	require.NoError(t, err)
	assert.True(t, res.BaseFixed)
	assert.Equal(t, 7, res.Number)
	assert.True(t, f.called("gh pr edit 7 --base feat/88-modo"))
	assert.False(t, f.called("gh pr create"))
}

func TestPipelinePR_Recusas(t *testing.T) {
	t.Run("corpo sem as seções do modelo", func(t *testing.T) {
		dir, id := prProject(t)
		testutil.WriteFile(t, dir, "curto.md", "## Spec\nx\n")
		f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "1"}
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "curto.md", Run: f.run, Gate: okGate})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rollback")
		assert.False(t, f.called("git push"))
	})
	t.Run("sem commit à frente da base", func(t *testing.T) {
		dir, id := prProject(t)
		f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "0"}
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "commit à frente")
		assert.False(t, f.called("git push"))
	})
	t.Run("Camada A reprovada", func(t *testing.T) {
		dir, id := prProject(t)
		f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "2"}
		bad := func(string, string, string, string) ([]string, error) { return []string{"[P-04] segredo"}, nil }
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: bad})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "P-04")
		assert.False(t, f.called("git push"))
	})
	t.Run("branch atual diferente da do job", func(t *testing.T) {
		dir, id := prProject(t)
		f := &fakeRun{branch: "main", ahead: "2"}
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate})
		require.Error(t, err)
		assert.False(t, f.called("git push"))
	})
	t.Run("corpo fora do projeto", func(t *testing.T) {
		dir, id := prProject(t)
		f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "2"}
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "../fora.md", Run: f.run, Gate: okGate})
		require.Error(t, err)
	})
	t.Run("push falha e nada é registrado", func(t *testing.T) {
		dir, id := prProject(t)
		f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "2", failOn: "git push"}
		_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate})
		require.Error(t, err)
		j, _, _ := bus.Get(dir, layout.RoleBuilder, id)
		assert.Nil(t, j.PR.Number)
	})
}

func TestPipelinePR_BranchIgualABaseERecusada(t *testing.T) {
	dir, id := prProject(t)
	testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/queue.json", `{"base_branch":"feat/77-auth-oauth","items":[]}`)
	f := &fakeRun{branch: "feat/77-auth-oauth", ahead: "2"}

	_, err := OpenPR(dir, PROptions{JobID: id, Worker: "forge-builder", BodyFile: "corpo.md", Run: f.run, Gate: okGate})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "FLX-02")
	assert.False(t, f.called("git push"), fmt.Sprint(f.calls))
}
