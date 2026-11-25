package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gb = uint64(1) << 30
)

// discoFalso simula um volume de 100 GB com o uso dado em GB.
func discoFalso(usadoGB uint64) func(string) (UsoDisco, error) {
	return func(string) (UsoDisco, error) { return UsoDisco{Total: 100 * gb, Usado: usadoGB * gb}, nil }
}

func memFalsa(mbDisp int) func() (int, error) { return func() (int, error) { return mbDisp, nil } }

func paramsBase() ParamsTeto {
	return ParamsTeto{
		Disco: UsoDisco{Total: 100 * gb, Usado: 50 * gb}, LimiteDiscoPct: 80, TamanhoWorktree: 5 * gb,
		LarguraOnda: 100, MemoriaDisponivelM: -1, JanelaPct: -1,
	}
}

func TestTeto_Disco(t *testing.T) {
	p := paramsBase()
	assert.Equal(t, 8, VagasDeDisco(p.Disco, 80, 5*gb), "50% de 100 GB, worktree de 5 GB: a reserva fica em 10 GB (menor que 20%), 8 vagas")
	assert.Equal(t, 8, TetoDeParalelo(p).Pares)

	p.Disco.Usado = 89 * gb
	assert.Equal(t, 0, TetoDeParalelo(p).Pares, "com 11 GB livres e 10 de reserva não cabe um worktree de 5 GB")

	p.Disco.Usado = 90 * gb
	assert.Equal(t, 0, VagasDeDisco(p.Disco, 80, 5*gb), "acima do limite nunca fica negativo")

	p = paramsBase()
	p.LarguraOnda = 3
	assert.Equal(t, Teto{Vagas: 8, VagasMemoria: semLimite, Largura: 3, Janela: semLimite, Fixo: semLimite, Pares: 3, Motivo: "onda"}, TetoDeParalelo(p))
}

// Caso real: Mac de 228 GB com 87% ocupado por outras coisas e 29 GB livres. A regra antiga (20%
// livre) dava 0 vagas e travava a esteira; com a reserva de 10 GB cabem dezenas de worktrees.
func TestTeto_DiscoGrandeCheioNaoTrava(t *testing.T) {
	u := UsoDisco{Total: 228 * gb, Usado: 199 * gb}
	assert.Equal(t, 38, VagasDeDisco(u, 80, gb/2))
	u = UsoDisco{Total: 40 * gb, Usado: 33 * gb}
	assert.Equal(t, 0, VagasDeDisco(u, 80, gb/2), "disco pequeno mantém os 20% de reserva")
}

func TestTeto_Janela(t *testing.T) {
	p := paramsBase()
	p.JanelaPct = 90
	p.LimiteJanelaPct = 85
	assert.Equal(t, 1, TetoDeParalelo(p).Pares, "janela em 90% com limite 85% derruba para 1 par")
	p.JanelaPct = 85
	assert.Equal(t, 8, TetoDeParalelo(p).Pares, "no limite ainda vale o teto normal")
	p.JanelaPct = 99
	p.LimiteJanelaPct = 0
	assert.Equal(t, 1, TetoDeParalelo(p).Pares, "limite omitido vale 85")
}

func TestTeto_MemoriaLimitaAntesDoDisco(t *testing.T) {
	p := paramsBase() // disco dá 8
	p.ReservaMemoriaMB = 2048
	p.MemoriaDisponivelM = 2048 + 2*MemoriaPorParPadraoMB + 100 // 2 pares
	got := TetoDeParalelo(p)
	assert.Equal(t, 2, got.VagasMemoria)
	assert.Equal(t, 2, got.Pares)
	assert.Equal(t, "memória", got.Motivo)

	p.MemoriaDisponivelM = 2048 // só a reserva
	assert.Equal(t, 0, TetoDeParalelo(p).Pares)

	p.MemoriaPorParMB = 500 // medida real menor
	p.MemoriaDisponivelM = 2048 + 1000
	assert.Equal(t, 2, TetoDeParalelo(p).VagasMemoria)

	p.MemoriaDisponivelM = -1 // memória desconhecida não trava
	assert.Equal(t, 8, TetoDeParalelo(p).Pares)
}

func TestTeto_SemTetoFixoSobeAlemDeOito(t *testing.T) {
	p := paramsBase()
	p.Disco.Usado = 0
	p.TamanhoWorktree = 1 * gb
	p.MaxParalelo = 0 // sem teto fixo
	got := TetoDeParalelo(p)
	assert.Equal(t, 90, got.Pares, "só o recurso limita (100 GB menos 10 de reserva)")

	p.MaxParalelo = 12 // teto opcional escolhido
	got = TetoDeParalelo(p)
	assert.Equal(t, 12, got.Pares)
	assert.Equal(t, "builder.max_parallel_tasks", got.Motivo)
}

