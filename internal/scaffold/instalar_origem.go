package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LerOrigemInstalacao lê um diretório de origem (clone feito fora do Go) num mapa chave relativa
// (separador "/") para conteúdo. Ignora `.git`; recusa link simbólico e arquivo não regular.
func LerOrigemInstalacao(origem string) (map[string][]byte, error) {
	fi, err := os.Stat(origem)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("origem %q não é um diretório", origem)
	}
	arquivos := map[string][]byte{}
	err = filepath.WalkDir(origem, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s é link simbólico ou não é arquivo regular; recusado (SEG-02)", p)
		}
		rel, err := filepath.Rel(origem, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		arquivos[filepath.ToSlash(rel)] = data
		return nil
	})
	return arquivos, err
}
