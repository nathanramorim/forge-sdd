package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUpgradeMantemStackDoSddrc(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/.sddrc", `{"project":"demo","version":"5.0.0","agents":["claude","gemini"],"lang":"pt-BR","naming_convention":"sequencial"}`)
	testutil.WriteFile(t, dir, "sdd/.sdd-version", "5.0.0")
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "| Runtime | node | |\n| DB | postgres | |\n")
	testutil.WriteFile(t, dir, "CLAUDE.md", "demo: stack node. Fluxo\n")

	_, err := runRoot(t, "update", dir, "--yes", "--version", "5.0.1")
	require.NoError(t, err)

	rc, err := os.ReadFile(filepath.Join(dir, "sdd/.sddrc"))
	require.NoError(t, err)
	assert.Contains(t, string(rc), `"stack": "node"`)
	assert.Contains(t, string(rc), `"db": "postgres"`)
	cl, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	assert.True(t, strings.Contains(string(cl), "stack node."), string(cl))
	gs, _ := os.ReadFile(filepath.Join(dir, ".gemini/system_instructions.md"))
	assert.Contains(t, string(gs), "- **Banco:** postgres")
}
