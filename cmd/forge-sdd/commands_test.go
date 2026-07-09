package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctorCommand_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Redireciona o stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Executa o comando via rootCmd
	rootCmd.SetArgs([]string{"doctor", tempDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Iniciando diagnóstico da metodologia")
	assert.Contains(t, output, "Estrutura Forge-SDD não encontrada (sdd/.sddrc ausente)")
}

func TestDoctorAndDestroyCommands(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Inicializa uma estrutura básica fake no diretório temporário para simular a presença do SDD
	sddDir := filepath.Join(tempDir, "sdd")
	err := os.MkdirAll(filepath.Join(sddDir, "memory"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(sddDir, "features"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{
		"project": "test-proj",
		"version": "1.6.1-beta.1",
		"agents": ["gemini"],
		"stack": "go",
		"db": "postgres",
		"lang": "pt-BR"
	}`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sddDir, ".sdd-version"), []byte("1.6.1-beta.1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sddDir, "memory", "progress.md"), []byte("# Progress"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(sddDir, "memory", "constitution.md"), []byte("# Constitution"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "GEMINI.md"), []byte("# Gemini"), 0644)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tempDir, ".gemini"), 0755)
	require.NoError(t, err)

	// 2. Executa o doctor e valida a saída positiva
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"doctor", tempDir})
	err = rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Diretório base da metodologia (sdd/) e arquivo de configuração encontrados.")
	assert.Contains(t, output, "Projeto: test-proj")
	assert.Contains(t, output, "Gemini (Google): Configurado e integrado (GEMINI.md + .gemini/)")
	assert.Contains(t, output, "Constituição (sdd/memory/constitution.md) ativa.")

	// 3. Executa o destroy com --dry-run
	r, w, _ = os.Pipe()
	os.Stdout = w

	_ = destroyCmd.Flags().Set("dry-run", "true")
	_ = destroyCmd.Flags().Set("yes", "false")
	rootCmd.SetArgs([]string{"destroy", tempDir, "--dry-run"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	_, _ = io.Copy(&buf, r)
	dryOutput := buf.String()

	assert.Contains(t, dryOutput, "[DRY-RUN] Seria removido:")
	assert.Contains(t, dryOutput, "[DRY-RUN] Fim da simulação.")

	// Garante que os arquivos NÃO foram removidos pelo dry-run
	assert.DirExists(t, sddDir)
	assert.FileExists(t, filepath.Join(tempDir, "GEMINI.md"))

	// 4. Executa o destroy com --yes (remoção de fato)
	r, w, _ = os.Pipe()
	os.Stdout = w

	_ = destroyCmd.Flags().Set("dry-run", "false")
	_ = destroyCmd.Flags().Set("yes", "true")
	rootCmd.SetArgs([]string{"destroy", tempDir, "--yes"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	_, _ = io.Copy(&buf, r)
	destroyOutput := buf.String()

	assert.Contains(t, destroyOutput, "Removido com sucesso:")
	assert.Contains(t, destroyOutput, "Estrutura Forge-SDD removida com sucesso")

	// Garante que os arquivos FORAM removidos
	assert.NoDirExists(t, sddDir)
	assert.NoFileExists(t, filepath.Join(tempDir, "GEMINI.md"))
	assert.NoDirExists(t, filepath.Join(tempDir, ".gemini"))
}
