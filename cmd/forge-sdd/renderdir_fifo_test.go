//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FIFO só existe em sistemas POSIX; no Windows o teste não compila.
func TestRenderDirFIFONaoTravaNemSubstitui(t *testing.T) {
	dir := t.TempDir()
	initClaude(t, dir)
	alvo := filepath.Join(dir, ".claude", "commands", "revisar.md")
	require.NoError(t, os.Remove(alvo))
	require.NoError(t, syscall.Mkfifo(alvo, 0o644))
	done := make(chan struct{})
	go func() { defer close(done); updateAntigo(t, dir) }()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("update travou no FIFO")
	}
	fi, err := os.Lstat(alvo)
	require.NoError(t, err)
	assert.NotZero(t, fi.Mode()&os.ModeNamedPipe)
}
