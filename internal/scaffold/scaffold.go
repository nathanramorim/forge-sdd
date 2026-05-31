package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
// Retorna a lista de arquivos criados.
func Run(cfg config.Config, targetDir string) ([]string, error) {
	var created []string

	err := fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// destPath: strip "templates/" prefix e ".tmpl" suffix
		rel := strings.TrimPrefix(path, "templates/")
		rel = strings.TrimSuffix(rel, ".tmpl")
		dest := filepath.Join(targetDir, rel)

		// criar diretório pai
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}

		// ler conteúdo do template
		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}

		// renderizar
		tmpl, err := template.New(path).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", path, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("render template %s: %w", path, err)
		}

		// escrever arquivo
		if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}

		created = append(created, dest)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}
