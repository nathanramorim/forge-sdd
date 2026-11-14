package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
)

// RoleTokens consolida a contabilidade de tokens acumulados de um papel.
type RoleTokens struct {
	Role         string `json:"role"`
	TokensInput  int    `json:"tokens_input"`
	TokensOutput int    `json:"tokens_output"`
	TotalTokens  int    `json:"total_tokens"`
}

// ProjectTokenMetrics agrega o total de tokens do projeto e por papel.
type ProjectTokenMetrics struct {
	Roles       map[string]RoleTokens `json:"roles"`
	TotalInput  int                   `json:"total_input"`
	TotalOutput int                   `json:"total_output"`
	TotalTokens int                   `json:"total_tokens"`
}

type sessionRecord struct {
	Feature      string   `json:"feature"`
	AgentPath    []string `json:"agent_path"`
	TokensInput  int      `json:"tokens_input"`
	TokensOutput int      `json:"tokens_output"`
}

// NormalizeRole normaliza nomes e apelidos de papéis para as três chaves canônicas:
// discovery, builder e reviewer.
func NormalizeRole(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(s, "builder"), strings.Contains(s, "bld"):
		return layout.RoleBuilder
	case strings.Contains(s, "reviewer"), strings.Contains(s, "revisor"), strings.Contains(s, "governanca"), strings.Contains(s, "governance"), strings.Contains(s, "rev"):
		return layout.RoleReviewer
	case strings.Contains(s, "discovery"), strings.Contains(s, "lider"), strings.Contains(s, "planner"), strings.Contains(s, "specifier"), strings.Contains(s, "orquestrador"), strings.Contains(s, "disc"):
		return layout.RoleDiscovery
	default:
		return ""
	}
}

// TokenMetricsPerRole lê todos os registros de telemetria da sessão em sdd/.metrics/
// e calcula o total acumulado de tokens (in, out, total) para cada papel (CA-01).
func TokenMetricsPerRole(dir string) (map[string]RoleTokens, error) {
	metrics, err := CalculateTokenMetrics(dir)
	if err != nil {
		return nil, err
	}
	return metrics.Roles, nil
}

// CalculateTokenMetrics lê as métricas de telemetria em sdd/.metrics e consolida os totais.
func CalculateTokenMetrics(dir string) (*ProjectTokenMetrics, error) {
	rolesMap := map[string]RoleTokens{
		layout.RoleDiscovery: {Role: layout.RoleDiscovery},
		layout.RoleBuilder:   {Role: layout.RoleBuilder},
		layout.RoleReviewer:  {Role: layout.RoleReviewer},
	}

	res := &ProjectTokenMetrics{
		Roles: rolesMap,
	}

	metricsDir := layout.Metrics(dir)
	entries, err := os.ReadDir(metricsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return res, nil
		}
		return nil, fmt.Errorf("falha ao ler diretório de métricas %s: %w", metricsDir, err)
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "session-") || !strings.HasSuffix(name, ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(metricsDir, name))
		if err != nil {
			continue
		}

		var rec sessionRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}

		role := inferRoleFromSession(rec)
		if role == "" {
			continue
		}

		current := res.Roles[role]
		current.TokensInput += rec.TokensInput
		current.TokensOutput += rec.TokensOutput
		current.TotalTokens += (rec.TokensInput + rec.TokensOutput)
		res.Roles[role] = current

		res.TotalInput += rec.TokensInput
		res.TotalOutput += rec.TokensOutput
		res.TotalTokens += (rec.TokensInput + rec.TokensOutput)
	}

	return res, nil
}

func inferRoleFromSession(rec sessionRecord) string {
	for _, p := range rec.AgentPath {
		if r := NormalizeRole(p); r != "" {
			return r
		}
	}

	f := strings.ToLower(rec.Feature)
	if strings.Contains(f, "discovery") {
		return layout.RoleDiscovery
	}
	if strings.Contains(f, "criteria") || strings.Contains(f, "features") || strings.Contains(f, "fix") {
		return layout.RoleBuilder
	}

	return ""
}

// FormatTokenBadge formata o badge de tokens para as bordas dos painéis do tmux.
// Ex: "🪙 12k tok" ou "🪙 78k tok ⚠️" quando próximo do limiar de 80k.
func FormatTokenBadge(tokens int) string {
	var k float64
	if tokens < 1000 && tokens > 0 {
		return fmt.Sprintf("🪙 %.1fk tok", float64(tokens)/1000.0)
	}
	k = float64(tokens) / 1000.0

	// Alerta visual quando atinge ou se aproxima de 75k-80k tokens
	if tokens >= 75000 {
		return fmt.Sprintf("🪙 %.0fk tok ⚠️", k)
	}
	return fmt.Sprintf("🪙 %.0fk tok", k)
}

// FormatTokensShort formata quantidade de tokens de forma compacta (ex: "12k", "0.8k").
func FormatTokensShort(tokens int) string {
	if tokens <= 0 {
		return "0k"
	}
	if tokens < 1000 {
		return fmt.Sprintf("%.1fk", float64(tokens)/1000.0)
	}
	return fmt.Sprintf("%.0fk", float64(tokens)/1000.0)
}
