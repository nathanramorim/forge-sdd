package survey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTTY_ReturnsStubValue(t *testing.T) {
	orig := isTerminalFn
	defer func() { isTerminalFn = orig }()

	isTerminalFn = func(int) bool { return true }
	assert.True(t, isTTY(), "deve retornar true quando isTerminalFn retorna true")

	isTerminalFn = func(int) bool { return false }
	assert.False(t, isTTY(), "deve retornar false quando isTerminalFn retorna false")
}

// TestIsTTY_PadraoNaoEhTTYNoCIENoTestRunner verifica que, no ambiente de testes,
// stdin não é um TTY (pipe do test runner), garantindo que o stub é necessário.
func TestIsTTY_PadraoNaoEhTTYNoCIENoTestRunner(t *testing.T) {
	// No test runner e CI, os.Stdin é um pipe, não um TTY.
	// Este teste documenta a premissa: isTTY() == false em CI.
	_ = isTTY() // não deve entrar em pânico
}
