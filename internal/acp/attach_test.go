package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachBridge_AttachAndBroadcast(t *testing.T) {
	bridge := NewAttachBridge(nil)

	// Anexa 2 clientes à mesma estação ("builder")
	ch1, detach1, err := bridge.Attach(queue.StationBuilder)
	require.NoError(t, err)
	defer detach1()

	ch2, detach2, err := bridge.Attach(queue.StationBuilder)
	require.NoError(t, err)
	defer detach2()

	// Anexa 1 cliente a outra estação ("planner")
	chPlanner, detachPlanner, err := bridge.Attach(queue.StationPlanner)
	require.NoError(t, err)
	defer detachPlanner()

	assert.Equal(t, 2, bridge.AttachedClients(queue.StationBuilder))
	assert.Equal(t, 1, bridge.AttachedClients(queue.StationPlanner))

	// Transmite logs concorrentemente para "builder"
	const numMessages = 10
	var wg sync.WaitGroup
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bridge.BroadcastLog(queue.StationBuilder, SessionLogEntry{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Type:      "step",
				Message:   fmt.Sprintf("Executando tarefa %d", idx),
			})
		}(i)
	}
	wg.Wait()

	// Verifica recebimento no canal 1
	received1 := 0
	timeout := time.After(2 * time.Second)
recvLoop1:
	for {
		select {
		case entry, ok := <-ch1:
			require.True(t, ok)
			assert.Contains(t, entry.Message, "Executando tarefa")
			received1++
			if received1 == numMessages {
				break recvLoop1
			}
		case <-timeout:
			t.Fatalf("timeout aguardando mensagens no canal 1 (recebidas: %d/%d)", received1, numMessages)
		}
	}

	// Verifica recebimento no canal 2
	received2 := 0
	timeout2 := time.After(2 * time.Second)
recvLoop2:
	for {
		select {
		case entry, ok := <-ch2:
			require.True(t, ok)
			assert.Contains(t, entry.Message, "Executando tarefa")
			received2++
			if received2 == numMessages {
				break recvLoop2
			}
		case <-timeout2:
			t.Fatalf("timeout aguardando mensagens no canal 2 (recebidas: %d/%d)", received2, numMessages)
		}
	}

	// Planner não deve ter recebido nenhuma mensagem de builder
	select {
	case entry := <-chPlanner:
		t.Fatalf("planner recebeu mensagem indevida: %+v", entry)
	default:
		// Esperado: canal vazio
	}

	// Verifica histórico gravado no bridge
	history := bridge.GetHistory(queue.StationBuilder)
	assert.Len(t, history, numMessages)
}

func TestAttachBridge_GracefulDetach(t *testing.T) {
	tempDir := t.TempDir()
	dispatcher := NewDispatcher()
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	_, err := pool.Warmup(context.Background(), tempDir)
	require.NoError(t, err)

	bridge := NewAttachBridge(pool)
	dispatcher.SetAttachBridge(bridge)

	// Sessão builder antes do attach
	sessBefore, ok := pool.GetSessionByStation(queue.StationBuilder)
	require.True(t, ok)
	assert.Equal(t, SessionStateStandby, sessBefore.State)
	assert.Equal(t, 0, sessBefore.AttachedClients)

	// 1. Anexar cliente
	ch, detach, err := bridge.Attach(queue.StationBuilder)
	require.NoError(t, err)

	sessAttached, ok := pool.GetSessionByStation(queue.StationBuilder)
	require.True(t, ok)
	assert.Equal(t, 1, sessAttached.AttachedClients)
	assert.Equal(t, 1, bridge.AttachedClients(queue.StationBuilder))

	// 2. Desconectar graciosamente (detach)
	detach()

	// Desanexação repetida deve ser idempotente
	assert.NotPanics(t, func() {
		detach()
	})

	// Canal deve ter sido fechado
	_, ok = <-ch
	assert.False(t, ok, "canal deveria estar fechado após detach")

	// Sessão deve continuar em standby e não cancelada
	sessAfter, ok := pool.GetSessionByStation(queue.StationBuilder)
	require.True(t, ok)
	assert.Equal(t, SessionStateStandby, sessAfter.State, "desconexão nunca deve matar ou alterar o estado da sessão no pool")
	assert.Equal(t, 0, sessAfter.AttachedClients)
	assert.Equal(t, 0, bridge.AttachedClients(queue.StationBuilder))

	// 3. Testa Run() com cancelamento gracioso via contexto
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancela de imediato

	var outBuf bytes.Buffer
	err = bridge.Run(ctx, queue.StationBuilder, strings.NewReader(""), &outBuf)
	assert.NoError(t, err, "Run deve retornar nil após cancelamento gracioso de contexto")
	assert.Contains(t, outBuf.String(), "🔌 FORGE-SDD ATTACH")
	assert.Contains(t, outBuf.String(), "🟢 Standby")
	assert.Contains(t, outBuf.String(), "[detach]")

	// Sessão ainda no pool em standby
	sessFinal, ok := pool.GetSessionByStation(queue.StationBuilder)
	require.True(t, ok)
	assert.Equal(t, SessionStateStandby, sessFinal.State)
	assert.Equal(t, 0, sessFinal.AttachedClients)

	// 4. Testa Run() com comando "exit"
	var outBuf2 bytes.Buffer
	err = bridge.Run(context.Background(), queue.StationBuilder, strings.NewReader("exit\n"), &outBuf2)
	assert.NoError(t, err, "Run deve retornar nil ao receber comando exit")
	assert.Contains(t, outBuf2.String(), "comando exit")
}

