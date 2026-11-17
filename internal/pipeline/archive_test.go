package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func onBranch(branch string) queue.GitFunc {
	return func(string, ...string) (string, error) { return branch, nil }
}

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
	res, err := AutoArchiveInitiative(tempDir, "79-clareza-pipeline", now, onBranch("feat/79-clareza-pipeline"))
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
	res, err := AutoArchiveInitiative(tempDir, "80-status-executivo", now, onBranch("feat/80-status-executivo"))
	require.NoError(t, err)
	assert.False(t, res.Archived, "iniciativa com tarefas pendentes NÃO deve ser arquivada")
	assert.Contains(t, res.Reason, "ainda pendente")

	// Pasta de arquivo não deve ter sido criada
	archiveDir := filepath.Join(tempDir, "sdd", "archive", "discovery", "80-status-executivo")
	_, err = os.Stat(archiveDir)
	assert.True(t, os.IsNotExist(err))
}

func TestApplyVerdict_GatilhoAutoArchive_OrigemTocaAi(t *testing.T) {
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
	q.Origem = queue.OrigemTocaAi
	qData, _ := json.Marshal(q)
	testutil.WriteFile(t, initDir, "queue.json", string(qData))
	archiveGit = onBranch(layout.InitiativeBranch(job.Feature))
	t.Cleanup(func() { archiveGit = nil })

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

	// O archive automático (origem toca-ai) move a pasta para sdd/archive
	archiveDir := filepath.Join(dir, "sdd", "archive", "discovery", job.Feature)
	assert.FileExists(t, filepath.Join(archiveDir, "criteria.md"))
	assert.NoDirExists(t, initDir)
}

func TestArchive_AoAprovar_SemOrigemSoAvisa(t *testing.T) {
	dir, job := setupProjectWithJob(t, layout.RoleReviewer)
	initDir := filepath.Join(dir, "sdd", "discovery", job.Feature)
	require.NoError(t, os.MkdirAll(initDir, 0o755))
	testutil.WriteFile(t, initDir, "criteria.md", "# Criteria\n")
	qData, _ := json.Marshal(queue.Queue{Items: []queue.QueueItem{{TaskID: "01", Status: queue.StatusApproved}}})
	testutil.WriteFile(t, initDir, "queue.json", string(qData))
	archiveGit = onBranch(layout.InitiativeBranch(job.Feature))
	t.Cleanup(func() { archiveGit = nil })
	_, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", bus.DefaultLeaseTTL, now)
	require.NoError(t, err)
	deps, _ := notifyDeps(nil)

	_, err = ApplyVerdict(dir, job.ID, "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "ok"}, deps, now)

	require.NoError(t, err)
	assert.DirExists(t, initDir, "sem origem toca-ai, nada é movido")
	var found bool
	for _, it := range mustInbox(t, dir) {
		if strings.Contains(it.Title, "pronta para arquivar") {
			found = true
		}
	}
	assert.True(t, found, "o inbox deve avisar que a iniciativa está pronta para arquivar")
}

func archiveFixture(t *testing.T, status string) string {
	t.Helper()
	dir := t.TempDir()
	initDir := filepath.Join(dir, "sdd", "discovery", "79-clareza")
	require.NoError(t, os.MkdirAll(filepath.Join(initDir, "features"), 0o755))
	testutil.WriteFile(t, initDir, "status.md", "# Status\n")
	testutil.WriteFile(t, initDir, "features/feat-79-01-a.md", "# a\n")
	testutil.WriteFile(t, initDir, "queue.json", `{"items":[{"task_id":"79-01","status":"`+status+`"}]}`)
	testutil.WriteFile(t, dir, "sdd/.metrics/session-2026-11-14T100000Z.json",
		`{"feature":"sdd/discovery/79-clareza/criteria.md","phase":"job-1","agent_path":["builder"],"tokens_input":10,"tokens_output":20,"turns":4,"medicao":"real"}`)
	return dir
}

func TestArchive_MoveIdempotente(t *testing.T) {
	dir := archiveFixture(t, queue.StatusApproved)
	at := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	git := onBranch("feat/79-clareza")

	res, err := AutoArchiveInitiative(dir, "79-clareza", at, git)

	require.NoError(t, err)
	require.True(t, res.Archived, res.Reason)
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "discovery", "79-clareza"), "a origem deixa de existir")
	assert.FileExists(t, filepath.Join(dir, "sdd", "archive", "discovery", "79-clareza", "features", "feat-79-01-a.md"))
	metric, err := os.ReadFile(res.TelemetryPath)
	require.NoError(t, err)
	assert.Contains(t, string(metric), `"tokens_input": 10`)
	assert.Contains(t, string(metric), `"turns": 4`)

	again, err := AutoArchiveInitiative(dir, "79-clareza", at.Add(time.Hour), git)

	require.NoError(t, err)
	assert.False(t, again.Archived)
	assert.Contains(t, again.Reason, "já arquivada")
	logData, _ := os.ReadFile(filepath.Join(dir, "sdd", "memory", "progress-log.md"))
	assert.Equal(t, 1, strings.Count(string(logData), "79-clareza concluída"))
	files, _ := os.ReadDir(filepath.Join(dir, "sdd", ".metrics"))
	assert.Len(t, files, 2, "a segunda execução não grava outra telemetria")
}

