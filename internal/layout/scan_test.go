package layout

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.email=t@t", "-c", "user.name=t"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
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
	assert.Equal(t, "AB12", w)

	for _, bad := range []string{"", "../x", "a b", strings.Repeat("a", 40), "x;y"} {
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
