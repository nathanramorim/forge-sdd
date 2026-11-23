package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hookProject cria um projeto com (ou sem) scripts/pr_guard.py e aponta CLAUDE_PROJECT_DIR para ele.
func hookProject(t *testing.T, comScript bool) string {
	t.Helper()
	dir := t.TempDir()
	if comScript {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "pr_guard.py"), []byte("# stub\n"), 0o644))
	}
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	return dir
}

// fakeUV instala um `uv` FALSO no PATH: grava os argumentos e o stdin e sai com o código dado.
func fakeUV(t *testing.T, exit int) (argsFile, stdinFile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub de uv em shell POSIX")
	}
	bin := t.TempDir()
	out := t.TempDir()
	argsFile, stdinFile = filepath.Join(out, "args"), filepath.Join(out, "stdin")
	script := "#!/bin/sh\n/bin/echo \"$@\" > '" + argsFile + "'\n/bin/cat > '" + stdinFile + "'\n/bin/echo aviso-do-guard >&2\nexit " + strconv.Itoa(exit) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "uv"), []byte(script), 0o755))
	t.Setenv("PATH", bin)
	return argsFile, stdinFile
}

func runHookPRGuard(t *testing.T, stdin string) (code int, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errb)
	rootCmd.SetArgs([]string{"hook", "pr-guard"})
	t.Cleanup(func() { rootCmd.SetIn(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
	err := rootCmd.Execute()
	if err == nil {
		return 0, errb.String()
	}
	var ec exitCoder
	require.True(t, errors.As(err, &ec), "erro inesperado: %v", err)
	return ec.ExitCode(), errb.String()
}

func TestHookPRGuardSemScriptSaiZero(t *testing.T) {
	hookProject(t, false)
	fakeUV(t, 2) // mesmo com uv presente, sem script não bloqueia
	code, stderr := runHookPRGuard(t, `{"tool_name":"Bash"}`)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "scripts")
	assert.Contains(t, stderr, "pr_guard.py")
}

func TestHookPRGuardRepassaCodigo2(t *testing.T) {
	hookProject(t, true)
	args, _ := fakeUV(t, 2)
	code, stderr := runHookPRGuard(t, `{}`)
	assert.Equal(t, 2, code, "o bloqueio real do guard permanece")
	assert.Contains(t, stderr, "aviso-do-guard")
	got, _ := os.ReadFile(args)
	assert.Equal(t, "run "+filepath.Join("scripts", "pr_guard.py")+" --hook", strings.TrimSpace(string(got)))
}

func TestHookPRGuardRepassaStdin(t *testing.T) {
	hookProject(t, true)
	_, stdin := fakeUV(t, 0)
	payload := `{"tool_name":"Bash","tool_input":{"command":"ls"}}`
	code, _ := runHookPRGuard(t, payload)
	assert.Equal(t, 0, code)
	got, _ := os.ReadFile(stdin)
	assert.Equal(t, payload, string(got))
}

func TestHookPRGuardSemUvSaiZero(t *testing.T) {
	hookProject(t, true)
	t.Setenv("PATH", t.TempDir()) // PATH sem uv
	code, stderr := runHookPRGuard(t, `{}`)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "uv")
}

func TestDoctorReconheceOsDoisComandosDoHook(t *testing.T) {
	for nome, cmd := range map[string]string{"novo": "forge hook pr-guard", "antigo": "uv run scripts/pr_guard.py --hook"} {
		t.Run(nome, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
			cfg := `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"` + cmd + `"}]}]}}`
			require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(cfg), 0o644))
			out := captureStdout(t, func() { reportPRGuardCoverage(dir, []string{"claude"}) })
			assert.Contains(t, out, "Hook PreToolUse instalado")
			assert.NotContains(t, out, "Hook PreToolUse ausente")
		})
	}
}
