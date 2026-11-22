package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoDeTeste cria um repositório git mínimo com ci-local.sh copiado e um módulo Go.
func repoDeTeste(t *testing.T, testeOk bool) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go ausente")
	}
	raiz := t.TempDir()
	dir, _ := os.Getwd()
	sh, err := os.ReadFile(filepath.Join(dir, "ci-local.sh"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(raiz, "scripts"), 0o755)
	if err := os.WriteFile(filepath.Join(raiz, "scripts", "ci-local.sh"), sh, 0o755); err != nil {
		t.Fatal(err)
	}
	asserta := "t.Log(1)"
	if !testeOk {
		asserta = "t.Fatal(\"quebrado\")"
	}
	_ = os.WriteFile(filepath.Join(raiz, "go.mod"), []byte("module x\n\ngo 1.21\n"), 0o644)
	_ = os.WriteFile(filepath.Join(raiz, "a.go"), []byte("package x\n"), 0o644)
	_ = os.WriteFile(filepath.Join(raiz, "a_test.go"), []byte("package x\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) { "+asserta+" }\n"), 0o644)
	for _, a := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "x"}} {
		c := exec.Command("git", a...)
		c.Dir = raiz
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", a, err, out)
		}
	}
	return raiz
}

func rodarCILocal(raiz string) (string, error) {
	c := exec.Command("bash", "scripts/ci-local.sh")
	c.Dir = raiz
	out, err := c.CombinedOutput()
	return string(out), err
}

func TestCILocalFalhaComTesteQuebradoESemSelo(t *testing.T) {
	raiz := repoDeTeste(t, false)
	out, err := rodarCILocal(raiz)
	if err == nil {
		t.Fatalf("deveria falhar com teste quebrado: %s", out)
	}
	if _, e := os.Stat(filepath.Join(raiz, ".forge", "runtime", "ci-local.json")); e == nil {
		t.Fatalf("falha não pode gravar selo")
	}
}

func TestCILocalPassaEGravaSeloComHEAD(t *testing.T) {
	raiz := repoDeTeste(t, true)
	out, err := rodarCILocal(raiz)
	if err != nil {
		t.Fatalf("deveria passar: %v %s", err, out)
	}
	head, _ := exec.Command("git", "-C", raiz, "rev-parse", "HEAD").Output()
	selo, e := os.ReadFile(filepath.Join(raiz, ".forge", "runtime", "ci-local.json"))
	if e != nil || !strings.Contains(string(selo), strings.TrimSpace(string(head))) {
		t.Fatalf("selo ausente ou sem o HEAD: %v %s", e, selo)
	}
}

func TestCILocalSeloNaoVersionado(t *testing.T) {
	dir, _ := os.Getwd()
	gi, err := os.ReadFile(filepath.Join(dir, "..", ".gitignore"))
	if err != nil || !strings.Contains(string(gi), ".forge/") {
		t.Fatalf(".forge/ deve estar no .gitignore: %v", err)
	}
}
