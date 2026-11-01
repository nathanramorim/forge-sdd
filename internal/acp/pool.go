package acp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
)

// CanonicalSessionDef associa o ID canônico da sessão à sua estação de trabalho.
type CanonicalSessionDef struct {
	ID      string
	Station string
}

// CanonicalPoolSessions define as 5 sessões canônicas persistentes pré-aquecidas no boot.
var CanonicalPoolSessions = []CanonicalSessionDef{
	{ID: "sess-01-lider", Station: queue.StationLider},
	{ID: "sess-02-pm", Station: queue.StationPMArchitect},
	{ID: "sess-03-planner", Station: queue.StationPlanner},
	{ID: "sess-04-builder", Station: queue.StationBuilder},
	{ID: "sess-05-governance", Station: queue.StationGovernance},
}

// SessionPool gerencia em memória as 5 sessões ACP persistentes das estações canônicas,
// provendo ciclo de vida thread-safe, isolamento de workspace e pre-warming (ARQ-03, SEG-01, SEG-02).
type SessionPool struct {
	mu              sync.RWMutex
	sm              *SessionManager
	dispatcher      *RPCDispatcher
	sessions        map[string]*PooledSession // Chave: sessionId
	byStation       map[string]*PooledSession // Chave: station
	stalledNotified map[string]bool           // Chave: sessionId
	workspaceDir    string
	warmed          bool
	heartbeat       *HeartbeatMonitor
	orchestrator    *HandoffOrchestrator
}

// NewSessionPool instancia um SessionPool configurado com SessionManager e RPCDispatcher opcionais.
func NewSessionPool(sm *SessionManager, dispatcher *RPCDispatcher) *SessionPool {
	if sm == nil {
		sm = NewSessionManager()
	}
	p := &SessionPool{
		sm:              sm,
		dispatcher:      dispatcher,
		sessions:        make(map[string]*PooledSession, len(CanonicalPoolSessions)),
		byStation:       make(map[string]*PooledSession, len(CanonicalPoolSessions)),
		stalledNotified: make(map[string]bool, len(CanonicalPoolSessions)),
	}
	p.heartbeat = NewHeartbeatMonitor(p, dispatcher)
	return p
}

// SetSessionManager atualiza o SessionManager associado ao pool.
func (p *SessionPool) SetSessionManager(sm *SessionManager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sm = sm
}

// HeartbeatMonitor retorna o monitor de batimento cardíaco vinculado ao pool.
func (p *SessionPool) HeartbeatMonitor() *HeartbeatMonitor {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.heartbeat
}

// SetHeartbeatMonitor atribui um monitor de batimento cardíaco customizado ao pool.
func (p *SessionPool) SetHeartbeatMonitor(hm *HeartbeatMonitor) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.heartbeat = hm
}

// Orchestrator retorna o HandoffOrchestrator vinculado ao pool.
func (p *SessionPool) Orchestrator() *HandoffOrchestrator {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.orchestrator
}

// SetOrchestrator atribui o HandoffOrchestrator ao pool.
func (p *SessionPool) SetOrchestrator(o *HandoffOrchestrator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.orchestrator = o
}

// StartHeartbeat inicia o monitor de batimentos em background.
func (p *SessionPool) StartHeartbeat(ctx context.Context) {
	p.mu.RLock()
	hm := p.heartbeat
	p.mu.RUnlock()

	if hm != nil {
		hm.Start(ctx)
	}
}

// StopHeartbeat para o monitor de batimentos se estiver ativo.
func (p *SessionPool) StopHeartbeat() {
	p.mu.RLock()
	hm := p.heartbeat
	p.mu.RUnlock()

	if hm != nil {
		hm.Stop()
	}
}

// Close encerra o SessionPool e libera recursos em background (COD-01).
func (p *SessionPool) Close() error {
	p.StopHeartbeat()
	return nil
}

// IsWarmed indica se o pool já foi pré-aquecido com sucesso.
func (p *SessionPool) IsWarmed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.warmed
}

// WorkspaceDir retorna o diretório de workspace ativo associado ao pool.
func (p *SessionPool) WorkspaceDir() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.workspaceDir
}

