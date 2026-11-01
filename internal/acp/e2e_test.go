package acp_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clientMessage representa uma mensagem JSON-RPC recebida pelo cliente (resposta ou notificação).
type clientMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *acp.RPCError   `json:"error,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// TestACPEndToEnd valida a jornada ponta a ponta (E2E) de um cliente IDE externo (como Zed ou Cursor)
// comunicando-se com o servidor forge-sdd acp via transporte bidirecional stdio (pipes).
//
// Etapas da jornada simulada:
// 1. Handshake "initialize"
// 2. Abertura de sessão isolada com "session/new"
// 3. Enfileiramento de demanda com "session/prompt"
// 4. Recepção e validação de notificações de streaming assíncrono ("session/notification")
// 5. Homologação do Portão FLX-01 com handoff determinístico via "gate/approve"
// 6. Encerramento ordenado da sessão com "session/cancel"
func TestACPEndToEnd(t *testing.T) {
	// Cria workspace isolado para a sessão de teste
	wsDir := t.TempDir()

	// Prepara a estrutura mínima de diretórios e a especificação da feature para validação de FLX-01
	specRelPath := filepath.Join("sdd", "features", "feat-e2e-demo", "feat-e2e-demo.md")
	specAbsPath := filepath.Join(wsDir, specRelPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(specAbsPath), 0o755))

	validSpecContent := `# feat-e2e-demo — Validação E2E do Protocolo ACP

## Modelo do módulo
- **Modelo:** enxuto.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Protocolo ACP | internal/acp | StdioTransport | reusar | Reuso de infraestrutura canônica |

## Dependências novas
- Nenhuma dependência externa.

