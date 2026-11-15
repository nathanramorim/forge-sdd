package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
)

// Session representa uma sessão de trabalho ativa no protocolo ACP ancorada em um workspace.
type Session struct {
	ID           string            `json:"sessionId"`
	WorkspaceDir string            `json:"workspaceDir"`
	Feature      string            `json:"feature,omitempty"`
	Branch       string            `json:"branch,omitempty"`
	Station      string            `json:"station"`
	Status       string            `json:"status"` // "active", "loaded", "cancelled", etc.
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SessionManager gerencia o ciclo de vida e a concorrência de sessões ACP com isolamento de workspace (SEG-01, SEG-02).
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	allowedRoot string
}

// NewSessionManager instancia um SessionManager thread-safe com limite de raiz de segurança opcional.
func NewSessionManager(allowedRoot ...string) *SessionManager {
	var root string
	if len(allowedRoot) > 0 && strings.TrimSpace(allowedRoot[0]) != "" {
		if abs, err := filepath.Abs(allowedRoot[0]); err == nil {
			root = abs
		} else {
			root = filepath.Clean(allowedRoot[0])
		}
	}
	return &SessionManager{
		sessions:    make(map[string]*Session),
		allowedRoot: root,
	}
}

// SetAllowedRoot configura o diretório raiz permitido para workspaces (SEG-02).
func (sm *SessionManager) SetAllowedRoot(root string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if root != "" {
		if abs, err := filepath.Abs(root); err == nil {
			sm.allowedRoot = abs
			return
		}
	}
	sm.allowedRoot = root
}

// AllowedRoot retorna a raiz permitida atualmente configurada.
func (sm *SessionManager) AllowedRoot() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.allowedRoot
}

// ValidateWorkspace valida se o diretório existe e bloqueia tentativas de path traversal (SEG-01, SEG-02).
func (sm *SessionManager) ValidateWorkspace(rawPath string) (string, error) {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		sm.mu.RLock()
		root := sm.allowedRoot
		sm.mu.RUnlock()

		if root != "" {
			return root, nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("falha ao resolver diretório atual: %w", err)
		}
		return filepath.Clean(cwd), nil
	}

	// 1. Rejeição de caracteres nulos ou controle
	if strings.ContainsRune(trimmed, 0) {
		return "", errors.New("caminho do workspace contém caracteres inválidos")
	}

	cleanPath := filepath.Clean(trimmed)

	sm.mu.RLock()
	root := sm.allowedRoot
	sm.mu.RUnlock()

	var absPath string
	if filepath.IsAbs(cleanPath) {
		absPath = cleanPath
	} else {
		base := root
		if base == "" {
			var err error
			base, err = os.Getwd()
			if err != nil {
				return "", fmt.Errorf("falha ao resolver diretório base: %w", err)
			}
		}
		absPath = filepath.Clean(filepath.Join(base, cleanPath))
	}

	// 2. Prevenção de path traversal fora dos limites seguros
	if root != "" {
		absRoot, err := filepath.Abs(root)
		if err == nil {
			rel, err := filepath.Rel(absRoot, absPath)
			if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
				return "", fmt.Errorf("tentativa de path traversal fora dos limites permitidos: %s", rawPath)
			}
		}
	} else {
		if !filepath.IsAbs(cleanPath) && (cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || strings.Contains(cleanPath, string(filepath.Separator)+"..")) {
			cwd, _ := os.Getwd()
			rel, err := filepath.Rel(cwd, absPath)
			if err != nil || strings.HasPrefix(rel, "..") {
				return "", fmt.Errorf("tentativa de path traversal fora do diretório de execução: %s", rawPath)
			}
		}
	}

	// 3. Verificação de existência no sistema de arquivos
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("diretório de workspace não existe: %s", absPath)
		}
		return "", fmt.Errorf("erro ao verificar diretório de workspace: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("caminho de workspace não é um diretório: %s", absPath)
	}

	return absPath, nil
}

// CreateSession instancia e armazena uma nova sessão com workspace validado.
func (sm *SessionManager) CreateSession(params SessionNewParams) (*Session, error) {
	wsDir, err := sm.ValidateWorkspace(params.WorkspaceDir)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	session := &Session{
		ID:           sessionID,
		WorkspaceDir: wsDir,
		Feature:      params.Feature,
		Branch:       params.Branch,
		Station:      queue.StationPMArchitect, // Estação padrão inicial (02 Concepção)
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]string),
	}

	sm.sessions[sessionID] = session
	cp := *session
	return &cp, nil
}

// RegisterSession registra ou substitui uma sessão gerenciada (thread-safe).
func (sm *SessionManager) RegisterSession(sess *Session) {
	if sm == nil || sess == nil {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sess.ID] = sess
}

// GetSession busca uma sessão ativa pelo identificador (thread-safe).
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sess, ok := sm.sessions[id]
	if !ok {
		return nil, false
	}
	cp := *sess
	if sess.Metadata != nil {
		cp.Metadata = make(map[string]string, len(sess.Metadata))
		for k, v := range sess.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp, true
}