// Warmup inicializa e pré-aquece as 5 sessões canônicas ancoradas no workspace validado (SEG-01, SEG-02).
func (p *SessionPool) Warmup(ctx context.Context, wsDir string) (*PoolWarmupResult, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	p.mu.RLock()
	sm := p.sm
	p.mu.RUnlock()

	if sm == nil {
		return nil, errors.New("session manager indisponível no pool")
	}

	validatedWs, err := sm.ValidateWorkspace(wsDir)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	now := time.Now().UTC()

	sessionsList := make([]*PooledSession, 0, len(CanonicalPoolSessions))
	for _, def := range CanonicalPoolSessions {
		existing, ok := p.sessions[def.ID]
		var sess *PooledSession
		if ok {
			sess = existing
			sess.WorkspaceDir = validatedWs
			sess.State = SessionStateStandby
			sess.ActiveItemID = ""
			sess.LastSeenAt = now
			sess.UpdatedAt = now
			sess.ProcessingStartedAt = time.Time{}
			sess.LastTelemetryAt = now
		} else {
			sess = &PooledSession{
				SessionID:           def.ID,
				Station:             def.Station,
				WorkspaceDir:        validatedWs,
				State:               SessionStateStandby,
				ActiveItemID:        "",
				LastSeenAt:          now,
				AttachedClients:     0,
				CreatedAt:           now,
				UpdatedAt:           now,
				Metadata:            make(map[string]string),
				ProcessingStartedAt: time.Time{},
				LastTelemetryAt:     now,
			}
			p.sessions[def.ID] = sess
			p.byStation[def.Station] = sess
		}
		p.stalledNotified[def.ID] = false

		// Registra ou sincroniza no SessionManager
		sm.RegisterSession(&Session{
			ID:           sess.SessionID,
			WorkspaceDir: sess.WorkspaceDir,
			Station:      sess.Station,
			Status:       "active",
			CreatedAt:    sess.CreatedAt,
			UpdatedAt:    now,
			Metadata:     make(map[string]string),
		})

		sessionsList = append(sessionsList, sess.Clone())
	}

	p.workspaceDir = validatedWs
	p.warmed = true
	dispatcher := p.dispatcher
	p.mu.Unlock()

	result := &PoolWarmupResult{
		Status:        "warmed",
		SessionsCount: len(sessionsList),
		Sessions:      sessionsList,
	}

	// Emite notificação pool/warmed
	if dispatcher != nil {
		dispatcher.BroadcastNotification(NewNotification("pool/warmed", result))
		dispatcher.BroadcastSessionNotification(SessionNotification{
			Type:      "pool_warmed",
			Message:   fmt.Sprintf("SessionPool pré-aquecido com %d sessões no workspace %s", len(sessionsList), validatedWs),
			Payload:   result,
			Timestamp: now.Format(time.RFC3339),
		})
	}

	return result, nil
}

// GetSession busca uma sessão pelo seu ID canônico, retornando cópia defensiva thread-safe.
func (p *SessionPool) GetSession(sessionID string) (*PooledSession, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sess, ok := p.sessions[sessionID]
	if !ok {
		return nil, false
	}
	return sess.Clone(), true
}

// GetSessionByStation busca uma sessão pelo identificador da estação, retornando cópia defensiva thread-safe.
func (p *SessionPool) GetSessionByStation(station string) (*PooledSession, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sess, ok := p.byStation[station]
	if !ok {
		return nil, false
	}
	return sess.Clone(), true
}

// UpdateSessionState altera o estado operacional e item ativo de uma sessão de forma atômica.
func (p *SessionPool) UpdateSessionState(sessionID string, state SessionState, activeItemID string) (*PooledSession, error) {
	if !IsValidSessionState(state) {
		return nil, fmt.Errorf("estado de sessão inválido: %q", state)
	}

	p.mu.Lock()
	sess, ok := p.sessions[sessionID]
	if !ok {
		p.mu.Unlock()
		return nil, fmt.Errorf("sessão %q não encontrada no pool", sessionID)
	}

	now := time.Now().UTC()
	sess.State = state
	sess.ActiveItemID = activeItemID
	sess.UpdatedAt = now
	sess.LastSeenAt = now
	p.stalledNotified[sessionID] = false

	if state == SessionStateProcessing {
		sess.ProcessingStartedAt = now
		sess.LastTelemetryAt = now
	} else {
		sess.ProcessingStartedAt = time.Time{}
	}

	cloned := sess.Clone()
	dispatcher := p.dispatcher
	p.mu.Unlock()

	if dispatcher != nil {
		dispatcher.BroadcastSessionNotification(SessionNotification{
			SessionID: cloned.SessionID,
			Type:      "station_state",
			Station:   cloned.Station,
			Message:   fmt.Sprintf("Sessão %s mudou para estado %s", cloned.SessionID, state),
			Payload:   cloned,
			Timestamp: now.Format(time.RFC3339),
		})
	}

	return cloned, nil
}

