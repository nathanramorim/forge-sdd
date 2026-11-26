package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNuvem_ValidaCamposQueVaoParaOScript(t *testing.T) {
	assert.NoError(t, NuvemConfig{Ativo: true, Repos: []string{"api", "dono/web"}, Instalar: "git", Ref: "claude/x-1"}.Validate())
	for nome, n := range map[string]NuvemConfig{
		"instalar desconhecido": {Instalar: "curl"},
		"ref com shell":         {Ref: "main;rm -rf /"},
		"ref com espaço":        {Ref: "a b"},
		"repo com ..":           {Repos: []string{".."}},
		"repo com barra dupla":  {Repos: []string{"a/b/c"}},
		"repo com $":            {Repos: []string{"$(x)"}},
	} {
		assert.Error(t, n.Validate(), nome)
	}
	assert.Equal(t, NuvemInstalarNPM, NuvemConfig{}.InstalarEfetivo())
}

func TestNuvem_ForgeJSONSemChaveNaoGravaNuvem(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, SaveForgeConfig(dir, DefaultForgeConfig()))
	raw, err := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"nuvem"`)

	cfg := DefaultForgeConfig()
	cfg.Nuvem = &NuvemConfig{Ativo: true, Repos: []string{"api"}}
	require.NoError(t, SaveForgeConfig(dir, cfg))
	got, _, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, got.NuvemAtiva())
	assert.Equal(t, []string{"api"}, got.NuvemOuPadrao().Repos)
}

func TestPerfilClaudeDa41_MigraParaStreamJSONEPermissoes(t *testing.T) {
	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"cli":"claude","permissions":"full","cli_profiles":{"claude":{"command":"claude","args":["-p","{prompt}","--agent","{agent}","-n","{name}"]}}}`)
	cfg, _, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	p := cfg.CLIProfiles[CLIClaude]
	assert.Contains(t, p.Args, "stream-json")
	assert.Contains(t, p.Args, MarkerPermissions)
	raw, err := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), "stream-json", "a migração fica gravada")

	// argv personalizado não é tocado
	dir = t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"cli":"claude","cli_profiles":{"claude":{"command":"claude","args":["-p","{prompt}","--verbose"]}}}`)
	cfg, _, err = LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"-p", "{prompt}", "--verbose"}, cfg.CLIProfiles[CLIClaude].Args)
}
