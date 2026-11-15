package acp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/queue"
)

// PendingGateInfo registra informações de uma demanda aguardando chancela humana no Portão FLX-01.
type PendingGateInfo struct {
	TaskID       string    `json:"taskId"`
	Station      string    `json:"station"`
	ArtifactPath string    `json:"artifactPath"`
	RequestedAt  time.Time `json:"requestedAt"`
}

// HandoffOrchestrator coordena a transição contínua e autônoma entre as 5 estações canônicas do SDD,
// garantindo o bloqueio obrigatório no Portão FLX-01 e prevenindo tentativas de bypass (ARQ-01, FLX-01).
type HandoffOrchestrator struct {
	mu            sync.RWMutex
	pool          *SessionPool
	dir           string
	dispatcher    *RPCDispatcher
	pendingGates  map[string]*PendingGateInfo // chave: taskID
	approvedGates map[string]bool             // chave: taskID
}

// NewHandoffOrchestrator instancia um novo orquestrador de handoff conectado ao SessionPool e RPCDispatcher.
func NewHandoffOrchestrator(pool *SessionPool, dir string, dispatcher *RPCDispatcher) *HandoffOrchestrator {
	if dir == "" {
		dir = "."
	}
	orch := &HandoffOrchestrator{
		pool:          pool,
		dir:           dir,
		dispatcher:    dispatcher,
		pendingGates:  make(map[string]*PendingGateInfo),
		approvedGates: make(map[string]bool),
	}
	if pool != nil {
		pool.SetOrchestrator(orch)
	}
	return orch
}

// Dir retorna o diretório base configurado para operações e arquivos de fila.
func (o *HandoffOrchestrator) Dir() string {
	return o.workspaceDir()
}

// SetDir atualiza o diretório de workspace do orquestrador.
func (o *HandoffOrchestrator) SetDir(dir string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.dir = dir
}

// Pool retorna o SessionPool associado ao orquestrador.
func (o *HandoffOrchestrator) Pool() *SessionPool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.pool
}

// SetPool atualiza o SessionPool associado.
func (o *HandoffOrchestrator) SetPool(pool *SessionPool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.pool = pool
}

// Dispatcher retorna o RPCDispatcher associado.
func (o *HandoffOrchestrator) Dispatcher() *RPCDispatcher {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.dispatcher
}

// SetDispatcher atualiza o RPCDispatcher associado.
func (o *HandoffOrchestrator) SetDispatcher(dispatcher *RPCDispatcher) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.dispatcher = dispatcher
}

func (o *HandoffOrchestrator) workspaceDir() string {
	o.mu.RLock()
	d := o.dir
	p := o.pool
	o.mu.RUnlock()

	if d != "" && d != "." {
		return d
	}
	if p != nil {
		if wd := p.WorkspaceDir(); wd != "" {
			return wd
		}
	}
	if d != "" {
		return d
	}
	return "."
}

// NextStation retorna a próxima estação na esteira contínua canônica do SDD.
func NextStation(currentStation string) (string, bool) {
	curr := NormalizeStation(currentStation)
	switch curr {
	case queue.StationLider:
		return queue.StationPMArchitect, true
	case queue.StationPMArchitect:
		return queue.StationPlanner, true
	case queue.StationPlanner:
		return queue.StationBuilder, true
	case queue.StationBuilder:
		return queue.StationGovernance, true
	default:
		return "", false
	}
}

// NormalizeStation converte identificadores de sessão ("sess-02-pm") ou nomes de estação para o canônico.
func NormalizeStation(station string) string {
	switch strings.TrimSpace(station) {
	case "sess-01-lider", queue.StationLider:
		return queue.StationLider
	case "sess-02-pm", queue.StationPMArchitect:
		return queue.StationPMArchitect
	case "sess-03-planner", queue.StationPlanner:
		return queue.StationPlanner
	case "sess-04-builder", queue.StationBuilder:
		return queue.StationBuilder
	case "sess-05-governance", queue.StationGovernance:
		return queue.StationGovernance
	default:
		return strings.TrimSpace(station)
	}
}

