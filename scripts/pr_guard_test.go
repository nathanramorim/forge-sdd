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

// TestPRGuardBloqueiaMergeComBaseForaDaPrincipal: um PR com base em outra branch entra por
// squash e deixa os commits de cima órfãos — o arquivo some da principal (feat-106-24d).
// O teste usa um `gh` falso no PATH para simular a base do PR.
func TestPRGuardBloqueiaMergeComBaseForaDaPrincipal(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 ausente")
	}
	dir, _ := os.Getwd()
	raizProjeto, _ := filepath.Abs(filepath.Join(dir, ".."))

	// Veredito conforme para o HEAD, para o guard chegar até a checagem da base.
	head := strings.TrimSpace(strings.Split(string(mustRun(t, raizProjeto, "git", "rev-parse", "HEAD")), "\n")[0])

	casos := map[string]struct {
		base     string
		quebrar  bool
		bloqueia bool
	}{
		"base na main passa":                 {base: "main"},
		"base em branch de feature bloqueia": {base: "feat/106-24b-leque-por-worktree", bloqueia: true},
	}

	for nome, c := range casos {
		t.Run(nome, func(t *testing.T) {
			bin := t.TempDir()
			script := "#!/bin/sh\n"
			if c.quebrar {
				script += "exit 1\n"
			} else {
				script += "case \"$1 $2\" in\n"
				script += "  'pr view') echo '{\"baseRefName\":\"" + c.base + "\"}' ;;\n"
				script += "  'repo view') echo '{\"defaultBranchRef\":{\"name\":\"main\"}}' ;;\n"
				script += "  *) echo '{}' ;;\n"
				script += "esac\n"
			}
			ghPath := filepath.Join(bin, "gh")
			if err := os.WriteFile(ghPath, []byte(script), 0o755); err != nil {
				t.Fatalf("escrever gh falso: %v", err)
			}

			veredito := filepath.Join(t.TempDir(), "veredito.json")
			conteudo := `{"veredito":"conforme","head_sha":"` + head + `","ponto":"entrega","feature":"feat-106-24d"}`
			if err := os.WriteFile(veredito, []byte(conteudo), 0o644); err != nil {
				t.Fatalf("escrever veredito: %v", err)
			}

			cmd := exec.Command(py, "-I", filepath.Join(dir, "pr_guard.py"),
				"--check", "gh pr merge 7 --squash", "--veredito", veredito, "--head", head)
			cmd.Dir = raizProjeto
			cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"))
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			err := cmd.Run()

			saida := out.String()
			if c.quebrar {
				// Sem conseguir ler a base, o guard não inventa: deixa passar com o veredito.
				if err != nil {
					t.Fatalf("sem base legível não deveria bloquear: %s", saida)
				}
				return
			}
			if c.bloqueia {
				if err == nil {
					t.Fatalf("deveria bloquear merge com base fora da principal: %s", saida)
				}
				if !strings.Contains(saida, "RN-22") || !strings.Contains(saida, "base") {
					t.Fatalf("a mensagem precisa citar a RN-22 e a base: %s", saida)
				}
				return
			}
			if err != nil {
				t.Fatalf("merge com base na principal deveria passar: %s", saida)
			}
		})
	}
}

// mustRun roda um comando e devolve a saída; falha o teste se o comando errar.
func mustRun(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
	}
	return out
}
