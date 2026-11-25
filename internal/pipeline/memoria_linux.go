//go:build linux

package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MemoriaDisponivelMB lê MemAvailable de /proc/meminfo (só stdlib).
func MemoriaDisponivelMB() (int, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return memAvailableDe(bufio.NewScanner(f))
}

func memAvailableDe(sc *bufio.Scanner) (int, error) {
	for sc.Scan() {
		campos := strings.Fields(sc.Text())
		if len(campos) >= 2 && campos[0] == "MemAvailable:" {
			kb, err := strconv.Atoi(campos[1])
			if err != nil {
				return 0, err
			}
			return kb / 1024, nil
		}
	}
	return 0, fmt.Errorf("MemAvailable ausente em /proc/meminfo")
}
