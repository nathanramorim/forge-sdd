package scaffold

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/forge-sdd/cli/internal/config"
)

// commandOrder define a ordem pedagógica dos comandos SDD exibidos no cheat-sheet,
// da configuração inicial até o uso ad-hoc de utilitários.
var commandOrder = []string{
	"constitution",
	"tutorial",
	"discovery",
	"split-features",
	"nova-feature",
	"novo-fix",
	"implementar",
	"proxima-feature",
	"revisar",
	"status",
	"archive",
	"doctor",
	"upgrade-sdd",
	"c4-architecture",
	"inbox",
	"implement",
}

// CommandInfo descreve um comando SDD para exibição no cheat-sheet do CLI.
type CommandInfo struct {
	Command     string
	Description string
}

// agentPromptDir retorna o diretório embutido onde os prompts de slash-command
// de um agente residem. Retorna "" para agentes desconhecidos.
func agentPromptDir(agent string) string {
	switch agent {
	case config.AgentClaude:
		return "templates/agents/claude/.claude/commands"
	case config.AgentGemini:
		return "templates/agents/gemini/.gemini/prompts"
	case config.AgentOpenAI:
		return "templates/agents/openai/.openai/prompts"
	case config.AgentCopilot:
		return "templates/.github/prompts"
	default:
		return ""
	}
}

// agentPromptSuffix retorna o sufixo de nome de arquivo usado pelos templates de
// slash-command de um agente. O Claude Code descobre comandos pelo nome do
// arquivo sem ".md" — por isso usa só ".md.tmpl", sem o infixo ".prompt" que os
// demais agentes mantêm por convenção própria (Copilot exige ".prompt.md").
func agentPromptSuffix(agent string) string {
	if agent == config.AgentClaude {
		return ".md.tmpl"
	}
	return ".prompt.md.tmpl"
}

// aliasCommands são comandos mantidos só por compatibilidade: têm adaptador, mas não
// aparecem no cheat-sheet (o comando canônico já aparece).
var aliasCommands = map[string]bool{"proxima-feature": true, "implement": true}

// commandSkill mapeia cada comando à skill de estação dona da sua referência canônica
// (.agents/skills/<skill>/references/<comando>.md), onde a descrição real ("**Uso:**") vive.
// Os adaptadores por agente só redirecionam para lá.
var commandSkill = map[string]string{
	"constitution":    "sdd-concepcao",
	"discovery":       "sdd-concepcao",
	"nova-feature":    "sdd-concepcao",
	"novo-fix":        "sdd-concepcao",
	"c4-architecture": "sdd-concepcao",
	"split-features":  "sdd-planejamento",
	"implementar":     "sdd-build",
	"revisar":         "sdd-governanca",
	"status":          "sdd-operacao",
	"doctor":          "sdd-operacao",
	"archive":         "sdd-operacao",
	"upgrade-sdd":     "sdd-operacao",
	"tutorial":        "sdd-operacao",
	"inbox":           "sdd-operacao",
}

// canonicalCommandPath devolve o caminho embutido da referência canônica de um comando.
func canonicalCommandPath(cmdName string) (string, bool) {
	skill, ok := commandSkill[cmdName]
	if !ok {
		return "", false
	}
	return "templates/.agents/skills/" + skill + "/references/" + cmdName + ".md.tmpl", true
}

var frontMatterDescRe = regexp.MustCompile(`(?m)^description:\s*"?([^"\n]+)"?\s*$`)
var usoLineRe = regexp.MustCompile(`^\*\*Uso:\*\*\s*(.+)$`)
var labelOnlyLineRe = regexp.MustCompile(`^\*\*[^*:]+:\*\*\s*$`)
var listMarkerRe = regexp.MustCompile(`^\d+\.\s+`)

const descMaxLen = 100

// extractDescription deriva uma descrição de uma linha do próprio prompt.
// Prioridade: (1) campo "description:" do front-matter YAML (formato Copilot),
// (2) a linha "**Uso:** ..." (formato Claude/Gemini), (3) a primeira linha com
// conteúdo substantivo (pulando títulos markdown e rótulos em negrito sem
// texto, ex: "**Ação:**").
func extractDescription(content string) string {
	if strings.HasPrefix(content, "---") {
		if end := strings.Index(content[3:], "---"); end != -1 {
			frontMatter := content[:end+3]
			if m := frontMatterDescRe.FindStringSubmatch(frontMatter); m != nil {
				return truncateDescription(cleanLine(m[1]))
			}
		}
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := usoLineRe.FindStringSubmatch(line); m != nil {
			return truncateDescription(cleanLine(m[1]))
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" || strings.HasPrefix(line, "#") || labelOnlyLineRe.MatchString(line) {
			continue
		}
		return truncateDescription(cleanLine(line))
	}
	return ""
}

// cleanLine remove marcações markdown que poluem a descrição de uma linha
// (numeração de lista e ênfase em negrito), preservando o texto legível.
func cleanLine(s string) string {
	s = listMarkerRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "**", "")
	return strings.TrimSpace(s)
}

// truncateDescription corta no limite de runas na última quebra de palavra,
// evitando cortar uma palavra ao meio.
func truncateDescription(s string) string {
	runes := []rune(s)
	if len(runes) <= descMaxLen {
		return s
	}
	cut := runes[:descMaxLen]
	if idx := strings.LastIndex(string(cut), " "); idx > 0 {
		cut = []rune(string(cut)[:idx])
	}
	return string(cut) + "…"
}

// CommandCheatSheet retorna, na ordem pedagógica do ciclo SDD, os comandos
// disponíveis para os agentes selecionados em cfg, com descrição extraída dos
// próprios templates embutidos (fonte única da verdade — sem duplicar texto).
// Os templates são renderizados com cfg antes da extração, para que placeholders
// como {{.Project}} apareçam já interpolados na descrição.
func CommandCheatSheet(cfg config.Config) []CommandInfo {
	agents := cfg.Agents
	if len(agents) == 0 {
		agents = []string{config.AgentCopilot}
	}

	var result []CommandInfo
	for _, cmdName := range commandOrder {
		if aliasCommands[cmdName] {
			continue
		}
		for _, agent := range agents {
			dir := agentPromptDir(agent)
			if dir == "" {
				continue
			}
			if _, err := templatesFS.ReadFile(dir + "/" + cmdName + agentPromptSuffix(agent)); err != nil {
				continue
			}
			canonical, ok := canonicalCommandPath(cmdName)
			if !ok {
				continue
			}
			data, err := templatesFS.ReadFile(canonical)
			if err != nil {
				continue
			}
			result = append(result, CommandInfo{
				Command:     "/" + cmdName,
				Description: extractDescription(renderCommandTemplate(cmdName, string(data), cfg)),
			})
			break
		}
	}
	return result
}

// renderCommandTemplate executa o template do prompt com cfg, igual ao pipeline
// de scaffold real (renderDir), para que a descrição reflita o texto final visto
// pelo usuário. Em caso de erro de parse/execução, devolve o conteúdo bruto.
func renderCommandTemplate(name, content string, cfg config.Config) string {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return content
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return content
	}
	return buf.String()
}
