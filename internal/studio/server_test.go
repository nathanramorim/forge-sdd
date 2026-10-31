package studio_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/studio"
)

func setupTestStudioServer(t *testing.T) (*studio.Server, string) {
	t.Helper()
	tempDir := t.TempDir()

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	if err != nil {
		t.Fatalf("falha ao instanciar servidor studio: %v", err)
	}

	if err := srv.Listen(); err != nil {
		t.Fatalf("falha no listen do servidor: %v", err)
	}

	go func() {
		_ = srv.Start()
	}()
	t.Cleanup(func() {
		_ = srv.Stop()
	})

	return srv, tempDir
}

func TestAPISpec_Success(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Cria estrutura de documentação SDD com spec válida
	specRelPath := filepath.Join("sdd", "features", "feat-42-test", "discovery.md")
	fullPath := filepath.Join(tempDir, specRelPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("falha ao criar pastas: %v", err)
	}

	specContent := `# Discovery 42 — Test Spec

**Status:** ` + "`in_progress`" + `

Especificação de arquitetura e critérios de teste.`

	if err := os.WriteFile(fullPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("falha ao gravar spec de teste: %v", err)
	}

	// 1. Requisição bem-sucedida para subpasta de feature
	reqURL := srv.URL() + "/api/spec?path=sdd/features/feat-42-test/discovery.md"
	resp, err := http.Get(reqURL)
	if err != nil {
		t.Fatalf("falha na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 OK, obtido %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("esperado Content-Type application/json, obtido %s", ct)
	}

	var data studio.SpecResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("falha ao decodificar JSON de resposta: %v", err)
	}

	if data.Path != "sdd/features/feat-42-test/discovery.md" {
		t.Errorf("Path incorreto: esperado 'sdd/features/feat-42-test/discovery.md', obtido %q", data.Path)
	}
	if data.AbsolutePath != fullPath {
		t.Errorf("AbsolutePath incorreto: esperado %q, obtido %q", fullPath, data.AbsolutePath)
	}
	if !strings.Contains(data.Content, "Discovery 42 — Test Spec") {
		t.Errorf("Content não contém título esperado: %q", data.Content)
	}
	if data.Size != int64(len(specContent)) {
		t.Errorf("Size incorreto: esperado %d, obtido %d", len(specContent), data.Size)
	}
	if data.Status != "in_progress" {
		t.Errorf("Status incorreto: esperado 'in_progress', obtido %q", data.Status)
	}
	if data.Station != "pm-architect" {
		t.Errorf("Station incorreta: esperado 'pm-architect', obtido %q", data.Station)
	}
	if data.FeatureID != "feat-42-test" {
		t.Errorf("FeatureID incorreto: esperado 'feat-42-test', obtido %q", data.FeatureID)
	}

	// 2. Requisição para arquivo direto em sdd/features/
	flatRelPath := filepath.Join("sdd", "features", "feat-01-direct.md")
	flatFullPath := filepath.Join(tempDir, flatRelPath)
	flatContent := `# Feature Direta

status: approved

Conteúdo direto.`
	if err := os.WriteFile(flatFullPath, []byte(flatContent), 0644); err != nil {
		t.Fatalf("falha ao gravar flat spec: %v", err)
	}

	respFlat, err := http.Get(srv.URL() + "/api/spec?path=sdd/features/feat-01-direct.md")
	if err != nil {
		t.Fatalf("falha na requisição flat: %v", err)
	}
	defer respFlat.Body.Close()

	if respFlat.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 OK na flat spec, obtido %d", respFlat.StatusCode)
	}

	var dataFlat studio.SpecResponse
	if err := json.NewDecoder(respFlat.Body).Decode(&dataFlat); err != nil {
		t.Fatalf("falha ao decodificar JSON flat: %v", err)
	}

	if dataFlat.FeatureID != "feat-01-direct" {
		t.Errorf("FeatureID incorreto para arquivo flat: esperado 'feat-01-direct', obtido %q", dataFlat.FeatureID)
	}
	if dataFlat.Status != "approved" {
		t.Errorf("Status incorreto para arquivo flat: esperado 'approved', obtido %q", dataFlat.Status)
	}
}

