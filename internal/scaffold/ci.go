package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/notify"
)

// Workflow opt-in da feat-118-05. O template vive em templates/ci (fora de globalRoots) para nunca
// ser renderizado sem --ci, nem com o Copilot (que põe templates/.github em globalRoots).
const (
	forgeProntoTemplate = "templates/ci/forge-pronto.yml.tmpl"
	forgeProntoRel      = ".github/workflows/forge-pronto.yml"
	ciGeneratedMarker   = "# forge-sdd:generated"
)

// ForgeProntoPath devolve o destino do workflow no projeto.
func ForgeProntoPath(targetDir string) string {
	return filepath.Join(targetDir, filepath.FromSlash(forgeProntoRel))
}

func isGeneratedCI(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for i := 0; i < 3 && sc.Scan(); i++ {
		if strings.TrimSpace(sc.Text()) == ciGeneratedMarker {
			return true
		}
	}
	return false
}

// RenderCI grava .github/workflows/forge-pronto.yml. Só é chamado com --ci. Arquivo existente sem o
// marcador é do usuário: preservado com aviso (SEG-02). O ci.yml do usuário nunca é aberto.
// Devolve os destinos tratados (em dry-run, os que seriam gravados).
func RenderCI(cfg config.Config, targetDir string) ([]string, error) {
	dest := ForgeProntoPath(targetDir)
	if cfg.DryRun {
		fmt.Printf("[DRY] %s\n", dest)
		return []string{dest}, nil
	}
	if _, err := os.Lstat(dest); err == nil && !isGeneratedCI(dest) {
		fmt.Fprintf(os.Stderr, "aviso: %s já existe e não é do forge-sdd (sem o marcador %s); preservado\n",
			notify.Sanitize(forgeProntoRel, 300, false), ciGeneratedMarker)
		return nil, nil
	}
	data, err := templatesFS.ReadFile(forgeProntoTemplate)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", forgeProntoTemplate, err)
	}
	rendered, err := renderTemplate(forgeProntoTemplate, data, cfg)
	if err != nil {
		return nil, err
	}
	if err := writeFileSafeMode(targetDir, dest, rendered, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %s não instalado: %s\n", notify.Sanitize(forgeProntoRel, 300, false), notify.Sanitize(err.Error(), 400, false))
		return nil, nil
	}
	return []string{dest}, nil
}
