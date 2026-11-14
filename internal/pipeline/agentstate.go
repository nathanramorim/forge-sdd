package pipeline

import (
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/gate"
)

var (
	// Mensagens inequívocas de agente indisponível (autenticação, conta, plano): valem em qualquer das últimas 12 linhas.
	agentUnavailableStrongRe = regexp.MustCompile(`(?i)(IneligibleTierError|no longer supported for|migrate to the Antigravity|not logged in|please (?:log ?in|run .* login|authenticate)|invalid api key|authentication (?:failed|error|required))`)

	// Genéricas: só valem na última linha útil ou numa linha com marcador de erro, acompanhadas de api/token/credential.
	agentUnavailableGenericRe = regexp.MustCompile(`(?i)(\b401\b|unauthorized|forbidden|permission[_ ]denied)`)
	credentialMarkerRe       = regexp.MustCompile(`(?i)(api|token|credential)`)
	unavailableErrorMarkerRe = regexp.MustCompile(`(?i)(✗|\berror\b|api error|fatal)`)
)

// AgentUnavailable descreve uma falha de ambiente, autenticação ou plano de um agente.
type AgentUnavailable struct {
	Message string
	Hint    string
}

// DetectAgentUnavailable confere se as últimas linhas da saída indicam agente indisponível.
// Só vale com saída diferente de zero (o chamador verifica o código de saída).
func DetectAgentUnavailable(tail []string, cli string) (AgentUnavailable, bool) {
	for i := len(tail) - 1; i >= 0; i-- {
		line := tail[i]
		strong := agentUnavailableStrongRe.MatchString(line)
		generic := agentUnavailableGenericRe.MatchString(line) && credentialMarkerRe.MatchString(line) && (i == len(tail)-1 || unavailableErrorMarkerRe.MatchString(line))
		if !strong && !generic {
			continue
		}
		msg := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "✗"))
		msg = gate.MaskSecrets(msg)
		return AgentUnavailable{
			Message: msg,
			Hint:    HintForCLI(cli),
		}, true
	}
	return AgentUnavailable{}, false
}

// HintForCLI devolve a instrução de ação apropriada para o agente indisponível.
func HintForCLI(cli string) string {
	switch strings.ToLower(cli) {
	case "gemini":
		return "o plano gratuito do Gemini Code Assist foi descontinuado para este cliente: troque o agente desta sessão (forge → Configurar) ou use o agy (Antigravity)"
	case "claude":
		return "rode claude login"
	case "copilot":
		return "rode gh auth login"
	case "codex":
		return "rode codex login"
	case "agy":
		return "entre pelo agy e tente de novo"
	default:
		return "revise o login da CLI ou troque o agente desta sessão"
	}
}
