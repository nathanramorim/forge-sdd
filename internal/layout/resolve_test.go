package layout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoComIniciativas cria um repositório na branch dada com as pastas sdd/discovery/<nome>.
func repoComIniciativas(t *testing.T, branch string, nomes ...string) string {
	t.Helper()
	t.Setenv("FORGE_INICIATIVA", "")
	SetOverride("")
	t.Cleanup(func() { SetOverride("") })
	dir := gitInit(t, branch)
	for _, n := range nomes {
		require.NoError(t, os.MkdirAll(filepath.Join(Discovery(dir), n), 0o755))
	}
	return dir
}

func TestResolveIniciativaPorPrefixoNumerico(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-11-parecer-com-head-sha", "106-pr-draft-revisao-e-ci", "10-outra", "107-x")
	id, path, err := ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Equal(t, "106-pr-draft-revisao-e-ci", id)
	assert.Equal(t, filepath.Join(Discovery(dir), id), path)

	aid, apath, ok := ActiveInitiative(dir)
	assert.True(t, ok)
	assert.Equal(t, id, aid)
	assert.Equal(t, path, apath)
	assert.Equal(t, filepath.Join(path, QueueFileName), QueuePath(dir))
	sp, ok := StatusPath(dir)
	assert.True(t, ok)
	assert.Equal(t, filepath.Join(path, StatusFileName), sp)
	assert.NoError(t, RequireInitiative(dir))
}

func TestResolvePrefixoAmbiguoFalha(t *testing.T) {
	dir := repoComIniciativas(t, "fix/106-01-algo", "106-a", "106-b", "107-c")
	id, path, err := ResolveInitiative(dir)
	require.ErrorIs(t, err, ErrAmbiguousInitiative)
	assert.Empty(t, id)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "106-a")
	assert.Contains(t, err.Error(), "106-b")
	assert.NotContains(t, err.Error(), "107-c")
	assert.Contains(t, err.Error(), "--iniciativa")
	assert.ErrorIs(t, RequireInitiative(dir), ErrAmbiguousInitiative)

	// adaptadores: ambiguidade vira ok=false, sem escolher uma
	_, _, ok := ActiveInitiative(dir)
	assert.False(t, ok)
	_, ok = StatusPath(dir)
	assert.False(t, ok)
	assert.Equal(t, LegacyQueueFile(dir), QueuePath(dir))
}

func TestBranchSemPrefixoNumericoMantemComportamento(t *testing.T) {
	for _, branch := range []string{"main", "fix/ordem-da-fila-e-prioridade", "chore/106-x", "feat/semid"} {
		dir := repoComIniciativas(t, branch, "106-a", "ordem-da-fila")
		id, path, err := ResolveInitiative(dir)
		require.NoError(t, err, branch)
		assert.Empty(t, id, branch)
		assert.Empty(t, path, branch)
		assert.NoError(t, RequireInitiative(dir), branch)
		assert.Equal(t, LegacyQueueFile(dir), QueuePath(dir), branch)
	}
	// numérico sem pasta correspondente: legado, como hoje
	dir := repoComIniciativas(t, "feat/999-zzz", "106-a")
	id, _, err := ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestResolveRecusaBranchHostil(t *testing.T) {
	dir := repoComIniciativas(t, "main", "106-a", "10-b", "1-c")
	hostis := []string{
		"feat/106-a/../../etc",
		"feat/106-a/extra",
		"feat/../106-a",
		"feat/106-..",
		"feat/106-a\nfeat/x",
		"feat/106-a\x00",
		"feat/106-A",
		"feat/x106-a",
		"feat/ 106-a",
		"chore/106-a",
		"feat/1066-a",
		"feat/1060-a",
		"feat/1234567-a",
		"feat/106-" + strings.Repeat("a", 201),
		"",
	}
	for _, b := range hostis {
		id, err := initiativeOfBranch(dir, b)
		assert.NoError(t, err, "%q", b)
		assert.Empty(t, id, "%q", b)
	}
	// 10 não casa 106 (segmento inteiro) e 1 não casa 10
	id, err := initiativeOfBranch(dir, "feat/10-qualquer")
	require.NoError(t, err)
	assert.Equal(t, "10-b", id)
	id, err = initiativeOfBranch(dir, "feat/1-qualquer")
	require.NoError(t, err)
	assert.Equal(t, "1-c", id)
	// limite de 200: exatamente 200 é aceito
	ok := "feat/106-" + strings.Repeat("a", 200-len("feat/106-"))
	require.Len(t, ok, 200)
	id, err = initiativeOfBranch(dir, ok)
	require.NoError(t, err)
	assert.Equal(t, "106-a", id)
	// pasta de nome inválido nunca vira id
	require.NoError(t, os.MkdirAll(filepath.Join(Discovery(dir), "106-Ruim"), 0o755))
	id, err = initiativeOfBranch(dir, "feat/106-x")
	require.NoError(t, err)
	assert.Equal(t, "106-a", id)
}

func TestBranchPorIniciativaNaoRegride(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-pr-draft-revisao-e-ci", "106-pr-draft-revisao-e-ci", "106-outra")
	id, path, err := ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Equal(t, "106-pr-draft-revisao-e-ci", id)
	assert.Equal(t, filepath.Join(Discovery(dir), id), path)
	// FromBranch segue puro
	fid, ok := FromBranch("feat/106-11-parecer")
	assert.True(t, ok)
	assert.Equal(t, "106-11-parecer", fid)
}

func TestPrecedenciaFlagVariavelBranch(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-11-x", "106-a", "107-b", "108-c")
	id, _, err := ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Equal(t, "106-a", id)

	t.Setenv("FORGE_INICIATIVA", "107-b")
	id, _, err = ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Equal(t, "107-b", id, "variável vence a branch")

	SetOverride("108-c")
	id, path, err := ResolveInitiative(dir)
	require.NoError(t, err)
	assert.Equal(t, "108-c", id, "flag vence a variável")
	assert.Equal(t, filepath.Join(Discovery(dir), "108-c"), path)

	// inválida ou inexistente: erro, sem cair na branch
	for _, bad := range []string{"../x", "999-nada", "106"} {
		SetOverride(bad)
		id, _, err = ResolveInitiative(dir)
		assert.Error(t, err, bad)
		assert.Empty(t, id, bad)
		assert.Error(t, RequireInitiative(dir), bad)
	}
	SetOverride("")
	t.Setenv("FORGE_INICIATIVA", "999-nada")
	_, _, err = ResolveInitiative(dir)
	assert.Error(t, err)

	// a flag resolve a ambiguidade
	dir2 := repoComIniciativas(t, "feat/106-1-x", "106-a", "106-b")
	SetOverride("106-b")
	id, _, err = ResolveInitiative(dir2)
	require.NoError(t, err)
	assert.Equal(t, "106-b", id)
}

func TestListInitiativesMarcaAtualEmBranchPorFeature(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-11-x", "106-a", "107-b")
	list, err := ListInitiatives(dir, false)
	require.NoError(t, err)
	marcadas := map[string]bool{}
	for _, it := range list {
		marcadas[it.ID] = it.Current
	}
	assert.True(t, marcadas["106-a"])
	assert.False(t, marcadas["107-b"])

	// ambiguidade: ninguém é marcado, sem erro
	dir2 := repoComIniciativas(t, "feat/106-11-x", "106-a", "106-b")
	list, err = ListInitiatives(dir2, false)
	require.NoError(t, err)
	for _, it := range list {
		assert.False(t, it.Current, it.ID)
	}
}
