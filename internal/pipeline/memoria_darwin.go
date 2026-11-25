//go:build darwin

package pipeline

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
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
