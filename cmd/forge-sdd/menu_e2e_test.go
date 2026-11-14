package main

import (
	"bytes"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_MenuPipeline percorre menu, seletor, checagem e o `up` real (em --dry-run, para não abrir
// terminais) com entrada simulada, e confere os três comandos gerados e a escolha gravada.
func TestE2E_MenuPipeline(t *testing.T) {
	dir := pipelineDir(t, nil)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	defer rootCmd.SetOut(nil)

	p := &scripted{
		selects: map[string][]string{
			"Forge SDD":                  {valUp, "<esc>"},
			"Agente da sessão":           {config.CLIClaude, config.CLICodex, config.CLIGemini},
			"Permissões do Builder e do": {config.PermissionProject},
		},
		confirms: []bool{true},
	}
	env := &menuEnv{
		dir: dir, out: &out, p: p, preflightEnv: allFound(),
		run: func(args ...string) error {
			if args[0] == "up" {
				args = append(args, "--dry-run", "--terminal", "native") // nunca abre terminais num teste
			}
			resetWorkerFlags()
			return runSubcommand(args...)
		},
	}
	require.NoError(t, runMenu(env))

	got := out.String()
	assert.Contains(t, got, "claude -n forge-discovery --agent lider", "Discovery com o agente escolhido")
	assert.Contains(t, got, "--role builder --name forge-builder --cli codex")
	assert.Contains(t, got, "--role reviewer --name forge-reviewer --cli gemini")
	assert.Contains(t, got, "[DRY-RUN]")

	cfg, _, err := config.LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, config.CLICodex, cfg.CLIFor(config.SessionRoleBuilder), "a escolha ficou gravada para a próxima vez")
}
