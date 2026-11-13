package layout

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return testutil.Git(t, dir, args...)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	testutil.WriteFile(t, root, rel, content)
}

const sampleStatus = "# Status\n\n## Status\n```\n70-01 [x] done\n70-02 [~] building\n```\n\n## Proximo passo\nBuilder: 70-02 e depois 70-03.\n"
const sampleQueue = `{"items":[{"task_id":"70-01","status":"approved"},{"task_id":"70-02","status":"building"},{"task_id":"70-03","status":"queued"},{"task_id":"70-04","status":"queued"}]}`

// repo com main e uma branch feat/70-alfa que contém a iniciativa 70-alfa.
func repoComIniciativaEmOutraBranch(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "sdd/discovery/10-antiga/discovery.md", "x")
	writeFile(t, dir, "sdd/discovery/discovery-05-legado.md", "x")
	writeFile(t, dir, "sdd/features/feat-08-legado.md", "x")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base")
	git(t, dir, "checkout", "-q", "-b", "feat/70-alfa")
	writeFile(t, dir, "sdd/discovery/70-alfa/status.md", sampleStatus)
	writeFile(t, dir, "sdd/discovery/70-alfa/queue.json", sampleQueue)
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "70")
	git(t, dir, "checkout", "-q", "main")
	return dir
}

func TestNextID_SequencialVeMaiorIdDeOutraBranch(t *testing.T) {
	dir := repoComIniciativaEmOutraBranch(t)
	// Na árvore de trabalho (main) só existe 10; a 70 existe apenas em feat/70-alfa.
	id, err := NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "71", id)
}

func TestNextID_SequencialConsideraLegadoENomesDeBranch(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "sdd/features/feat-43a2-x.md", "x")
	writeFile(t, dir, "sdd/discovery/plan-53-y.md", "x")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base")
	git(t, dir, "branch", "fix/88-so-na-branch")

	id, err := NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "89", id)
}

func TestNextID_ForaDeRepositorioUsaArvoreDeTrabalho(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sdd/discovery/12-x/discovery.md", "x")
	id, err := NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "13", id)

	vazio, err := NextID(t.TempDir(), "", "")
	require.NoError(t, err)
	assert.Equal(t, "01", vazio)
}

func TestNextID_HashEWorkitem(t *testing.T) {
	dir := repoComIniciativaEmOutraBranch(t)
	h, err := NextID(dir, "hash", "")
	require.NoError(t, err)
	assert.Regexp(t, `^[0-9a-f]{4}$`, h)

	w, err := NextID(dir, "workitem", "AB12")
	require.NoError(t, err)
	assert.Equal(t, "ab12", w, "workitem é normalizado para minúsculas, como o slug da iniciativa")

	for _, bad := range []string{"", "../x", "a b", strings.Repeat("a", 40), "x;y", "70"} {
		_, err := NextID(dir, "workitem", bad)
		assert.Error(t, err, "%q", bad)
	}
	_, err = NextID(dir, "inexistente", "")
	assert.Error(t, err)
}

func TestListInitiatives_ArvoreDeTrabalhoENaoMescladas(t *testing.T) {
	dir := repoComIniciativaEmOutraBranch(t)

	local, err := ListInitiatives(dir, false)
	require.NoError(t, err)
	require.Len(t, local, 1)
	assert.Equal(t, "10-antiga", local[0].ID)
	assert.Equal(t, WorkTreeLabel, local[0].Branch)

	all, err := ListInitiatives(dir, true)
	require.NoError(t, err)
	require.Len(t, all, 2)
	var alfa InitiativeInfo
	for _, it := range all {
		if it.ID == "70-alfa" {
			alfa = it
		}
	}
	require.Equal(t, "70-alfa", alfa.ID)
	assert.Equal(t, "feat/70-alfa", alfa.Branch)
	assert.Contains(t, alfa.Status, "70-02 [~] building")
	assert.Equal(t, "Builder: 70-02 e depois 70-03.", alfa.Next)
	assert.Equal(t, map[string]int{"approved": 1, "building": 1, "queued": 2}, alfa.Tasks)
}

func TestListInitiatives_BranchAtualUsaArvoreDeTrabalho(t *testing.T) {
	dir := repoComIniciativaEmOutraBranch(t)
	git(t, dir, "checkout", "-q", "feat/70-alfa")
	writeFile(t, dir, "sdd/discovery/70-alfa/status.md", "# S\n\n## Status\n```\nmudou localmente\n```\n")

	all, err := ListInitiatives(dir, true)
	require.NoError(t, err)
	for _, it := range all {
		if it.ID == "70-alfa" {
			assert.True(t, it.Current)
			assert.Equal(t, WorkTreeLabel, it.Branch)
			assert.Equal(t, "mudou localmente", it.Status)
		}
	}
}

func TestRefsMaliciosasSaoRecusadas(t *testing.T) {
	for _, ref := range []string{"", "--upload-pack=x", "-x", "a b", "a:b", "a\nb", `a\b`} {
		assert.False(t, safeRef(ref), "%q", ref)
		_, err := gitShow(t.TempDir(), ref, "sdd/x")
		assert.Error(t, err, "%q", ref)
		_, err = namesInRef(t.TempDir(), ref, "sdd/")
		assert.Error(t, err, "%q", ref)
	}
	assert.True(t, safeRef("feat/69-x"))
	assert.True(t, safeRef("origin/feat/69-x"))
}

