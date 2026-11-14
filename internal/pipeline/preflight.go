package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
)

// Nível de um item da checagem prévia.
const (
	LevelOK    = "ok"
	LevelWarn  = "aviso"
	LevelBlock = "bloqueio"
)

// Item é o resultado de uma verificação, no formato do doctor: ok, aviso ou bloqueio.
type Item struct {
	Session string // papel (discovery, builder, reviewer) ou vazio para o projeto
	Agent   string
	Level   string
	Message string
	Action  string // o que fazer, quando houver
}

// PreflightEnv é o ambiente da checagem; injetável nos testes. Nada aqui executa a CLI do agente.
type PreflightEnv struct {
	LookPath func(string) (string, error)
	Exists   func(string) bool
}

// DefaultPreflightEnv usa o PATH e o sistema de arquivos reais.
func DefaultPreflightEnv() PreflightEnv {
	return PreflightEnv{
		LookPath: exec.LookPath,
		Exists:   func(p string) bool { _, err := os.Stat(p); return err == nil },
	}
}

// agentFiles são os arquivos de instrução que cada CLI lê (documentação oficial) e o agente do
// scaffold que os gera (`forge-sdd update --agent <x>`).
var agentFiles = map[string]struct {
	Files     []string // qualquer um basta
	UpdateAs  string
	InstallIt string
}{
	config.CLIClaude:  {[]string{"CLAUDE.md"}, "claude", "npm install -g @anthropic-ai/claude-code"},
	config.CLIGemini:  {[]string{"GEMINI.md"}, "gemini", "npm install -g @google/antigravity-cli"},
	config.CLICopilot: {[]string{".github/copilot-instructions.md", "AGENTS.md"}, "copilot", "npm install -g @github/copilot"},
	config.CLICodex:   {[]string{"AGENTS.md"}, "openai", "npm install -g @openai/codex"},
	config.CLIAgy:     {[]string{"GEMINI.md", "AGENTS.md"}, "gemini", ""},
}

// Preflight confere, antes de abrir as sessões, se cada agente escolhido tem a CLI no PATH, os
// arquivos de instrução no projeto e, para o claude, as definições de agente que o worker usa.
// CLI ausente e configuração inválida bloqueiam; arquivo ausente avisa e indica o comando.
func Preflight(dir string, cfg config.ForgeConfig, env PreflightEnv) []Item {
	if env.LookPath == nil || env.Exists == nil {
		d := DefaultPreflightEnv()
		if env.LookPath == nil {
			env.LookPath = d.LookPath
		}
		if env.Exists == nil {
			env.Exists = d.Exists
		}
	}
	var items []Item
	if err := cfg.Validate(); err != nil {
		return []Item{{Level: LevelBlock, Message: "forge.json inválido: " + firstLine(err.Error()), Action: "corrija .forge/config/forge.json ou rode forge-sdd pipeline init"}}
	}
	seen := map[string]bool{}
	for _, role := range config.SessionRoles() {
		cli := cfg.CLIFor(role)
		profile, _ := cfg.Profile(cli)

		if _, err := env.LookPath(profile.Command); err != nil {
			item := Item{Session: role, Agent: cli, Level: LevelBlock, Message: fmt.Sprintf("a CLI %q não está no PATH", profile.Command)}
			if hint := agentFiles[cli].InstallIt; hint != "" {
				item.Action = "instale (confirme na documentação da CLI): " + hint
			} else {
				item.Action = "instale a CLI ou escolha outro agente"
			}
			items = append(items, item)
		} else {
			items = append(items, Item{Session: role, Agent: cli, Level: LevelOK, Message: "CLI " + profile.Command + " encontrada"})
		}

		if seen[cli] {
			continue
		}
		seen[cli] = true
		info := agentFiles[cli]
		found := false
		for _, f := range info.Files {
			if env.Exists(filepath.Join(dir, filepath.FromSlash(f))) {
				found = true
				break
			}
		}
		if !found {
			items = append(items, Item{
				Session: role, Agent: cli, Level: LevelWarn,
				Message: fmt.Sprintf("falta o arquivo de instruções que o %s lê (%s)", cli, strings.Join(info.Files, " ou ")),
				Action:  "forge-sdd update --agent " + info.UpdateAs,
			})
		}
		if cli == config.CLIClaude {
			for _, ag := range claudeAgentsFor(cfg) {
				if !env.Exists(filepath.Join(dir, ".claude", "agents", ag+".md")) {
					items = append(items, Item{Session: role, Agent: cli, Level: LevelBlock,
						Message: fmt.Sprintf("falta .claude/agents/%s.md, que a sessão usa com --agent", ag),
						Action:  "forge-sdd update --agent claude"})
				}
			}
		}
	}

	if cfg.Terminal != config.TerminalNative && runtime.GOOS != "windows" {
		if _, err := env.LookPath("tmux"); err != nil {
			ienv := InstallEnv{GOOS: runtime.GOOS, LookPath: env.LookPath}
			action := "instale o tmux ou use --terminal native"
			if argv, _, ok := InstallPlan(ienv); ok {
				action = "instale: " + InstallCommandLine(argv) + " (ou use --terminal native)"
			}
			items = append(items, Item{
				Level:   LevelWarn,
				Message: "tmux ausente: as sessões abrirão em janelas soltas",
				Action:  action,
			})
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "sdd", "memory", "constitution.md")); err == nil {
		if b, err := gate.BuildBrief(dir, 0); err == nil && b.State == gate.BriefRascunho {
			msg := "Constituição em rascunho"
			if len(b.Missing) > 0 {
				msg += ": " + strings.Join(b.Missing, ", ")
			}
			items = append(items, Item{
				Level:   LevelWarn,
				Message: msg,
				Action:  "/constitution",
			})
		}
	}
	return items
}

// claudeAgentsFor lista as definições de agente do claude usadas pelas sessões que o escolheram.
func claudeAgentsFor(cfg config.ForgeConfig) []string {
	var out []string
	for _, pair := range []struct{ role, agent string }{
		{config.SessionRoleDiscovery, "lider"}, {config.SessionRoleBuilder, layout.RoleBuilder}, {config.SessionRoleReviewer, layout.RoleReviewer},
	} {
		if cfg.CLIFor(pair.role) == config.CLIClaude {
			out = append(out, pair.agent)
		}
	}
	return out
}

// PreflightBlocked diz se algum item impede abrir as sessões.
func PreflightBlocked(items []Item) bool {
	for _, it := range items {
		if it.Level == LevelBlock {
			return true
		}
	}
	return false
}

// FormatPreflight resume os itens em linhas curtas para o terminal e para o menu.
func FormatPreflight(items []Item) string {
	var b strings.Builder
	for _, it := range items {
		mark := map[string]string{LevelOK: "✓", LevelWarn: "!", LevelBlock: "✗"}[it.Level]
		who := it.Session
		if it.Agent != "" {
			who += "/" + it.Agent
		}
		if who != "" {
			who += ": "
		}
		fmt.Fprintf(&b, "  %s %s%s\n", mark, who, it.Message)
		if it.Action != "" {
			fmt.Fprintf(&b, "      → %s\n", it.Action)
		}
	}
	return b.String()
}
