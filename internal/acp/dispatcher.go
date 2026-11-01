package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// HandlerFunc processa uma chamada RPC e retorna um resultado ou um erro canônico.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (any, *RPCError)

// NotificationHandlerFunc processa uma notificação unidirecional (sem resposta).
type NotificationHandlerFunc func(ctx context.Context, params json.RawMessage)

// RPCDispatcher gerencia o registro e roteamento thread-safe de métodos JSON-RPC 2.0.
type RPCDispatcher struct {
	mu                      sync.RWMutex
	handlers                map[string]HandlerFunc
	notificationHandlers    map[string]NotificationHandlerFunc
	subMu                   sync.Mutex
	notificationSubscribers []chan Notification
	serverVersion           string
	initialized             bool
	sessionManager          *SessionManager
}

// NewDispatcher instancia um despachador com versão padrão e métodos base registrados.
func NewDispatcher() *RPCDispatcher {
	return NewDispatcherWithVersion("3.2.1-beta")
}

// NewDispatcherWithVersion instancia o despachador com versão configurada.
func NewDispatcherWithVersion(version string) *RPCDispatcher {
	if version == "" {
		version = "3.2.1-beta"
	}
	d := &RPCDispatcher{
		handlers:                make(map[string]HandlerFunc),
		notificationHandlers:    make(map[string]NotificationHandlerFunc),
		notificationSubscribers: make([]chan Notification, 0),
		serverVersion:           version,
		sessionManager:          NewSessionManager(),
	}

	d.registerDefaultHandlers()
	return d
}

// registerDefaultHandlers registra os métodos canônicos do protocolo ACP:
// initialize, session/new, session/load, session/cancel, session/prompt, gate/approve.
func (d *RPCDispatcher) registerDefaultHandlers() {
	d.RegisterHandler("initialize", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		var initParams InitializeParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &initParams); err != nil {
				return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid initialize params: %v", err))
			}
		}

		d.mu.Lock()
		d.initialized = true
		serverVersion := d.serverVersion
		d.mu.Unlock()

		return InitializeResult{
			ProtocolVersion: DefaultProtocolVersion,
			ServerInfo: ServerInfo{
				Name:    ServerName,
				Version: serverVersion,
			},
			Capabilities: ServerCapabilities{
				Sessions: true,
				Gates:    true,
				Prompts:  true,
			},
		}, nil
	})

	d.RegisterHandler("session/new", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		var p SessionNewParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid session/new params: %v", err))
			}
		}

		sess, err := d.sessionManager.CreateSession(p)
		if err != nil {
			return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("invalid session/new: %v", err))
		}

		return SessionResult{
			SessionID:    sess.ID,
			Status:       sess.Status,
			WorkspaceDir: sess.WorkspaceDir,
		}, nil
	})

	d.RegisterHandler("session/load", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		var p SessionLoadParams
		if err := json.Unmarshal(params, &p); err != nil || p.SessionID == "" {
			return nil, NewRPCError(CodeInvalidParams, "sessionId is required")
		}

		sess, ok := d.sessionManager.GetSession(p.SessionID)
		if !ok {
			return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("session %q not found", p.SessionID))
		}

		return SessionResult{
			SessionID:    sess.ID,
			Status:       "loaded",
			WorkspaceDir: sess.WorkspaceDir,
		}, nil
	})

	d.RegisterHandler("session/cancel", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		var p SessionCancelParams
		if err := json.Unmarshal(params, &p); err != nil || p.SessionID == "" {
			return nil, NewRPCError(CodeInvalidParams, "sessionId is required")
		}

		sess, err := d.sessionManager.CancelSession(p.SessionID, p.Reason)
		if err != nil {
			return nil, NewRPCError(CodeInvalidParams, err.Error())
		}

		return SessionResult{
			SessionID:    sess.ID,
			Status:       sess.Status,
			WorkspaceDir: sess.WorkspaceDir,
		}, nil
	})

	d.RegisterHandler("session/prompt", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		return HandleSessionPrompt(ctx, d, params)
	})

	d.RegisterHandler("gate/approve", func(ctx context.Context, params json.RawMessage) (any, *RPCError) {
		return HandleGateApprove(ctx, d, params)
	})
}

// RegisterHandler registra ou sobrescreve um handler para um método JSON-RPC (thread-safe).
func (d *RPCDispatcher) RegisterHandler(method string, handler HandlerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[method] = handler
}

// RegisterNotificationHandler registra um handler para notificações (sem retorno ao cliente).
func (d *RPCDispatcher) RegisterNotificationHandler(method string, handler NotificationHandlerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.notificationHandlers[method] = handler
}

// IsInitialized retorna se o handshake de inicialização já foi concluído.
func (d *RPCDispatcher) IsInitialized() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.initialized
}

