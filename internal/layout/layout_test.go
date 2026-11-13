package layout

import (
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathsEstaveis(t *testing.T) {
	d := "/proj"
	assert.Equal(t, filepath.Join(d, "sdd", ".sddrc"), Sddrc(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".sdd-version"), Version(d))
	assert.Equal(t, filepath.Join(d, "sdd", "memory", "constitution.md"), Constitution(d))
	assert.Equal(t, filepath.Join(d, "sdd", "memory", "progress.md"), Progress(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".metrics"), Metrics(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".queue", "queue.json"), LegacyQueueFile(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".runtime", "stations"), StationsDir(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".runtime", "activity.json"), ActivityFile(d))
	assert.Equal(t, filepath.Join(d, "sdd", "discovery"), Discovery(d))
	assert.Equal(t, filepath.Join(d, "sdd", "archive"), Archive(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".runtime"), Runtime(d))
}

func TestInitiative_Conformes(t *testing.T) {
	for _, id := range []string{"69-otimizacao-estrutura", "5ae2-curva-aprendizado", "1024-x-y"} {
		got, err := Initiative("/p", id)
		require.NoError(t, err, id)
		assert.Equal(t, filepath.Join("/p", "sdd", "discovery", id), got)
	}
	st, err := InitiativeStatus("/p", "69-x")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/p", "sdd", "discovery", "69-x", "status.md"), st)
	q, err := InitiativeQueue("/p", "69-x")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/p", "sdd", "discovery", "69-x", "queue.json"), q)
}

func TestInitiative_Violacoes(t *testing.T) {
	invalid := []string{
		"", "69", "..", "../etc", "69-x/../..", "/abs/69-x", "69-x\x00", "69-x\n", "69_x", "69-X",
		"69-x/y", `69-x\y`, "-69-x", "69--x", "69-", "69-x.md",
	}
	for _, id := range invalid {
		_, err := Initiative("/p", id)
		require.Error(t, err, "%q deveria ser recusado", id)
		assert.ErrorIs(t, err, ErrInvalidInitiative)
		_, err = InitiativeStatus("/p", id)
		assert.Error(t, err)
		_, err = InitiativeQueue("/p", id)
		assert.Error(t, err)
		_, err = InitiativeFeatures("/p", id)
		assert.Error(t, err)
		_, err = InitiativeFixes("/p", id)
		assert.Error(t, err)
	}
}

func TestFromBranch(t *testing.T) {
	cases := map[string]struct {
		id string
		ok bool
	}{
		"feat/69-otimizacao-estrutura": {"69-otimizacao-estrutura", true},
		"fix/12-corrige-algo":          {"12-corrige-algo", true},
		"feat/5ae2-curva-aprendizado":  {"5ae2-curva-aprendizado", true},
		"main":                         {"", false},
		"release/v3.0.1":               {"", false},
		"chore/release-2.4.0":          {"", false},
		"feat/semid":                   {"", false},
		"feat/69-x/extra":              {"", false},
		"feat/../69-x":                 {"", false},
		"":                             {"", false},
	}
	for branch, want := range cases {
		id, ok := FromBranch(branch)
		assert.Equal(t, want.ok, ok, branch)
		assert.Equal(t, want.id, id, branch)
	}
}

func gitInit(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	testutil.InitRepo(t, dir, branch)
	return dir
}

func TestCurrent(t *testing.T) {
	dir := gitInit(t, "feat/69-teste-de-layout")
	id, ok, err := Current(dir)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "69-teste-de-layout", id)

	main := gitInit(t, "main")
	id, ok, err = Current(main)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, id)
}

func TestCurrent_HeadDestacadoEForaDeRepositorio(t *testing.T) {
	dir := gitInit(t, "feat/69-teste-de-layout")
	testutil.Git(t, dir, "checkout", "-q", "--detach")
	id, ok, err := Current(dir)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, id)

	_, _, err = Current(t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "layout:")
}

func TestForgePaths(t *testing.T) {
	d := "/proj"
	assert.Equal(t, filepath.Join(d, ".forge", "config", "forge.json"), ForgeConfigFile(d))
	assert.Equal(t, filepath.Join(d, ".forge", "config", "notifications.json"), NotificationsFile(d))
	assert.Equal(t, filepath.Join(d, ".forge", "runtime"), ForgeRuntime(d))
	assert.Equal(t, filepath.Join(d, ".forge", "runtime", "worktrees"), Worktrees(d))
	assert.Equal(t, filepath.Join(d, ".forge", "runtime", "queues", "discovery", "inbox"), Inbox(d))
	assert.Equal(t, filepath.Join(d, ".forge", "runtime", "queues", "discovery", "archive"), InboxArchive(d))

	for _, role := range []string{RoleBuilder, RoleReviewer} {
		for _, st := range []string{StatePending, StateProcessing, StateDone, StateFailed} {
			p, err := QueueState(d, role, st)
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(d, ".forge", "runtime", "queues", role, st), p)
		}
	}
}

func TestQueueState_Violacoes(t *testing.T) {
	cases := [][2]string{
		{"", "pending"}, {"builder", ""}, {"../x", "pending"}, {"builder", "../pending"}, {"builder", "inbox"},
		{"reviewer", "archive"}, {"discovery", "pending"}, {"discovery", "done"}, {"Builder", "pending"}, {"builder", "pending/x"},
	}
	for _, c := range cases {
		_, err := QueueState("/p", c[0], c[1])
		require.Error(t, err, "%q/%q", c[0], c[1])
		assert.ErrorIs(t, err, ErrInvalidQueue)
	}
	assert.Equal(t, "", func() string { p, _ := QueueState("/p", "x", "y"); return p }())
}
