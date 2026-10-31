package studio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEHub gerencia clientes conectados ao stream de eventos /api/events.
type SSEHub struct {
	mu        sync.RWMutex
	clients   map[chan string]bool
	agg       *Aggregator
	lastHash  string
	stopChan  chan struct{}
}

// NewSSEHub instancia um novo despachante de eventos SSE.
func NewSSEHub(agg *Aggregator) *SSEHub {
	return &SSEHub{
		clients:  make(map[chan string]bool),
		agg:      agg,
		stopChan: make(chan struct{}),
	}
}

// Start inicia o loop de monitoramento periódico (polling passivo a cada 600ms).
func (h *SSEHub) Start() {
	ticker := time.NewTicker(600 * time.Millisecond)
	go func() {
		for {
			select {
			case <-h.stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				h.checkAndBroadcast()
			}
		}
	}()
}

// Stop encerra o loop de monitoramento.
func (h *SSEHub) Stop() {
	close(h.stopChan)
}

// NotifyChange força uma atualização e emissão imediata do estado para os clientes conectados.
func (h *SSEHub) NotifyChange() {
	snap, err := h.agg.GetSnapshot()
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastHash = snap.StateHash
	data, err := json.Marshal(snap)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("event: state\ndata: %s\n\n", string(data))
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (h *SSEHub) checkAndBroadcast() {
	snap, err := h.agg.GetSnapshot()
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if snap.StateHash != h.lastHash {
		h.lastHash = snap.StateHash
		data, err := json.Marshal(snap)
		if err != nil {
			return
		}
		msg := fmt.Sprintf("event: state\ndata: %s\n\n", string(data))
		for ch := range h.clients {
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

// ServeHTTP atende conexões SSE no endpoint /api/events.
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string, 10)

	h.mu.Lock()
	h.clients[clientChan] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		close(clientChan)
		h.mu.Unlock()
	}()

	// Envia snapshot inicial imediatamente na conexão
	if snap, err := h.agg.GetSnapshot(); err == nil {
		if data, err := json.Marshal(snap); err == nil {
			fmt.Fprintf(w, "event: state\ndata: %s\n\n", string(data))
			flusher.Flush()
		}
	}

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-clientChan:
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}
