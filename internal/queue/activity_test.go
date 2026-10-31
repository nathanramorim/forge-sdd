package queue_test

import (
	"testing"

	"github.com/forge-sdd/cli/internal/queue"
)

func TestRecordAndLoadActivity(t *testing.T) {
	tempDir := t.TempDir()

	err := queue.RecordActivity(tempDir, queue.StationLider, "run_command", "git status", "Verificando status", "success")
	if err != nil {
		t.Fatalf("falha ao registrar atividade: %v", err)
	}

	err = queue.RecordActivity(tempDir, queue.StationBuilder, "edit_file", "server.go", "Ajustando porta", "running")
	if err != nil {
		t.Fatalf("falha ao registrar segunda atividade: %v", err)
	}

	items, err := queue.LoadActivities(tempDir)
	if err != nil {
		t.Fatalf("falha ao ler atividades: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("esperava 2 itens, obteve %d", len(items))
	}

	if items[0].Station != queue.StationBuilder || items[0].Tool != "edit_file" {
		t.Errorf("item 0 inesperado: %+v", items[0])
	}

	if items[1].Station != queue.StationLider || items[1].Tool != "run_command" {
		t.Errorf("item 1 inesperado: %+v", items[1])
	}
}
