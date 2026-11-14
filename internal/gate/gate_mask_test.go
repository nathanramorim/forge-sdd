package gate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskSecrets(t *testing.T) {
	sk := "sk-" + strings.Repeat("a", 24)
	gh := "ghp_" + strings.Repeat("b", 24)
	aws := "AKIA" + strings.Repeat("C", 16)
	for in, leaked := range map[string]string{
		"export K=" + sk:                      sk,
		"git push https://" + gh + "@x/y":     gh,
		"id " + aws:                           aws,
		"curl -H 'Authorization: Bearer xyz'": "xyz",
		"PASSWORD=hunter2 ./run":              "hunter2",
		"api_key: segredo123":                 "segredo123",
	} {
		out := MaskSecrets(in)
		assert.NotContains(t, out, leaked, in)
		assert.Contains(t, out, "***", in)
	}
	assert.Equal(t, "npm test -- --watch=false", MaskSecrets("npm test -- --watch=false"), "texto comum não muda")
}
