package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateDryRun_ImprimeMovimentosENaoGravaNada(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"demo","version":"3.2.2-beta","agents":["claude"],"lang":"pt-BR","naming_convention":"sequencial"}`)
	testutil.WriteFile(t, dir, "sdd/.sdd-version", "3.2.2-beta")
	testutil.WriteFile(t, dir, "sdd/discovery/discovery-01-alfa.md", "d1")
	testutil.WriteFile(t, dir, "sdd/features/index.md", "indice")
	testutil.WriteFile(t, dir, "sdd/features/feat-01-alfa.md", "f1")

	out, err := runRoot(t, "update", dir, "--dry-run")
	require.NoError(t, err)

	assert.Contains(t, out, "[DRY-RUN]")
	assert.Contains(t, out, "mover   sdd/discovery/discovery-01-alfa.md -> sdd/archive/discovery/discovery-01-alfa.md")
	assert.Contains(t, out, "mover   sdd/features/index.md -> sdd/archive/features/index.md")

	// Nada foi movido nem criado.
	assert.FileExists(t, filepath.Join(dir, "sdd", "features", "index.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "discovery", "discovery-01-alfa.md"))
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "archive"))
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "backups"))
	_, err = os.Stat(filepath.Join(dir, ".agents"))
	assert.True(t, os.IsNotExist(err), "dry-run não pode criar .agents")
}

func TestUpdateDryRun_SemLegadoInformaQueNadaHaAFazer(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"demo","version":"4.0.0-beta","agents":["claude"],"lang":"pt-BR"}`)
	out, err := runRoot(t, "update", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "Nenhuma migração de layout")
}
