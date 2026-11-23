package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/notify"
)

const schemaMetricasRel = "sdd/.metrics/schema.json"

// regraGerenciada diz se o caminho (relativo, com "/") é uma rule sdd-* ou o schema de métricas,
// arquivos que o forge entregou e atualiza por igualdade byte a byte (feat-118-04). Rules de
// outros nomes pertencem ao usuário e nunca entram.
func regraGerenciada(rel string) bool {
	if rel == schemaMetricasRel {
		return true
	}
	return strings.HasPrefix(rel, ".agents/rules/sdd-") && strings.HasSuffix(rel, ".md")
}

// arquivoDoForge: o arquivo existente é regular e igual ao que o forge entregou na 5.0 (tabela embed,
// reusa renderV50, DUP-06) ou ao template atual (nada a mudar). Qualquer outra coisa é do usuário.
func arquivoDoForge(rel, dest string, atual []byte, cfg config.Config) bool {
	fi, err := os.Lstat(dest)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		return false
	}
	if bytes.Equal(got, atual) {
		return true
	}
	want, ok := renderV50(rel, cfg)
	return ok && bytes.Equal(got, want)
}

// avisaRegraPreservada avisa em stderr e grava `<nome>.novo.<ext>` com a versão nova do forge.
func avisaRegraPreservada(targetDir, dest, rel string, rendered []byte) {
	novo := novoPath(dest)
	msg := "aviso: %s tem versão nova do forge e foi mantido (você o editou); revise o conteúdo novo em %s\n"
	if err := writeFileSafeMode(targetDir, novo, rendered, 0644); err != nil {
		refusals.Add(1)
		msg = "aviso: %s tem versão nova do forge e foi mantido (você o editou); %s não gravado\n"
	}
	fmt.Fprintf(os.Stderr, msg, notify.Sanitize(rel, 300, false), notify.Sanitize(filepath.ToSlash(novo), 300, false))
}

// PlanoRegras classifica as rules sdd-* e o schema.json existentes: atualizar (do forge, diferem do
// template atual) e manter (editados pelo usuário). Não escreve nada (ARQ-05).
func PlanoRegras(targetDir string, cfg config.Config) (atualizar, manter []string) {
	_ = fs.WalkDir(templatesFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(p, "templates/"), ".tmpl")
		if !regraGerenciada(rel) {
			return nil
		}
		dest := filepath.Join(targetDir, filepath.FromSlash(rel))
		if _, serr := os.Lstat(dest); serr != nil {
			return nil
		}
		data, rerr := templatesFS.ReadFile(p)
		if rerr != nil {
			return nil
		}
		rendered, rerr := renderTemplate(p, data, cfg)
		if rerr != nil {
			return nil
		}
		got, _ := os.ReadFile(dest)
		switch {
		case bytes.Equal(got, rendered):
		case arquivoDoForge(rel, dest, rendered, cfg):
			atualizar = append(atualizar, rel)
		default:
			manter = append(manter, rel)
		}
		return nil
	})
	return atualizar, manter
}
