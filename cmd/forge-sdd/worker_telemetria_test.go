package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jobDaTask(id, task string) bus.Job {
	return bus.Job{ID: id, Type: tipoTaskBuild, Initiative: "122-x", TaskIDs: []string{task},
		SpecPath: "sdd/discovery/122-x/features/" + task + ".md"}
}

func argDe(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func TestTelemetriaPorTask_ArgsUsamSpecDaTaskEPapel(t *testing.T) {
	u := monitor.JobUsage{TokensInput: 10, Medicao: monitor.MedicaoReal}
	args := sessionRecordArgs("/p", "builder", jobDaTask("j1", "feat-122-03"), "feat-122-03", "done", u)
	assert.Equal(t, "sdd/discovery/122-x/features/feat-122-03.md", argDe(args, "--feature"))
	assert.Equal(t, "builder", argDe(args, "--phase"))
	assert.Equal(t, "real", argDe(args, "--medicao"))
	assert.Equal(t, "approved", argDe(args, "--outcome"))
}

func TestTelemetriaPorTask_JobPorFeatureMantemContrato(t *testing.T) {
	j := bus.Job{ID: "job-1", Type: bus.TypeFeatureBuild, TaskIDs: []string{"a", "b"}, SpecPath: "sdd/x.md"}
	assert.Equal(t, "", taskIDDoJob(j))
	args := sessionRecordArgs("/p", "builder", j, "", "failed", monitor.JobUsage{})
	assert.Equal(t, "job-1", argDe(args, "--phase"))
	assert.Equal(t, "blocked", argDe(args, "--outcome"))
	// Job por task com mais de uma task não é válido: sem id de task.
	assert.Equal(t, "", taskIDDoJob(bus.Job{Type: tipoTaskBuild, TaskIDs: []string{"a", "b"}}))
	assert.Equal(t, "a", taskIDDoJob(bus.Job{Type: tipoTaskBuild, TaskIDs: []string{"a"}}))
}

func TestSessionRecordPorTask_UmRegistroPorTaskEPapelComUsoReal(t *testing.T) {
	dir := t.TempDir()
	writeTestSddrc(t, dir, true)
	require.NoError(t, os.MkdirAll(layout.Logs(dir), 0o755))
	result := `{"type":"result","num_turns":3,"duration_ms":5000,"total_cost_usd":0.5,"usage":{"input_tokens":100,"output_tokens":20,"cache_read_input_tokens":7,"cache_creation_input_tokens":3},"modelUsage":{"claude-x":{}}}` + "\n"
	pares := []struct{ role, job, task string }{
		{"builder", "jb-1", "feat-122-03"}, {"reviewer", "jr-1", "feat-122-03"}, {"builder", "jb-2", "feat-122-07"},
	}
	for _, p := range pares {
		require.NoError(t, os.WriteFile(filepath.Join(layout.Logs(dir), p.role+"-"+p.job+".log"), []byte("ruído\n"+result), 0o644))
		j := jobDaTask(p.job, p.task)
		u := monitor.ReadJobUsage(dir, p.role, j.ID, j.SpecPath)
		require.Equal(t, monitor.MedicaoReal, u.Medicao)
		rootCmd.SetArgs(sessionRecordArgs(dir, p.role, j, p.task, "done", u))
		require.NoError(t, rootCmd.Execute())
	}
	files, err := filepath.Glob(filepath.Join(dir, "sdd", ".metrics", "session-*.json"))
	require.NoError(t, err)
	require.Len(t, files, 3, "um registro por task e papel")
	vistos := map[string]bool{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		require.NoError(t, err)
		var m monitor.SessionMetrics
		require.NoError(t, json.Unmarshal(raw, &m))
		assert.Equal(t, "forge-sdd/metrics/1.0", m.Schema)
		assert.Equal(t, "real", m.Medicao)
		assert.Equal(t, 100, m.TokensInput)
		assert.Equal(t, 7, m.TokensCacheRead)
		vistos[m.Feature+"|"+m.Phase] = true
	}
	assert.Len(t, vistos, 3)
	assert.True(t, vistos["sdd/discovery/122-x/features/feat-122-03.md|reviewer"])
}

func TestTelemetriaPorTask_SemEventoResultFicaEstimado(t *testing.T) {
	dir := t.TempDir()
	j := jobDaTask("j9", "feat-122-03")
	u := monitor.ReadJobUsage(dir, "builder", j.ID, j.SpecPath)
	assert.Equal(t, monitor.MedicaoEstimado, u.Medicao)
}
