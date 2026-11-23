package scaffold

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/vcs"
)

const secaoFerramentas = "## Ferramentas e Integrações"

var (
	linhaTabelaRe = regexp.MustCompile(`^\|\s*([^|]+?)\s*\|`)
	separadorRe   = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)
)

// AcrescentaCamposDeModo é o passo à parte do `update` (ADR-0028): depois de Run, acrescenta à
// tabela `## Ferramentas e Integrações` da Constituição do usuário só as linhas de modo que
// faltam, com os padrões seguros. Nunca altera linha existente (SEG-02); sem linhas a
// acrescentar não grava. Seção sem tabela: não grava e avisa em warn. Devolve quantas linhas
// acrescentou. Constituição ausente: nada a fazer.
func AcrescentaCamposDeModo(targetDir string, warn io.Writer) (int, error) {
	path, out, rotulos, err := calculaCamposDeModo(targetDir, warn)
	if err != nil || len(rotulos) == 0 {
		return 0, err
	}
	if err := writeFileSafe(targetDir, path, []byte(out)); err != nil {
		return 0, fmt.Errorf("gravar a Constituição: %w", err)
	}
	return len(rotulos), nil
}

// PlanoCamposDeModo devolve os rótulos de modo que AcrescentaCamposDeModo acrescentaria, sem gravar
// nada (ARQ-05). Mesma decisão do update real (DUP-06).
func PlanoCamposDeModo(targetDir string, warn io.Writer) ([]string, error) {
	_, _, rotulos, err := calculaCamposDeModo(targetDir, warn)
	return rotulos, err
}

// calculaCamposDeModo decide o que acrescentar e devolve o novo texto da Constituição, sem gravar.
func calculaCamposDeModo(targetDir string, warn io.Writer) (path, out string, rotulos []string, err error) {
	path = filepath.Join(targetDir, "sdd", "memory", "constitution.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, "", nil, nil
		}
		return path, "", nil, fmt.Errorf("ler a Constituição: %w", err)
	}
	texto := string(data)
	linhas := strings.Split(texto, "\n")

	ini := -1
	for i, l := range linhas {
		if strings.TrimRight(l, " \t\r") == secaoFerramentas {
			ini = i
			break
		}
	}

	if ini < 0 {
		var b strings.Builder
		b.WriteString(strings.TrimRight(texto, "\n"))
		b.WriteString("\n\n" + secaoFerramentas + "\n| Campo | Valor |\n|-------|-------|\n")
		for _, c := range vcs.CamposDeModoPadrao() {
			fmt.Fprintf(&b, "| %s | %s |\n", c.Rotulo, c.Valor)
			rotulos = append(rotulos, c.Rotulo)
		}
		out = b.String()
	} else {
		fim := len(linhas)
		for i := ini + 1; i < len(linhas); i++ {
			if strings.HasPrefix(linhas[i], "## ") {
				fim = i
				break
			}
		}
		sep, ultima := -1, -1
		existentes := map[string]bool{}
		for i := ini + 1; i < fim; i++ {
			l := strings.TrimRight(linhas[i], "\r")
			if sep < 0 {
				if separadorRe.MatchString(l) && i > ini+1 && strings.HasPrefix(linhas[i-1], "|") {
					sep, ultima = i, i
				}
				continue
			}
			if !strings.HasPrefix(l, "|") {
				break
			}
			ultima = i
			if m := linhaTabelaRe.FindStringSubmatch(l); m != nil {
				existentes[m[1]] = true
			}
		}
		if sep < 0 {
			fmt.Fprintf(warn, "aviso: %q em sdd/memory/constitution.md não tem tabela legível; campos de modo não acrescentados (padrões seguros seguem valendo)\n", secaoFerramentas)
			return path, "", nil, nil
		}
		var add []string
		for _, c := range vcs.CamposDeModoPadrao() {
			if !existentes[c.Rotulo] {
				add = append(add, fmt.Sprintf("| %s | %s |", c.Rotulo, c.Valor))
				rotulos = append(rotulos, c.Rotulo)
			}
		}
		if len(add) == 0 {
			return path, "", nil, nil
		}
		res := append([]string{}, linhas[:ultima+1]...)
		res = append(res, add...)
		res = append(res, linhas[ultima+1:]...)
		out = strings.Join(res, "\n")
	}
	return path, out, rotulos, nil
}
