package queue

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const specComReuse = "# Spec\n\n## O que já existe\n\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| x | y | z | reusar | w |\n"

func writeSpec(t *testing.T, dir, rel string) {
	t.Helper()
	testutil.WriteFile(t, dir, rel, specComReuse)
}

func commitAll(t *testing.T, dir, msg string) {
	t.Helper()
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", msg)
}

func TestQueueFilePath_IniciativaAtivaUsaPastaDaIniciativa(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "feat/70-alfa")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "discovery", "70-alfa"), 0o755))

	assert.Equal(t, filepath.Join(dir, "sdd", "discovery", "70-alfa", "queue.json"), QueueFilePath(dir))
}

func TestQueueFilePath_SemIniciativaUsaFilaLegada(t *testing.T) {
	// Sem git: cai na fila global legada.
	assert.Equal(t, filepath.Join("/p", "sdd", ".queue", "queue.json"), QueueFilePath("/p"))

	// Branch de iniciativa, mas sem a pasta em sdd/discovery: também legada.
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "feat/71-sem-pasta")
	assert.Equal(t, filepath.Join(dir, "sdd", ".queue", "queue.json"), QueueFilePath(dir))
}

// A beta nasce da alfa, que já tem fila: se a fila fosse global, a beta herdaria o item da alfa.
func TestEnqueue_IniciativaNovaNaoHerdaItensDeOutra(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")

	testutil.Git(t, dir, "checkout", "-q", "-b", "feat/70-alfa")
	specAlfa := filepath.Join("sdd", "discovery", "70-alfa", "features", "feat-70-01.md")
	writeSpec(t, dir, specAlfa)
	require.NoError(t, Enqueue(dir, "feat-70", "feat/70-alfa", "70-01", "t", specAlfa, nil))
	commitAll(t, dir, "alfa")

	testutil.Git(t, dir, "checkout", "-q", "-b", "feat/71-beta", "feat/70-alfa")
	specBeta := filepath.Join("sdd", "discovery", "71-beta", "features", "feat-71-01.md")
	writeSpec(t, dir, specBeta)
	require.NoError(t, Enqueue(dir, "feat-71", "feat/71-beta", "71-01", "t", specBeta, nil))

	q, err := Load(dir)
	require.NoError(t, err)
	require.Len(t, q.Items, 1, "a fila da beta deve ter apenas o próprio item")
	assert.Equal(t, "71-01", q.Items[0].TaskID)

	// A fila da alfa continua intacta no próprio arquivo.
	alfa, err := os.ReadFile(filepath.Join(dir, "sdd", "discovery", "70-alfa", "queue.json"))
	require.NoError(t, err)
	assert.Contains(t, string(alfa), "70-01")
	assert.NotContains(t, string(alfa), "71-01")
}

func TestEnqueue_RecusaBranchDeOutraIniciativa(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "feat/70-alfa")
	spec := filepath.Join("sdd", "discovery", "70-alfa", "features", "feat-70-01.md")
	writeSpec(t, dir, spec)

	err := Enqueue(dir, "feat-71", "feat/71-beta", "71-01", "t", spec, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "71-beta")
	assert.NoFileExists(t, filepath.Join(dir, "sdd", "discovery", "70-alfa", "queue.json"), "nada deve ser gravado na fila errada")
}

// Duas branches paralelas criam iniciativas, enfileiram tasks e gravam status.md;
// o merge das duas termina sem conflito (feat 69-03, TST-04).
func TestMergeDeDuasIniciativasParalelasSemConflito(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")
	testutil.WriteFile(t, dir, ".gitignore", "sdd/.runtime/\n")
	commitAll(t, dir, "gitignore")

	for _, br := range []struct{ branch, id string }{{"feat/70-alfa", "70-alfa"}, {"feat/71-beta", "71-beta"}} {
		testutil.Git(t, dir, "checkout", "-q", "-B", br.branch, "main")
		spec := filepath.Join("sdd", "discovery", br.id, "features", "feat-"+br.id[:2]+"-01.md")
		writeSpec(t, dir, spec)
		require.NoError(t, Enqueue(dir, "feat-"+br.id[:2], br.branch, br.id[:2]+"-01", "t", spec, nil))
		testutil.WriteFile(t, dir, "sdd/discovery/"+br.id+"/status.md", "# Status "+br.id+"\n")
		require.NoError(t, RecordActivity(dir, StationBuilder, "run_command", "go test", "x", "running"))
		commitAll(t, dir, "iniciativa "+br.id)
	}

	testutil.Git(t, dir, "checkout", "-q", "-B", "integracao", "feat/70-alfa")
	out := testutil.Git(t, dir, "merge", "--no-edit", "feat/71-beta")
	assert.NotContains(t, out, "CONFLICT")

	for _, id := range []string{"70-alfa", "71-beta"} {
		for _, f := range []string{"queue.json", "status.md"} {
			assert.FileExists(t, filepath.Join(dir, "sdd", "discovery", id, f), "%s/%s deve existir após o merge", id, f)
		}
	}
	assert.Empty(t, testutil.Git(t, dir, "ls-files", "sdd/.runtime"), "estado de execução fica fora do git")
}

// Controle negativo: com a fila global legada, as mesmas duas frentes conflitam.
// Prova que o teste acima detecta o problema que a feat 69-03 resolve.
func TestControle_FilaGlobalLegadaConflitaEntreBranchesParalelas(t *testing.T) {
	dir := t.TempDir()
	testutil.InitRepo(t, dir, "main")

	for _, br := range []string{"feat/70-alfa", "feat/71-beta"} {
		testutil.Git(t, dir, "checkout", "-q", "-B", br, "main")
		// Sem pasta de iniciativa em sdd/discovery: a fila cai no arquivo global sdd/.queue/queue.json.
		spec := filepath.Join("sdd", "features", "feat-"+br[5:7]+"-01.md")
		writeSpec(t, dir, spec)
		require.NoError(t, Enqueue(dir, "feat-"+br[5:7], br, br[5:7]+"-01", "t", spec, nil))
		require.FileExists(t, filepath.Join(dir, "sdd", ".queue", "queue.json"))
		commitAll(t, dir, br)
	}

	testutil.Git(t, dir, "checkout", "-q", "-B", "integracao", "feat/70-alfa")
	full := append([]string{"-C", dir, "-c", "user.email=t@t", "-c", "user.name=t"}, "merge", "--no-edit", "feat/71-beta")
	out, _ := execGit(full...)
	assert.Contains(t, out, "CONFLICT", "a fila global legada deve conflitar")
}

// execGit executa o git sem falhar o teste: usado quando o conflito é o resultado esperado.
func execGit(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	return string(out), err
}
