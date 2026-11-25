package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func modArquivos() map[string][]byte {
	return map[string][]byte{
		".claude-plugin/plugin.json": []byte(`{"name":"meu-mod","version":"0.1.0","description":"Mod de teste"}`),
		".claude-plugin/forge.json": []byte(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0",
			"permissions":{"network":false,"write":["sdd/.metrics"],"hooks":["session.measure"]}}`),
		"hooks/hooks.json":   []byte(`{"modules":["./register.tsx"]}`),
		"hooks/register.tsx": []byte("export const register = () => {}"),
	}
}

func TestInstalaModGravaSoEmAgentsPlugins(t *testing.T) {
	home := isolaHome(t)
	p := novoForge(t, t.TempDir())

	res, err := InstalarPlugin(p, p, "meu-mod", modArquivos())
	require.NoError(t, err)
	assert.Len(t, res.Gravados, 4)
	for _, g := range res.Gravados {
		assert.Contains(t, g, filepath.Join(p, ".agents", "plugins", "meu-mod")+string(filepath.Separator))
	}
	assert.NoDirExists(t, filepath.Join(home, ".claude"))
	assert.NoDirExists(t, filepath.Join(p, ".claude"))
}

func TestInstalaModSemContratoNaoGravaNada(t *testing.T) {
	p := novoForge(t, t.TempDir())
	a := modArquivos()
	delete(a, ".claude-plugin/forge.json")

	_, err := InstalarPlugin(p, p, "meu-mod", a)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forge.json ausente")
	assert.NoDirExists(t, filepath.Join(p, ".agents"), "recusa antes de gravar o primeiro arquivo")
}

func TestInstalaModSemPermissoesOuVersaoMinimaEhRecusado(t *testing.T) {
	for nome, forge := range map[string]string{
		"sem permissões":          `{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0"}`,
		"sem versão do claude":    `{"schema":"forge-sdd/mod/1.0","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`,
		"sem versão do forge":     `{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`,
		"escrita fora do projeto": `{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":["~/.claude"],"hooks":["a.b"]}}`,
	} {
		t.Run(nome, func(t *testing.T) {
			p := novoForge(t, t.TempDir())
			a := modArquivos()
			a[".claude-plugin/forge.json"] = []byte(forge)
			_, err := InstalarPlugin(p, p, "meu-mod", a)
			require.Error(t, err)
			assert.NoDirExists(t, filepath.Join(p, ".agents"))
		})
	}
}

func TestInstalaPluginSemModulosContinuaSemContrato(t *testing.T) {
	// ADR-0030: só quem traz módulo de ganchos ou forge.json é mod; o diagram-studio segue como estava
	p := novoForge(t, t.TempDir())
	_, err := InstalarPlugin(p, p, "diagram-studio", pluginFiles())
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(p, ".agents", "plugins", "diagram-studio", ".claude-plugin", "plugin.json"))
}

func TestInstalaModDoRepositorioPeloLerOrigem(t *testing.T) {
	origem := filepath.Join("..", "..", ".agents", "plugins", "forge-painel")
	if _, err := os.Stat(origem); err != nil {
		t.Skip("mod de referência ausente")
	}
	arquivos, err := LerOrigemInstalacao(origem)
	require.NoError(t, err)

	p := novoForge(t, t.TempDir())
	res, err := InstalarPlugin(p, p, "forge-painel", arquivos)
	require.NoError(t, err)
	assert.NotEmpty(t, res.Gravados)
	assert.FileExists(t, filepath.Join(p, ".agents", "plugins", "forge-painel", ".claude-plugin", "forge.json"))
}
