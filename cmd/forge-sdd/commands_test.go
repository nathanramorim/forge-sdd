package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
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
		"version": "1.9.1-beta",
		"agents": ["gemini"],
		"stack": "go",
		"db": "postgres",
		"lang": "pt-BR"
	}`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sddDir, ".sdd-version"), []byte("1.9.1-beta"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sddDir, "memory", "progress.md"), []byte("# Progress"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(sddDir, "memory", "constitution.md"), []byte("# Constitution"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "GEMINI.md"), []byte("# Gemini"), 0644)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tempDir, ".gemini"), 0755)
	require.NoError(t, err)

	// Cria estrutura de features aninhadas
	featuresDir := filepath.Join(sddDir, "features")
	err = os.MkdirAll(filepath.Join(featuresDir, "feat-02-checkout"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(featuresDir, "feat-01-auth.md"), []byte("# Auth"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(featuresDir, "feat-02-checkout", "feat-02-01-api.md"), []byte("# API Checkout"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(featuresDir, "feat-02-checkout", "task-01-database.md"), []byte("# DB Checkout"), 0644)
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
	assert.Contains(t, output, "Features Registradas: 3 cadastrada(s)")
	assert.Contains(t, output, "sdd/features/feat-01-auth.md")
	assert.Contains(t, output, "sdd/features/feat-02-checkout/feat-02-01-api.md")
	assert.Contains(t, output, "sdd/features/feat-02-checkout/task-01-database.md")

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

func TestDoctorCommand_NamingConventionConsistent(t *testing.T) {
	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "memory"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "features"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{"project":"test-proj","version":"1.9.1-beta","agents":[],"stack":"go","db":"none","lang":"pt-BR"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "memory", "progress.md"), []byte("# Progress"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "memory", "constitution.md"), []byte("# Constitution"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "features", "feat-00-foundation.md"), []byte("# Foundation"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "features", "feat-01-auth.md"), []byte("# Auth"), 0644))

	output := runDoctorCapturingStdout(t, tempDir)

	assert.Contains(t, output, "Convenção de Nomenclatura:")
	assert.Contains(t, output, "Convenção de nomenclatura consistente.")
	assert.NotContains(t, output, "Deriva de convenção detectada")
}

func TestDoctorCommand_NamingConventionDriftDetected(t *testing.T) {
	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "memory"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "features"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "discovery"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{"project":"test-proj","version":"1.9.1-beta","agents":[],"stack":"go","db":"none","lang":"pt-BR"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "memory", "progress.md"), []byte("# Progress"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "memory", "constitution.md"), []byte("# Constitution"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "features", "feat-00-foundation.md"), []byte("# Foundation"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "features", "feat-43a2-fix.md"), []byte("# Fix"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "discovery", "discovery-9b2f-tema.md"), []byte("# Discovery"), 0644))

	output := runDoctorCapturingStdout(t, tempDir)

	assert.Contains(t, output, "Deriva de convenção detectada")
	assert.Contains(t, output, "sdd/features/feat-00-foundation.md")
	assert.Contains(t, output, "sdd/features/feat-43a2-fix.md")
	assert.Contains(t, output, "sdd/discovery/discovery-9b2f-tema.md")
}

func runDoctorCapturingStdout(t *testing.T, targetDir string) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"doctor", targetDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestInitCommand(t *testing.T) {
	// Cenário 1: Init com argumento de pasta (Caso A)
	tempDirA := t.TempDir()
	
	_ = initCmd.Flags().Set("yes", "true")
	_ = initCmd.Flags().Set("name", "")
	_ = initCmd.Flags().Set("stack", "go")
	_ = initCmd.Flags().Set("db", "postgres")
	rootCmd.SetArgs([]string{"init", tempDirA})
	
	err := rootCmd.Execute()
	require.NoError(t, err)

	// Valida que os arquivos foram gerados na raiz da pasta e não em uma subpasta
	assert.FileExists(t, filepath.Join(tempDirA, "sdd", ".sddrc"))
	assert.FileExists(t, filepath.Join(tempDirA, "sdd", ".sdd-version"))

	// Cenário 2: Init sem argumentos (Caso B)
	tempDirB := t.TempDir()
	
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(tempDirB)
	require.NoError(t, err)
	
	_ = initCmd.Flags().Set("yes", "true")
	_ = initCmd.Flags().Set("name", "sub-proj")
	_ = initCmd.Flags().Set("stack", "node")
	_ = initCmd.Flags().Set("db", "postgres")
	rootCmd.SetArgs([]string{"init"})
	
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Valida que a subpasta "sub-proj" foi criada
	subProjDir := filepath.Join(tempDirB, "sub-proj")
	assert.DirExists(t, subProjDir)
	assert.FileExists(t, filepath.Join(subProjDir, "sdd", ".sddrc"))
}

func TestInitCommand_TutorialFlagPrintsHint(t *testing.T) {
	tempDir := t.TempDir()

	_ = initCmd.Flags().Set("yes", "true")
	_ = initCmd.Flags().Set("name", "demo")
	_ = initCmd.Flags().Set("stack", "go")
	_ = initCmd.Flags().Set("db", "none")
	_ = initCmd.Flags().Set("tutorial", "true")
	defer func() { _ = initCmd.Flags().Set("tutorial", "false") }()

	output := runInitCapturingStdout(t, tempDir)

	assert.Contains(t, output, "Modo tutorial ativado")
	assert.Contains(t, output, "/tutorial")
}

func TestInitCommand_TutorialFlagSkippedOnDryRun(t *testing.T) {
	tempDir := t.TempDir()

	_ = initCmd.Flags().Set("yes", "true")
	_ = initCmd.Flags().Set("name", "demo")
	_ = initCmd.Flags().Set("stack", "go")
	_ = initCmd.Flags().Set("db", "none")
	_ = initCmd.Flags().Set("tutorial", "true")
	_ = initCmd.Flags().Set("dry-run", "true")
	defer func() {
		_ = initCmd.Flags().Set("tutorial", "false")
		_ = initCmd.Flags().Set("dry-run", "false")
	}()

	output := runInitCapturingStdout(t, tempDir)

	assert.NotContains(t, output, "Modo tutorial ativado")
}

func runInitCapturingStdout(t *testing.T, targetDir string) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"init", targetDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDoctorCommand_AutoHealNamingConvention(t *testing.T) {
	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "memory"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sddDir, "features"), 0755))

	// sddrc sem naming_convention
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{"project":"test-proj","version":"1.9.1-beta","agents":[],"stack":"go","db":"none","lang":"pt-BR"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "memory", "progress.md"), []byte("# Progress"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, "features", "feat-43a2-telemetry.md"), []byte("# Telemetry"), 0644))

	output := runDoctorCapturingStdout(t, tempDir)

	assert.Contains(t, output, "Convenção de nomenclatura ausente corrigida automaticamente para: hash")

	// verifica se o .sddrc agora contem a convention
	data, err := os.ReadFile(filepath.Join(sddDir, ".sddrc"))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"naming_convention": "hash"`)
}

