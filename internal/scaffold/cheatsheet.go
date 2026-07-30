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
	"proxima-feature",
	"revisar",
	"status",
	"archive",
	"doctor",
	"upgrade-sdd",
	"c4-architecture",
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
		for _, agent := range agents {
			dir := agentPromptDir(agent)
			if dir == "" {
				continue
			}
			data, err := templatesFS.ReadFile(dir + "/" + cmdName + ".prompt.md.tmpl")
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
