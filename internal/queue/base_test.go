package queue

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeGit(out string, err error) GitFunc {
	return func(string, ...string) (string, error) { return out, err }
}

func initiativeWithQueue(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	base := filepath.Join(dir, "sdd", "discovery", "90-x")
	require.NoError(t, os.MkdirAll(base, 0o755))
	if body != "" {
		require.NoError(t, os.WriteFile(filepath.Join(base, "queue.json"), []byte(body), 0o644))
	}
	return dir
}

func TestBaseDaIniciativa_OrdemDasFontes(t *testing.T) {
	created := fakeGit("branch: Created from feat/88-modo\n", nil)

	dir := initiativeWithQueue(t, `{"base_branch":"feat/da-fila","items":[]}`)
	b, src := ResolveBase(dir, "90-x", "feat/da-flag", "main", created)
	assert.Equal(t, []string{"feat/da-flag", BaseFromFlag}, []string{b, src})

	b, src = ResolveBase(dir, "90-x", "", "main", created)
	assert.Equal(t, []string{"feat/da-fila", BaseFromQueue}, []string{b, src})

	dir = initiativeWithQueue(t, `{"items":[]}`)
	b, src = ResolveBase(dir, "90-x", "", "main", created)
	assert.Equal(t, []string{"feat/88-modo", BaseFromGit}, []string{b, src})

	b, src = ResolveBase(dir, "90-x", "", "main", fakeGit("branch: Created from HEAD", nil))
	assert.Equal(t, []string{"main", BaseFromConfig}, []string{b, src})

	b, src = ResolveBase(dir, "90-x", "", "main", fakeGit("", errors.New("sem reflog")))
	assert.Equal(t, []string{"main", BaseFromConfig}, []string{b, src})
}

func TestBaseDaIniciativa_RecusaNomeInvalido(t *testing.T) {
	dir := initiativeWithQueue(t, `{"base_branch":"--upload-pack=x","items":[]}`)

	b, src := ResolveBase(dir, "90-x", "a b", "main", fakeGit("", errors.New("x")))

	assert.Equal(t, []string{"main", BaseFromConfig}, []string{b, src})
	require.Error(t, ValidBranchName("-rf"))
	require.Error(t, ValidBranchName("a..b"))
	require.NoError(t, ValidBranchName("feat/88-modo-pipeline"))
}

func TestBaseDaIniciativa_OrigemFicaNaFila(t *testing.T) {
	dir := initiativeWithQueue(t, `{"items":[]}`)

	require.NoError(t, SetOrigem(dir, "90-x", OrigemTocaAi))

	q, err := LoadInitiative(dir, "90-x")
	require.NoError(t, err)
	assert.Equal(t, OrigemTocaAi, q.Origem)
}
