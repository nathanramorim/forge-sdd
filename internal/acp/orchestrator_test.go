package acp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "sdd", ".queue", "stations"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "sdd", "features", "feat-65"), 0o755))
	return tempDir
}

func createValidSpecFile(t *testing.T, dir, relPath string) string {
	t.Helper()
	fullPath := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))

	content := `# feat-65-02 — Orquestrador de Handoff

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Fila de Estações | internal/queue | StationsDir | reusar | Reuso de infraestrutura |

## Dependências novas
- Nenhuma dependência externa.
`
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	return relPath
}

// TestHandoffOrchestrator_PMToPlanner_GateBlocking valida que a conclusão de uma spec pelo PM (Estação 02)
// transiciona sess-02-pm para approval_pending, emite a notificação gate/approval_required
// e bloqueia qualquer despacho para sess-03-planner (FLX-01).
func TestHandoffOrchestrator_PMToPlanner_GateBlocking(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	specRel := "sdd/features/feat-65/feat-65-02.md"
	createValidSpecFile(t, wsDir, specRel)

	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	orch := dispatcher.Orchestrator()
	require.NotNil(t, orch)
	orch.SetDir(wsDir)

	notifCh := make(chan acp.Notification, 20)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	// Estado inicial: sess-02-pm e sess-03-planner em Standby
	pmSess, ok := pool.GetSessionByStation(queue.StationPMArchitect)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateStandby, pmSess.State)

	planSess, ok := pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateStandby, planSess.State)

	// Estação 02 conclui elaboração da spec
	taskID := "feat-65-02"
	err = orch.HandleTaskCompleted(context.Background(), queue.StationPMArchitect, taskID, specRel)
	require.NoError(t, err)

	// sess-02-pm deve transicionar para approval_pending com task vinculada
	pmSess, ok = pool.GetSessionByStation(queue.StationPMArchitect)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateApprovalPending, pmSess.State)
	assert.Equal(t, taskID, pmSess.ActiveItemID)

	// sess-03-planner NÃO deve ser ativada (permanece em standby)
	planSess, ok = pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateStandby, planSess.State)
	assert.Empty(t, planSess.ActiveItemID)

	// Valida emissão da notificação gate/approval_required
	var receivedGateNotif bool
	deadline := time.After(1 * time.Second)
	for !receivedGateNotif {
		select {
		case n := <-notifCh:
			if n.Method == "gate/approval_required" {
				receivedGateNotif = true
				raw, err := json.Marshal(n.Params)
				require.NoError(t, err)
				var payload map[string]any
				require.NoError(t, json.Unmarshal(raw, &payload))
				assert.Equal(t, queue.StationPMArchitect, payload["station"])
				assert.Equal(t, taskID, payload["itemId"])
				assert.Equal(t, specRel, payload["specPath"])
				assert.Equal(t, "FLX-01", payload["gate"])
			}
		case <-deadline:
			t.Fatal("timeout aguardando notificação gate/approval_required")
		}
	}

	// Tentativa de despacho para sess-03-planner deve ser bloqueada
	err = orch.DispatchTask(context.Background(), queue.StationPlanner, taskID, specRel)
	require.Error(t, err)
	var rpcErr *acp.RPCError
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, acp.CodeServerError, rpcErr.Code)

	// Estado do planner permanece inalterado em standby
	planSess, _ = pool.GetSessionByStation(queue.StationPlanner)
	assert.Equal(t, acp.SessionStateStandby, planSess.State)
}

// TestHandoffOrchestrator_GateApproval_TransitionsToPlanner valida a aprovação determinística do
// Portão FLX-01, transição de sess-02-pm para standby e ativação contínua de sess-03-planner em processing.
func TestHandoffOrchestrator_GateApproval_TransitionsToPlanner(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	specRel := "sdd/features/feat-65/feat-65-02.md"
	createValidSpecFile(t, wsDir, specRel)

	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	orch := dispatcher.Orchestrator()
	orch.SetDir(wsDir)

	notifCh := make(chan acp.Notification, 20)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	taskID := "feat-65-02"
	err = orch.HandleTaskCompleted(context.Background(), queue.StationPMArchitect, taskID, specRel)
	require.NoError(t, err)

	// Homologa o Portão FLX-01 via ApproveGate
	res, err := orch.ApproveGate(context.Background(), queue.StationPMArchitect, taskID, specRel)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, taskID, res.FeatureID)
	assert.Equal(t, "FLX-01", res.Gate)
	assert.Equal(t, queue.StatusApproved, res.Status)
	assert.True(t, res.Approved)
	assert.Equal(t, queue.StationPMArchitect, res.FromStation)
	assert.Equal(t, queue.StationPlanner, res.ToStation)
	assert.Equal(t, queue.StationPlanner, res.NextStation)
	require.NotNil(t, res.HandoffItem)
	assert.Equal(t, queue.StatusQueued, res.HandoffItem.Status)

	// sess-02-pm deve ter retornado para standby
	pmSess, ok := pool.GetSessionByStation(queue.StationPMArchitect)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateStandby, pmSess.State)

	// sess-03-planner deve ter sido ativada automaticamente em processing
	planSess, ok := pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateProcessing, planSess.State)
	assert.Equal(t, taskID, planSess.ActiveItemID)

	// Filas em disco devem refletir o handoff determinístico
	pmSQ, err := queue.LoadStationQueue(wsDir, queue.StationPMArchitect)
	require.NoError(t, err)
	require.NotEmpty(t, pmSQ.Items)
	assert.Equal(t, queue.StatusApproved, pmSQ.Items[0].Status)

	planSQ, err := queue.LoadStationQueue(wsDir, queue.StationPlanner)
	require.NoError(t, err)
	require.NotEmpty(t, planSQ.Items)
	assert.Equal(t, queue.StatusQueued, planSQ.Items[0].Status)

	// Verifica se a task agora é reconhecida como aprovada
	assert.True(t, orch.IsGateApproved(taskID))
	assert.False(t, orch.IsGatePending(taskID))

	// Notificação session/notification do tipo station_state deve ter sido emitida
	var receivedStationState bool
	deadline := time.After(1 * time.Second)
	for !receivedStationState {
		select {
		case n := <-notifCh:
			if n.Method == "session/notification" {
				raw, err := json.Marshal(n.Params)
				require.NoError(t, err)
				var sn acp.SessionNotification
				require.NoError(t, json.Unmarshal(raw, &sn))
				if sn.Type == "station_state" && sn.Station == queue.StationPlanner {
					receivedStationState = true
				}
			}
		case <-deadline:
			t.Fatal("timeout aguardando notificação station_state para planner")
		}
	}
}

