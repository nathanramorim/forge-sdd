package scripts

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const headFalso = "abcdef1234567890abcdef1234567890abcdef12"

// rodar executa o script com python3 -I, sem shell e sem rede; devolve saída e código.
func rodar(t *testing.T, args ...string) (string, int) {
	t.Helper()
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 ausente: teste do script não roda (informar no PR)")
	}
	dir, _ := os.Getwd()
	cmd := exec.Command(py, append([]string{"-I", filepath.Join(dir, "validar_entrega.py")}, args...)...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err = cmd.Run()
	rc := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		rc = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("falha ao executar: %v", err)
	}
	return out.String(), rc
}

func fixture(nome string) string {
	d, _ := os.Getwd()
	return filepath.Join(d, "tests", "fixtures", nome)
}

func veredito(t *testing.T, conteudo string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "veredito.json")
	if err := os.WriteFile(p, []byte(conteudo), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func vJSON(sha, ponto, ver string) string {
	return `{"veredito":"` + ver + `","achados":[],"resumo":"ok","head_sha":"` + sha + `","ponto":"` + ponto + `","modo":"local","feature":"feat-900-01"}`
}

func TestValidarEntregaSpecValidaPassa(t *testing.T) {
	out, rc := rodar(t, "--json", "--sem-llm", "--fixture", fixture("valido"))
	if rc != 0 {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
	var r struct {
		OK         bool     `json:"ok"`
		Julgamento []string `json:"julgamento"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil || !r.OK || r.Julgamento == nil {
		t.Fatalf("json inesperado (%v): %s", err, out)
	}
}

func TestValidarEntregaRuleInexistenteRecusada(t *testing.T) {
	out, rc := rodar(t, "--json", "--sem-llm", "--fixture", fixture("rule_inexistente"))
	if rc != 1 || !strings.Contains(out, "ZZZ-99") || !strings.Contains(out, `"ok": false`) {
		t.Fatalf("esperava rc=1 citando ZZZ-99; rc=%d out=%s", rc, out)
	}
}

func TestValidarEntregaVereditoValidoConfereComHEAD(t *testing.T) {
	v := veredito(t, vJSON(headFalso, "entrega", "conforme"))
	out, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
	if rc != 0 {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
}

func TestValidarEntregaVereditoComSHADiferenteDoHEADRecusado(t *testing.T) {
	v := veredito(t, vJSON("1111111", "entrega", "conforme"))
	out, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
	if rc != 1 || !strings.Contains(out, "head_sha") {
		t.Fatalf("esperava rc=1 por SHA divergente; rc=%d out=%s", rc, out)
	}
}

func TestValidarEntregaVereditoViolacaoNaoPassa(t *testing.T) {
	v := veredito(t, vJSON(headFalso, "entrega", "violacao"))
	_, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
	if rc != 1 {
		t.Fatalf("rc=%d", rc)
	}
}

func TestValidarEntregaPontoDoVereditoDiferenteRecusado(t *testing.T) {
	v := veredito(t, vJSON(headFalso, "spec", "conforme"))
	_, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
	if rc != 1 {
		t.Fatalf("rc=%d", rc)
	}
}

func TestValidarEntregaVereditoMalformadoSaiCom2(t *testing.T) {
	for nome, conteudo := range map[string]string{
		"nao-json":    "{isto nao e json",
		"sem-sha":     `{"veredito":"conforme","ponto":"entrega"}`,
		"sha-nao-hex": vJSON("zzzzzzz; rm -rf /", "entrega", "conforme"),
		"sha-curto":   vJSON("abc12", "entrega", "conforme"),
		"raiz-lista":  `[1,2]`,
	} {
		v := veredito(t, conteudo)
		out, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
		if rc != 2 {
			t.Fatalf("%s: esperava rc=2; rc=%d out=%s", nome, rc, out)
		}
	}
}

func TestValidarEntregaRecusaCaminhoComPontoPonto(t *testing.T) {
	_, rc := rodar(t, "--json", "--sem-llm", "--fixture", fixture("valido")+"/../valido")
	if rc != 2 {
		t.Fatalf("fixture com ..: rc=%d", rc)
	}
	_, rc = rodar(t, "--json", "--sem-llm", "--fixture", fixture("valido"), "--veredito", "../veredito.json")
	if rc != 2 {
		t.Fatalf("veredito com ..: rc=%d", rc)
	}
}

func TestValidarEntregaHEADInvalidoSaiCom2(t *testing.T) {
	_, rc := rodar(t, "--json", "--sem-llm", "--fixture", fixture("valido"), "--head", "$(id)")
	if rc != 2 {
		t.Fatalf("rc=%d", rc)
	}
}

func TestValidarEntregaSaidaSemCaracteresDeControle(t *testing.T) {
	v := veredito(t, `{"veredito":"conforme\u001b[31m","head_sha":"`+headFalso+`","ponto":"entrega"}`)
	out, rc := rodar(t, "--json", "--sem-llm", "--ponto", "entrega", "--fixture", fixture("valido"), "--veredito", v, "--head", headFalso)
	if rc == 2 {
		t.Fatalf("rc=2 inesperado: %s", out)
	}
	if strings.ContainsRune(out, 0x1b) {
		t.Fatalf("saída com ESC: %q", out)
	}
}

func TestValidarEntregaPontoForaDoEnumSaiCom2(t *testing.T) {
	_, rc := rodar(t, "--json", "--sem-llm", "--ponto", "outro", "--fixture", fixture("valido"))
	if rc != 2 {
		t.Fatalf("rc=%d", rc)
	}
}
