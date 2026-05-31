package scaffold

import (
	"os"
	"strings"
	"testing"
	"github.com/forge-sdd/cli/internal/config"
)

func TestRunIntegration(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project: "demo", Stack: "go", DB: "postgres",
		Telemetry: true, Lang: "pt-BR", SddVersion: "1.1.0",
	}
	files, err := Run(cfg, dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("nenhum arquivo criado")
	}
	t.Logf("✓ %d arquivos criados", len(files))

	for _, f := range []string{
		dir + "/sdd/memory/constitution.md",
		dir + "/.github/copilot-instructions.md",
		dir + "/.vscode/mcp.json",
	} {
		if _, e := os.Stat(f); e != nil {
			t.Errorf("arquivo faltando: %s", f)
		}
	}

	data, err := os.ReadFile(dir + "/sdd/memory/constitution.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "demo") {
		t.Error("'demo' não encontrado em constitution.md")
	}
}
