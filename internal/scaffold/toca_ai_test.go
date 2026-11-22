package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTocaAIRamoChatSemWorker(t *testing.T) {
	dir := renderAllAgents(t)
	b, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "sdd-operacao", "references", "toca-ai.md"))
	require.NoError(t, err)
	s := string(b)
	require.Contains(t, s, "Com workers ativos")
	require.Contains(t, s, "Sem workers")
	require.Contains(t, s, "subagentes desta sessão")
	require.Contains(t, s, "Nunca sugira `forge up`")
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, "forge up") {
			require.Contains(t, l, "Nunca", "a skill não pode mandar iniciar forge up: %s", l)
		}
	}
	r, err := os.ReadFile(filepath.Join(dir, ".agents", "rules", "sdd-fluxo-e-portao.md"))
	require.NoError(t, err)
	require.Contains(t, string(r), "Não presuma `forge up`")
}
