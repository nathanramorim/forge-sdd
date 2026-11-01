package acp_test

import (
	"context"
	"encoding/json"
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

// TestACPSessionLifecycle valida o ciclo de vida completo de sessões, isolamento de workspace (SEG-01, SEG-02)
// e o despacho de comandos via session/prompt (CA-01).
func TestACPSessionLifecycle(t *testing.T) {
	dispatcher := acp.NewDispatcher()
	wsDir := t.TempDir()

	// 1. Criação de sessão válida
	newReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      100,
		"method":  "session/new",
		"params": acp.SessionNewParams{
			WorkspaceDir: wsDir,
			Feature:      "feat-64-02",
			Branch:       "feat/64-acp-sessions",
		},
	}
	rawNew, err := json.Marshal(newReq)
	require.NoError(t, err)

	respNew := dispatcher.Dispatch(context.Background(), rawNew)
	require.NotNil(t, respNew)
	require.Nil(t, respNew.Error)

	var resNew acp.SessionResult
	rawRes, err := json.Marshal(respNew.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawRes, &resNew))

	assert.NotEmpty(t, resNew.SessionID)
	assert.Equal(t, "active", resNew.Status)
	assert.Equal(t, wsDir, resNew.WorkspaceDir)

	// 2. Carregamento de sessão existente (session/load)
	loadReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      101,
		"method":  "session/load",
		"params": acp.SessionLoadParams{
			SessionID: resNew.SessionID,
		},
	}
	rawLoad, err := json.Marshal(loadReq)
	require.NoError(t, err)

	respLoad := dispatcher.Dispatch(context.Background(), rawLoad)
	require.NotNil(t, respLoad)
	require.Nil(t, respLoad.Error)

	var resLoad acp.SessionResult
	rawResLoad, err := json.Marshal(respLoad.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawResLoad, &resLoad))
	assert.Equal(t, resNew.SessionID, resLoad.SessionID)
	assert.Equal(t, "loaded", resLoad.Status)

	// 3. Isolamento e validação de Workspace (SEG-01, SEG-02)
	t.Run("Rejeita diretório inexistente", func(t *testing.T) {
		nonExistentDir := filepath.Join(wsDir, "dir-que-nao-existe")
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      102,
			"method":  "session/new",
			"params": acp.SessionNewParams{
				WorkspaceDir: nonExistentDir,
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "não existe")
	})

	t.Run("Rejeita caminho com caractere nulo", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      103,
			"method":  "session/new",
			"params": acp.SessionNewParams{
				WorkspaceDir: wsDir + "\x00/hack",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})

	t.Run("Rejeita path traversal fora dos limites permitidos", func(t *testing.T) {
		customMgr := acp.NewSessionManager(wsDir)
		customDispatcher := acp.NewDispatcher()
		customDispatcher.SetSessionManager(customMgr)

		outsideDir := t.TempDir()
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      104,
			"method":  "session/new",
			"params": acp.SessionNewParams{
				WorkspaceDir: outsideDir,
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := customDispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "path traversal")
	})

	// 4. Inscrição em streaming de notificações (session/notification)
	notifCh := make(chan acp.Notification, 10)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	// 5. Envio de prompt para enfileiramento na estação ativa (session/prompt)
	promptReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      105,
		"method":  "session/prompt",
		"params": acp.SessionPromptParams{
			SessionID: resNew.SessionID,
			Prompt:    "Implementar feature de sessão ACP com fila",
			Station:   queue.StationPMArchitect,
			Metadata: map[string]string{
				"itemId": "prompt-01",
			},
		},
	}
	rawPrompt, err := json.Marshal(promptReq)
	require.NoError(t, err)

	respPrompt := dispatcher.Dispatch(context.Background(), rawPrompt)
	require.NotNil(t, respPrompt)
	require.Nil(t, respPrompt.Error)

	var resPrompt acp.SessionPromptResult
	rawResPrompt, err := json.Marshal(respPrompt.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawResPrompt, &resPrompt))

	assert.Equal(t, resNew.SessionID, resPrompt.SessionID)
	assert.Equal(t, queue.StationPMArchitect, resPrompt.Station)
	assert.Equal(t, queue.StatusQueued, resPrompt.Status)
	assert.Equal(t, "prompt-01", resPrompt.Item.ID)

	// Verifica se item foi fisicamente gravado na fila da estação (internal/queue)
	sq, err := queue.LoadStationQueue(wsDir, queue.StationPMArchitect)
	require.NoError(t, err)
	require.NotEmpty(t, sq.Items)
	assert.Equal(t, "prompt-01", sq.Items[0].ID)

	// Verifica recebimento da notificação de streaming
	select {
	case n := <-notifCh:
		assert.Equal(t, "session/notification", n.Method)
		notifBytes, err := json.Marshal(n.Params)
		require.NoError(t, err)
		var sn acp.SessionNotification
		require.NoError(t, json.Unmarshal(notifBytes, &sn))
		assert.Equal(t, "prompt_enqueued", sn.Type)
		assert.Equal(t, queue.StationPMArchitect, sn.Station)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout aguardando evento no canal session/notification")
	}

	// 6. Testes negativos de session/prompt
	t.Run("Rejeita prompt vazio", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      106,
			"method":  "session/prompt",
			"params": acp.SessionPromptParams{
				SessionID: resNew.SessionID,
				Prompt:    "   ",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})

	t.Run("Rejeita sessionId inexistente", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      107,
			"method":  "session/prompt",
			"params": acp.SessionPromptParams{
				SessionID: "sessao-inexistente",
				Prompt:    "algo",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})

	t.Run("Rejeita estação inválida", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      108,
			"method":  "session/prompt",
			"params": acp.SessionPromptParams{
				SessionID: resNew.SessionID,
				Prompt:    "algo",
				Station:   "estacao-invalida",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)
		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})

	// 7. Cancelamento de sessão (session/cancel)
	cancelReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      109,
		"method":  "session/cancel",
		"params": acp.SessionCancelParams{
			SessionID: resNew.SessionID,
			Reason:    "finalizado com sucesso",
		},
	}
	rawCancel, err := json.Marshal(cancelReq)
	require.NoError(t, err)
	respCancel := dispatcher.Dispatch(context.Background(), rawCancel)
	require.NotNil(t, respCancel)
	require.Nil(t, respCancel.Error)

	var resCancel acp.SessionResult
	rawResCancel, err := json.Marshal(respCancel.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawResCancel, &resCancel))
	assert.Equal(t, "cancelled", resCancel.Status)
}

