package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// Ações de um item do plano do update.
const (
	AcaoCriar      = "criar"
	AcaoReescrever = "reescrever"
	AcaoIgual      = "igual"
	AcaoPreservar  = "preservar"
)

// ItemPlano é o destino de um template e o que o update faria com ele.
type ItemPlano struct {
	Rel  string // caminho relativo ao projeto, com "/"
	Acao string
	Tipo string // "instrucao", "regra" ou "outro"
}

// PlanoDoUpdate devolve, sem gravar nada (ARQ-05), o que Run faria com cada template de cfg. Usa a
// mesma lista de raízes e a mesma decisão (decideDestino) do update real (DUP-06).
func PlanoDoUpdate(targetDir string, cfg config.Config) ([]ItemPlano, error) {
	var itens []ItemPlano
	for _, r := range raizesDoRun(cfg) {
		err := fs.WalkDir(templatesFS, r.root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel := strings.TrimSuffix(strings.TrimPrefix(p, r.strip+"/"), ".tmpl")
			dest := filepath.Join(targetDir, rel)
			data, err := templatesFS.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read template %s: %w", p, err)
			}
			rendered, err := renderTemplate(p, data, cfg)
			if err != nil {
				return err
			}
			relSlash := filepath.ToSlash(rel)
			dec := decideDestino(relSlash, dest, targetDir, rendered, cfg)
			item := ItemPlano{Rel: relSlash, Tipo: dec.Tipo}
			switch got, rerr := os.ReadFile(dest); {
			case dec.Preservar:
				item.Acao = AcaoPreservar
			case rerr != nil:
				item.Acao = AcaoCriar
			case bytes.Equal(got, rendered):
				item.Acao = AcaoIgual
			default:
				item.Acao = AcaoReescrever
			}
			itens = append(itens, item)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return itens, nil
}

// PlanoHookInbox informa se o hook do inbox (UserPromptSubmit) seria instalado, sem gravar.
func PlanoHookInbox(targetDir string) (bool, error) {
	return mergeClaudeHook(targetDir, "UserPromptSubmit", inboxHookCommand, []string{"forge inbox", "forge-sdd inbox"}, true)
}
