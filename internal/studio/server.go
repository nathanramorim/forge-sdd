package studio

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"time"
)

//go:embed assets/*
var assetsFS embed.FS

// Server implementa o servidor local Read-Only HTTP/SSE do Forge Studio.
type Server struct {
	Dir        string
	Host       string
	Port       int
	Aggregator *Aggregator
	Hub        *SSEHub
	httpServer *http.Server
	listener   net.Listener
	mu         sync.RWMutex
}

// Config contém parâmetros de inicialização do Studio.
type Config struct {
	Dir  string
	Host string
	Port int
}

// NewServer cria uma nova instância do servidor Studio.
func NewServer(cfg Config) (*Server, error) {
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 4321
	}

	agg := NewAggregator(cfg.Dir)
	hub := NewSSEHub(agg)

	s := &Server{
		Dir:        cfg.Dir,
		Host:       cfg.Host,
		Port:       cfg.Port,
		Aggregator: agg,
		Hub:        hub,
	}

	mux := http.NewServeMux()

	// 1. API Read-Only
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/events", s.handleEvents)

	// 2. Static Assets (embed.FS)
	subFS, err := fs.Sub(assetsFS, "assets")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(subFS)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Forge Studio Assets Loading...")
		})
	}

	// Middleware de Segurança (SEG-01 / SEG-02 / SEG-06): Rejeita qualquer mutação
	readOnlyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed (Studio é estritamente Read-Only)", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	})

	s.httpServer = &http.Server{
		Handler:      readOnlyHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE requer escrita contínua sem timeout rígido
	}

	return s, nil
}

// Listen inicializa o listener de rede sincronamente antes de servir requisições.
func (s *Server) Listen() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("falha ao abrir porta %s para forge-sdd studio: %w", addr, err)
	}
	s.listener = ln
	return nil
}

// Start inicia o servidor no host e porta configurados.
func (s *Server) Start() error {
	if err := s.Listen(); err != nil {
		return err
	}

	s.Hub.Start()

	s.mu.RLock()
	ln := s.listener
	s.mu.RUnlock()

	return s.httpServer.Serve(ln)
}

// URL retorna o endereço completo onde o servidor está operando.
func (s *Server) URL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return fmt.Sprintf("http://%s:%d", s.Host, s.Port)
	}
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

// Stop finaliza o servidor graciosa e silenciosamente.
func (s *Server) Stop() error {
	s.Hub.Stop()
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	snap, err := s.Aggregator.GetSnapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(snap)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	s.Hub.ServeHTTP(w, r)
}
