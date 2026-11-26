package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func forgeEm(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte("{}"), 0o644))
}

func rodaSkills(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(append([]string{"skills"}, args...))
	err := rootCmd.Execute()
	for _, c := range [][2]string{{"tipo", ""}, {"slug", ""}} {
		_ = skillsDestinosCmd.Flags().Set(c[0], c[1])
	}
	for _, f := range []string{"de", "destino"} {
		_ = skillsInstalarCmd.Flags().Set(f, "")
	}
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return buf.String(), err
}

func TestSkillsInstalarNoOpComHub(t *testing.T) {
	hub := t.TempDir()
	forgeEm(t, hub)
	sub := filepath.Join(hub, "api")
	forgeEm(t, sub)
	origem := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(origem, "SKILL.md"), []byte("# x"), 0o644))

	out, err := rodaSkills(t, "instalar", "skill", "minha-skill", "--dir", hub, "--de", origem)
	require.NoError(t, err)
	assert.Contains(t, out, "1 gravado")
	assert.FileExists(t, filepath.Join(hub, ".agents", "skills", "minha-skill", "SKILL.md"))

	out, err = rodaSkills(t, "instalar", "skill", "minha-skill", "--dir", sub, "--de", origem)
	require.NoError(t, err)
	assert.Contains(t, out, "já disponível em "+hub)
	assert.NoDirExists(t, filepath.Join(sub, ".agents"))

	out, err = rodaSkills(t, "destinos", "--dir", sub, "--tipo", "skill", "--slug", "minha-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "já disponível em "+hub)

	out, err = rodaSkills(t, "destinos", "--dir", hub)
	require.NoError(t, err)
	assert.Equal(t, hub+"\n"+sub+"\n", out)

	_, err = rodaSkills(t, "instalar", "skill", "minha-skill", "--dir", hub, "--de", origem, "--destino", t.TempDir())
	require.Error(t, err)
	_, err = rodaSkills(t, "instalar", "nada", "x-skill", "--dir", hub, "--de", origem)
	require.Error(t, err)
}

func TestSkillsInstalarModRecusaManifestoIncompleto(t *testing.T) {
	proj := t.TempDir()
	forgeEm(t, proj)
	origem := t.TempDir()
	for rel, conteudo := range map[string]string{
		".claude-plugin/plugin.json": `{"name":"meu-mod","version":"0.1.0","description":"Mod de teste"}`,
		"hooks/hooks.json":           `{"modules":["./register.tsx"]}`,
		"hooks/register.tsx":         "export const register = () => {}",
	} {
		alvo := filepath.Join(origem, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(alvo), 0o755))
		require.NoError(t, os.WriteFile(alvo, []byte(conteudo), 0o644))
	}

	_, err := rodaSkills(t, "instalar", "plugin", "meu-mod", "--dir", proj, "--de", origem)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forge.json ausente")
	assert.NoDirExists(t, filepath.Join(proj, ".agents"), "nada é gravado quando o manifesto é recusado")

	// com o contrato completo a mesma origem instala
	forge := `{"schema":"forge-sdd/mod/1.0","minClaudeCode":"2.1.331","minForge":"1.0.0","permissions":{"network":false,"write":[],"hooks":["session.start"]}}`
	require.NoError(t, os.WriteFile(filepath.Join(origem, ".claude-plugin", "forge.json"), []byte(forge), 0o644))
	out, err := rodaSkills(t, "instalar", "plugin", "meu-mod", "--dir", proj, "--de", origem)
	require.NoError(t, err)
	assert.Contains(t, out, "4 gravado")
	assert.FileExists(t, filepath.Join(proj, ".agents", "plugins", "meu-mod", ".claude-plugin", "forge.json"))
}

func TestSkillsInstalarModDeReferenciaDoRepositorio(t *testing.T) {
	proj := t.TempDir()
	forgeEm(t, proj)
	origem := filepath.Join("..", "..", ".agents", "plugins", "forge-painel")
	out, err := rodaSkills(t, "instalar", "plugin", "forge-painel", "--dir", proj, "--de", origem)
	require.NoError(t, err)
	assert.Contains(t, out, "gravado")
	assert.FileExists(t, filepath.Join(proj, ".agents", "plugins", "forge-painel", "hooks", "register.tsx"))
}
