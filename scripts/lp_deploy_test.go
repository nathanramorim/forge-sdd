package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoLP monta um repo git mínimo com lp-deploy.sh e um novidades_lp.py falso.
// Sem rede: o wrangler nunca é chamado nestes testes.
func repoLP(t *testing.T, checkOk bool) string {
	t.Helper()
	raiz := t.TempDir()
	dir, _ := os.Getwd()
	sh, err := os.ReadFile(filepath.Join(dir, "lp-deploy.sh"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(raiz, "scripts"), 0o755)
	_ = os.MkdirAll(filepath.Join(raiz, "docs"), 0o755)
	_ = os.WriteFile(filepath.Join(raiz, "scripts", "lp-deploy.sh"), sh, 0o755)
	py := "import sys\nsys.exit(0)\n"
	if !checkOk {
		py = "import sys\nsys.exit(1 if '--check' in sys.argv else 0)\n"
	}
	_ = os.WriteFile(filepath.Join(raiz, "scripts", "novidades_lp.py"), []byte(py), 0o644)
	_ = os.WriteFile(filepath.Join(raiz, "docs", "index.html"), []byte("x"), 0o644)
	for _, a := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "x"}} {
		c := exec.Command("git", a...)
		c.Dir = raiz
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", a, err, out)
		}
	}
	return raiz
}

func rodaLP(t *testing.T, raiz string, args ...string) (string, error) {
	t.Helper()
	c := exec.Command("bash", append([]string{"scripts/lp-deploy.sh"}, args...)...)
	c.Dir = raiz
	// PATH sem npx: qualquer chamada ao wrangler falharia alto.
	c.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	out, err := c.CombinedOutput()
	return string(out), err
}

