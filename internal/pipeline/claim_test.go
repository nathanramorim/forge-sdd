package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const claimInit = "70-alfa"

func claimRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	q := queue.Queue{FeatureID: "feat-70", Items: []queue.QueueItem{
		{TaskID: "feat-70-01-a", Status: queue.StatusQueued, Rules: []string{}, SkillsUsed: []string{}},
		{TaskID: "feat-70-02-b", Status: queue.StatusQueued, Rules: []string{}, SkillsUsed: []string{}},
	}}
	data, err := json.Marshal(q)
	require.NoError(t, err)
	testutil.WriteFile(t, dir, "sdd/discovery/"+claimInit+"/queue.json", string(data))
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", "plano")
	return dir
}

func TestClaimPelaExistenciaDaBranch(t *testing.T) {
	dir := claimRepo(t)
	require.NoError(t, ClaimFeature(dir, "feat/70-alfa-01", "main"))
	out := testutil.Git(t, dir, "branch", "--list", "feat/70-alfa-01")
	assert.Contains(t, out, "feat/70-alfa-01")

	// Branch já existente (aqui só a referência remota) também conta como claim feito.
	testutil.Git(t, dir, "update-ref", "refs/remotes/origin/feat/70-alfa-02", "main")
	assert.Error(t, ClaimFeature(dir, "feat/70-alfa-02", "main"))
	// Nome inválido nunca vira argumento do git.
	assert.Error(t, ClaimFeature(dir, "--delete", "main"))
}

func TestSegundoClaimDaMesmaFeatureFalha(t *testing.T) {
	dir := claimRepo(t)
	require.NoError(t, ClaimFeature(dir, "feat/70-alfa-01", "main"))
	err := ClaimFeature(dir, "feat/70-alfa-01", "main")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "já")
}

func TestDuasFeaturesNaoEscrevemNoMesmoArquivo(t *testing.T) {
	dir := claimRepo(t)
	pa, err := queue.StatePath(dir, claimInit, "feat-70-01-a")
	require.NoError(t, err)
	pb, err := queue.StatePath(dir, claimInit, "feat-70-02-b")
	require.NoError(t, err)
	require.NotEqual(t, pa, pb)
	require.NoError(t, os.MkdirAll(filepath.Dir(pa), 0o755))
	planBefore, err := os.ReadFile(filepath.Join(dir, "sdd/discovery", claimInit, "queue.json"))
	require.NoError(t, err)

	n, err := queue.SetTaskStatuses(dir, claimInit, []string{"70-01"}, queue.StatusBuilding)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	assert.FileExists(t, pa)
	assert.NoFileExists(t, pb, "a feature 02 não pode ser tocada pela 01")
	planAfter, err := os.ReadFile(filepath.Join(dir, "sdd/discovery", claimInit, "queue.json"))
	require.NoError(t, err)
	assert.Equal(t, string(planBefore), string(planAfter), "queue.json vira plano: não é regravado")
}

// Duas branches de feature mudam o estado cada uma no seu arquivo; o merge não conflita.
func TestEstadoDeFeaturesParalelasSemConflito(t *testing.T) {
	dir := claimRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd/discovery", claimInit, queue.StateDirName), 0o755))
	testutil.WriteFile(t, dir, "sdd/discovery/"+claimInit+"/state/.gitkeep", "")
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", "state")

	for _, c := range []struct{ branch, task string }{{"feat/70-alfa-01", "70-01"}, {"feat/70-alfa-02", "70-02"}} {
		require.NoError(t, ClaimFeature(dir, c.branch, "main"))
		testutil.Git(t, dir, "checkout", "-q", c.branch)
		n, err := queue.SetTaskStatuses(dir, claimInit, []string{c.task}, queue.StatusBuilding)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		testutil.Git(t, dir, "add", "-A")
		testutil.Git(t, dir, "commit", "-q", "-m", "estado "+c.task)
	}
	testutil.Git(t, dir, "checkout", "-q", "main")
	testutil.Git(t, dir, "merge", "-q", "--no-edit", "feat/70-alfa-01")
	testutil.Git(t, dir, "merge", "-q", "--no-edit", "feat/70-alfa-02")

	q, err := queue.LoadForInitiative(dir, claimInit)
	require.NoError(t, err)
	for _, it := range q.Items {
		assert.Equal(t, queue.StatusBuilding, it.Status, it.TaskID)
	}
}
