package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTmuxDeliverer struct {
	captured string
	err      error
}

func (m mockTmuxDeliverer) CapturePane(target string) (string, error) {
	return m.captured, m.err
}
func (mockTmuxDeliverer) LoadBuffer(name string, data []byte) error { return nil }
func (mockTmuxDeliverer) PasteBuffer(name, target string) error    { return nil }
func (mockTmuxDeliverer) SendKeys(target string, keys ...string) error {
	return nil
}
func (mockTmuxDeliverer) SetPaneTitle(target, title string) error { return nil }

func TestFormatLiveStatus_FallbackLogSemTmux(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	// Cria HEAD apontando para feat/80-status para simular git repo sem gastar processo git init
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/feat/80-status\n"), 0o644))

	logsDir := filepath.Join(dir, ".forge", "runtime", "logs")
	require.NoError(t, os.MkdirAll(logsDir, 0o755))
	logFile := filepath.Join(logsDir, "builder-job-123.log")
	require.NoError(t, os.WriteFile(logFile, []byte("Linha antiga\nLinha 1 do builder\nLinha 2 do builder\n"), 0o644))

	sddDir := filepath.Join(dir, "sdd", ".queue")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	queueFile := filepath.Join(sddDir, "queue.json")
	require.NoError(t, os.WriteFile(queueFile, []byte(`{"items":[{"task_id":"80-01","title":"Tarefa pronta","status":"approved"},{"task_id":"80-02","title":"Tarefa pendente","status":"queued"}]}`), 0o644))

	gitFn := func(d string, max int) ([]string, error) {
		return []string{"abc1234 feat: commit recente", "def5678 fix: commit anterior"}, nil
	}

	start := time.Now()
	out, err := FormatLiveStatusWith(dir, mockTmuxDeliverer{err: os.ErrNotExist}, gitFn)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond, "deve executar em menos de 100ms (CA-01)")

	assert.Contains(t, out, "### Status da Esteira —")
	assert.Contains(t, out, "1. Ao Vivo no Painel")
	assert.Contains(t, out, "Linha 1 do builder")
	assert.Contains(t, out, "Linha 2 do builder")

	assert.Contains(t, out, "2. Commits Recentes no Git:")
	assert.Contains(t, out, "abc1234 feat: commit recente")
	assert.Contains(t, out, "def5678 fix: commit anterior")

	assert.Contains(t, out, "3. Fila de Tarefas:")
	assert.Contains(t, out, "• [x] 80-01: Tarefa pronta")
	assert.Contains(t, out, "• [ ] 80-02: Tarefa pendente")

	assert.Contains(t, out, "Próximo passo: Acompanhar o Builder")

	// Orçamento de tokens: palavras aproximadas
	words := len(strings.Fields(out))
	assert.Less(t, words, 200, "deve respeitar o orçamento de tokens (< 250 tokens)")
}

func TestFormatLiveStatus_ComTodasTarefasAprovadas(t *testing.T) {
	dir := t.TempDir()

	sddDir := filepath.Join(dir, "sdd", ".queue")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	queueFile := filepath.Join(sddDir, "queue.json")
	require.NoError(t, os.WriteFile(queueFile, []byte(`{"items":[{"task_id":"80-01","title":"Tarefa 1","status":"approved"},{"task_id":"80-02","title":"Tarefa 2","status":"approved"}]}`), 0o644))

	gitFn := func(d string, max int) ([]string, error) {
		return nil, nil
	}

	out, err := FormatLiveStatusWith(dir, nil, gitFn)
	require.NoError(t, err)
	assert.Contains(t, out, "Próximo passo: /archive ou novo /discovery")
	assert.Contains(t, out, "(nenhum commit recente)")
}
