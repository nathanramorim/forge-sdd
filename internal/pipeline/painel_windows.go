//go:build windows

package pipeline

import (
	"os/exec"
	"syscall"
)

// Sinalizadores de criação de processo do Windows (CREATE_NEW_PROCESS_GROUP e DETACHED_PROCESS).
const (
	createNewProcessGroup = 0x00000200
	detachedProcess       = 0x00000008
)

// detach solta o processo do console que o abriu.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup | detachedProcess}
}

// processAlive não tem sonda barata no Windows sem dependência: considera que o pid não vive e o
// worker é iniciado de novo (o barramento resolve a disputa por job com rename atômico).
func processAlive(pid int) bool { return false }
