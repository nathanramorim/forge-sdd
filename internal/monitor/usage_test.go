package monitor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const resultLog = `{"type":"system","subtype":"init"}
linha que não é json
{"type":"result","num_turns":3,"duration_ms":4500,"total_cost_usd":0.5,"usage":{"input_tokens":10,"output_tokens":20,"cache_read_input_tokens":100,"cache_creation_input_tokens":7},"modelUsage":{"claude-sonnet-5":{}}}
{"type":"result","num_turns":2,"duration_ms":1600,"total_cost_usd":0.25,"usage":{"input_tokens":5,"output_tokens":6,"cache_read_input_tokens":50,"cache_creation_input_tokens":1},"modelUsage":{"claude-sonnet-5":{}}}
`

func writeLog(t *testing.T, dir, role, id, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(layout.Logs(dir), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(layout.Logs(dir), role+"-"+id+".log"), []byte(body), 0o644))
}

func TestUsoReal_SomaOsEventosResult(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "builder", "j1", resultLog)

	u := ReadJobUsage(dir, "builder", "j1", "")

	assert.Equal(t, MedicaoReal, u.Medicao)
	assert.Equal(t, 15, u.TokensInput)
	assert.Equal(t, 26, u.TokensOutput)
	assert.Equal(t, 150, u.TokensCacheRead)
	assert.Equal(t, 8, u.TokensCacheCreation)
	assert.Equal(t, 5, u.Turns)
	assert.Equal(t, 6, u.DurationSeconds)
	assert.Equal(t, "claude-sonnet-5", u.Model)
	assert.InDelta(t, 0.75, u.CostUSDTabela, 1e-9)
}

func TestUsoEstimado_SemResult(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "builder", "j2", `{"type":"assistant"}`+"\n")

	u := ReadJobUsage(dir, "builder", "j2", "")

	assert.Equal(t, MedicaoEstimado, u.Medicao)
	assert.Positive(t, u.TokensInput)
	assert.Zero(t, u.Turns)
}

func TestUsoEstimado_LogInexistente(t *testing.T) {
	u := ReadJobUsage(t.TempDir(), "builder", "nada", "")
	assert.Equal(t, MedicaoEstimado, u.Medicao)
}

func TestSumUsage_SeparaRealDeEstimado(t *testing.T) {
	ms := []SessionMetrics{
		{Medicao: MedicaoReal, TokensInput: 1, Turns: 2},
		{Medicao: MedicaoReal, TokensInput: 3, Turns: 1},
		{TokensInput: 100},
	}
	t1 := SumUsage(ms)
	assert.Equal(t, 4, t1.Real.TokensInput)
	assert.Equal(t, 3, t1.Real.Turns)
	assert.Equal(t, 1, t1.Estimado.Sessions)
	assert.True(t, BelongsToInitiative(SessionMetrics{Feature: "sdd/discovery/90-x/criteria.md"}, "90-x"))
	assert.False(t, BelongsToInitiative(SessionMetrics{Feature: "sdd/discovery/91-x/criteria.md"}, "90-x"))
}

func TestVersaoDoBinario_TelemetriaNaoTemVersaoFixa(t *testing.T) {
	data, err := os.ReadFile("session_record.go")
	require.NoError(t, err)
	assert.NotContains(t, string(data), "4.5.0")
}

func TestEstimateSessionUsage_SemFeature(t *testing.T) {
	dir := t.TempDir()
	u := EstimateSessionUsage(dir, "", nil)
	assert.Positive(t, u.TokensInput)
	assert.Positive(t, u.TokensOutput)
	assert.Equal(t, MedicaoEstimado, u.Medicao)
	expectedInput := EstimateTokensFromBytes(estimateFallbackInputBytes * 4)
	assert.Equal(t, expectedInput, u.TokensInput)
}

func TestEstimateSessionUsage_SemFiles(t *testing.T) {
	dir := t.TempDir()
	u := EstimateSessionUsage(dir, "", []string{})
	assert.Positive(t, u.TokensOutput)
	expectedOutput := EstimateTokensFromBytes(estimateFallbackOutputBytes)
	assert.Equal(t, expectedOutput, u.TokensOutput)
}

func TestEstimateSessionUsage_ComFeature(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "sdd", "discovery", "97-test", "features")
	require.NoError(t, os.MkdirAll(featDir, 0o755))
	featFile := filepath.Join(featDir, "feat-97-01-test.md")
	require.NoError(t, os.WriteFile(featFile, []byte(strings.Repeat("x", 4000)), 0o644))

	// Testa com caminho relativo direto
	u := EstimateSessionUsage(dir, "sdd/discovery/97-test/features/feat-97-01-test.md", nil)
	assert.Positive(t, u.TokensInput)
	assert.Equal(t, MedicaoEstimado, u.Medicao)
	// 4000 bytes × 4 (mesmo multiplicador de EstimateJobUsage) → EstimateTokensFromBytes(16000) = 4000 tokens
	assert.Equal(t, EstimateTokensFromBytes(4000*4), u.TokensInput)
}

func TestEstimateSessionUsage_ComFilesTouched(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.go"), []byte(strings.Repeat("x", 2000)), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.go"), []byte(strings.Repeat("x", 2000)), 0o644))

	u := EstimateSessionUsage(dir, "", []string{"file1.go", "file2.go"})
	// 4000 bytes total → EstimateTokensFromBytes(4000) = 1000 tokens
	assert.Equal(t, EstimateTokensFromBytes(4000), u.TokensOutput)
}
