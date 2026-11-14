package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 13, 14, 15, 0, 0, time.UTC)

const constitution = `# Constituição
- **ARQ-01:** Fronteiras entre cmd e internal.
- **DUP-02:** Reuse Check obrigatório.
- **SEG-01:** Entrada validada na borda.
- **TST-01:** Critério executável.
`

const taskSpec = `# feat-77-01 — Autenticação

**Branch:** feat/77-auth-oauth

## Modelo do módulo
- enxuto

## O que já existe (Reuse Check — DUP-02)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| auth | internal/auth | nada | criar | novo |

## Dependências novas
- Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- ` + "`ARQ-01`, `SEG-01`, `TST-01`" + `.
- Rules: ` + "`sdd-seguranca.md`" + `. ADR: ` + "`adr-0003-estrategia-de-auth`" + `.

## Critérios de Aceitação Executáveis
` + "```bash\ngo test ./...\n```" + `
`

const criteria = `# Critérios — OAuth

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| auth | internal/auth | nada | criar | novo |
`

func evidence(feature string, at time.Time) string {
	return fmt.Sprintf(`{"provider":"clickup","id":"CU-8421","url":"https://app.clickup.com/t/8421","status":"To Do","feature":%q,"checked_at":%q}`, feature, at.UTC().Format(time.RFC3339))
}

// project monta um projeto válido no modo pipeline; mut permite quebrar uma peça por vez.
func project(t *testing.T, mut func(dir string)) string {
	t.Helper()
	dir := t.TempDir()
	cfg, _ := json.Marshal(config.DefaultForgeConfig())
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(cfg))
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", constitution)
	testutil.WriteFile(t, dir, ".agents/rules/sdd-seguranca.md", "---\ndescription: s\nglobs: [\"**/*\"]\nregras: [SEG-01, DUP-02]\n---\ncorpo\n")
	testutil.WriteFile(t, dir, ".agents/rules/sdd-arquitetura.md", "---\ndescription: a\nregras: [ARQ-01]\n---\ncorpo\n")
	testutil.WriteFile(t, dir, "sdd/spec/adr/adr-0003-estrategia-de-auth.md", "# ADR-0003\n")
	base := "sdd/discovery/77-auth-oauth"
	testutil.WriteFile(t, dir, base+"/discovery.md", "# Discovery 77 — Autenticação OAuth2\n")
	testutil.WriteFile(t, dir, base+"/criteria.md", criteria)
	testutil.WriteFile(t, dir, base+"/plan.md", "# Plano\n")
	testutil.WriteFile(t, dir, base+"/features/feat-77-01-autenticacao.md", taskSpec)
	testutil.WriteFile(t, dir, base+"/card-evidence.json", evidence("auth-oauth", now.Add(-5*time.Minute)))
	if mut != nil {
		mut(dir)
	}
	return dir
}

func builderPending(t *testing.T, dir string) []bus.Entry {
	t.Helper()
	list, err := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	require.NoError(t, err)
	return list
}

func TestDispatch_ComCardCriaJob(t *testing.T) {
	dir := project(t, nil)
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	assert.True(t, res.Created)
	assert.Equal(t, "77-auth-oauth", res.Initiative)

	jobs := builderPending(t, dir)
	require.Len(t, jobs, 1)
	j := jobs[0].Job
	assert.Equal(t, bus.TypeFeatureBuild, j.Type)
	assert.Equal(t, bus.PriorityNormal, j.Priority)
	assert.Equal(t, "auth-oauth", j.Feature)
	assert.Equal(t, "77-auth-oauth", j.Initiative)
	assert.Equal(t, []string{"77-01"}, j.TaskIDs)
	assert.Equal(t, "sdd/discovery/77-auth-oauth/criteria.md", j.SpecPath)
	assert.Equal(t, []string{"sdd/spec/adr/adr-0003-estrategia-de-auth.md"}, j.ADRRefs)
	assert.Equal(t, []string{".agents/rules/sdd-arquitetura.md", ".agents/rules/sdd-seguranca.md"}, j.RulesRefs)
	assert.Equal(t, bus.Card{Provider: "clickup", ID: "CU-8421", URL: "https://app.clickup.com/t/8421"}, j.Card)
	assert.Equal(t, "feat/77-auth-oauth", j.PR.Branch)
	assert.Equal(t, "main", j.PR.BaseBranch)
	assert.Nil(t, j.PR.Number)
	assert.Equal(t, 1, j.Attempt)
	assert.Equal(t, 3, j.MaxReworkAttempts)
	assert.Equal(t, "Discovery 77 — Autenticação OAuth2", j.Title)
	assert.Contains(t, res.CardInstruction, "CU-8421")
	assert.Contains(t, res.CardInstruction, "In Progress")

	// Segunda chamada com o mesmo instante é idempotente.
	again, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Len(t, builderPending(t, dir), 1)
}

