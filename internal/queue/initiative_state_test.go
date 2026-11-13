package queue

import (
	"os"
	"os/exec"
	"path/filepath"
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

func writeSpec(t *testing.T, dir, rel string) {
	t.Helper()
	spec := "# Spec\n\n## O que já existe\n\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| x | y | z | reusar | w |\n"
	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, rel), []byte(spec), 0o644))
}

func TestQueueFilePath_IniciativaAtivaUsaPastaDaIniciativa(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "feat/70-alfa")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", "70-alfa"), 0o755))

	assert.Equal(t, filepath.Join(dir, "sdd", "discovery", "70-alfa", "queue.json"), QueueFilePath(dir))
}

func TestQueueFilePath_SemIniciativaUsaFilaLegada(t *testing.T) {
	// Sem git: cai na fila global legada.
	assert.Equal(t, filepath.Join("/p", "sdd", ".queue", "queue.json"), QueueFilePath("/p"))

	// Branch de iniciativa, mas sem a pasta em sdd/discovery: também legada.
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "feat/71-sem-pasta")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	assert.Equal(t, filepath.Join(dir, "sdd", ".queue", "queue.json"), QueueFilePath(dir))
}

func TestEnqueue_IniciativaNovaNaoHerdaItensDeOutra(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "init")

	for _, br := range []struct{ branch, id string }{{"feat/70-alfa", "70-alfa"}, {"feat/71-beta", "71-beta"}} {
		git(t, dir, "checkout", "-q", "-B", br.branch, "main")
		spec := filepath.Join("sdd", "discovery", br.id, "features", "feat-"+br.id[:2]+"-01.md")
		writeSpec(t, dir, spec)
		require.NoError(t, Enqueue(dir, "feat-"+br.id[:2], br.branch, br.id[:2]+"-01", "t", spec, nil))
		q, err := Load(dir)
		require.NoError(t, err)
		require.Len(t, q.Items, 1, "a fila de %s deve ter apenas o próprio item", br.id)
		assert.Equal(t, br.id[:2]+"-01", q.Items[0].TaskID)
		git(t, dir, "add", "-A")
		git(t, dir, "commit", "-q", "-m", "iniciativa "+br.id)
	}
}

// Duas branches paralelas criam iniciativas, enfileiram tasks e gravam status.md;
// o merge das duas termina sem conflito (feat 69-03, TST-04).
func TestMergeDeDuasIniciativasParalelasSemConflito(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("sdd/.runtime/\n"), 0o644))
	git(t, dir, "add", ".gitignore")
	git(t, dir, "commit", "-q", "-m", "init")

	for _, br := range []struct{ branch, id string }{{"feat/70-alfa", "70-alfa"}, {"feat/71-beta", "71-beta"}} {
		git(t, dir, "checkout", "-q", "-B", br.branch, "main")
		spec := filepath.Join("sdd", "discovery", br.id, "features", "feat-"+br.id[:2]+"-01.md")
		writeSpec(t, dir, spec)
		require.NoError(t, Enqueue(dir, "feat-"+br.id[:2], br.branch, br.id[:2]+"-01", "t", spec, nil))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", "discovery", br.id, "status.md"), []byte("# Status "+br.id+"\n"), 0o644))
		require.NoError(t, RecordActivity(dir, StationBuilder, "run_command", "go test", "x", "running"))
		git(t, dir, "add", "-A")
		git(t, dir, "commit", "-q", "-m", "iniciativa "+br.id)
	}

	git(t, dir, "checkout", "-q", "-B", "integracao", "feat/70-alfa")
	out := git(t, dir, "merge", "--no-edit", "feat/71-beta")
	assert.NotContains(t, out, "CONFLICT")

	for _, id := range []string{"70-alfa", "71-beta"} {
		for _, f := range []string{"queue.json", "status.md"} {
			_, err := os.Stat(filepath.Join(dir, "sdd", "discovery", id, f))
			assert.NoError(t, err, "%s/%s deve existir após o merge", id, f)
		}
	}

	// Estado de execução fica fora do git quando .runtime/ está ignorado.
	tracked := git(t, dir, "ls-files", "sdd/.runtime")
	assert.Empty(t, tracked)
}