func TestAttachBridge_InvalidStation(t *testing.T) {
	bridge := NewAttachBridge(nil)

	// Estação desconhecida
	ch, detach, err := bridge.Attach("estacao-inexistente")
	assert.Error(t, err)
	assert.Nil(t, ch)
	assert.Nil(t, detach)
	assert.Contains(t, err.Error(), "estação inválida")
	for _, canonical := range queue.CanonicalStations() {
		assert.Contains(t, err.Error(), canonical)
	}

	// Estação vazia
	ch2, detach2, err2 := bridge.Attach("")
	assert.Error(t, err2)
	assert.Nil(t, ch2)
	assert.Nil(t, detach2)

	// Broadcast para estação inválida é no-op seguro
	assert.NotPanics(t, func() {
		bridge.BroadcastLog("invalida", SessionLogEntry{
			Message: "teste",
		})
	})
}

func TestRPCSessionAttach_Handler(t *testing.T) {
	tempDir := t.TempDir()
	dispatcher := NewDispatcher()
	pool := dispatcher.Pool()
	require.NotNil(t, pool)

	_, err := pool.Warmup(context.Background(), tempDir)
	require.NoError(t, err)

	bridge := dispatcher.AttachBridge()
	require.NotNil(t, bridge)

	// Emite um log prévio na estação builder
	bridge.BroadcastLog(queue.StationBuilder, SessionLogEntry{
		Timestamp: "2026-11-01T02:00:00Z",
		Type:      "info",
		Message:   "Sessão inicializada em standby",
	})

	// 1. Chamada session/attach com estação válida
	reqBytes := []byte(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "session/attach",
		"params": {
			"station": "builder"
		}
	}`)
	resp := dispatcher.Dispatch(context.Background(), reqBytes)
	require.NotNil(t, resp)
	assert.Nil(t, resp.Error)

	var result SessionAttachResult
	resultBytes, _ := json.Marshal(resp.Result)
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "sess-04-builder", result.SessionID)
	assert.Equal(t, queue.StationBuilder, result.Station)
	assert.Equal(t, SessionStateStandby, result.State)
	assert.NotEmpty(t, result.History)
	assert.Equal(t, "Sessão inicializada em standby", result.History[0].Message)

	// 2. Chamada session/attach com sessionId
	reqBytes2 := []byte(`{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "session/attach",
		"params": {
			"sessionId": "sess-03-planner"
		}
	}`)
	resp2 := dispatcher.Dispatch(context.Background(), reqBytes2)
	require.NotNil(t, resp2)
	assert.Nil(t, resp2.Error)

	var result2 SessionAttachResult
	resultBytes2, _ := json.Marshal(resp2.Result)
	err = json.Unmarshal(resultBytes2, &result2)
	require.NoError(t, err)
	assert.Equal(t, "sess-03-planner", result2.SessionID)
	assert.Equal(t, queue.StationPlanner, result2.Station)

	// 3. Chamada session/attach com estação inválida
	reqBytesInvalid := []byte(`{
		"jsonrpc": "2.0",
		"id": 3,
		"method": "session/attach",
		"params": {
			"station": "devops"
		}
	}`)
	resp3 := dispatcher.Dispatch(context.Background(), reqBytesInvalid)
	require.NotNil(t, resp3)
	assert.NotNil(t, resp3.Error)
	assert.Equal(t, CodeInvalidParams, resp3.Error.Code)
	assert.Contains(t, resp3.Error.Message, "estação inválida")
}
