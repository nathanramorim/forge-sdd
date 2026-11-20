package monitor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var evalNow = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

const goodCriteria = "set -o pipefail\ngo test -list 'TestXNegativo' ./texto | grep -F 'TestXNegativo' >/dev/null\n# negativo: caso ruim reprova\ngo test ./texto -run 'TestXNegativo'\n"

const goodSim = `{"sessions":{"aprovado":true,"voltas_retorno":1,"tokens":100,"segundos":10,"intervencoes":1},
"chat":{"aprovado":false,"voltas_retorno":2,"tokens":80,"segundos":5,"intervencoes":0}}`

func writeRef(t *testing.T, refsDir, id, criteria, sim string) {
	t.Helper()
	d := filepath.Join(refsDir, id)
	require.NoError(t, os.MkdirAll(d, 0o755))
	spec := "# " + id + "\n\n## Critérios de Aceitação Executáveis\n```bash\n" + criteria + "```\n"
	require.NoError(t, os.WriteFile(filepath.Join(d, "spec.md"), []byte(spec), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(d, evalSimFile), []byte(sim), 0o644))
}

// evalProject cria um projeto com n referências válidas e devolve a raiz e sdd/eval/refs.
func evalProject(t *testing.T, n int) (string, string) {
	t.Helper()
	dir := t.TempDir()
	refs := EvalRefsDir(dir)
	for i := 1; i <= n; i++ {
		writeRef(t, refs, fmt.Sprintf("ref-%02d", i), goodCriteria, goodSim)
	}
	return dir, refs
}

func TestEvalCriteriosTemNegativo(t *testing.T) {
	t.Run("conjunto real do repositório passa", func(t *testing.T) {
		refs, err := LoadEvalRefs(EvalRefsDir(filepath.Join("..", "..")))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(refs), EvalMinRefs)
		assert.LessOrEqual(t, len(refs), EvalMaxRefs)
	})
	t.Run("referência sem linha de negativo reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-01", "go test ./texto -run 'TestXNegativo'\n", goodSim)
		_, err := LoadEvalRefs(refs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ref-01")
		assert.Contains(t, err.Error(), "# negativo:")
	})
	t.Run("negativo que roda teste não listado reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-02", "# negativo: caso ruim\ngo test ./texto -run 'TestInexistente'\n", goodSim)
		_, err := LoadEvalRefs(refs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TestInexistente")
	})
	t.Run("simulação sem um dos modos reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-03", goodCriteria, `{"sessions":{"aprovado":true}}`)
		_, err := LoadEvalRefs(refs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `sem o modo "chat"`)
	})
	t.Run("simulação com campo desconhecido reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-04", goodCriteria, `{"sessions":{"aprovado":true,"extra":1},"chat":{}}`)
		_, err := LoadEvalRefs(refs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extra")
	})
	t.Run("conjunto fora da faixa reprova", func(t *testing.T) {
		for _, n := range []int{EvalMinRefs - 1, EvalMaxRefs + 1} {
			_, refs := evalProject(t, n)
			_, err := LoadEvalRefs(refs)
			require.Error(t, err, "n=%d", n)
			assert.Contains(t, err.Error(), fmt.Sprintf("tem %d features", n))
		}
	})
	t.Run("diretório ausente devolve erro com contexto", func(t *testing.T) {
		_, err := LoadEvalRefs(filepath.Join(t.TempDir(), "nada"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "falha ao ler o conjunto de avaliação")
	})
}

func TestEvalComparaModos(t *testing.T) {
	t.Run("grava os dois modos e imprime a tabela", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		s, path, err := RunEval(dir, EvalModeSessions, evalNow)
		require.NoError(t, err)
		assert.Equal(t, EvalResultPath(dir, EvalModeSessions), path)
		assert.Equal(t, EvalResult{Mode: "sessions", Executor: "simulador", GeneratedAt: "2026-11-20T12:00:00Z",
			Features: 10, Aprovadas: 10, TaxaAprovacao: 1, VoltasRetorno: 10, Tokens: 1000, Segundos: 100, Intervencoes: 10}, s)
		c, _, err := RunEval(dir, EvalModeChat, evalNow)
		require.NoError(t, err)
		assert.Equal(t, 0, c.Aprovadas)
		assert.Equal(t, 0.0, c.TaxaAprovacao)

		var out bytes.Buffer
		require.NoError(t, CompareEval(dir, &out))
		assert.Contains(t, out.String(), "sessions")
		assert.Regexp(t, `taxa de aprovação\s+100%\s+0%`, out.String())
		assert.Regexp(t, `tokens\s+1000\s+800`, out.String())
		assert.Regexp(t, `intervenções humanas\s+10\s+0`, out.String())
	})
	t.Run("comparar sem um dos modos falha com orientação", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		_, _, err := RunEval(dir, EvalModeSessions, evalNow)
		require.NoError(t, err)
		err = CompareEval(dir, &bytes.Buffer{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "forge eval run --mode chat")
	})
	t.Run("arquivo com modo trocado é rejeitado", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		_, path, err := RunEval(dir, EvalModeSessions, evalNow)
		require.NoError(t, err)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(EvalResultPath(dir, EvalModeChat), data, 0o644))
		_, err = ReadEvalResult(dir, EvalModeChat)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `esperado "chat"`)
	})
	t.Run("modo inválido é rejeitado", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		_, _, err := RunEval(dir, "terminal", evalNow)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "modo de avaliação inválido")
		_, err = SimulateEval(nil, "", evalNow)
		require.Error(t, err)
	})
	t.Run("conjunto inválido não grava resultado", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs-1)
		_, _, err := RunEval(dir, EvalModeChat, evalNow)
		require.Error(t, err)
		_, statErr := os.Stat(EvalResultPath(dir, EvalModeChat))
		assert.True(t, os.IsNotExist(statErr))
	})
}