func TestTeto_PisoDeUmComVagaDeDiscoEMemoria(t *testing.T) {
	p := paramsBase()
	p.LarguraOnda = 1
	p.MemoriaDisponivelM = 2048 + MemoriaPorParPadraoMB
	p.ReservaMemoriaMB = 2048
	assert.Equal(t, 1, TetoDeParalelo(p).Pares)
	p.LarguraOnda = 0
	assert.Equal(t, 0, TetoDeParalelo(p).Pares, "sem task elegível não há piso")
}

func TestDisco_MedicaoRealEEstadoPorMaquina(t *testing.T) {
	dir := t.TempDir()
	u, err := MedirDisco(filepath.Join(dir, "nao", "existe"))
	require.NoError(t, err, "mede o primeiro pai que existe")
	assert.Greater(t, u.Total, uint64(0))
	assert.LessOrEqual(t, u.Usado, u.Total)

	// 1 KiB de arquivo fora de .git e de .forge/runtime
	testutil.WriteFile(t, dir, "a.txt", string(make([]byte, 1024)))
	testutil.WriteFile(t, dir, ".git/objeto", string(make([]byte, 4096)))
	testutil.WriteFile(t, dir, ".forge/runtime/x", string(make([]byte, 4096)))
	tam, err := MedirTamanhoDir(dir)
	require.NoError(t, err)
	assert.Equal(t, uint64(1024), tam)

	got, err := TamanhoDoWorktree(dir, true)
	require.NoError(t, err)
	assert.Equal(t, uint64(1024), got)
	e, ok := LerEstadoDisco(dir)
	require.True(t, ok)
	assert.Equal(t, uint64(1024), e.TamanhoWorktree)

	// dry-run não grava
	dir2 := t.TempDir()
	_, err = TamanhoDoWorktree(dir2, false)
	require.NoError(t, err)
	_, ok = LerEstadoDisco(dir2)
	assert.False(t, ok)
}

func TestDisco_RemedeACadaVinteTasks(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "a.txt", string(make([]byte, 1000)))
	_, err := TamanhoDoWorktree(dir, true)
	require.NoError(t, err)
	require.NoError(t, RegistrarMemoriaDoPar(dir, 900))

	testutil.WriteFile(t, dir, "b.txt", string(make([]byte, 1000))) // o projeto cresceu
	for i := 0; i < RemedirACada-1; i++ {
		require.NoError(t, RegistrarTaskConcluida(dir))
	}
	got, _ := TamanhoDoWorktree(dir, true)
	assert.Equal(t, uint64(1000), got, "antes da vigésima task vale a medida guardada")

	require.NoError(t, RegistrarTaskConcluida(dir))
	got, _ = TamanhoDoWorktree(dir, true)
	assert.Equal(t, uint64(2000), got, "na vigésima remede")
	e, _ := LerEstadoDisco(dir)
	assert.Equal(t, 0, e.TasksDesde)
	assert.Equal(t, 900, e.MemoriaParMB, "a medida de memória sobrevive à remedição")
	assert.Equal(t, 900, MemoriaDoPar(dir))
	assert.Error(t, RegistrarMemoriaDoPar(dir, 0))
}

func TestDisco_EstadoCorrompidoEhIgnorado(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(layout.ForgeRuntime(dir), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(layout.ForgeRuntime(dir), "disco.json"), []byte("{lixo"), 0o644))
	_, ok := LerEstadoDisco(dir)
	assert.False(t, ok)
	assert.Equal(t, MemoriaPorParPadraoMB, MemoriaDoPar(dir))
}

func TestDisco_MemoriaDaMaquinaEhLegivel(t *testing.T) {
	v, err := MemoriaDisponivelMB()
	if err != nil {
		t.Skipf("sem leitura de memória neste ambiente: %v", err)
	}
	assert.Greater(t, v, 0)
}

// planoLargo cria n tasks independentes e devolve o plano e o job pai.
func planoLargo(t *testing.T, n int) (string, [][]Task, bus.Job) {
	t.Helper()
	dir := projetoPorTask(t, map[string]string{"feat-77-01-autenticacao.md": specTask("feat-77-01", "T1", "")})
	for i := 2; i <= n; i++ {
		id := "feat-77-" + twoDigits(i)
		testutil.WriteFile(t, dir, base77+"/features/"+id+"-t.md", specTask(id, "T", ""))
	}
	plano, err := planoDeTasks(dir, "77-auth-oauth", 64)
	require.NoError(t, err)
	pai := bus.Job{
		SchemaVersion: bus.SchemaVersion, ID: "job-pai", Type: bus.TypeFeatureBuild, Priority: bus.PriorityNormal, Feature: "auth-oauth",
		Title: "x", Initiative: "77-auth-oauth", SpecPath: base77 + "/criteria.md", PR: bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		Attempt: 1, MaxReworkAttempts: 3,
	}
	for _, b := range plano {
		for _, tk := range b {
			pai.TaskIDs = append(pai.TaskIDs, idCurtoDaTask(tk.ID))
		}
	}
	return dir, plano, pai
}