// TestHandoffOrchestrator_BypassPrevention garante que qualquer tentativa de ativar o planner
// ou avançar do PM para o planner sem aprovação do portão seja rejeitada com erro canônico -32001.
func TestHandoffOrchestrator_BypassPrevention(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	orch := dispatcher.Orchestrator()
	orch.SetDir(wsDir)

	taskID := "feat-bypass-test"
	err = orch.HandleTaskCompleted(context.Background(), queue.StationPMArchitect, taskID, "sdd/features/spec.md")
	require.NoError(t, err)

	// 1. Tentativa de DispatchTask direta para planner
	err = orch.DispatchTask(context.Background(), queue.StationPlanner, taskID, "")
	require.Error(t, err)
	var rpcErr *acp.RPCError
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, acp.CodeServerError, rpcErr.Code)
	assert.Contains(t, rpcErr.Message, "bypass")

	// 2. Tentativa via ActivateStation
	err = orch.ActivateStation(context.Background(), queue.StationPlanner, taskID)
	require.Error(t, err)
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, acp.CodeServerError, rpcErr.Code)

	// 3. Tentativa via AdvanceStation a partir de pm-architect
	err = orch.AdvanceStation(context.Background(), queue.StationPMArchitect, taskID)
	require.Error(t, err)
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, acp.CodeServerError, rpcErr.Code)

	// 4. Tentativa de aprovação com arquivo inexistente (P-02 violação)
	_, err = orch.ApproveGate(context.Background(), queue.StationPMArchitect, taskID, "arquivo-inexistente.md")
	require.Error(t, err)
	require.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, acp.CodeServerError, rpcErr.Code)
	assert.Contains(t, rpcErr.Message, "Portão FLX-01 rejeitado")

	// 5. Verifica que os estados das sessões permaneceram intocados
	pmSess, ok := pool.GetSessionByStation(queue.StationPMArchitect)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateApprovalPending, pmSess.State)

	planSess, ok := pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateStandby, planSess.State)
}

// TestHandoffOrchestrator_ConcurrentHandoffs valida a thread-safety do orquestrador sob execução
// concorrente pesada com detecção de corrida (go test -race).
func TestHandoffOrchestrator_ConcurrentHandoffs(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	orch := dispatcher.Orchestrator()
	orch.SetDir(wsDir)

	const concurrency = 16
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		idx := i
		go func() {
			defer wg.Done()
			taskID := fmt.Sprintf("feat-concurrent-%02d", idx)
			specRel := fmt.Sprintf("sdd/features/feat-65/feat-%02d.md", idx)
			createValidSpecFile(t, wsDir, specRel)

			// 1. Sinaliza conclusão
			cErr := orch.HandleTaskCompleted(context.Background(), queue.StationPMArchitect, taskID, specRel)
			assert.NoError(t, cErr)

			// 2. Consulta estado concorrente
			_ = orch.IsGatePending(taskID)
			_ = orch.PendingGates()

			// 3. Tentativa de bypass concorrente (deve ser rejeitada ou passar somente após aprovação)
			_ = orch.DispatchTask(context.Background(), queue.StationPlanner, taskID, specRel)

			// 4. Aprovação formal do portão
			res, aErr := orch.ApproveGate(context.Background(), queue.StationPMArchitect, taskID, specRel)
			assert.NoError(t, aErr)
			assert.NotNil(t, res)

			// 5. Pós-aprovação, validação idempotente
			assert.True(t, orch.IsGateApproved(taskID))
		}()
	}

	wg.Wait()
}
