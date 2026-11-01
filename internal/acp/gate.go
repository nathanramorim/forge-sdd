package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
	FromStation string             `json:"fromStation"`
	ToStation   string             `json:"toStation"`
	HandoffItem *queue.StationItem `json:"handoffItem"`
	Checks      []gate.CheckResult `json:"checks"`
}

// HandleGateApprove executa a homologação do Portão FLX-01 via ACP.
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

	wsDir := strings.TrimSpace(p.WorkspaceDir)
	branch := strings.TrimSpace(p.Branch)

	if p.SessionID != "" && d != nil && d.sessionManager != nil {
		if sess, ok := d.sessionManager.GetSession(p.SessionID); ok {
			if wsDir == "" {
				wsDir = sess.WorkspaceDir
			}
			if branch == "" {
				branch = sess.Branch
			}
		}
	}

	if wsDir == "" {
		wsDir = "."
	}

	// 1. Validação canônica do Portão FLX-01 (reuso estrito de internal/gate - DUP-01, DUP-02)
	checks, passed := gate.ValidateFLX01WithBranch(wsDir, p.SpecPath, branch)
	if !passed {
		var failureEvidences []string
		for _, c := range checks {
			if !c.Passed {
				failureEvidences = append(failureEvidences, fmt.Sprintf("[%s/%s]: %s", c.ID, c.Rule, c.Evidence))
			}
		}
		errMsg := "Portão FLX-01 rejeitado: " + strings.Join(failureEvidences, "; ")
		return nil, NewRPCError(CodeServerError, errMsg, map[string]any{
			"gate":   "FLX-01",
			"checks": checks,
		})
	}

	// 2. Determinação de identificadores e estações
	fromStation := strings.TrimSpace(p.FromStation)
	if fromStation == "" {
		fromStation = queue.StationPMArchitect
	}
	toStation := strings.TrimSpace(p.ToStation)
	if toStation == "" {
		toStation = queue.StationPlanner
	}

	if !queue.IsValidStation(fromStation) {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid fromStation: %q", fromStation))
	}
	if !queue.IsValidStation(toStation) {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid toStation: %q", toStation))
	}

	itemID := strings.TrimSpace(p.FeatureID)
	if itemID == "" {
		itemID = strings.TrimSpace(p.TaskID)
	}
	if itemID == "" {
		base := filepath.Base(p.SpecPath)
		itemID = strings.TrimSuffix(base, filepath.Ext(base))
	}

	title := strings.TrimSpace(p.Title)
	if title == "" {
		title = fmt.Sprintf("Feature %s", itemID)
	}

	// 3. Transição atômica da Estação 02 (origem) para 'approved'
	fromSQ, err := queue.LoadStationQueue(wsDir, fromStation)
	if err == nil {
		found := false
		for _, it := range fromSQ.Items {
			if it.ID == itemID || (p.SpecPath != "" && it.ArtifactPath == p.SpecPath) {
				found = true
				break
			}
		}
		if !found {
			_, _, _ = queue.EnqueueStationItem(wsDir, queue.StationItem{
				ID:           itemID,
				Title:        title,
				Station:      fromStation,
				Operation:    "spec",
				Status:       queue.StatusApproved,
				ArtifactPath: p.SpecPath,
			})
		} else {
			_, _ = queue.UpdateStationItemStatus(wsDir, fromStation, itemID, queue.StatusApproved)
		}
	}

	// 4. Handoff determinístico para Estação de destino (queue.Handoff)
	handoffItem, _, err := queue.Handoff(
		wsDir,
		fromStation,
		toStation,
		itemID,
		title,
		p.SpecPath,
		p.InputContract,
		p.OutputContract,
		p.IdempotencyKey,
	)
	if err != nil {
		return nil, NewRPCError(CodeInternalError, fmt.Sprintf("handoff failed: %v", err))
	}

	// 5. Notificação de streaming via session/notification
	if d != nil {
		d.BroadcastSessionNotification(SessionNotification{
			SessionID: p.SessionID,
			Type:      "gate_approved",
			Station:   toStation,
			Message:   fmt.Sprintf("Portão FLX-01 aprovado para %s; handoff realizado para estação %s", itemID, toStation),
			Payload:   handoffItem,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}

	return GateApproveResult{
		FeatureID:   itemID,
		SpecPath:    p.SpecPath,
		Gate:        "FLX-01",
		Status:      queue.StatusApproved,
		FromStation: fromStation,
		ToStation:   toStation,
		HandoffItem: handoffItem,
		Checks:      checks,
	}, nil
}
