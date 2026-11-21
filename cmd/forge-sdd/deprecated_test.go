package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runSep executa o rootCmd separando stdout de stderr.
func runSep(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var o, e bytes.Buffer
	rootCmd.SetOut(&o)
	rootCmd.SetErr(&e)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return o.String(), e.String(), err
}

func TestComandosDepreciadosAvisam(t *testing.T) {
	dir := t.TempDir()
	casos := [][]string{
		{"queue", "stations", "--dir", dir},
		{"queue", "handoff", "--dir", dir}, // falta flag: erro, mas o aviso sai antes
	}
	for _, args := range casos {
		_, stderr, _ := runSep(t, args...)
		assert.Contains(t, stderr, "AVISO", strings.Join(args, " "))
		assert.Contains(t, stderr, "depreciado", strings.Join(args, " "))
		assert.NotContains(t, stderr, "\x1b[", "sem ANSI")
	}
	// attach: o aviso roda no PreRun, depois da validação de argumentos.
	assert.NotNil(t, attachCmd.PreRun)
	assert.Contains(t, attachCmd.Short, "[depreciado]")
	assert.Contains(t, queueStationsCmd.Short, "[depreciado]")
	assert.Contains(t, queueHandoffCmd.Short, "[depreciado]")
}

func TestDepreciadosContinuamFuncionando(t *testing.T) {
	dir := t.TempDir()
	_, stderr, err := runSep(t, "queue", "stations", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, stderr, "AVISO")

	// handoff segue validando as flags obrigatórias (mesmo erro de antes).
	_, _, err = runSep(t, "queue", "handoff", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flags --from, --to e --id são obrigatórias")

	// attach segue rejeitando estação inválida (mesmo erro), sem aviso antes da validação.
	_, _, err = runSep(t, "attach", "inexistente", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "estação inválida")

	// nada foi removido: continuam registrados e listados.
	assert.Equal(t, queueCmd, queueStationsCmd.Parent())
	assert.Equal(t, queueCmd, queueHandoffCmd.Parent())
	assert.Equal(t, rootCmd, attachCmd.Parent())
	assert.True(t, attachCmd.IsAvailableCommand(), "attach segue listado no --help")
	assert.True(t, queueStationsCmd.IsAvailableCommand())
	assert.True(t, queueHandoffCmd.IsAvailableCommand())
}

func TestComandosVigentesSemAviso(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"queue", "--dir", dir},
		{"queue", "enqueue", "--dir", dir},
	} {
		_, stderr, _ := runSep(t, args...)
		assert.NotContains(t, stderr, "depreciado", strings.Join(args, " "))
	}
}
