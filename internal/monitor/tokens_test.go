package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenMetricsPerRole(t *testing.T) {
	tempDir := t.TempDir()
	metricsDir := filepath.Join(tempDir, "sdd", ".metrics")
	require.NoError(t, os.MkdirAll(metricsDir, 0o755))

	sessions := []struct {
		filename string
		role     string
		in       int
		out      int
	}{
		{"session-2026-11-14T100000Z.json", "discovery", 10000, 2000},
		{"session-2026-11-14T103000Z.json", "builder", 60000, 18000},
		{"session-2026-11-14T110000Z.json", "reviewer", 22000, 6000},
	}

	for _, s := range sessions {
		rec := map[string]interface{}{
			"$schema":       "forge-sdd/metrics/1.0",
			"feature":       fmt.Sprintf("sdd/%s/spec.md", s.role),
			"agent_path":    []string{s.role},
			"tokens_input":  s.in,
			"tokens_output": s.out,
		}
		data, err := json.Marshal(rec)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(metricsDir, s.filename), data, 0o644))
	}

	// Executa TokenMetricsPerRole conforme CA-01
	roles, err := TokenMetricsPerRole(tempDir)
	require.NoError(t, err)
	require.NotNil(t, roles)

	// Valida discovery
	disc := roles[layout.RoleDiscovery]
	assert.Equal(t, 10000, disc.TokensInput)
	assert.Equal(t, 2000, disc.TokensOutput)
	assert.Equal(t, 12000, disc.TotalTokens)

	// Valida builder
	bld := roles[layout.RoleBuilder]
	assert.Equal(t, 60000, bld.TokensInput)
	assert.Equal(t, 18000, bld.TokensOutput)
	assert.Equal(t, 78000, bld.TotalTokens)

	// Valida reviewer
	rev := roles[layout.RoleReviewer]
	assert.Equal(t, 22000, rev.TokensInput)
	assert.Equal(t, 6000, rev.TokensOutput)
	assert.Equal(t, 28000, rev.TotalTokens)

	// Valida CalculateTokenMetrics
	projMetrics, err := CalculateTokenMetrics(tempDir)
	require.NoError(t, err)
	assert.Equal(t, 92000, projMetrics.TotalInput)
	assert.Equal(t, 26000, projMetrics.TotalOutput)
	assert.Equal(t, 118000, projMetrics.TotalTokens)

	// Valida badges
	assert.Equal(t, "🪙 12k tok", FormatTokenBadge(12000))
	assert.Equal(t, "🪙 78k tok ⚠️", FormatTokenBadge(78000))
	assert.Equal(t, "🪙 0.8k tok", FormatTokenBadge(800))
	assert.Equal(t, "🪙 0k tok", FormatTokenBadge(0))
}