// HandleTaskCompleted processa a sinalização de conclusão de tarefa por uma estação de trabalho.
// Se station == queue.StationPMArchitect, transiciona sess-02-pm para approval_pending com bloqueio
// obrigatório do Portão FLX-01, emite a notificação gate/approval_required e impede qualquer despacho
// para sess-03-planner até aprovação explícita.
func (o *HandoffOrchestrator) HandleTaskCompleted(ctx context.Context, station string, taskID string, artifactPath string) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}

	normStation := NormalizeStation(station)
	if !queue.IsValidStation(normStation) {
		return fmt.Errorf("estação inválida: %q", station)
	}

	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		base := filepath.Base(artifactPath)
		taskID = strings.TrimSuffix(base, filepath.Ext(base))
	}

	o.mu.Lock()
	o.pendingGates[taskID] = &PendingGateInfo{
		TaskID:       taskID,
		Station:      normStation,
		ArtifactPath: artifactPath,
		RequestedAt:  time.Now().UTC(),
	}
	delete(o.approvedGates, taskID)
	pool := o.pool
	dispatcher := o.dispatcher
	o.mu.Unlock()

	if normStation == queue.StationPMArchitect {
		// 1. Transiciona sess-02-pm para SessionStateApprovalPending (FLX-01)
		if pool != nil {
			if sess, ok := pool.GetSessionByStation(queue.StationPMArchitect); ok {
				_, _ = pool.UpdateSessionState(sess.SessionID, SessionStateApprovalPending, taskID)
			}
		}

		// 2. Emite notificação gate/approval_required
		payload := map[string]any{
			"station":   queue.StationPMArchitect,
			"itemId":    taskID,
			"taskId":    taskID,
			"specPath":  artifactPath,
			"gate":      "FLX-01",
			"rules":     []string{"P-02", "P-03"},
			"message":   fmt.Sprintf("Portão FLX-01 pendente de aprovação humana para a task %s", taskID),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		if dispatcher != nil {
			dispatcher.BroadcastNotification(NewNotification("gate/approval_required", payload))
			dispatcher.BroadcastSessionNotification(SessionNotification{
				Type:      "gate/approval_required",
				Station:   queue.StationPMArchitect,
				Message:   fmt.Sprintf("Portão FLX-01 pendente de aprovação humana para a task %s", taskID),
				Payload:   payload,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			})
		}
	} else {
		// Para outras estações, retorna a sessão correspondente para standby
		if pool != nil {
			if sess, ok := pool.GetSessionByStation(normStation); ok {
				_, _ = pool.UpdateSessionState(sess.SessionID, SessionStateStandby, "")
			}
		}
	}

	return nil
}

// OnSpecCreated é um alias conveniente para HandleTaskCompleted da Estação 02 (PM & Arquiteto).
func (o *HandoffOrchestrator) OnSpecCreated(station string, itemID string, specPath string) error {
	return o.HandleTaskCompleted(context.Background(), station, itemID, specPath)
}

