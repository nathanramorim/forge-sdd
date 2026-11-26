package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A cópia embutida é idêntica à fonte de desenvolvimento em .agents/plugins (sem divergência).
func TestModsEmbutidosIguaisAoRepositorio(t *testing.T) {
	raiz := filepath.Join("..", "..", ".agents", "plugins")
	for _, slug := range ModsEmbutidos() {
		embutido, ok := ModEmbutido(slug)
		require.True(t, ok, slug)
		fonte, err := LerOrigemInstalacao(filepath.Join(raiz, slug))
		require.NoError(t, err)
		require.Equal(t, len(fonte), len(embutido), "%s: rode cp -r .agents/plugins/%s internal/scaffold/mods/", slug, slug)
		for rel, data := range fonte {
			assert.Equal(t, string(data), string(embutido[rel]), "%s/%s divergiu da fonte", slug, rel)
		}
	}
}

func TestModEmbutidoInstalaSemOrigem(t *testing.T) {
	assert.Contains(t, ModsEmbutidos(), "forge-painel")
	arq, ok := ModEmbutido("forge-painel")
	require.True(t, ok)
	assert.Contains(t, arq, ".claude-plugin/forge.json")
	proj := t.TempDir()
	res, err := InstalarPlugin(proj, proj, "forge-painel", arq)
	require.NoError(t, err)
	assert.NotEmpty(t, res.Gravados)
	_, err = os.Stat(filepath.Join(proj, ".agents", "plugins", "forge-painel", ".claude-plugin", "forge.json"))
	assert.NoError(t, err)

	_, ok = ModEmbutido("../etc")
	assert.False(t, ok, "slug hostil")
	_, ok = ModEmbutido("nao-existe")
	assert.False(t, ok)
}
