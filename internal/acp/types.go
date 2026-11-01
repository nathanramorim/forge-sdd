package acp

import (
	"encoding/json"
	"fmt"
	"time"
)

// Constantes canônicas do protocolo JSON-RPC 2.0 e ACP
const (
	JSONRPCVersion         = "2.0"
	DefaultProtocolVersion = "1.0.0"
	ServerName             = "forge-sdd-acp"
)

// Códigos padronizados de erro JSON-RPC 2.0 e ACP
const (
	CodeParseError     = -32700 // Erro ao interpretar JSON
	CodeInvalidRequest = -32600 // Estrutura da requisição inválida
	CodeMethodNotFound = -32601 // Método não encontrado
	CodeInvalidParams  = -32602 // Parâmetros inválidos
	CodeInternalError  = -32603 // Erro interno do servidor
	CodeServerError    = -32001 // Erro de servidor/protocolo (ex: rejeição de portão ou sessão)
)

// RPCError representa um erro JSON-RPC 2.0 canônico.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error satisfaz a interface standard error de Go (COD-03).
func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}
	if e.Data != nil {
		return fmt.Sprintf("rpc error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// NewRPCError constrói uma instância de RPCError enriquecida.
func NewRPCError(code int, message string, data ...any) *RPCError {
	var d any
	if len(data) > 0 {
		d = data[0]
	}
	return &RPCError{
		Code:    code,
		Message: message,
		Data:    d,
	}
}

// Request representa uma requisição JSON-RPC 2.0.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response representa uma resposta JSON-RPC 2.0 canônica.
type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// MarshalJSON garante a conformidade estrita com a especificação JSON-RPC 2.0:
// quando Error estiver preenchido, Result é omitido; quando Error for nil, Result é emitido.
func (r Response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		type errResponse struct {
			JSONRPC string    `json:"jsonrpc"`
			ID      any       `json:"id"`
			Error   *RPCError `json:"error"`
		}
		return json.Marshal(errResponse{
			JSONRPC: r.JSONRPC,
			ID:      r.ID,
			Error:   r.Error,
		})
	}

	type okResponse struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id"`
		Result  any    `json:"result"`
	}
	return json.Marshal(okResponse{
		JSONRPC: r.JSONRPC,
		ID:      r.ID,
		Result:  r.Result,
	})
}

// UnmarshalJSON desserializa uma resposta JSON-RPC 2.0.
func (r *Response) UnmarshalJSON(data []byte) error {
	type rawResponse struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *RPCError       `json:"error,omitempty"`
	}
	var raw rawResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.JSONRPC = raw.JSONRPC
	r.ID = raw.ID
	r.Error = raw.Error
	if len(raw.Result) > 0 {
		r.Result = raw.Result
	}
	return nil
}

// Notification representa uma mensagem unidirecional sem ID (Request sem resposta obrigatória).
type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// ClientInfo descreve o cliente conectado (IDE, Studio ou script).
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ClientCapabilities descreve capacidades oferecidas pelo cliente externo.
type ClientCapabilities struct {
	Fs       map[string]any `json:"fs,omitempty"`
	Terminal bool           `json:"terminal,omitempty"`
	Prompts  bool           `json:"prompts,omitempty"`
}

// InitializeParams parâmetros recebidos no handshake "initialize".
type InitializeParams struct {
	ProtocolVersion string              `json:"protocolVersion"`
	ClientInfo      ClientInfo          `json:"clientInfo"`
	Capabilities    *ClientCapabilities `json:"capabilities,omitempty"`
}

// ServerInfo descreve a identificação do servidor ACP forge-sdd.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities expõe os recursos do servidor forge-sdd para o cliente.
type ServerCapabilities struct {
	Sessions bool `json:"sessions"`
	Gates    bool `json:"gates"`
	Prompts  bool `json:"prompts"`
}

// InitializeResult resultado retornado pelo handshake "initialize".
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

// SessionNewParams parâmetros para abertura de nova sessão (session/new).
type SessionNewParams struct {
	WorkspaceDir string `json:"workspaceDir,omitempty"`
	Feature      string `json:"feature,omitempty"`
	Branch       string `json:"branch,omitempty"`
}

