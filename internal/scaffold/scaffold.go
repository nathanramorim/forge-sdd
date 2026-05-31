package scaffold

import (
	"embed"
	"io/fs"

	"github.com/forge-sdd/cli/internal/config"
)

//go:embed all:templates
var templatesFS embed.FS

// Walk retorna a lista de paths dos templates embutidos (sem o prefixo "templates/").
func Walk() []string {
	var paths []string
	fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

// Run renderiza os templates com cfg e escreve os arquivos em targetDir.
// Implementado em feat-02-init-interactive.
func Run(cfg config.Config, targetDir string) ([]string, error) {
	return nil, nil
}
