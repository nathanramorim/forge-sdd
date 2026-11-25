//go:build !windows

package pipeline

import (
	"fmt"
	"syscall"
)

// medirVolume lê capacidade e uso do volume que contém path (só stdlib, ARQ-04).
func medirVolume(path string) (UsoDisco, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return UsoDisco{}, fmt.Errorf("medir disco de %s: %w", path, err)
	}
	bs := uint64(st.Bsize)
	total := uint64(st.Blocks) * bs
	usado := (uint64(st.Blocks) - uint64(st.Bfree)) * bs
	return UsoDisco{Total: total, Usado: usado}, nil
}
