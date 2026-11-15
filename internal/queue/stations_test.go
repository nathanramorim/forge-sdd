package queue_test

import (
	"sync"
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
)

func TestStationQueuesAtomic(t *testing.T) {
	tempDir := t.TempDir()

	// Testa carregamento de estação inexistente
	sq, err := queue.LoadStationQueue(tempDir, queue.StationLider)
	if err != nil {
		t.Fatalf("erro ao carregar fila inexistente: %v", err)
	}
	if sq.Station != queue.StationLider {
		t.Errorf("esperado station lider, recebido %s", sq.Station)
	}
	if sq.Version != 0 {
		t.Errorf("esperado version 0 inicial, recebido %d", sq.Version)
	}

	// Escrita concorrente atômica
	var wg sync.WaitGroup
	workers := 10
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			item := queue.StationItem{
				ID:        queue.StationLider + "-task",
				Title:     "Demanda do Líder",
				Station:   queue.StationLider,
				Operation: "plan",
				Status:    queue.StatusBuilding,
				Attempt:   workerID,
			}
			_, _, _ = queue.EnqueueStationItem(tempDir, item)
		}(i)
	}
	wg.Wait()

	finalSQ, err := queue.LoadStationQueue(tempDir, queue.StationLider)
	if err != nil {
		t.Fatalf("erro ao recarregar fila: %v", err)
	}
	if finalSQ.Version == 0 {
		t.Errorf("esperado incremento monotônico de version, recebido %d", finalSQ.Version)
	}
	if len(finalSQ.Items) != 1 {
		t.Errorf("esperado exatamente 1 item atualizado sem duplicatas, recebido %d", len(finalSQ.Items))
	}
}

func TestIdempotencyKey(t *testing.T) {
	tempDir := t.TempDir()

	item := queue.StationItem{
		ID:           "feat-59-01",
		Title:        "Motor de Filas",
		Station:      queue.StationBuilder,
		Operation:    "build",
		ArtifactPath: "sdd/features/feat-59/task.md",
		Attempt:      1,
	}

	// Primeiro enqueue
	item1, hit1, err := queue.EnqueueStationItem(tempDir, item)
	if err != nil {
		t.Fatalf("erro no 1º enqueue: %v", err)
	}
	if hit1 {
		t.Errorf("1º enqueue não deveria ser idempotent hit")
	}

	// Segundo enqueue com mesma chamada
	item2, hit2, err := queue.EnqueueStationItem(tempDir, item)
	if err != nil {
		t.Fatalf("erro no 2º enqueue: %v", err)
	}
	if !hit2 {
		t.Errorf("2º enqueue DEVERIA ser idempotent hit")
	}
	if item1.IdempotencyKey != item2.IdempotencyKey {
		t.Errorf("chaves de idempotência divergem: %s vs %s", item1.IdempotencyKey, item2.IdempotencyKey)
	}

	// Verifica persistência das 5 estações
	stations := queue.CanonicalStations()
	if len(stations) != 5 {
		t.Fatalf("esperado 5 estações canônicas, recebido %d", len(stations))
	}
}

func TestHandoffBetweenStations(t *testing.T) {
	tempDir := t.TempDir()

	item, hit, err := queue.Handoff(tempDir, queue.StationLider, queue.StationPMArchitect, "demand-01", "Demanda Estruturada", "discovery.md", "Entrada Humano", "Saida Discovery", "")
	if err != nil {
		t.Fatalf("erro no handoff: %v", err)
	}
	if hit {
		t.Errorf("primeiro handoff não deveria ser hit idempotente")
	}
	if item.Station != queue.StationPMArchitect {
		t.Errorf("esperado destino pm-architect, recebido %s", item.Station)
	}

	// Handoff repetido deve acusar hit idempotente
	itemRepeated, hitRepeated, err := queue.Handoff(tempDir, queue.StationLider, queue.StationPMArchitect, "demand-01", "Demanda Estruturada", "discovery.md", "Entrada Humano", "Saida Discovery", "")
	if err != nil {
		t.Fatalf("erro no handoff repetido: %v", err)
	}
	if !hitRepeated {
		t.Errorf("handoff repetido deveria acusar hit idempotente")
	}
	if item.IdempotencyKey != itemRepeated.IdempotencyKey {
		t.Errorf("chaves divergem: %s vs %s", item.IdempotencyKey, itemRepeated.IdempotencyKey)
	}
}

func TestSyncLegacyQueue(t *testing.T) {
	tempDir := t.TempDir()

	legacyQ := &queue.Queue{
		FeatureID: "feat-59",
		Items: []queue.QueueItem{
			{
				TaskID:   "task-01",
				Title:    "Scaffold",
				SpecPath: "spec.md",
				Status:   queue.StatusBuilding,
				Attempts: 1,
			},
		},
	}

	if err := queue.Save(tempDir, legacyQ); err != nil {
		t.Fatalf("erro ao salvar fila legado: %v", err)
	}

	// Verifica se builder.json foi sincronizado automaticamente
	builderSQ, err := queue.LoadStationQueue(tempDir, queue.StationBuilder)
	if err != nil {
		t.Fatalf("erro ao carregar builder.json: %v", err)
	}
	if len(builderSQ.Items) != 1 {
		t.Fatalf("esperado 1 item sincronizado em builder.json, recebido %d", len(builderSQ.Items))
	}
	if builderSQ.Items[0].ID != "task-01" {
		t.Errorf("esperado ID task-01, recebido %s", builderSQ.Items[0].ID)
	}
}

func TestUpdateStationItemStatus(t *testing.T) {
	tempDir := t.TempDir()

	item := queue.StationItem{
		ID:        "builder-task-01",
		Title:     "Implementar recurso",
		Station:   queue.StationBuilder,
		Operation: "build",
		Status:    queue.StatusQueued,
	}

	_, _, err := queue.EnqueueStationItem(tempDir, item)
	if err != nil {
		t.Fatalf("erro ao enfileirar: %v", err)
	}

	// Atualiza para building
	updated, err := queue.UpdateStationItemStatus(tempDir, queue.StationBuilder, "builder-task-01", queue.StatusBuilding)
	if err != nil {
		t.Fatalf("erro ao atualizar status: %v", err)
	}
	if updated.Status != queue.StatusBuilding {
		t.Errorf("esperado status building, recebido %s", updated.Status)
	}

	// Atualiza para review_pending
	updated2, err := queue.UpdateStationItemStatus(tempDir, queue.StationBuilder, "builder-task-01", queue.StatusReviewPending)
	if err != nil {
		t.Fatalf("erro ao atualizar status 2: %v", err)
	}
	if updated2.Status != queue.StatusReviewPending {
		t.Errorf("esperado status review_pending, recebido %s", updated2.Status)
	}

	// Testa erro em item inexistente
	_, err = queue.UpdateStationItemStatus(tempDir, queue.StationBuilder, "non-existent", queue.StatusApproved)
	if err == nil {
		t.Errorf("esperado erro para item inexistente")
	}

	// Testa erro em estação inválida
	_, err = queue.UpdateStationItemStatus(tempDir, "invalid-station", "builder-task-01", queue.StatusApproved)
	if err == nil {
		t.Errorf("esperado erro para estação inválida")
	}
}
