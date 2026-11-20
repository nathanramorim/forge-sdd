package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/deprecation"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projetoSdd(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	return dir
}

func TestComandoDepreciadoRegistraEAvisa(t *testing.T) {
	t.Setenv("CLAUDECODE", "1")
	dir := projetoSdd(t)
	_, stderr, err := runSep(t, "queue", "stations", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, stderr, "AVISO")
	b, err := os.ReadFile(filepath.Join(dir, deprecation.RelPath))
	require.NoError(t, err)
	linhas := strings.Split(strings.TrimSpace(string(b)), "\n")
	require.Len(t, linhas, 2)
	assert.Contains(t, linhas[1], ",queue stations,stations --dir,claude-code,")
	assert.Contains(t, linhas[1], "ADR-0021")
	assert.NotContains(t, string(b), dir, "sem caminho absoluto")
}

func TestFalhaDeEscritaNaoQuebraComando(t *testing.T) {
	dir := projetoSdd(t)
	// o CSV é um diretório: a gravação falha
	require.NoError(t, os.MkdirAll(filepath.Join(dir, deprecation.RelPath), 0o755))
	_, stderr, err := runSep(t, "queue", "stations", "--dir", dir)
	require.NoError(t, err, "falha de registro não quebra o comando")
	assert.Contains(t, stderr, "depreciado")
	assert.Contains(t, stderr, "não gravado")
}

func TestDryRunNaoGravaRegistro(t *testing.T) {
	dir := projetoSdd(t)
	// nenhum depreciado atual tem --dry-run; usa comando de teste com a flag e entrada no catálogo.
	deprecation.Catalogo = append(deprecation.Catalogo, deprecation.Entrada{Comando: "fake", Substituto: "x", Desde: "ADR-0021"})
	t.Cleanup(func() { deprecation.Catalogo = deprecation.Catalogo[:len(deprecation.Catalogo)-1] })
	exec := func(args ...string) {
		root := &cobra.Command{Use: "forge"}
		fake := &cobra.Command{Use: "fake", PreRun: avisoDepreciado(), Run: func(*cobra.Command, []string) {}}
		fake.Flags().Bool("dry-run", false, "")
		fake.Flags().String("dir", "", "")
		root.AddCommand(fake)
		root.SetArgs(args)
		require.NoError(t, root.Execute())
	}
	exec("fake", "--dry-run", "--dir", dir)
	_, err := os.Stat(filepath.Join(dir, deprecation.RelPath))
	assert.True(t, os.IsNotExist(err), "--dry-run não grava")

	// sem --dry-run, o mesmo comando grava
	exec("fake", "--dir", dir)
	_, err = os.Stat(filepath.Join(dir, deprecation.RelPath))
	assert.NoError(t, err)
}

func TestComandoVigenteNaoRegistra(t *testing.T) {
	dir := projetoSdd(t)
	for _, args := range [][]string{{"queue", "--dir", dir}, {"queue", "enqueue", "--dir", dir}} {
		_, _, _ = runSep(t, args...)
	}
	_, err := os.Stat(filepath.Join(dir, deprecation.RelPath))
	assert.True(t, os.IsNotExist(err))
}

func TestGitattributesTemMergeUnionDoCSV(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".gitattributes"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "sdd/memory/logs/*.csv merge=union")
}
