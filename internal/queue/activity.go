package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MaxOutputLength define o limite máximo de tamanho de logs/saída em bytes (8 KB - SEG-01, SEG-05).
const MaxOutputLength = 8192

// ActivityItem representa uma ferramenta ou ação executada por uma estação em tempo real.
type ActivityItem struct {
	ID          string   `json:"id,omitempty"`
	Timestamp   string   `json:"timestamp"`
	Station     string   `json:"station"` // lider, pm-architect, planner, builder, governance
	Tool        string   `json:"tool"`    // run_command, read_file, edit_file, gate, git, context7, etc.
	Action      string   `json:"action"`  // argumento principal / comando / caminho de arquivo
	Details     string   `json:"details"` // descrição legível do objetivo
	Status      string   `json:"status"`  // running, success, error
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	FilePath    string   `json:"filePath,omitempty"`
	Output      string   `json:"output,omitempty"`
	Error       string   `json:"error,omitempty"`
	DurationMs  int64    `json:"durationMs,omitempty"`
	CompletedAt string   `json:"completedAt,omitempty"`
	MascotState string   `json:"mascotState,omitempty"`
}

var activityMu sync.Mutex

func truncateOutput(s string) string {
	if len(s) <= MaxOutputLength {
		return s
	}
	return s[:MaxOutputLength] + "\n... [output truncado em 8KB]"
}

// RecordActivity persiste atomicamente um evento de atividade na fila de logs recentes (máximo 50 itens).
// Mantém 100% de retrocompatibilidade com chamadas existentes.
func RecordActivity(dir, station, tool, action, details, status string) error {
	return RecordActivityDetailed(dir, ActivityItem{
		Station: station,
		Tool:    tool,
		Action:  action,
		Details: details,
		Status:  status,
	})
}

// RecordActivityDetailed persiste atomicamente um evento enriquecido, aplicando reconciliação
// automática em atividades anteriores da mesma estação e truncamento seguro de payloads.
func RecordActivityDetailed(dir string, item ActivityItem) error {
	activityMu.Lock()
	defer activityMu.Unlock()

	items, _ := LoadActivities(dir)
	now := time.Now().UTC().Format(time.RFC3339)

	if item.Timestamp == "" {
		item.Timestamp = now
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("act-%d", time.Now().UnixNano())
	}
	if item.Status == "" {
		item.Status = "running"
	}
	if item.Output != "" {
		item.Output = truncateOutput(item.Output)
	}
	if item.Error != "" {
		item.Error = truncateOutput(item.Error)
	}
	if item.Status != "running" && item.CompletedAt == "" {
		item.CompletedAt = now
	}

	// Reconciliação automática: fecha atividades anteriores em 'running' da mesma estação
	for i := range items {
		if items[i].Station == item.Station && items[i].Status == "running" {
			items[i].Status = "success"
			if items[i].CompletedAt == "" {
				items[i].CompletedAt = now
			}
		}
	}

	// Insere no início (mais recente primeiro)
	newItems := append([]ActivityItem{item}, items...)
	if len(newItems) > 50 {
		newItems = newItems[:50]
	}

	actPath := filepath.Join(dir, "sdd", ".queue", "activity.json")
	return WriteJSONAtomic(actPath, newItems, 0o644)
}

// ReconcileStationActivities fecha todas as atividades em 'running' da estação especificada para 'success'.
func ReconcileStationActivities(dir, station string) error {
	activityMu.Lock()
	defer activityMu.Unlock()

	items, err := LoadActivities(dir)
	if err != nil || len(items) == 0 {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	changed := false
	for i := range items {
		if items[i].Station == station && items[i].Status == "running" {
			items[i].Status = "success"
			if items[i].CompletedAt == "" {
				items[i].CompletedAt = now
			}
			changed = true
		}
	}

	if !changed {
		return nil
	}

	actPath := filepath.Join(dir, "sdd", ".queue", "activity.json")
	return WriteJSONAtomic(actPath, items, 0o644)
}

// SanitizeOrphanActivities varre activity.json e encerra qualquer atividade órfã presa em 'running'.
func SanitizeOrphanActivities(dir string) error {
	activityMu.Lock()
	defer activityMu.Unlock()

	items, err := LoadActivities(dir)
	if err != nil || len(items) == 0 {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	changed := false
	for i := range items {
		if items[i].Status == "running" {
			items[i].Status = "success"
			if items[i].CompletedAt == "" {
				items[i].CompletedAt = now
			}
			changed = true
		}
	}

	if !changed {
		return nil
	}

	actPath := filepath.Join(dir, "sdd", ".queue", "activity.json")
	return WriteJSONAtomic(actPath, items, 0o644)
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
