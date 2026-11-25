package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
)

// IniciativasPorTask lista as iniciativas com esteira por task em andamento (pai salvo).
func IniciativasPorTask(dir string) []string {
	ents, err := os.ReadDir(filepath.Join(layout.ForgeRuntime(dir), "tarefas"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() && slugSeguro(e.Name()) {
			if _, err := os.Stat(filepath.Join(dirTarefas(dir, e.Name()), "pai.json")); err == nil {
				out = append(out, e.Name())
			}
		}
	}
	sort.Strings(out)
	return out
}

// tokensPorTask soma os tokens reais e estimados de sdd/.metrics por task do plano: o registro do
// worker usa o caminho da spec da task em `feature` (sessionRecordArgs).
func tokensPorTask(dir, ini string) map[string]int {
	out := map[string]int{}
	for _, m := range monitor.ReadSessions(dir, ini) {
		base := filepath.Base(filepath.ToSlash(m.Feature))
		if !strings.HasPrefix(base, "feat-") {
			continue
		}
		partes := strings.SplitN(strings.TrimSuffix(base, ".md"), "-", 4)
		if len(partes) < 3 {
			continue
		}
		id := strings.Join(partes[:3], "-")
		out[id] += m.TokensInput + m.TokensOutput + m.TokensCacheRead + m.TokensCacheCreation
	}
	return out
}

// FormatEsteiraPorTask resume a esteira por task (feat-122-08): pares ativos sobre o teto da
// máquina, o que segura o teto, e cada task com estado e tokens. Vazio quando não há esteira.
func FormatEsteiraPorTask(dir string, now time.Time) string {
	inis := IniciativasPorTask(dir)
	if len(inis) == 0 {
		return ""
	}
	cfg, _, _ := config.LoadForgeConfig(dir)
	var b strings.Builder
	b.WriteString("─── 🔀 ESTEIRA POR TASK ──────────────────────────────────────────────────────\n")
	for _, ini := range inis {
		plano, err := planoDeTasks(dir, ini, cfg.Builder.LarguraDePlano())
		if err != nil {
			fmt.Fprintf(&b, "%s · plano ilegível: %s\n", ini, firstLine(err.Error()))
			continue
		}
		est, _ := EstadosDaIniciativa(dir, ini)
		tok := tokensPorTask(dir, ini)
		ativos, mescladas, total := 0, 0, 0
		for _, onda := range plano {
			for _, t := range onda {
				total++
				switch est[t.ID] {
				case OndaEmAndamento:
					ativos++
				case TaskMesclada:
					mescladas++
				}
			}
		}
		largura := len(ElegiveisDoPlano(plano, est)) + ativos
		linha := fmt.Sprintf("%s · %d/%d mescladas · pares ativos %d", ini, mescladas, total, ativos)
		if largura > 0 {
			if teto, terr := calculaTetoMem(dir, cfg.Builder, largura, nil, func() (int, bool) { return LerJanelaDoPainel(dir, now) }, nil, false); terr == nil {
				linha += fmt.Sprintf(" de %d (%s) · vagas de disco %d", teto.Pares, teto.Motivo, teto.Vagas)
			}
		}
		b.WriteString(linha + "\n")
		for _, onda := range plano {
			for _, t := range onda {
				e := est[t.ID]
				if e == "" {
					e = TaskPendente
				}
				extra := ""
				if n := tok[t.ID]; n > 0 {
					extra = " · " + monitor.FormatTokensShort(n) + " tokens"
				}
				fmt.Fprintf(&b, "  %-14s %s%s\n", t.ID, strings.ReplaceAll(e, "-", " "), extra)
			}
		}
		if mescladas == total && total > 0 {
			fmt.Fprintf(&b, "  → feature pronta em %s: abra o PR em draft e use squash no merge\n", layout.InitiativeBranch(ini))
		}
	}
	return b.String()
}
