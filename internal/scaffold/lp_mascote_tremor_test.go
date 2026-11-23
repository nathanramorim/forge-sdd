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

const (
	marcaTremorInicio = "/* tremor:inicio */"
	marcaTremorFim    = "/* tremor:fim */"
)

func lerLPIndex(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "index.html"))
	require.NoError(t, err, "docs/index.html deve existir")
	return string(raw)
}

func blocoTremor(t *testing.T, html string) string {
	t.Helper()
	i := strings.Index(html, marcaTremorInicio)
	f := strings.Index(html, marcaTremorFim)
	require.GreaterOrEqual(t, i, 0, "marcador %s deve existir", marcaTremorInicio)
	require.Greater(t, f, i, "marcador %s deve vir depois do início", marcaTremorFim)
	return html[i+len(marcaTremorInicio) : f]
}

func TestLP_MascoteTremorSoNaEntrada(t *testing.T) {
	html := lerLPIndex(t)

	assert.NotContains(t, html, "Math.sin(t * 28)", "o tremor contínuo não pode existir")
	assert.Contains(t, html, "prefers-reduced-motion: reduce")

	bloco := blocoTremor(t, html)
	assert.Contains(t, bloco, "Math.sin(")
	assert.Contains(t, bloco, "TREMOR_ERRO")

	fora := strings.Replace(html, bloco, "", 1)
	for _, l := range strings.Split(fora, "\n") {
		if strings.Contains(l, "Math.sin(") && strings.Contains(l, "frequencia") {
			t.Errorf("Math.sin com frequencia fora do bloco: %s", strings.TrimSpace(l))
		}
	}

	ini := strings.Index(html, "function setState(")
	require.GreaterOrEqual(t, ini, 0, "setState deve existir")
	fim := strings.Index(html[ini:], "\n  }\n")
	require.Greater(t, fim, 0)
	assert.Contains(t, html[ini:ini+fim], "errorSince = null", "setState reinicia errorSince")

	upd := strings.Index(html, "update: (t, lookX, lookY)")
	require.GreaterOrEqual(t, upd, 0, "update do mascote deve existir")
	assert.Contains(t, html[upd:], "errorSince = t", "update fixa errorSince com o relógio do update")
}

const scriptTremorNode = `
const D = 10, A = TREMOR_ERRO.amplitude, dur = TREMOR_ERRO.duracao;
const f = (dt, red) => deslocamentoTremor(D + dt, D, red === undefined ? false : red);
const falhas = [];
const ok = (c, m) => { if (!c) falhas.push(m); };
ok(f(0) === 0, 't=desde deve dar 0');
for (const dt of [0.1, 0.2]) ok(Math.abs(f(dt)) > 0 && Math.abs(f(dt)) <= A, 'amplitude em ' + dt);
const maxJ = (a, b) => { let m = 0; for (let x = a; x < b; x += 0.001) m = Math.max(m, Math.abs(f(x))); return m; };
ok(maxJ(0, dur / 2) > maxJ(dur / 2, dur), 'envoltória deve decair');
for (const dt of [0.5, 0.6, 10]) ok(f(dt) === 0, 'depois da janela deve dar 0 em ' + dt);
for (const dt of [0, 0.1, 0.3, 5]) ok(f(dt, true) === 0, 'reduzMovimento deve dar 0');
ok(deslocamentoTremor(D + 0.1, null, false) === 0, 'desde nulo deve dar 0');
if (falhas.length) { console.error(falhas.join('\n')); process.exit(1); }
console.log('ok');
`

func TestLP_MascoteTremorComportamento(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node não encontrado no PATH")
	}
	bloco := blocoTremor(t, lerLPIndex(t))
	out, err := exec.Command(node, "-e", bloco+scriptTremorNode).CombinedOutput()
	require.NoError(t, err, "node: %s", string(out))
}