func TestDispatch_SemCardBloqueia(t *testing.T) {
	base := "sdd/discovery/77-auth-oauth"
	cases := map[string]struct {
		mut  func(dir string)
		want string
	}{
		"sem evidência do card": {func(dir string) { _ = os.Remove(filepath.Join(dir, base, "card-evidence.json")) }, "falta a evidência do card"},
		"evidência vencida": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", evidence("auth-oauth", now.Add(-31*time.Minute)))
		}, "vencida"},
		"evidência de outra feature": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", evidence("outra-coisa", now))
		}, "outra-coisa"},
		"evidência no futuro": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", evidence("auth-oauth", now.Add(2*time.Hour)))
		}, "futuro"},
		"provider fora da lista": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", `{"provider":"trello","id":"1","url":"https://t/1","status":"x","feature":"auth-oauth","checked_at":"`+now.Format(time.RFC3339)+`"}`)
		}, "provider"},
		"card sem id": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", `{"provider":"clickup","id":"","url":"","status":"x","feature":"auth-oauth","checked_at":"`+now.Format(time.RFC3339)+`"}`)
		}, "card precisa existir"},
		"evidência com campo extra": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/card-evidence.json", `{"provider":"clickup","id":"1","url":"https://t/1","status":"x","feature":"auth-oauth","checked_at":"`+now.Format(time.RFC3339)+`","token":"abc"}`)
		}, "inválida"},
		"evidência com JSON quebrado": {func(dir string) { testutil.WriteFile(t, dir, base+"/card-evidence.json", `{`) }, "inválida"},
		"spec sem O que já existe": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/criteria.md", "# Critérios\nsem a tabela\n")
		}, "FLX-01"},
		"task sem O que já existe": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/features/feat-77-01-autenticacao.md", strings.Replace(taskSpec, "## O que já existe", "## Outra coisa", 1))
		}, "feat-77-01"},
		"task sem seção obrigatória": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/features/feat-77-01-autenticacao.md", strings.Replace(taskSpec, "## Dependências novas", "## Deps", 1))
		}, "dependências novas"},
		"ADR citada inexistente": {func(dir string) { _ = os.Remove(filepath.Join(dir, "sdd/spec/adr/adr-0003-estrategia-de-auth.md")) }, "adr-0003"},
		"ID inexistente na Constituição": {func(dir string) {
			testutil.WriteFile(t, dir, base+"/features/feat-77-01-autenticacao.md", strings.Replace(taskSpec, "`ARQ-01`", "`ARQ-99`", 1))
		}, "ARQ-99"},
		"rule citada não existe (outra rule cobre)": {func(dir string) { _ = os.Remove(filepath.Join(dir, ".agents/rules/sdd-seguranca.md")) }, "sdd-seguranca.md citada não existe"},
		"nenhuma rule cobre os IDs": {func(dir string) {
			_ = os.Remove(filepath.Join(dir, ".agents/rules/sdd-seguranca.md"))
			_ = os.Remove(filepath.Join(dir, ".agents/rules/sdd-arquitetura.md"))
		}, "nenhuma rule"},
		"sem tasks":     {func(dir string) { _ = os.Remove(filepath.Join(dir, base, "features/feat-77-01-autenticacao.md")) }, "nenhuma task"},
		"sem plan.md":   {func(dir string) { _ = os.Remove(filepath.Join(dir, base, "plan.md")) }, "plan.md"},
		"sem discovery": {func(dir string) { _ = os.Remove(filepath.Join(dir, base, "discovery.md")) }, "discovery.md"},
	}
	for name, c := range cases {
		dir := project(t, c.mut)
		res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
		require.NoError(t, err, name)
		require.True(t, res.Blocked(), "%s: o gate deveria bloquear", name)
		assert.Contains(t, strings.Join(res.Pending, " | "), c.want, name)
		assert.Empty(t, builderPending(t, dir), "%s: nada pode ser criado em builder/pending", name)
		assert.Nil(t, res.Job, name)
	}
}

func TestDispatch_ListaTodasAsPendenciasDeUmaVez(t *testing.T) {
	dir := project(t, func(dir string) {
		_ = os.Remove(filepath.Join(dir, "sdd/discovery/77-auth-oauth/card-evidence.json"))
		_ = os.Remove(filepath.Join(dir, "sdd/discovery/77-auth-oauth/plan.md"))
		_ = os.Remove(filepath.Join(dir, "sdd/spec/adr/adr-0003-estrategia-de-auth.md"))
	})
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.Pending), 3)
}

