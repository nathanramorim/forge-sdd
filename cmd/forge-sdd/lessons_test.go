package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendLesson_CreatesFileWithHeader(t *testing.T) {
	dir := t.TempDir()
	path, err := AppendLesson(dir, "duplicação de lógica entre prompts", "extrair bloco único referenciado", "fix-50")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "Lições Aprendidas")
	assert.Contains(t, content, "duplicação de lógica entre prompts → extrair bloco único referenciado (fix-50)")
}

func TestAppendLesson_NewestFirst(t *testing.T) {
	dir := t.TempDir()
	_, err := AppendLesson(dir, "padrão A", "fix A", "")
	require.NoError(t, err)
	path, err := AppendLesson(dir, "padrão B", "fix B", "")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	idxB := strings.Index(content, "padrão B")
	idxA := strings.Index(content, "padrão A")
	require.NotEqual(t, -1, idxA)
	require.NotEqual(t, -1, idxB)
	assert.Less(t, idxB, idxA, "entrada mais recente deve aparecer primeiro")
}

func TestAppendLesson_RespectsBudget(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 200; i++ {
		_, err := AppendLesson(dir, "padrão repetido para estourar o orçamento de tamanho do arquivo", "correção aplicada nesta sessão de teste", "ref-teste")
		require.NoError(t, err)
	}

	path := filepath.Join(dir, "sdd", "memory", "lessons.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(data), lessonsBudgetBytes, "arquivo deve respeitar o orçamento configurado")
}

func TestAppendRuleLesson_CreatesRuleFileAndValidatesConstitution(t *testing.T) {
	dir := t.TempDir()
	memDir := filepath.Join(dir, "sdd", "memory")
	require.NoError(t, os.MkdirAll(memDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "constitution.md"), []byte("# Constituição\n- **COD-02:** go vet\n- **DUP-10:** fonte única\n"), 0o644))

	path, err := AppendRuleLesson(dir, "COD-02", "faltou go vet", "rodar go vet ./...", "feat-106-12")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "sdd", "memory", "licoes", "COD-02.md"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "COD-02")
	assert.Contains(t, string(data), "faltou go vet → rodar go vet ./... (feat-106-12)")

	_, err = AppendRuleLesson(dir, "ZZZ-99", "padrão", "fix", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ZZZ-99")

	_, err = AppendRuleLesson(dir, "../COD-02", "padrão", "fix", "")
	require.Error(t, err)
}
