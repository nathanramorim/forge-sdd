package scaffold

import (
	"embed"
	"io/fs"
	"path"

	"github.com/forge-sdd/cli/internal/skills"
)

// Mods de referência que viajam dentro do binário (ADR-0030): o pacote do npm leva só o binário,
// então `forge skills instalar plugin <slug>` sem --de usa esta cópia. A fonte de desenvolvimento é
// .agents/plugins/<slug>; TestModsEmbutidosIguaisAoRepositorio impede que as duas divirjam.
//
//go:embed all:mods
var modsFS embed.FS

// ModEmbutido devolve os arquivos do mod embutido no binário, com caminhos relativos à pasta do mod.
func ModEmbutido(slug string) (map[string][]byte, bool) {
	raiz := path.Join("mods", slug)
	if skills.ValidateSlug(slug) != nil {
		return nil, false
	}
	if st, err := fs.Stat(modsFS, raiz); err != nil || !st.IsDir() {
		return nil, false
	}
	arquivos := map[string][]byte{}
	err := fs.WalkDir(modsFS, raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := modsFS.ReadFile(p)
		if err != nil {
			return err
		}
		arquivos[p[len(raiz)+1:]] = data
		return nil
	})
	if err != nil || len(arquivos) == 0 {
		return nil, false
	}
	return arquivos, true
}

// ModsEmbutidos lista os slugs dos mods que viajam no binário.
func ModsEmbutidos() []string {
	ents, _ := fs.ReadDir(modsFS, "mods")
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