func TestArchive_MoveIdempotente_RecusaForaDaBranchEComPendencia(t *testing.T) {
	dir := archiveFixture(t, queue.StatusApproved)
	res, err := AutoArchiveInitiative(dir, "79-clareza", now, onBranch("main"))
	require.NoError(t, err)
	assert.False(t, res.Archived)
	assert.Contains(t, res.Reason, "não é a da iniciativa")
	assert.DirExists(t, filepath.Join(dir, "sdd", "discovery", "79-clareza"))

	pend := archiveFixture(t, queue.StatusReviewPending)
	res, err = AutoArchiveInitiative(pend, "79-clareza", now, onBranch("feat/79-clareza"))
	require.NoError(t, err)
	assert.False(t, res.Archived)
	assert.Contains(t, res.Reason, "ainda pendente")
}

func TestArchive_MoveIdempotente_SobrepoeCopiaAntiga(t *testing.T) {
	dir := archiveFixture(t, queue.StatusApproved)
	testutil.WriteFile(t, dir, "sdd/archive/discovery/79-clareza/status.md", "# velho\n")

	res, err := AutoArchiveInitiative(dir, "79-clareza", now, onBranch("feat/79-clareza"))

	require.NoError(t, err)
	require.True(t, res.Archived)
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "discovery", "79-clareza"))
	data, _ := os.ReadFile(filepath.Join(dir, "sdd", "archive", "discovery", "79-clareza", "status.md"))
	assert.Equal(t, "# Status\n", string(data))
}

func TestArchive_MoveIdempotente_FixPendenteBloqueiaEFixFeitoNao(t *testing.T) {
	dir := archiveFixture(t, queue.StatusApproved)
	fixes := filepath.Join(dir, "sdd", "discovery", "79-clareza", "fixes")
	testutil.WriteFile(t, fixes, "fix-79-01-a.md", "# fix\n\n**Status:** `done`\n")
	testutil.WriteFile(t, fixes, "fix-79-02-b.md", "# fix\n\n**Status:** `ready`\n")
	git := onBranch("feat/79-clareza")

	res, err := AutoArchiveInitiative(dir, "79-clareza", now, git)
	require.NoError(t, err)
	assert.False(t, res.Archived)
	assert.Contains(t, res.Reason, "fix fix-79-02 ainda pendente")

	testutil.WriteFile(t, fixes, "fix-79-02-b.md", "# fix\n\n**Status:** `ready`\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "discovery", "79-clareza", "queue.json"),
		[]byte(`{"items":[{"task_id":"79-01","status":"approved"},{"task_id":"fix-79-02","status":"approved"}]}`), 0o644))
	res, err = AutoArchiveInitiative(dir, "79-clareza", now, git)
	require.NoError(t, err)
	assert.True(t, res.Archived, "fix aprovado na fila não bloqueia: %s", res.Reason)
}
