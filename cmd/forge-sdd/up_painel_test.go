package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func instalaModPainel(t *testing.T, dir string) {
	t.Helper()
	marca := filepath.Join(dir, ".agents", "plugins", "forge-painel", ".claude-plugin", "plugin.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(marca), 0o755))
	require.NoError(t, os.WriteFile(marca, []byte("{}"), 0o644))
}

// iniciadorFalso troca o iniciador real: registra o que seria iniciado, sem processo algum.
func iniciadorFalso(t *testing.T) *[]pipeline.PainelWorker {
	t.Helper()
	var chamados []pipeline.PainelWorker
	orig := newPainelStarter
	t.Cleanup(func() { newPainelStarter = orig })
	pid := 4000
	newPainelStarter = func(w pipeline.PainelWorker, _ string) (int, error) {
		chamados = append(chamados, w)
		pid++
		return pid, nil
	}
	return &chamados
}

func TestUp_PainelDryRunListaWorkersEmSegundoPlanoSemJanela(t *testing.T) {
	dir := pipelineDir(t, nil)
	instalaModPainel(t, dir)
	chamados := iniciadorFalso(t)
	resetWorkerFlags()

	out, err := runRoot(t, "up", "--dir", dir, "--terminal", "painel", "--dry-run")
	require.NoError(t, err)

	assert.Contains(t, out, "worker --role builder --name forge-builder --cli claude")
	assert.Contains(t, out, "worker --role reviewer --name forge-reviewer --cli claude")
	assert.Contains(t, out, filepath.Join(dir, ".forge", "runtime", "logs", "builder-worker.log"), "log em arquivo")
	assert.Contains(t, out, "--plugin-dir "+filepath.Join(dir, ".agents", "plugins", "forge-painel"))
	assert.Contains(t, out, "[DRY-RUN] Nada foi iniciado")
	for _, janela := range []string{"osascript", "gnome-terminal", "konsole", "x-terminal-emulator", "tmux new-session", "wt.exe", "cmd.exe"} {
		assert.NotContains(t, out, janela, "o modo painel não abre janela nem tmux")
	}
	assert.Empty(t, *chamados, "--dry-run não inicia nada")
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "logs"), "--dry-run não grava nada (ARQ-05)")
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "painel"))
}

// Um comando só: sem o mod, o forge up o instala do binário (dry-run só avisa e não grava nada).
func TestUp_PainelSemOModInstalaSozinho(t *testing.T) {
	dir := pipelineDir(t, nil)
	chamados := iniciadorFalso(t)
	resetWorkerFlags()
	out, err := runRoot(t, "up", "--dir", dir, "--terminal", "painel", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "o mod forge-painel seria instalado")
	assert.NoDirExists(t, filepath.Join(dir, ".agents", "plugins", "forge-painel"))

	resetWorkerFlags()
	out, err = runRoot(t, "up", "--dir", dir, "--terminal", "painel")
	require.NoError(t, err)
	assert.Contains(t, out, "✓ mod forge-painel instalado ou atualizado")
	assert.FileExists(t, filepath.Join(dir, ".agents", "plugins", "forge-painel", ".claude-plugin", "forge.json"))
	assert.Len(t, *chamados, 2)

	// De novo, já igual: nada a gravar.
	resetWorkerFlags()
	out, err = runRoot(t, "up", "--dir", dir, "--terminal", "painel")
	require.NoError(t, err)
	assert.NotContains(t, out, "instalado ou atualizado")
}

func TestUp_PainelIniciaOsWorkersEmSegundoPlano(t *testing.T) {
	dir := pipelineDir(t, nil)
	instalaModPainel(t, dir)
	chamados := iniciadorFalso(t)
	resetWorkerFlags()

	out, err := runRoot(t, "up", "--dir", dir, "--terminal", "painel")
	require.NoError(t, err)

	require.Len(t, *chamados, 2)
	assert.Equal(t, "forge-builder", (*chamados)[0].Title)
	assert.Equal(t, "forge-reviewer", (*chamados)[1].Title)
	assert.Equal(t, []string{"worker", "--role", "builder", "--name", "forge-builder", "--cli", "claude"}, (*chamados)[0].Argv[1:])
	assert.Contains(t, out, "forge-builder iniciado em segundo plano (pid 4001)")
	assert.Contains(t, out, "forge-reviewer iniciado em segundo plano (pid 4002)")
	assert.Contains(t, out, "Para acompanhar, abra o Claude Code com o mod: claude")
	assert.FileExists(t, filepath.Join(dir, ".forge", "runtime", "painel", "builder.pid"))
}

func TestUp_PainelComDiscoveryForaDoClaudeRecusa(t *testing.T) {
	dir := pipelineDir(t, nil)
	instalaModPainel(t, dir)
	chamados := iniciadorFalso(t)
	resetWorkerFlags()

	_, err := runRoot(t, "up", "--dir", dir, "--terminal", "painel", "--discovery", "gemini", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exige o claude")
	assert.Empty(t, *chamados)
}

func TestUp_PainelVindoDoForgeJSON(t *testing.T) {
	dir := pipelineDir(t, func(c *config.ForgeConfig) { c.Terminal = config.TerminalPainel })
	instalaModPainel(t, dir)
	iniciadorFalso(t)
	resetWorkerFlags()

	out, err := runRoot(t, "up", "--dir", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "[DRY-RUN] Nada foi iniciado")
	assert.False(t, strings.Contains(out, "modo headless"), "não cai no caminho das janelas")
}

func TestUp_TerminalInvalidoContinuaRecusado(t *testing.T) {
	dir := pipelineDir(t, nil)
	resetWorkerFlags()
	_, err := runRoot(t, "up", "--dir", dir, "--terminal", "painelzinho", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "painel")
}
