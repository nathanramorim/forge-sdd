package studio

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
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
	// Se cfg.Port for 0, net.Listen alocará dinamicamente uma porta efêmera livre (ideal para testes).

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

	// 1. API Read-Only e Portão FLX-01
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/spec", s.handleSpec)
	mux.HandleFunc("/api/spec/approve", s.handleSpecApprove)

	// 2. Static Assets (embed.FS)
	subFS, err := fs.Sub(assetsFS, "assets")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(subFS)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Forge Studio Assets Loading...")
		})
	}

	// Middleware de Segurança (SEG-01 / SEG-02 / SEG-06 / FLX-01): Permite POST exclusivamente em /api/spec/approve
	readOnlyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if r.URL.Path != "/api/spec/approve" {
				http.Error(w, "Method Not Allowed (Studio é estritamente Read-Only)", http.StatusMethodNotAllowed)
				return
			}
		} else if r.Method != http.MethodGet && r.Method != http.MethodHead {
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

// SpecResponse encapsula os dados retornados pelo endpoint GET /api/spec.
type SpecResponse struct {
	Path         string `json:"path"`
	AbsolutePath string `json:"absolute_path"`
	Content      string `json:"content"`
	Size         int64  `json:"size"`
	Status       string `json:"status,omitempty"`
	Station      string `json:"station,omitempty"`
	FeatureID    string `json:"feature_id,omitempty"`
}

// SpecApproveRequest representa o payload esperado por POST /api/spec/approve.
type SpecApproveRequest struct {
	SpecPath string `json:"spec_path"`
	Station  string `json:"station,omitempty"`
	TaskID   string `json:"task_id,omitempty"`
}

// SpecApproveResponse representa o payload de retorno de POST /api/spec/approve.
type SpecApproveResponse struct {
	Status   string `json:"status"`
	SpecPath string `json:"spec_path"`
	TaskID   string `json:"task_id,omitempty"`
	Station  string `json:"station,omitempty"`
	Message  string `json:"message,omitempty"`
}

// sanitizeSpecPath valida e normaliza o caminho relativo de uma especificação (SEG-01, SEG-04).
func sanitizeSpecPath(baseDir, rawPath string) (cleanSlash string, fullPathAbs string, fi os.FileInfo, err error, statusCode int) {
	if strings.TrimSpace(rawPath) == "" {
		return "", "", nil, fmt.Errorf("parâmetro 'path' é obrigatório"), http.StatusBadRequest
	}

	// 1. Rejeição de caracteres nulos ou controle
	if strings.ContainsRune(rawPath, 0) {
		return "", "", nil, fmt.Errorf("caminho contém caracteres inválidos"), http.StatusBadRequest
	}

	// 2. Bloqueio imediato de caminhos absolutos ou prefixos que indicam traversal
	if filepath.IsAbs(rawPath) || strings.HasPrefix(rawPath, "/") || strings.HasPrefix(rawPath, "\\") {
		return "", "", nil, fmt.Errorf("caminho absoluto não permitido"), http.StatusBadRequest
	}

	// 3. Normalização e sanitização com filepath.Clean
	cleanRel := filepath.Clean(filepath.ToSlash(rawPath))
	cleanRel = strings.TrimPrefix(cleanRel, "./")

	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, "../") || strings.Contains(cleanRel, "/../") {
		return "", "", nil, fmt.Errorf("tentativa de path traversal detectada"), http.StatusBadRequest
	}

	// 4. Validação de escopo estrito: deve residir sob o diretório sdd/ (SEG-04)
	cleanSlash = filepath.ToSlash(cleanRel)
	if cleanSlash != "sdd" && !strings.HasPrefix(cleanSlash, "sdd/") {
		return "", "", nil, fmt.Errorf("acesso permitido apenas a documentos sob o diretório 'sdd/'"), http.StatusForbidden
	}

	// 5. Resolução do caminho absoluto ancorado no diretório-base do projeto
	if baseDir == "" {
		baseDir = "."
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", "", nil, fmt.Errorf("erro ao resolver caminho base do projeto"), http.StatusInternalServerError
	}

	fullPath := filepath.Join(absBase, filepath.FromSlash(cleanSlash))
	fullPathAbs, err = filepath.Abs(fullPath)
	if err != nil {
		return "", "", nil, fmt.Errorf("caminho inválido"), http.StatusBadRequest
	}

	// 6. Confirmação com filepath.Rel de que não há escape de diretório (SEG-01 / SEG-02)
	relCheck, err := filepath.Rel(absBase, fullPathAbs)
	if err != nil || strings.HasPrefix(relCheck, "..") || filepath.IsAbs(relCheck) {
		return "", "", nil, fmt.Errorf("acesso negado: caminho fora da raiz do projeto"), http.StatusForbidden
	}

	// 7. Verificação de existência e tipo (não pode ser diretório)
	fi, err = os.Stat(fullPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil, fmt.Errorf("especificação não encontrada"), http.StatusNotFound
		}
		return "", "", nil, fmt.Errorf("falha ao acessar arquivo: %v", err), http.StatusInternalServerError
	}

	if fi.IsDir() {
		return "", "", nil, fmt.Errorf("o caminho especificado é um diretório, não um arquivo de especificação"), http.StatusBadRequest
	}

	return cleanSlash, fullPathAbs, fi, nil, http.StatusOK
}

