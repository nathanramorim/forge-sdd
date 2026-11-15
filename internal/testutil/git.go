// Package testutil reúne auxiliares compartilhados pelos testes (DUP-05): repositório git
// temporário e escrita de arquivos. Não é usado em código de produção.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Git executa o git em dir com argv explícito e falha o teste se der erro. Devolve a saída.
func Git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.email=t@t", "-c", "user.name=t"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// InitRepo cria um repositório em dir na branch indicada, com um commit vazio inicial.
func InitRepo(t testing.TB, dir, branch string) {
	t.Helper()
	Git(t, dir, "init", "-q", "-b", branch)
	Git(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
}

// WriteFile grava content em root/rel criando as pastas necessárias.
func WriteFile(t testing.TB, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}
