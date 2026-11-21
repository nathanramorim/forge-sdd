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

// rodarPRGuard executa scripts/pr_guard.py com python3 -I, com stdin opcional.
func rodarPRGuard(t *testing.T, stdin string, args ...string) (string, int) {
	t.Helper()
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 ausente: teste do script não roda")
	}
	dir, _ := os.Getwd()
	cmd := exec.Command(py, append([]string{"-I", filepath.Join(dir, "pr_guard.py")}, args...)...)
	cmd.Dir = dir
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
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

func TestPRGuardPermiteCriarPRComDraft(t *testing.T) {
	casos := []string{
		"gh pr create --draft --fill",
		"gh pr create -d --title 'feat: x'",
		"gh -R owner/repo pr create --draft --base main",
		"az repos pr create --draft true --title x",
		"az repos pr create --draft=true --title x",
		"az repos pr update --id 7 --title 'novo titulo'",
		"gh pr list --state open",
		"gh pr view 7",
		"git status && go test ./...",
	}
	for _, c := range casos {
		out, rc := rodarPRGuard(t, "", "--check", c)
		if rc != 0 {
			t.Fatalf("comando %q deveria passar (rc=0), mas rc=%d out=%s", c, rc, out)
		}
	}
}

func TestPRGuardBloqueiaCriarSemDraftEMergeSemVeredito(t *testing.T) {
	casos := []string{
		"gh pr create --fill",
		"gh pr create --draft=false --fill",
		"gh pr merge 7",
		"gh pr ready 7",
		"gh pr ready",
		"az repos pr create --title x",
		"az repos pr create --draft false --title x",
		"az repos pr update --id 7 --draft false",
		"az repos pr update --id 7 --draft=false",
		"az repos pr update --id 7 --status completed",
		"az repos pr update --id 7 --status=completed",
	}
	for _, c := range casos {
		out, rc := rodarPRGuard(t, "", "--check", c)
		if rc == 0 {
			t.Fatalf("comando %q deveria falhar (rc!=0), mas passou! out=%s", c, out)
		}
		if !strings.Contains(out, "Claude Code") {
			t.Fatalf("esperava aviso de cobertura do Claude Code na saída de %q: %s", c, out)
		}
	}
}

func TestPRGuardBloqueiaTentativasDeBypassComEncadeamento(t *testing.T) {
	casos := []string{
		"echo ok && gh pr merge 7",
		"echo ok ; gh pr create --fill",
		"false || gh pr ready 7",
		"echo ok | gh pr merge 7",
		"echo $(gh pr merge 7)",
		"echo `gh pr create --fill`",
		"(gh pr merge 7)",
		"env GH_TOKEN=abc gh pr create --fill",
		"command gh pr merge 7",
		"bash -c 'gh pr merge 7'",
		"sh -c \"az repos pr update --id 7 --draft false\"",
		"echo linha1\ngh pr create --fill",
	}
	for _, c := range casos {
		out, rc := rodarPRGuard(t, "", "--check", c)
		if rc == 0 {
			t.Fatalf("bypass %q deveria ser bloqueado (rc!=0), mas passou! out=%s", c, out)
		}
	}
}

func TestPRGuardReadyEMergePassamComVereditoConformeDoHEAD(t *testing.T) {
	v := veredito(t, vJSON(headFalso, "entrega", "conforme"))
	for _, c := range []string{
		"gh pr ready 7",
		"gh pr merge 7",
		"az repos pr update --id 7 --draft false",
		"az repos pr update --id 7 --status completed",
	} {
		out, rc := rodarPRGuard(t, "", "--check", c, "--veredito", v, "--head", headFalso)
		if rc != 0 {
			t.Fatalf("comando %q com veredito conforme do HEAD deveria passar: rc=%d out=%s", c, rc, out)
		}
	}

	// Negativo: SHA diferente do HEAD continua bloqueando
	vOutroSHA := veredito(t, vJSON("1111111", "entrega", "conforme"))
	_, rc := rodarPRGuard(t, "", "--check", "gh pr merge 7", "--veredito", vOutroSHA, "--head", headFalso)
	if rc == 0 {
		t.Fatalf("merge com SHA divergente do HEAD deveria ser bloqueado")
	}

	// Negativo: veredito violacao continua bloqueando
	vViolacao := veredito(t, vJSON(headFalso, "entrega", "violacao"))
	_, rc = rodarPRGuard(t, "", "--check", "gh pr ready 7", "--veredito", vViolacao, "--head", headFalso)
	if rc == 0 {
		t.Fatalf("ready com veredito violacao deveria ser bloqueado")
	}
}

func TestPRGuardModoHookPreToolUseViaStdin(t *testing.T) {
	okPayload := `{"tool_name":"Bash","tool_input":{"command":"gh pr create --draft --fill"}}`
	out, rc := rodarPRGuard(t, okPayload, "--hook")
	if rc != 0 {
		t.Fatalf("hook com --draft deveria passar: rc=%d out=%s", rc, out)
	}

	bloqueadoPayload := `{"tool_name":"Bash","tool_input":{"command":"gh pr create --fill"}}`
	_, rc = rodarPRGuard(t, bloqueadoPayload, "--hook")
	if rc == 0 {
		t.Fatalf("hook sem --draft deveria bloquear")
	}

	invalidoPayload := `{nao-json`
	_, rc = rodarPRGuard(t, invalidoPayload, "--hook")
	if rc == 0 {
		t.Fatalf("hook com JSON inválido deveria falhar")
	}
}
