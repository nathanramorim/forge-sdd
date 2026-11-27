package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/forge-sdd/cli/internal/notify"
)

// marcadorGitattributes identifica o bloco que o forge acrescenta ao .gitattributes do projeto.
const marcadorGitattributes = "# forge-sdd:eol"

// blocoGitattributes fixa LF só nos arquivos do forge (specs, rules, skills e comandos), para o
// checkout no Windows não gravar CRLF neles. O código do usuário fica de fora.
const blocoGitattributes = marcadorGitattributes + " (arquivos do Forge SDD sempre com LF)\n" +
	"sdd/** text eol=lf\n" +
	".agents/** text eol=lf\n" +
	".claude/** text eol=lf\n" +
	".forge/** text eol=lf\n"

// garantirGitattributes acrescenta o bloco do forge ao .gitattributes (cria o arquivo se faltar).
// Idempotente: com o marcador presente, não toca no arquivo. Falha vira aviso e não aborta.
func garantirGitattributes(targetDir string) {
	dest := filepath.Join(targetDir, ".gitattributes")
	atual, err := os.ReadFile(dest)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "aviso: .gitattributes não lido: %s\n", notify.Sanitize(err.Error(), 400, false))
		return
	}
	if bytes.Contains(atual, []byte(marcadorGitattributes)) {
		return
	}
	novo := append([]byte{}, atual...)
	if len(novo) > 0 {
		if !bytes.HasSuffix(novo, []byte("\n")) {
			novo = append(novo, '\n')
		}
		novo = append(novo, '\n')
	}
	novo = append(novo, blocoGitattributes...)
	if err := writeFileSafeMode(targetDir, dest, novo, 0644); err != nil {
		refusals.Add(1)
		fmt.Fprintf(os.Stderr, "aviso: .gitattributes não atualizado: %s\n", notify.Sanitize(err.Error(), 400, false))
	}
}