func TestLPDeployDryRunSemRede(t *testing.T) {
	raiz := repoLP(t, true)
	out, err := rodaLP(t, raiz, "--dry-run")
	if err != nil {
		t.Fatalf("dry-run deveria sair 0: %v %s", err, out)
	}
	for _, want := range []string{"novidades_lp.py", "--check", "wrangler@4.165.0 deploy"} {
		if !strings.Contains(out, want) {
			t.Errorf("saida sem %q: %s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(raiz, ".forge", "runtime", "lp-deploy.json")); err == nil {
		t.Error("dry-run nao deve gravar o registro")
	}
}

func TestLPDeployRecusaQuandoCheckFalha(t *testing.T) {
	raiz := repoLP(t, false)
	out, err := rodaLP(t, raiz)
	if err == nil {
		t.Fatalf("deveria recusar: %s", out)
	}
	if !strings.Contains(out, "recusado") {
		t.Errorf("sem mensagem de recusa: %s", out)
	}
	if _, err := os.Stat(filepath.Join(raiz, ".forge", "runtime", "lp-deploy.json")); err == nil {
		t.Error("recusa nao deve gravar o registro")
	}
}

func TestLPDeployStatusAvisaDesatualizada(t *testing.T) {
	raiz := repoLP(t, true)
	out, err := rodaLP(t, raiz, "--status")
	if err != nil || !strings.Contains(out, "LP desatualizada") {
		t.Fatalf("esperava aviso: %v %s", err, out)
	}
	sha, _ := exec.Command("git", "-C", raiz, "log", "-1", "--format=%H", "--", "docs").Output()
	_ = os.MkdirAll(filepath.Join(raiz, ".forge", "runtime"), 0o755)
	_ = os.WriteFile(filepath.Join(raiz, ".forge", "runtime", "lp-deploy.json"),
		[]byte(`{"docs_commit": "`+strings.TrimSpace(string(sha))+`"}`), 0o644)
	out, _ = rodaLP(t, raiz, "--status")
	if strings.Contains(out, "desatualizada") || !strings.Contains(out, "em dia") {
		t.Errorf("deveria estar em dia: %s", out)
	}
}

const contaTeste = "0123456789abcdef0123456789abcdef"

func TestLPDeployRecusaSemContaExplicita(t *testing.T) {
	raiz := repoLP(t, true)
	out, err := rodaLP(t, raiz)
	if err == nil {
		t.Fatalf("sem conta deveria recusar: %s", out)
	}
	if !strings.Contains(out, "recusado") || !strings.Contains(out, "conta do Cloudflare") {
		t.Errorf("mensagem sem recusa de conta: %s", out)
	}
	if _, err := os.Stat(filepath.Join(raiz, ".forge", "runtime", "lp-deploy.json")); err == nil {
		t.Error("recusa nao deve gravar o registro")
	}
}

func TestLPDeployRecusaContaMalformada(t *testing.T) {
	raiz := repoLP(t, true)
	out, err := rodaLP(t, raiz, "--conta", "abc")
	if err == nil || !strings.Contains(out, "recusado") {
		t.Fatalf("conta malformada deveria recusar: %v %s", err, out)
	}
}

func TestLPDeployDryRunMostraContaMascarada(t *testing.T) {
	raiz := repoLP(t, true)
	out, err := rodaLP(t, raiz, "--dry-run", "--conta", contaTeste)
	if err != nil {
		t.Fatalf("dry-run deveria sair 0: %v %s", err, out)
	}
	if !strings.Contains(out, "****cdef") {
		t.Errorf("sem a conta mascarada: %s", out)
	}
	if strings.Contains(out, contaTeste) {
		t.Errorf("o id completo da conta nao pode ser impresso: %s", out)
	}
}

func TestLPDeployLeContaDoArquivoLocal(t *testing.T) {
	raiz := repoLP(t, true)
	_ = os.MkdirAll(filepath.Join(raiz, ".forge", "config"), 0o755)
	_ = os.WriteFile(filepath.Join(raiz, ".forge", "config", "lp-deploy.json"), []byte(`{"account_id": "`+contaTeste+`"}`), 0o644)
	out, err := rodaLP(t, raiz, "--dry-run")
	if err != nil || !strings.Contains(out, "****cdef") {
		t.Fatalf("deveria ler a conta do arquivo local: %v %s", err, out)
	}
}

func TestLPDeployIgnoraContaDoAmbienteDoShell(t *testing.T) {
	raiz := repoLP(t, true)
	c := exec.Command("bash", "scripts/lp-deploy.sh")
	c.Dir = raiz
	c.Env = append(os.Environ(), "PATH=/usr/bin:/bin", "CLOUDFLARE_ACCOUNT_ID="+contaTeste)
	out, err := c.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "recusado") {
		t.Fatalf("CLOUDFLARE_ACCOUNT_ID do ambiente nao pode bastar: %v %s", err, out)
	}
}

func TestLPDeployLeContaDeJSONEmVariasLinhas(t *testing.T) {
	raiz := repoLP(t, true)
	_ = os.MkdirAll(filepath.Join(raiz, ".forge", "config"), 0o755)
	_ = os.WriteFile(filepath.Join(raiz, ".forge", "config", "lp-deploy.json"), []byte("{\n  \"account_id\":\n    \""+contaTeste+"\"\n}\n"), 0o644)
	out, err := rodaLP(t, raiz, "--dry-run")
	if err != nil || !strings.Contains(out, "****cdef") {
		t.Fatalf("deveria ler a conta de JSON em varias linhas: %v %s", err, out)
	}
}

func TestDeployLPWorkflowFornecaAConta(t *testing.T) {
	dir, _ := os.Getwd()
	wf, err := os.ReadFile(filepath.Join(dir, "..", ".github", "workflows", "deploy-lp.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wf), "LP_CLOUDFLARE_ACCOUNT_ID: ${{ vars.LP_CLOUDFLARE_ACCOUNT_ID }}") {
		t.Error("o workflow manual precisa passar LP_CLOUDFLARE_ACCOUNT_ID ao scripts/lp-deploy.sh")
	}
}
