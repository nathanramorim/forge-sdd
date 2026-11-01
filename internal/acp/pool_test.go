package acp_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionPool_Warmup valida a inicialização e pré-aquecimento das 5 sessões canônicas (CA-01).
func TestSessionPool_Warmup(t *testing.T) {
	wsDir := t.TempDir()
	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	assert.False(t, pool.IsWarmed())
	assert.Empty(t, pool.WorkspaceDir())

	// Inscreve canal para capturar a notificação "pool/warmed"
	notifCh := make(chan acp.Notification, 5)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	// Executa Warmup
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := pool.Warmup(ctx, wsDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "warmed", result.Status)
	assert.Equal(t, 5, result.SessionsCount)
	assert.Len(t, result.Sessions, 5)
	assert.True(t, pool.IsWarmed())
	assert.Equal(t, wsDir, pool.WorkspaceDir())

	expectedSessions := map[string]string{
		"sess-01-lider":      queue.StationLider,
		"sess-02-pm":         queue.StationPMArchitect,
		"sess-03-planner":    queue.StationPlanner,
		"sess-04-builder":    queue.StationBuilder,
		"sess-05-governance": queue.StationGovernance,
	}

	for id, expectedStation := range expectedSessions {
		sess, ok := pool.GetSession(id)
		require.True(t, ok, "sessão %s deve existir no pool", id)
		assert.Equal(t, expectedStation, sess.Station)
		assert.Equal(t, wsDir, sess.WorkspaceDir)
		assert.Equal(t, acp.SessionStateStandby, sess.State)
		assert.Empty(t, sess.ActiveItemID)
		assert.False(t, sess.CreatedAt.IsZero())
		assert.False(t, sess.LastSeenAt.IsZero())

		byStation, okStation := pool.GetSessionByStation(expectedStation)
		require.True(t, okStation, "sessão da estação %s deve existir no pool", expectedStation)
		assert.Equal(t, id, byStation.SessionID)

		// Verifica sincronização no SessionManager
		smSess, smOk := dispatcher.SessionManager().GetSession(id)
		require.True(t, smOk, "sessão %s deve ser refletida no SessionManager", id)
		assert.Equal(t, expectedStation, smSess.Station)
		assert.Equal(t, "active", smSess.Status)
	}

	// Verifica recebimento da notificação pool/warmed
	select {
	case notif := <-notifCh:
		assert.Equal(t, "pool/warmed", notif.Method)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout aguardando notificação pool/warmed")
	}

	// Teste de concorrência com -race
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sessions := pool.ListSessions()
			assert.Len(t, sessions, 5)

			s, ok := pool.GetSession("sess-04-builder")
			assert.True(t, ok)
			assert.Equal(t, queue.StationBuilder, s.Station)

			st, okSt := pool.GetSessionByStation(queue.StationPlanner)
			assert.True(t, okSt)
			assert.Equal(t, "sess-03-planner", st.SessionID)
		}(i)
	}
	wg.Wait()
}

// TestSessionPool_States valida a máquina de estados, transições e isolamento de cópias defensivas.
func TestSessionPool_States(t *testing.T) {
	wsDir := t.TempDir()
	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	t.Run("Valid state transitions on UpdateSessionState", func(t *testing.T) {
		updated, err := pool.UpdateSessionState("sess-04-builder", acp.SessionStateProcessing, "task-101")
		require.NoError(t, err)
		assert.Equal(t, acp.SessionStateProcessing, updated.State)
		assert.Equal(t, "task-101", updated.ActiveItemID)

		// Transição para approval_pending
		updated, err = pool.UpdateSessionState("sess-04-builder", acp.SessionStateApprovalPending, "task-101")
		require.NoError(t, err)
		assert.Equal(t, acp.SessionStateApprovalPending, updated.State)

		// Transição para error
		updated, err = pool.UpdateSessionState("sess-04-builder", acp.SessionStateError, "")
		require.NoError(t, err)
		assert.Equal(t, acp.SessionStateError, updated.State)

		// Transição para offline
		updated, err = pool.UpdateSessionState("sess-04-builder", acp.SessionStateOffline, "")
		require.NoError(t, err)
		assert.Equal(t, acp.SessionStateOffline, updated.State)

		// Retorno para standby
		updated, err = pool.UpdateSessionState("sess-04-builder", acp.SessionStateStandby, "")
		require.NoError(t, err)
		assert.Equal(t, acp.SessionStateStandby, updated.State)
	})

	t.Run("Valid state transitions on TransitionState by Station", func(t *testing.T) {
		err := pool.TransitionState(queue.StationPMArchitect, acp.SessionStateApprovalPending, "aguardando revisao gate")
		require.NoError(t, err)

		sess, ok := pool.GetSessionByStation(queue.StationPMArchitect)
		require.True(t, ok)
		assert.Equal(t, acp.SessionStateApprovalPending, sess.State)
		assert.Equal(t, "aguardando revisao gate", sess.Metadata["transitionReason"])
	})

	t.Run("Rejection of invalid states and invalid stations", func(t *testing.T) {
		_, err := pool.UpdateSessionState("sess-04-builder", acp.SessionState("invalid_state"), "")
		assert.Error(t, err)

		_, err = pool.UpdateSessionState("sess-inexistente", acp.SessionStateStandby, "")
		assert.Error(t, err)

		err = pool.TransitionState("estacao_invalida", acp.SessionStateStandby, "")
		assert.Error(t, err)

		err = pool.TransitionState(queue.StationLider, acp.SessionState("estado_errado"), "")
		assert.Error(t, err)
	})

	t.Run("Defensive copy protection", func(t *testing.T) {
		sess, ok := pool.GetSession("sess-01-lider")
		require.True(t, ok)

		// Mutação externa não pode afetar o pool
		sess.State = acp.SessionStateError
		sess.ActiveItemID = "hacked"
		if sess.Metadata == nil {
			sess.Metadata = make(map[string]string)
		}
		sess.Metadata["injected"] = "val"

		fresh, ok := pool.GetSession("sess-01-lider")
		require.True(t, ok)
		assert.Equal(t, acp.SessionStateStandby, fresh.State)
		assert.Empty(t, fresh.ActiveItemID)
		assert.Empty(t, fresh.Metadata["injected"])
	})

	t.Run("Concurrent state updates", func(t *testing.T) {
		var wg sync.WaitGroup
		stations := []string{
			queue.StationLider,
			queue.StationPMArchitect,
			queue.StationPlanner,
			queue.StationBuilder,
			queue.StationGovernance,
		}

		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				station := stations[idx%len(stations)]
				states := acp.ValidSessionStates()
				targetState := states[idx%len(states)]
				_ = pool.TransitionState(station, targetState, "concurrent test")
			}(i)
		}
		wg.Wait()
	})
}