// ApproveGate valida deterministicamente os critérios P-02 e P-03 do Portão FLX-01 via internal/gate,
// atualiza a fila para approved, devolve sess-02-pm para standby, ativa sess-03-planner em processing
// e emite as notificações de streaming correspondentes.
func (o *HandoffOrchestrator) ApproveGate(ctx context.Context, station string, taskID string, specPath string) (*GateApproveResult, error) {
	res, rpcErr := o.ApproveGateWithParams(ctx, GateApproveParams{
		FromStation: station,
		TaskID:      taskID,
		FeatureID:   taskID,
		SpecPath:    specPath,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return res, nil
}

// OnGateApproved é um alias de compatibilidade para homologação direta de portão.
func (o *HandoffOrchestrator) OnGateApproved(station string, itemID string, specPath string) (*GateApproveResult, error) {
	return o.ApproveGate(context.Background(), station, itemID, specPath)
}

// ApproveGateWithParams executa a aprovação do Portão FLX-01 com suporte completo aos parâmetros de entrada.
func (o *HandoffOrchestrator) ApproveGateWithParams(ctx context.Context, p GateApproveParams) (*GateApproveResult, *RPCError) {
	if ctx != nil && ctx.Err() != nil {
		return nil, NewRPCError(CodeInternalError, ctx.Err().Error())
	}

	specPath := strings.TrimSpace(p.SpecPath)
	if specPath == "" {
		return nil, NewRPCError(CodeInvalidParams, "specPath is required")
	}

	wsDir := strings.TrimSpace(p.WorkspaceDir)
	if wsDir == "" {
		wsDir = o.workspaceDir()
	}

	branch := strings.TrimSpace(p.Branch)
	o.mu.RLock()
	d := o.dispatcher
	pool := o.pool
	o.mu.RUnlock()

	if p.SessionID != "" && d != nil && d.sessionManager != nil {
		if sess, ok := d.sessionManager.GetSession(p.SessionID); ok {
			if wsDir == "." && sess.WorkspaceDir != "" {
				wsDir = sess.WorkspaceDir
			}
			if branch == "" {
				branch = sess.Branch
			}
		}
	}

	// 1. Validação canônica do Portão FLX-01 (reuso estrito de internal/gate - DUP-01, DUP-02)
	checks, passed := gate.ValidateFLX01WithBranch(wsDir, specPath, branch)
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
	fromStation := NormalizeStation(p.FromStation)
	if fromStation == "" {
		fromStation = queue.StationPMArchitect
	}
	toStation := NormalizeStation(p.ToStation)
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
		base := filepath.Base(specPath)
		itemID = strings.TrimSuffix(base, filepath.Ext(base))
	}

	title := strings.TrimSpace(p.Title)
	if title == "" {
		title = fmt.Sprintf("Feature %s", itemID)
	}

	// 3. Atualização de estado interno do orquestrador
	o.mu.Lock()
	o.approvedGates[itemID] = true
	delete(o.pendingGates, itemID)
	o.mu.Unlock()

	// 4. Transição atômica da Estação de origem para 'approved'
	fromSQ, err := queue.LoadStationQueue(wsDir, fromStation)
	if err == nil {
		found := false
		for _, it := range fromSQ.Items {
			if it.ID == itemID || (specPath != "" && it.ArtifactPath == specPath) {
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
				ArtifactPath: specPath,
			})
		} else {
			_, _ = queue.UpdateStationItemStatus(wsDir, fromStation, itemID, queue.StatusApproved)
		}
	}

	// 5. Handoff determinístico para Estação de destino (queue.Handoff)
	handoffItem, _, err := queue.Handoff(
		wsDir,
		fromStation,
		toStation,
		itemID,
		title,
		specPath,
		p.InputContract,
		p.OutputContract,
		p.IdempotencyKey,
	)
	if err != nil {
		return nil, NewRPCError(CodeInternalError, fmt.Sprintf("handoff failed: %v", err))
	}

	// 6. Notificação primária de homologação de portão
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

	// 7. Transições automáticas de sessão no SessionPool:
	// - sess-02-pm retorna para Standby
	// - sess-03-planner transiciona automaticamente para Processing vinculada a taskID
	if pool != nil {
		if pmSess, ok := pool.GetSessionByStation(queue.StationPMArchitect); ok {
			_, _ = pool.UpdateSessionState(pmSess.SessionID, SessionStateStandby, "")
		}
		if planSess, ok := pool.GetSessionByStation(queue.StationPlanner); ok {
			_, _ = pool.UpdateSessionState(planSess.SessionID, SessionStateProcessing, itemID)
		}
	}

	return &GateApproveResult{
		FeatureID:   itemID,
		SpecPath:    specPath,
		Gate:        "FLX-01",
		Status:      queue.StatusApproved,
		Approved:    true,
		FromStation: fromStation,
		ToStation:   toStation,
		NextStation: toStation,
		HandoffItem: handoffItem,
		Checks:      checks,
	}, nil
}

