package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var referenciasConcepcao = []string{"discovery", "nova-feature", "novo-fix"}

func lerReferenciaConcepcao(t *testing.T, nome string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("templates", ".agents", "skills", "sdd-concepcao", "references", nome+".md.tmpl"))
	require.NoError(t, err)
	return string(b)
}

func TestReferencesConcepcaoCitamForgeQueueCard(t *testing.T) {
	for _, n := range referenciasConcepcao {
		s := lerReferenciaConcepcao(t, n)
		require.Contains(t, s, "forge queue card", n)
		require.Contains(t, s, "Gestor desligado", n)
		require.Contains(t, s, "nada a fazer", n)
	}
}

func TestReferencesConcepcaoNaoPressupoemForgeUp(t *testing.T) {
	for _, n := range referenciasConcepcao {
		require.False(t, strings.Contains(lerReferenciaConcepcao(t, n), "forge up"), n)
	}
}
