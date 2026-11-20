package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureAll captura todo o stdout (sem o limite do captureStdout) lendo em paralelo.
func captureAll(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	done := make(chan string)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	fn()
	require.NoError(t, w.Close())
	os.Stdout = old
	return <-done
}

func doctorProjeto(t *testing.T, sddVersion string, features int) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", "001-x", "features"), 0o755))
	rc := fmt.Sprintf(`{"project":"p","version":%q,"agents":["claude"],"stack":"go","db":"none","lang":"pt-BR","naming_convention":"sequencial"}`, sddVersion)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(rc), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sdd-version"), []byte(sddVersion), 0o644))
	for i := 0; i < features; i++ {
		f := filepath.Join(dir, "sdd", "discovery", "001-x", "features", fmt.Sprintf("feat-001-%02d.md", i))
		require.NoError(t, os.WriteFile(f, []byte("# f\n"), 0o644))
	}
	return dir
}

func doctorSaida(t *testing.T, binario, dir string, args ...string) string {
	t.Helper()
	config.SetBinaryVersion(binario)
	t.Cleanup(func() { config.SetBinaryVersion(version) })
	_ = doctorCmd.Flags().Set("verbose", "false")
	return captureAll(t, func() {
		_, err := runRoot(t, append([]string{"doctor", dir}, args...)...)
		require.NoError(t, err)
	})
}

func TestDoctorVersaoComparaComBinario(t *testing.T) {
	ok := doctorSaida(t, "5.0.0", doctorProjeto(t, "5.0.0", 0))
	assert.Contains(t, ok, "✓ Projeto em sincronia com o binário: v5.0.0")

	antigo := doctorSaida(t, "5.1.0", doctorProjeto(t, "5.0.0", 0))
	assert.Contains(t, antigo, "forge update")
	assert.NotContains(t, antigo, "em sincronia")
}

func TestDoctorBinarioMaisAntigoFalha(t *testing.T) {
	out := doctorSaida(t, "4.9.1", doctorProjeto(t, "5.0.0", 0))
	assert.Contains(t, out, "✗ Binário na v4.9.1 é mais antigo que o projeto")
	assert.NotContains(t, out, "em sincronia")
}

func TestDoctorVersaoSaneada(t *testing.T) {
	dir := doctorProjeto(t, "5.0.0", 0)
	sujo := "5.0.0\x1b[31mVERMELHO\r\x07"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sdd-version"), []byte(sujo), 0o644))
	out := doctorSaida(t, "5.0.0", dir)
	assert.NotContains(t, out, "\x1b")
	assert.NotContains(t, out, "\r")
	assert.NotContains(t, out, "\x07")
}

func TestDoctorListaSoComVerbose(t *testing.T) {
	dir := doctorProjeto(t, "5.0.0", 30)
	before := snapshotArquivos(t, dir)

	sem := doctorSaida(t, "5.0.0", dir)
	assert.Contains(t, sem, "30 cadastrada(s)")
	assert.NotContains(t, sem, "Caminhos:")
	assert.NotContains(t, sem, "feat-001-05.md")
	assert.LessOrEqual(t, len(strings.Split(sem, "\n")), 40, "saída padrão deve ser curta")

	com := doctorSaida(t, "5.0.0", dir, "--verbose")
	assert.Contains(t, com, "Caminhos:")
	assert.Contains(t, com, "feat-001-05.md")

	assert.Equal(t, before, snapshotArquivos(t, dir), "doctor não pode escrever no disco")
}

// snapshotArquivos junta caminho e conteúdo de todos os arquivos para provar ausência de escrita.
func snapshotArquivos(t *testing.T, dir string) string {
	t.Helper()
	var sb strings.Builder
	require.NoError(t, filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, _ := os.ReadFile(p)
		sb.WriteString(p + "|" + string(b) + "\n")
		return nil
	}))
	return sb.String()
}
