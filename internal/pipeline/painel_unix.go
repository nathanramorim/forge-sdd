//go:build !windows

package pipeline

import (
	"os"
	"os/exec"
	"syscall"
)

// detach põe o processo em uma sessão própria: ele sobrevive ao fechamento do terminal que o abriu.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// processAlive testa o pid com o sinal 0 (não envia nada).
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
