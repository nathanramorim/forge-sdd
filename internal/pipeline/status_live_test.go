package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatLiveStatus_LeOLogMaisRecente(t *testing.T) {
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
	out, err := FormatLiveStatusWith(dir, gitFn)
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

	out, err := FormatLiveStatusWith(dir, gitFn)
	require.NoError(t, err)
	assert.Contains(t, out, "Próximo passo: /archive ou novo /discovery")
	assert.Contains(t, out, "(nenhum commit recente)")
}

func TestFormatRichDashboard(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/feat/80-status\n"), 0o644))

	sddDir := filepath.Join(dir, "sdd", ".queue")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	queueFile := filepath.Join(sddDir, "queue.json")
	require.NoError(t, os.WriteFile(queueFile, []byte(`{"items":[{"task_id":"80-01","title":"Tarefa Concluida","status":"approved"},{"task_id":"80-02","title":"Tarefa em Andamento","status":"in_progress"}]}`), 0o644))

	metricsDir := filepath.Join(dir, "sdd", ".metrics")
	require.NoError(t, os.MkdirAll(metricsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(metricsDir, "session-1.json"), []byte(`{"agent_path":["builder"],"tokens_input":1200,"tokens_output":300}`), 0o644))

	gitFn := func(d string, max int) ([]string, error) {
		return []string{"abc1234 feat: painel vivo implementado"}, nil
	}

	out, err := FormatRichDashboard(dir, gitFn, nil)
	require.NoError(t, err)

	assert.Contains(t, out, "feat/80-status")
	assert.Contains(t, out, "PROCESSAMENTO & CONSUMO POR SESSÃO")
	assert.Contains(t, out, "PROGRESSO DA ESTEIRA (DAG)")
	assert.Contains(t, out, "50% (1/2 tasks)")
	assert.Contains(t, out, "[✓] 80-01: Tarefa Concluida")
	assert.Contains(t, out, "[▶] 80-02: Tarefa em Andamento")
	assert.Contains(t, out, "COMMITS RECENTES NO GIT")
	assert.Contains(t, out, "abc1234 feat: painel vivo implementado")
}

func TestRunLiveDashboard_Once(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/feat/80-status\n"), 0o644))

	var buf strings.Builder
	err := RunLiveDashboard(context.Background(), dir, &buf, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "PROCESSAMENTO & CONSUMO POR SESSÃO")
}

func TestFormatRichDashboard_LimitaTarefasEmFilaGrande(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/feat/80-status\n"), 0o644))

	sddDir := filepath.Join(dir, "sdd", ".queue")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))

	// Cria fila com 24 tarefas simulando barbearia-app
	var itemsJSON strings.Builder
	itemsJSON.WriteString(`{"items":[`)
	for i := 1; i <= 24; i++ {
		if i > 1 {
			itemsJSON.WriteString(",")
		}
		status := "queued"
		if i == 3 {
			status = "in_progress"
		}
		itemsJSON.WriteString(fmt.Sprintf(`{"task_id":"feat-%02d","title":"Tarefa %d","status":"%s"}`, i, i, status))
	}
	itemsJSON.WriteString(`]}`)
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "queue.json"), []byte(itemsJSON.String()), 0o644))

	out, err := FormatRichDashboard(dir, nil, nil)
	require.NoError(t, err)

	assert.Contains(t, out, "0% (0/24 tasks)")
	assert.Contains(t, out, "[▶] feat-03: Tarefa 3")
	assert.Contains(t, out, "... e mais 18 tarefas (use 'forge status' para a lista completa)")
}

func TestAoVivoSemJSONCru(t *testing.T) {
	log := `{"type":"assistant","message":{"content":[{"type":"text","text":"lendo a spec"},{"type":"tool_use","name":"Read","input":{"file_path":"/x/y/criteria.md"}}]}}` + "\n" +
		`{"type":"result","usage":{"input_tokens":1}}` + "\n" +
		`{"type":"assistant","message":{"con` + "\n" +
		"saída em texto comum\n"

	lines := extractLastInformativeLines(log, 5)

	require.NotEmpty(t, lines)
	for _, l := range lines {
		assert.NotContains(t, l, `{"`)
	}
	assert.Contains(t, strings.Join(lines, "\n"), "lendo a spec")
	assert.Equal(t, "saída em texto comum", lines[len(lines)-1])
}
