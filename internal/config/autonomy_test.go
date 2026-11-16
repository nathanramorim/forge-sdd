package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutonomy_PadraoNormal(t *testing.T) {
	cfg := DefaultForgeConfig()
	assert.False(t, cfg.IsAdvanced())
	assert.True(t, ValidAutonomy(cfg.Autonomy))
	require.NoError(t, cfg.Validate())
}

func TestAutonomy_ValorInvalido(t *testing.T) {
	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"autonomy":"turbo"}`)
	_, _, err := LoadForgeConfig(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "autonomy")
}

func TestAutonomy_AvancadoExigeFullEIsolado(t *testing.T) {
	cfg := DefaultForgeConfig()
	cfg.Autonomy = AutonomyAdvanced
	require.Error(t, cfg.Validate(), "sem permissions full")

	cfg.Permissions = PermissionFull
	require.Error(t, cfg.Validate(), "sem isolated_confirmed")

	cfg.IsolatedConfirmed = true
	require.NoError(t, cfg.Validate())
	assert.True(t, cfg.IsAdvanced())
}

func TestAutonomy_ArquivoAntigoValido(t *testing.T) {
	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"cli":"claude"}`)
	cfg, found, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, found)
	assert.False(t, cfg.IsAdvanced())
}

func TestDefaultCard_Falso(t *testing.T) {
	assert.False(t, DefaultForgeConfig().PMMCP.RequireExistingCard)

	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"pm_mcp":{"provider":"auto","supported":["jira"],"require_existing_card":true,"status_on_implement":"A","status_on_pr_approved":"B","evidence_max_age_minutes":30}}`)
	cfg, _, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.PMMCP.RequireExistingCard, "forge.json existente mantém o valor")
}

func TestAutonomy_AvisoDoDoctor(t *testing.T) {
	dir := t.TempDir()
	assert.Empty(t, AdvancedWarning(dir))
	writeForge(t, dir, ".forge/config/forge.json", `{"autonomy":"advanced","permissions":"full"}`)
	assert.Contains(t, AdvancedWarning(dir), "isolado")
	writeForge(t, dir, ".forge/config/forge.json", `{"autonomy":"advanced","permissions":"full","isolated_confirmed":true}`)
	assert.Empty(t, AdvancedWarning(dir))
}
