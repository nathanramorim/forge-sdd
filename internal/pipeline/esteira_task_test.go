package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ini77 = "77-auth-oauth"

// jobDaTask devolve o task_build enfileirado da task no Builder.
func jobDaTask(t *testing.T, dir, task string) bus.Job {
	t.Helper()
	for _, st := range []string{layout.StatePending, layout.StateProcessing, layout.StateDone} {
		es, err := bus.List(dir, layout.RoleBuilder, st)
		require.NoError(t, err)
		for _, e := range es {
			if e.Job.TaskID == task {
				return e.Job
			}
		}
	}
	t.Fatalf("task %s sem job", task)
	return bus.Job{}
}

// rodaTaskComCommit roda o worker do Builder uma vez: o agente falso commita no worktree da task
// e grava o retorno dentro dele (como o agente real faria).
func rodaTaskComCommit(t *testing.T, dir string, arquivo func(task string) (string, string)) {
	t.Helper()
	runner := &fakeRunner{fn: func(spec RunSpec) (RunResult, error) {
		task := idDoPromptRe.FindStringSubmatch(spec.Prompt)[1]
		jobID := regexp.MustCompile(`(?m)^id: (job-\S+)$`).FindStringSubmatch(spec.Prompt)[1]
		require.NotEqual(t, dir, spec.Dir, "o Builder da task nunca roda no worktree principal")
		nome, corpo := arquivo(task)
		testutil.WriteFile(t, spec.Dir, nome, corpo)
		testutil.Git(t, spec.Dir, "add", nome)
		testutil.Git(t, spec.Dir, "commit", "-q", "-m", "task "+task)
		sha := strings.TrimSpace(testutil.Git(t, spec.Dir, "rev-parse", "HEAD"))
		testutil.WriteFile(t, spec.Dir, RetornoPathRel(jobID), fmt.Sprintf(`{"task_id":%q,"status":"pronto","commit":%q,"resumo":"feito"}`, task, sha))
		return RunResult{}, nil
	}}
	rodaBuilder(t, dir, runner)
}