// TransitionState altera o estado de uma estação canônica registrando o motivo opcional.
func (p *SessionPool) TransitionState(station string, newState SessionState, reason string) error {
	if !IsValidSessionState(newState) {
		return fmt.Errorf("estado de sessão inválido: %q", newState)
	}
	if !queue.IsValidStation(station) {
		return fmt.Errorf("estação inválida: %q", station)
	}

	p.mu.Lock()
	sess, ok := p.byStation[station]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("nenhuma sessão associada à estação %q", station)
	}

	now := time.Now().UTC()
	sess.State = newState
	sess.UpdatedAt = now
	sess.LastSeenAt = now
	if reason != "" {
		if sess.Metadata == nil {
			sess.Metadata = make(map[string]string)
		}
		sess.Metadata["transitionReason"] = reason
	}
	p.stalledNotified[sess.SessionID] = false

	if newState == SessionStateProcessing {
		sess.ProcessingStartedAt = now
		sess.LastTelemetryAt = now
	} else {
		sess.ProcessingStartedAt = time.Time{}
	}

	cloned := sess.Clone()
	dispatcher := p.dispatcher
	p.mu.Unlock()

	if dispatcher != nil {
		dispatcher.BroadcastSessionNotification(SessionNotification{
			SessionID: cloned.SessionID,
			Type:      "station_state",
			Station:   cloned.Station,
			Message:   fmt.Sprintf("Estação %s transitou para %s (%s)", station, newState, reason),
			Payload:   cloned,
			Timestamp: now.Format(time.RFC3339),
		})
	}

	return nil
}

// TouchSession atualiza o timestamp de atividade e telemetria da sessão.
func (p *SessionPool) TouchSession(sessionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sess, ok := p.sessions[sessionID]
	if !ok {
		return fmt.Errorf("sessão %q não encontrada no pool", sessionID)
	}

	now := time.Now().UTC()
	sess.LastSeenAt = now
	sess.LastTelemetryAt = now
	p.stalledNotified[sessionID] = false
	return nil
}

// RecordTelemetry é um alias semântico para atualizar a telemetria recente de uma sessão.
func (p *SessionPool) RecordTelemetry(sessionID string) error {
	return p.TouchSession(sessionID)
}

// ListSessions retorna a lista ordenada com cópias defensivas das 5 sessões canônicas.
func (p *SessionPool) ListSessions() []*PooledSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	res := make([]*PooledSession, 0, len(CanonicalPoolSessions))
	for _, def := range CanonicalPoolSessions {
		if s, ok := p.sessions[def.ID]; ok {
			res = append(res, s.Clone())
		}
	}
	return res
}

// Status inspeciona a integridade e telemetria geral do SessionPool.
func (p *SessionPool) Status() *PoolStatusResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sessions := make([]*PooledSession, 0, len(CanonicalPoolSessions))
	healthy := p.warmed && len(p.sessions) == len(CanonicalPoolSessions)

	for _, def := range CanonicalPoolSessions {
		if s, ok := p.sessions[def.ID]; ok {
			if s.State == SessionStateError || s.State == SessionStateOffline {
				healthy = false
			}
			sessions = append(sessions, s.Clone())
		} else {
			healthy = false
		}
	}

	return &PoolStatusResult{
		Healthy:  healthy,
		Sessions: sessions,
	}
}

// updateHeartbeat é invocado pelo HeartbeatMonitor a cada ciclo periódico para atualizar LastSeenAt
// e detectar sessões em stall.
func (p *SessionPool) updateHeartbeat(now time.Time, stallTimeout time.Duration) []*PooledSession {
	p.mu.Lock()
	defer p.mu.Unlock()

	var stalled []*PooledSession
	for id, s := range p.sessions {
		s.LastSeenAt = now

		if s.State == SessionStateProcessing && stallTimeout > 0 {
			ref := s.LastTelemetryAt
			if ref.IsZero() {
				ref = s.ProcessingStartedAt
			}
			if !ref.IsZero() && now.Sub(ref) > stallTimeout {
				if !p.stalledNotified[id] {
					p.stalledNotified[id] = true
					stalled = append(stalled, s.Clone())
				}
			}
		}
	}
	return stalled
}
