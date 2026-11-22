package doctor

import (
	"fmt"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
)

// checaFilaSemWorker avisa (nunca falha) quando há job na fila (estado pending) e nenhum worker
// ativo. Usa o detector de workers do monitor e a fila do bus.
func checaFilaSemWorker(dir string, d Deps) Checagem {
	const id = "fila-sem-worker"
	if d.Runner == nil {
		return ok(id, "listagem de processos indisponível")
	}
	jobs := 0
	for _, role := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		es, err := bus.List(dir, role, layout.StatePending)
		if err != nil {
			return ok(id, "fila ilegível: nada a avisar")
		}
		jobs += len(es)
	}
	if jobs == 0 {
		return ok(id, "nenhum job na fila")
	}
	n, err := monitor.ActiveWorkers(dir, d.Runner)
	if err == nil && n > 0 {
		return ok(id, fmt.Sprintf("%d job(s) na fila e %d worker(s) ativo(s)", jobs, n))
	}
	return aviso(id, fmt.Sprintf("%d job(s) na fila e nenhum worker ativo", jobs), "no chat, oriente os subagentes (veja a skill); ou inicie os workers se quiser a esteira")
}
