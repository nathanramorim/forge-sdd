package acp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestACPInitialize valida o handshake "initialize", parsing JSON-RPC 2.0 e transporte Stdio (CA-01).
func TestACPInitialize(t *testing.T) {
	t.Run("Direct Dispatcher Handshake", func(t *testing.T) {
		dispatcher := acp.NewDispatcherWithVersion("3.2.1-beta")
		assert.False(t, dispatcher.IsInitialized())

		reqPayload := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": acp.InitializeParams{
				ProtocolVersion: "1.0.0",
				ClientInfo: acp.ClientInfo{
					Name:    "zed",
					Version: "0.140.0",
				},
				Capabilities: &acp.ClientCapabilities{
					Terminal: true,
					Prompts:  true,
				},
			},
		}
		raw, err := json.Marshal(reqPayload)
		require.NoError(t, err)

		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Equal(t, float64(1), resp.ID)
		assert.Nil(t, resp.Error)
		assert.True(t, dispatcher.IsInitialized())

		initResult, ok := resp.Result.(acp.InitializeResult)
		require.True(t, ok)
		assert.Equal(t, acp.DefaultProtocolVersion, initResult.ProtocolVersion)
		assert.Equal(t, acp.ServerName, initResult.ServerInfo.Name)
		assert.Equal(t, "3.2.1-beta", initResult.ServerInfo.Version)
		assert.True(t, initResult.Capabilities.Sessions)
		assert.True(t, initResult.Capabilities.Gates)
		assert.True(t, initResult.Capabilities.Prompts)
	})

	t.Run("StdioTransport Handshake via Pipe", func(t *testing.T) {
		dispatcher := acp.NewDispatcherWithVersion("3.2.1-beta")
		inReader, inWriter := io.Pipe()
		var outBuf bytes.Buffer

		transport := acp.NewStdioTransport(dispatcher, inReader, &outBuf)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- transport.Run(ctx)
		}()

		// Envia handshake initialize
		initMsg := `{"jsonrpc":"2.0","id":"init-42","method":"initialize","params":{"protocolVersion":"1.0.0","clientInfo":{"name":"cursor","version":"0.40"}}}` + "\n"
		_, err := inWriter.Write([]byte(initMsg))
		require.NoError(t, err)

		// Fecha writer para simular encerramento do stream
		require.NoError(t, inWriter.Close())

		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout aguardando término do transporte stdio")
		}

		outData := outBuf.Bytes()
		lines := strings.Split(strings.TrimSpace(string(outData)), "\n")
		require.NotEmpty(t, lines)

		var resp acp.Response
		err = json.Unmarshal([]byte(lines[0]), &resp)
		require.NoError(t, err)

		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Equal(t, "init-42", resp.ID)
		assert.Nil(t, resp.Error)

		var result acp.InitializeResult
		resultBytes, err := json.Marshal(resp.Result)
		require.NoError(t, err)
		err = json.Unmarshal(resultBytes, &result)
		require.NoError(t, err)

		assert.Equal(t, acp.ServerName, result.ServerInfo.Name)
		assert.Equal(t, "3.2.1-beta", result.ServerInfo.Version)
	})
}

