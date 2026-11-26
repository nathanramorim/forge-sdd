package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNaNuvem(t *testing.T) {
	assert.True(t, NaNuvem(func(string) string { return "true" }))
	assert.False(t, NaNuvem(func(string) string { return "" }))
}

func TestReposFaltando_HubComPastasIrmas(t *testing.T) {
	raiz := t.TempDir()
	hub := filepath.Join(raiz, "hub")
	require.NoError(t, os.MkdirAll(hub, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(raiz, "api", ".git"), 0o755))
	got := ReposFaltando(hub, []string{"hub", "dono/api", "dono/web"})
	assert.Equal(t, []string{"dono/web"}, got, "o próprio projeto e os clonados não faltam")
	assert.Empty(t, ReposFaltando(hub, nil))
}

func TestChecklistNuvem(t *testing.T) {
	t.Run("npm", func(t *testing.T) {
		s := ChecklistNuvem(config.NuvemConfig{Repos: []string{"api"}}, "hub")
		assert.Contains(t, s, "hub, api")
		assert.Contains(t, s, "registry.npmjs.org")
	})
	t.Run("git", func(t *testing.T) {
		s := ChecklistNuvem(config.NuvemConfig{Instalar: "git"}, "app")
		assert.Contains(t, s, "proxy.golang.org")
	})
}

func TestResumoNuvem_SoReapareceQuandoAEsteiraMuda(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "runtime", "queues", "builder", "pending"), 0o755))
	agora := time.Date(2026, 11, 26, 1, 0, 0, 0, time.UTC)
	primeiro := ResumoNuvem(dir, agora)
	assert.Contains(t, primeiro, "DIAGNÓSTICO")
	assert.Empty(t, ResumoNuvem(dir, agora.Add(40*time.Second)), "nada mudou: não repete")

	_, _, err := bus.Enqueue(dir, layout.RoleBuilder, jobDeTeste("job-x", "77-a", bus.TypeFeatureBuild, "feat/77-a"), agora)
	require.NoError(t, err)
	assert.Contains(t, ResumoNuvem(dir, agora.Add(time.Minute)), "1 pendente", "um job novo reaparece")
}
