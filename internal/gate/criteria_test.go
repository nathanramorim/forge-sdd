package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goodCriteriaSpec = "# feat\n\n## Critérios de Aceitação Executáveis\n```bash\nset -o pipefail\ngrep -o 'a' f.txt | wc -l\n# negativo: arquivo ausente reprova\ntest ! -e nao-existe.txt\n```\n- Teste negativo: arquivo ausente.\n"

func hasProblem(problems []string, sub string) bool {
	for _, p := range problems {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}

func TestCriteriaVerifyAceitaBlocoBom(t *testing.T) {
	block, ok := ExtractCriteriaBlock(goodCriteriaSpec)
	require.True(t, ok)
	assert.Empty(t, CriteriaProblems(block))
}

func TestCriteriaVerifyRejeitaGrepC(t *testing.T) {
	// Caso real: grep -c conta linhas; num SVG de uma linha só sempre dá 1.
	data, err := os.ReadFile(filepath.Join("testdata", "criterios", "grep-c-svg.md"))
	require.NoError(t, err)
	block, ok := ExtractCriteriaBlock(string(data))
	require.True(t, ok)
	assert.True(t, hasProblem(CriteriaProblems(block), "grep -c"))

	for _, line := range []string{"grep -qc x f", "grep -cE 'x' f", "grep --count x f", "grep -c x f | head -1"} {
		assert.True(t, hasProblem(CriteriaProblems("set -o pipefail\n"+line+"\n# negativo: x\nfalse || true"), "grep -c"), line)
	}
	assert.False(t, hasProblem(CriteriaProblems("set -o pipefail\ngrep -o x f | wc -l\n# negativo: x\ntrue"), "grep -c"))
	assert.False(t, hasProblem(CriteriaProblems("go test -count=1 ./...\n# negativo: x\ntrue"), "grep -c"))
}

func TestCriteriaVerifyExigeNegativo(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "criterios", "sem-negativo.md"))
	require.NoError(t, err)
	block, ok := ExtractCriteriaBlock(string(data))
	require.True(t, ok)
	assert.True(t, hasProblem(CriteriaProblems(block), "teste negativo"))

	// Marcador sem comando depois não conta.
	assert.True(t, hasProblem(CriteriaProblems("true\n# negativo: nada"), "não é seguido de comando"))
	// Marcador inline conta.
	assert.Empty(t, CriteriaProblems("test ! -e x # negativo: ausente reprova"))
}

func TestCriteriaVerifyIgnoraComoRegenerar(t *testing.T) {
	// Caso real: o extrator juntava o bloco de "Como regenerar" aos critérios.
	spec := goodCriteriaSpec + "\n## Como regenerar\n```bash\nrm -rf /tmp/x && make regen\n```\n"
	block, ok := ExtractCriteriaBlock(spec)
	require.True(t, ok)
	assert.NotContains(t, block, "make regen")
	assert.Contains(t, block, "wc -l")

	// Bloco só fora da seção de critérios não é critério.
	_, ok = ExtractCriteriaBlock("# x\n\n## Como regenerar\n```bash\nmake regen\n```\n")
	assert.False(t, ok)

	// Subseção ### dentro dos critérios continua na seção; um "## " dentro de cerca não encerra.
	spec2 := "## Critérios de Aceitação Executáveis\n### Unitários\n```bash\necho '## não é título'\ntrue # negativo: x\n```\n## Outra\n"
	block, ok = ExtractCriteriaBlock(spec2)
	require.True(t, ok)
	assert.Contains(t, block, "não é título")
	assert.Contains(t, block, "negativo")
}

func TestCriteriaVerifyRejeitaPipelineSemPipefail(t *testing.T) {
	assert.True(t, hasProblem(CriteriaProblems("go test -list X ./a | grep -q X\n# negativo: x\ntrue"), "pipefail"))
	assert.False(t, hasProblem(CriteriaProblems("go vet ./... || exit 1\n# negativo: x\ntrue"), "pipefail"))
}

func TestCriteriaVerifyRejeitaSaidaVaziaSobBashE(t *testing.T) {
	// Caso real: critério que só passa quando imprime algo quebra sob bash -e com saída vazia.
	block := "set -o pipefail\nprintf '' | grep -v x\n# negativo: x\ntrue"
	assert.True(t, hasProblem(CriteriaProblems(block), "grep -v"))
	err := RunCriteriaBlock(t.TempDir(), block)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bash -e")

	// '!' sem '|| exit' nunca falha sob bash -e.
	assert.True(t, hasProblem(CriteriaProblems("! true\n# negativo: x\ntrue"), "'!'"))
	assert.False(t, hasProblem(CriteriaProblems("! false || exit 1\n# negativo: x\ntrue"), "'!'"))
}

func TestCriteriaVerifyRejeitaSintaxe(t *testing.T) {
	assert.True(t, hasProblem(CriteriaProblems("if true; then\n# negativo: x\ntrue"), "bash -n"))
}

func TestRunChecksP25(t *testing.T) {
	dir := setupGateDir(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f.txt"), []byte("aaa"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "boa.md"), []byte(goodCriteriaSpec), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ruim.md"), []byte("## Critérios de Aceitação Executáveis\n```bash\nfalse # negativo: x\n```\n"), 0o644))

	results, err := RunChecks(dir, Options{SpecPath: "boa.md"})
	require.NoError(t, err)
	_, found := findCheck(results, "P-25")
	assert.False(t, found, "P-25 só roda com Criteria")

	results, err = RunChecks(dir, Options{SpecPath: "boa.md", Criteria: true, CriteriaExec: true})
	require.NoError(t, err)
	p25, found := findCheck(results, "P-25")
	require.True(t, found)
	assert.True(t, p25.Passed, p25.Evidence)

	results, err = RunChecks(dir, Options{SpecPath: "ruim.md", Criteria: true, CriteriaExec: true})
	require.NoError(t, err)
	p25, _ = findCheck(results, "P-25")
	assert.False(t, p25.Passed)
	assert.Contains(t, p25.Evidence, "bash -e")

	results, err = RunChecks(dir, Options{SpecPath: "ausente.md", Criteria: true})
	require.NoError(t, err)
	p25, _ = findCheck(results, "P-25")
	assert.False(t, p25.Passed)
}
