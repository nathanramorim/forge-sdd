package monitor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var evalNow = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

const goodCriteria = "set -o pipefail\ngo test -list 'TestXNegativo' ./texto | grep -F 'TestXNegativo' >/dev/null\n# negativo: caso ruim reprova\ngo test ./texto -run 'TestXNegativo'\n"

const goodSim = `{"aprovado":true,"voltas_retorno":1,"tokens":100,"segundos":10,"intervencoes":1}`

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
	t.Run("simulação com valores por modo reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-03", goodCriteria, `{"sessions":{"aprovado":true},"chat":{"aprovado":false}}`)
		_, err := LoadEvalRefs(refs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "igual nos dois modos")
	})
	t.Run("simulação com campo desconhecido reprova o conjunto", func(t *testing.T) {
		_, refs := evalProject(t, EvalMinRefs)
		writeRef(t, refs, "ref-04", goodCriteria, `{"aprovado":true,"extra":1}`)
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
		assert.Equal(t, EvalResult{Mode: "sessions", Origem: "simulado", Executor: "simulador", GeneratedAt: "2026-11-20T12:00:00Z",
			Features: 10, Aprovadas: 10, TaxaAprovacao: 1, VoltasRetorno: 10, Tokens: 1000, Segundos: 100, Intervencoes: 10}, s)
		_, _, err = RunEval(dir, EvalModeChat, evalNow)
		require.NoError(t, err)

		var out bytes.Buffer
		require.NoError(t, CompareEval(dir, &out, true))
		lines := strings.Split(out.String(), "\n")
		assert.True(t, strings.HasPrefix(lines[0], EvalAvisoSimulado), "aviso no topo: %q", lines[0])
		assert.Regexp(t, `origem\s+simulado\s+simulado`, out.String())
		assert.Regexp(t, `taxa de aprovação\s+100%\s+100%`, out.String())
	})
	t.Run("compare recusa dados simulados sem a flag", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		for _, m := range []string{EvalModeSessions, EvalModeChat} {
			_, _, err := RunEval(dir, m, evalNow)
			require.NoError(t, err)
		}
		var out bytes.Buffer
		err := CompareEval(dir, &out, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "comparação recusada")
		assert.Empty(t, out.String())
	})
	t.Run("compare sem aviso quando os dois lados são reais", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Dir(EvalResultPath(dir, "chat")), 0o755))
		for _, m := range []string{EvalModeSessions, EvalModeChat} {
			body := fmt.Sprintf(`{"mode":%q,"origem":"real","executor":"agente","features":10}`, m)
			require.NoError(t, os.WriteFile(EvalResultPath(dir, m), []byte(body), 0o644))
		}
		var out bytes.Buffer
		require.NoError(t, CompareEval(dir, &out, false))
		assert.NotContains(t, out.String(), "SIMULADO")
	})
	t.Run("arquivo sem origem é rejeitado", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Dir(EvalResultPath(dir, "chat")), 0o755))
		require.NoError(t, os.WriteFile(EvalResultPath(dir, "chat"), []byte(`{"mode":"chat"}`), 0o644))
		_, err := ReadEvalResult(dir, EvalModeChat)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sem origem válida")
	})
	t.Run("linha do run avisa que é simulado", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		res, path, err := RunEval(dir, EvalModeChat, evalNow)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(FormatEvalRun(res, path), EvalAvisoSimulado))
		res.Origem = EvalOrigemReal
		assert.NotContains(t, FormatEvalRun(res, path), "SIMULADO")
	})
	t.Run("comparar sem um dos modos falha com orientação", func(t *testing.T) {
		dir, _ := evalProject(t, EvalMinRefs)
		_, _, err := RunEval(dir, EvalModeSessions, evalNow)
		require.NoError(t, err)
		err = CompareEval(dir, &bytes.Buffer{}, true)
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

// TestEvalSimulacaoNeutra falha se o simulador der números diferentes aos dois modos no conjunto real:
// valores simulados não podem sugerir vantagem de um desenho.
func TestEvalSimulacaoNeutra(t *testing.T) {
	refs, err := LoadEvalRefs(EvalRefsDir(filepath.Join("..", "..")))
	require.NoError(t, err)
	s, err := SimulateEval(refs, EvalModeSessions, evalNow)
	require.NoError(t, err)
	c, err := SimulateEval(refs, EvalModeChat, evalNow)
	require.NoError(t, err)
	c.Mode = s.Mode
	assert.Equal(t, s, c, "o simulador favorece um modo")
}
