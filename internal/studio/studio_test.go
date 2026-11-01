package studio_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/studio"
)

func TestStudioServer_StateAndReadOnly(t *testing.T) {
	tempDir := t.TempDir()

	// Cria itens em lider e builder
	_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "item-lider",
		Title:     "Demanda do Líder",
		Station:   queue.StationLider,
		Operation: "conception",
		Status:    queue.StatusBuilding,
	})
	if err != nil {
		t.Fatalf("erro ao criar item no lider: %v", err)
	}

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0, // porta dinâmica
	})
	if err != nil {
		t.Fatalf("erro ao criar studio server: %v", err)
	}

	if err := srv.Listen(); err != nil {
		t.Fatalf("erro ao abrir porta: %v", err)
	}

	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	url := srv.URL()

	// 1. Testa GET /api/state
	resp, err := http.Get(url + "/api/state")
	if err != nil {
		t.Fatalf("falha ao requisitar /api/state: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado status 200, recebido %d", resp.StatusCode)
	}

	var snap studio.StudioSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("falha ao decodificar snapshot: %v", err)
	}

	if len(snap.Stations) != 5 {
		t.Fatalf("esperado 5 estações no snapshot, recebido %d", len(snap.Stations))
	}
	if snap.ActiveStation != queue.StationLider {
		t.Errorf("esperado active station lider, recebido %s", snap.ActiveStation)
	}

	// 2. Testa Rejeição de Métodos de Mutação (POST, PUT, DELETE) -> 405 Method Not Allowed
	mutations := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, m := range mutations {
		req, _ := http.NewRequest(m, url+"/api/state", strings.NewReader(`{"bad":"mutation"}`))
		mResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("erro na requisição de mutação %s: %v", m, err)
		}
		if mResp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("esperado 405 Method Not Allowed para %s, recebido %d", m, mResp.StatusCode)
		}
		mResp.Body.Close()
	}
}

func TestStudioServer_SSE(t *testing.T) {
	tempDir := t.TempDir()

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	if err != nil {
		t.Fatalf("erro ao criar studio: %v", err)
	}

	if err := srv.Listen(); err != nil {
		t.Fatalf("erro ao abrir porta: %v", err)
	}

	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL()+"/api/events", nil)
	if err != nil {
		t.Fatalf("erro ao criar req SSE: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("falha ao conectar no SSE: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado status 200 no SSE, recebido %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("esperado Content-Type text/event-stream, recebido %s", ct)
	}
}

func TestStudioState_PlannerQueuedItemActivation(t *testing.T) {
	tempDir := t.TempDir()

	// Enfileira item no Planejador com status queued (vindo de handoff do PM)
	_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "task-plan-01",
		Title:     "Planejamento da Feature 63",
		Station:   queue.StationPlanner,
		Operation: "planning",
		Status:    queue.StatusQueued,
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar item no planejador: %v", err)
	}

	agg := studio.NewAggregator(tempDir)
	snap, err := agg.GetSnapshot()
	if err != nil {
		t.Fatalf("falha ao obter snapshot: %v", err)
	}

	if snap.ActiveStation != queue.StationPlanner {
		t.Errorf("esperado ActiveStation %s, obtido %s", queue.StationPlanner, snap.ActiveStation)
	}

	var plannerStation *studio.StationDTO
	for i := range snap.Stations {
		if snap.Stations[i].ID == queue.StationPlanner {
			plannerStation = &snap.Stations[i]
			break
		}
	}
	if plannerStation == nil {
		t.Fatalf("estação planner não encontrada no snapshot")
	}

	if plannerStation.MascotState != "thinking" {
		t.Errorf("esperado MascotState 'thinking', obtido %s", plannerStation.MascotState)
	}
	if !strings.HasPrefix(plannerStation.CurrentFocus, "Fila:") {
		t.Errorf("esperado CurrentFocus iniciando com 'Fila:', obtido %s", plannerStation.CurrentFocus)
	}
	if plannerStation.ExecutingTask == nil {
		t.Fatalf("esperado ExecutingTask não nulo para o Planejador")
	}
	if plannerStation.ExecutingTask.ID != "task-plan-01" {
		t.Errorf("esperado ExecutingTask.ID 'task-plan-01', obtido %s", plannerStation.ExecutingTask.ID)
	}
	if plannerStation.ActiveItems != 1 {
		t.Errorf("esperado ActiveItems 1, obtido %d", plannerStation.ActiveItems)
	}
}

