package studio_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStudioACPRPC valida o despacho de mensagens JSON-RPC 2.0 via HTTP POST em /api/acp/rpc (CA-01).
func TestStudioACPRPC(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	t.Run("Initialize Handshake", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","clientInfo":{"name":"zed","version":"0.140"}}}`
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))

		var rpcResp acp.Response
		err = json.NewDecoder(resp.Body).Decode(&rpcResp)
		require.NoError(t, err)

		assert.Equal(t, "2.0", rpcResp.JSONRPC)
		assert.Equal(t, float64(1), rpcResp.ID)
		assert.Nil(t, rpcResp.Error)

		var initRes acp.InitializeResult
		resultBytes, err := json.Marshal(rpcResp.Result)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(resultBytes, &initRes))
		assert.Equal(t, acp.DefaultProtocolVersion, initRes.ProtocolVersion)
		assert.Equal(t, acp.ServerName, initRes.ServerInfo.Name)
	})

	t.Run("Session Lifecycle", func(t *testing.T) {
		// 1. session/new
		newReq := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"session/new","params":{"workspaceDir":%q}}`, tempDir)
		respNew, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(newReq))
		require.NoError(t, err)
		defer respNew.Body.Close()

		require.Equal(t, http.StatusOK, respNew.StatusCode)
		var newResp acp.Response
		require.NoError(t, json.NewDecoder(respNew.Body).Decode(&newResp))
		require.Nil(t, newResp.Error)

		var sessRes acp.SessionResult
		resBytes, err := json.Marshal(newResp.Result)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(resBytes, &sessRes))
		sessID := sessRes.SessionID
		require.NotEmpty(t, sessID)

		// 2. session/load
		loadReq := fmt.Sprintf(`{"jsonrpc":"2.0","id":3,"method":"session/load","params":{"sessionId":%q}}`, sessID)
		respLoad, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(loadReq))
		require.NoError(t, err)
		defer respLoad.Body.Close()

		require.Equal(t, http.StatusOK, respLoad.StatusCode)
		var loadResp acp.Response
		require.NoError(t, json.NewDecoder(respLoad.Body).Decode(&loadResp))
		require.Nil(t, loadResp.Error)

		var loadRes acp.SessionResult
		loadBytes, err := json.Marshal(loadResp.Result)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(loadBytes, &loadRes))
		assert.Equal(t, "loaded", loadRes.Status)

		// 3. session/cancel
		cancelReq := fmt.Sprintf(`{"jsonrpc":"2.0","id":4,"method":"session/cancel","params":{"sessionId":%q,"reason":"teste"}}`, sessID)
		respCancel, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(cancelReq))
		require.NoError(t, err)
		defer respCancel.Body.Close()

		require.Equal(t, http.StatusOK, respCancel.StatusCode)
		var cancelResp acp.Response
		require.NoError(t, json.NewDecoder(respCancel.Body).Decode(&cancelResp))
		require.Nil(t, cancelResp.Error)

		var cancelRes acp.SessionResult
		cancelBytes, err := json.Marshal(cancelResp.Result)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(cancelBytes, &cancelRes))
		assert.Equal(t, "cancelled", cancelRes.Status)
	})

	t.Run("Notification Returns 204 No Content", func(t *testing.T) {
		notifReq := `{"jsonrpc":"2.0","method":"session/notification","params":{"type":"ping"}}`
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(notifReq))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Parse Error", func(t *testing.T) {
		malformedJSON := `{jsonrpc: "2.0", broken`
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(malformedJSON))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var rpcResp acp.Response
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&rpcResp))
		require.NotNil(t, rpcResp.Error)
		assert.Equal(t, acp.CodeParseError, rpcResp.Error.Code)
	})

	t.Run("Method Not Found", func(t *testing.T) {
		unknownReq := `{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(unknownReq))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var rpcResp acp.Response
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&rpcResp))
		require.NotNil(t, rpcResp.Error)
		assert.Equal(t, acp.CodeMethodNotFound, rpcResp.Error.Code)
	})

	t.Run("Invalid Request Version", func(t *testing.T) {
		invalidReq := `{"jsonrpc":"1.0","id":6,"method":"initialize"}`
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(invalidReq))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var rpcResp acp.Response
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&rpcResp))
		require.NotNil(t, rpcResp.Error)
		assert.Equal(t, acp.CodeInvalidRequest, rpcResp.Error.Code)
	})

	t.Run("Payload Exceeding 2MB Limit", func(t *testing.T) {
		// Gera payload ligeiramente superior a 2 MB (SEG-01, SEG-05)
		oversized := make([]byte, 2*1024*1024+10)
		for i := range oversized {
			oversized[i] = 'a'
		}
		resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", bytes.NewReader(oversized))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("Method Not Allowed On GET", func(t *testing.T) {
		resp, err := http.Get(srv.URL() + "/api/acp/rpc")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// TestStudioACPEventsSSE valida a recepção de eventos SSE dedicados em /api/acp/events (CA-02).
func TestStudioACPEventsSSE(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

	t.Run("Method Not Allowed On POST", func(t *testing.T) {
		resp, err := http.Post(srv.URL()+"/api/acp/events", "application/json", strings.NewReader("{}"))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Connect And Receive Event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL()+"/api/acp/events", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
		assert.Equal(t, "keep-alive", resp.Header.Get("Connection"))

		eventReceived := make(chan string, 1)
		go func() {
			reader := bufio.NewReader(resp.Body)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if strings.HasPrefix(line, "data:") {
					eventReceived <- strings.TrimSpace(strings.TrimPrefix(line, "data:"))
					return
				}
			}
		}()

		// Dispara notificação via despachador ACP
		srv.ACPDispatcher().BroadcastSessionNotification(acp.SessionNotification{
			SessionID: "sess-sse-test",
			Station:   "planner",
			Type:      "prompt_enqueued",
			Message:   "Iniciando planejamento da feature",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})

		select {
		case data := <-eventReceived:
			assert.Contains(t, data, "sess-sse-test")
			assert.Contains(t, data, "prompt_enqueued")
			assert.Contains(t, data, "planner")
		case <-time.After(2 * time.Second):
			t.Fatal("timeout aguardando evento SSE em /api/acp/events")
		}
	})
}

// TestStudioACPTransport valida concorrência multicliente e ausência de regressão em rotas legadas (CA-03).
func TestStudioACPTransport(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	t.Run("Multiclient Concurrency", func(t *testing.T) {
		const numSSEListeners = 5
		const numRPCCallers = 15

		var wg sync.WaitGroup
		sseCtx, sseCancel := context.WithCancel(context.Background())
		defer sseCancel()

		receivedCounts := make([]int, numSSEListeners)
		var mu sync.Mutex

		// 1. Inicia múltiplos ouvintes SSE concorrentes
		for i := 0; i < numSSEListeners; i++ {
			idx := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				req, err := http.NewRequestWithContext(sseCtx, http.MethodGet, srv.URL()+"/api/acp/events", nil)
				if err != nil {
					return
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				reader := bufio.NewReader(resp.Body)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if strings.HasPrefix(line, "data:") {
						mu.Lock()
						receivedCounts[idx]++
						mu.Unlock()
					}
				}
			}()
		}

		// Aguarda estabelecimento inicial das conexões SSE
		time.Sleep(100 * time.Millisecond)

		// 2. Dispara múltiplas chamadas JSON-RPC concorrentes
		var rpcWg sync.WaitGroup
		for i := 0; i < numRPCCallers; i++ {
			callIdx := i
			rpcWg.Add(1)
			go func() {
				defer rpcWg.Done()
				reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"initialize","params":{"protocolVersion":"1.0.0"}}`, callIdx)
				resp, err := http.Post(srv.URL()+"/api/acp/rpc", "application/json", strings.NewReader(reqBody))
				if err != nil {
					t.Errorf("erro no cliente rpc %d: %v", callIdx, err)
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				if resp.StatusCode != http.StatusOK {
					t.Errorf("status inesperado no cliente rpc %d: %d, body: %s", callIdx, resp.StatusCode, string(body))
				}
			}()
		}
		rpcWg.Wait()

		// 3. Emite broadcast e verifica entrega a todos os ouvintes SSE
		srv.ACPDispatcher().BroadcastSessionNotification(acp.SessionNotification{
			SessionID: "concurrent-sess",
			Station:   "builder",
			Type:      "gate_approved",
			Message:   "Concorrência confirmada",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})

		time.Sleep(200 * time.Millisecond)
		sseCancel()
		wg.Wait()

		mu.Lock()
		for i, count := range receivedCounts {
			assert.GreaterOrEqual(t, count, 1, "ouvinte SSE %d deveria receber pelo menos 1 notificação", i)
		}
		mu.Unlock()
	})

	t.Run("No Regression On Legacy Routes", func(t *testing.T) {
		// 1. GET /api/state retorna 200 OK e JSON válido
		respState, err := http.Get(srv.URL() + "/api/state")
		require.NoError(t, err)
		defer respState.Body.Close()
		assert.Equal(t, http.StatusOK, respState.StatusCode)

		var stateMap map[string]any
		require.NoError(t, json.NewDecoder(respState.Body).Decode(&stateMap))
		assert.Contains(t, stateMap, "stations")

		// 2. POST /api/state é estritamente proibido (Read-Only)
		respPostState, err := http.Post(srv.URL()+"/api/state", "application/json", strings.NewReader("{}"))
		require.NoError(t, err)
		defer respPostState.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, respPostState.StatusCode)

		// 3. SSE legado em /api/events continua funcional
		reqEvents, err := http.NewRequest(http.MethodGet, srv.URL()+"/api/events", nil)
		require.NoError(t, err)
		respEvents, err := http.DefaultClient.Do(reqEvents)
		require.NoError(t, err)
		defer respEvents.Body.Close()
		assert.Equal(t, http.StatusOK, respEvents.StatusCode)
		assert.Contains(t, respEvents.Header.Get("Content-Type"), "text/event-stream")

		// Lê primeiro evento de estado emitido imediatamente
		reader := bufio.NewReader(respEvents.Body)
		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "data:"))
	})

	_ = tempDir
}