func TestUpdateCommand_UpgradeYesResolvesFromNpmRegistry(t *testing.T) {
	// Mock do NPM Registry com uma beta mais recente do que a versão local compilada no binário de teste.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"dist-tags": map[string]string{
				"latest": "9.0.0",
				"beta":   "9.9.9-beta",
			},
		})
	}))
	defer server.Close()

	oldURL := config.NpmRegistryURL
	config.NpmRegistryURL = server.URL
	defer func() { config.NpmRegistryURL = oldURL }()

	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{
		"project": "test-proj",
		"version": "1.0.0",
		"agents": ["gemini"],
		"lang": "pt-BR"
	}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sdd-version"), []byte("1.0.0"), 0644))

	_ = updateCmd.Flags().Set("yes", "true")
	_ = updateCmd.Flags().Set("upgrade", "true")
	defer func() {
		_ = updateCmd.Flags().Set("yes", "false")
		_ = updateCmd.Flags().Set("upgrade", "false")
	}()

	rootCmd.SetArgs([]string{"update", tempDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	// A versão gravada deve vir do registry mockado (a beta, por ser mais recente que o latest mockado),
	// nunca da constante `version` compilada localmente no binário de teste (1.7.1-beta.5).
	gotVersion, err := os.ReadFile(filepath.Join(sddDir, ".sdd-version"))
	require.NoError(t, err)
	assert.Equal(t, "9.9.9-beta", strings.TrimSpace(string(gotVersion)))
}

func TestUpdateCommand_UpgradeYesNpmFailureReturnsExplicitError(t *testing.T) {
	oldURL := config.NpmRegistryURL
	config.NpmRegistryURL = "http://127.0.0.1:1" // porta inválida: falha de conexão imediata
	defer func() { config.NpmRegistryURL = oldURL }()

	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{
		"project": "test-proj",
		"version": "1.0.0",
		"agents": ["gemini"],
		"lang": "pt-BR"
	}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sdd-version"), []byte("1.0.0"), 0644))

	_ = updateCmd.Flags().Set("yes", "true")
	_ = updateCmd.Flags().Set("upgrade", "true")
	defer func() {
		_ = updateCmd.Flags().Set("yes", "false")
		_ = updateCmd.Flags().Set("upgrade", "false")
	}()

	rootCmd.SetArgs([]string{"update", tempDir})
	err := rootCmd.Execute()

	require.Error(t, err, "esperava erro explícito quando o NPM Registry está inacessível durante --upgrade --yes")
	assert.Contains(t, err.Error(), "NPM Registry")
}

func TestUpdateCommand_NamingConventionFlagPersisted(t *testing.T) {
	tempDir := t.TempDir()
	sddDir := filepath.Join(tempDir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{
		"project": "test-proj",
		"version": "1.9.1",
		"agents": ["gemini"],
		"lang": "pt-BR",
		"naming_convention": "sequencial"
	}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sdd-version"), []byte("1.9.1"), 0644))

	_ = updateCmd.Flags().Set("yes", "true")
	_ = updateCmd.Flags().Set("naming-convention", "workitem")
	defer func() {
		_ = updateCmd.Flags().Set("yes", "false")
		_ = updateCmd.Flags().Set("naming-convention", "")
	}()

	rootCmd.SetArgs([]string{"update", tempDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(sddDir, ".sddrc"))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"naming_convention": "workitem"`,
		"a flag --naming-convention deve ser persistida em sdd/.sddrc mesmo sem --agent/--upgrade/--version")
}
