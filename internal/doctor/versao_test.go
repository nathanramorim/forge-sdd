package doctor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompararVersao(t *testing.T) {
	casos := []struct {
		p, b string
		want VersaoStatus
	}{
		{"5.0.0", "5.0.0", VersaoOK},
		{"v5.0.0", "5.0.0", VersaoOK},
		{"4.9.0", "5.0.0", VersaoProjetoAntigo},
		{"5.0.0", "4.9.1", VersaoBinarioAntigo},
		{"5.0.0", "dev", VersaoNaoComparavel},
		{"", "5.0.0", VersaoNaoComparavel},
	}
	for _, c := range casos {
		got, _ := CompararVersao(c.p, c.b)
		assert.Equal(t, c.want, got, "%q vs %q", c.p, c.b)
	}
	_, msg := CompararVersao("5.0.0\x1b[31m\r", "5.0.0")
	assert.False(t, strings.ContainsAny(msg, "\x1b\r"))
}
