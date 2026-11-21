package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const initEstado = "79-clareza"

func iniciativaComStatus(t *testing.T, status string) string {
	t.Helper()
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/discovery/"+initEstado+"/features/feat-79-01-a.md", "# a\n")
	if status != "" {
		testutil.WriteFile(t, dir, "sdd/discovery/"+initEstado+"/status.md", status)
	}
	return dir
}

// O marcador do Gate 1 passa a ser a linha `Gate 1: aprovado` do status.md; queue.json segue valendo
// para iniciativa antiga (sem regressão).
func TestGoalSpecApprovedPeloStatusDaPrincipal(t *testing.T) {
	cfg := config.DefaultForgeConfig()

	com := iniciativaComStatus(t, "# Status\n\nGate 1: aprovado em 2026-11-20\n")
	assert.True(t, goalSpecApproved(com, initEstado, cfg), "marcador no status.md, sem queue.json")

	negrito := iniciativaComStatus(t, "# Status\n\n**Gate 1:** aprovado\n")
	assert.True(t, goalSpecApproved(negrito, initEstado, cfg))

	sem := iniciativaComStatus(t, "# Status\n\nGate 1 pendente\n")
	assert.False(t, goalSpecApproved(sem, initEstado, cfg), "sem marcador e sem queue.json")

	prosa := iniciativaComStatus(t, "# Status\n\nainda falta o Gate 1: aprovado pelo dono\n")
	assert.False(t, goalSpecApproved(prosa, initEstado, cfg), "o marcador vale só como linha própria")

	semStatus := iniciativaComStatus(t, "")
	assert.False(t, goalSpecApproved(semStatus, initEstado, cfg))

	legada := iniciativaComStatus(t, "# Status\n")
	testutil.WriteFile(t, legada, "sdd/discovery/"+initEstado+"/queue.json", `{"items":[]}`)
	assert.True(t, goalSpecApproved(legada, initEstado, cfg), "iniciativa antiga: queue.json aprova como hoje")

	assert.False(t, goalSpecApproved(com, "../fora", cfg), "id inválido")

	// status.md que é link simbólico não é lido (SEG-02).
	lnk := iniciativaComStatus(t, "")
	alvo := filepath.Join(t.TempDir(), "x.md")
	require.NoError(t, os.WriteFile(alvo, []byte("Gate 1: aprovado\n"), 0o644))
	require.NoError(t, os.Symlink(alvo, filepath.Join(lnk, "sdd", "discovery", initEstado, "status.md")))
	assert.False(t, goalSpecApproved(lnk, initEstado, cfg))
}

func iniciativaComEstado(t *testing.T, status string) string {
	t.Helper()
	dir := t.TempDir()
	base := "sdd/discovery/" + initEstado
	testutil.WriteFile(t, dir, base+"/status.md", "# Status\n")
	testutil.WriteFile(t, dir, base+"/queue.json", `{"items":[{"task_id":"79-01","status":"queued"}]}`)
	require.NoError(t, queue.WriteFeatureState(dir, initEstado, queue.FeatureState{TaskID: "79-01", Status: status}))
	return dir
}

// archive roda da principal e enxerga o estado por feature, não o status do plano.
func TestArchiveRodaDaPrincipal(t *testing.T) {
	at := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)

	dir := iniciativaComEstado(t, queue.StatusApproved)
	assert.Equal(t, []string{initEstado}, ListArchiveCandidates(dir), "estado approved libera, mesmo com o plano queued")
	res, err := AutoArchiveInitiative(dir, initEstado, at, onBranch("main"))
	require.NoError(t, err)
	require.True(t, res.Archived, res.Reason)
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "discovery", initEstado))

	pend := iniciativaComEstado(t, queue.StatusReviewPending)
	assert.Empty(t, ListArchiveCandidates(pend))
	res, err = AutoArchiveInitiative(pend, initEstado, at, onBranch("main"))
	require.NoError(t, err)
	assert.False(t, res.Archived)
	assert.Contains(t, res.Reason, "ainda pendente")

	outra := iniciativaComEstado(t, queue.StatusApproved)
	res, err = AutoArchiveInitiative(outra, initEstado, at, onBranch("chore/qualquer"))
	require.NoError(t, err)
	assert.False(t, res.Archived)
	assert.Contains(t, res.Reason, "principal")

	legada := archiveFixture(t, queue.StatusApproved)
	res, err = AutoArchiveInitiative(legada, "79-clareza", at, onBranch("feat/79-clareza"))
	require.NoError(t, err)
	assert.True(t, res.Archived, "a branch da iniciativa segue valendo (compatibilidade)")
}

// O archive decide pelo que está na árvore da principal; a fila de uma branch da iniciativa que
// ficou para trás (com task pendente) não é lida.
func TestBuilderNaoLeAFilaDaBranchDaIniciativa(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	base := "sdd/discovery/" + initEstado
	testutil.WriteFile(t, dir, base+"/status.md", "# Status\n")
	testutil.WriteFile(t, dir, base+"/queue.json", `{"items":[{"task_id":"79-01","status":"approved"}]}`)
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", "plano")
	testutil.Git(t, dir, "checkout", "-q", "-b", "feat/79-clareza")
	testutil.WriteFile(t, dir, base+"/queue.json", `{"items":[{"task_id":"79-01","status":"review_pending"}]}`)
	testutil.Git(t, dir, "commit", "-q", "-am", "fila velha da branch")
	testutil.Git(t, dir, "checkout", "-q", "main")

	res, err := AutoArchiveInitiative(dir, initEstado, time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC), nil)
	require.NoError(t, err)
	require.True(t, res.Archived, res.Reason)
}
