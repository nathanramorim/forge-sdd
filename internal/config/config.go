package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Agentes suportados.
const (
	AgentCopilot = "copilot"
	AgentClaude  = "claude"
	AgentGemini  = "gemini"
)

var validAgents = map[string]bool{
	AgentCopilot: true,
	AgentClaude:  true,
	AgentGemini:  true,
}

// ParseAgents converte uma string csv ("copilot,claude") em []string validado.
func ParseAgents(csv string) ([]string, error) {
	if csv == "" {
		return []string{AgentCopilot}, nil
	}
	parts := strings.Split(csv, ",")
	seen := make(map[string]bool, len(parts))
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if !validAgents[p] {
			return nil, fmt.Errorf("agente desconhecido: %q (válidos: copilot, claude, gemini)", p)
		}
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result, nil
}

// Config contém todas as configurações coletadas durante o init.
type Config struct {
	Project    string   // nome do projeto
	Stack      string   // runtime principal: go, node, python, rust, other
	DB         string   // banco de dados: postgres, sqlite, mongo, none
	Telemetry  bool     // habilitar telemetria local
	Lang       string   // idioma dos templates: pt-BR, en
	SddVersion string   // versão Forge-SDD a usar
	DryRun     bool     // exibir árvore sem criar arquivos
	Agents     []string // agentes de IA: copilot, claude, gemini
}

// Defaults retorna uma Config com valores padrão.
func Defaults() Config {
	return Config{
		Project:    "meu-projeto",
		Stack:      "go",
		DB:         "none",
		Telemetry:  true,
		Lang:       "pt-BR",
		SddVersion: "1.3.2",
		DryRun:     false,
		Agents:     []string{AgentCopilot},
	}
}

// FromFlags preenche a Config a partir das flags registradas no cobra.Command.
// Começa com Defaults() e sobrescreve apenas as flags que foram explicitamente passadas.
func FromFlags(cmd *cobra.Command) Config {
	cfg := Defaults()
	if f := cmd.Flags().Lookup("name"); f != nil && f.Changed {
		cfg.Project = f.Value.String()
	}
	if f := cmd.Flags().Lookup("stack"); f != nil && f.Changed {
		cfg.Stack = f.Value.String()
	}
	if f := cmd.Flags().Lookup("db"); f != nil && f.Changed {
		cfg.DB = f.Value.String()
	}
	if f := cmd.Flags().Lookup("lang"); f != nil && f.Changed {
		cfg.Lang = f.Value.String()
	}
	if f := cmd.Flags().Lookup("version"); f != nil && f.Changed {
		cfg.SddVersion = f.Value.String()
	}
	if f := cmd.Flags().Lookup("no-telemetry"); f != nil && f.Changed {
		v, _ := cmd.Flags().GetBool("no-telemetry")
		cfg.Telemetry = !v
	}
	if f := cmd.Flags().Lookup("agent"); f != nil && f.Changed {
		agents, err := ParseAgents(f.Value.String())
		if err != nil {
			// erro de validação — mantém default; será relatado em RunE
			cfg.Agents = []string{AgentCopilot}
		} else {
			cfg.Agents = agents
		}
	}
	return cfg
}
