//go:build darwin

package pipeline

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	vmPaginaRe = regexp.MustCompile(`page size of (\d+) bytes`)
	vmLinhaRe  = regexp.MustCompile(`(?m)^Pages (free|inactive|speculative):\s+(\d+)\.`)
)

// MemoriaDisponivelMB soma as páginas livres, inativas e especulativas do vm_stat (sem argumentos
// vindos de fora).
func MemoriaDisponivelMB() (int, error) {
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0, err
	}
	return memDoVMStat(string(out))
}

// MemoriaTotalMB lê hw.memsize pelo sysctl (sem argumentos vindos de fora).
func MemoriaTotalMB() (int, error) {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(b >> 20), nil
}

func memDoVMStat(s string) (int, error) {
	m := vmPaginaRe.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("vm_stat sem tamanho de página")
	}
	pg, _ := strconv.Atoi(m[1])
	var paginas int
	for _, l := range vmLinhaRe.FindAllStringSubmatch(s, -1) {
		n, _ := strconv.Atoi(l[2])
		paginas += n
	}
	if paginas == 0 {
		return 0, fmt.Errorf("vm_stat sem páginas livres")
	}
	return paginas * pg / (1 << 20), nil
}
