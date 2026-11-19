package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tocaAiProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, config.SaveForgeConfig(dir, config.DefaultForgeConfig()))
	base := filepath.Join(dir, "sdd", "discovery", "90-meta-longa")
	require.NoError(t, os.MkdirAll(filepath.Join(base, "features"), 0o755))
	for f, body := range map[string]string{
		"plan.md":                  "# Plano\n",
		"queue.json":               `{"spec_status":"reviewed","items":[]}`,
		"features/feat-90-01-a.md": "# a\n\n**Depende de:** nenhuma\n\n## Critérios de Aceitação Executáveis\n```bash\ngo test ./...\n```\n",
		"features/feat-90-02-b.md": "# b\n\n**Depende de:** feat-90-01\n\n## Critérios de Aceitação Executáveis\n```bash\ngo test ./...\n```\n",
		"requisitos.json":          `{"consentimentos":[{"id":"voz-propria","texto":"a voz é minha"}]}`,
	} {
		require.NoError(t, os.WriteFile(filepath.Join(base, f), []byte(body), 0o644))
	}
	return dir
}

func resetTocaAiFlags() {
	for _, n := range []string{"preflight", "dry-run"} {
		_ = pipelineTocaAiCmd.Flags().Set(n, "false")
	}
}

func TestPipelineTocaAi_PreFooListaPendenciaENaoEnfileira(t *testing.T) {
	dir := tocaAiProject(t)
	resetTocaAiFlags()
	out, err := runRoot(t, "pipeline", "toca-ai", "90-meta-longa", "--dir", dir, "--preflight")
	require.Error(t, err)
	assert.Contains(t, out, "voz-propria")
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "queues"))
}

func TestPipelineTocaAi_ConsentSoInterativoEEnfileiraDepois(t *testing.T) {
	dir := tocaAiProject(t)
	oldI, oldA := isInteractive, askConsent
	defer func() { isInteractive, askConsent = oldI, oldA }()

	isInteractive = func() bool { return false }
	_, err := runRoot(t, "pipeline", "consent", "voz-propria", "--dir", dir, "--iniciativa", "90-meta-longa")
	require.Error(t, err, "um agente (sem terminal) não concede consentimento")
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "runtime", "consents.json"))

	isInteractive = func() bool { return true }
	askConsent = func(string) (bool, error) { return false, nil }
	_, err = runRoot(t, "pipeline", "consent", "voz-propria", "--dir", dir, "--iniciativa", "90-meta-longa")
	require.Error(t, err)
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "runtime", "consents.json"))

	askConsent = func(string) (bool, error) { return true, nil }
	_, err = runRoot(t, "pipeline", "consent", "voz-propria", "--dir", dir, "--iniciativa", "90-meta-longa")
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, ".forge", "runtime", "consents.json"))

	resetTocaAiFlags()
	out, err := runRoot(t, "pipeline", "toca-ai", "90-meta-longa", "--dir", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "enfileiraria")
	assert.NoDirExists(t, filepath.Join(dir, ".forge", "runtime", "queues", "builder", "pending"))

	resetTocaAiFlags()
	out, err = runRoot(t, "pipeline", "toca-ai", "90-meta-longa", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "feat-90-01 → feat-90-02")
	list, err := os.ReadDir(filepath.Join(dir, ".forge", "runtime", "queues", "builder", "pending"))
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
