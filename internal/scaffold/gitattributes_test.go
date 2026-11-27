package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fix-118-04: o init grava o bloco eol=lf dos arquivos do forge; rodar de novo não duplica.
func TestInitGravaGitattributesComLF(t *testing.T) {
	dir := t.TempDir()
	cfg := cfgInstrucao()
	for i := 0; i < 2; i++ {
		_, err := Run(cfg, dir)
		require.NoError(t, err)
	}
	got := lerRel(t, dir, ".gitattributes")
	assert.Equal(t, 1, strings.Count(got, marcadorGitattributes))
	for _, linha := range []string{"sdd/** text eol=lf", ".agents/** text eol=lf", ".claude/** text eol=lf"} {
		assert.Contains(t, got, linha)
	}
	assert.NotContains(t, got, "* text=auto", "não mexe no código do usuário")
}

func TestGitattributesDoUsuarioPreservado(t *testing.T) {
	dir := t.TempDir()
	meu := "*.png binary\nsdd/memory/logs/*.csv merge=union"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(meu), 0o644))
	garantirGitattributes(dir)
	got := lerRel(t, dir, ".gitattributes")
	assert.True(t, strings.HasPrefix(got, meu+"\n\n"), got)
	assert.Contains(t, got, blocoGitattributes)
}