// CancelSession encerra uma sessão ativa registrando a justificativa.
func (sm *SessionManager) CancelSession(id string, reason string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, ok := sm.sessions[id]
	if !ok {
		return nil, fmt.Errorf("sessão %q não encontrada", id)
	}

	sess.Status = "cancelled"
	sess.UpdatedAt = time.Now().UTC()
	if reason != "" {
		if sess.Metadata == nil {
			sess.Metadata = make(map[string]string)
		}
		sess.Metadata["cancelReason"] = reason
	}

	cp := *sess
	return &cp, nil
}

// UpdateSessionStation atualiza a estação ativa vinculada à sessão.
func (sm *SessionManager) UpdateSessionStation(id string, station string) (*Session, error) {
	if !queue.IsValidStation(station) {
		return nil, fmt.Errorf("estação inválida: %q", station)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, ok := sm.sessions[id]
	if !ok {
		return nil, fmt.Errorf("sessão %q não encontrada", id)
	}

	sess.Station = station
	sess.UpdatedAt = time.Now().UTC()
	cp := *sess
	return &cp, nil
}

// ListSessions retorna a lista de todas as sessões registradas.
func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	res := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		cp := *s
		res = append(res, &cp)
	}
	return res
}

// SessionPromptParams define os parâmetros para envio de prompt ou demanda via ACP.
type SessionPromptParams struct {
	SessionID string            `json:"sessionId"`
	Prompt    string            `json:"prompt"`
	Station   string            `json:"station,omitempty"`
	Operation string            `json:"operation,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SessionPromptResult define o resultado retornado pelo comando session/prompt.
type SessionPromptResult struct {
	SessionID     string             `json:"sessionId"`
	Station       string             `json:"station"`
	Status        string             `json:"status"`
	Item          *queue.StationItem `json:"item"`
	IdempotentHit bool               `json:"idempotentHit"`
}

// SessionNotification representa o payload emitido no canal unidirecional de streaming "session/notification".
type SessionNotification struct {
	SessionID string `json:"sessionId,omitempty"`
	Type      string `json:"type"` // "station_state", "mascot_state", "tool_call", "prompt_enqueued", "gate_approved"
	Station   string `json:"station,omitempty"`
	Mascot    string `json:"mascot,omitempty"`
	Tool      string `json:"tool,omitempty"`
	Message   string `json:"message,omitempty"`
	Payload   any    `json:"payload,omitempty"`
	Timestamp string `json:"timestamp"`
}

// HandleSessionPrompt processa a chamada RPC "session/prompt", validando parâmetros (SEG-01)
// e despachando deterministicamente para a fila da estação ativa via queue.EnqueueStationItem.
func HandleSessionPrompt(ctx context.Context, d *RPCDispatcher, params json.RawMessage) (any, *RPCError) {
	if len(params) == 0 {
		return nil, NewRPCError(CodeInvalidParams, "params required for session/prompt")
	}

	var p SessionPromptParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid session/prompt params: %v", err))
	}

	p.SessionID = strings.TrimSpace(p.SessionID)
	if p.SessionID == "" {
		return nil, NewRPCError(CodeInvalidParams, "sessionId is required")
	}

	p.Prompt = strings.TrimSpace(p.Prompt)
	if p.Prompt == "" {
		return nil, NewRPCError(CodeInvalidParams, "prompt cannot be empty")
	}

	if d == nil || d.sessionManager == nil {
		return nil, NewRPCError(CodeInternalError, "session manager unavailable")
	}

	session, ok := d.sessionManager.GetSession(p.SessionID)
	if !ok {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("session %q not found", p.SessionID))
	}

	if session.Status == "cancelled" {
		return nil, NewRPCError(CodeServerError, fmt.Sprintf("session %q is cancelled", p.SessionID))
	}

	targetStation := strings.TrimSpace(p.Station)
	if targetStation == "" {
		targetStation = session.Station
	}
	if targetStation == "" {
		targetStation = queue.StationPMArchitect
	}
	if !queue.IsValidStation(targetStation) {
		return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid station: %q", targetStation))
	}

	itemID := fmt.Sprintf("prompt-%d", time.Now().UnixNano())
	if p.Metadata != nil && strings.TrimSpace(p.Metadata["itemId"]) != "" {
		itemID = strings.TrimSpace(p.Metadata["itemId"])
	}

	op := strings.TrimSpace(p.Operation)
	if op == "" {
		op = "prompt"
	}

	item := queue.StationItem{
		ID:        itemID,
		Title:     p.Prompt,
		Station:   targetStation,
		Operation: op,
		Status:    queue.StatusQueued,
		Metadata:  p.Metadata,
	}

	enqueued, hit, err := queue.EnqueueStationItem(session.WorkspaceDir, item)
	if err != nil {
		return nil, NewRPCError(CodeInternalError, fmt.Sprintf("failed to enqueue station item: %v", err))
	}

	// Propaga notificação via streaming ACP
	d.BroadcastSessionNotification(SessionNotification{
		SessionID: session.ID,
		Type:      "prompt_enqueued",
		Station:   targetStation,
		Message:   fmt.Sprintf("Item %s enfileirado na estação %s", enqueued.ID, targetStation),
		Payload:   enqueued,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})

	return SessionPromptResult{
		SessionID:     session.ID,
		Station:       targetStation,
		Status:        enqueued.Status,
		Item:          enqueued,
		IdempotentHit: hit,
	}, nil
}