// SetInitialized altera explicitamente o status de inicialização.
func (d *RPCDispatcher) SetInitialized(v bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.initialized = v
}

// ServerVersion retorna a versão do servidor ACP.
func (d *RPCDispatcher) ServerVersion() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.serverVersion
}

// SessionManager retorna o gerenciador de sessões ativas do ACP.
func (d *RPCDispatcher) SessionManager() *SessionManager {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.sessionManager
}

// SetSessionManager configura um gerenciador de sessões customizado no despachador.
func (d *RPCDispatcher) SetSessionManager(sm *SessionManager) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sessionManager = sm
}

// BroadcastSessionNotification emite uma notificação no canal unidirecional "session/notification".
func (d *RPCDispatcher) BroadcastSessionNotification(n SessionNotification) {
	if n.Timestamp == "" {
		n.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	d.BroadcastNotification(NewNotification("session/notification", n))
}

// SubscribeNotifications inscreve um canal para receber notificações emitidas pelo dispatcher.
func (d *RPCDispatcher) SubscribeNotifications(ch chan Notification) {
	d.subMu.Lock()
	defer d.subMu.Unlock()
	d.notificationSubscribers = append(d.notificationSubscribers, ch)
}

// UnsubscribeNotifications remove a inscrição de um canal de notificações.
func (d *RPCDispatcher) UnsubscribeNotifications(ch chan Notification) {
	d.subMu.Lock()
	defer d.subMu.Unlock()
	for i, c := range d.notificationSubscribers {
		if c == ch {
			d.notificationSubscribers = append(d.notificationSubscribers[:i], d.notificationSubscribers[i+1:]...)
			break
		}
	}
}

// BroadcastNotification envia uma notificação para todos os ouvintes inscritos de forma não-bloqueante.
func (d *RPCDispatcher) BroadcastNotification(n Notification) {
	d.subMu.Lock()
	defer d.subMu.Unlock()
	for _, ch := range d.notificationSubscribers {
		select {
		case ch <- n:
		default:
			// Canal ocupado: evita bloquear o despachador
		}
	}
}

// rawEnvelope é uma estrutura intermediária para inspecionar requisições e notificações JSON-RPC 2.0.
type rawEnvelope struct {
	JSONRPC *string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Method  *string          `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// Dispatch processa de forma síncrona uma mensagem bruta em JSON.
// Retorna um ponteiro para Response se for uma Request, ou nil se for uma Notification ou mensagem vazia.
func (d *RPCDispatcher) Dispatch(ctx context.Context, raw []byte) *Response {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}

	var env rawEnvelope
	if err := json.Unmarshal(trimmed, &env); err != nil {
		resp := NewErrorResponse(nil, NewRPCError(CodeParseError, fmt.Sprintf("Parse error: %v", err)))
		return &resp
	}

	var idVal any
	if env.ID != nil {
		_ = json.Unmarshal(*env.ID, &idVal)
	}

	// Validação de conformidade com JSON-RPC 2.0
	if env.JSONRPC == nil || *env.JSONRPC != JSONRPCVersion {
		resp := NewErrorResponse(idVal, NewRPCError(CodeInvalidRequest, "Invalid Request: 'jsonrpc' must be exactly '2.0'"))
		return &resp
	}

	if env.Method == nil || *env.Method == "" {
		resp := NewErrorResponse(idVal, NewRPCError(CodeInvalidRequest, "Invalid Request: 'method' field is required"))
		return &resp
	}

	// Notificação: ausência de membro "id"
	if env.ID == nil {
		d.mu.RLock()
		notifHandler, ok := d.notificationHandlers[*env.Method]
		d.mu.RUnlock()

		if ok && notifHandler != nil {
			notifHandler(ctx, env.Params)
		}
		return nil // Notificações nunca geram resposta (RFC JSON-RPC 2.0)
	}

	// Requisição com id
	d.mu.RLock()
	handler, found := d.handlers[*env.Method]
	d.mu.RUnlock()

	if !found || handler == nil {
		resp := NewErrorResponse(idVal, NewRPCError(CodeMethodNotFound, fmt.Sprintf("Method %q not found", *env.Method)))
		return &resp
	}

	result, rpcErr := handler(ctx, env.Params)
	if rpcErr != nil {
		resp := NewErrorResponse(idVal, rpcErr)
		return &resp
	}

	resp := NewSuccessResponse(idVal, result)
	return &resp
}

// DispatchAsync despacha o processamento da mensagem em uma goroutine assíncrona.
func (d *RPCDispatcher) DispatchAsync(ctx context.Context, raw []byte, onResponse func(*Response)) {
	go func() {
		resp := d.Dispatch(ctx, raw)
		if resp != nil && onResponse != nil {
			onResponse(resp)
		}
	}()
}