func TestAPISpec_PathTraversalAndScope(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Cria arquivo fora de sdd/ para testar isolamento de escopo
	outsideFile := filepath.Join(tempDir, "sensitive.txt")
	if err := os.WriteFile(outsideFile, []byte("SECRET_TOKEN"), 0644); err != nil {
		t.Fatalf("falha ao criar arquivo de teste externo: %v", err)
	}

	// Cria pasta sdd/features para testar requisição contra diretório
	sddDir := filepath.Join(tempDir, "sdd", "features")
	if err := os.MkdirAll(sddDir, 0755); err != nil {
		t.Fatalf("falha ao criar diretório sdd: %v", err)
	}

	tests := []struct {
		name         string
		urlQuery     string
		expectedCode int
	}{
		{
			name:         "parâmetro ausente",
			urlQuery:     "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "parâmetro vazio",
			urlQuery:     "path=",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "apenas espaços em branco",
			urlQuery:     "path=%20%20",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "caractere nulo no caminho",
			urlQuery:     "path=sdd%00features",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "caminho absoluto unix",
			urlQuery:     "path=/etc/passwd",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "caminho absoluto windows",
			urlQuery:     "path=\\etc\\passwd",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "path traversal relativo direto",
			urlQuery:     "path=../../etc/passwd",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "path traversal simples ponto-ponto",
			urlQuery:     "path=..",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "path traversal dentro do prefixo sdd",
			urlQuery:     "path=sdd/../../sensitive.txt",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "path traversal profundo a partir de features",
			urlQuery:     "path=sdd/features/../../../sensitive.txt",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "arquivo fora do escopo sdd",
			urlQuery:     "path=sensitive.txt",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "diretório em vez de arquivo",
			urlQuery:     "path=sdd/features",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			endpoint := srv.URL() + "/api/spec"
			if tc.urlQuery != "" {
				endpoint += "?" + tc.urlQuery
			}

			resp, err := http.Get(endpoint)
			if err != nil {
				t.Fatalf("erro ao executar requisição HTTP: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("status code divergente: esperado %d, obtido %d", tc.expectedCode, resp.StatusCode)
			}

			// Resposta deve ser um JSON de erro estruturado
			var errPayload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errPayload); err != nil {
				t.Fatalf("falha ao decodificar JSON de erro: %v", err)
			}

			if errMsg, ok := errPayload["error"]; !ok || strings.TrimSpace(errMsg) == "" {
				t.Errorf("campo 'error' vazio ou ausente na resposta: %v", errPayload)
			}
		})
	}
}

func TestAPISpec_NotFound(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

	resp, err := http.Get(srv.URL() + "/api/spec?path=sdd/features/feat-inexistente/discovery.md")
	if err != nil {
		t.Fatalf("erro ao executar requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("esperado status 404 Not Found, obtido %d", resp.StatusCode)
	}

	var errPayload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errPayload); err != nil {
		t.Fatalf("falha ao decodificar JSON de erro: %v", err)
	}

	if !strings.Contains(errPayload["error"], "não encontrada") {
		t.Errorf("mensagem de erro inesperada: %v", errPayload)
	}
}

