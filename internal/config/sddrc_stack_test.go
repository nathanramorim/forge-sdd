package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func escreveSddrc(t *testing.T, conteudo string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(conteudo), 0o644))
	return dir
}

func TestSddrcGravaEPreservaStackEBanco(t *testing.T) {
	dir := escreveSddrc(t, `{"project":"p","version":"5.1.0","agents":["claude"],"lang":"pt-BR","stack":"node","db":"postgres"}`)
	rc, err := ReadSddrc(dir)
	require.NoError(t, err)
	cfg := rc.ToConfig()
	assert.Equal(t, "node", cfg.Stack)
	assert.Equal(t, "postgres", cfg.DB)
}

func TestSddrcDe50SemStackNaoAssumeDefault(t *testing.T) {
	dir := escreveSddrc(t, `{"project":"p","version":"5.0.0","agents":["claude"],"lang":"pt-BR"}`)
	rc, err := ReadSddrc(dir)
	require.NoError(t, err)
	cfg := rc.ToConfig()
	assert.Empty(t, cfg.Stack)
	assert.Empty(t, cfg.DB)
}
