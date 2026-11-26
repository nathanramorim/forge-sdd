package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNuvem_AtivarStatusEDesativar(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, runSubcommand("pipeline", "init", "--dir", dir))
	require.NoError(t, runSubcommand("nuvem", "ativar", "--dir", dir, "--repos", "api, web", "--instalar", "git"))
	raw, err := os.ReadFile(filepath.Join(dir, ".forge", "config", "forge.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"ativo": true`)
	assert.Contains(t, string(raw), `"api"`)
	_, err = os.Stat(filepath.Join(dir, ".claude", "hooks", "forge-nuvem.sh"))
	require.NoError(t, err)

	require.Error(t, runSubcommand("nuvem", "ativar", "--dir", dir, "--ref", "main;rm"), "valor com shell é recusado")

	require.NoError(t, runSubcommand("nuvem", "desativar", "--dir", dir))
	_, err = os.Stat(filepath.Join(dir, ".claude", "hooks", "forge-nuvem.sh"))
	assert.True(t, os.IsNotExist(err))
}

func TestNuvemIniciar_ForaDaNuvemNaoFazNada(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, nuvemIniciar(&out, t.TempDir(), false, func(string) string { return "" }))
	assert.Contains(t, out.String(), "fora da nuvem")
}
