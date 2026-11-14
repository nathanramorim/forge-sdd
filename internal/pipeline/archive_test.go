package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoArchive_IniciativaCompleta(t *testing.T) {
	tempDir := t.TempDir()
	initDir := filepath.Join(tempDir, "sdd", "discovery", "79-clareza-pipeline")
	require.NoError(t, os.MkdirAll(initDir, 0o755))

	// Cria arquivos da iniciativa
	testutil.WriteFile(t, initDir, "criteria.md", "# Criteria\n")
	testutil.WriteFile(t, initDir, "status.md", "# Status\n")

	// Cria fila com todas as tarefas aprovadas
	q := queue.Queue{
		FeatureID:  "79",
		SpecStatus: "reviewed",
		Items: []queue.QueueItem{
			{TaskID: "79-01", Status: queue.StatusApproved},
			{TaskID: "79-02", Status: queue.StatusApproved},
		},
	}
	qData, err := json.Marshal(q)
	require.NoError(t, err)
	testutil.WriteFile(t, initDir, "queue.json", string(qData))

	now := time.Date(2026, 11, 14, 15, 0, 0, 0, time.UTC)
	res, err := AutoArchiveInitiative(tempDir, "79-clareza-pipeline", now)
	require.NoError(t, err)
	require.True(t, res.Archived, "iniciativa sem pendências deve ser arquivada")

	// 1. Verifica consolidação em sdd/archive/discovery/79-clareza-pipeline
	archiveDir := filepath.Join(tempDir, "sdd", "archive", "discovery", "79-clareza-pipeline")
	assert.FileExists(t, filepath.Join(archiveDir, "criteria.md"))
	assert.FileExists(t, filepath.Join(archiveDir, "status.md"))
	assert.FileExists(t, filepath.Join(archiveDir, "queue.json"))

	// 2. Verifica registro em sdd/memory/progress-log.md
	logContent, err := os.ReadFile(filepath.Join(tempDir, "sdd", "memory", "progress-log.md"))
	require.NoError(t, err)
	assert.Contains(t, string(logContent), "79-clareza-pipeline concluída e arquivada com sucesso")

	// 3. Verifica gravação de telemetria em sdd/.metrics/
	require.NotEmpty(t, res.TelemetryPath)
	assert.FileExists(t, res.TelemetryPath)

	metricData, err := os.ReadFile(res.TelemetryPath)
	require.NoError(t, err)
	assert.Contains(t, string(metricData), "approved")
	assert.Contains(t, string(metricData), `"criterio_atendido": true`)
}

func TestAutoArchive_ComTarefasPendentesNaoArquiva(t *testing.T) {
	tempDir := t.TempDir()
	initDir := filepath.Join(tempDir, "sdd", "discovery", "80-status-executivo")
	require.NoError(t, os.MkdirAll(initDir, 0o755))

	// Cria fila com uma tarefa ainda pendente
	q := queue.Queue{
		FeatureID:  "80",
		SpecStatus: "reviewed",
		Items: []queue.QueueItem{
			{TaskID: "80-01", Status: queue.StatusApproved},
			{TaskID: "80-02", Status: queue.StatusBuilding},
		},
	}
	qData, err := json.Marshal(q)
	require.NoError(t, err)
	testutil.WriteFile(t, initDir, "queue.json", string(qData))

	now := time.Date(2026, 11, 14, 15, 0, 0, 0, time.UTC)
	res, err := AutoArchiveInitiative(tempDir, "80-status-executivo", now)
	require.NoError(t, err)
	assert.False(t, res.Archived, "iniciativa com tarefas pendentes NÃO deve ser arquivada")
	assert.Contains(t, res.Reason, "ainda pendente")

	// Pasta de arquivo não deve ter sido criada
	archiveDir := filepath.Join(tempDir, "sdd", "archive", "discovery", "80-status-executivo")
	_, err = os.Stat(archiveDir)
	assert.True(t, os.IsNotExist(err))
}

func TestApplyVerdict_GatilhoAutoArchive(t *testing.T) {
	// Verifica a integração do término de task aprovada com o gatilho de auto-archive (CA-04)
	dir, job := setupProjectWithJob(t, layout.RoleReviewer)

	// Cria pasta da iniciativa compatível com job.Feature
	initDir := filepath.Join(dir, "sdd", "discovery", job.Feature)
	require.NoError(t, os.MkdirAll(initDir, 0o755))
	testutil.WriteFile(t, initDir, "criteria.md", "# Criteria\n")

	// Fila com todas as tarefas aprovadas
	q := queue.Queue{
		FeatureID:  job.Feature,
		SpecStatus: "reviewed",
		Items: []queue.QueueItem{
			{TaskID: "01", Status: queue.StatusApproved},
		},
	}
	qData, _ := json.Marshal(q)
	testutil.WriteFile(t, initDir, "queue.json", string(qData))

	// Reivindica o job para processamento antes de aplicar o veredito
	_, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", bus.DefaultLeaseTTL, now)
	require.NoError(t, err)

	deps, _ := notifyDeps(nil)
	v := queue.Verdict{
		Veredito: "conforme",
		Resumo:   "Tudo em conformidade com as regras e spec.",
	}

	res, err := ApplyVerdict(dir, job.ID, "forge-reviewer", v, deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, res.Action)

	// Verifica se auto-archive consolidou em sdd/archive
	archiveDir := filepath.Join(dir, "sdd", "archive", "discovery", job.Feature)
	assert.FileExists(t, filepath.Join(archiveDir, "criteria.md"))
}
