package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "memory"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0o755))

	constContent := `# Constituição — forge-sdd

- **ARQ-01**: Separação estrita entre pacotes de domínio e interface CLI.
- **DUP-02**: Toda spec deve documentar o que já existe antes de criar código novo.
- **SEG-04**: Segredos e chaves privadas nunca são versionados no repositório.
- **SEG-07**: Dependências externas exigem justificativa explícita na Spec.
- **FLX-01**: Nenhuma implementação começa sem Spec revisada.
- **FLX-02**: Commits diretos em main/master são proibidos.
- **GOV-02**: Exceções à Constituição exigem bloco constitution-exception com prazo.
- **GOV-04**: Rules em .agents/rules/ citam IDs da Constituição sem duplicar seu texto.
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "constitution.md"), []byte(constContent), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "progress.md"), []byte("# Progress\nStatus: ok\n"), 0o644))

	ruleContent := `---
description: Regras de arquitetura
globs: ["internal/**/*.go"]
regras: [ARQ-01, DUP-02]
---

# Como cumprir
Mantenha os pacotes coesos e consulte os pacotes existentes antes de criar novos.
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".agents", "rules", "sdd-arquitetura.md"), []byte(ruleContent), 0o644))

	return dir
}

func findCheck(results []CheckResult, id string) (CheckResult, bool) {
	for _, r := range results {
		if r.ID == id {
			return r, true
		}
	}
	return CheckResult{}, false
}

func TestRunChecksAllPassing(t *testing.T) {
	dir := setupGateDir(t)

	specRel := "sdd/features/feat-58/task-01.md"
	specFull := filepath.Join(dir, specRel)
	require.NoError(t, os.MkdirAll(filepath.Dir(specFull), 0o755))
	specBody := `# Task 01

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
| --- | --- | --- | --- | --- |
| Gate | internal/ | Nada | criar | Novo pacote internal/gate |

## Dependências novas
- spf13/cobra: subcomando CLI
`
	require.NoError(t, os.WriteFile(specFull, []byte(specBody), 0o644))

	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	prText := `Descrição do PR

constitution-exception:
  regra: ARQ-01
  motivo: Migração gradual
  prazo: 2026-12-31
  responsavel: @nathan
`
	results, err := RunChecks(dir, Options{
		Branch:       "feat/58-governanca",
		SpecPath:     specRel,
		GoModChanged: true,
		PRBodyText:   prText,
		Now:          now,
	})
	require.NoError(t, err)
	assert.False(t, HasFailures(results))

	for _, id := range []string{"P-01", "P-02", "P-03", "P-13", "P-14", "P-20", "P-21", "P-23"} {
		chk, ok := findCheck(results, id)
		require.True(t, ok, "esperava encontrar check %s", id)
		assert.True(t, chk.Passed, "check %s deveria passar: %s", id, chk.Evidence)
	}
}

func TestCheckP01Branch(t *testing.T) {
	dir := setupGateDir(t)

	results, err := RunChecks(dir, Options{Branch: "main"})
	require.NoError(t, err)
	p01, ok := findCheck(results, "P-01")
	require.True(t, ok)
	assert.False(t, p01.Passed)

	resultsMaster, err := RunChecks(dir, Options{Branch: "master"})
	require.NoError(t, err)
	p01Master, _ := findCheck(resultsMaster, "P-01")
	assert.False(t, p01Master.Passed)
}

