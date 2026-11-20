package config

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// agentModelFiles liga cada papel de agent_models aos agentes de .claude/agents que herdam o modelo
// dele (os subagentes do papel herdam, ADR-0015). Agentes fora da tabela seguem em `model: inherit`.
var agentModelFiles = map[string][]string{
	SessionRoleDiscovery: {"pm-architect", "planner"},
	SessionRoleBuilder:   {"builder"},
	SessionRoleReviewer:  {"reviewer", "governance"},
}

// AgentModelFiles devolve os nomes dos agentes (sem .md) que recebem o modelo do papel, em ordem estável.
func AgentModelFiles(role string) []string {
	return append([]string(nil), agentModelFiles[role]...)
}

// validateAgentModels valida agent_models na borda (SEG-01): papel conhecido e modelo no formato aceito
// por ValidModel (não vazio, sem caractere de controle, até 80). Não depende da CLI das sessões: o
// cabeçalho é lido pelo Claude Code, que aceita alias (haiku, sonnet, opus, inherit) ou ID completo.
func (c ForgeConfig) validateAgentModels() error {
	roles := make([]string, 0, len(c.AgentModels))
	for role := range c.AgentModels {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		if !contains(SessionRoles(), role) {
			return fmt.Errorf("agent_models: papel %q desconhecido (use discovery, builder ou reviewer)", role)
		}
		v := c.AgentModels[role]
		if !ValidModel(v) {
			return fmt.Errorf("agent_models.%s: %q inválido (use letras, números, ponto, dois-pontos, hífen e sublinhado, de 1 a 80)", role, v)
		}
		if reason := yamlUnsafe(v); reason != "" {
			return fmt.Errorf("agent_models.%s: %q %s no cabeçalho YAML do agente", role, v, reason)
		}
	}
	return nil
}

var yamlReserved = map[string]bool{"null": true, "true": true, "false": true, "yes": true, "no": true, "on": true, "off": true, "y": true, "n": true, "~": true}

// yamlUnsafe diz por que o valor não é lido como texto no cabeçalho YAML (vazio quando é seguro). Vale só
// para agent_models; ValidModel segue como está para os outros usos.
func yamlUnsafe(v string) string {
	switch {
	case !unicode.IsLetter([]rune(v)[0]):
		return "não começa por letra e seria lido como outro tipo ou erro"
	case strings.HasSuffix(v, ":") || strings.HasSuffix(v, "-"):
		return "termina em ':' ou '-' e quebraria a leitura"
	case yamlReserved[strings.ToLower(v)]:
		return "é um valor reservado do YAML (nulo ou booleano)"
	}
	return ""
}
