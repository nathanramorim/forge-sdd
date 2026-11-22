package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIPoliticaRunsPorEvento(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 ausente")
	}
	dir, _ := os.Getwd()
	runs := `{"workflow_runs":[
	 {"event":"pull_request","created_at":"2026-11-22T10:00:00Z","jobs":[{"name":"test","conclusion":"success","started_at":"2026-11-22T10:00:00Z","completed_at":"2026-11-22T10:02:00Z"}]},
	 {"event":"pull_request","created_at":"2026-11-22T11:00:00Z","jobs":[]},
	 {"event":"push","created_at":"2026-11-21T09:00:00Z","jobs":[]}]}`
	arq := filepath.Join(t.TempDir(), "runs.json")
	if err := os.WriteFile(arq, []byte(runs), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(py, "-I", filepath.Join(dir, "ci-politica.py"), "--offline", "--relatorio", arq).CombinedOutput()
	if err != nil {
		t.Fatalf("falhou: %v %s", err, out)
	}
	s := string(out)
	for _, quer := range []string{"runs por dia por evento", "2026-11-22", "pull_request", "push"} {
		if !strings.Contains(s, quer) {
			t.Fatalf("faltou %q na saída:\n%s", quer, s)
		}
	}
}
