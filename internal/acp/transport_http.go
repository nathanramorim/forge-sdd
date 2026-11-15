package acp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPHandler gerencia a recepção e despacho de comandos JSON-RPC 2.0 e streaming de SSE via HTTP.
type HTTPHandler struct {
	dispatcher     *RPCDispatcher
	maxMessageSize int64
}

// NewHTTPHandler instancia um novo manipulador HTTP conectado ao RPCDispatcher fornecido.
func NewHTTPHandler(dispatcher *RPCDispatcher) *HTTPHandler {
	if dispatcher == nil {
		dispatcher = NewDispatcher()
	}
	return &HTTPHandler{
		dispatcher:     dispatcher,
		maxMessageSize: DefaultMaxMessageSize, // 2 MB (SEG-01, SEG-05)
	}
}

// SetMaxMessageSize configura o tamanho máximo permitido de payload (SEG-01).
func (h *HTTPHandler) SetMaxMessageSize(size int64) {
	if size > 0 {
		h.maxMessageSize = size
	}
}

// MaxMessageSize retorna o tamanho máximo configurado.
func (h *HTTPHandler) MaxMessageSize() int64 {
	return h.maxMessageSize
}

// Dispatcher retorna o despachador RPC associado.
func (h *HTTPHandler) Dispatcher() *RPCDispatcher {
	return h.dispatcher
}

// ServeHTTP processa requisições JSON-RPC 2.0 via HTTP POST (ex: /api/acp/rpc).
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.Error(w, "Method Not Allowed (ACP RPC requires HTTP POST)", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	limit := h.maxMessageSize
	if limit <= 0 {
		limit = DefaultMaxMessageSize
	}

	// SEG-01 / SEG-05: Proteção contra payloads excessivos via io.LimitReader (+1 para detecção de estouro)
	body, err := io.ReadAll(io.LimitReader(r.Body, limit+1))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := NewErrorResponse(nil, NewRPCError(CodeParseError, fmt.Sprintf("Failed to read request body: %v", err)))
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if int64(len(body)) > limit {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		resp := NewErrorResponse(nil, NewRPCError(CodeInvalidRequest, fmt.Sprintf("Request payload exceeds limit of %d bytes", limit)))
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := NewErrorResponse(nil, NewRPCError(CodeParseError, "Parse error: empty request body"))
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := h.dispatcher.Dispatch(r.Context(), trimmed)
	if resp == nil {
		// Notificações JSON-RPC não geram resposta no protocolo HTTP (RFC JSON-RPC 2.0 over HTTP)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ServeEvents transmite notificações ACP em tempo real via Server-Sent Events (SSE) (ex: /api/acp/events).
func (h *HTTPHandler) ServeEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.Error(w, "Method Not Allowed (ACP Events requires HTTP GET)", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	clientChan := make(chan Notification, 64)
	h.dispatcher.SubscribeNotifications(clientChan)
	defer func() {
		h.dispatcher.UnsubscribeNotifications(clientChan)
	}()

	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case notif, ok := <-clientChan:
			if !ok {
				return
			}
			data, err := json.Marshal(notif)
			if err != nil {
				continue
			}
			eventName := notif.Method
			if eventName == "" {
				eventName = "message"
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, string(data))
			flusher.Flush()
		}
	}
}

// EventsHandler retorna um http.Handler dedicado ao endpoint de streaming SSE.
func (h *HTTPHandler) EventsHandler() http.Handler {
	return http.HandlerFunc(h.ServeEvents)
}