func twoDigits(i int) string {
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

func TestDespachar_Liberacao_QuatroVagasQuatroTasks(t *testing.T) {
	dir, plano, pai := planoLargo(t, 10)
	cfg := config.DefaultForgeConfig().Builder
	cfg.MaxParallelTasks = 0
	require.NoError(t, GravarEstadoDisco(dir, EstadoDisco{TamanhoWorktree: 5 * gb}))

	// 60 GB usados de 100, reserva de 10 GB, worktree 5 GB: 6 vagas; com 2 ativos sobram 4.
	estados := map[string]string{"feat-77-01": OndaEmAndamento, "feat-77-02": OndaEmAndamento}
	res, err := Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Estados: estados, Builder: cfg, Disco: discoFalso(60), Memoria: memFalsa(64 * 1024), Now: now})
	require.NoError(t, err)
	assert.Equal(t, 6, res.Teto.Pares)
	assert.Equal(t, 2, res.Ativos)
	require.Len(t, res.Jobs, 4, "as 4 vagas livres começam na mesma rodada")
	assert.Equal(t, "77-03", res.Jobs[0].TaskID, "na ordem do plano, sem repetir as ativas")
	assert.Len(t, builderPending(t, dir), 4)

	// idempotente: a mesma rodada não duplica
	_, err = Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Estados: estados, Builder: cfg, Disco: discoFalso(60), Memoria: memFalsa(64 * 1024), Now: now})
	require.NoError(t, err)
	assert.Len(t, builderPending(t, dir), 4)
}

func TestDespachar_SemVagaNaoComecaNada(t *testing.T) {
	dir, plano, pai := planoLargo(t, 3)
	cfg := config.DefaultForgeConfig().Builder
	require.NoError(t, GravarEstadoDisco(dir, EstadoDisco{TamanhoWorktree: 5 * gb}))

	res, err := Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Builder: cfg, Disco: discoFalso(89), Memoria: memFalsa(64 * 1024), Now: now})
	require.NoError(t, err)
	assert.Empty(t, res.Jobs)
	assert.Empty(t, builderPending(t, dir))

	// memória baixa também segura, mesmo com disco folgado
	res, err = Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Builder: cfg, Disco: discoFalso(10), Memoria: memFalsa(1000), Now: now})
	require.NoError(t, err)
	assert.Empty(t, res.Jobs)
	assert.Equal(t, "memória", res.Teto.Motivo)
}

func TestDespachar_SemTetoFixoSobeAlemDeOito(t *testing.T) {
	dir, plano, pai := planoLargo(t, 12)
	cfg := config.DefaultForgeConfig().Builder
	cfg.MaxParallelTasks = 0
	require.NoError(t, GravarEstadoDisco(dir, EstadoDisco{TamanhoWorktree: 1 * gb}))
	res, err := Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Builder: cfg, Disco: discoFalso(10), Memoria: memFalsa(64 * 1024), Now: now})
	require.NoError(t, err)
	assert.Len(t, res.Jobs, 12, "sem teto fixo, todas as 12 elegíveis cabem")

	// com teto opcional, ele vale
	dir2, plano2, pai2 := planoLargo(t, 12)
	cfg.MaxParallelTasks = 5
	require.NoError(t, GravarEstadoDisco(dir2, EstadoDisco{TamanhoWorktree: 1 * gb}))
	res, err = Despachar(EntradaDespacho{Dir: dir2, Pai: pai2, Plano: plano2, Builder: cfg, Disco: discoFalso(10), Memoria: memFalsa(64 * 1024), Now: now})
	require.NoError(t, err)
	assert.Len(t, res.Jobs, 5)
}

func TestDespachar_DependenteEsperaAPredecessoraMesclada(t *testing.T) {
	dir := projetoPorTask(t, map[string]string{
		"feat-77-01-autenticacao.md": specTask("feat-77-01", "A", ""),
		"feat-77-02-tokens.md":       specTask("feat-77-02", "B", "feat-77-01"),
	})
	plano, err := planoDeTasks(dir, "77-auth-oauth", 64)
	require.NoError(t, err)
	assert.Len(t, ElegiveisDoPlano(plano, nil), 1)
	assert.Len(t, ElegiveisDoPlano(plano, map[string]string{"feat-77-01": OndaEmAndamento}), 0, "01 em andamento não libera a 02")
	el := ElegiveisDoPlano(plano, map[string]string{"feat-77-01": TaskMesclada})
	require.Len(t, el, 1)
	assert.Equal(t, "feat-77-02", el[0].ID)
}

func TestDespachar_DiscoIndisponivelEhErro(t *testing.T) {
	dir, plano, pai := planoLargo(t, 2)
	_, err := Despachar(EntradaDespacho{Dir: dir, Pai: pai, Plano: plano, Builder: config.DefaultForgeConfig().Builder,
		Disco: func(string) (UsoDisco, error) { return UsoDisco{}, errors.New("sem acesso") }, Now: now})
	assert.Error(t, err)
}
