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

func TestMaskSecrets_VariaveisComPrefixoAspasJSONEUserinfo(t *testing.T) {
	skAnt := "sk-ant-api03-" + strings.Repeat("a", 30)
	skProj := "sk-proj-" + strings.Repeat("b", 30)
	pat := "github_pat_" + strings.Repeat("c", 30)
	slack := "xoxb-" + strings.Repeat("1", 12)
	jwt := "eyJ" + strings.Repeat("a", 12) + "." + strings.Repeat("b", 12) + "." + strings.Repeat("c", 12)
	for in, leaked := range map[string]string{
		"export ANTHROPIC_API_KEY=" + skAnt:          skAnt,
		"GITHUB_TOKEN=abc123def":                     "abc123def",
		"AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI":        "wJalrXUtnFEMI",
		`password="abc 123"`:                         "abc",
		`token: 'tok999'`:                            "tok999",
		`{"password": "hunter2", "ok": true}`:        "hunter2",
		"mysql --password hunter2 -h x":              "hunter2",
		"git clone https://user:senha123@host/x.git": "senha123",
		"key " + skProj:                              skProj,
		"x " + pat:                                   pat,
		"x " + slack:                                 slack,
		"Authorization: Bearer " + jwt:               jwt,
		"-----BEGIN RSA PRIVATE" + " KEY-----\nMIIEabc\n-----END RSA PRIVATE" + " KEY-----": "MIIEabc",
	} {
		out := MaskSecrets(in)
		assert.NotContains(t, out, leaked, in)
		assert.Contains(t, out, "***", in)
	}
	// Texto comum e nomes sem valor secreto não mudam.
	for _, in := range []string{"npm test -- --watch=false", "export PATH=/usr/bin", "git push origin main", "go test ./..."} {
		assert.Equal(t, in, MaskSecrets(in), in)
	}
}

func TestMaskSecrets_FalsosPositivosELacunasResiduais(t *testing.T) {
	// Palavras que terminam em "sk-" e caminhos com slugs não são chaves.
	for _, in := range []string{
		"sdd/features/feat-72-x/task-72-01-visibilidade-tmux.md",
		"risk-management-framework-review-document",
		"max_tokens=100 e tokens: 5",
	} {
		assert.Equal(t, in, MaskSecrets(in), in)
	}
	// Lacunas apontadas na reconferência.
	for in, leaked := range map[string]string{
		`password = "my secret pass"`:          "secret pass",
		`{\"password\":\"s3cr3t\"}`:            "s3cr3t",
		"Authorization: Basic dXNlcjpwYXNz":    "dXNlcjpwYXNz",
		"git clone https://u:pa/ss@host/x.git": "pa/ss",
	} {
		out := MaskSecrets(in)
		assert.NotContains(t, out, leaked, in)
		assert.Contains(t, out, "***", in)
	}
	assert.Contains(t, MaskSecrets("key sk-"+strings.Repeat("a", 24)), "***")
}
