package pipeline

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goalInit = "77-auth-oauth"
const goalBase = "sdd/discovery/" + goalInit

func goalStep(dir, n, deps string) {
	body := "# feat-77-" + n + "\n\n**Depende de:** " + deps + "\n"
	_ = os.WriteFile(filepath.Join(dir, goalBase, "features", "feat-77-"+n+"-etapa.md"), []byte(body), 0o644)
}

// goalProject: 3 etapas (02 e 03 dependem da 01), Gate 1 humano feito (queue.json).
func goalProject(t *testing.T, advanced bool) string {
	t.Helper()
	dir := advancedProject(t, advanced)
	goalStep(dir, "01", "nenhuma")
	goalStep(dir, "02", "feat-77-01")
	goalStep(dir, "03", "feat-77-02")
	_ = os.Remove(filepath.Join(dir, goalBase, "features", "feat-77-01-autenticacao.md"))
	if !advanced {
		writeQueue(t, dir)
	}
	return dir
}

func builderIDs(t *testing.T, dir string) []string {
	var ids []string
	for _, e := range builderPending(t, dir) {
		ids = append(ids, e.Job.ID)
	}
	return ids
}

func TestGoal_TresEtapasNaOrdemDoDAGSemDuplicar(t *testing.T) {
	dir := goalProject(t, false)
	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.NoError(t, err)
	assert.Equal(t, []string{"feat-77-01", "feat-77-02", "feat-77-03"}, res.Steps)
	assert.Len(t, res.Created, 3)

	jobs := builderPending(t, dir)
	require.Len(t, jobs, 3)
	for i, step := range res.Steps {
		j := jobs[i].Job
		assert.Equal(t, goalStepJobID(goalInit, step), j.ID, "ordem de consumo = ordem do DAG")
		assert.Equal(t, bus.TypeFeatureBuild, j.Type)
		assert.Equal(t, []string{step}, j.TaskIDs)
		assert.Equal(t, goalInit, j.Initiative)
		assert.Equal(t, "feat/"+goalInit+"-"+step, j.PR.Branch)
		assert.True(t, strings.HasPrefix(j.SpecPath, goalBase+"/features/"+step))
	}

	again, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now.Add(time.Hour)})
	require.NoError(t, err)
	assert.Empty(t, again.Created)
	assert.Len(t, again.Existing, 3)
	assert.Len(t, builderPending(t, dir), 3, "reenviar não duplica")
}

func TestGoal_CicloRejeitadoComAListaDasEtapas(t *testing.T) {
	dir := goalProject(t, false)
	goalStep(dir, "01", "feat-77-03")
	_, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ciclo")
	assert.Contains(t, err.Error(), "feat-77-01")
	assert.Contains(t, err.Error(), "feat-77-03")
	assert.Empty(t, builderPending(t, dir))
}

func TestGoal_SemPipelineOrientaInit(t *testing.T) {
	dir := t.TempDir()
	_, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	assert.True(t, errors.Is(err, ErrNotPipelineMode))
}

func TestGoal_NormalSemAprovacaoHumanaNaoEnfileira(t *testing.T) {
	dir := goalProject(t, false)
	require.NoError(t, os.Remove(filepath.Join(dir, goalBase, "queue.json")))
	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.NoError(t, err)
	assert.Contains(t, res.Awaiting, "Gate 1")
	assert.Empty(t, builderPending(t, dir))
}

func TestGoal_AvancadoSoEnfileiraAposParecerAprovado(t *testing.T) {
	dir := goalProject(t, true)
	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.NoError(t, err)
	assert.NotEmpty(t, res.SpecJobID)
	assert.Empty(t, builderPending(t, dir), "nenhuma etapa antes do parecer")
	require.Len(t, reviewerPending(t, dir), 1)

	// Reenviar antes do parecer não duplica o spec_review.
	_, err = EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.NoError(t, err)
	require.Len(t, reviewerPending(t, dir), 1)

	e, err := bus.Claim(dir, layout.RoleReviewer, "forge-reviewer", time.Hour, now)
	require.NoError(t, err)
	deps, _ := notifyDeps(nil)
	ar, err := ApplySpecVerdict(dir, e.Job.ID, "forge-reviewer", conforme(), deps, now)
	require.NoError(t, err)
	assert.Equal(t, ActionApproved, ar.Action)
	assert.Equal(t, 3, len(builderPending(t, dir)), "um job por etapa, sem build único da iniciativa")
	for _, id := range builderIDs(t, dir) {
		assert.True(t, strings.HasPrefix(id, goalJobPrefix), id)
	}

	again, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
	require.NoError(t, err)
	assert.Empty(t, again.Created)
	assert.Len(t, builderPending(t, dir), 3)
}

func snapshot(t *testing.T, dir string) []string {
	var files []string
	require.NoError(t, filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			files = append(files, rel+info.ModTime().String())
		}
		return nil
	}))
	return files
}

func TestGoal_DryRunNaoGravaNada(t *testing.T) {
	for _, adv := range []bool{false, true} {
		dir := goalProject(t, adv)
		before := snapshot(t, dir)
		_, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now, DryRun: true})
		require.NoError(t, err)
		assert.Equal(t, before, snapshot(t, dir), "advanced=%v", adv)
	}
}

