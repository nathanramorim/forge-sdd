//go:build windows

package pipeline

import (
	"fmt"
	"syscall"
	"unsafe"
)

type memoriaStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// MemoriaDisponivelMB usa GlobalMemoryStatusEx via syscall (só stdlib).
func MemoriaDisponivelMB() (int, error) {
	st := memoriaStatusEx{Length: uint32(unsafe.Sizeof(memoriaStatusEx{}))}
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	r, _, e := proc.Call(uintptr(unsafe.Pointer(&st)))
	if r == 0 {
		return 0, fmt.Errorf("GlobalMemoryStatusEx: %w", e)
	}
	return int(st.AvailPhys >> 20), nil
}
