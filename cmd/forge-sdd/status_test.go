package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	// O rootCmd é global: restaura os padrões das flags entre execuções.
	_ = statusCmd.Flags().Set("all-branches", "false")
	_ = statusCmd.Flags().Set("initiatives", "false")
	_ = statusCmd.Flags().Set("live", "false")
	_ = statusCmd.Flags().Set("once", "false")
	_ = statusCmd.Flags().Set("compact", "false")
	_ = nextIDCmd.Flags().Set("kind", "discovery")
	_ = nextIDCmd.Flags().Set("workitem", "")
	_ = updateCmd.Flags().Set("dry-run", "false")
	_ = contextCmd.Flags().Set("station", "build")
	_ = pipelineInitCmd.Flags().Set("dry-run", "false")
	_ = dispatchCmd.Flags().Set("card-evidence", "")
	_ = dispatchCmd.Flags().Set("dry-run", "false")
	_ = briefCmd.Flags().Set("dir", ".")
	_ = briefCmd.Flags().Set("max-bytes", "0")
	_ = briefCmd.Flags().Set("check", "false")
	resetWorkerFlags()
	resetReviewFlags()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return out.String(), err
}

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	testutil.Git(t, dir, args...)
}

func TestStatusENextID_ComIniciativaEmOutraBranch(t *testing.T) {
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", "10-antiga"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "discovery", "10-antiga", "discovery.md"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(`{"naming_convention":"sequencial"}`), 0o644))
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "base")
	gitT(t, dir, "checkout", "-q", "-b", "feat/70-alfa")
	base := filepath.Join(dir, "sdd", "discovery", "70-alfa")
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "status.md"), []byte("## Status\n```\n70-01 [x] done\n```\n## Proximo passo\nSeguir 70-02.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "queue.json"), []byte(`{"items":[{"task_id":"70-01","status":"approved"},{"task_id":"70-02","status":"queued"}]}`), 0o644))
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-q", "-m", "70")
	gitT(t, dir, "checkout", "-q", "main")

	out, err := runRoot(t, "next-id", "--dir", dir)
	require.NoError(t, err)
	assert.Equal(t, "71\n", out)

	out, err = runRoot(t, "status", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "10-antiga")
	assert.NotContains(t, out, "70-alfa")

	out, err = runRoot(t, "status", "--dir", dir, "--all-branches")
	require.NoError(t, err)
	assert.Contains(t, out, "70-alfa")
	assert.Contains(t, out, "feat/70-alfa")
	assert.Contains(t, out, "70-01 [x] done")
	assert.Contains(t, out, "Seguir 70-02.")
	assert.Contains(t, out, "approved=1 queued=1")
}

func TestStatus_SemIniciativas(t *testing.T) {
	out, err := runRoot(t, "status", "--dir", t.TempDir())
	require.NoError(t, err)
	assert.Contains(t, out, "Nenhuma iniciativa")
}

func TestNextID_KindInvalidoEWorkitemSemValor(t *testing.T) {
	_, err := runRoot(t, "next-id", "--dir", t.TempDir(), "--kind", "feature")
	require.Error(t, err)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(`{"naming_convention":"workitem"}`), 0o644))
	_, err = runRoot(t, "next-id", "--dir", dir)
	require.Error(t, err)

	out, err := runRoot(t, "next-id", "--dir", dir, "--workitem", "1024")
	require.NoError(t, err)
	assert.Equal(t, "1024\n", out)
}

func TestNextID_SddrcInvalidoFalhaEmVezDeAdivinhar(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte("{quebrado"), 0o644))
	_, err := runRoot(t, "next-id", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inválido")
}

func TestContext_PorTaskIDEPorCaminho(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "- **ARQ-01:** Fronteiras.\n- **TST-01:** Executável.\n")
	testutil.WriteFile(t, dir, ".agents/skills/sdd-build/SKILL.md", "---\nname: sdd-build\nstatus: approved\nscope: core\n---\n")
	testutil.WriteFile(t, dir, "sdd/discovery/70-x/features/feat-70-01.md", "# Spec\nCita ARQ-01.\n")
	testutil.WriteFile(t, dir, "sdd/.queue/queue.json", `{"items":[{"task_id":"70-01","spec_path":"sdd/discovery/70-x/features/feat-70-01.md","status":"queued","rules":["TST-01"]}]}`)

	out, err := runRoot(t, "context", "70-01", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "- **ARQ-01:** Fronteiras.")
	assert.Contains(t, out, "- **TST-01:** Executável.", "IDs da fila também entram")
	assert.Contains(t, out, ".agents/skills/sdd-build/SKILL.md")

	out, err = runRoot(t, "context", "sdd/discovery/70-x/features/feat-70-01.md", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "ARQ-01")

	_, err = runRoot(t, "context", "99-99", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "70-01", "o erro lista as tasks disponíveis")
}

