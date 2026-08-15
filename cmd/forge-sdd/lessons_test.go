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
