package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func penv(tools []string, files ...string) PreflightEnv {
	return penvFor("/proj", tools, files...)
}

func penvFor(dir string, tools []string, files ...string) PreflightEnv {
	has := map[string]bool{}
	for _, t := range tools {
		has[t] = true
	}
	exist := map[string]bool{}
	for _, f := range files {
		exist[filepath.Join(dir, filepath.FromSlash(f))] = true
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
	env := penv([]string{"claude", "codex", "agy", "tmux"}, "CLAUDE.md", "AGENTS.md", "GEMINI.md", ".claude/agents/lider.md")
	items := Preflight("/proj", cfg, env)
	assert.False(t, PreflightBlocked(items), FormatPreflight(items))
	for _, it := range items {
		assert.Equal(t, LevelOK, it.Level, "%+v", it)
	}
}

func TestPreflight_TmuxAusenteAvisaComComando(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Terminal = config.TerminalAuto
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIClaude, config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}
	env := penv([]string{"claude", "codex", "agy", "brew"}, "CLAUDE.md", "AGENTS.md", "GEMINI.md", ".claude/agents/lider.md")
	items := Preflight("/proj", cfg, env)
	assert.False(t, PreflightBlocked(items), "tmux ausente só avisa")
	var tmuxWarn *Item
	for i := range items {
		if strings.Contains(items[i].Message, "tmux ausente") {
			tmuxWarn = &items[i]
			break
		}
	}
	require.NotNil(t, tmuxWarn)
	assert.Equal(t, LevelWarn, tmuxWarn.Level)
	assert.Contains(t, tmuxWarn.Action, "instale")
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
	wantCalled := 3
	if runtime.GOOS != "windows" {
		wantCalled = 4 // 3 sessões + tmux
	}
	assert.Equal(t, wantCalled, called, "só LookPath, uma vez por sessão e uma para o tmux")
}

func TestPreflight_ConstituicaoRascunho(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "constitution.md"), []byte("# Constituição — Teste\n## Missão\n[Uma frase]\n"), 0o644))

	cfg := config.DefaultForgeConfig()
	cfg.Sessions = map[string]string{config.SessionRoleDiscovery: config.CLIClaude, config.SessionRoleBuilder: config.CLICodex, config.SessionRoleReviewer: config.CLIGemini}
	env := penvFor(dir, []string{"claude", "codex", "gemini", "agy", "tmux"}, "CLAUDE.md", "AGENTS.md", "GEMINI.md", ".claude/agents/lider.md")

	items := Preflight(dir, cfg, env)
	assert.False(t, PreflightBlocked(items), FormatPreflight(items))

	var constWarn *Item
	for i := range items {
		if strings.Contains(items[i].Message, "Constituição em rascunho") {
			constWarn = &items[i]
			break
		}
	}
	require.NotNil(t, constWarn, "deve gerar aviso de Constituição em rascunho")
	assert.Equal(t, LevelWarn, constWarn.Level)
	assert.Equal(t, "/constitution", constWarn.Action)

	// Agora com projeto pronta: não deve avisar
	dirPronta := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dirPronta, "sdd", "memory"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dirPronta, "sdd", "spec"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "memory", "constitution.md"), []byte("# Constituição — App\n## Missão\nMissao real.\n## Regras\n1. Regra.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "modules.md"), []byte("# Módulos\n## Core\nModulo core.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "flows.md"), []byte("# Fluxos\n## Fluxo 1\nFluxo real.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "decisions.md"), []byte("# Decisões\n\n## Decisões Resolvidas\n| ID | Decisão | Escolha |\n|---|---|---|\n| D01 | Banco | SQLite |\n"), 0o644))

	envPronta := penvFor(dirPronta, []string{"claude", "codex", "gemini", "agy", "tmux"}, "CLAUDE.md", "AGENTS.md", "GEMINI.md", ".claude/agents/lider.md")
	itemsPronta := Preflight(dirPronta, cfg, envPronta)
	assert.False(t, PreflightBlocked(itemsPronta), FormatPreflight(itemsPronta))
	for _, it := range itemsPronta {
		assert.NotContains(t, it.Message, "Constituição em rascunho")
	}
}

func TestPreflight_GeminiChecaBinarioAgy(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	// Simula forge.json legado onde o comando era "gemini"
	cfg.CLIProfiles[config.CLIGemini] = config.CLIProfile{
		Command: "gemini",
		Args:    []string{"-p", "{prompt}", "-o", "stream-json", config.MarkerPermissions},
	}
	cfg.Sessions = map[string]string{config.SessionRoleReviewer: config.CLIGemini}

	// Com agy no PATH (e sem o binário gemini): deve passar como OK
	envComAgy := penv([]string{"claude", "agy", "tmux"}, "CLAUDE.md", "GEMINI.md", ".claude/agents/lider.md", ".claude/agents/builder.md")
	items := Preflight("/proj", cfg, envComAgy)
	it := findItem(items, config.SessionRoleReviewer, LevelOK)
	require.NotNil(t, it, "deve reconhecer CLI agy para sessão gemini")
	assert.Contains(t, it.Message, "agy")

	// Sem agy no PATH: deve bloquear mesmo que haja um binário legado gemini
	envSemAgy := penv([]string{"claude", "gemini", "tmux"}, "CLAUDE.md", "GEMINI.md", ".claude/agents/lider.md", ".claude/agents/builder.md")
	itemsBloq := Preflight("/proj", cfg, envSemAgy)
	itBloq := findItem(itemsBloq, config.SessionRoleReviewer, LevelBlock)
	require.NotNil(t, itBloq, "sem agy no PATH deve bloquear")
	assert.Contains(t, itBloq.Message, "agy")
	assert.Contains(t, itBloq.Action, "@google/antigravity-cli")
}
