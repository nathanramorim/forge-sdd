package queue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stInit = "80-beta"

func stateRepo(t *testing.T, comState bool) string {
	t.Helper()
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "feat/80-beta")
	spec := filepath.Join("sdd", "discovery", stInit, "features", "feat-80-01.md")
	writeSpec(t, dir, spec)
	require.NoError(t, Enqueue(dir, "feat-80", "feat/80-beta", "80-01", "t", spec, nil))
	if comState {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", stInit, StateDirName), 0o755))
	}
	return dir
}

func planBytes(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "sdd", "discovery", stInit, "queue.json"))
	require.NoError(t, err)
	return string(b)
}

func TestFilaAntigaSemStateFuncionaComoHoje(t *testing.T) {
	dir := stateRepo(t, false)
	n, err := SetTaskStatuses(dir, stInit, []string{"80-01"}, StatusReviewPending)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Contains(t, planBytes(t, dir), `"review_pending"`)
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "discovery", stInit, StateDirName))
	q, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, StatusReviewPending, q.Items[0].Status)
}

func TestMutadoresGravamEstadoPorFeature(t *testing.T) {
	dir := stateRepo(t, true)
	antes := planBytes(t, dir)

	n, err := BlockTasks(dir, stInit, []string{"80-01"}, "sem rede")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, antes, planBytes(t, dir), "queue.json não muda")

	st, err := ReadFeatureState(dir, stInit, "feat-80-01")
	require.NoError(t, err)
	require.NotNil(t, st)
	assert.Equal(t, StatusBlocked, st.Status)
	assert.Equal(t, "sem rede", st.BlockedReason)
	assert.Equal(t, StateSchemaVersion, st.Schema)

	q, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, StatusBlocked, q.Items[0].Status)

	// Aprovada nunca regride, também pelo estado.
	st.Status = StatusApproved
	require.NoError(t, WriteFeatureState(dir, stInit, *st))
	n, err = SetTaskStatuses(dir, stInit, []string{"80-01"}, StatusQueued)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestNextECompleteUsamEstadoEReusamAttempt(t *testing.T) {
	dir := stateRepo(t, true)
	antes := planBytes(t, dir)
	it, err := Next(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, it.Attempts)
	st, err := ReadFeatureState(dir, stInit, "80-01")
	require.NoError(t, err)
	assert.Equal(t, StatusBuilding, st.Status)
	assert.Equal(t, 1, st.Attempt, "Attempt é o contador da fila, não um novo")

	require.NoError(t, Complete(dir, "80-01"))
	st, err = ReadFeatureState(dir, stInit, "80-01")
	require.NoError(t, err)
	assert.Equal(t, StatusReviewPending, st.Status)
	assert.Equal(t, 1, st.Attempt)
	assert.Equal(t, antes, planBytes(t, dir))
}

func TestEstadoInvalidoERecusado(t *testing.T) {
	cases := map[string]string{
		"campo extra":     `{"schema":1,"task_id":"80-01","status":"queued","attempt":0,"x":1}`,
		"status invalido": `{"schema":1,"task_id":"80-01","status":"voando","attempt":0}`,
		"schema errado":   `{"schema":9,"task_id":"80-01","status":"queued","attempt":0}`,
		"attempt neg":     `{"schema":1,"task_id":"80-01","status":"queued","attempt":-1}`,
		"task divergente": `{"schema":1,"task_id":"80-02","status":"queued","attempt":0}`,
		"sha ruim":        `{"schema":1,"task_id":"80-01","status":"queued","attempt":0,"head_sha":"../x"}`,
		"json quebrado":   `{`,
	}
	for nome, conteudo := range cases {
		t.Run(nome, func(t *testing.T) {
			dir := stateRepo(t, true)
			testutil.WriteFile(t, dir, "sdd/discovery/"+stInit+"/state/80-01.json", conteudo)
			_, err := ReadFeatureState(dir, stInit, "80-01")
			assert.Error(t, err)
			_, err = Load(dir)
			assert.Error(t, err, "Load recusa o estado inválido")
		})
	}
}

func TestEstadoRecusaLinkSimbolicoEIdInvalido(t *testing.T) {
	dir := stateRepo(t, true)
	alvo := filepath.Join(t.TempDir(), "fora.json")
	require.NoError(t, os.WriteFile(alvo, []byte(`{}`), 0o644))
	require.NoError(t, os.Symlink(alvo, filepath.Join(dir, "sdd", "discovery", stInit, StateDirName, "80-01.json")))
	_, err := ReadFeatureState(dir, stInit, "80-01")
	assert.Error(t, err)
	assert.Error(t, WriteFeatureState(dir, stInit, FeatureState{TaskID: "80-01", Status: StatusQueued}))

	_, err = StatePath(dir, stInit, "../etc/passwd")
	assert.Error(t, err)
	_, err = StatePath(dir, "../x", "80-01")
	assert.Error(t, err)
}

func TestSaveNaoGravaEstadoNoPlano(t *testing.T) {
	dir := stateRepo(t, true)
	_, err := BlockTasks(dir, stInit, []string{"80-01"}, "x")
	require.NoError(t, err)
	require.NoError(t, SetBase(dir, "main"))
	assert.NotContains(t, planBytes(t, dir), `"blocked"`)
}