// TestACPGateApproval valida a homologação e o handoff determinístico do Portão FLX-01,
// além de rejeições canônicas com código de erro -32001 (CA-02).
func TestACPGateApproval(t *testing.T) {
	dispatcher := acp.NewDispatcher()
	wsDir := t.TempDir()

	specRel := "sdd/features/feat-64-protocolo-acp/feat-64-02.md"
	specFull := filepath.Join(wsDir, specRel)
	require.NoError(t, os.MkdirAll(filepath.Dir(specFull), 0o755))

	validSpecContent := `# feat-64-02 — Protocolo ACP

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor de Filas | internal/queue | EnqueueStationItem | reusar | Reuso de infraestrutura |

## Dependências novas
- Nenhuma dependência externa.
`
	require.NoError(t, os.WriteFile(specFull, []byte(validSpecContent), 0o644))

	// Inscreve subscriber para validar notificação de aprovação
	notifCh := make(chan acp.Notification, 10)
	dispatcher.SubscribeNotifications(notifCh)
	defer dispatcher.UnsubscribeNotifications(notifCh)

	// Cria sessão no workspace
	newResp := dispatcher.Dispatch(context.Background(), []byte(fmt.Sprintf(
		`{"jsonrpc":"2.0","id":1,"method":"session/new","params":{"workspaceDir":%q,"branch":"feat/64-acp"}}`,
		wsDir,
	)))
	require.NotNil(t, newResp)
	require.Nil(t, newResp.Error)

	var sessRes acp.SessionResult
	rawSess, err := json.Marshal(newResp.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawSess, &sessRes))

	t.Run("Aprovação com sucesso do Portão FLX-01", func(t *testing.T) {
		approveReq := map[string]any{
			"jsonrpc": "2.0",
			"id":      200,
			"method":  "gate/approve",
			"params": acp.GateApproveParams{
				SessionID: sessRes.SessionID,
				SpecPath:  specRel,
				FeatureID: "feat-64-02",
				Title:     "Sessões ACP e Portão FLX-01",
			},
		}
		rawApprove, err := json.Marshal(approveReq)
		require.NoError(t, err)

		respApprove := dispatcher.Dispatch(context.Background(), rawApprove)
		require.NotNil(t, respApprove)
		require.Nil(t, respApprove.Error)

		var resGate acp.GateApproveResult
		rawResGate, err := json.Marshal(respApprove.Result)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(rawResGate, &resGate))

		assert.Equal(t, "feat-64-02", resGate.FeatureID)
		assert.Equal(t, "FLX-01", resGate.Gate)
		assert.Equal(t, queue.StatusApproved, resGate.Status)
		assert.Equal(t, queue.StationPMArchitect, resGate.FromStation)
		assert.Equal(t, queue.StationPlanner, resGate.ToStation)
		require.NotNil(t, resGate.HandoffItem)
		assert.Equal(t, "feat-64-02", resGate.HandoffItem.ID)
		assert.Equal(t, queue.StatusQueued, resGate.HandoffItem.Status)

		// 1. Verifica se Estação 02 (origem) está marcada como 'approved'
		pmSQ, err := queue.LoadStationQueue(wsDir, queue.StationPMArchitect)
		require.NoError(t, err)
		require.NotEmpty(t, pmSQ.Items)
		assert.Equal(t, queue.StatusApproved, pmSQ.Items[0].Status)

		// 2. Verifica se Estação 03 (destino) recebeu o handoff
		plannerSQ, err := queue.LoadStationQueue(wsDir, queue.StationPlanner)
		require.NoError(t, err)
		require.NotEmpty(t, plannerSQ.Items)
		assert.Equal(t, "feat-64-02", plannerSQ.Items[0].ID)
		assert.Equal(t, queue.StatusQueued, plannerSQ.Items[0].Status)

		// 3. Notificação recebida no streaming
		select {
		case n := <-notifCh:
			assert.Equal(t, "session/notification", n.Method)
			var sn acp.SessionNotification
			rawN, err := json.Marshal(n.Params)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(rawN, &sn))
			assert.Equal(t, "gate_approved", sn.Type)
			assert.Equal(t, queue.StationPlanner, sn.Station)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout aguardando notificação de gate/approve")
		}
	})

	t.Run("Rejeição com erro -32001 quando Spec não existe (P-02)", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      201,
			"method":  "gate/approve",
			"params": acp.GateApproveParams{
				SessionID: sessRes.SessionID,
				SpecPath:  "sdd/features/spec-inexistente.md",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeServerError, resp.Error.Code) // -32001
		assert.Contains(t, resp.Error.Message, "P-02")
	})

	t.Run("Rejeição com erro -32001 quando tabela de reuso ausente (P-03 / DUP-02)", func(t *testing.T) {
		invalidSpecRel := "sdd/features/feat-invalida.md"
		invalidSpecFull := filepath.Join(wsDir, invalidSpecRel)
		require.NoError(t, os.WriteFile(invalidSpecFull, []byte("# Feature Sem Reúso\nCorpo sem tabela.\n"), 0o644))

		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      202,
			"method":  "gate/approve",
			"params": acp.GateApproveParams{
				SessionID: sessRes.SessionID,
				SpecPath:  invalidSpecRel,
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeServerError, resp.Error.Code) // -32001
		assert.Contains(t, resp.Error.Message, "P-03")
	})

	t.Run("Rejeição com erro -32001 quando branch é main ou master (P-01 / FLX-02)", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      203,
			"method":  "gate/approve",
			"params": acp.GateApproveParams{
				SessionID: sessRes.SessionID,
				SpecPath:  specRel,
				Branch:    "main",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeServerError, resp.Error.Code) // -32001
		assert.Contains(t, resp.Error.Message, "P-01")
	})

	t.Run("Rejeição com erro -32602 quando specPath vazio", func(t *testing.T) {
		req := map[string]any{
			"jsonrpc": "2.0",
			"id":      204,
			"method":  "gate/approve",
			"params": acp.GateApproveParams{
				SessionID: sessRes.SessionID,
				SpecPath:  "",
			},
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		resp := dispatcher.Dispatch(context.Background(), raw)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, acp.CodeInvalidParams, resp.Error.Code)
	})
}

