package pipeline

import (
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAgentUnavailable_PadroesFortes(t *testing.T) {
	cases := []struct {
		name    string
		cli     string
		lines   []string
		wantMsg string
		wantHnt string
	}{
		{
			name: "gemini ineligible tier error do relato real",
			cli:  config.CLIGemini,
			lines: []string{
				"iniciando sessão",
				"IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals. ... migrate to the Antigravity suite of products",
				"detalhes adicionais",
			},
			wantMsg: "IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals. ... migrate to the Antigravity suite of products",
			wantHnt: "o plano gratuito do Gemini Code Assist foi descontinuado para este cliente: troque o agente desta sessão (forge → Configurar) ou use o agy (Antigravity)",
		},
		{
			name: "claude not logged in",
			cli:  config.CLIClaude,
			lines: []string{
				"carregando projeto",
				"✗ You are not logged in. Please log in first.",
			},
			wantMsg: "You are not logged in. Please log in first.",
			wantHnt: "rode claude login",
		},
		{
			name: "copilot please authenticate",
			cli:  config.CLICopilot,
			lines: []string{
				"please authenticate using gh auth login",
			},
			wantMsg: "please authenticate using gh auth login",
			wantHnt: "rode gh auth login",
		},
		{
			name: "codex invalid api key",
			cli:  config.CLICodex,
			lines: []string{
				"connecting to api",
				"invalid api key provided",
			},
			wantMsg: "invalid api key provided",
			wantHnt: "rode codex login",
		},
		{
			name: "agy authentication failed",
			cli:  config.CLIAgy,
			lines: []string{
				"authentication failed",
			},
			wantMsg: "authentication failed",
			wantHnt: "entre pelo agy e tente de novo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			au, ok := DetectAgentUnavailable(tc.lines, tc.cli)
			require.True(t, ok)
			assert.Equal(t, tc.wantMsg, au.Message)
			assert.Equal(t, tc.wantHnt, au.Hint)
		})
	}
}

func TestDetectAgentUnavailable_PadraoGenerico(t *testing.T) {
	// Linha genérica na última posição com menção a token/credential/api casa
	lastLineGeneric := []string{
		"passo 1",
		"error: 401 unauthorized token",
	}
	au, ok := DetectAgentUnavailable(lastLineGeneric, "custom")
	require.True(t, ok)
	assert.Contains(t, au.Message, "401 unauthorized token")
	assert.Equal(t, "revise o login da CLI ou troque o agente desta sessão", au.Hint)

	// Texto de código ou conversa com 401 no meio da saída NÃO casa
	codeInMiddle := []string{
		"const code = 401;",
		"return res.status(code);",
		"processamento concluído",
	}
	_, ok = DetectAgentUnavailable(codeInMiddle, config.CLIClaude)
	assert.False(t, ok, "401 no meio do código sem marcador de erro não deve casar")
}

func TestDetectAgentUnavailable_MascaraSegredo(t *testing.T) {
	fakeToken := "ghp_" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	lines := []string{
		"authentication error: token " + fakeToken + " invalid",
	}
	au, ok := DetectAgentUnavailable(lines, config.CLICopilot)
	require.True(t, ok)
	assert.NotContains(t, au.Message, fakeToken)
	assert.Contains(t, au.Message, "***")
}
