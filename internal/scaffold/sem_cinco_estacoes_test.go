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

// isentoCincoEstacoes lista arquivos cuja troca é de outra task. O agente
// lider tem a descrição trocada na feat-105-06; remover a isenção lá.
func isentoCincoEstacoes(path string) bool {
	base := filepath.Base(path)
	return base == "lider.md.tmpl" || base == "lider.md"
}

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
