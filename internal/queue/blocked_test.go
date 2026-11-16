package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockedNaoBloqueiaIndependentes(t *testing.T) {
	dir := t.TempDir()
	path := writeInitiativeQueue(t, dir,
		QueueItem{TaskID: "feat-88-01", Status: StatusQueued},
		QueueItem{TaskID: "feat-88-03", Status: StatusQueued},
	)
	n, err := BlockTasks(dir, testInitiative, []string{"feat-88-01"}, "spec escalada")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, StatusBlocked, readStatuses(t, path)["feat-88-01"])

	q := &Queue{Items: []QueueItem{
		{TaskID: "feat-88-01", Status: StatusBlocked},
		{TaskID: "feat-88-03", Status: StatusQueued},
	}}
	el, err := EligibleTasks(q, 0)
	require.NoError(t, err)
	require.Len(t, el, 1)
	assert.Equal(t, "feat-88-03", el[0].TaskID)
}

func TestBlockedBloqueiaDependentes(t *testing.T) {
	q := &Queue{Items: []QueueItem{
		{TaskID: "feat-88-01", Status: StatusBlocked},
		{TaskID: "feat-88-02", Status: StatusQueued, DependsOn: []string{"feat-88-01"}},
	}}
	el, err := EligibleTasks(q, 0)
	require.NoError(t, err)
	assert.Empty(t, el)
}

func TestBlockedNaoRegrideAprovada(t *testing.T) {
	dir := t.TempDir()
	path := writeInitiativeQueue(t, dir, QueueItem{TaskID: "feat-88-01", Status: StatusApproved})
	n, err := BlockTasks(dir, testInitiative, []string{"feat-88-01"}, "x")
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Equal(t, StatusApproved, readStatuses(t, path)["feat-88-01"])
}

func TestBlockedPainelMostraMotivo(t *testing.T) {
	out := FormatExecutivePanel(&Queue{Items: []QueueItem{{TaskID: "t", Title: "T", Status: StatusBlocked, BlockedReason: "limite"}}})
	assert.Contains(t, out, "bloqueada: limite")
}
