package queue

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenPR_GhNaoEncontrado(t *testing.T) {
	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "", errors.New("not found") }

	callCount := 0
	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })
	prRunCmd = func(_, name string, _ ...string) (string, error) {
		callCount++
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: "feat/100-test"}
	item := &QueueItem{TaskID: "100-01", GovernanceVerdict: &Verdict{Veredito: VerdictConforme, Resumo: "ok"}}
	err := OpenPR(t.TempDir(), q, item, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "gh não encontrado")
	assert.Equal(t, 0, callCount, "nenhum comando deve ser executado quando gh está ausente")
}

func TestOpenPR_BranchVazia(t *testing.T) {
	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "/usr/bin/gh", nil }

	callCount := 0
	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })
	prRunCmd = func(_, name string, _ ...string) (string, error) {
		callCount++
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: ""}
	item := &QueueItem{TaskID: "100-01", GovernanceVerdict: &Verdict{Veredito: VerdictConforme, Resumo: "ok"}}
	err := OpenPR(t.TempDir(), q, item, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "branch não definida")
	assert.Equal(t, 0, callCount, "nenhum comando deve ser executado com branch vazia")
}

func TestOpenPR_SemOrigin(t *testing.T) {
	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "/usr/bin/gh", nil }

	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })
	pushCalled := false
	prRunCmd = func(_, name string, args ...string) (string, error) {
		if name == "git" && len(args) > 0 && args[0] == "config" {
			return "", errors.New("no origin")
		}
		if name == "git" && len(args) > 0 && args[0] == "push" {
			pushCalled = true
		}
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: "feat/100-test"}
	item := &QueueItem{TaskID: "100-01", GovernanceVerdict: &Verdict{Veredito: VerdictConforme, Resumo: "ok"}}
	err := OpenPR(t.TempDir(), q, item, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "remote origin não encontrado")
	assert.False(t, pushCalled, "push não deve ser chamado sem origin")
}

func TestOpenPR_PRJaExiste(t *testing.T) {
	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "/usr/bin/gh", nil }

	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })
	pushCalled := false
	prRunCmd = func(_, name string, args ...string) (string, error) {
		if name == "git" && len(args) > 0 && args[0] == "config" {
			return "git@github.com:org/repo.git", nil
		}
		if name == "gh" && len(args) > 0 && args[0] == "pr" && args[1] == "list" {
			return `[{"url":"https://github.com/org/repo/pull/42"}]`, nil
		}
		if name == "git" && len(args) > 0 && args[0] == "push" {
			pushCalled = true
		}
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: "feat/100-test"}
	item := &QueueItem{TaskID: "100-01", GovernanceVerdict: &Verdict{Veredito: VerdictConforme, Resumo: "ok"}}
	err := OpenPR(t.TempDir(), q, item, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "PR já existe")
	assert.Contains(t, buf.String(), "https://github.com/org/repo/pull/42")
	assert.False(t, pushCalled, "push não deve ser chamado quando PR já existe")
}

func TestOpenPR_CaminhoFeliz(t *testing.T) {
	dir := t.TempDir()
	specPath := writeSpecFile(t, dir, "sdd/features/feat-100/task-01.md", validSpecContent)

	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "/usr/bin/gh", nil }

	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })

	var calls []string
	prRunCmd = func(_, name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		if name == "git" && len(args) > 0 && args[0] == "config" {
			return "git@github.com:org/repo.git", nil
		}
		if name == "gh" && len(args) > 0 && args[0] == "pr" && args[1] == "list" {
			return `[]`, nil
		}
		if name == "gh" && len(args) > 0 && args[0] == "pr" && args[1] == "create" {
			return "https://github.com/org/repo/pull/99", nil
		}
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: "feat/100-auto-pr", BaseBranch: ""}
	item := &QueueItem{
		TaskID:   "100-01",
		SpecPath: specPath,
		GovernanceVerdict: &Verdict{
			Veredito: VerdictConforme,
			Resumo:   "implementação conforme",
		},
	}
	err := OpenPR(dir, q, item, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "→ PR aberto:")
	assert.Contains(t, buf.String(), "https://github.com/org/repo/pull/99")

	// Verifica push chamado com a branch correta
	var pushCall string
	for _, c := range calls {
		if strings.HasPrefix(c, "git push") {
			pushCall = c
		}
	}
	assert.Contains(t, pushCall, "feat/100-auto-pr", "push deve usar a branch da fila")

	// Verifica gh pr create com --base main (BaseBranch vazio → "main")
	var createCall string
	for _, c := range calls {
		if strings.HasPrefix(c, "gh pr create") {
			createCall = c
		}
	}
	assert.Contains(t, createCall, "--base main", "base deve ser main quando BaseBranch está vazio")
}

func TestOpenPR_CaminhoFelizComBaseBranch(t *testing.T) {
	dir := t.TempDir()
	specPath := writeSpecFile(t, dir, "sdd/features/feat-100/task-01.md", validSpecContent)

	orig := prLookPath
	t.Cleanup(func() { prLookPath = orig })
	prLookPath = func(string) (string, error) { return "/usr/bin/gh", nil }

	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })

	var calls []string
	prRunCmd = func(_, name string, args ...string) (string, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "git" && len(args) > 0 && args[0] == "config" {
			return "git@github.com:org/repo.git", nil
		}
		if name == "gh" && len(args) > 0 && args[0] == "pr" && args[1] == "list" {
			return `[]`, nil
		}
		if name == "gh" && len(args) > 0 && args[0] == "pr" && args[1] == "create" {
			return "https://github.com/org/repo/pull/100", nil
		}
		return "", nil
	}

	var buf bytes.Buffer
	q := &Queue{Branch: "feat/100-auto-pr", BaseBranch: "develop"}
	item := &QueueItem{
		TaskID:   "100-01",
		SpecPath: specPath,
		GovernanceVerdict: &Verdict{
			Veredito: VerdictConforme,
			Resumo:   "conforme",
		},
	}
	err := OpenPR(dir, q, item, &buf)
	require.NoError(t, err)

	var createCall string
	for _, c := range calls {
		if strings.HasPrefix(c, "gh pr create") {
			createCall = c
		}
	}
	assert.Contains(t, createCall, "--base develop", "base deve ser BaseBranch quando definido")
	assert.Contains(t, createCall, "--draft", "PR nasce em draft (RN-7)")
}

func TestOpenPR_VCSNenhumNaoExecutaComando(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "constitution.md"),
		[]byte("| VCS / Work Item System | nenhum |\n"), 0o644))
	origRun := prRunCmd
	t.Cleanup(func() { prRunCmd = origRun })
	chamadas := 0
	prRunCmd = func(_, _ string, _ ...string) (string, error) { chamadas++; return "", nil }

	var buf bytes.Buffer
	err := OpenPR(dir, &Queue{Branch: "feat/x"}, &QueueItem{TaskID: "1"}, &buf)
	require.NoError(t, err)
	assert.Zero(t, chamadas)
	assert.Contains(t, buf.String(), "VCS nenhum")

	// VCS desconhecido é recusado
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "memory", "constitution.md"),
		[]byte("| VCS / Work Item System | gitlab |\n"), 0o644))
	assert.Error(t, OpenPR(dir, &Queue{Branch: "feat/x"}, &QueueItem{TaskID: "1"}, &buf))
	assert.Zero(t, chamadas)
}
