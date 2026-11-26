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

// MemoriaTotalMB lê MemTotal de /proc/meminfo.
func MemoriaTotalMB() (int, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return campoMeminfo(bufio.NewScanner(f), "MemTotal:")
}

func memAvailableDe(sc *bufio.Scanner) (int, error) {
	return campoMeminfo(sc, "MemAvailable:")
}

func campoMeminfo(sc *bufio.Scanner, campo string) (int, error) {
	for sc.Scan() {
		campos := strings.Fields(sc.Text())
		if len(campos) >= 2 && campos[0] == campo {
			kb, err := strconv.Atoi(campos[1])
			if err != nil {
				return 0, err
			}
			return kb / 1024, nil
		}
	}
	return 0, fmt.Errorf("%s ausente em /proc/meminfo", campo)
}