// TestACPSessionAndGate valida concorrência intensiva (-race) entre sessões,
// despacho de prompts e homologação de portões nas 5 estações (CA-03).
func TestACPSessionAndGate(t *testing.T) {
	dispatcher := acp.NewDispatcher()
	wsDir := t.TempDir()

	// Prepara arquivo de Spec válido para testes de aprovação
	specRel := "sdd/features/concorrente/spec.md"
	specFull := filepath.Join(wsDir, specRel)
	require.NoError(t, os.MkdirAll(filepath.Dir(specFull), 0o755))
	specContent := `# Concorrência Spec

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Queue | internal/queue | LoadStationQueue | reusar | Reúso concorrente |
`
	require.NoError(t, os.WriteFile(specFull, []byte(specContent), 0o644))

	// Cria sessão inicial
	newReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      300,
		"method":  "session/new",
		"params": acp.SessionNewParams{
			WorkspaceDir: wsDir,
			Feature:      "feat-race-test",
			Branch:       "feat/race",
		},
	}
	rawNew, err := json.Marshal(newReq)
	require.NoError(t, err)
	respNew := dispatcher.Dispatch(context.Background(), rawNew)
	require.NotNil(t, respNew)
	require.Nil(t, respNew.Error)

	var sessRes acp.SessionResult
	rawSess, err := json.Marshal(respNew.Result)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rawSess, &sessRes))

	const concurrentWorkers = 25
	var wg sync.WaitGroup
	wg.Add(concurrentWorkers * 2)

	// Goroutines enviando session/prompt concorrentemente
	for i := 0; i < concurrentWorkers; i++ {
		go func(idx int) {
			defer wg.Done()
			req := map[string]any{
				"jsonrpc": "2.0",
				"id":      4000 + idx,
				"method":  "session/prompt",
				"params": acp.SessionPromptParams{
					SessionID: sessRes.SessionID,
					Prompt:    fmt.Sprintf("Demanda concorrente #%d", idx),
					Station:   queue.StationPMArchitect,
					Metadata: map[string]string{
						"itemId": fmt.Sprintf("concurrent-%d", idx),
					},
				},
			}
			raw, err := json.Marshal(req)
			if err != nil {
				return
			}
			resp := dispatcher.Dispatch(context.Background(), raw)
			if resp == nil || resp.Error != nil {
				t.Errorf("falha no prompt concorrente #%d: %v", idx, resp)
			}
		}(i)
	}

	// Goroutines executando gate/approve concorrentemente (idempotência e atomicidade)
	for i := 0; i < concurrentWorkers; i++ {
		go func(idx int) {
			defer wg.Done()
			req := map[string]any{
				"jsonrpc": "2.0",
				"id":      5000 + idx,
				"method":  "gate/approve",
				"params": acp.GateApproveParams{
					SessionID:      sessRes.SessionID,
					SpecPath:       specRel,
					FeatureID:      "feat-race-test",
					IdempotencyKey: fmt.Sprintf("race-gate-key-%d", idx%3),
				},
			}
			raw, err := json.Marshal(req)
			if err != nil {
				return
			}
			resp := dispatcher.Dispatch(context.Background(), raw)
			if resp == nil || resp.Error != nil {
				t.Errorf("falha no gate approve concorrente #%d: %v", idx, resp)
			}
		}(i)
	}

	wg.Wait()

	// Valida integridade final das filas persistidas
	pmSQ, err := queue.LoadStationQueue(wsDir, queue.StationPMArchitect)
	require.NoError(t, err)
	assert.NotEmpty(t, pmSQ.Items)

	plannerSQ, err := queue.LoadStationQueue(wsDir, queue.StationPlanner)
	require.NoError(t, err)
	assert.NotEmpty(t, plannerSQ.Items)
}
