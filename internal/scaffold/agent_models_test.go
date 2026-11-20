package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scaffoldClaude(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	_, err := Run(pipelineCfg(config.AgentClaude), dir)
	require.NoError(t, err)
	return dir
}

func writeForge(t *testing.T, dir string, models map[string]string) {
	t.Helper()
	cfg := config.DefaultForgeConfig()
	cfg.AgentModels = models
	require.NoError(t, config.SaveForgeConfig(dir, cfg))
}

func headerModel(t *testing.T, dir, agent string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "agents", agent+".md"))
	require.NoError(t, err)
	for _, l := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(l, "model:") {
			return strings.TrimSpace(strings.TrimPrefix(l, "model:"))
		}
	}
	t.Fatalf("sem model: em %s", agent)
	return ""
}

func TestAgentModelsEscreveCabecalho(t *testing.T) {
	dir := scaffoldClaude(t)
	writeForge(t, dir, map[string]string{"builder": "haiku", "reviewer": "claude-sonnet-4-5"})

	changed, err := applyAgentModels(dir)
	require.NoError(t, err)
	assert.Len(t, changed, 3) // builder, reviewer, governance

	assert.Equal(t, "haiku", headerModel(t, dir, "builder"))
	assert.Equal(t, "claude-sonnet-4-5", headerModel(t, dir, "reviewer"))
	assert.Equal(t, "claude-sonnet-4-5", headerModel(t, dir, "governance"), "subagente herda o papel")
	assert.Equal(t, "inherit", headerModel(t, dir, "planner"), "papel sem chave segue inherit")
	assert.Equal(t, "inherit", headerModel(t, dir, "lider"))

	// idempotente
	changed, err = applyAgentModels(dir)
	require.NoError(t, err)
	assert.Empty(t, changed)
}

func TestAgentModelsSemChaveNaoMuda(t *testing.T) {
	dir := scaffoldClaude(t)
	before, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "builder.md"))

	changed, err := applyAgentModels(dir) // sem forge.json
	require.NoError(t, err)
	assert.Empty(t, changed)

	writeForge(t, dir, nil) // forge.json sem agent_models
	changed, err = applyAgentModels(dir)
	require.NoError(t, err)
	assert.Empty(t, changed)
	after, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "builder.md"))
	assert.Equal(t, string(before), string(after))
	assert.Equal(t, "inherit", headerModel(t, dir, "builder"))
}

func TestAgentModelsRecusaValorInvalido(t *testing.T) {
	for name, v := range map[string]string{
		"vazio":     "",
		"controle":  "haiku\nmodel: x",
		"espaco":    "opus 4",
		"longo":     strings.Repeat("a", 81),
		"separador": "a;b",
	} {
		t.Run(name, func(t *testing.T) {
			cfg := config.DefaultForgeConfig()
			cfg.AgentModels = map[string]string{"builder": v}
			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "agent_models.builder")
		})
	}

	t.Run("papel desconhecido", func(t *testing.T) {
		cfg := config.DefaultForgeConfig()
		cfg.AgentModels = map[string]string{"lider": "haiku"}
		require.Error(t, cfg.Validate())
	})

	t.Run("nenhum arquivo é gravado", func(t *testing.T) {
		dir := scaffoldClaude(t)
		before, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "builder.md"))
		cfg := config.DefaultForgeConfig()
		cfg.AgentModels = map[string]string{"builder": "haiku\nx"}
		require.Error(t, cfg.Validate())
		raw := `{"cli":"claude","agent_models":{"builder":"haiku\nx"}}`
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), []byte(raw), 0o644))
		_, err := applyAgentModels(dir)
		require.Error(t, err)
		after, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "builder.md"))
		assert.Equal(t, string(before), string(after))
	})

	t.Run("teto de 420 bytes", func(t *testing.T) {
		dir := scaffoldClaude(t)
		writeForge(t, dir, map[string]string{"reviewer": strings.Repeat("m", 80)})
		_, err := applyAgentModels(dir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "420")
		assert.Equal(t, "inherit", headerModel(t, dir, "reviewer"), "sem gravação parcial")
	})
}

func TestAgentModelsDryRunNaoGrava(t *testing.T) {
	dir := scaffoldClaude(t)
	writeForge(t, dir, map[string]string{"builder": "haiku"})
	cfg := pipelineCfg(config.AgentClaude)
	cfg.DryRun = true
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	assert.Equal(t, "inherit", headerModel(t, dir, "builder"))
}

func TestAgentModelsAplicadoPeloRun(t *testing.T) {
	dir := scaffoldClaude(t)
	writeForge(t, dir, map[string]string{"discovery": "opus"})
	_, err := Run(pipelineCfg(config.AgentClaude), dir) // forge update re-renderiza e reaplica
	require.NoError(t, err)
	assert.Equal(t, "opus", headerModel(t, dir, "pm-architect"))
	assert.Equal(t, "opus", headerModel(t, dir, "planner"))
}

func TestAgentModelsRecusaEscalarYAML(t *testing.T) {
	for _, v := range []string{"a:", "-", "-1", "1.5", "0x1F", "9lives", "a-", "null", "NULL", "True", "false", "yes", "No", "on", "OFF", "y", "N", "~", ".x", "_x"} {
		cfg := config.DefaultForgeConfig()
		cfg.AgentModels = map[string]string{"builder": v}
		err := cfg.Validate()
		require.Error(t, err, v)
		assert.Contains(t, err.Error(), "agent_models.builder", v)
	}
}

func TestAgentModelsAceitaNomesReais(t *testing.T) {
	for _, v := range []string{"inherit", "sonnet", "opus", "haiku", "claude-sonnet-4-5", "claude-opus-4-1-20250805", "gpt-5.1", "gemini-2.5-pro", "us.anthropic.claude-sonnet-4:0"} {
		cfg := config.DefaultForgeConfig()
		cfg.AgentModels = map[string]string{"builder": v}
		assert.NoError(t, cfg.Validate(), v)
	}
}
