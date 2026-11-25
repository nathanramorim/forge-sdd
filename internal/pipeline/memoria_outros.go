//go:build !linux && !darwin && !windows

package pipeline

import "errors"

// MemoriaDisponivelMB não tem leitura neste sistema: a trava de memória fica desligada.
func MemoriaDisponivelMB() (int, error) {
	return 0, errors.New("memória disponível não suportada neste sistema")
}
