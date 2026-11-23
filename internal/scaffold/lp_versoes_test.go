package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lerIndexLP(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "index.html"))
	require.NoError(t, err, "docs/index.html deve existir")
	return string(raw)
}

func extrairBlocoVersoesLP(t *testing.T, html string) string {
	t.Helper()
	const ini = "/* versoes:inicio */"
	const fim = "/* versoes:fim */"
	i := strings.Index(html, ini)
	require.NotEqual(t, -1, i, "marcador versoes:inicio ausente em docs/index.html")
	j := strings.Index(html[i:], fim)
	require.NotEqual(t, -1, j, "marcador versoes:fim ausente em docs/index.html")
	return html[i : i+j+len(fim)]
}

func TestLP_VersoesFallbackEstatico(t *testing.T) {
	html := lerIndexLP(t)

	assert.Regexp(t, `data-versao="latest"[^>]*>5\.0\.0<`, html)
	assert.Regexp(t, `data-versao="beta"[^>]*>5\.2\.3-beta<`, html)
	assert.Contains(t, html, "npm install -g @nathanramorim/forge-sdd")
	assert.Contains(t, html, "npm install -g @nathanramorim/forge-sdd@beta")
	assert.Contains(t, html, "Para testar novidades. A beta não é a versão recomendada para produção.")
	assert.Contains(t, html, "As versões vêm do registro público do npm; se ele não responder, mostramos as últimas conhecidas.")
	assert.Contains(t, html, "Agora (v5.0) · 3 Sessões em Pipeline")
}

func TestLP_VersoesChamadaSegura(t *testing.T) {
	html := lerIndexLP(t)
	bloco := extrairBlocoVersoesLP(t, html)

	assert.Contains(t, bloco, "credentials: 'omit'")
	assert.Contains(t, bloco, "referrerPolicy: 'no-referrer'")
	assert.Contains(t, bloco, "AbortController")
	assert.NotContains(t, bloco, "innerHTML")
	assert.NotContains(t, bloco, "setInterval")
	assert.NotContains(t, bloco, "localStorage")
	assert.NotContains(t, html, "/-/package/")
}

func TestLP_VersoesFuncaoPura(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node não encontrado no PATH")
	}

	bloco := extrairBlocoVersoesLP(t, lerIndexLP(t))
	script := bloco + `
const assert = require('assert');
const fb = { latest: '5.0.0', beta: '5.1.0-beta' };

assert.strictEqual(versaoValida('5.0.0', 'latest'), true);
assert.strictEqual(versaoValida('5.1.0-beta', 'beta'), true);
assert.strictEqual(versaoValida('5.1.0-beta.2', 'beta'), true);

assert.strictEqual(versaoValida('5.0.0<img src=x onerror=1>', 'latest'), false);
assert.strictEqual(versaoValida('javascript:alert(1)', 'latest'), false);
assert.strictEqual(versaoValida('5.0', 'latest'), false);
assert.strictEqual(versaoValida('5.1.0-beta', 'latest'), false);
assert.strictEqual(versaoValida('5.0.0', 'beta'), false);
assert.strictEqual(versaoValida(null, 'latest'), false);
assert.strictEqual(versaoValida({}, 'beta'), false);

assert.deepStrictEqual(
  escolherVersoes({ version: '5.2.0' }, { version: '5.3.0-beta.1' }, fb),
  { latest: '5.2.0', beta: '5.3.0-beta.1' }
);
assert.deepStrictEqual(
  escolherVersoes({ version: '5.0.0<img src=x onerror=1>' }, { version: 'javascript:alert(1)' }, fb),
  fb
);
assert.deepStrictEqual(
  escolherVersoes({ version: '5.1.0-beta' }, { version: '5.0.0' }, fb),
  fb
);
assert.deepStrictEqual(
  escolherVersoes(null, {}, fb),
  fb
);
assert.deepStrictEqual(
  escolherVersoes({ version: '5.1.0' }, { version: 'invalida' }, fb),
  { latest: '5.1.0', beta: '5.1.0-beta' }
);
console.log('ok');
`
	cmd := exec.Command(nodePath, "-e", script)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "falha ao validar função pura no node: %s", string(out))
	assert.Equal(t, "ok", strings.TrimSpace(string(out)))
}