// SessionLoadParams parâmetros para carregamento de sessão existente (session/load).
type SessionLoadParams struct {
	SessionID string `json:"sessionId"`
}

// SessionCancelParams parâmetros para cancelamento de sessão ativa (session/cancel).
type SessionCancelParams struct {
	SessionID string `json:"sessionId"`
	Reason    string `json:"reason,omitempty"`
}

// SessionResult resultado genérico de operações sobre sessões ACP.
type SessionResult struct {
	SessionID    string `json:"sessionId"`
	Status       string `json:"status"`
	WorkspaceDir string `json:"workspaceDir,omitempty"`
}

// Helpers para construção de respostas
func NewSuccessResponse(id any, result any) Response {
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id any, err *RPCError) Response {
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error:   err,
	}
}

func NewNotification(method string, params any) Notification {
	return Notification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
	}
}

// SessionState define a máquina de estados operacionais de uma sessão no pool ACP.
type SessionState string

const (
	SessionStateStandby         SessionState = "standby"
	SessionStateProcessing      SessionState = "processing"
	SessionStateApprovalPending SessionState = "approval_pending"
	SessionStateError           SessionState = "error"
	SessionStateOffline         SessionState = "offline"
)

// ValidSessionStates lista todos os estados canônicos de sessão válidos.
func ValidSessionStates() []SessionState {
	return []SessionState{
		SessionStateStandby,
		SessionStateProcessing,
		SessionStateApprovalPending,
		SessionStateError,
		SessionStateOffline,
	}
}

// IsValidSessionState verifica se o estado fornecido pertence aos estados canônicos.
func IsValidSessionState(state SessionState) bool {
	for _, s := range ValidSessionStates() {
		if s == state {
			return true
		}
	}
	return false
}

// PooledSession representa uma sessão canônica pré-aquecida e gerenciada pelo SessionPool.
type PooledSession struct {
	SessionID           string            `json:"sessionId"`
	Station             string            `json:"station"`
	WorkspaceDir        string            `json:"workspaceDir"`
	State               SessionState      `json:"state"`
	ActiveItemID        string            `json:"activeItemId,omitempty"`
	LastSeenAt          time.Time         `json:"lastSeenAt"`
	AttachedClients     int               `json:"attachedClients"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	ProcessingStartedAt time.Time         `json:"processingStartedAt,omitempty"`
	LastTelemetryAt     time.Time         `json:"lastTelemetryAt,omitempty"`
}

// Clone cria uma cópia defensiva profunda da sessão no pool.
func (s *PooledSession) Clone() *PooledSession {
	if s == nil {
		return nil
	}
	cp := *s
	if s.Metadata != nil {
		cp.Metadata = make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}

// SessionSummary representa um resumo condensado da telemetria de uma sessão no pool.
type SessionSummary struct {
	SessionID       string       `json:"sessionId"`
	Station         string       `json:"station"`
	State           SessionState `json:"state"`
	LastSeenAt      time.Time    `json:"lastSeenAt"`
	ActiveItemID    string       `json:"activeItemId,omitempty"`
	AttachedClients int          `json:"attachedClients"`
}

// Summary extrai um SessionSummary a partir de uma PooledSession.
func (s *PooledSession) Summary() SessionSummary {
	return SessionSummary{
		SessionID:       s.SessionID,
		Station:         s.Station,
		State:           s.State,
		LastSeenAt:      s.LastSeenAt,
		ActiveItemID:    s.ActiveItemID,
		AttachedClients: s.AttachedClients,
	}
}

// PoolWarmupParams define os parâmetros para inicialização/pré-aquecimento do SessionPool.
type PoolWarmupParams struct {
	WorkspaceDir string `json:"workspaceDir,omitempty"`
}

// PoolWarmupResult define o resultado retornado pelo comando pool/warmup.
type PoolWarmupResult struct {
	Status        string           `json:"status"`
	SessionsCount int              `json:"sessionsCount"`
	Sessions      []*PooledSession `json:"sessions"`
}

// PoolStatusResult define o resultado de telemetria e integridade retornado por pool/status.
type PoolStatusResult struct {
	Healthy  bool             `json:"healthy"`
	Sessions []*PooledSession `json:"sessions"`
}