func TestDispatch_DryRunNaoGravaEFeatureAmbigua(t *testing.T) {
	dir := project(t, nil)
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, DryRun: true})
	require.NoError(t, err)
	require.False(t, res.Blocked())
	require.NotNil(t, res.Job)
	assert.Empty(t, res.JobPath)
	assert.NoDirExists(t, layout.Forge(dir)+"/runtime", "dry-run não cria diretórios de fila")

	testutil.WriteFile(t, dir, "sdd/discovery/78-auth-saml/discovery.md", "# x\n")
	res, err = Dispatch(dir, "auth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.True(t, res.Blocked())
	assert.Contains(t, res.Pending[0], "ambígua")
	assert.Contains(t, res.Pending[0], "77-auth-oauth")
	assert.Contains(t, res.Pending[0], "78-auth-saml")

	res, err = Dispatch(dir, "inexistente", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.Contains(t, res.Pending[0], "nenhuma iniciativa")
}

func TestDispatch_ModoPipelineInativoEViolacoesDeEntrada(t *testing.T) {
	dir := t.TempDir()
	_, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.ErrorIs(t, err, ErrNotPipelineMode)

	dir = project(t, nil)
	for _, f := range []string{"../x", "a/b", "A B", "", "-x", "a;b", strings.Repeat("a", 90)} {
		_, err := Dispatch(dir, f, DispatchOptions{Now: now})
		require.Error(t, err, "%q", f)
	}
	assert.Empty(t, builderPending(t, dir))

	testutil.WriteFile(t, dir, ".forge/config/forge.json", `{"cli":"gpt"}`)
	_, err = Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.Error(t, err, "forge.json inválido é erro, não modo legado")
	assert.NotErrorIs(t, err, ErrNotPipelineMode)
}

func TestDispatch_EvidenciaConfinadaAoProjeto(t *testing.T) {
	dir := project(t, nil)
	outside := t.TempDir()
	testutil.WriteFile(t, outside, "ev.json", evidence("auth-oauth", now))
	for name, p := range map[string]string{"absoluto": filepath.Join(outside, "ev.json"), "com ..": "../" + filepath.Base(outside) + "/ev.json"} {
		res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, EvidencePath: p})
		require.NoError(t, err, name)
		require.True(t, res.Blocked(), name)
		assert.Empty(t, builderPending(t, dir), name)
	}
	// Caminho relativo legítimo dentro do projeto é aceito.
	testutil.WriteFile(t, dir, "evidencias/ev.json", evidence("auth-oauth", now))
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now, EvidencePath: "evidencias/ev.json"})
	require.NoError(t, err)
	assert.False(t, res.Blocked(), "%v", res.Pending)
}

func TestDispatch_ADRNaoCitadaEmRulesNaoBloqueiaMasEntraComoReferencia(t *testing.T) {
	dir := project(t, func(dir string) {
		testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/criteria.md", criteria+"\nVer `adr-0003-estrategia-de-auth` e a futura `adr-0099-a-criar`.\n")
		// A ADR 0099 só é mencionada fora da seção de rules: não bloqueia.
		testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/features/feat-77-01-autenticacao.md", strings.Replace(taskSpec, " ADR: `adr-0003-estrategia-de-auth`.", "", 1))
	})
	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	assert.Equal(t, []string{"sdd/spec/adr/adr-0003-estrategia-de-auth.md"}, res.Job.ADRRefs)
}

func TestDispatch_SemCardObrigatorio(t *testing.T) {
	// Card desligado: nenhuma evidência é exigida e o job nasce sem card.
	dir := project(t, func(dir string) {
		cfg := config.DefaultForgeConfig()
		cfg.PMMCP.RequireExistingCard = false
		data, _ := json.Marshal(cfg)
		testutil.WriteFile(t, dir, ".forge/config/forge.json", string(data))
		_ = os.Remove(filepath.Join(dir, "sdd/discovery/77-auth-oauth/card-evidence.json"))
	})
	_, statErr := os.Stat(filepath.Join(dir, "sdd/discovery/77-auth-oauth/card-evidence.json"))
	require.True(t, os.IsNotExist(statErr), "o teste parte sem arquivo de evidência")

	res, err := Dispatch(dir, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	require.False(t, res.Blocked(), "%v", res.Pending)
	assert.True(t, res.Created)
	assert.Empty(t, res.CardInstruction, "sem card não há card para mover")
	jobs := builderPending(t, dir)
	require.Len(t, jobs, 1)
	assert.Equal(t, bus.Card{}, jobs[0].Job.Card)

	// Card ligado (padrão): sem evidência continua bloqueando.
	dir2 := project(t, func(dir string) {
		_ = os.Remove(filepath.Join(dir, "sdd/discovery/77-auth-oauth/card-evidence.json"))
	})
	res, err = Dispatch(dir2, "auth-oauth", DispatchOptions{Now: now})
	require.NoError(t, err)
	assert.True(t, res.Blocked())
	assert.Contains(t, res.Pending[0], "evidência")
	assert.Empty(t, builderPending(t, dir2))
}
