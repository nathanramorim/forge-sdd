package survey

import (
	"os"

	"golang.org/x/term"
)

// isTerminalFn é a função de detecção de TTY; substituível nos testes.
var isTerminalFn = func(fd int) bool { return term.IsTerminal(fd) }

// isTTY informa se a stdin atual é um terminal interativo que suporta modo raw
// (setas, bubbletea). Retorna false quando o stdin for um pipe, redirecionamento
// ou um console sem Virtual Terminal Processing (cmd.exe clássico sem VT).
func isTTY() bool {
	return isTerminalFn(int(os.Stdin.Fd()))
}