func TestStudioState_GovernanceQueueItemActivation(t *testing.T) {
	statuses := []string{queue.StatusReviewPending, "in_progress", "running"}

	for _, st := range statuses {
		t.Run("status_"+st, func(t *testing.T) {
			tempDir := t.TempDir()

			_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
				ID:        "gov-audit-01",
				Title:     "Auditoria de Conformidade ARQ-01",
				Station:   queue.StationGovernance,
				Operation: "audit",
				Status:    st,
			})
			if err != nil {
				t.Fatalf("erro ao enfileirar item na governança: %v", err)
			}

			agg := studio.NewAggregator(tempDir)
			snap, err := agg.GetSnapshot()
			if err != nil {
				t.Fatalf("falha ao obter snapshot: %v", err)
			}

			if snap.ActiveStation != queue.StationGovernance {
				t.Errorf("esperado ActiveStation %s, obtido %s", queue.StationGovernance, snap.ActiveStation)
			}

			var govStation *studio.StationDTO
			for i := range snap.Stations {
				if snap.Stations[i].ID == queue.StationGovernance {
					govStation = &snap.Stations[i]
					break
				}
			}
			if govStation == nil {
				t.Fatalf("estação governance não encontrada")
			}

			if govStation.MascotState != "approval" {
				t.Errorf("esperado MascotState 'approval', obtido %s", govStation.MascotState)
			}
			expectedFocus := "Auditando Conformidade: Auditoria de Conformidade ARQ-01"
			if govStation.CurrentFocus != expectedFocus {
				t.Errorf("esperado CurrentFocus %q, obtido %q", expectedFocus, govStation.CurrentFocus)
			}
			if govStation.ExecutingTask == nil {
				t.Fatalf("esperado ExecutingTask preenchido")
			}
			if govStation.ExecutingTask.ID != "gov-audit-01" {
				t.Errorf("esperado ExecutingTask.ID 'gov-audit-01', obtido %s", govStation.ExecutingTask.ID)
			}
			if govStation.ExecutingTask.Status != "review_pending" {
				t.Errorf("esperado ExecutingTask.Status 'review_pending', obtido %s", govStation.ExecutingTask.Status)
			}
			if govStation.ExecutingTask.Operation != "audit" {
				t.Errorf("esperado ExecutingTask.Operation 'audit', obtido %s", govStation.ExecutingTask.Operation)
			}
		})
	}
}

func TestStudioState_GovernanceRecentActivityActivation(t *testing.T) {
	for _, actStatus := range []string{"running", "working"} {
		t.Run("status_"+actStatus, func(t *testing.T) {
			tempDir := t.TempDir()

			err := queue.RecordActivity(tempDir, queue.StationGovernance, "read_file", "internal/studio/state.go", "Auditoria Camada B", actStatus)
			if err != nil {
				t.Fatalf("erro ao registrar atividade: %v", err)
			}

			agg := studio.NewAggregator(tempDir)
			snap, err := agg.GetSnapshot()
			if err != nil {
				t.Fatalf("falha ao obter snapshot: %v", err)
			}

			if snap.ActiveStation != queue.StationGovernance {
				t.Errorf("esperado ActiveStation %s, obtido %s", queue.StationGovernance, snap.ActiveStation)
			}

			var govStation *studio.StationDTO
			for i := range snap.Stations {
				if snap.Stations[i].ID == queue.StationGovernance {
					govStation = &snap.Stations[i]
					break
				}
			}
			if govStation == nil {
				t.Fatalf("estação governance não encontrada")
			}

			if govStation.MascotState != "approval" {
				t.Errorf("esperado MascotState 'approval', obtido %s", govStation.MascotState)
			}
			expectedFocus := "Auditoria Camada B em andamento"
			if govStation.CurrentFocus != expectedFocus {
				t.Errorf("esperado CurrentFocus %q, obtido %q", expectedFocus, govStation.CurrentFocus)
			}
		})
	}
}

