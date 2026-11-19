package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
)

// SessionMetrics espelha sdd/.metrics/schema.json.
type SessionMetrics struct {
	Schema           string   `json:"$schema"`
	Feature          string   `json:"feature"`
	Phase            string   `json:"phase,omitempty"`
	AgentPath        []string `json:"agent_path"`
	TokensInput      int      `json:"tokens_input"`
	TokensOutput     int      `json:"tokens_output"`
	Turns            int      `json:"turns"`
	DurationSeconds  int      `json:"duration_seconds"`
	CriterioAtendido bool     `json:"criterio_atendido"`
	Model            string   `json:"model"`
	ReworkLines      int      `json:"rework_lines"`
	Context7Calls    int      `json:"context7_calls"`
	GitCalls         int      `json:"git_calls"`
	FilesTouched     []string `json:"files_touched"`
	Outcome          string   `json:"outcome"`
	SddVersion       string   `json:"sdd_version"`

	TokensCacheRead     int     `json:"tokens_cache_read,omitempty"`
	TokensCacheCreation int     `json:"tokens_cache_creation,omitempty"`
	CostUSDTabela       float64 `json:"cost_usd_tabela,omitempty"`
	Medicao             string  `json:"medicao,omitempty"`

	Complexidade *string `json:"complexidade"` // reservado; preenchimento definido por ADR futura
}

// RecordSessionMetrics grava um arquivo de telemetria em sdd/.metrics/session-<ISO8601>.json.
func RecordSessionMetrics(targetDir string, metrics SessionMetrics, at time.Time) (string, error) {
	metricsDir := layout.Metrics(targetDir)
	if err := os.MkdirAll(metricsDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar sdd/.metrics: %w", err)
	}

	base := fmt.Sprintf("session-%s", at.Format("2006-01-02T150405Z"))
	filename := base + ".json"
	path := filepath.Join(metricsDir, filename)
	for i := 2; pathExists(path); i++ {
		filename = fmt.Sprintf("%s-%d.json", base, i)
		path = filepath.Join(metricsDir, filename)
	}

	if metrics.Schema == "" {
		metrics.Schema = "forge-sdd/metrics/1.0"
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("falha ao serializar telemetria: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
	return path, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
