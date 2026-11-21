package scaffold

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func extrairBlocoNovidadesLP(t *testing.T, html string) string {
	t.Helper()
	const ini = "/* novidades:inicio */"
	const fim = "/* novidades:fim */"
	i := strings.Index(html, ini)
	require.NotEqual(t, -1, i, "marcador novidades:inicio ausente em docs/index.html")
	j := strings.Index(html[i:], fim)
	require.NotEqual(t, -1, j, "marcador novidades:fim ausente em docs/index.html")
	return html[i : i+j+len(fim)]
}

func lerJSONNovidades(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "data", "releases.json"))
	require.NoError(t, err, "docs/data/releases.json deve existir")
	var dados map[string]any
	require.NoError(t, json.Unmarshal(raw, &dados), "releases.json precisa ser JSON válido")
	return dados
}

func TestLP_NovidadesMarcadores(t *testing.T) {
	html := lerIndexLP(t)

	assert.Contains(t, html, `id="novidades"`, "a seção #novidades precisa existir")
	assert.Contains(t, html, "novidades:inicio")
	assert.Contains(t, html, "novidades:fim")
	assert.Contains(t, html, "releases/history.md", "a seção aponta para as release notes completas")
	assert.Contains(t, html, `data-nv="stable"`)
	assert.Contains(t, html, `data-nv="beta"`)

	// As abas reutilizam os mesmos marcadores de versão do topo: uma fonte só.
	assert.Regexp(t, `data-versao="latest"[^>]*>5\.0\.0<`, html)
	assert.Regexp(t, `data-versao="beta"[^>]*>5\.1\.0-beta<`, html)

	// A seção fica antes do rodapé e depois do fim do conteúdo.
	iNov := strings.Index(html, `id="novidades"`)
	iFooter := strings.Index(html, "<footer>")
	require.NotEqual(t, -1, iNov)
	require.NotEqual(t, -1, iFooter)
	assert.Less(t, iNov, iFooter, "a seção de novidades vem antes do rodapé")
}

func TestLP_NovidadesJsonVersionado(t *testing.T) {
	dados := lerJSONNovidades(t)

	versoes, ok := dados["versoes"].([]any)
	require.True(t, ok, "releases.json precisa ter o campo versoes")
	require.NotEmpty(t, versoes)

	canais := map[string]bool{}
	for _, v := range versoes {
		item, ok := v.(map[string]any)
		require.True(t, ok)
		canal, _ := item["canal"].(string)
		canais[canal] = true
		rotulo, _ := item["rotulo"].(string)
		assert.NotEmpty(t, rotulo, "toda versão do JSON precisa de rótulo")
		itens, ok := item["itens"].([]any)
		require.True(t, ok, "versão %s sem itens", rotulo)
		assert.NotEmpty(t, itens, "versão %s sem novidades", rotulo)
	}
	assert.True(t, canais["stable"], "o JSON precisa da versão estável")
	assert.True(t, canais["beta"], "o JSON precisa da versão beta")
}

func TestLP_NovidadesFuncaoPura(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node não encontrado no PATH")
	}

	bloco := extrairBlocoNovidadesLP(t, lerIndexLP(t))

	// O bloco é só lógica pura aqui: sem DOM e sem rede, para poder ser testado.
	assert.NotContains(t, bloco, "innerHTML")
	assert.NotContains(t, bloco, "setInterval")
	assert.NotContains(t, bloco, "localStorage")
	assert.Contains(t, bloco, "AbortController")
	assert.Contains(t, bloco, "credentials: 'omit'")
	assert.Contains(t, bloco, "referrerPolicy: 'no-referrer'")

	// Recorta só as funções puras (até a primeira que toca o DOM).
	i := strings.Index(bloco, "function nvAtivarAbas")
	require.NotEqual(t, -1, i, "function nvAtivarAbas ausente")
	puro := bloco[:i]

	script := puro + `
const assert = require('assert');

// normalização aceita o válido
const ok = nvNormalizar([{ canal: 'stable', rotulo: '5.0.0', itens: [{ titulo: 'A', texto: 'b' }] }]);
assert.strictEqual(ok.length, 1);
assert.strictEqual(ok[0].rotulo, '5.0.0');
assert.strictEqual(ok[0].itens[0].titulo, 'A');

// descarta marcação e esquema executável
assert.deepStrictEqual(nvNormalizar([{ canal: 'stable', itens: [{ titulo: '<img src=x onerror=1>', texto: 'ok' }] }]), []);
assert.deepStrictEqual(nvNormalizar([{ canal: 'stable', rotulo: 'javascript:alert(1)', itens: [{ titulo: 'a' }] }]), []);
assert.deepStrictEqual(nvNormalizar([{ canal: 'stable', itens: [{ texto: 'data:text/html,<b>x</b>' }] }]), []);

// descarta canal inválido, versão sem itens e entrada não-objeto
assert.deepStrictEqual(nvNormalizar([{ canal: 'royal', itens: [{ titulo: 'a' }] }]), []);
assert.deepStrictEqual(nvNormalizar([{ canal: 'stable', itens: [] }]), []);
assert.deepStrictEqual(nvNormalizar(null), []);
assert.deepStrictEqual(nvNormalizar([null, 1, 'x']), []);

// o parser de histórico entende as duas formas de marcar o canal
const hist = [
  '### Versão 5.0.0 — Estável (Estável)',
  '*   **Destaque:** texto',
  '### Versão 5.1.0 — Beta nova (Beta)',
  '*   **Outro:** mais texto',
  ''
].join('\n');
const parsed = nvParseHistorico(hist);
assert.strictEqual(parsed.length, 2);
assert.strictEqual(parsed[0].canal, 'stable');
assert.strictEqual(parsed[0].versao, '5.0.0');
assert.strictEqual(parsed[1].canal, 'beta');
assert.strictEqual(parsed[1].rotulo, '5.1.0-beta');
assert.strictEqual(parsed[1].itens[0].titulo, 'Outro');

console.log('ok');
`
	cmd := exec.Command(nodePath, "-e", script)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "falha ao validar a lógica de novidades no node: %s", string(out))
	assert.Equal(t, "ok", strings.TrimSpace(string(out)))
}

func TestNovidadesGerador(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 não encontrado no PATH")
	}
	raiz := filepath.Join("..", "..")
	cmd := exec.Command(py, "-I", "scripts/novidades_lp.py", "--check", "--quiet")
	cmd.Dir = raiz
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "docs/data/releases.json está fora de dia com o history.md: %s", string(out))
}