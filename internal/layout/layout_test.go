package layout

import (
	"os/exec"
	"path/filepath"
	"testing"

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
	assert.Equal(t, filepath.Join(d, "sdd", ".queue", "queue.json"), QueueFile(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".queue", "stations"), StationsDir(d))
	assert.Equal(t, filepath.Join(d, "sdd", ".queue", "activity.json"), ActivityFile(d))
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
	for _, args := range [][]string{
		{"init", "-q", "-b", branch},
		{"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
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
	out, err := exec.Command("git", "-C", dir, "checkout", "-q", "--detach").CombinedOutput()
	require.NoError(t, err, string(out))
	id, ok, err := Current(dir)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, id)

	_, _, err = Current(t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "layout:")
}
