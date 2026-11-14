package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func snapshotDir(t *testing.T, dir string) map[string]int64 {
	t.Helper()
	snap := map[string]int64{}
	err := filepath.Walk(dir, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		snap[rel] = info.Size()
		return nil
	})
	require.NoError(t, err)
	return snap
}

func TestBriefComando(t *testing.T) {
	dir := t.TempDir()

	const constText = `# Constituição — forge-cmd-demo

## Missão
Prover comando brief determinístico.

## Stack
| Item | Escolha |
|---|---|
| Go | 1.23 |

## Regras
1. Regra única.

## IDs da Constituição
- **ARQ-01:** Regra de arquitetura.
`
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", constText)
	testutil.WriteFile(t, dir, "sdd/spec/modules.md", "# Módulos\n\n## CLI\nInterface.\n")
	testutil.WriteFile(t, dir, "sdd/spec/flows.md", "# Fluxos\n\n## Brief\nMontagem do contexto.\n")
	testutil.WriteFile(t, dir, "sdd/spec/decisions.md", "# Decisões\n\n## Decisões Resolvidas\n| ID | Decisão | Status |\n|---|---|---|\n| D01 | Arquitetura | resolvido |\n")

	beforeSnap := snapshotDir(t, dir)

	// 1. brief padrão imprime resumo
	out, err := runRoot(t, "brief", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "# Contexto do projeto — forge-cmd-demo")
	assert.Contains(t, out, "Estado: pronta")
	assert.Contains(t, out, "Missão: Prover comando brief determinístico.")

	// 2. brief --check em projeto pronta sai com código 0
	outCheck, errCheck := runRoot(t, "brief", "--dir", dir, "--check")
	require.NoError(t, errCheck)
	assert.Equal(t, "pronta\n", outCheck)

	// 3. brief --check em projeto rascunho sai com código 3
	dirRascunho := t.TempDir()
	testutil.WriteFile(t, dirRascunho, "sdd/memory/constitution.md", "# Constituição\n\n## Missão\n[Uma frase]\n")
	testutil.WriteFile(t, dirRascunho, "sdd/spec/modules.md", "# Módulos\n")

	outRascunho, errRascunho := runRoot(t, "brief", "--dir", dirRascunho, "--check")
	require.Error(t, errRascunho)
	assert.Contains(t, outRascunho, "rascunho")
	var ec exitCoder
	require.True(t, errors.As(errRascunho, &ec))
	assert.Equal(t, 3, ec.ExitCode())

	// 4. Árvore de arquivos idêntica antes e depois (nunca escreve em disco)
	afterSnap := snapshotDir(t, dir)
	assert.Equal(t, beforeSnap, afterSnap, "árvore de arquivos não pode ser alterada pelo brief")
}
