package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineInit_CriaSemSobrescreverEDryRunNaoGrava(t *testing.T) {
	dir := t.TempDir()
	_ = pipelineInitCmd.Flags().Set("dry-run", "false")

	out, err := runRoot(t, "pipeline", "init", "--dir", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "[DRY]")
	assert.NoDirExists(t, filepath.Join(dir, ".forge"))

	_ = pipelineInitCmd.Flags().Set("dry-run", "false")
	out, err = runRoot(t, "pipeline", "init", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "criado    .forge/config/forge.json")
	assert.FileExists(t, filepath.Join(dir, ".forge", "runtime", ".gitignore"))

	cfg, found, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, found, "init ativa o modo pipeline")
	assert.Equal(t, config.DefaultForgeConfig(), cfg)
	_, err = config.LoadNotificationsConfig(dir)
	require.NoError(t, err)

	// Edição do usuário sobrevive à segunda execução.
	p := filepath.Join(dir, ".forge", "config", "forge.json")
	require.NoError(t, os.WriteFile(p, []byte(`{"cli":"agy"}`), 0o644))
	out, err = runRoot(t, "pipeline", "init", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "mantido   .forge/config/forge.json")
	data, _ := os.ReadFile(p)
	assert.Equal(t, `{"cli":"agy"}`, string(data))
}
