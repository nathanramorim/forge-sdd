package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForgeProfiles_QuatroAgentesValidos(t *testing.T) {
	cfg := DefaultForgeConfig()
	require.NoError(t, cfg.Validate())
	for _, cli := range OfferedCLIs() {
		p, ok := cfg.Profile(cli)
		require.True(t, ok, cli)
		assert.NotEmpty(t, p.Command, cli)
		assert.NotEmpty(t, p.Args, cli)
		assert.NotEmpty(t, p.InteractiveArgs, cli)
		assert.Contains(t, p.Args, MarkerPermissions, "%s: o worker precisa do marcador de permissão", cli)
		assert.NoError(t, validateProfile(cli, p), cli)
	}
	// O -p é prompt no claude, no gemini e no copilot, mas perfil no codex: o headless do codex é exec.
	assert.Equal(t, "exec", cfg.CLIProfiles[CLICodex].Args[0])
	assert.Equal(t, "-p", cfg.CLIProfiles[CLIGemini].Args[0])
	assert.NotContains(t, cfg.CLIProfiles[CLICodex].Args, "-p")
}

func TestForgeProfiles_ForgeJSONDa41SemSessionsContinuaValido(t *testing.T) {
	dir := t.TempDir()
	old := `{"cli":"claude","cli_profiles":{"claude":{"command":"claude","args":["-p","{prompt}","--agent","{agent}","-n","{name}"]}}}`
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".forge", "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".forge", "config", "forge.json"), []byte(old), 0o644))
	cfg, found, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	require.True(t, found)
	for _, role := range SessionRoles() {
		assert.Equal(t, CLIClaude, cfg.CLIFor(role), role)
	}
	p, ok := cfg.Profile(CLIClaude)
	require.True(t, ok)
	assert.Equal(t, []string{"-n", "{name}", "--agent", "{agent}"}, p.InteractiveArgs, "completa o que a 4.1 não trazia")
	assert.Equal(t, PermissionProject, cfg.Permissions)
}

func TestForgeProfiles_SessionsPorPapel(t *testing.T) {
	cfg := DefaultForgeConfig()
	cfg.Sessions = map[string]string{SessionRoleDiscovery: CLIClaude, SessionRoleBuilder: CLICodex, SessionRoleReviewer: CLIGemini}
	require.NoError(t, cfg.Validate())
	assert.Equal(t, CLICodex, cfg.CLIFor(SessionRoleBuilder))
	assert.Equal(t, CLIGemini, cfg.CLIFor(SessionRoleReviewer))
	cfg.Sessions = map[string]string{SessionRoleBuilder: CLICodex}
	assert.Equal(t, CLIClaude, cfg.CLIFor(SessionRoleDiscovery), "papel sem escolha cai no cli")
}

func TestForgeProfiles_Violacoes(t *testing.T) {
	bad := map[string]func(*ForgeConfig){
		"papel desconhecido":   func(c *ForgeConfig) { c.Sessions = map[string]string{"chefe": CLIClaude} },
		"agente fora da lista": func(c *ForgeConfig) { c.Sessions = map[string]string{SessionRoleBuilder: "gpt"} },
		"agente sem perfil": func(c *ForgeConfig) {
			delete(c.CLIProfiles, CLICodex)
			c.Sessions = map[string]string{SessionRoleBuilder: CLICodex}
		},
		"permissions inválido":      func(c *ForgeConfig) { c.Permissions = "tudo" },
		"allow_commands com espaço": func(c *ForgeConfig) { c.AllowCommands = []string{"git push"} },
		"allow_commands com shell":  func(c *ForgeConfig) { c.AllowCommands = []string{"git;rm"} },
		"cli desconhecido":          func(c *ForgeConfig) { c.CLI = "cursor" },
		"comando com metacaractere": func(c *ForgeConfig) {
			c.CLIProfiles[CLICodex] = CLIProfile{Command: "codex;rm", Args: []string{"exec"}}
		},
		"marcador desconhecido": func(c *ForgeConfig) {
			c.CLIProfiles[CLICodex] = CLIProfile{Command: "codex", Args: []string{"{secret}"}}
		},
		"permissions embutido": func(c *ForgeConfig) {
			c.CLIProfiles[CLICodex] = CLIProfile{Command: "codex", Args: []string{"--x={permissions}"}}
		},
		"interativo com controle": func(c *ForgeConfig) {
			c.CLIProfiles[CLICodex] = CLIProfile{Command: "codex", Args: []string{"exec"}, InteractiveArgs: []string{"a\nb"}}
		},
	}
	for name, mut := range bad {
		cfg := DefaultForgeConfig()
		mut(&cfg)
		assert.Error(t, cfg.Validate(), name)
	}
}

func TestPermissionArgs_TabelaPorCLIeNivel(t *testing.T) {
	allow := []string{"git", "go"}
	assert.Equal(t, []string{"--permission-mode", "acceptEdits", "--allowedTools", "Read", "Edit", "Write", "Glob", "Grep", "Bash(git *)", "Bash(go *)"}, PermissionArgs(CLIClaude, PermissionProject, allow))
	assert.Equal(t, []string{"--dangerously-skip-permissions"}, PermissionArgs(CLIClaude, PermissionFull, allow))
	assert.Equal(t, []string{"--approval-mode", "auto_edit"}, PermissionArgs(CLIGemini, PermissionProject, allow))
	assert.Equal(t, []string{"--approval-mode", "yolo"}, PermissionArgs(CLIGemini, PermissionFull, allow))
	assert.Equal(t, []string{"--allow-tool=write", "--allow-tool=shell(git:*)", "--allow-tool=shell(go:*)"}, PermissionArgs(CLICopilot, PermissionProject, allow))
	assert.Equal(t, []string{"--allow-all-tools"}, PermissionArgs(CLICopilot, PermissionFull, allow))
	assert.Equal(t, []string{"--sandbox", "workspace-write"}, PermissionArgs(CLICodex, PermissionProject, allow))
	assert.Equal(t, []string{"--dangerously-bypass-approvals-and-sandbox"}, PermissionArgs(CLICodex, PermissionFull, allow))
	assert.Nil(t, PermissionArgs("cursor", PermissionProject, allow))
	assert.Nil(t, PermissionArgs(CLIClaude, "tudo", allow))
	assert.Contains(t, PermissionSummary(PermissionFull, allow), "ambiente isolado")
	assert.Contains(t, PermissionSummary(PermissionProject, allow), "git, go")
}

func TestSaveForgeConfig_PreservaCamposEGravaAtomico(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultForgeConfig()
	cfg.Sessions = map[string]string{SessionRoleBuilder: CLICodex}
	cfg.Reviewer.MaxReworkAttempts = 5
	cfg.Permissions = PermissionFull
	require.NoError(t, SaveForgeConfig(dir, cfg))
	got, found, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, CLICodex, got.CLIFor(SessionRoleBuilder))
	assert.Equal(t, 5, got.Reviewer.MaxReworkAttempts)
	assert.Equal(t, PermissionFull, got.Permissions)

	entries, _ := os.ReadDir(filepath.Join(dir, ".forge", "config"))
	for _, e := range entries {
		assert.False(t, strings.Contains(e.Name(), ".tmp"), "sem arquivo temporário: %s", e.Name())
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	assert.True(t, json.Valid(raw))

	// Config inválida não grava nada novo.
	bad := got
	bad.Permissions = "tudo"
	assert.Error(t, SaveForgeConfig(dir, bad))
	again, _, _ := LoadForgeConfig(dir)
	assert.Equal(t, PermissionFull, again.Permissions)
}
