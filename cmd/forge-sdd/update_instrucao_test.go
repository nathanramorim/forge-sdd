package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateAvisaInstrucaoPreservada(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"demo","version":"5.0.0","agents":["claude"],"lang":"pt-BR","stack":"node","db":"none","naming_convention":"sequencial"}`)
	testutil.WriteFile(t, dir, "sdd/.sdd-version", "5.0.0")
	testutil.WriteFile(t, dir, "CLAUDE.md", "# CLAUDE.md — demo\n\nMINHA REGRA\n")

	// --dry-run lista o arquivo e não grava nada
	out, err := runRoot(t, "update", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "mantido CLAUDE.md")
	assert.NoFileExists(t, filepath.Join(dir, "CLAUDE.novo.md"))

	old := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	_, err = runRoot(t, "update", dir, "--yes", "--version", "5.0.1")
	os.Stderr = old
	_ = w.Close()
	require.NoError(t, err)
	errOut, _ := io.ReadAll(r)

	assert.Contains(t, string(errOut), "CLAUDE.md é seu")
	assert.Contains(t, string(errOut), "CLAUDE.novo.md")
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	assert.Equal(t, "# CLAUDE.md — demo\n\nMINHA REGRA\n", string(b))
	assert.FileExists(t, filepath.Join(dir, "CLAUDE.novo.md"))
}