func TestListInitiatives_SemCommitsEForaDeRepositorio(t *testing.T) {
	// Fora de repositório: modo árvore de trabalho funciona; --all-branches falha com contexto.
	dir := t.TempDir()
	writeFile(t, dir, "sdd/discovery/20-x/status.md", sampleStatus)
	local, err := ListInitiatives(dir, false)
	require.NoError(t, err)
	require.Len(t, local, 1)
	_, err = ListInitiatives(dir, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "layout:")

	// Repositório sem commits: sem refs, sem erro.
	vazio := t.TempDir()
	git(t, vazio, "init", "-q", "-b", "main")
	all, err := ListInitiatives(vazio, true)
	require.NoError(t, err)
	assert.Empty(t, all)
}

func TestSummarizeStatus(t *testing.T) {
	st, next := summarizeStatus(sampleStatus)
	assert.Equal(t, "70-01 [x] done | 70-02 [~] building", st)
	assert.Equal(t, "Builder: 70-02 e depois 70-03.", next)
	st, next = summarizeStatus("")
	assert.Empty(t, st)
	assert.Empty(t, next)
	long := "## Status\n" + strings.Repeat("x", 300)
	st, _ = summarizeStatus(long)
	assert.LessOrEqual(t, len([]rune(st)), 100)
}

func TestListInitiatives_SanitizaTextoDeRefsNaoConfiaveis(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	git(t, dir, "checkout", "-q", "-b", "feat/70-evil")
	writeFile(t, dir, "sdd/discovery/70-evil/status.md", "## Status\n```\n\x1b[2J\x1b]0;pwned\x07limpo\n```\n## Proximo passo\nfaça \x1b[31mvermelho\x9b\n")
	writeFile(t, dir, "sdd/discovery/70-evil/queue.json", `{"items":[{"status":"approved"},{"status":"\u001b[31mRED\nlinha"},{"status":"UPPER"}]}`)
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "evil")
	git(t, dir, "checkout", "-q", "main")

	all, err := ListInitiatives(dir, true)
	require.NoError(t, err)
	require.Len(t, all, 1)
	it := all[0]
	for _, field := range []string{it.Status, it.Next} {
		for _, r := range field {
			assert.False(t, r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f), "caractere de controle %q vazou em %q", r, field)
		}
	}
	assert.Contains(t, it.Status, "limpo")
	assert.Equal(t, map[string]int{"approved": 1, "outro": 2}, it.Tasks, "chaves fora do padrão são agrupadas em 'outro'")
}

func TestListInitiatives_RecusaBlobGrandeDeRef(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	git(t, dir, "checkout", "-q", "-b", "feat/70-grande")
	writeFile(t, dir, "sdd/discovery/70-grande/status.md", "## Status\n"+strings.Repeat("x", maxStateFileBytes+10))
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "grande")
	git(t, dir, "checkout", "-q", "main")

	all, err := ListInitiatives(dir, true)
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Empty(t, all[0].Status, "status.md acima do limite não é lido")

	_, err = gitShow(dir, "feat/70-grande", "sdd/discovery/70-grande/status.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "limite")
}

func TestNextID_VeFixAninhadoEmOutraBranchELocal(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	writeFile(t, dir, "sdd/discovery/10-base/discovery.md", "x")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base")
	git(t, dir, "checkout", "-q", "-b", "feat/10-base")
	writeFile(t, dir, "sdd/discovery/10-base/fixes/fix-70-bug.md", "x")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "fix aninhado")
	git(t, dir, "checkout", "-q", "main")

	id, err := NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "71", id, "fix aninhado em outra branch deve contar")

	writeFile(t, dir, "sdd/discovery/10-base/features/feat-75-x.md", "x")
	id, err = NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "76", id, "feature aninhada na árvore de trabalho deve contar")
}

func TestNextID_WorkitemEmUsoEhRecusado(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	writeFile(t, dir, "sdd/discovery/1024-algo/discovery.md", "x")
	_, err := NextID(dir, "workitem", "1024")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "em uso")
	id, err := NextID(dir, "workitem", "2048")
	require.NoError(t, err)
	assert.Equal(t, "2048", id)
}

func TestErrosDoGitSaoPropagadosForaDoNotARepo(t *testing.T) {
	// Fora de repositório: árvore de trabalho vale e não há erro.
	dir := t.TempDir()
	writeFile(t, dir, "sdd/discovery/12-x/discovery.md", "x")
	id, err := NextID(dir, "sequencial", "")
	require.NoError(t, err)
	assert.Equal(t, "13", id)

	// Diretório inexistente: o git falha de verdade (não é "sem repositório") e o erro sobe.
	_, err = NextID(filepath.Join(dir, "nao-existe"), "sequencial", "")
	require.Error(t, err, "falha real do git não pode virar 'sem refs'")
	assert.NotErrorIs(t, err, errNotRepo)
}
