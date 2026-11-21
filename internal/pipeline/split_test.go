package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// iniciativaDeTeste cria a pasta da iniciativa com status.md e plan.md.
func iniciativaDeTeste(t *testing.T, status, plano string) string {
	t.Helper()
	dir := t.TempDir()
	base := filepath.Join(dir, "sdd", "discovery", "106-pr-draft-revisao-e-ci")
	require.NoError(t, os.MkdirAll(filepath.Join(base, "features"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "status.md"), []byte(status), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "plan.md"), []byte(plano), 0o644))
	return dir
}

func specDeTeste(t *testing.T, dir, sub, nome, titulo string) {
	t.Helper()
	p := filepath.Join(dir, "sdd", "discovery", "106-pr-draft-revisao-e-ci", sub, nome)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	corpo := "# " + titulo + "\n\n## O que já existe\n| Procurei | Onde | Achei | Decisão |\n|---|---|---|---|\n| x | y | z | **Reusar** |\n"
	require.NoError(t, os.WriteFile(p, []byte(corpo), 0o644))
}

const planoDeTeste = `# Plano — Iniciativa 106

## Onda 1
| # | Task | Depende de |
|---|---|---|
| 1.1 | **feat-106-21** — queue review | — |
| 1.2 | **feat-106-22a** — PR da spec | feat-106-21 |

## Onda 2
- **feat-106-23** Encerrar
`

func TestBuilderFazSplitSemSegundaAprovacao(t *testing.T) {
	dir := iniciativaDeTeste(t, "**Gate 1:** aprovado em 21/11/2026\n", planoDeTeste)
	for _, id := range []string{"feat-106-21", "feat-106-22a", "feat-106-23"} {
		specDeTeste(t, dir, "features", id+"-x.md", id+" — título")
	}

	// Com o Gate 1 aprovado, o Builder enfileira direto, sem nova pergunta.
	enfileiradas, jaExistiam, err := SplitPeloBuilder(dir, "106-pr-draft-revisao-e-ci", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"feat-106-21", "feat-106-22a", "feat-106-23"}, enfileiradas,
		"a ordem é a do plano")
	assert.Empty(t, jaExistiam)

	// Rodar de novo não duplica: as três já estão na fila.
	novas, existentes, err := SplitPeloBuilder(dir, "106-pr-draft-revisao-e-ci", nil)
	require.NoError(t, err)
	assert.Empty(t, novas)
	assert.Len(t, existentes, 3)
}

func TestIdsDasTasksDoPlano(t *testing.T) {
	ids := IdsDasTasksDoPlano(planoDeTeste)
	assert.Equal(t, []string{"feat-106-21", "feat-106-22a", "feat-106-23"}, ids)

	// Ignora o que não é ID de task.
	assert.Empty(t, IdsDasTasksDoPlano("nada aqui, só prosa"))
	assert.Empty(t, IdsDasTasksDoPlano("feat-x-y e fix-abc"))

	// Aceita fix e sufixo de letra.
	assert.Equal(t, []string{"fix-106-01", "feat-106-24b"},
		IdsDasTasksDoPlano("`fix-106-01`, `feat-106-24b`"))
}

func TestSplitForaDoPlanoRecusado(t *testing.T) {
	dir := iniciativaDeTeste(t, "**Gate 1:** aprovado\n", planoDeTeste)
	specDeTeste(t, dir, "features", "feat-106-99-x.md", "feat-106-99 — inventada")

	// Task que não está no plano é recusada antes de qualquer escrita na fila.
	enfileiradas, _, err := SplitPeloBuilder(dir, "106-pr-draft-revisao-e-ci", []string{"feat-106-99"})
	require.ErrorIs(t, err, ErrSplitForaDoPlano)
	assert.Empty(t, enfileiradas)

	// Lista mista: recusa a inteira, sem enfileirar nenhuma.
	enfileiradas, _, err = SplitPeloBuilder(dir, "106-pr-draft-revisao-e-ci",
		[]string{"feat-106-21", "feat-106-99"})
	require.ErrorIs(t, err, ErrSplitForaDoPlano)
	assert.Empty(t, enfileiradas, "nada foi enfileirado quando uma task está fora do plano")
}

func TestSplitSemGate1Recusado(t *testing.T) {
	// status.md sem o marcador do Gate 1: o Builder não fatia.
	dir := iniciativaDeTeste(t, "# Status\n\nAguardando revisão.\n", planoDeTeste)
	_, _, err := SplitPeloBuilder(dir, "106-pr-draft-revisao-e-ci", nil)
	require.ErrorIs(t, err, ErrSplitSemGate1)

	// Plano vazio também é recusado.
	vazio := iniciativaDeTeste(t, "**Gate 1:** aprovado\n", "# Plano\n\nsem tasks\n")
	_, _, err = SplitPeloBuilder(vazio, "106-pr-draft-revisao-e-ci", nil)
	require.ErrorIs(t, err, ErrSplitVazio)
}

func TestOrdemDeEnfileiramentoRespeitaPlano(t *testing.T) {
	p := PlanoAprovado{Tasks: []string{"feat-1-a", "feat-1-b", "feat-1-c"}}
	ordem, err := p.OrdemDeEnfileiramento([]string{"feat-1-c", "feat-1-a"})
	require.NoError(t, err)
	assert.Equal(t, []string{"feat-1-a", "feat-1-c"}, ordem, "a ordem é a do plano, não a da chamada")

	_, err = p.OrdemDeEnfileiramento([]string{"feat-1-z"})
	require.ErrorIs(t, err, ErrSplitForaDoPlano)

	// Sem pedidas, devolve o plano inteiro.
	todas, err := p.OrdemDeEnfileiramento(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"feat-1-a", "feat-1-b", "feat-1-c"}, todas)
}
