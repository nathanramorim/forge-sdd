package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ActivityItem representa uma ferramenta ou ação executada por uma estação em tempo real.
type ActivityItem struct {
	Timestamp string `json:"timestamp"`
	Station   string `json:"station"`   // lider, pm-architect, planner, builder, governance
	Tool      string `json:"tool"`      // run_command, read_file, edit_file, gate, git, context7, etc.
	Action    string `json:"action"`    // argumento principal / comando / caminho de arquivo
	Details   string `json:"details"`   // descrição legível do objetivo
	Status    string `json:"status"`    // running, success, error
}

var activityMu sync.Mutex

// RecordActivity persiste atomicamente um evento de atividade na fila de logs recentes (máximo 50 itens).
func RecordActivity(dir, station, tool, action, details, status string) error {
	activityMu.Lock()
	defer activityMu.Unlock()

	items, _ := LoadActivities(dir)
	now := time.Now().UTC().Format(time.RFC3339)
	newItem := ActivityItem{
		Timestamp: now,
		Station:   station,
		Tool:      tool,
		Action:    action,
		Details:   details,
		Status:    status,
	}

	// Insere no início (mais recente primeiro)
	newItems := append([]ActivityItem{newItem}, items...)
	if len(newItems) > 50 {
		newItems = newItems[:50]
	}

	actPath := filepath.Join(dir, "sdd", ".queue", "activity.json")
	return WriteJSONAtomic(actPath, newItems, 0o644)
}

// LoadActivities lê os registros recentes de atividade de sdd/.queue/activity.json.
func LoadActivities(dir string) ([]ActivityItem, error) {
	actPath := filepath.Join(dir, "sdd", ".queue", "activity.json")
	data, err := os.ReadFile(actPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []ActivityItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}
