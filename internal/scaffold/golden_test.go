package scaffold

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
)

// -update regenera os arquivos golden em testdata/golden/
var update = flag.Bool("update", false, "regenera golden fixtures")

func TestGoldenInit(t *testing.T) {
	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "postgres",
		Telemetry:  true,
		Lang:       "pt-BR",
		SddVersion: "1.5.2",
	}

	goldenDir := filepath.Join("testdata", "golden")

	if *update {
		// regenera: limpa e recria o golden dir
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatalf("removeAll golden: %v", err)
		}
		files, err := Run(cfg, goldenDir)
		if err != nil {
			t.Fatalf("Run (update): %v", err)
		}
		t.Logf("✓ golden atualizado: %d arquivos em %s", len(files), goldenDir)
		return
	}

	// cria golden se não existir
	if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
		t.Logf("golden não existe, criando em %s", goldenDir)
		if _, err := Run(cfg, goldenDir); err != nil {
			t.Fatalf("Run (init golden): %v", err)
		}
	}

	// roda scaffold em tmpdir
	tmpDir := t.TempDir()
	got, err := Run(cfg, tmpDir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// lista arquivos esperados
	var want []string
	err = filepath.WalkDir(goldenDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(goldenDir, path)
		want = append(want, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir golden: %v", err)
	}

	// normaliza lista got para caminhos relativos
	gotRel := make([]string, len(got))
	for i, p := range got {
		rel, _ := filepath.Rel(tmpDir, p)
		gotRel[i] = rel
	}

	// verifica contagem
	if len(gotRel) != len(want) {
		t.Errorf("contagem: got %d, want %d", len(gotRel), len(want))
	}

	// compara conteúdo de cada arquivo
	wantSet := make(map[string]bool, len(want))
	for _, w := range want {
		wantSet[w] = true
	}
	for _, rel := range gotRel {
		if !wantSet[rel] {
			t.Errorf("arquivo inesperado: %s", rel)
			continue
		}
		gotData, err := os.ReadFile(filepath.Join(tmpDir, rel))
		if err != nil {
			t.Errorf("read got %s: %v", rel, err)
			continue
		}
		wantData, err := os.ReadFile(filepath.Join(goldenDir, rel))
		if err != nil {
			t.Errorf("read want %s: %v", rel, err)
			continue
		}
		if strings.TrimSpace(string(gotData)) != strings.TrimSpace(string(wantData)) {
			t.Errorf("conteúdo diverge: %s\n--- want ---\n%s\n--- got ---\n%s",
				rel, string(wantData), string(gotData))
		}
	}

	t.Logf("✓ TestGoldenInit: %d arquivos comparados com golden", len(gotRel))
}
