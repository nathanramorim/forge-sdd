package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pipelineCfg(agents ...string) config.Config {
	return config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: "4.1.0", Agents: agents}
}

func TestPipeline_ScriptsDeLauncherInstaladosEExecutaveis(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(pipelineCfg(config.AgentClaude), dir)
	require.NoError(t, err)

	sh := filepath.Join(dir, ".forge", "forge-up.sh")
	info, err := os.Stat(sh)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.NotZero(t, info.Mode().Perm()&0o100, "forge-up.sh precisa do bit de execução")
	}
	data, _ := os.ReadFile(sh)
	assert.True(t, strings.HasPrefix(string(data), "#!/usr/bin/env bash"))
	assert.Contains(t, string(data), `exec forge-sdd up "$@"`)
	assert.NotContains(t, string(data), "osascript", "a lógica de terminais fica em forge-sdd up (D-8)")

	ps1, err := os.ReadFile(filepath.Join(dir, ".forge", "forge-up.ps1"))
	require.NoError(t, err)
	assert.Contains(t, string(ps1), "& forge-sdd up @args")
	assert.Contains(t, string(ps1), "Set-Location")
	assert.NotContains(t, string(ps1), "bash", "o .ps1 roda sem Git Bash")

	if bash, err := exec.LookPath("bash"); err == nil {
		out, err := exec.Command(bash, "-n", sh).CombinedOutput()
		assert.NoError(t, err, string(out))
	}
}

func TestPipeline_ForgeUpFalhaComMensagemSemForgeSdd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("script bash")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash indisponível")
	}
	dir := t.TempDir()
	_, err = Run(pipelineCfg(config.AgentClaude), dir)
	require.NoError(t, err)
	cmd := exec.Command(bash, filepath.Join(dir, ".forge", "forge-up.sh"), "--dry-run")
	cmd.Env = []string{"PATH=/nonexistent"}
	out, err := cmd.CombinedOutput()
	require.Error(t, err)
	assert.Contains(t, string(out), "forge-sdd não está no PATH")
}

func TestPipeline_HookDoInboxSoComClaudeESemApagarConfiguracao(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, ".claude/settings.json", `{"model":"opus","hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo ok"}]}]}}`)
	_, err := Run(pipelineCfg(config.AgentClaude), dir)
	require.NoError(t, err)
	m := readSettings(t, dir)
	assert.Equal(t, "opus", m["model"])
	assert.Equal(t, []string{inboxHookCommand}, hookCommands(m))
	assert.NotNil(t, m["hooks"].(map[string]any)["PreToolUse"])

	_, err = Run(pipelineCfg(config.AgentClaude), dir) // update: idempotente
	require.NoError(t, err)
	assert.Equal(t, []string{inboxHookCommand}, hookCommands(readSettings(t, dir)))

	other := t.TempDir()
	_, err = Run(pipelineCfg(config.AgentGemini), other)
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(other, ".claude", "settings.json"), "sem Claude não há hook")
}

func TestPipeline_JSONInvalidoDoUsuarioNaoQuebraOUpdate(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, ".claude/settings.json", `{"hooks":`)
	_, err := Run(pipelineCfg(config.AgentClaude), dir)
	require.NoError(t, err, "o update continua mesmo sem poder instalar o hook")
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	assert.Equal(t, `{"hooks":`, string(data), "arquivo do usuário intacto")
}

func TestPipeline_AdaptadoresEAgentesNasTresFerramentas(t *testing.T) {
	dir := renderAllAgents(t)
	for _, f := range []string{
		".claude/commands/inbox.md", ".claude/commands/implement.md", ".claude/agents/reviewer.md",
		".gemini/prompts/inbox.prompt.md", ".gemini/prompts/implement.prompt.md", ".gemini/skills/reviewer.chatmode.md",
		".github/prompts/inbox.prompt.md", ".github/prompts/implement.prompt.md", ".github/chatmodes/reviewer.chatmode.md",
		".agents/skills/sdd-build/references/worker-job.md", ".agents/skills/sdd-governanca/references/pr-review.md",
		".agents/skills/sdd-operacao/references/inbox.md", ".agents/skills/sdd-operacao/references/pipeline.md",
	} {
		assert.FileExists(t, filepath.Join(dir, filepath.FromSlash(f)), f)
	}
	agent, err := os.ReadFile(filepath.Join(dir, ".claude", "agents", "reviewer.md"))
	require.NoError(t, err)
	assert.Contains(t, string(agent), "tools: Read, Grep, Glob, Bash")
	assert.NotContains(t, string(agent), "Edit")
	assert.NotContains(t, string(agent), "Write")
	assert.NoFileExists(t, filepath.Join(dir, ".forge", "config", "forge.json"), "o scaffold nunca ativa o pipeline")
}

func TestPipeline_CheatSheetListaInboxEOcultaAliases(t *testing.T) {
	var names []string
	for _, c := range CommandCheatSheet(pipelineCfg(config.AgentClaude)) {
		names = append(names, c.Command)
	}
	assert.Contains(t, names, "/inbox")
	assert.Contains(t, names, "/implementar")
	assert.NotContains(t, names, "/implement")
	assert.NotContains(t, names, "/proxima-feature")
}
