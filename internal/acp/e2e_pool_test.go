package acp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/studio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_PoolLifecycleAndStudioIntegration valida a integração completa do SessionPool com o Forge Studio:
// 1. Arranque do Studio e pré-aquecimento automático das 5 sessões (standby)
// 2. Despacho e consulta de status via HTTP JSON-RPC (/api/acp)
// 3. Transição operacional da Estação 02 (PM & Arquiteto) -> processing (mascote working)
// 4. Conclusão da Spec -> bloqueio estrito em approval_pending (mascote approval)
// 5. Homologação do Portão FLX-01 via RPC gate/approve e handoff determinístico para Estação 03 (Planejador)
// 6. Anexação de cliente terminal via AttachBridge (pub/sub de logs em tempo real)
// 7. Desconexão graciosa sem encerramento das sessões no pool
func TestE2E_PoolLifecycleAndStudioIntegration(t *testing.T) {
	wsDir := t.TempDir()

	// 1. Prepara estrutura SDD mínima e arquivo de especificação para homologação do Portão FLX-01
	specRelPath := filepath.Join("sdd", "features", "feat-e2e-pool", "feat-e2e-pool.md")
	specAbsPath := filepath.Join(wsDir, specRelPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(specAbsPath), 0o755))

	validSpecContent := `# feat-e2e-pool — Teste E2E do Pool de Sessões

## Modelo do módulo
- **Modelo:** enxuto.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| SessionPool | internal/acp | SessionPool | reusar | Gerenciador canônico de pool |

## Dependências novas
- Nenhuma dependência externa.
`
	require.NoError(t, os.WriteFile(specAbsPath, []byte(validSpecContent), 0o644))

	// 2. Inicializa o servidor Forge Studio
	srv, err := studio.NewServer(studio.Config{
		Dir:  wsDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	require.NoError(t, err)
	require.NoError(t, srv.Listen())
	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	serverURL := srv.URL()
	require.NotEmpty(t, serverURL)

	// 3. Valida se o SessionPool foi pré-aquecido com as 5 sessões canônicas em standby
	pool := srv.Pool()
	require.NotNil(t, pool)

	sessions := pool.ListSessions()
	require.Len(t, sessions, 5)
	for _, s := range sessions {
		assert.Equal(t, acp.SessionStateStandby, s.State)
		assert.NotEmpty(t, s.SessionID)
	}

	// 4. Consulta status do pool via HTTP POST /api/acp (JSON-RPC 2.0)
	statusReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "pool/status",
		"params":  map[string]any{},
	}
	body, err := json.Marshal(statusReq)
	require.NoError(t, err)

	httpResp, err := http.Post(fmt.Sprintf("%s/api/acp", serverURL), "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer httpResp.Body.Close()
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	var rpcResp struct {
		JSONRPC string             `json:"jsonrpc"`
		ID      int                `json:"id"`
		Result  acp.PoolStatusResult `json:"result"`
	}
	require.NoError(t, json.NewDecoder(httpResp.Body).Decode(&rpcResp))
	assert.True(t, rpcResp.Result.Healthy)
	assert.Len(t, rpcResp.Result.Sessions, 5)

	// 5. Simula Estação 02 iniciando elaboração (processing)
	require.NoError(t, pool.TransitionState(queue.StationPMArchitect, acp.SessionStateProcessing, "conception"))

	// Snapshot reflete mascote working
	snap, err := srv.Aggregator.GetSnapshot()
	require.NoError(t, err)
	for _, st := range snap.Stations {
		if st.ID == queue.StationPMArchitect {
			assert.Equal(t, "working", st.MascotState)
		}
	}

	// 6. Conclui elaboração e bloqueia em approval_pending (Portão FLX-01)
	require.NoError(t, pool.TransitionState(queue.StationPMArchitect, acp.SessionStateApprovalPending, "spec completed"))

	snap, err = srv.Aggregator.GetSnapshot()
	require.NoError(t, err)
	for _, st := range snap.Stations {
		if st.ID == queue.StationPMArchitect {
			assert.Equal(t, "approval", st.MascotState)
		}
	}

	// 7. Homologação do Portão FLX-01 via chamada RPC gate/approve
	gateReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "gate/approve",
		"params": map[string]any{
			"gate":         "FLX-01",
			"featureId":    "feat-e2e-pool",
			"specPath":     specRelPath,
			"workspaceDir": wsDir,
		},
	}
	gateBody, err := json.Marshal(gateReq)
	require.NoError(t, err)

	gateHttpResp, err := http.Post(fmt.Sprintf("%s/api/acp", serverURL), "application/json", bytes.NewReader(gateBody))
	require.NoError(t, err)
	defer gateHttpResp.Body.Close()
	require.Equal(t, http.StatusOK, gateHttpResp.StatusCode)

	var gateResp struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Result  acp.GateApproveResult `json:"result"`
	}
	require.NoError(t, json.NewDecoder(gateHttpResp.Body).Decode(&gateResp))
	assert.True(t, gateResp.Result.Approved)
	assert.Equal(t, queue.StationPlanner, gateResp.Result.NextStation)

	// Handoff automático: Estação 02 volta a standby e Estação 03 transiciona para processing
	plannerSess, ok := pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateProcessing, plannerSess.State)

	// Snapshot reflete mascote do Planejador em thinking
	snap, err = srv.Aggregator.GetSnapshot()
	require.NoError(t, err)
	for _, st := range snap.Stations {
		if st.ID == queue.StationPlanner {
			assert.Equal(t, "thinking", st.MascotState)
		}
	}

	// 8. Anexação de cliente terminal via AttachBridge
	bridge := srv.AttachBridge()
	require.NotNil(t, bridge)

	stream, detach, err := bridge.Attach(queue.StationPlanner)
	require.NoError(t, err)
	defer detach()

	// Broadcast de logs em tempo real
	bridge.BroadcastLog(queue.StationPlanner, acp.SessionLogEntry{
		Type:    "info",
		Message: "Planejador decompondo tarefas determinísticas...",
	})

	select {
	case entry := <-stream:
		assert.Equal(t, "Planejador decompondo tarefas determinísticas...", entry.Message)
		assert.Equal(t, "info", entry.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout aguardando log do Planejador via streaming de terminal")
	}

	// 9. Desconexão graciosa do terminal
	detach()

	// Verifica se a sessão do Planejador continua íntegra e ativa no pool
	plannerSessAfter, ok := pool.GetSessionByStation(queue.StationPlanner)
	require.True(t, ok)
	assert.Equal(t, acp.SessionStateProcessing, plannerSessAfter.State)
}
