package pipeline

import (
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Conforme: agente morto por sinal volta sozinho à fila uma vez, sem gastar tentativa; morto de novo,
// falha com o sinal e a causa provável no motivo (antes era só "saiu com código -1").
func TestWorker_MortoPorSinalRetomaUmaVezEDepoisExplica(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	runner := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) {
		return RunResult{ExitCode: -1, Sinal: "killed"}, nil
	}}
	var outcomes []string
	out := runOnce(t, dir, layout.RoleBuilder, runner, func(j bus.Job, o string) { outcomes = append(outcomes, o) })
	assert.Contains(t, out, "voltou à fila e será retomado")
	assert.Equal(t, []string{OutcomeDeferred, OutcomeFailed}, outcomes, "uma retomada, depois falha")
	assert.Len(t, runner.prompts, 2)

	e, err := bus.Find(dir, layout.RoleBuilder, "job-1")
	require.NoError(t, err)
	assert.Equal(t, layout.StateFailed, e.State)
	assert.Equal(t, 1, e.Job.Attempt, "a retomada não gasta tentativa")
	nota := e.Job.Events[len(e.Job.Events)-1].Note
	assert.Contains(t, nota, "sinal killed (provável falta de memória")
	assert.Contains(t, nota, "já foi retomado uma vez")
	assert.Equal(t, 0, deferrals(e.Job), "retomada por sinal não conta como limite de uso")
}

// Negativo: saída com código comum (erro do agente) não é retomada.
func TestWorker_CodigoComumNaoRetoma(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	runner := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) { return RunResult{ExitCode: 2}, nil }}
	runOnce(t, dir, layout.RoleBuilder, runner, nil)
	assert.Len(t, runner.prompts, 1)
	e, err := bus.Find(dir, layout.RoleBuilder, "job-1")
	require.NoError(t, err)
	assert.Equal(t, layout.StateFailed, e.State)
	assert.Contains(t, e.Job.Events[len(e.Job.Events)-1].Note, "saiu com código 2")
}

func TestNomeDaSessaoPrincipal(t *testing.T) {
	dir := t.TempDir()
	assert.NotEqual(t, SessionDiscovery, NomeDaSessaoPrincipal(dir), "sem .sddrc usa a pasta")
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"barbearia app"}`)
	assert.Equal(t, "barbearia-app", NomeDaSessaoPrincipal(dir))
}
