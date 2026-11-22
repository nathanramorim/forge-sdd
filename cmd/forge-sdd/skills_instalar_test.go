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
