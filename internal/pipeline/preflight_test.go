package pipeline

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func penv(tools []string, files ...string) PreflightEnv {
	has := map[string]bool{}
	for _, t := range tools {
		has[t] = true
	}
	exist := map[string]bool{}
	for _, f := range files {
		exist[filepath.FromSlash("/proj/"+f)] = true
	}
	return PreflightEnv{
		LookPath: func(n string) (string, error) {
			if has[n] {
				return "/usr/bin/" + n, nil
			}
			return "", errors.New("não encontrado")
		},
		Exists: func(p string) bool { return exist[p] },
	}
}

func findItem(items []Item, role, level string) *Item {
	for i := range items {
		if items[i].Session == role && items[i].Level == level {
			return &items[i]
		}
	}
	return nil
}

func TestPreflight_TudoCerto(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIClaude, config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}
	env := penv([]string{"claude", "codex", "gemini"}, "CLAUDE.md", "AGENTS.md", "GEMINI.md", ".claude/agents/lider.md")
	items := Preflight("/proj", cfg, env)
	assert.False(t, PreflightBlocked(items), FormatPreflight(items))
	for _, it := range items {
		assert.Equal(t, LevelOK, it.Level, "%+v", it)
	}
}

func TestPreflight_CLIAusenteBloqueia(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleBuilder: config.CLICodex}
	env := penv([]string{"claude"}, "CLAUDE.md", "AGENTS.md", ".claude/agents/lider.md", ".claude/agents/reviewer.md")
	items := Preflight("/proj", cfg, env)
	assert.True(t, PreflightBlocked(items))
	it := findItem(items, config.SessionRoleBuilder, LevelBlock)
	require.NotNil(t, it)
	assert.Contains(t, it.Message, "codex")
	assert.Contains(t, it.Action, "@openai/codex")
	assert.Contains(t, FormatPreflight(items), "✗")
}

func TestPreflight_ArquivoDeInstrucaoAusenteAvisaComAAcao(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleBuilder: config.CLICodex}
	env := penv([]string{"claude", "codex"}, "CLAUDE.md", ".claude/agents/lider.md", ".claude/agents/reviewer.md") // sem AGENTS.md
	items := Preflight("/proj", cfg, env)
	assert.False(t, PreflightBlocked(items), "arquivo ausente só avisa")
	it := findItem(items, config.SessionRoleBuilder, LevelWarn)
	require.NotNil(t, it)
	assert.Contains(t, it.Message, "AGENTS.md")
	assert.Equal(t, "forge-sdd update --agent openai", it.Action, "codex é o agente openai do scaffold")
}

func TestPreflight_CadaCLIApontaParaOArquivoCerto(t *testing.T) {
	want := map[string][]string{
		config.CLIClaude:  {"CLAUDE.md"},
		config.CLIGemini:  {"GEMINI.md"},
		config.CLICopilot: {".github/copilot-instructions.md", "AGENTS.md"},
		config.CLICodex:   {"AGENTS.md"},
	}
	for cli, files := range want {
		assert.Equal(t, files, agentFiles[cli].Files, cli)
	}
	// O Codex não lê OPENAI.md: com só OPENAI.md o aviso aparece.
	cfg := config.DefaultForgeConfig()
	cfg.CLI = config.CLICodex
	items := Preflight("/proj", cfg, penv([]string{"codex"}, "OPENAI.md"))
	require.NotNil(t, findItem(items, config.SessionRoleDiscovery, LevelWarn), "o aviso sai uma vez por agente, na primeira sessão que o usa")
}

func TestPreflight_ClaudeSemDefinicaoDeAgenteBloqueia(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	env := penv([]string{"claude"}, "CLAUDE.md", ".claude/agents/lider.md") // falta o reviewer
	items := Preflight("/proj", cfg, env)
	assert.True(t, PreflightBlocked(items))
	var msgs []string
	for _, it := range items {
		if it.Level == LevelBlock {
			msgs = append(msgs, it.Message)
		}
	}
	assert.Contains(t, strings.Join(msgs, "|"), ".claude/agents/reviewer.md")
}

func TestPreflight_ConfigInvalidaBloqueia(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Permissions = "tudo"
	items := Preflight("/proj", cfg, penv(nil))
	require.Len(t, items, 1)
	assert.Equal(t, LevelBlock, items[0].Level)
	assert.Contains(t, items[0].Message, "forge.json inválido")
}

func TestPreflight_NaoExecutaAAgenteNemLeOAmbiente(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	called := 0
	env := PreflightEnv{
		LookPath: func(n string) (string, error) { called++; return "/x/" + n, nil },
		Exists:   func(string) bool { return true },
	}
	_ = Preflight("/proj", cfg, env)
	assert.Equal(t, 3, called, "só LookPath, uma vez por sessão")
}