func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSONError(w, "Method Not Allowed (Studio é estritamente Read-Only)", http.StatusMethodNotAllowed)
		return
	}

	rawPath := r.URL.Query().Get("path")
	cleanSlash, fullPathAbs, fi, err, statusCode := sanitizeSpecPath(s.Dir, rawPath)
	if err != nil {
		writeJSONError(w, err.Error(), statusCode)
		return
	}

	// Leitura do arquivo (UTF-8)
	data, err := os.ReadFile(fullPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, "especificação não encontrada", http.StatusNotFound)
			return
		}
		writeJSONError(w, fmt.Sprintf("erro ao ler arquivo: %v", err), http.StatusInternalServerError)
		return
	}

	contentStr := string(data)

	// Extração de metadados auxiliares (status, station, feature_id)
	status := "pending_review"
	for _, line := range strings.Split(contentStr, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "**Status:**") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "**Status:**"))
			val = strings.Trim(val, "` \t")
			if val != "" {
				status = val
				break
			}
		} else if strings.HasPrefix(trimmed, "status:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "status:"))
			val = strings.Trim(val, "`\"' \t")
			if val != "" {
				status = val
				break
			}
		}
	}

	featureID := ""
	if strings.HasPrefix(cleanSlash, "sdd/features/") {
		rest := strings.TrimPrefix(cleanSlash, "sdd/features/")
		parts := strings.Split(rest, "/")
		if len(parts) > 0 {
			featureID = strings.TrimSuffix(parts[0], ".md")
		}
	}

	resp := SpecResponse{
		Path:         cleanSlash,
		AbsolutePath: fullPathAbs,
		Content:      contentStr,
		Size:         fi.Size(),
		Status:       status,
		Station:      queue.StationPMArchitect,
		FeatureID:    featureID,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSpecApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method Not Allowed (apenas POST é permitido nesta rota)", http.StatusMethodNotAllowed)
		return
	}

	var req SpecApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "JSON inválido no corpo da requisição", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.SpecPath) == "" {
		writeJSONError(w, "campo 'spec_path' é obrigatório", http.StatusBadRequest)
		return
	}

	cleanSlash, _, _, err, statusCode := sanitizeSpecPath(s.Dir, req.SpecPath)
	if err != nil {
		writeJSONError(w, err.Error(), statusCode)
		return
	}

	taskID := req.TaskID
	if taskID == "" {
		taskID = strings.TrimSuffix(filepath.Base(cleanSlash), ".md")
	}

	targetStation := req.Station
	if targetStation == "" {
		targetStation = queue.StationPMArchitect
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// 1. Atualiza ou adiciona item aprovado na fila da estação pm-architect
	pmSQ, err := queue.LoadStationQueue(s.Dir, queue.StationPMArchitect)
	if err != nil || pmSQ == nil {
		pmSQ = &queue.StationQueue{
			Station:       queue.StationPMArchitect,
			ProcessedKeys: []string{},
			Items:         []queue.StationItem{},
		}
	}

	foundInPM := false
	for i := range pmSQ.Items {
		if pmSQ.Items[i].ArtifactPath == cleanSlash || pmSQ.Items[i].ID == taskID {
			pmSQ.Items[i].Status = queue.StatusApproved
			pmSQ.Items[i].UpdatedAt = now
			foundInPM = true
		}
	}
	if !foundInPM {
		idemKey := queue.CalculateIdempotencyKey(queue.StationPMArchitect, taskID, "approve", cleanSlash, 1)
		pmSQ.Items = append(pmSQ.Items, queue.StationItem{
			ID:             taskID,
			Title:          "Especificação: " + filepath.Base(cleanSlash),
			Station:        queue.StationPMArchitect,
			Operation:      "approve",
			Status:         queue.StatusApproved,
			ArtifactPath:   cleanSlash,
			IdempotencyKey: idemKey,
			Attempt:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		pmSQ.ProcessedKeys = append(pmSQ.ProcessedKeys, idemKey)
	}
	_ = queue.SaveStationQueue(s.Dir, pmSQ)

	// 2. Se houver fila unificada queue.json, sincroniza SpecStatus para approved
	q, _ := queue.Load(s.Dir)
	if q != nil {
		q.SpecStatus = queue.StatusApproved
		for i := range q.Items {
			if q.Items[i].SpecPath == cleanSlash || q.Items[i].TaskID == taskID {
				q.Items[i].Status = queue.StatusApproved
			}
		}
		_ = queue.Save(s.Dir, q)
	}

	// 3. Registra atividade na telemetria viva
	_ = queue.RecordActivity(s.Dir, queue.StationPMArchitect, "human_gate", "approve "+cleanSlash, "Portão FLX-01 Superado: Spec aprovada pelo humano", "success")

	// 4. Notifica SSEHub para disparar broadcast instantâneo aos clientes
	if s.Hub != nil {
		s.Hub.NotifyChange()
	}

	resp := SpecApproveResponse{
		Status:   "approved",
		SpecPath: cleanSlash,
		TaskID:   taskID,
		Station:  targetStation,
		Message:  "Especificação aprovada com sucesso no Portão FLX-01",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

