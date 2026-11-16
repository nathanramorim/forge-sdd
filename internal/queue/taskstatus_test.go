package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testInitiative = "77-auth-oauth"

func writeInitiativeQueue(t *testing.T, dir string, items ...QueueItem) string {
	t.Helper()
	base, err := layout.Initiative(dir, testInitiative)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(base, 0o755))
	path := filepath.Join(base, layout.QueueFileName)
	data, err := json.Marshal(Queue{SpecStatus: "reviewed", Items: items})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
	return path
}

func readStatuses(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var q Queue
	require.NoError(t, json.Unmarshal(data, &q))
	out := map[string]string{}
	for _, it := range q.Items {
		out[it.TaskID] = it.Status
	}
	return out
}

func TestSetTaskStatuses_AprovaTasksDoJobPorIdEPrefixo(t *testing.T) {
	dir := t.TempDir()
	path := writeInitiativeQueue(t, dir,
		QueueItem{TaskID: "01", Status: StatusQueued},
		QueueItem{TaskID: "02-credenciais", Status: StatusChangesRequested},
		QueueItem{TaskID: "03", Status: StatusQueued},
	)

	n, err := SetTaskStatuses(dir, testInitiative, []string{"01", "02"}, StatusApproved)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, map[string]string{"01": StatusApproved, "02-credenciais": StatusApproved, "03": StatusQueued}, readStatuses(t, path))
}

func TestSetTaskStatuses_IdempotenteENaoRegrideAprovada(t *testing.T) {
	dir := t.TempDir()
	path := writeInitiativeQueue(t, dir,
		QueueItem{TaskID: "01", Status: StatusApproved},
		QueueItem{TaskID: "02", Status: StatusQueued},
	)

	n, err := SetTaskStatuses(dir, testInitiative, []string{"01", "02"}, StatusReviewPending)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "só a 02 muda; a 01 já aprovada não regride")
	assert.Equal(t, map[string]string{"01": StatusApproved, "02": StatusReviewPending}, readStatuses(t, path))

	n, err = SetTaskStatuses(dir, testInitiative, []string{"01", "02"}, StatusReviewPending)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "reaplicar não muda nada")
}

func TestSetTaskStatuses_SemEfeitoQuandoNaoHaFilaOuIniciativa(t *testing.T) {
	dir := t.TempDir()

	n, err := SetTaskStatuses(dir, testInitiative, []string{"01"}, StatusApproved)
	require.NoError(t, err, "fila ausente não é erro")
	assert.Zero(t, n)

	path := writeInitiativeQueue(t, dir, QueueItem{TaskID: "01", Status: StatusQueued})
	for name, c := range map[string]struct {
		initiative string
		ids        []string
	}{
		"sem task ids":        {testInitiative, nil},
		"sem iniciativa":      {"", []string{"01"}},
		"iniciativa inválida": {"../fora", []string{"01"}},
		"task desconhecida":   {testInitiative, []string{"99"}},
	} {
		n, err := SetTaskStatuses(dir, c.initiative, c.ids, StatusApproved)
		require.NoError(t, err, name)
		assert.Zero(t, n, name)
	}
	assert.Equal(t, map[string]string{"01": StatusQueued}, readStatuses(t, path))
}

func TestMatchesTask_IgnoraPrefixoFeatENuncaCasaNumeroMaior(t *testing.T) {
	cases := []struct {
		item, id string
		want     bool
	}{
		{"feat-88-01", "88-01", true},
		{"feat-88-01-modo-avancado", "88-01", true},
		{"feat-84-01-nome", "feat-84-01", true},
		{"88-01", "88-01", true},
		{"feat-88-10-nome", "88-1", false},
		{"feat-88-10", "88-1", false},
		{"feat-88-02", "88-01", false},
		{"feat-88-01", "", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, matchesTask(c.item, c.id), "%s × %s", c.item, c.id)
	}
}

func TestSetTaskStatuses_IdsDoJobCasamComIdsRealDaFila(t *testing.T) {
	dir := t.TempDir()
	path := writeInitiativeQueue(t, dir,
		QueueItem{TaskID: "feat-88-01-modo", Status: StatusQueued},
		QueueItem{TaskID: "feat-88-02", Status: StatusApproved},
		QueueItem{TaskID: "feat-88-10-outro", Status: StatusQueued},
	)

	n, err := SetTaskStatuses(dir, testInitiative, []string{"88-01", "88-02", "88-1"}, StatusApproved)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, map[string]string{"feat-88-01-modo": StatusApproved, "feat-88-02": StatusApproved, "feat-88-10-outro": StatusQueued}, readStatuses(t, path))

	n, err = SetTaskStatuses(dir, testInitiative, []string{"88-01"}, StatusApproved)
	require.NoError(t, err)
	assert.Zero(t, n, "idempotente")
}

func TestCountMatches(t *testing.T) {
	dir := t.TempDir()
	writeInitiativeQueue(t, dir, QueueItem{TaskID: "feat-88-01-x", Status: StatusApproved})

	n, err := CountMatches(dir, testInitiative, []string{"88-01"})
	require.NoError(t, err)
	assert.Equal(t, 1, n, "conta mesmo task já aprovada")

	n, err = CountMatches(dir, testInitiative, []string{"99-01"})
	require.NoError(t, err)
	assert.Zero(t, n)
}