// DispatchTask despacha uma tarefa para a sessão correspondente no SessionPool, aplicando estritamente
// o bloqueio constitucional do Portão FLX-01 e prevenindo tentativas de bypass (erro canônico -32001).
func (o *HandoffOrchestrator) DispatchTask(ctx context.Context, targetStation string, taskID string, artifactPath string) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}

	target := NormalizeStation(targetStation)
	if !queue.IsValidStation(target) {
		return fmt.Errorf("estação de destino inválida: %q", targetStation)
	}

	taskID = strings.TrimSpace(taskID)

	// Prevenção de bypass do Portão FLX-01:
	// A ativação do Planejador (Estação 03) é estritamente proibida sem aprovação prévia do portão.
	if target == queue.StationPlanner {
		o.mu.RLock()
		isApproved := o.approvedGates[taskID]
		isPending := o.pendingGates[taskID] != nil
		pool := o.pool
		o.mu.RUnlock()

		var pmPendingApproval bool
		if pool != nil {
			if pmSess, ok := pool.GetSessionByStation(queue.StationPMArchitect); ok {
				if pmSess.State == SessionStateApprovalPending {
					pmPendingApproval = true
				}
			}
		}

		if !isApproved || isPending || pmPendingApproval {
			return NewRPCError(CodeServerError, fmt.Sprintf("tentativa de bypass do Portão FLX-01 detectada: ativação da estação %s bloqueada para task %s", target, taskID))
		}
	}

	o.mu.RLock()
	pool := o.pool
	o.mu.RUnlock()

	if pool != nil {
		if sess, ok := pool.GetSessionByStation(target); ok {
			_, err := pool.UpdateSessionState(sess.SessionID, SessionStateProcessing, taskID)
			return err
		}
	}

	return nil
}

// ActivateStation ativa uma estação de trabalho para processar uma taskID, impedindo bypass de portão.
func (o *HandoffOrchestrator) ActivateStation(ctx context.Context, station string, taskID string) error {
	return o.DispatchTask(ctx, station, taskID, "")
}

// AdvanceStation avança deterministicamente da estação atual para a próxima estação do pipeline.
func (o *HandoffOrchestrator) AdvanceStation(ctx context.Context, currentStation string, taskID string) error {
	curr := NormalizeStation(currentStation)
	next, ok := NextStation(curr)
	if !ok {
		return fmt.Errorf("estação final ou desconhecida: %q", currentStation)
	}
	return o.DispatchTask(ctx, next, taskID, "")
}

// TriggerHandoff dispara o handoff manual entre estações, validando as regras do portão.
func (o *HandoffOrchestrator) TriggerHandoff(fromStation, toStation, taskID string, payload any) error {
	_ = fromStation
	_ = payload
	return o.DispatchTask(context.Background(), toStation, taskID, "")
}

// IsGateApproved retorna se um item/tarefa foi aprovado no Portão FLX-01.
func (o *HandoffOrchestrator) IsGateApproved(taskID string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.approvedGates[taskID]
}

// IsGatePending retorna se um item/tarefa está aguardando chancela no Portão FLX-01.
func (o *HandoffOrchestrator) IsGatePending(taskID string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.pendingGates[taskID] != nil
}

// PendingGates retorna uma cópia defensiva de todas as demandas aguardando aprovação.
func (o *HandoffOrchestrator) PendingGates() []*PendingGateInfo {
	o.mu.RLock()
	defer o.mu.RUnlock()
	list := make([]*PendingGateInfo, 0, len(o.pendingGates))
	for _, g := range o.pendingGates {
		cp := *g
		list = append(list, &cp)
	}
	return list
}
