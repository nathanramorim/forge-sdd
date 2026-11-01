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

	// 4. Verifica se a fila planner.json recebeu o item enfileirado para planejamento
	plannerSQ, err := queue.LoadStationQueue(tempDir, queue.StationPlanner)
	if err != nil {
		t.Fatalf("falha ao carregar fila do planejador (planner.json): %v", err)
	}
	var foundPlannerItem bool
	for _, it := range plannerSQ.Items {
		if it.ID == "feat-63-01" && (it.Status == queue.StatusQueued || it.Status == queue.StatusBuilding) && it.Station == queue.StationPlanner {
			foundPlannerItem = true
			if it.ArtifactPath != filepath.ToSlash(specRel) {
				t.Errorf("artifact_path no planner inesperado: %q", it.ArtifactPath)
			}
			break
		}
	}
	if !foundPlannerItem {
		t.Errorf("item 'feat-63-01' não foi enfileirado em planner.json: %+v", plannerSQ.Items)
	}

	// 5. Teste de Idempotência: aprovar novamente a mesma spec não causa erro nem duplica
	resp2, err := http.Post(srv.URL()+"/api/spec/approve", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("falha na segunda requisição de aprovação: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200 OK na segunda chamada idempotente, obtido %d", resp2.StatusCode)
	}

	plannerSQ2, _ := queue.LoadStationQueue(tempDir, queue.StationPlanner)
	count := 0
	for _, it := range plannerSQ2.Items {
		if it.ID == "feat-63-01" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("item duplicado em planner.json após aprovação repetida: count=%d", count)
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

func TestAPIHandoff(t *testing.T) {
	t.Run("Success", TestAPIHandoff_Success)
	t.Run("ValidationAndSecurity", TestAPIHandoff_ValidationAndSecurity)
	t.Run("MethodNotAllowed", TestAPIHandoff_MethodNotAllowed)
	t.Run("SSEBroadcast", TestAPIHandoff_SSEBroadcast)
}

func TestAPIHandoff_Success(t *testing.T) {
	srv, tempDir := setupTestStudioServer(t)

	// Prepara arquivo de artefato
	artRel := filepath.Join("sdd", "features", "feat-99", "plan.md")
	artFull := filepath.Join(tempDir, artRel)
	_ = os.MkdirAll(filepath.Dir(artFull), 0755)
	_ = os.WriteFile(artFull, []byte("# Plano de Implementação"), 0644)

	// 1. Inicializa item na estação de origem (planner)
	origItem := queue.StationItem{
		ID:           "task-99-01",
		Title:        "Planejamento Concluído",
		Station:      queue.StationPlanner,
		Operation:    "plan",
		Status:       "review_pending",
		ArtifactPath: filepath.ToSlash(artRel),
	}
	if _, _, err := queue.EnqueueStationItem(tempDir, origItem); err != nil {
		t.Fatalf("falha ao enfileirar na estação de origem: %v", err)
	}

	// 2. Dispara POST /api/handoff de planner -> builder
	payload := studio.HandoffRequest{
		From:         queue.StationPlanner,
		To:           queue.StationBuilder,
		TaskID:       "task-99-01",
		Title:        "Construção da Feature 99",
		ArtifactPath: filepath.ToSlash(artRel),
		Input:        "Plano Aprovado",
		Output:       "Código e Testes",
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(srv.URL()+"/api/handoff", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("falha no POST /api/handoff: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 OK, obtido %d", resp.StatusCode)
	}

	var resData studio.HandoffResponse
	if err := json.NewDecoder(resp.Body).Decode(&resData); err != nil {
		t.Fatalf("falha ao decodificar JSON de resposta: %v", err)
	}

	if resData.Status != "ok" {
		t.Errorf("status inesperado: %q", resData.Status)
	}
	if resData.From != queue.StationPlanner {
		t.Errorf("From inesperado: %q", resData.From)
	}
	if resData.To != queue.StationBuilder {
		t.Errorf("To inesperado: %q", resData.To)
	}
	if resData.ItemID != "task-99-01" {
		t.Errorf("ItemID inesperado: %q", resData.ItemID)
	}

	// 3. Verifica se a fila de destino (builder.json) contém o item enfileirado
	builderSQ, err := queue.LoadStationQueue(tempDir, queue.StationBuilder)
	if err != nil {
		t.Fatalf("falha ao carregar fila do builder: %v", err)
	}
	var foundInBuilder bool
	for _, it := range builderSQ.Items {
		if it.ID == "task-99-01" && it.Station == queue.StationBuilder && it.Status == queue.StatusQueued {
			foundInBuilder = true
			if it.ArtifactPath != filepath.ToSlash(artRel) {
				t.Errorf("ArtifactPath inesperado no builder: %q", it.ArtifactPath)
			}
			break
		}
	}
	if !foundInBuilder {
		t.Errorf("item 'task-99-01' não foi encontrado na fila do builder: %+v", builderSQ.Items)
	}

	// 4. Verifica se a estação de origem (planner.json) teve o item marcado como approved
	plannerSQ, err := queue.LoadStationQueue(tempDir, queue.StationPlanner)
	if err != nil {
		t.Fatalf("falha ao carregar fila do planner: %v", err)
	}
	for _, it := range plannerSQ.Items {
		if it.ID == "task-99-01" && it.Status != queue.StatusApproved {
			t.Errorf("item na estação de origem deveria estar approved, obtido %s", it.Status)
		}
	}

	// 5. Teste de Idempotência: novo POST idêntico não duplica
	resp2, err := http.Post(srv.URL()+"/api/handoff", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("falha no segundo POST /api/handoff: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 na repetição de handoff, obtido %d", resp2.StatusCode)
	}

	builderSQ2, _ := queue.LoadStationQueue(tempDir, queue.StationBuilder)
	count := 0
	for _, it := range builderSQ2.Items {
		if it.ID == "task-99-01" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("item duplicado na fila após chamada idempotente: count=%d", count)
	}
}

func TestAPIHandoff_ValidationAndSecurity(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

	tests := []struct {
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:         "JSON mal formatado",
			payload:      `{bad json`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Estação de origem inválida",
			payload:      `{"from": "alien", "to": "builder", "task_id": "t1"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Estação de destino inválida",
			payload:      `{"from": "pm-architect", "to": "martian", "task_id": "t1"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Ambas estações vazias",
			payload:      `{"from": "", "to": "", "task_id": "t1"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Path traversal no artifact_path",
			payload:      `{"from": "pm-architect", "to": "planner", "task_id": "t1", "artifact_path": "../../etc/passwd"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Caminho absoluto no artifact_path",
			payload:      `{"from": "pm-architect", "to": "planner", "task_id": "t1", "artifact_path": "/var/log/secret"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Caractere nulo no artifact_path",
			payload:      `{"from": "pm-architect", "to": "planner", "task_id": "t1", "artifact_path": "sdd/\u0000bad"}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(srv.URL()+"/api/handoff", "application/json", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatalf("falha ao enviar POST: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("esperado status %d, obtido %d", tt.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestAPIHandoff_MethodNotAllowed(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			req, err := http.NewRequest(m, srv.URL()+"/api/handoff", nil)
			if err != nil {
				t.Fatalf("falha ao criar requisição %s: %v", m, err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("falha na requisição %s: %v", m, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("esperado 405 Method Not Allowed para método %s em /api/handoff, obtido %d", m, resp.StatusCode)
			}
		})
	}
}

func TestAPIHandoff_SSEBroadcast(t *testing.T) {
	srv, _ := setupTestStudioServer(t)

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

	// Consome snapshot inicial
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

	// Canal para capturar broadcast do handoff
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

	// Dispara POST /api/handoff
	handoffPayload := `{"from": "builder", "to": "governance", "task_id": "feat-sse-handoff", "title": "Build Finalizado"}`
	respHandoff, err := http.Post(srv.URL()+"/api/handoff", "application/json", strings.NewReader(handoffPayload))
	if err != nil {
		t.Fatalf("falha ao enviar POST de handoff: %v", err)
	}
	defer respHandoff.Body.Close()

	if respHandoff.StatusCode != http.StatusOK {
		t.Fatalf("esperado status 200 no handoff, obtido %d", respHandoff.StatusCode)
	}

	// Aguarda broadcast SSE
	select {
	case snapJSON := <-broadcastReceived:
		if !strings.Contains(snapJSON, "governance") {
			t.Errorf("snapshot SSE não contém dados da estação de destino: %s", snapJSON)
		}
	case <-time.After(1500 * time.Millisecond):
		t.Fatalf("timeout: broadcast SSE não foi recebido após POST /api/handoff")
	}
}

func TestStudioServer_AutoRunFlow(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Enfileira tarefa no planner com status queued
	taskID := "autorun-feat-100"
	_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:           taskID,
		Title:        "Feature Autônoma Completa",
		Station:      queue.StationPlanner,
		Operation:    "plan",
		Status:       queue.StatusQueued,
		ArtifactPath: "sdd/features/feat-100/plan.md",
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task no planner: %v", err)
	}

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	if err != nil {
		t.Fatalf("falha ao inicializar servidor: %v", err)
	}
	if err := srv.Listen(); err != nil {
		t.Fatalf("falha no listen: %v", err)
	}
	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	// Dispara o TriggerAutoRun na estação Planner
	srv.TriggerAutoRun(queue.StationPlanner)

	// Aguarda a esteira autônoma percorrer Planner -> Builder -> Governance -> Líder
	deadline := time.Now().Add(4 * time.Second)
	var completedAtLeader bool

	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)

		liderSQ, err := queue.LoadStationQueue(tempDir, queue.StationLider)
		if err == nil {
			for _, it := range liderSQ.Items {
				if it.ID == taskID && it.Status == queue.StatusApproved {
					completedAtLeader = true
					break
				}
			}
		}
		if completedAtLeader {
			break
		}
	}

	if !completedAtLeader {
		t.Fatalf("esteira autônoma não concluiu o ciclo até o Líder dentro do prazo esperado")
	}

	// Verifica Governança
	govSQ, err := queue.LoadStationQueue(tempDir, queue.StationGovernance)
	if err != nil {
		t.Fatalf("falha ao carregar fila de governança: %v", err)
	}
	var foundGovApproved bool
	for _, it := range govSQ.Items {
		if it.ID == taskID && it.Status == queue.StatusApproved && it.GovernanceVerdict != nil {
			if it.GovernanceVerdict.Veredito == queue.VerdictConforme {
				foundGovApproved = true
				break
			}
		}
	}
	if !foundGovApproved {
		t.Errorf("tarefa não foi devidamente aprovada com veredito conforme na Governança: %+v", govSQ.Items)
	}

	// Verifica atividades registradas
	acts, err := queue.LoadActivities(tempDir)
	if err != nil {
		t.Fatalf("falha ao carregar atividades: %v", err)
	}
	hasPlannerSuccess := false
	hasBuilderPortaoA := false
	hasGovAudit := false
	for _, a := range acts {
		if a.Station == queue.StationPlanner && a.Status == "success" {
			hasPlannerSuccess = true
		}
		if a.Station == queue.StationBuilder && a.Tool == "portao_a" && a.Status == "success" {
			hasBuilderPortaoA = true
		}
		if a.Station == queue.StationGovernance && a.Tool == "governance" && a.Status == "success" {
			hasGovAudit = true
		}
	}
	if !hasPlannerSuccess {
		t.Errorf("telemetria não registrou atividade de sucesso no Planner")
	}
	if !hasBuilderPortaoA {
		t.Errorf("telemetria não registrou atividade de Portão A no Builder")
	}
	if !hasGovAudit {
		t.Errorf("telemetria não registrou atividade de auditoria conforme na Governança")
	}
}

func TestStudioServer_AutoRunBuilderFailure(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("FORGE_TEST_FAIL_PORTAO_A", "true")
	defer os.Unsetenv("FORGE_TEST_FAIL_PORTAO_A")

	taskID := "builder-fail-task"
	_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:           taskID,
		Title:        "Builder com Teste Quebrado",
		Station:      queue.StationBuilder,
		Operation:    "build",
		Status:       queue.StatusQueued,
		ArtifactPath: "sdd/features/fail/plan.md",
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task: %v", err)
	}

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	if err != nil {
		t.Fatalf("falha ao inicializar servidor: %v", err)
	}
	if err := srv.Listen(); err != nil {
		t.Fatalf("falha no listen: %v", err)
	}
	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	srv.TriggerAutoRun(queue.StationBuilder)

	// Aguarda processamento
	time.Sleep(350 * time.Millisecond)

	builderSQ, err := queue.LoadStationQueue(tempDir, queue.StationBuilder)
	if err != nil {
		t.Fatalf("falha ao ler fila do builder: %v", err)
	}
	var foundChangesRequested bool
	for _, it := range builderSQ.Items {
		if it.ID == taskID && it.Status == queue.StatusChangesRequested {
			foundChangesRequested = true
			break
		}
	}
	if !foundChangesRequested {
		t.Errorf("esperado status changes_requested no builder após falha nos testes, obtido: %+v", builderSQ.Items)
	}

	// Garante que NÃO houve handoff para governança
	govSQ, _ := queue.LoadStationQueue(tempDir, queue.StationGovernance)
	if govSQ != nil {
		for _, it := range govSQ.Items {
			if it.ID == taskID {
				t.Errorf("tarefa com falha no Portão A não deveria ter sido transferida para Governança")
			}
		}
	}
}
