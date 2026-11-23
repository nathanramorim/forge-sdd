package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// stackValueRe é a lista de caracteres aceitos para stack e banco recuperados dos arquivos do projeto
// (SEG-01): sem aspas, barra invertida, quebra de linha, espaço ou chaves; tamanho máximo 40.
var stackValueRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+#-]{0,39}$`)

var stackProseRe = regexp.MustCompile(`stack (\S+)\.`)

// ValidStackValue informa se v é um valor de stack/banco seguro para gravar no .sddrc e nos adaptadores.
func ValidStackValue(v string) bool {
	return stackValueRe.MatchString(v) && !strings.Contains(v, "{{")
}

func readProjectFile(targetDir, rel string) string {
	p := filepath.Join(targetDir, filepath.FromSlash(rel))
	fi, err := os.Lstat(p)
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > 1<<20 {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// tableCell devolve a escolha (2ª coluna) da linha de tabela `| <label> | valor | ...`.
func tableCell(content, label string) string {
	for _, line := range strings.Split(content, "\n") {
		cols := strings.Split(strings.TrimSpace(line), "|")
		if len(cols) < 3 || strings.TrimSpace(cols[1]) != label {
			continue
		}
		if v := strings.TrimSpace(cols[2]); ValidStackValue(v) {
			return v
		}
	}
	return ""
}

// bulletValue devolve o valor da linha `- **<label>:** valor`.
func bulletValue(content, label string) string {
	prefix := "- **" + label + ":**"
	for _, line := range strings.Split(content, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), prefix); ok {
			if v := strings.TrimSpace(rest); ValidStackValue(v) {
				return v
			}
		}
	}
	return ""
}

// RecoverStack recupera stack e banco de um projeto da 5.0 cujo .sddrc não os guarda. Por campo, vence a
// primeira fonte com valor válido: (1) Constituição e spec/stack.md; (2) system_instructions de gemini e
// openai; (3) prosa `stack X.` dos adaptadores (só stack). Sem valor, devolve vazio: nunca assume go/none.
func RecoverStack(targetDir string) (stack, db string) {
	for _, rel := range []string{"sdd/memory/constitution.md", "sdd/spec/stack.md"} {
		c := readProjectFile(targetDir, rel)
		if stack == "" {
			stack = tableCell(c, "Runtime")
		}
		if db == "" {
			db = tableCell(c, "DB")
		}
	}
	for _, rel := range []string{".gemini/system_instructions.md", ".openai/system_instructions.md"} {
		c := readProjectFile(targetDir, rel)
		if stack == "" {
			stack = bulletValue(c, "Stack")
		}
		if db == "" {
			db = bulletValue(c, "Banco")
		}
	}
	for _, rel := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md", ".github/copilot-instructions.md"} {
		if stack != "" {
			break
		}
		if m := stackProseRe.FindStringSubmatch(readProjectFile(targetDir, rel)); m != nil && ValidStackValue(m[1]) {
			stack = m[1]
		}
	}
	return stack, db
}