func aprovaRevisao(t *testing.T, dir string) *ReviewResult {
	t.Helper()
	e, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", bus.DefaultLeaseTTL, now)
	require.NoError(t, err)
	require.NotNil(t, e)
	deps, _ := notifyDeps(nil)
	res, err := ApplyVerdict(dir, e.Job.ID, "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "ok"}, deps, now)
	require.NoError(t, err)
	return res
}

// Ciclo completo de uma task: worktree efêmero, commit, worktree apagado, revisão, squash na
// feature, branch da task apagada, estado mesclada (CA-03, CA-06).
func TestEsteira_CicloDaTaskTerminaMescladaNaFeature(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	featAntes := strings.TrimSpace(testutil.Git(t, dir, "rev-parse", "feat/77-auth-oauth"))

	rodaTaskComCommit(t, dir, func(task string) (string, string) { return "a-" + task + ".go", "package a\n" })
	wt, _ := WorktreePath(dir, ini77, "feat-77-01")
	assert.NoDirExists(t, wt, "worktree limpo é apagado depois do commit")
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/*"), "forge/77-auth-oauth-feat-77-01", "a branch da task fica até a junção")

	res := aprovaRevisao(t, dir)
	assert.Equal(t, ActionApproved, res.Action)
	require.NotNil(t, res.Juncao)
	assert.Equal(t, JuncaoIntegrada, res.Juncao.Estado)
	assert.True(t, res.Juncao.FeatPronta)

	log := testutil.Git(t, dir, "log", "--format=%s", featAntes+"..feat/77-auth-oauth")
	assert.Equal(t, 1, strings.Count(strings.TrimSpace(log), "\n")+1, "squash: um commit por task")
	assert.Contains(t, log, "feat(77-01)")
	assert.FileExists(t, filepath.Join(dir, "a-77-01.go"), "o principal na branch da feature avança por fast forward")
	assert.Empty(t, strings.TrimSpace(testutil.Git(t, dir, "branch", "--list", "forge/*")))
	est, err := EstadosDaIniciativa(dir, ini77)
	require.NoError(t, err)
	assert.Equal(t, TaskMesclada, est["feat-77-01"])
}

// Uma task que depende de outra só entra na fila quando a anterior é mesclada (RN-7).
func TestEsteira_DespachoContinuoLiberaDependente(t *testing.T) {
	dir := projetoPorTask(t, map[string]string{
		"feat-77-01-autenticacao.md": specTask("feat-77-01", "A", ""),
		"feat-77-02-tokens.md":       specTask("feat-77-02", "B", "feat-77-01"),
	})
	testutil.InitRepo(t, dir, "feat/77-auth-oauth")
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	require.Len(t, res.TaskJobs, 1)

	rodaTaskComCommit(t, dir, func(task string) (string, string) { return "a-" + task + ".go", "package a\n" })
	assert.Len(t, builderPending(t, dir), 0, "a dependente espera a junção")
	r := aprovaRevisao(t, dir)
	assert.Equal(t, 1, r.Despachados)
	assert.Equal(t, "77-02", jobDaTask(t, dir, "77-02").TaskID)
}

// Conflito na junção bloqueia a task, preserva a branch e escala (negativo, SEG-03).
func TestEsteira_JuncaoComConflitoBloqueiaEPreserva(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	rodaTaskComCommit(t, dir, func(string) (string, string) { return "mesmo.txt", "da task\n" })
	// Enquanto isso, a feature recebeu outra versão do mesmo arquivo.
	testutil.WriteFile(t, dir, "mesmo.txt", "da feature\n")
	testutil.Git(t, dir, "add", "mesmo.txt")
	testutil.Git(t, dir, "commit", "-q", "-m", "concorrente")

	res := aprovaRevisao(t, dir)
	assert.Equal(t, ActionEscalated, res.Action)
	require.NotNil(t, res.Juncao)
	assert.Equal(t, JuncaoBloqueada, res.Juncao.Estado)
	assert.Contains(t, res.Juncao.Conflitos, "mesmo.txt")
	assert.Contains(t, testutil.Git(t, dir, "branch", "--list", "forge/*"), "forge/77-auth-oauth-feat-77-01")
	got, _ := os.ReadFile(filepath.Join(dir, "mesmo.txt"))
	assert.Equal(t, "da feature\n", string(got))
	est, _ := EstadosDaIniciativa(dir, ini77)
	assert.Equal(t, OndaBloqueada, est["feat-77-01"])
}

// Queda no meio: o worktree sujo fica, a próxima execução retoma dele com o handoff compacto
// (RN-16, RN-17).
func TestRetomada_WorktreeSujoViraRetomada(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	job := jobDaTask(t, dir, "77-01")

	wt, err := PrepararWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	assert.Empty(t, wt.Retomada)
	testutil.WriteFile(t, wt.Path, "meio.go", "package x\n")

	apagado, err := EncerrarWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	assert.False(t, apagado, "worktree sujo nunca é apagado")

	de2, err := PrepararWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	assert.Equal(t, wt.Path, de2.Path)
	assert.Equal(t, []string{"meio.go"}, de2.Retomada)
	sec := SecaoRetomada(job, de2)
	assert.Contains(t, sec, "<retomada>")
	assert.Contains(t, sec, "arquivos_sujos: meio.go")
	assert.Contains(t, sec, "wip(77-01)")
}

// Retrabalho ou queda depois do commit: o worktree é recriado da branch da task (RN-18).
func TestRetomada_ReabreDaBranchDaTask(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	job := jobDaTask(t, dir, "77-01")
	wt, err := PrepararWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	testutil.WriteFile(t, wt.Path, "feito.go", "package x\n")
	testutil.Git(t, wt.Path, "add", "feito.go")
	testutil.Git(t, wt.Path, "commit", "-q", "-m", "wip(77-01): passo 1")
	apagado, err := EncerrarWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	require.True(t, apagado)

	de2, err := PrepararWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	assert.True(t, de2.Reaberto)
	assert.FileExists(t, filepath.Join(de2.Path, "feito.go"))
	assert.Contains(t, SecaoRetomada(job, de2), "ultimo_commit:")
}

// Limite de uso no meio da task: o job volta à fila e o worktree fica para retomar (RN-19).
func TestRetomada_LimiteDeUsoMantemWorktree(t *testing.T) {
	dir := projetoGitComTasks(t, 1)
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	job := jobDaTask(t, dir, "77-01")
	wt, err := PrepararWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	testutil.WriteFile(t, wt.Path, "meio.go", "package x\n")
	// O caminho de adiamento do worker sai antes de encerrar o worktree; aqui confere o contrato.
	apagado, err := EncerrarWorktreeDaTask(dir, job, "forge-builder")
	require.NoError(t, err)
	assert.False(t, apagado)
	assert.FileExists(t, filepath.Join(wt.Path, "meio.go"))
}

func TestEsteira_JanelaDoPainelSoValeRecente(t *testing.T) {
	dir := t.TempDir()
	_, ok := LerJanelaDoPainel(dir, now)
	assert.False(t, ok, "sem arquivo, janela desconhecida")

	escreve := func(at time.Time, cinco, sete float64) {
		testutil.WriteFile(t, dir, ".forge/runtime/painel/janela.json",
			fmt.Sprintf(`{"five_hour_pct":%v,"seven_day_pct":%v,"at":%q}`, cinco, sete, at.UTC().Format(time.RFC3339)))
	}
	escreve(now.Add(-2*time.Minute), 87.4, 40)
	pct, ok := LerJanelaDoPainel(dir, now)
	require.True(t, ok)
	assert.Equal(t, 87, pct, "vale a maior das janelas")

	escreve(now.Add(-20*time.Minute), 99, 99)
	_, ok = LerJanelaDoPainel(dir, now)
	assert.False(t, ok, "leitura velha não decide o teto")

	testutil.WriteFile(t, dir, ".forge/runtime/painel/janela.json", `{lixo`)
	_, ok = LerJanelaDoPainel(dir, now)
	assert.False(t, ok)
}

func TestEsteira_IdsHostisSaoRecusados(t *testing.T) {
	dir := t.TempDir()
	assert.Error(t, GravarEstadoTask(dir, "../x", "feat-1", TaskMesclada, "", "", now))
	assert.Error(t, GravarEstadoTask(dir, "77-a", "../../etc", TaskMesclada, "", "", now))
	_, err := LerPai(dir, "../x")
	assert.Error(t, err)
	_, err = JuntarTaskNaFeature(dir, bus.Job{ID: "j", TaskID: "1", Initiative: "../x"}, "w", now)
	assert.Error(t, err)
	_, err = os.Stat(filepath.Join(dir, ".forge"))
	assert.True(t, os.IsNotExist(err))
}

// forge status mostra pares ativos, vagas de disco e cada task com estado (feat-122-08, CA-10).
func TestStatus_EsteiraPorTaskMostraParesETasks(t *testing.T) {
	dir := projetoPorTask(t, map[string]string{
		"feat-77-01-autenticacao.md": specTask("feat-77-01", "A", ""),
		"feat-77-02-tokens.md":       specTask("feat-77-02", "B", "feat-77-01"),
	})
	testutil.InitRepo(t, dir, "feat/77-auth-oauth")
	assert.Empty(t, FormatEsteiraPorTask(dir, now), "sem esteira por task, nada aparece")
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, Disco: discoFalso(10), Memoria: memFalsa(65536)})
	require.NoError(t, err)
	out := FormatEsteiraPorTask(dir, now)
	assert.Contains(t, out, "77-auth-oauth · 0/2 mescladas · pares ativos 1")
	assert.Contains(t, out, "feat-77-01     em andamento")
	assert.Contains(t, out, "feat-77-02     pendente")
}