// TestACPErrors valida o tratamento de mensagens malformadas e códigos padronizados (CA-02).
func TestACPErrors(t *testing.T) {
	dispatcher := acp.NewDispatcher()

	t.Run("Parse Error (-32700)", func(t *testing.T) {
		malformedJSON := []byte(`{jsonrpc: 2.0, "method": broken`)
		resp := dispatcher.Dispatch(context.Background(), malformedJSON)

		require.NotNil(t, resp)
		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Nil(t, resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeParseError, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "Parse error")
	})

	t.Run("Invalid Request - Missing jsonrpc (-32600)", func(t *testing.T) {
		invalidReq := []byte(`{"id": 1, "method": "test"}`)
		resp := dispatcher.Dispatch(context.Background(), invalidReq)

		require.NotNil(t, resp)
		assert.Equal(t, float64(1), resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidRequest, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "jsonrpc")
	})

	t.Run("Invalid Request - Wrong jsonrpc version (-32600)", func(t *testing.T) {
		invalidReq := []byte(`{"jsonrpc": "1.0", "id": "req-1", "method": "test"}`)
		resp := dispatcher.Dispatch(context.Background(), invalidReq)

		require.NotNil(t, resp)
		assert.Equal(t, "req-1", resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidRequest, resp.Error.Code)
	})

	t.Run("Invalid Request - Empty Method (-32600)", func(t *testing.T) {
		invalidReq := []byte(`{"jsonrpc": "2.0", "id": 2, "method": ""}`)
		resp := dispatcher.Dispatch(context.Background(), invalidReq)

		require.NotNil(t, resp)
		assert.Equal(t, float64(2), resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidRequest, resp.Error.Code)
	})

	t.Run("Method Not Found (-32601)", func(t *testing.T) {
		unknownReq := []byte(`{"jsonrpc": "2.0", "id": 3, "method": "unknown/nonexistent"}`)
		resp := dispatcher.Dispatch(context.Background(), unknownReq)

		require.NotNil(t, resp)
		assert.Equal(t, float64(3), resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeMethodNotFound, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "unknown/nonexistent")
	})

	t.Run("Invalid Params (-32602)", func(t *testing.T) {
		invalidParamsReq := []byte(`{"jsonrpc": "2.0", "id": 4, "method": "session/load", "params": "not-an-object"}`)
		resp := dispatcher.Dispatch(context.Background(), invalidParamsReq)

		require.NotNil(t, resp)
		assert.Equal(t, float64(4), resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})

	t.Run("Custom Server Error (-32001)", func(t *testing.T) {
		dispatcher.RegisterHandler("gate/reject-test", func(ctx context.Context, params json.RawMessage) (any, *acp.RPCError) {
			return nil, acp.NewRPCError(acp.CodeServerError, "Portão FLX-01 rejeitado", map[string]string{
				"check": "P-02",
			})
		})

		req := []byte(`{"jsonrpc": "2.0", "id": 5, "method": "gate/reject-test"}`)
		resp := dispatcher.Dispatch(context.Background(), req)

		require.NotNil(t, resp)
		assert.Equal(t, float64(5), resp.ID)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeServerError, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "Portão FLX-01 rejeitado")
		assert.NotNil(t, resp.Error.Data)
	})

	t.Run("Internal Error (-32603)", func(t *testing.T) {
		dispatcher.RegisterHandler("panic/test", func(ctx context.Context, params json.RawMessage) (any, *acp.RPCError) {
			return nil, acp.NewRPCError(acp.CodeInternalError, "falha interna irrecuperável")
		})

		req := []byte(`{"jsonrpc": "2.0", "id": 6, "method": "panic/test"}`)
		resp := dispatcher.Dispatch(context.Background(), req)

		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInternalError, resp.Error.Code)
	})
}

// TestACPNotifications testa que notificações não produzem resposta e broadcast funciona.
func TestACPNotifications(t *testing.T) {
	dispatcher := acp.NewDispatcher()

	var notifReceived bool
	var notifMu sync.Mutex

	dispatcher.RegisterNotificationHandler("session/notify", func(ctx context.Context, params json.RawMessage) {
		notifMu.Lock()
		defer notifMu.Unlock()
		notifReceived = true
	})

	// Mensagem de notificação (sem campo "id")
	notifMsg := []byte(`{"jsonrpc": "2.0", "method": "session/notify", "params": {"message": "hello"}}`)
	resp := dispatcher.Dispatch(context.Background(), notifMsg)
	assert.Nil(t, resp, "notificações JSON-RPC 2.0 não devem gerar resposta")

	notifMu.Lock()
	assert.True(t, notifReceived, "handler de notificação deve ser acionado")
	notifMu.Unlock()

	// Teste de inscrição e broadcast de notificação
	subCh := make(chan acp.Notification, 5)
	dispatcher.SubscribeNotifications(subCh)
	defer dispatcher.UnsubscribeNotifications(subCh)

	expectedNotif := acp.NewNotification("station/changed", map[string]string{
		"station": "04-builder",
	})
	dispatcher.BroadcastNotification(expectedNotif)

	select {
	case received := <-subCh:
		assert.Equal(t, "station/changed", received.Method)
	case <-time.After(1 * time.Second):
		t.Fatal("notificação não recebida no subscriber")
	}
}

