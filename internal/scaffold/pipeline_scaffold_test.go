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

func TestScaffold_OpenAIGeraAGENTSmdEPreservaODoUsuario(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Project: "demo", Stack: "go", DB: "none", Lang: "pt-BR", Agents: []string{config.AgentOpenAI}, SddVersion: "4.2.0-beta"}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(agents), "# AGENTS.md — demo"))
	assert.LessOrEqual(t, len(agents), 1536)
	openai, _ := os.ReadFile(filepath.Join(dir, "OPENAI.md"))
	assert.Contains(t, string(openai), "AGENTS.md", "OPENAI.md vira ponteiro, sem duplicar o conteúdo")
	assert.True(t, IsGeneratedAgentsMD(filepath.Join(dir, "AGENTS.md")))

	// Um AGENTS.md do usuário nunca é sobrescrito.
	own := "# Meu AGENTS.md\nregras do time\n"
	dir2 := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "AGENTS.md"), []byte(own), 0o644))
	_, err = Run(cfg, dir2)
	require.NoError(t, err)
	got, _ := os.ReadFile(filepath.Join(dir2, "AGENTS.md"))
	assert.Equal(t, own, string(got))
	assert.False(t, IsGeneratedAgentsMD(filepath.Join(dir2, "AGENTS.md")))
}

func TestAgentsMD_TituloIgualDoUsuarioNaoEhGerado(t *testing.T) {
	dir := t.TempDir()
	own := "# AGENTS.md — Regras do time\n\nNão mexer.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(own), 0o644))
	assert.False(t, IsGeneratedAgentsMD(filepath.Join(dir, "AGENTS.md")), "só o título não basta: falta o marcador")

	cfg := config.Config{Project: "demo", Stack: "go", DB: "none", Lang: "pt-BR", Agents: []string{config.AgentOpenAI}, SddVersion: "4.2.0-beta"}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	got, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	assert.Equal(t, own, string(got), "init/update não sobrescrevem o arquivo do usuário, mesmo com o título parecido")

	// O gerado tem o marcador e é reconhecido como gerado.
	dir2 := t.TempDir()
	_, err = Run(cfg, dir2)
	require.NoError(t, err)
	assert.True(t, IsGeneratedAgentsMD(filepath.Join(dir2, "AGENTS.md")))
	gen, _ := os.ReadFile(filepath.Join(dir2, "AGENTS.md"))
	assert.Contains(t, string(gen), "<!-- forge-sdd:generated -->")
}

func TestScaffold_SkillsDoModoPipelineDizemQueADiscoveryNaoExecuta(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{Project: "demo", Stack: "go", DB: "none", Lang: "pt-BR", SddVersion: "4.2.0-beta", Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		return string(b)
	}
	split := read(".agents/skills/sdd-planejamento/references/split-features.md")
	assert.Contains(t, split, "forge-sdd dispatch <iniciativa>")
	assert.Contains(t, split, "entregue ao Builder")
	assert.Contains(t, split, "Esta sessão não implementa")
	impl := read(".agents/skills/sdd-build/references/implementar.md")
	assert.Contains(t, impl, "nunca edite `forge.json` nem o código da feature")
	assert.Contains(t, read(".agents/skills/sdd-concepcao/references/discovery.md"), "esta sessão nunca constrói")
}

func TestScaffold_ProtocoloBrief(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{Project: "demo", Stack: "go", DB: "none", Lang: "pt-BR", SddVersion: "4.2.0-beta", Agents: []string{config.AgentClaude, config.AgentGemini, config.AgentOpenAI}}, dir)
	require.NoError(t, err)
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		return string(b)
	}

	assert.Contains(t, read("CLAUDE.md"), "Discovery começa por `forge-sdd brief`")
	assert.Contains(t, read("GEMINI.md"), "Discovery começa por `forge-sdd brief`")
	assert.Contains(t, read("AGENTS.md"), "Discovery começa por `forge-sdd brief`")

	concepcao := read(".agents/skills/sdd-concepcao/SKILL.md")
	assert.Contains(t, concepcao, "rode `forge-sdd brief` antes de tudo")
	assert.Contains(t, concepcao, "Decisões do projeto aplicadas")

	for _, ref := range []string{
		".agents/skills/sdd-concepcao/references/discovery.md",
		".agents/skills/sdd-concepcao/references/nova-feature.md",
		".agents/skills/sdd-concepcao/references/novo-fix.md",
		".agents/skills/sdd-planejamento/references/split-features.md",
	} {
		assert.Contains(t, read(ref), "forge-sdd brief", ref)
	}
}

func TestScaffold_ConstitutionConteudo(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(config.Config{
		Project: "demo", Stack: "go", DB: "none", Lang: "pt-BR", SddVersion: "4.2.0-beta",
		Agents: []string{config.AgentClaude, config.AgentGemini, config.AgentCopilot},
	}, dir)
	require.NoError(t, err)

	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		return string(b)
	}

	targets := []string{
		".claude/commands/constitution.md",
		".gemini/prompts/constitution.prompt.md",
		".github/prompts/constitution.prompt.md",
		".agents/skills/sdd-concepcao/references/constitution.md",
	}

	for _, path := range targets {
		c := read(path)
		assert.Contains(t, c, "módulos", path)
		assert.Contains(t, c, "fluxos", path)
		assert.Contains(t, c, "decisões", path)
		assert.Contains(t, c, "ADR", path)
		assert.Contains(t, c, "segurança", path)
		assert.Contains(t, c, "brief --check", path)
	}
}

func TestScaffold_DocsCitamBriefConstituicaoETmux(t *testing.T) {
	root := filepath.Join("..", "..")

	docs := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "docs", "pipeline.md"),
		filepath.Join("templates", ".forge", "LEIA-ME.md.tmpl"),
		filepath.Join(root, "site", "public", "guia-pipeline.html"),
	}

	for _, doc := range docs {
		data, err := os.ReadFile(doc)
		require.NoError(t, err, "arquivo de documentação deve existir: %s", doc)
		content := string(data)

		assert.True(t, strings.Contains(content, "brief") || strings.Contains(content, "forge-sdd brief"), "%s deve citar brief", doc)
		assert.True(t, strings.Contains(content, "/constitution") || strings.Contains(content, "Constituição"), "%s deve citar constitution", doc)
		assert.True(t, strings.Contains(content, "tmux"), "%s deve citar tmux", doc)
	}
}