// TestSessionPool_Heartbeat valida ciclos periódicos do monitor, atualização de lastSeenAt e detecção de stalls.
func TestSessionPool_Heartbeat(t *testing.T) {
	wsDir := t.TempDir()
	dispatcher := acp.NewDispatcher()
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	_, err := pool.Warmup(context.Background(), wsDir)
	require.NoError(t, err)

	notifCh := make(chan acp.Notification, 10)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	// HeartbeatMonitor com intervalo ultra rápido para testes unitários
	monitor := acp.NewHeartbeatMonitorConfig(pool, dispatcher, 20*time.Millisecond, 50*time.Millisecond)
	pool.SetHeartbeatMonitor(monitor)

	t.Run("Tick updates lastSeenAt on all sessions", func(t *testing.T) {
		beforeTick := time.Now().UTC()
		time.Sleep(5 * time.Millisecond)

		monitor.Tick()

		for _, s := range pool.ListSessions() {
			assert.True(t, s.LastSeenAt.After(beforeTick) || s.LastSeenAt.Equal(beforeTick))
		}
	})

	t.Run("Detect stalled session in processing and emit session/stalled", func(t *testing.T) {
		// Coloca a sessão builder em processamento
		_, err := pool.UpdateSessionState("sess-04-builder", acp.SessionStateProcessing, "item-slow-1")
		require.NoError(t, err)

		// Aguarda tempo superior ao stallTimeout configurado (50ms)
		time.Sleep(70 * time.Millisecond)

		// Executa Tick
		monitor.Tick()

		// Deve emitir notificação "session/stalled"
		select {
		case notif := <-notifCh:
			// Pode ter recebido pool_warmed ou station_state antes; procura session/stalled
			if notif.Method != "session/stalled" {
				for {
					select {
					case nextNotif := <-notifCh:
						if nextNotif.Method == "session/stalled" {
							notif = nextNotif
							goto Found
						}
					case <-time.After(500 * time.Millisecond):
						t.Fatalf("esperava notificação session/stalled mas recebeu %s", notif.Method)
					}
				}
			}
		Found:
			assert.Equal(t, "session/stalled", notif.Method)
			rawBytes, _ := json.Marshal(notif.Params)
			var stalledSess acp.PooledSession
			err := json.Unmarshal(rawBytes, &stalledSess)
			require.NoError(t, err)
			assert.Equal(t, "sess-04-builder", stalledSess.SessionID)
			assert.Equal(t, queue.StationBuilder, stalledSess.Station)
			assert.Equal(t, acp.SessionStateProcessing, stalledSess.State)
			assert.Equal(t, "item-slow-1", stalledSess.ActiveItemID)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout aguardando notificação session/stalled")
		}
	})

	t.Run("TouchSession resets stall tracking", func(t *testing.T) {
		err := pool.TouchSession("sess-04-builder")
		require.NoError(t, err)
	})

	t.Run("Start and Stop lifecycle zero goroutine leak", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		assert.False(t, monitor.IsRunning())
		monitor.Start(ctx)
		assert.True(t, monitor.IsRunning())

		// Segundo start idempotente
		monitor.Start(ctx)
		assert.True(t, monitor.IsRunning())

		time.Sleep(50 * time.Millisecond)

		monitor.Stop()
		assert.False(t, monitor.IsRunning())

		// Segundo stop idempotente
		monitor.Stop()
		assert.False(t, monitor.IsRunning())
	})
}

