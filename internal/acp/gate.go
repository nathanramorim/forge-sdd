package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/queue"
)

// GateApproveParams define os parâmetros para validação e homologação do Portão FLX-01 via gate/approve.
type GateApproveParams struct {
	SessionID      string `json:"sessionId,omitempty"`
	WorkspaceDir   string `json:"workspaceDir,omitempty"`
	FeatureID      string `json:"featureId,omitempty"`
	TaskID         string `json:"taskId,omitempty"`
	Title          string `json:"title,omitempty"`
	SpecPath       string `json:"specPath"`
	Branch         string `json:"branch,omitempty"`
	FromStation    string `json:"fromStation,omitempty"` // Padrão: queue.StationPMArchitect (Estação 02)
	ToStation      string `json:"toStation,omitempty"`   // Padrão: queue.StationPlanner (Estação 03)
	InputContract  string `json:"inputContract,omitempty"`
	OutputContract string `json:"outputContract,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// GateApproveResult define o resultado retornado após a homologação e o handoff determinístico do Portão FLX-01.
type GateApproveResult struct {
	FeatureID   string             `json:"featureId"`
	SpecPath    string             `json:"specPath"`
	Gate        string             `json:"gate"`   // "FLX-01"
	Status      string             `json:"status"` // "approved"
	Approved    bool               `json:"approved,omitempty"`
	FromStation string             `json:"fromStation"`
	ToStation   string             `json:"toStation"`
	NextStation string             `json:"nextStation,omitempty"`
	HandoffItem *queue.StationItem `json:"handoffItem"`
	Checks      []gate.CheckResult `json:"checks"`
}

// HandleGateApprove executa a homologação do Portão FLX-01 via ACP delegando ao HandoffOrchestrator.
// Reutiliza o validador canônico em internal/gate (P-01, P-02, P-03), rejeita com erro -32001
// caso critérios sejam violados, e em caso de sucesso realiza escrita atômica transicionando
// a fila da Estação 02 para approved e acionando o handoff determinístico (queue.Handoff).
func HandleGateApprove(ctx context.Context, d *RPCDispatcher, params json.RawMessage) (any, *RPCError) {
	if len(params) == 0 {
		return nil, NewRPCError(CodeInvalidParams, "params required for gate/approve")
	}

	var p GateApproveParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid gate/approve params: %v", err))
	}

	p.SpecPath = strings.TrimSpace(p.SpecPath)
	if p.SpecPath == "" {
		return nil, NewRPCError(CodeInvalidParams, "specPath is required")
	}

	var orch *HandoffOrchestrator
	if d != nil {
		orch = d.Orchestrator()
	}
	if orch == nil {
		orch = NewHandoffOrchestrator(nil, ".", d)
	}

	res, rpcErr := orch.ApproveGateWithParams(ctx, p)
	if rpcErr != nil {
		return nil, rpcErr
	}

	return *res, nil
}
