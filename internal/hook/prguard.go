// Package hook guarda o wrapper dos hooks do Claude Code (fix-106-14b, ADR-0027).
package hook

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PRGuardScript é o caminho do guard relativo à raiz do projeto.
var PRGuardScript = filepath.Join("scripts", "pr_guard.py")

// ProjectRoot resolve a raiz do projeto: CLAUDE_PROJECT_DIR, senão o toplevel do git, senão o cwd.
func ProjectRoot() string {
	if d := os.Getenv("CLAUDE_PROJECT_DIR"); d != "" {
		return d
	}
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		if d := strings.TrimSpace(string(out)); d != "" {
			return d
		}
	}
	d, _ := os.Getwd()
	return d
}

// RunPRGuard executa o pr_guard do projeto em root e devolve o código de saída do hook.
// Sem o script ou sem `uv`, avisa em stderr e devolve 0 (fail-open consciente: o job "pronto?"
// do CI é o apoio). Com os dois, repassa stdin/stdout/stderr e devolve o código do guard.
func RunPRGuard(root string, stdin io.Reader, stdout, stderr io.Writer) int {
	script := filepath.Join(root, PRGuardScript)
	if fi, err := os.Stat(script); err != nil || fi.IsDir() {
		fmt.Fprintf(stderr, "forge: %s não existe; barreira de PR ignorada (rode 'forge update')\n", PRGuardScript)
		return 0
	}
	uv, err := exec.LookPath("uv")
	if err != nil {
		fmt.Fprintln(stderr, "forge: 'uv' não está no PATH; barreira de PR ignorada (instale em https://docs.astral.sh/uv/)")
		return 0
	}
	c := exec.Command(uv, "run", PRGuardScript, "--hook")
	c.Dir = root
	c.Stdin, c.Stdout, c.Stderr = stdin, stdout, stderr
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() >= 0 {
			return ee.ExitCode()
		}
		fmt.Fprintf(stderr, "forge: falha ao executar o guard: %v\n", err)
		return 0
	}
	return 0
}
