package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Os exemplos entregues pelo scaffold são exatamente os padrões do código (DUP-04) e nunca ativam o pipeline.
func TestForgeExamples_IguaisAosPadroesENaoAtivamPipeline(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: "4.1.0", Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(dir, ".forge", "config", "forge.json"), "só o pipeline init cria o arquivo ativo")
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "config", "notifications.json"))
	_, found, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.False(t, found)

	var forge config.ForgeConfig
	data, err := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json.example"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &forge))
	assert.Equal(t, config.DefaultForgeConfig(), forge)
	require.NoError(t, forge.Validate())

	var notif config.NotificationsConfig
	data, err = os.ReadFile(filepath.Join(dir, ".forge", "config", "notifications.json.example"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &notif))
	require.NoError(t, notif.Validate())
	assert.True(t, notif.Channels.DiscoverySession.Enabled)
	assert.False(t, notif.Channels.Email.Enabled)
	assert.False(t, notif.Channels.Webhook.Enabled)
}

func TestForgeRuntime_NasceIgnoradoPeloGit(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: "4.1.0", Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)
	testutil.InitRepo(t, dir, "main")
	testutil.WriteFile(t, dir, ".forge/runtime/queues/builder/pending/1-x.json", "{}")
	testutil.WriteFile(t, dir, ".forge/runtime/worktrees/j/t/arquivo.go", "x")
	testutil.Git(t, dir, "add", "-A")
	tracked := testutil.Git(t, dir, "ls-files", ".forge/runtime")
	assert.Equal(t, ".forge/runtime/.gitignore\n", tracked, "só o .gitignore do runtime é versionado")
}

func TestUpdate_NaoSobrescreveForgeJsonAtivo(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, ".forge/config/forge.json", `{"cli":"agy"}`)
	_, err := Run(config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: "4.1.0", Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"cli":"agy"}`, string(data))
}
