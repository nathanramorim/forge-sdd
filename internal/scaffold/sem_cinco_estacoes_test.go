package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cincoEstacoes é o texto obsoleto (ADR-0021): o produto são 3 sessões.
const cincoEstacoes = "5 estações"

// isentoCincoEstacoes: nenhuma isenção; o lider foi corrigido na feat-105-06.
func isentoCincoEstacoes(string) bool { return false }

func varreSemCincoEstacoes(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || isentoCincoEstacoes(p) {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), cincoEstacoes) {
			t.Errorf("%s ainda diz %q; use \"3 sessões (Discovery, Builder, Reviewer)\"", p, cincoEstacoes)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("varredura de %s: %v", root, err)
	}
}

func TestTemplatesSemCincoEstacoes(t *testing.T) {
	varreSemCincoEstacoes(t, "templates")
}

// Negativo: o golden entregue ao usuário não volta a dizer "5 estações".
func TestGoldensSemTextoDeCincoEstacoes(t *testing.T) {
	varreSemCincoEstacoes(t, filepath.Join("testdata", "golden"))
}

// TestLiderDescricaoSemCincoEstacoes: o agente lider descreve a sessão Discovery.
func TestLiderDescricaoSemCincoEstacoes(t *testing.T) {
	for _, p := range []string{
		"templates/agents/claude/.claude/agents/lider.md.tmpl",
		"../../.claude/agents/lider.md",
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if strings.Contains(s, cincoEstacoes) || !strings.Contains(s, "sessão Discovery") {
			t.Errorf("%s: descrição do lider fora do papel real", p)
		}
		if len(s) > 1536 {
			t.Errorf("%s: %d bytes acima do orçamento (P-24)", p, len(s))
		}
	}
}

// Negativo: o lider continua no scaffold e no repositório (nada renomeado/removido).
func TestLiderContinuaNoScaffold(t *testing.T) {
	for _, p := range []string{
		"templates/agents/claude/.claude/agents/lider.md.tmpl",
		"../../.claude/agents/lider.md",
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("lider sumiu: %v", err)
		}
		if !strings.Contains(string(b), "name: lider\n") {
			t.Errorf("%s: name: lider ausente", p)
		}
	}
}
