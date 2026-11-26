package pipeline

import (
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Conforme: fim do /discovery manda a spec ao Reviewer em autonomia normal; aprovada, avisa no inbox
// e NÃO despacha ao Builder (quem decide é a pessoa; quem despacha é o /split-features).
func TestRevisarDiscovery_AprovadaVoltaParaAPessoa(t *testing.T) {
	dir := advancedProject(t, false)
	job, _, err := RevisarDiscovery(dir, "77-auth-oauth", now)
	require.NoError(t, err)
	assert.True(t, isRevisaoDiscovery(*job))
	require.Len(t, reviewerPending(t, dir), 1)

	e, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, now)
	require.NoError(t, err)
	deps, _ := notifyDeps(nil)
	res, err := ApplySpecVerdict(dir, e.Job.ID, "forge-reviewer", conforme(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)
	require.NotNil(t, res.Notification)
	assert.Contains(t, res.Notification.Message, "/split-features")
	assert.Empty(t, builderPending(t, dir), "aprovação não despacha sozinha")
}

// Negativo: reprovada pede ajuste com o comando certo; sem tasks não há o que bloquear.
func TestRevisarDiscovery_AjusteEEscalada(t *testing.T) {
	dir := advancedProject(t, false)
	_, _, err := RevisarDiscovery(dir, "77-auth-oauth", now)
	require.NoError(t, err)
	e, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, now)
	require.NoError(t, err)
	deps, _ := notifyDeps(nil)
	res, err := ApplySpecVerdict(dir, e.Job.ID, "forge-reviewer", violation(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionRework, res.Action)
	assert.Contains(t, res.Notification.Message, "forge revisar-spec 77-auth-oauth")

	_, _, err = RevisarDiscovery(dir, "99-nao-existe", now)
	assert.Error(t, err)
}