// TestACPSessionBaseMethods valida os métodos canônicos session/new, session/load, session/cancel.
func TestACPSessionBaseMethods(t *testing.T) {
	dispatcher := acp.NewDispatcher()

	// session/new
	newReq := []byte(`{"jsonrpc": "2.0", "id": 10, "method": "session/new", "params": {"workspaceDir": "/tmp/test", "feature": "feat-64"}}`)
	resp := dispatcher.Dispatch(context.Background(), newReq)
	require.NotNil(t, resp)
	require.Nil(t, resp.Error)

	res, ok := resp.Result.(acp.SessionResult)
	require.True(t, ok)
	assert.NotEmpty(t, res.SessionID)
	assert.Equal(t, "active", res.Status)
	assert.Equal(t, "/tmp/test", res.WorkspaceDir)

	// session/load
	loadReq := []byte(fmt.Sprintf(`{"jsonrpc": "2.0", "id": 11, "method": "session/load", "params": {"sessionId": %q}}`, res.SessionID))
	resp = dispatcher.Dispatch(context.Background(), loadReq)
	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	loadRes, ok := resp.Result.(acp.SessionResult)
	require.True(t, ok)
	assert.Equal(t, res.SessionID, loadRes.SessionID)
	assert.Equal(t, "loaded", loadRes.Status)

	// session/cancel
	cancelReq := []byte(fmt.Sprintf(`{"jsonrpc": "2.0", "id": 12, "method": "session/cancel", "params": {"sessionId": %q, "reason": "aborted"}}`, res.SessionID))
	resp = dispatcher.Dispatch(context.Background(), cancelReq)
	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	cancelRes, ok := resp.Result.(acp.SessionResult)
	require.True(t, ok)
	assert.Equal(t, res.SessionID, cancelRes.SessionID)
	assert.Equal(t, "cancelled", cancelRes.Status)
}

// TestACPConcurrency valida concorrência segura em alto volume com detecção de data race.
func TestACPConcurrency(t *testing.T) {
	dispatcher := acp.NewDispatcher()

	// Registra handler customizado
	dispatcher.RegisterHandler("compute", func(ctx context.Context, params json.RawMessage) (any, *acp.RPCError) {
		var req struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, acp.NewRPCError(acp.CodeInvalidParams, "invalid params")
		}
		return map[string]int{"square": req.Index * req.Index}, nil
	})

	inReader, inWriter := io.Pipe()
	var outBuf bytes.Buffer
	transport := acp.NewStdioTransport(dispatcher, inReader, &outBuf)
	transport.SetConcurrent(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		runErr <- transport.Run(ctx)
	}()

	const totalRequests = 100
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	var writeMu sync.Mutex
	for i := 0; i < totalRequests; i++ {
		go func(idx int) {
			defer wg.Done()
			msg := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"compute","params":{"index":%d}}`+"\n", idx, idx)
			writeMu.Lock()
			_, _ = inWriter.Write([]byte(msg))
			writeMu.Unlock()
		}(i)
	}

	wg.Wait()
	_ = inWriter.Close()

	select {
	case err := <-runErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout aguardando processamento concorrente do stdio transport")
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	assert.Equal(t, totalRequests, len(lines), "todas as requisições concorrentes devem ter gerado resposta")
}