func goalEnv(have ...string) PreflightEnv {
	set := map[string]bool{}
	for _, h := range have {
		set[h] = true
	}
	return PreflightEnv{
		LookPath: func(n string) (string, error) {
			if set[n] {
				return "/bin/" + n, nil
			}
			return "", errors.New("ausente")
		},
		Exists: func(p string) bool { return set[filepath.Base(p)] },
	}
}

const goalReq = `{"comandos":[["uv","run","voice-studio","perfil","criar"],["git","status"]],
"ferramentas":["uv","python3.11"],"arquivos":["amostra-voz.wav"],
"consentimentos":[{"id":"voz-propria","texto":"a voz da amostra é minha"}]}`

func TestPreflightGoal_ListaTudoDeUmaVezENaoEnfileira(t *testing.T) {
	dir := goalProject(t, false)
	testutil.WriteFile(t, dir, goalBase+"/"+RequirementsFile, goalReq)

	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now, Env: goalEnv("uv")})
	require.NoError(t, err)
	all := strings.Join(res.Pending, "\n")
	assert.Contains(t, all, `"python3.11"`)
	assert.NotContains(t, all, `ferramenta "uv"`)
	assert.Contains(t, all, `"amostra-voz.wav"`)
	assert.Contains(t, all, `comando "uv"`)
	assert.NotContains(t, all, `comando "git"`)
	assert.Contains(t, all, "voz-propria")
	assert.Len(t, res.Pending, 4)
	assert.Empty(t, builderPending(t, dir))
}

func TestPreflightGoal_PermissaoFullCobreComandos(t *testing.T) {
	dir := goalProject(t, false)
	cfg := config.DefaultForgeConfig()
	cfg.Permissions = config.PermissionFull
	require.NoError(t, config.SaveForgeConfig(dir, cfg))
	testutil.WriteFile(t, dir, goalBase+"/"+RequirementsFile, `{"comandos":[["uv","run"]]}`)
	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now, Env: goalEnv()})
	require.NoError(t, err)
	assert.Empty(t, res.Pending)
}

func TestPreflightGoal_RequisitosInvalidosSaoRecusados(t *testing.T) {
	for name, body := range map[string]string{
		"campo desconhecido": `{"shell":"rm -rf /"}`,
		"argv vazio":         `{"comandos":[[]]}`,
		"id de consent ruim": `{"consentimentos":[{"id":"../x","texto":"t"}]}`,
		"controle no texto":  `{"consentimentos":[{"id":"ok","texto":"a\nb"}]}`,
	} {
		dir := goalProject(t, false)
		testutil.WriteFile(t, dir, goalBase+"/"+RequirementsFile, body)
		_, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now})
		assert.Error(t, err, name)
		assert.Empty(t, builderPending(t, dir), name)
	}
}

func TestConsent_SoOGravadorHumanoLiberaAEtapa(t *testing.T) {
	dir := goalProject(t, false)
	testutil.WriteFile(t, dir, goalBase+"/"+RequirementsFile, `{"consentimentos":[{"id":"voz-propria","texto":"a voz é minha"}]}`)

	res, err := EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now, Env: goalEnv()})
	require.NoError(t, err)
	require.Len(t, res.Pending, 1)
	assert.Empty(t, builderPending(t, dir), "sem consentimento gravado a etapa nem começa")

	require.Error(t, RecordConsent(dir, "../x", "t", now))
	require.NoError(t, RecordConsent(dir, "voz-propria", "a voz é minha", now))
	require.NoError(t, RecordConsent(dir, "voz-propria", "a voz é minha", now), "regravar não duplica")
	recs, err := LoadConsents(dir)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	info, err := os.Stat(consentsPath(dir))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	res, err = EnqueueGoalSteps(dir, goalInit, GoalOptions{Now: now, Env: goalEnv()})
	require.NoError(t, err)
	assert.Empty(t, res.Pending)
	assert.Len(t, builderPending(t, dir), 3)
}

func TestPendencias_ListaUnicaBloqueiaEtapaEIndependentesSeguem(t *testing.T) {
	dir := goalProject(t, false)
	testutil.WriteFile(t, dir, goalBase+"/queue.json", `{"spec_status":"reviewed","items":[
{"task_id":"feat-77-01-etapa","status":"queued"},{"task_id":"feat-77-02-etapa","status":"queued"}]}`)
	deps, _ := notifyDeps(nil)
	items := []HumanPending{{TaskID: "feat-77-01", Item: "liberar o comando uv"}}
	require.NoError(t, ReportHumanPending(dir, goalInit, items, deps, now))
	require.NoError(t, ReportHumanPending(dir, goalInit, items, deps, now), "repetir não duplica")
	require.NoError(t, ReportHumanPending(dir, goalInit, []HumanPending{{TaskID: "feat-77-02", Item: "fornecer amostra-voz.wav"}}, deps, now))

	data, err := os.ReadFile(filepath.Join(dir, goalBase, HumanPendingFile))
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(data), "liberar o comando uv"))
	assert.Contains(t, string(data), "fornecer amostra-voz.wav")

	raw, err := os.ReadFile(filepath.Join(dir, goalBase, "queue.json"))
	require.NoError(t, err)
	var q queue.Queue
	require.NoError(t, json.Unmarshal(raw, &q))
	for _, it := range q.Items {
		assert.Equal(t, queue.StatusBlocked, it.Status)
	}
	assert.Contains(t, string(raw), "liberar o comando uv")

	require.Error(t, ReportHumanPending(dir, goalInit, []HumanPending{{TaskID: "x", Item: "y"}}, deps, now))
}
