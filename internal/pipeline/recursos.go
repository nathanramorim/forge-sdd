package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// Retrato de recursos da máquina para o painel e o diagnóstico. Quem mede é o worker do Builder, a
// cada GravarRecursosACada; o painel só lê o JSON (menos de 1 KB), nunca mede (desempenho é premissa).

// GravarRecursosACada é o intervalo entre dois retratos.
const GravarRecursosACada = 15 * time.Second

type WorktreeRetrato struct {
	Iniciativa string `json:"iniciativa"`
	Task       string `json:"task"`
	Estado     string `json:"estado"` // ativo, ocioso (retomada pendente) ou orfao (task já mesclada)
}

type Recursos struct {
	Em              time.Time         `json:"em"`
	Worktrees       []WorktreeRetrato `json:"worktrees"`
	TamanhoWorktree uint64            `json:"tamanho_worktree_bytes"`
	DiscoTotal      uint64            `json:"disco_total_bytes"`
	DiscoUsado      uint64            `json:"disco_usado_bytes"`
	MemoriaLivreMB  int               `json:"memoria_livre_mb"` // -1 = desconhecida
	MemoriaPorParMB int               `json:"memoria_por_par_mb"`
	Pares           int               `json:"pares"`
	Motivo          string            `json:"motivo"`
	ModoTask        bool              `json:"modo_task"`
}

func caminhoRecursos(dir string) string {
	return filepath.Join(layout.ForgeRuntime(dir), "painel", "recursos.json")
}

// MedirRecursos monta o retrato com medições baratas: statfs do volume, memória livre do sistema,
// tamanho do worktree já guardado em disco.json e a lista de pastas em .forge/runtime/worktrees.
func MedirRecursos(dir string, now time.Time) Recursos {
	r := Recursos{Em: now.UTC(), MemoriaLivreMB: -1, MemoriaPorParMB: MemoriaDoPar(dir)}
	if u, err := MedirDiscoDosWorktrees(dir); err == nil {
		r.DiscoTotal, r.DiscoUsado = u.Total, u.Usado
	}
	if e, ok := LerEstadoDisco(dir); ok {
		r.TamanhoWorktree = e.TamanhoWorktree
	}
	if mb, err := MemoriaDisponivelMB(); err == nil {
		r.MemoriaLivreMB = mb
	}
	ativos := map[string]bool{}
	for _, papel := range []string{layout.RoleBuilder, layout.RoleReviewer} {
		proc, _ := bus.List(dir, papel, layout.StateProcessing)
		for _, e := range proc {
			if e.Job.TaskID != "" {
				ativos[e.Job.Initiative+"/feat-"+e.Job.TaskID] = true
			}
		}
	}
	inis, _ := os.ReadDir(layout.Worktrees(dir))
	for _, ini := range inis {
		if !ini.IsDir() || !slugSeguro(ini.Name()) {
			continue
		}
		est, _ := EstadosDaIniciativa(dir, ini.Name())
		tasks, _ := os.ReadDir(filepath.Join(layout.Worktrees(dir), ini.Name()))
		for _, t := range tasks {
			if !t.IsDir() {
				continue
			}
			w := WorktreeRetrato{Iniciativa: ini.Name(), Task: t.Name(), Estado: "ocioso"}
			switch {
			case ativos[ini.Name()+"/"+t.Name()]:
				w.Estado = "ativo"
			case est[t.Name()] == TaskMesclada:
				w.Estado = "orfao"
			}
			r.Worktrees = append(r.Worktrees, w)
		}
	}
	sort.Slice(r.Worktrees, func(i, j int) bool {
		return r.Worktrees[i].Iniciativa+r.Worktrees[i].Task < r.Worktrees[j].Iniciativa+r.Worktrees[j].Task
	})
	if len(r.Worktrees) > 40 {
		r.Worktrees = r.Worktrees[:40]
	}
	if cfg, _, err := config.LoadForgeConfig(dir); err == nil {
		r.ModoTask = cfg.Builder.PorTask() && !cfg.IsAdvanced()
		if teto, terr := calculaTetoMem(dir, cfg.Builder, 1<<10, func(string) (UsoDisco, error) {
			return UsoDisco{Total: r.DiscoTotal, Usado: r.DiscoUsado}, nil
		}, nil, func() (int, error) {
			if r.MemoriaLivreMB < 0 {
				return 0, fmt.Errorf("desconhecida")
			}
			return r.MemoriaLivreMB, nil
		}, false); terr == nil {
			r.Pares, r.Motivo = teto.Pares, teto.Motivo
		}
	}
	return r
}

// GravarRecursos mede e grava o retrato (escrita atômica, só o worker escreve).
func GravarRecursos(dir string, now time.Time) error {
	r := MedirRecursos(dir, now)
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	p := caminhoRecursos(dir)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// LerRecursos devolve o último retrato, se existir.
func LerRecursos(dir string) (Recursos, bool) {
	var r Recursos
	data, err := os.ReadFile(caminhoRecursos(dir))
	if err != nil || len(data) > 16384 || json.Unmarshal(data, &r) != nil {
		return r, false
	}
	return r, true
}

// FormatRecursos é o bloco de recursos do diagnóstico.
func FormatRecursos(r Recursos, now time.Time) string {
	var b strings.Builder
	gb := func(v uint64) string { return fmt.Sprintf("%.1f GB", float64(v)/(1<<30)) }
	if r.DiscoTotal > 0 {
		fmt.Fprintf(&b, "disco: %s livres de %s (%d%% ocupado)\n", gb(r.DiscoTotal-r.DiscoUsado), gb(r.DiscoTotal), int(r.DiscoUsado*100/r.DiscoTotal))
	}
	if r.MemoriaLivreMB >= 0 {
		fmt.Fprintf(&b, "memória: %d MB livres · cerca de %d MB por par\n", r.MemoriaLivreMB, r.MemoriaPorParMB)
	}
	cont := map[string]int{}
	for _, w := range r.Worktrees {
		cont[w.Estado]++
	}
	fmt.Fprintf(&b, "worktrees: %d (ativos %d · ociosos %d · órfãos %d) · cerca de %s cada\n", len(r.Worktrees), cont["ativo"], cont["ocioso"], cont["orfao"], gb(r.TamanhoWorktree))
	if r.ModoTask {
		fmt.Fprintf(&b, "cabem agora: %d par(es) (limitado por %s)\n", r.Pares, r.Motivo)
	}
	if idade := now.Sub(r.Em); idade > time.Minute {
		fmt.Fprintf(&b, "(retrato de há %s; o worker do Builder atualiza a cada 15 s)\n", duracaoCurta(idade))
	}
	return b.String()
}