// TestSessionPool_SecurityWorkspace valida confinamento de diretório e bloqueio de path traversal (SEG-01, SEG-02).
func TestSessionPool_SecurityWorkspace(t *testing.T) {
	tempRoot := t.TempDir()
	dispatcher := acp.NewDispatcher()
	dispatcher.SessionManager().SetAllowedRoot(tempRoot)
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	t.Run("Reject path traversal traversal outside allowed root", func(t *testing.T) {
		_, err := pool.Warmup(context.Background(), "../etc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
		assert.False(t, pool.IsWarmed())
		assert.Empty(t, pool.WorkspaceDir())
	})

	t.Run("Reject non-existent directory", func(t *testing.T) {
		nonExistent := filepath.Join(tempRoot, "subpasta_inexistente_12345")
		_, err := pool.Warmup(context.Background(), nonExistent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "não existe")
		assert.False(t, pool.IsWarmed())
	})

	t.Run("Accept directory inside allowed root", func(t *testing.T) {
		result, err := pool.Warmup(context.Background(), tempRoot)
		require.NoError(t, err)
		assert.True(t, pool.IsWarmed())
		assert.Equal(t, tempRoot, result.Sessions[0].WorkspaceDir)
	})
}

// TestSessionPool_RPCHandlers valida os métodos JSON-RPC 2.0 pool/warmup e pool/status via RPCDispatcher.
func TestSessionPool_RPCHandlers(t *testing.T) {
	tempDir := t.TempDir()
	dispatcher := acp.NewDispatcher()
	dispatcher.SessionManager().SetAllowedRoot(tempDir)

	ctx := context.Background()

	t.Run("pool/status before warmup returns unhealthy", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "pool/status",
		}
		raw, _ := json.Marshal(req)
		resp := dispatcher.Dispatch(ctx, raw)
		require.NotNil(t, resp)
		assert.Nil(t, resp.Error)

		statusRes, ok := resp.Result.(*acp.PoolStatusResult)
		require.True(t, ok)
		assert.False(t, statusRes.Healthy)
		assert.Empty(t, statusRes.Sessions)
	})

	t.Run("pool/warmup with invalid traversal returns CodeInvalidParams (-32602)", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "pool/warmup",
			"params": acp.PoolWarmupParams{
				WorkspaceDir: "../etc",
			},
		}
		raw, _ := json.Marshal(req)
		resp := dispatcher.Dispatch(ctx, raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "path traversal")
	})

	t.Run("pool/warmup valid executes successfully", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      3,
			"method":  "pool/warmup",
			"params": acp.PoolWarmupParams{
				WorkspaceDir: tempDir,
			},
		}
		raw, _ := json.Marshal(req)
		resp := dispatcher.Dispatch(ctx, raw)
		require.NotNil(t, resp)
		assert.Nil(t, resp.Error)

		warmupRes, ok := resp.Result.(*acp.PoolWarmupResult)
		require.True(t, ok)
		assert.Equal(t, "warmed", warmupRes.Status)
		assert.Equal(t, 5, warmupRes.SessionsCount)
		assert.Len(t, warmupRes.Sessions, 5)
	})

	t.Run("pool/status after warmup returns healthy and 5 sessions", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      4,
			"method":  "pool/status",
		}
		raw, _ := json.Marshal(req)
		resp := dispatcher.Dispatch(ctx, raw)
		require.NotNil(t, resp)
		assert.Nil(t, resp.Error)

		statusRes, ok := resp.Result.(*acp.PoolStatusResult)
		require.True(t, ok)
		assert.True(t, statusRes.Healthy)
		assert.Len(t, statusRes.Sessions, 5)

		for _, s := range statusRes.Sessions {
			assert.NotEmpty(t, s.SessionID)
			assert.NotEmpty(t, s.Station)
			assert.Equal(t, acp.SessionStateStandby, s.State)
			assert.False(t, s.LastSeenAt.IsZero())
			assert.Equal(t, 0, s.AttachedClients)
		}
	})

	t.Run("pool/status with an error session marks healthy as false", func(t *testing.T) {
		_, err := dispatcher.Pool().UpdateSessionState("sess-02-pm", acp.SessionStateError, "")
		require.NoError(t, err)

		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      5,
			"method":  "pool/status",
		}
		raw, _ := json.Marshal(req)
		resp := dispatcher.Dispatch(ctx, raw)
		require.NotNil(t, resp)
		assert.Nil(t, resp.Error)

		statusRes, ok := resp.Result.(*acp.PoolStatusResult)
		require.True(t, ok)
		assert.False(t, statusRes.Healthy)
	})
}
