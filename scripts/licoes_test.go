package scripts

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func rodarLicoes(t *testing.T, args ...string) (string, int) {
	t.Helper()
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 ausente: teste do script não roda")
	}
	dir, _ := os.Getwd()
	cmd := exec.Command(py, append([]string{"-I", filepath.Join(dir, "licoes.py")}, args...)...)
	cmd.Dir = filepath.Dir(dir)
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

func TestLicoesRuleExistenteImprimeCOD02(t *testing.T) {
	out, rc := rodarLicoes(t, "--rule", "COD-02")
	if rc != 0 || !strings.Contains(out, "COD-02") {
		t.Fatalf("esperava rc=0 contendo COD-02; rc=%d out=%s", rc, out)
	}
}

func TestLicoesMigradasDUP10EFLX04(t *testing.T) {
	out, rc := rodarLicoes(t, "--rule", "DUP-10,FLX-04")
	if rc != 0 || !strings.Contains(out, "DUP-10") || !strings.Contains(out, "FLX-04") {
		t.Fatalf("esperava lições migradas DUP-10 e FLX-04; rc=%d out=%s", rc, out)
	}
}

func TestLicoesRuleInexistenteRecusada(t *testing.T) {
	out, rc := rodarLicoes(t, "--rule", "ZZZ-99")
	if rc == 0 || !strings.Contains(out, "ZZZ-99") {
		t.Fatalf("esperava falha citando ZZZ-99; rc=%d out=%s", rc, out)
	}
}

func TestLicoesRecusaPathTraversalEControle(t *testing.T) {
	_, rc := rodarLicoes(t, "--dir", "../outro", "--rule", "COD-02")
	if rc == 0 {
		t.Fatalf("esperava recusa para --dir com ..")
	}
	_, rc = rodarLicoes(t, "--rule", "../COD-02")
	if rc == 0 {
		t.Fatalf("esperava recusa para --rule com ..")
	}
}
