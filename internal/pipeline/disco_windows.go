//go:build windows

package pipeline

import (
	"fmt"
	"syscall"
	"unsafe"
)

// medirVolume lê capacidade e uso do volume que contém path com GetDiskFreeSpaceExW (só stdlib).
func medirVolume(path string) (UsoDisco, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return UsoDisco{}, fmt.Errorf("medir disco de %s: %w", path, err)
	}
	var livreAoUsuario, total, livre uint64
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetDiskFreeSpaceExW")
	r, _, e := proc.Call(uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&livreAoUsuario)), uintptr(unsafe.Pointer(&total)), uintptr(unsafe.Pointer(&livre)))
	if r == 0 {
		return UsoDisco{}, fmt.Errorf("medir disco de %s: %w", path, e)
	}
	return UsoDisco{Total: total, Usado: total - livre}, nil
}