## Critérios de Aceitação Executáveis
1. Validação E2E completa via stdio pipes.
`
	require.NoError(t, os.WriteFile(specAbsPath, []byte(validSpecContent), 0o644))

	// Configuração do despachador e do transporte stdio via pipes simulados
	dispatcher := acp.NewDispatcherWithVersion("3.2.1-beta")

	// serverIn: cliente escreve -> servidor lê
	serverInReader, clientWriter := io.Pipe()
	// serverOut: servidor escreve -> cliente lê
	clientReader, serverOutWriter := io.Pipe()

	transport := acp.NewStdioTransport(dispatcher, serverInReader, serverOutWriter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicia o loop do transporte stdio em goroutine dedicada
	transportDone := make(chan error, 1)
	go func() {
		transportDone <- transport.Run(ctx)
	}()

	// Canais para despachar mensagens recebidas pelo cliente externo
	responsesCh := make(chan clientMessage, 100)
	notificationsCh := make(chan clientMessage, 100)
	readerDone := make(chan struct{})

	// Goroutine do cliente para leitura de linhas do servidor via clientReader
	go func() {
		defer close(readerDone)
		scanner := bufio.NewScanner(clientReader)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var msg clientMessage
			if err := json.Unmarshal(line, &msg); err != nil {
				continue
			}

			if msg.ID != nil {
				responsesCh <- msg
			} else if msg.Method == "session/notification" {
				notificationsCh <- msg
			}
		}
	}()

	// Helper do cliente para enviar chamadas RPC
	sendRPC := func(id any, method string, params any) {
		reqMap := map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"method":  method,
		}
		if params != nil {
			reqMap["params"] = params
		}
		data, err := json.Marshal(reqMap)
		require.NoError(t, err)
		data = append(data, '\n')
		_, err = clientWriter.Write(data)
		require.NoError(t, err)
	}

	// Helper do cliente para aguardar resposta correspondente
	awaitResponse := func(expectedID any, timeout time.Duration) clientMessage {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case resp := <-responsesCh:
				if fmt.Sprint(resp.ID) == fmt.Sprint(expectedID) {
					return resp
				}
			case <-timer.C:
				t.Fatalf("timeout aguardando resposta para ID %v", expectedID)
				return clientMessage{}
			}
		}
	}

	// Helper do cliente para aguardar notificação de streaming
	awaitNotification := func(expectedType string, timeout time.Duration) acp.SessionNotification {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case notifMsg := <-notificationsCh:
				var notif acp.SessionNotification
				if err := json.Unmarshal(notifMsg.Params, &notif); err == nil {
					if expectedType == "" || notif.Type == expectedType {
						return notif
					}
				}
			case <-timer.C:
				t.Fatalf("timeout aguardando notificação do tipo %q", expectedType)
				return acp.SessionNotification{}
			}
		}
	}

	// =========================================================================
	// 1. Handshake "initialize"
	// =========================================================================
	t.Log("1. Executando handshake 'initialize'...")
	sendRPC(1, "initialize", acp.InitializeParams{
		ProtocolVersion: "1.0.0",
		ClientInfo: acp.ClientInfo{
			Name:    "zed-editor",
			Version: "0.140.0",
		},
		Capabilities: &acp.ClientCapabilities{
			Terminal: true,
			Prompts:  true,
		},
	})

	initResp := awaitResponse(1, 2*time.Second)
	require.Nil(t, initResp.Error, "handshake initialize não deve retornar erro")

	var initResult acp.InitializeResult
	require.NoError(t, json.Unmarshal(initResp.Result, &initResult))
	assert.Equal(t, acp.DefaultProtocolVersion, initResult.ProtocolVersion)
	assert.Equal(t, acp.ServerName, initResult.ServerInfo.Name)
	assert.Equal(t, "3.2.1-beta", initResult.ServerInfo.Version)
	assert.True(t, initResult.Capabilities.Sessions)
	assert.True(t, initResult.Capabilities.Gates)
	assert.True(t, initResult.Capabilities.Prompts)
	assert.True(t, dispatcher.IsInitialized())

	// =========================================================================
	// 2. Abertura de Sessão Isolada ("session/new")
	// =========================================================================
	t.Log("2. Abertura de nova sessão 'session/new'...")
	sendRPC(2, "session/new", acp.SessionNewParams{
		WorkspaceDir: wsDir,
		Feature:      "feat-e2e-demo",
		Branch:       "feat/e2e-demo",
	})

	newResp := awaitResponse(2, 2*time.Second)
	require.Nil(t, newResp.Error, "session/new não deve retornar erro")

	var newResult acp.SessionResult
	require.NoError(t, json.Unmarshal(newResp.Result, &newResult))
	assert.NotEmpty(t, newResult.SessionID)
	assert.Equal(t, "active", newResult.Status)
	assert.Equal(t, wsDir, newResult.WorkspaceDir)

	sessionID := newResult.SessionID

	// =========================================================================
	// 3. Enfileiramento de Demanda ("session/prompt")
	// =========================================================================
	t.Log("3. Enviando prompt para a estação de Concepção 'session/prompt'...")
	sendRPC(3, "session/prompt", acp.SessionPromptParams{
		SessionID: sessionID,
		Station:   queue.StationPMArchitect,
		Prompt:    "Elaborar arquitetura inicial para feat-e2e-demo",
		Operation: "spec",
		Metadata: map[string]string{
			"origem": "zed-acp-client",
		},
	})

	promptResp := awaitResponse(3, 2*time.Second)
	require.Nil(t, promptResp.Error, "session/prompt não deve retornar erro")

	var promptResult acp.SessionPromptResult
	require.NoError(t, json.Unmarshal(promptResp.Result, &promptResult))
	assert.Equal(t, sessionID, promptResult.SessionID)
	assert.Equal(t, queue.StationPMArchitect, promptResult.Station)
	assert.Equal(t, queue.StatusQueued, promptResult.Status)
	require.NotNil(t, promptResult.Item)
	assert.Equal(t, "Elaborar arquitetura inicial para feat-e2e-demo", promptResult.Item.Title)

	// Valida persistência da fila da Estação 02 no filesystem (sdd/.queue/stations/pm-architect.json)
	pmQueue, err := queue.LoadStationQueue(wsDir, queue.StationPMArchitect)
	require.NoError(t, err)
	assert.NotEmpty(t, pmQueue.Items)
	assert.Equal(t, promptResult.Item.ID, pmQueue.Items[len(pmQueue.Items)-1].ID)

	// =========================================================================
	// 4. Recepção de Notificações de Streaming Assíncrono
	// =========================================================================
	t.Log("4. Aguardando notificação 'prompt_enqueued' de streaming assíncrono...")
	promptNotif := awaitNotification("prompt_enqueued", 2*time.Second)
	assert.Equal(t, sessionID, promptNotif.SessionID)
	assert.Equal(t, queue.StationPMArchitect, promptNotif.Station)
	assert.Contains(t, promptNotif.Message, "enfileirado na estação pm-architect")

	// =========================================================================
	// 5. Homologação do Portão FLX-01 com Handoff ("gate/approve")
	// =========================================================================
	t.Log("5. Homologando Portão FLX-01 via 'gate/approve'...")
	sendRPC(4, "gate/approve", acp.GateApproveParams{
		SessionID:      sessionID,
		WorkspaceDir:   wsDir,
		FeatureID:      "feat-e2e-demo",
		SpecPath:       specRelPath,
		Branch:         "feat/e2e-demo",
		FromStation:    queue.StationPMArchitect,
		ToStation:      queue.StationPlanner,
		Title:          "Feature E2E Demo Homologada",
		InputContract:  "Spec de Arquitetura",
		OutputContract: "Plano de Quebra em Tasks",
	})

	gateResp := awaitResponse(4, 2*time.Second)
	require.Nil(t, gateResp.Error, "gate/approve deve aprovar o Portão FLX-01 com sucesso")

	var gateResult acp.GateApproveResult
	require.NoError(t, json.Unmarshal(gateResp.Result, &gateResult))
	assert.Equal(t, "feat-e2e-demo", gateResult.FeatureID)
	assert.Equal(t, "FLX-01", gateResult.Gate)
	assert.Equal(t, queue.StatusApproved, gateResult.Status)
	assert.Equal(t, queue.StationPMArchitect, gateResult.FromStation)
	assert.Equal(t, queue.StationPlanner, gateResult.ToStation)
	require.NotNil(t, gateResult.HandoffItem)
	assert.Equal(t, queue.StationPlanner, gateResult.HandoffItem.Station)

	// Valida persistência atômica do handoff na fila da Estação 03 (planner)
	plannerQueue, err := queue.LoadStationQueue(wsDir, queue.StationPlanner)
	require.NoError(t, err)
	assert.NotEmpty(t, plannerQueue.Items)
	assert.Equal(t, "feat-e2e-demo", plannerQueue.Items[len(plannerQueue.Items)-1].ID)

	// Valida recebimento da notificação de streaming 'gate_approved'
	t.Log("5.1 Aguardando notificação 'gate_approved'...")
	gateNotif := awaitNotification("gate_approved", 2*time.Second)
	assert.Equal(t, sessionID, gateNotif.SessionID)
	assert.Equal(t, queue.StationPlanner, gateNotif.Station)
	assert.Contains(t, gateNotif.Message, "Portão FLX-01 aprovado")

	// =========================================================================
	// 6. Encerramento Ordenado da Sessão ("session/cancel")
	// =========================================================================
	t.Log("6. Encerrando sessão com 'session/cancel'...")
	sendRPC(5, "session/cancel", acp.SessionCancelParams{
		SessionID: sessionID,
		Reason:    "Jornada de testes E2E concluída com 100% de sucesso",
	})

	cancelResp := awaitResponse(5, 2*time.Second)
	require.Nil(t, cancelResp.Error, "session/cancel não deve retornar erro")

	var cancelResult acp.SessionResult
	require.NoError(t, json.Unmarshal(cancelResp.Result, &cancelResult))
	assert.Equal(t, sessionID, cancelResult.SessionID)
	assert.Equal(t, "cancelled", cancelResult.Status)

	// Confirma que a sessão foi atualizada no SessionManager
	storedSess, ok := dispatcher.SessionManager().GetSession(sessionID)
	require.True(t, ok)
	assert.Equal(t, "cancelled", storedSess.Status)
	assert.Equal(t, "Jornada de testes E2E concluída com 100% de sucesso", storedSess.Metadata["cancelReason"])

	// =========================================================================
	// 7. Fechamento Limpo dos Streams e Encerramento
	// =========================================================================
	t.Log("7. Encerrando pipes e aguardando desligamento limpo do transporte...")
	require.NoError(t, clientWriter.Close())

	select {
	case err := <-transportDone:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout aguardando encerramento do transporte stdio")
	}

	require.NoError(t, serverOutWriter.Close())

	select {
	case <-readerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout aguardando encerramento do reader do cliente")
	}
}
