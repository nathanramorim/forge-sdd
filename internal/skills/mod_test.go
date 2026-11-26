package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func modValido() map[string][]byte {
	return map[string][]byte{
		".claude-plugin/plugin.json": []byte(`{"name":"meu-mod","version":"0.1.0","description":"Mod de teste"}`),
		".claude-plugin/forge.json": []byte(`{
			"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0",
			"permissions":{"network":false,"write":["sdd/.metrics"],"hooks":["session.measure"],"read":["sdd/.metrics"],"env":["FORGE_WORKER"],"model":false}
		}`),
		"hooks/hooks.json":   []byte(`{"modules":["./register.tsx"]}`),
		"hooks/register.tsx": []byte("export const register = () => {}"),
	}
}

// comForge troca o conteúdo de forge.json mantendo o resto do mod válido.
func comForge(forge string) map[string][]byte {
	a := modValido()
	a[".claude-plugin/forge.json"] = []byte(forge)
	return a
}

func TestInstalarModValidoPassa(t *testing.T) {
	require.NoError(t, ValidarMod("meu-mod", modValido(), "6.0.0-beta"))
}

func TestInstalarModDoRepositorioCumpreOContrato(t *testing.T) {
	raiz := filepath.Join("..", "..", ".agents", "plugins", "forge-painel")
	arquivos := map[string][]byte{}
	for _, rel := range []string{".claude-plugin/plugin.json", ".claude-plugin/forge.json", "hooks/hooks.json", "hooks/register.tsx"} {
		b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
		require.NoError(t, err)
		arquivos[rel] = b
	}
	assert.True(t, EhMod(arquivos))
	require.NoError(t, ValidarMod("forge-painel", arquivos, "6.0.0"))
}

func TestInstalarPluginSemModulosNaoEhMod(t *testing.T) {
	// plugin só de skills e comandos (diagram-studio) segue a ADR-0025, sem contrato
	arquivos := map[string][]byte{".claude-plugin/plugin.json": []byte("{}"), "scripts/diagrama.py": []byte("x")}
	assert.False(t, EhMod(arquivos))
	require.NoError(t, ValidarMod("diagram-studio", arquivos, "6.0.0"))
}

func TestInstalarModRecusaManifestoIncompleto(t *testing.T) {
	semForge := modValido()
	delete(semForge, ".claude-plugin/forge.json")

	semPlugin := modValido()
	delete(semPlugin, ".claude-plugin/plugin.json")

	semModulo := modValido()
	delete(semModulo, "hooks/register.tsx")

	nomeErrado := modValido()
	nomeErrado[".claude-plugin/plugin.json"] = []byte(`{"name":"outro","version":"0.1.0","description":"x"}`)

	versaoRuim := modValido()
	versaoRuim[".claude-plugin/plugin.json"] = []byte(`{"name":"meu-mod","version":"1","description":"x"}`)

	semDescricao := modValido()
	semDescricao[".claude-plugin/plugin.json"] = []byte(`{"name":"meu-mod","version":"0.1.0","description":" "}`)

	jsonRuim := modValido()
	jsonRuim[".claude-plugin/plugin.json"] = []byte(`{`)

	casos := map[string]struct {
		arquivos map[string][]byte
		trecho   string
	}{
		"sem forge.json":              {semForge, "forge.json ausente"},
		"sem plugin.json":             {semPlugin, "plugin.json ausente"},
		"módulo citado sem arquivo":   {semModulo, "que o plugin não traz"},
		"nome diferente do slug":      {nomeErrado, "deve ser igual ao slug"},
		"versão fora de X.Y.Z":        {versaoRuim, "deve ser X.Y.Z"},
		"sem descrição":               {semDescricao, "description é obrigatória"},
		"plugin.json inválido":        {jsonRuim, "plugin.json inválido"},
		"sem permissões":              {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0"}`), "permissions.network é obrigatório"},
		"sem versão mínima do claude": {comForge(`{"schema":"forge-sdd/mod/1.0","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`), "minClaudeCode"},
		"sem versão mínima do forge":  {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`), "minForge"},
		"sem ganchos":                 {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":[]}}`), "permissions.hooks é obrigatório"},
		"sem write":                   {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"hooks":["a.b"]}}`), "permissions.write é obrigatório"},
		"schema desconhecido":         {comForge(`{"schema":"x/9","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`), "schema"},
		"campo desconhecido":          {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"],"shell":true}}`), "unknown field"},
		"forge.json inválido":         {comForge(`{`), "forge.json inválido"},
		"minForge acima do forge":     {comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"9.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`), "exige forge 9.0.0"},
	}
	for nome, c := range casos {
		t.Run(nome, func(t *testing.T) {
			err := ValidarMod("meu-mod", c.arquivos, "6.0.0-beta")
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.trecho)
		})
	}
}

func TestInstalarModRecusaCaminhoForaDoProjeto(t *testing.T) {
	for _, caminho := range []string{"/etc", "~/.claude", "../fora", "sdd/../../x", ".claude/settings.json", ".git/hooks", ".agents/skills", "C:/x", "", "sdd//x"} {
		bruto, err := json.Marshal(map[string]any{
			"schema": ModSchema, "minClaudeCode": "2.1.331", "minForge": "6.0.0",
			"permissions": map[string]any{"network": false, "write": []string{caminho}, "hooks": []string{"a.b"}},
		})
		require.NoError(t, err)
		err = ValidarMod("meu-mod", comForge(string(bruto)), "6.0.0")
		require.Error(t, err, "write %q deveria ser recusado", caminho)
		assert.Contains(t, err.Error(), "permissions.write")
	}
	// a leitura segue a mesma regra
	err := ValidarMod("meu-mod", comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":[],"read":["../x"],"hooks":["a.b"]}}`), "6.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permissions.read")
}

func TestInstalarModVersaoDeDesenvolvimentoNaoCompara(t *testing.T) {
	a := comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"9.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"]}}`)
	require.NoError(t, ValidarMod("meu-mod", a, "dev"))
}

func TestInstalarModJuntaTodosOsProblemas(t *testing.T) {
	a := comForge(`{"schema":"x","permissions":{}}`)
	err := ValidarMod("meu-mod", a, "6.0.0")
	require.Error(t, err)
	for _, trecho := range []string{"schema", "minClaudeCode", "minForge", "network", "write", "hooks"} {
		assert.Contains(t, err.Error(), trecho)
	}
}

func TestInstalarModVariavelDeAmbienteInvalida(t *testing.T) {
	a := comForge(`{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"6.0.0","permissions":{"network":false,"write":[],"hooks":["a.b"],"env":["minha var"]}}`)
	err := ValidarMod("meu-mod", a, "6.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permissions.env")
}
