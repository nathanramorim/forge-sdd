package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lerArquivoRaiz(t *testing.T, partes ...string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(append([]string{"..", ".."}, partes...)...))
	require.NoError(t, err)
	return string(raw)
}

// analisarChipLP varre as tags <a> do HTML: conta <a> abertos dentro de
// <a class="brand"> e confere se algum <a href="#novidades"> contém
// data-versao="beta" (sem dependência nova, SEG-07).
func analisarChipLP(t *testing.T, doc string) (aninhados int, chipBeta bool) {
	t.Helper()
	tags := regexp.MustCompile(`(?s)<(/?)a(\s[^>]*)?>|<[^>]*data-versao="beta"[^>]*>`)
	var pilha []string
	for _, m := range tags.FindAllStringSubmatch(doc, -1) {
		switch {
		case strings.HasPrefix(m[0], "</a"):
			if len(pilha) > 0 {
				pilha = pilha[:len(pilha)-1]
			}
		case strings.HasPrefix(m[0], "<a"):
			if len(pilha) > 0 && strings.Contains(pilha[0], `class="brand"`) {
				aninhados++
			}
			pilha = append(pilha, m[0])
		default: // tag com data-versao="beta"
			for _, abertura := range pilha {
				if strings.Contains(abertura, `href="#novidades"`) {
					chipBeta = true
				}
			}
		}
	}
	return aninhados, chipBeta
}

func TestLPChipVersaoLeva(t *testing.T) {
	doc := lerIndexLP(t)
	aninhados, chipBeta := analisarChipLP(t, doc)
	assert.True(t, chipBeta, `o chip Beta deve ser <a href="#novidades"> contendo data-versao="beta"`)
	assert.Zero(t, aninhados, `nenhum <a> pode ficar dentro de <a class="brand">`)
	assert.Contains(t, doc, `data-nv-alvo="beta"`)
	assert.Regexp(t, `data-versao="beta"[^>]*>5\.2\.0-beta<`, doc)
	assert.Contains(t, doc, ".chip[href]:focus-visible")
	assert.NotRegexp(t, `<a[^>]*data-nv-alvo[^>]*aria-label`, doc)
}

func TestRegraRedacaoRelease(t *testing.T) {
	const marca = "Redação das notas: o título em negrito diz o ganho da pessoa"
	const tom = "sem marca de terceiros"
	regra := lerArquivoRaiz(t, ".agents", "rules", "sdd-fluxo-e-portao.md")
	tmpl := lerArquivoRaiz(t, "internal", "scaffold", "templates", ".agents", "rules", "sdd-fluxo-e-portao.md.tmpl")
	golden := lerArquivoRaiz(t, "internal", "scaffold", "testdata", "golden", ".agents", "rules", "sdd-fluxo-e-portao.md")
	for nome, txt := range map[string]string{"rule": regra, "template": tmpl, "golden": golden} {
		assert.Contains(t, txt, marca, nome)
		assert.Contains(t, txt, tom, nome)
	}
	trecho := func(txt string) string {
		var ls []string
		for _, l := range strings.Split(txt, "\n") {
			if strings.Contains(l, "Redação das notas:") || strings.Contains(l, "Tom com emoção") {
				ls = append(ls, strings.TrimSpace(l))
			}
		}
		return strings.Join(ls, "\n")
	}
	assert.Equal(t, trecho(regra), trecho(tmpl), "rule e template idênticos")
	assert.Equal(t, trecho(tmpl), trecho(golden), "template e golden idênticos")
}

func TestNotas520SemTecnicoNoTitulo(t *testing.T) {
	hist := lerArquivoRaiz(t, "sdd", "releases", "history.md")
	i := strings.Index(hist, "### Versão 5.2.0")
	require.NotEqual(t, -1, i)
	sec := hist[i+3:]
	if j := strings.Index(sec, "\n### "); j >= 0 {
		sec = sec[:j]
	}
	negrito := regexp.MustCompile(`\*\*(.+?)\*\*`)
	proibidos := []string{"(#", "fix-", ".go", ".md", ".py", ".sh", "--"}
	achou := 0
	for _, m := range negrito.FindAllStringSubmatch(sec, -1) {
		achou++
		for _, p := range proibidos {
			assert.NotContains(t, m[1], p, "título em negrito: %q", m[1])
		}
	}
	assert.Greater(t, achou, 5)
}