func TestAPISpec_MethodNotAllowed(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

	req, err := http.NewRequest(http.MethodPost, srv.URL()+"/api/spec?path=sdd/features/feat-42/discovery.md", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("erro ao construir requisição POST: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("erro ao executar POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("esperado status 405 Method Not Allowed, obtido %d", resp.StatusCode)
	}
}

func TestAPISpecApprove(t *testing.T) {
	t.Run("Success", TestAPISpecApprove_Success)
	t.Run("SecurityAndValidation", TestAPISpecApprove_SecurityAndValidation)
	t.Run("SSEBroadcast", TestAPISpecApprove_SSEBroadcast)
}

func TestAPISpecApprove_Success(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Cria arquivo de especificação válido
	specRel := filepath.Join("sdd", "features", "feat-63-test", "discovery.md")
	specFull := filepath.Join(tempDir, specRel)
	if err := os.MkdirAll(filepath.Dir(specFull), 0755); err != nil {
		t.Fatalf("falha ao criar pasta da spec: %v", err)
	}
	specContent := "# Discovery 63 — Teste de Aprovação\n\n**Status:** `pending_review`\n"
	if err := os.WriteFile(specFull, []byte(specContent), 0644); err != nil {
		t.Fatalf("falha ao criar spec de teste: %v", err)
	}

	// 1. Inicializa fila da estação pm-architect com tarefa pendente
	item := queue.StationItem{
		ID:           "feat-63-01",
		Title:        "Concepção da Feature 63",
		Station:      queue.StationPMArchitect,
		Operation:    "discovery",
		Status:       "pending_review",
		ArtifactPath: filepath.ToSlash(specRel),
		Attempt:      1,
	}
	if _, _, err := queue.EnqueueStationItem(tempDir, item); err != nil {
		t.Fatalf("falha ao enfileirar item na estação: %v", err)
	}

	// 2. Chama POST /api/spec/approve
	payload := []byte(`{"spec_path": "` + filepath.ToSlash(specRel) + `", "task_id": "feat-63-01"}`)
	req, err := http.NewRequest(http.MethodPost, srv.URL()+"/api/spec/approve", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("falha ao criar requisição POST: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("falha ao enviar POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 OK, obtido %d", resp.StatusCode)
	}

	var resData studio.SpecApproveResponse
	if err := json.NewDecoder(resp.Body).Decode(&resData); err != nil {
		t.Fatalf("falha ao decodificar JSON de resposta: %v", err)
	}

	if resData.Status != "approved" {
		t.Errorf("status inesperado no JSON: esperado 'approved', obtido %q", resData.Status)
	}
	if resData.SpecPath != filepath.ToSlash(specRel) {
		t.Errorf("spec_path inesperado: esperado %q, obtido %q", filepath.ToSlash(specRel), resData.SpecPath)
	}

	// 3. Verifica se a fila de pm-architect foi atualizada para approved
	sq, err := queue.LoadStationQueue(tempDir, queue.StationPMArchitect)
	if err != nil {
		t.Fatalf("falha ao carregar fila da estação após aprovação: %v", err)
	}
	var foundApproved bool
	for _, it := range sq.Items {
		if it.ID == "feat-63-01" && it.Status == queue.StatusApproved {
			foundApproved = true
			break
		}
	}
	if !foundApproved {
		t.Errorf("item 'feat-63-01' não foi marcado como 'approved' na fila: %+v", sq.Items)
	}

	// 4. Teste de Idempotência: aprovar novamente a mesma spec não causa erro
	resp2, err := http.Post(srv.URL()+"/api/spec/approve", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("falha na segunda requisição de aprovação: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200 OK na segunda chamada idempotente, obtido %d", resp2.StatusCode)
	}
}

func TestAPISpecApprove_SecurityAndValidation(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Cria arquivo válido dentro de sdd/
	specRel := filepath.Join("sdd", "features", "feat-valid", "spec.md")
	specFull := filepath.Join(tempDir, specRel)
	_ = os.MkdirAll(filepath.Dir(specFull), 0755)
	_ = os.WriteFile(specFull, []byte("# Valid Spec"), 0644)

	// Cria arquivo fora de sdd/ para teste de escopo
	outsideFile := filepath.Join(tempDir, "outside.txt")
	_ = os.WriteFile(outsideFile, []byte("sensitive"), 0644)

	tests := []struct {
		name         string
		method       string
		path         string
		payload      string
		expectedCode int
	}{
		{
			name:         "JSON mal formatado",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{invalid json`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "spec_path vazio",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": ""}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "caractere nulo",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": "sdd/\u0000features"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "caminho absoluto unix",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": "/etc/passwd"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "path traversal relativo",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": "../../outside.txt"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "fora do escopo sdd",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": "outside.txt"}`,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "especificação inexistente",
			method:       http.MethodPost,
			path:         "/api/spec/approve",
			payload:      `{"spec_path": "sdd/features/nonexistent.md"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "GET em /api/spec/approve não permitido",
			method:       http.MethodGet,
			path:         "/api/spec/approve",
			payload:      "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "POST em /api/state não permitido (Read-Only)",
			method:       http.MethodPost,
			path:         "/api/state",
			payload:      `{}`,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "POST em /api/spec não permitido (Read-Only)",
			method:       http.MethodPost,
			path:         "/api/spec",
			payload:      `{}`,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "POST em rota raiz / não permitido",
			method:       http.MethodPost,
			path:         "/",
			payload:      `{}`,
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body *strings.Reader
			if tc.payload != "" {
				body = strings.NewReader(tc.payload)
			} else {
				body = strings.NewReader("")
			}

			req, err := http.NewRequest(tc.method, srv.URL()+tc.path, body)
			if err != nil {
				t.Fatalf("erro ao construir requisição: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("erro ao executar requisição HTTP: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("status code divergente: esperado %d, obtido %d", tc.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestAPISpecApprove_SSEBroadcast(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Cria arquivo válido de spec
	specRel := filepath.Join("sdd", "features", "feat-sse-test", "discovery.md")
	specFull := filepath.Join(tempDir, specRel)
	_ = os.MkdirAll(filepath.Dir(specFull), 0755)
	_ = os.WriteFile(specFull, []byte("# SSE Test Spec"), 0644)

	// Conecta cliente SSE em /api/events
	respSSE, err := http.Get(srv.URL() + "/api/events")
	if err != nil {
		t.Fatalf("falha ao conectar no SSE /api/events: %v", err)
	}
	defer respSSE.Body.Close()

	if respSSE.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 no SSE, obtido %d", respSSE.StatusCode)
	}

	reader := bufio.NewReader(respSSE.Body)

	// Lê o primeiro snapshot que é enviado imediatamente na conexão
	doneInitial := make(chan struct{})
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data:") {
				close(doneInitial)
				return
			}
		}
	}()

	select {
	case <-doneInitial:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout aguardando snapshot inicial do SSE")
	}

	// Canal para capturar o broadcast após aprovação
	broadcastReceived := make(chan string, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data:") {
				dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				broadcastReceived <- dataStr
				return
			}
		}
	}()

	// Dispara aprovação da spec via POST /api/spec/approve
	approvePayload := `{"spec_path": "` + filepath.ToSlash(specRel) + `", "task_id": "feat-sse-01"}`
	respApprove, err := http.Post(srv.URL()+"/api/spec/approve", "application/json", strings.NewReader(approvePayload))
	if err != nil {
		t.Fatalf("falha ao enviar POST de aprovação: %v", err)
	}
	defer respApprove.Body.Close()

	if respApprove.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 na aprovação, obtido %d", respApprove.StatusCode)
	}

	// Verifica se o SSE recebeu o broadcast da mudança em menos de 1 segundo
	select {
	case snapJSON := <-broadcastReceived:
		if !strings.Contains(snapJSON, "pm-architect") {
			t.Errorf("snapshot SSE não contém dados da estação: %s", snapJSON)
		}
	case <-time.After(1500 * time.Millisecond):
		t.Fatalf("timeout: broadcast SSE não foi recebido após POST /api/spec/approve")
	}
}
