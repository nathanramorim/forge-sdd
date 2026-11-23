package vcs_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeModosAposAcrescimoDoUpdate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sdd", "memory", "constitution.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte("## Ferramentas e Integrações\n| Campo | Valor |\n|---|---|\n| VCS / Work Item System | nenhum |\n| WIP humano | 7 |\n"), 0o644))

	_, err := scaffold.AcrescentaCamposDeModo(dir, &bytes.Buffer{})
	require.NoError(t, err)
	m, err := vcs.LeModos(dir)
	require.NoError(t, err)
	assert.Equal(t, "nenhum", m.VCS)
	assert.Equal(t, 7, m.WIPHumano)
	assert.Equal(t, vcs.AtestadoLocal, m.Atestado)
	assert.Equal(t, vcs.RevisaoComHumano, m.RevisaoHumana)
	assert.Equal(t, vcs.ModulosNenhum, m.ModulosPermitidos)
	assert.Equal(t, vcs.InicioDesligado, m.InicioAutomatico)
}
