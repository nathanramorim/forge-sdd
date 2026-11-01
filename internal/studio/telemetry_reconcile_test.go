package studio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStudioServer_BootSanitizesOrphanActivities(t *testing.T) {
	tempDir := t.TempDir()

	// Prepara 3 atividades antigas em "running"
	_ = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationBuilder,
		Tool:    "run_command",
		Action:  "go test",
		Status:  "running",
	})
	_ = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationPMArchitect,
		Tool:    "read_file",
		Action:  "spec.md",
		Status:  "running",
	})
	_ = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationPlanner,
		Tool:    "edit_file",
		Action:  "plan.md",
		Status:  "running",
	})

	// Inicializa o Studio
	srv, err := NewServer(Config{Dir: tempDir})
	require.NoError(t, err)
	defer srv.Stop()

	// 1. Verifica se no disco foram sanitizadas para "success"
	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.Len(t, items, 3)
	for _, it := range items {
		assert.Equal(t, "success", it.Status)
		assert.NotEmpty(t, it.CompletedAt)
	}

	// 2. Verifica se a API /api/state reflete status limpo
	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var snap StudioSnapshot
	err = json.Unmarshal(rec.Body.Bytes(), &snap)
	require.NoError(t, err)

	for _, act := range snap.RecentActivity {
		assert.NotEqual(t, "running", act.Status)
	}
}

func TestStudioServer_StandbyTransitionReconciliation(t *testing.T) {
	tempDir := t.TempDir()

	srv, err := NewServer(Config{Dir: tempDir})
	require.NoError(t, err)
	defer srv.Stop()

	// Registra atividade em "running" para a estação builder
	err = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
		Station: queue.StationBuilder,
		Tool:    "run_command",
		Action:  "building binary",
		Status:  "running",
	})
	require.NoError(t, err)

	// Simula sessão builder transicionando para processing e depois para standby
	pool := srv.Pool()
	require.NotNil(t, pool)

	err = pool.TransitionState(queue.StationBuilder, acp.SessionStateProcessing, "feat-66-task")
	require.NoError(t, err)

	// Transiciona de volta para Standby
	err = pool.TransitionState(queue.StationBuilder, acp.SessionStateStandby, "task finished")
	require.NoError(t, err)

	// Aguarda um pequeno instante para os callbacks de listener processarem
	time.Sleep(50 * time.Millisecond)

	// Verifica se a atividade running foi reconciliada
	items, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, items)
	assert.Equal(t, "success", items[0].Status)
	assert.NotEmpty(t, items[0].CompletedAt)
}

func TestStudioServer_StandbyStationSnapshotSanitization(t *testing.T) {
	tempDir := t.TempDir()

	srv, err := NewServer(Config{Dir: tempDir})
	require.NoError(t, err)
	defer srv.Stop()

	// Injeta diretamente um arquivo com status running sem disparar reconciliação em disco
	actPath := tempDir + "/sdd/.queue/activity.json"
	rawItems := []queue.ActivityItem{
		{
			ID:        "act-raw",
			Station:   queue.StationLider,
			Tool:      "run_command",
			Action:    "silent test",
			Status:    "running",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	_ = queue.WriteJSONAtomic(actPath, rawItems, 0o644)

	// Obtém snapshot via Aggregator com pool ativo (estação líder está em standby no pool)
	snap, err := srv.Aggregator.GetSnapshot()
	require.NoError(t, err)

	for _, act := range snap.RecentActivity {
		if act.Station == queue.StationLider {
			assert.Equal(t, "success", act.Status, "estação em standby não pode emitir running")
		}
	}
}

// Concorrência sob race detector
func TestStudioServer_ConcurrentActivityAndState(t *testing.T) {
	tempDir := t.TempDir()

	srv, err := NewServer(Config{Dir: tempDir})
	require.NoError(t, err)
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 10)

	// Goroutine gravando atividades
	go func() {
		for i := 0; i < 30; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				_ = queue.RecordActivityDetailed(tempDir, queue.ActivityItem{
					Station: queue.StationBuilder,
					Tool:    "run_command",
					Action:  "step",
					Status:  "running",
				})
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Goroutine sanitizando
	go func() {
		for i := 0; i < 15; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				_, _ = queue.SanitizeOrphanActivities(tempDir)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Goroutine lendo snapshot
	for g := 0; g < 3; g++ {
		go func() {
			for i := 0; i < 20; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := srv.Aggregator.GetSnapshot()
					if err != nil {
						errChan <- err
						return
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}

	select {
	case err := <-errChan:
		t.Fatalf("erro durante concorrência: %v", err)
	case <-time.After(1 * time.Second):
		// Sucesso
	}
}

func TestStudioServer_TaskExecute_DetailedTelemetry(t *testing.T) {
	tempDir := t.TempDir()

	// Cria fila do Builder com uma tarefa
	sq := &queue.StationQueue{
		Station: queue.StationBuilder,
		Version: 1,
		Items: []queue.StationItem{
			{
				ID:           "task-builder-1",
				Title:        "Implementar Testes",
				Station:      queue.StationBuilder,
				Status:       queue.StatusQueued,
				ArtifactPath: "sdd/features/feat-01/task-01.md",
			},
		},
	}
	require.NoError(t, queue.SaveStationQueue(tempDir, sq))

	srv, err := NewServer(Config{Dir: tempDir})
	require.NoError(t, err)
	defer srv.Stop()

	// Executa a tarefa do Builder
	body := strings.NewReader(`{"station":"builder","task_id":"task-builder-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/task/execute", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// Verifica se a atividade registrada contém Command, FilePath, Output e CompletedAt
	activities, err := queue.LoadActivities(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, activities)

	latest := activities[0]
	assert.Equal(t, queue.StationBuilder, latest.Station)
	assert.Equal(t, "success", latest.Status)
	assert.Contains(t, latest.Command, "go test")
	assert.Equal(t, "sdd/features/feat-01/task-01.md", latest.FilePath)
	assert.NotEmpty(t, latest.Output)
	assert.NotEmpty(t, latest.CompletedAt)
}
