package survey

import (
	"os"

	"github.com/mattn/go-isatty"
)

// isTerminalFn é a função de detecção de TTY; substituível nos testes.
var isTerminalFn = func(fd uintptr) bool { return isatty.IsTerminal(fd) }

// isTTY informa se o stdin atual é um terminal interativo que suporta modo raw
// (setas, bubbletea). Retorna false quando stdin for um pipe, redirecionamento
// ou for chamado fora de um terminal (CI, subprocesso com stdin redirecionado).
//
// Nota: cmd.exe clássico e Windows Terminal são ambos terminais do ponto de vista
// do SO; IsTerminal retorna true para ambos. O modo acessível só ativa quando
// stdin não é terminal (pipe, /dev/null, CI), não para distinguir VT no cmd.exe.
func isTTY() bool {
	return isTerminalFn(os.Stdin.Fd())
}
