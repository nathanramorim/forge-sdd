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
