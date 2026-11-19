package survey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTTY_ReturnsStubValue(t *testing.T) {
	orig := isTerminalFn
	defer func() { isTerminalFn = orig }()

	isTerminalFn = func(uintptr) bool { return true }
	assert.True(t, isTTY(), "deve retornar true quando isTerminalFn retorna true")

	isTerminalFn = func(uintptr) bool { return false }
	assert.False(t, isTTY(), "deve retornar false quando isTerminalFn retorna false")
}

// TestIsTTY_NaoPanica verifica que a chamada padrão (sem stub) não entra em pânico,
// independente de o ambiente de teste ter ou não um TTY.
func TestIsTTY_NaoPanica(t *testing.T) {
	assert.NotPanics(t, func() { isTTY() })
}