func TestStatusLiveFlag(t *testing.T) {
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "feat/80-status")
	gitT(t, dir, "commit", "-q", "--allow-empty", "-m", "feat: commit inicial")

	logsDir := filepath.Join(dir, ".forge", "runtime", "logs")
	require.NoError(t, os.MkdirAll(logsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(logsDir, "builder-job-1.log"), []byte("✎ internal/status.go\n… trabalhando há 1m · 1 arquivo alterado\n"), 0o644))

	sddDir := filepath.Join(dir, "sdd", "discovery", "80-status")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "queue.json"), []byte(`{
		"items": [
			{"task_id": "80-01", "title": "Template Canônico", "status": "approved"},
			{"task_id": "80-02", "title": "Subcomando live", "status": "queued"}
		]
	}`), 0o644))

	// Testa flag --compact (Template Canônico tríplice para agentes de IA)
	start := time.Now()
	outCompact, err := runRoot(t, "status", "--dir", dir, "--compact")
	elapsedCompact := time.Since(start)
	require.NoError(t, err)
	assert.Less(t, elapsedCompact, 100*time.Millisecond, "CA-01: status compacto deve executar em menos de 100ms")

	assert.Contains(t, outCompact, "### Status da Esteira —")
	assert.Contains(t, outCompact, "1. Ao Vivo no Painel")
	assert.Contains(t, outCompact, "✎ internal/status.go")
	assert.Contains(t, outCompact, "2. Commits Recentes no Git:")
	assert.Contains(t, outCompact, "feat: commit inicial")
	assert.Contains(t, outCompact, "3. Fila de Tarefas:")
	assert.Contains(t, outCompact, "• [x] 80-01: Template Canônico")
	assert.Contains(t, outCompact, "• [ ] 80-02: Subcomando live")
	assert.Contains(t, outCompact, "Próximo passo: Acompanhar o Builder")

	tokensCompact := len(strings.Fields(outCompact))
	assert.Less(t, tokensCompact, 250, "CA-03: saída compacta deve respeitar o limite de 250 tokens")

	// Testa flag longa --live (Painel executivo vivo/living dashboard)
	startLive := time.Now()
	out, err := runRoot(t, "status", "--dir", dir, "--live", "--once")
	elapsedLive := time.Since(startLive)

	require.NoError(t, err)
	assert.Less(t, elapsedLive, 500*time.Millisecond, "painel vivo snapshot deve executar rapidamente (< 500ms)")

	assert.Contains(t, out, "feat/80-status")
	assert.Contains(t, out, "PROCESSAMENTO & CONSUMO POR SESSÃO")
	assert.Contains(t, out, "PROGRESSO DA ESTEIRA (DAG)")
	assert.Contains(t, out, "50% (1/2 tasks)")
	assert.Contains(t, out, "[✓] 80-01: Template Canônico")
	assert.Contains(t, out, "[ ] 80-02: Subcomando live")
	assert.Contains(t, out, "COMMITS RECENTES NO GIT")
	assert.Contains(t, out, "feat: commit inicial")
	assert.Contains(t, out, "Próximo passo: Acompanhar o Builder")

	// Testa flag curta -l
	outShort, err := runRoot(t, "status", "--dir", dir, "-l", "--once")
	require.NoError(t, err)
	assert.Equal(t, out, outShort, "flag curta -l deve produzir a mesma saída que --live")

	// Testa forge status padrão (snapshot no CLI)
	outSnapshot, err := runRoot(t, "status", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, outSnapshot, "PROCESSAMENTO & CONSUMO POR SESSÃO")
	assert.Contains(t, outSnapshot, "PROGRESSO DA ESTEIRA (DAG)")
	assert.Contains(t, outSnapshot, "INICIATIVAS (sdd/discovery)")
	assert.Contains(t, outSnapshot, "80-status")
}

