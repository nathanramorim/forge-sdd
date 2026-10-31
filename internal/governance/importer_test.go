package governance

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyBlock_InsertsAndUpdatesIdempotently(t *testing.T) {
	original := "# Constituição — demo\n\n## Regras\n1. Sem commit em main\n"
	blockV1 := "<!-- acme-governanca:inicio versao=0.1.0 -->\n## Governança Acme\n* **ARQ-01** Regra 1\n<!-- acme-governanca:fim -->"
	blockV2 := "<!-- acme-governanca:inicio versao=0.2.0 -->\n## Governança Acme\n* **ARQ-01** Regra 1 atualizada\n<!-- acme-governanca:fim -->"

	first := ApplyBlock(original, "<!-- acme-governanca:inicio", "<!-- acme-governanca:fim -->", blockV1)
	assert.Contains(t, first, "# Constituição — demo")
	assert.Contains(t, first, "versao=0.1.0")

	second := ApplyBlock(first, "<!-- acme-governanca:inicio", "<!-- acme-governanca:fim -->", blockV2)
	assert.Contains(t, second, "# Constituição — demo")
	assert.Contains(t, second, "versao=0.2.0")
	assert.NotContains(t, second, "versao=0.1.0")
}

func TestApplyFile_Modes(t *testing.T) {
	dir := t.TempDir()

	// 1. criar_se_ausente: cria na 1ª vez, pula na 2ª
	action, err := ApplyFile(dir, "docs/governanca/catalogo.md", ModeCriarSeAusente, []byte("catalogo inicial"))
	require.NoError(t, err)
	assert.Equal(t, "created", action)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs/governanca/catalogo.md"), []byte("catalogo editado pelo time"), 0644))
	action, err = ApplyFile(dir, "docs/governanca/catalogo.md", ModeCriarSeAusente, []byte("catalogo novo"))
	require.NoError(t, err)
	assert.Equal(t, "skipped", action)
	data, err := os.ReadFile(filepath.Join(dir, "docs/governanca/catalogo.md"))
	require.NoError(t, err)
	assert.Equal(t, "catalogo editado pelo time", string(data))

	// 2. substituir: sempre sobrescreve
	action, err = ApplyFile(dir, ".agents/rules/pb-seguranca.md", ModeSubstituir, []byte("v1"))
	require.NoError(t, err)
	assert.Equal(t, "created", action)
	action, err = ApplyFile(dir, ".agents/rules/pb-seguranca.md", ModeSubstituir, []byte("v2"))
	require.NoError(t, err)
	assert.Equal(t, "replaced", action)
	data, err = os.ReadFile(filepath.Join(dir, ".agents/rules/pb-seguranca.md"))
	require.NoError(t, err)
	assert.Equal(t, "v2", string(data))

	// 3. mesclar: preserva existente e anexa novo
	action, err = ApplyFile(dir, ".github/pull_request_template.md", ModeMesclar, []byte("## Tarefa"))
	require.NoError(t, err)
	assert.Equal(t, "created", action)
	action, err = ApplyFile(dir, ".github/pull_request_template.md", ModeMesclar, []byte("## Exceções"))
	require.NoError(t, err)
	assert.Equal(t, "merged", action)
	data, err = os.ReadFile(filepath.Join(dir, ".github/pull_request_template.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "## Tarefa")
	assert.Contains(t, string(data), "## Exceções")
}

func TestWriteVersionRecord(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 31, 14, 0, 0, 0, time.UTC)
	path, err := WriteVersionRecord(dir, "planning-brain-governanca", "0.2.1", "abc1234", now)
	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "pacote: planning-brain-governanca")
	assert.Contains(t, string(data), "versao: 0.2.1")
	assert.Contains(t, string(data), "sha: abc1234")
}
