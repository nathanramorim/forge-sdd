package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
)

// refsDaTask reúne as referências próprias da task a partir do pacote de contexto dela (RN-2).
func refsDaTask(dir string, adrDirs []string, trel string, pack gate.ContextPack, rulesSection string) *taskJobRefs {
	refs := &taskJobRefs{Spec: trel}
	for _, r := range pack.Rules {
		refs.Rules = append(refs.Rules, r.Path)
	}
	for _, tok := range adrTokensIn(rulesSection) {
		if p, ok := findADR(dir, adrDirs, tok); ok {
			refs.ADRs = append(refs.ADRs, p)
		}
	}
	return refs
}

// tetoDaRodada mede o disco e calcula o teto de pares para a rodada (RN-5). Usa as injeções de opts
// quando existem; com gravar, persiste o tamanho do worktree medido.
func tetoDaRodada(dir string, b config.BuilderConfig, largura int, opts DispatchOptions, gravar bool) (Teto, error) {
	return calculaTetoMem(dir, b, largura, opts.Disco, opts.JanelaPct, opts.Memoria, gravar)
}

func calculaTetoMem(dir string, b config.BuilderConfig, largura int, disco func(string) (UsoDisco, error), janela func() (int, bool), mem func() (int, error), gravar bool) (Teto, error) {
	if disco == nil {
		disco = MedirDiscoDosWorktrees
	}
	if mem == nil {
		mem = MemoriaDisponivelMB
	}
	uso, err := disco(dir)
	if err != nil {
		return Teto{}, err
	}
	tam, err := TamanhoDoWorktree(dir, gravar)
	if err != nil {
		return Teto{}, err
	}
	pct := -1
	if janela != nil {
		if v, ok := janela(); ok {
			pct = v
		}
	}
	memMB := -1 // desconhecida: a trava de memória não vale
	if v, merr := mem(); merr == nil {
		memMB = v
	}
	return TetoDeParalelo(ParamsTeto{
		Disco: uso, LimiteDiscoPct: b.LimiteDiscoEfetivo(), TamanhoWorktree: tam,
		LarguraOnda: largura, MaxParalelo: b.MaxParallelTasks, JanelaPct: pct,
		MemoriaDisponivelM: memMB, ReservaMemoriaMB: b.ReservaMemoriaEfetiva(), MemoriaPorParMB: MemoriaDoPar(dir),
	}), nil
}

// EntradaDespacho é o que uma rodada do despachante recebe. Estados é o quadro do plano indexado
// pelo id do plano ("feat-122-01"): pendente, em-andamento, mesclada ou bloqueada.
type EntradaDespacho struct {
	Dir     string
	Pai     bus.Job // o feature_build que nomeia os jobs (Initiative preenchida)
	Plano   [][]Task
	Estados map[string]string
	Builder config.BuilderConfig
	Disco   func(dir string) (UsoDisco, error)
	Janela  func() (int, bool)
	Memoria func() (int, error) // memória disponível em MB; nil lê a da máquina
	Now     time.Time
}

// ResultadoDespacho descreve a rodada.
type ResultadoDespacho struct {
	Teto      Teto
	Ativos    int
	Elegiveis int
	Jobs      []bus.Job // task_build enfileirados nesta rodada
}

// Despachar roda uma rodada: calcula o teto, desconta os pares já ativos e enfileira de uma vez
// todas as tasks elegíveis que cabem (RN-7). Chamar de novo a cada worktree apagado recalcula; várias
// vagas abertas juntas põem várias tasks para começar juntas. É idempotente: job existente não
// duplica.
func Despachar(in EntradaDespacho) (*ResultadoDespacho, error) {
	elegiveis := ElegiveisDoPlano(in.Plano, in.Estados)
	ativos := 0
	for _, e := range in.Estados {
		if e == OndaEmAndamento {
			ativos++
		}
	}
	res := &ResultadoDespacho{Ativos: ativos, Elegiveis: len(elegiveis)}
	if len(elegiveis) == 0 {
		return res, nil
	}
	teto, err := calculaTetoMem(in.Dir, in.Builder, len(elegiveis)+ativos, in.Disco, in.Janela, in.Memoria, true)
	if err != nil {
		return nil, err
	}
	res.Teto = teto
	livres := teto.Pares - ativos
	if livres <= 0 {
		return res, nil
	}
	if livres < len(elegiveis) {
		elegiveis = elegiveis[:livres]
	}
	adrDirs := []string{"sdd/spec/adr"}
	if cfg, found, cerr := config.LoadForgeConfig(in.Dir); cerr == nil && found {
		adrDirs = cfg.ADRDirs
	}
	var jobs []bus.Job
	for _, t := range elegiveis {
		trel, rerr := filepath.Rel(in.Dir, t.File)
		if rerr != nil {
			return nil, fmt.Errorf("task %s: %w", t.ID, rerr)
		}
		trel = filepath.ToSlash(trel)
		pack, cerr := gate.BuildContext(in.Dir, trel, nil, "build")
		if cerr != nil {
			return nil, fmt.Errorf("task %s: %w", t.ID, cerr)
		}
		data, derr := os.ReadFile(t.File)
		if derr != nil {
			return nil, fmt.Errorf("task %s: %w", t.ID, derr)
		}
		refs := refsDaTask(in.Dir, adrDirs, trel, pack, sectionBody(string(data), "rules e ids da constituição aplicáveis"))
		jobs = append(jobs, TaskJobFrom(in.Pai, idCurtoDaTask(t.ID), tituloDoArquivo(t.File, t.ID), *refs))
	}
	if _, _, err := EnfileiraTasks(in.Dir, jobs, in.Now); err != nil {
		return nil, err
	}
	res.Jobs = jobs
	return res, nil
}