func TestCheckP02AndP03SpecAndReuse(t *testing.T) {
	dir := setupGateDir(t)

	// Spec inexistente
	results, err := RunChecks(dir, Options{SpecPath: "sdd/features/inexistente.md"})
	require.NoError(t, err)
	p02, _ := findCheck(results, "P-02")
	p03, _ := findCheck(results, "P-03")
	assert.False(t, p02.Passed)
	assert.False(t, p03.Passed)

	// Spec sem decisão válida
	badSpec := filepath.Join(dir, "sdd/features/bad.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(badSpec), 0o755))
	require.NoError(t, os.WriteFile(badSpec, []byte("# Spec\n\n## O que já existe\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| X | Y | Z | ignorar | - |\n"), 0o644))

	results2, err := RunChecks(dir, Options{SpecPath: "sdd/features/bad.md"})
	require.NoError(t, err)
	p02Ok, _ := findCheck(results2, "P-02")
	p03Bad, _ := findCheck(results2, "P-03")
	assert.True(t, p02Ok.Passed)
	assert.False(t, p03Bad.Passed)
}

func TestCheckP13Secrets(t *testing.T) {
	dir := setupGateDir(t)
	 fakeToken := "ghp_" + strings.Repeat("aB9", 10)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "leaked.go"), []byte("package main\nvar tok = \""+fakeToken+"\"\n"), 0o644))

	results, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	p13, ok := findCheck(results, "P-13")
	require.True(t, ok)
	assert.False(t, p13.Passed)
	assert.Contains(t, p13.Evidence, "leaked.go")
}

func TestCheckP14NewDependencies(t *testing.T) {
	dir := setupGateDir(t)
	specRel := "sdd/features/spec.md"
	specFull := filepath.Join(dir, specRel)
	require.NoError(t, os.MkdirAll(filepath.Dir(specFull), 0o755))
	require.NoError(t, os.WriteFile(specFull, []byte("# Spec\n\n## O que já existe\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| A | B | C | criar | Novo |\n\n## Dependências novas\nNenhuma.\n"), 0o644))

	results, err := RunChecks(dir, Options{SpecPath: specRel, GoModChanged: true})
	require.NoError(t, err)
	p14, ok := findCheck(results, "P-14")
	require.True(t, ok)
	assert.False(t, p14.Passed)
}

func TestCheckP20ConstitutionException(t *testing.T) {
	dir := setupGateDir(t)
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	// Prazo expirado
	expiredPR := `constitution-exception:
  regra: ARQ-01
  motivo: Teste expirado
  prazo: 2026-10-01
  responsavel: @dev
`
	results, err := RunChecks(dir, Options{PRBodyText: expiredPR, Now: now})
	require.NoError(t, err)
	p20, ok := findCheck(results, "P-20")
	require.True(t, ok)
	assert.False(t, p20.Passed)
	assert.Contains(t, p20.Evidence, "expirado")

	// Faltando responsavel
	missingFieldPR := `constitution-exception:
  regra: ARQ-01
  motivo: Sem responsável
  prazo: 2026-12-01
`
	results2, err := RunChecks(dir, Options{PRBodyText: missingFieldPR, Now: now})
	require.NoError(t, err)
	p20Missing, _ := findCheck(results2, "P-20")
	assert.False(t, p20Missing.Passed)
	assert.Contains(t, p20Missing.Evidence, "responsavel")
}

func TestCheckP21ProgressBudget(t *testing.T) {
	dir := setupGateDir(t)
	bigContent := strings.Repeat("a", MaxProgressBytes+10)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "progress.md"), []byte(bigContent), 0o644))

	results, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	p21, ok := findCheck(results, "P-21")
	require.True(t, ok)
	assert.False(t, p21.Passed)
}

func TestCheckP23RulesGovernance(t *testing.T) {
	dir := setupGateDir(t)

	// 1. ID inexistente na constitution
	badIDRule := `---
description: Rule inválida
regras: [ARQ-99]
---
Texto próprio.
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".agents", "rules", "bad-id.md"), []byte(badIDRule), 0o644))
	results, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	p23, _ := findCheck(results, "P-23")
	assert.False(t, p23.Passed)
	assert.Contains(t, p23.Evidence, "ARQ-99")

	require.NoError(t, os.Remove(filepath.Join(dir, ".agents", "rules", "bad-id.md")))

	// 2. Cópia literal de linha inteira da constitution
	duplicatedLineRule := `---
description: Rule com cópia literal
regras: [ARQ-01]
---
- **ARQ-01**: Separação estrita entre pacotes de domínio e interface CLI.
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".agents", "rules", "dup-line.md"), []byte(duplicatedLineRule), 0o644))
	results2, err := RunChecks(dir, Options{})
	require.NoError(t, err)
	p23Dup, _ := findCheck(results2, "P-23")
	assert.False(t, p23Dup.Passed)
	assert.Contains(t, p23Dup.Evidence, "cópia literal")
}