func TestStudioServer_TaskExecute(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Enfileira tarefas de teste
	_, _, err := queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "builder-task-01",
		Title:     "Recurso Builder",
		Station:   queue.StationBuilder,
		Operation: "build",
		Status:    queue.StatusQueued,
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task no builder: %v", err)
	}

	_, _, err = queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "pm-task-01",
		Title:     "Spec PM",
		Station:   queue.StationPMArchitect,
		Operation: "spec",
		Status:    queue.StatusQueued,
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task no pm-architect: %v", err)
	}

	_, _, err = queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "planner-task-01",
		Title:     "Plan Decomposition",
		Station:   queue.StationPlanner,
		Operation: "plan",
		Status:    queue.StatusQueued,
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task no planner: %v", err)
	}

	_, _, err = queue.EnqueueStationItem(tempDir, queue.StationItem{
		ID:        "gov-task-01",
		Title:     "Governance Review",
		Station:   queue.StationGovernance,
		Operation: "review",
		Status:    queue.StatusQueued,
	})
	if err != nil {
		t.Fatalf("erro ao enfileirar task no governance: %v", err)
	}

	srv, err := studio.NewServer(studio.Config{
		Dir:  tempDir,
		Host: "127.0.0.1",
		Port: 0,
	})
	if err != nil {
		t.Fatalf("erro ao inicializar servidor: %v", err)
	}
	if err := srv.Listen(); err != nil {
		t.Fatalf("erro ao abrir porta: %v", err)
	}
	go func() {
		_ = srv.Start()
	}()
	defer func() {
		_ = srv.Stop()
	}()

	url := srv.URL()

	// 2. Testa validações de entrada
	t.Run("invalid_station", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "invalid", "task_id": "builder-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("esperado status 400, recebido %d", resp.StatusCode)
		}
	})

	t.Run("missing_task_id", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "builder", "task_id": ""}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("esperado status 400, recebido %d", resp.StatusCode)
		}
	})

	t.Run("task_not_found", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "builder", "task_id": "non-existent"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("esperado status 404, recebido %d", resp.StatusCode)
		}
	})

	// 3. Testa execução com sucesso no Builder (Portão A)
	t.Run("builder_execution_success", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "builder", "task_id": "builder-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperado status 200, recebido %d", resp.StatusCode)
		}

		var execResp studio.TaskExecuteResponse
		if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
			t.Fatalf("falha ao decodificar resposta: %v", err)
		}
		if execResp.Status != queue.StatusReviewPending {
			t.Errorf("esperado status review_pending, recebido %s", execResp.Status)
		}

		// Verifica na fila da estação
		sq, err := queue.LoadStationQueue(tempDir, queue.StationBuilder)
		if err != nil {
			t.Fatalf("erro ao carregar fila: %v", err)
		}
		if len(sq.Items) == 0 || sq.Items[0].Status != queue.StatusReviewPending {
			t.Errorf("esperado item status review_pending no builder.json, recebido %v", sq.Items)
		}

		// Verifica atividades registradas
		acts, err := queue.LoadActivities(tempDir)
		if err != nil {
			t.Fatalf("erro ao carregar atividades: %v", err)
		}
		hasPortaoA := false
		for _, a := range acts {
			if a.Tool == "portao_a" && a.Status == "success" {
				hasPortaoA = true
				break
			}
		}
		if !hasPortaoA {
			t.Errorf("esperado registro de atividade de sucesso para portao_a")
		}
	})

	// 4. Testa execução nas outras estações (pm-architect, planner, governance)
	t.Run("pm_architect_execution", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "pm-architect", "task_id": "pm-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperado status 200, recebido %d", resp.StatusCode)
		}
	})

	t.Run("planner_execution", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "planner", "task_id": "planner-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperado status 200, recebido %d", resp.StatusCode)
		}
	})

	t.Run("governance_execution", func(t *testing.T) {
		body := bytes.NewBufferString(`{"station": "governance", "task_id": "gov-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("esperado status 200, recebido %d", resp.StatusCode)
		}

		// Verifica veredito
		sq, _ := queue.LoadStationQueue(tempDir, queue.StationGovernance)
		if len(sq.Items) == 0 || sq.Items[0].GovernanceVerdict == nil {
			t.Errorf("esperado GovernanceVerdict populado na conclusão da governança")
		}
	})

	// 5. Testa simulação de erro no Builder
	t.Run("builder_execution_failure", func(t *testing.T) {
		os.Setenv("FORGE_TEST_FAIL_PORTAO_A", "true")
		defer os.Unsetenv("FORGE_TEST_FAIL_PORTAO_A")

		body := bytes.NewBufferString(`{"station": "builder", "task_id": "builder-task-01"}`)
		resp, err := http.Post(url+"/api/task/execute", "application/json", body)
		if err != nil {
			t.Fatalf("erro na requisição: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("esperado status 400 em falha simulada, recebido %d", resp.StatusCode)
		}
	})
}
