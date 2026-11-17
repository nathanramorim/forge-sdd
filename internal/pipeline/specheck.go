package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/queue"
)

var fixFileRe = regexp.MustCompile(`^(fix-[a-z0-9]+-[0-9]{2}[a-z]?)-`)

// taskSpecProblems confere, por task (feat-*.md e fix-*.md da iniciativa), o que o build exige da
// spec: dependências que existem na iniciativa e critério de aceitação com comando executável.
// Devolve tudo de uma vez, ordenado, para o erro sair antes do build e sem gastar tokens.
func taskSpecProblems(dir, base string) []string {
	files := map[string]string{} // id -> caminho
	for _, sub := range []string{"features", "fixes"} {
		entries, err := os.ReadDir(filepath.Join(base, sub))
		if err != nil {
			continue
		}
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".md") {
				continue
			}
			if m := taskFileRe.FindStringSubmatch(n); m != nil {
				files[m[1]] = filepath.Join(base, sub, n)
			} else if m := fixFileRe.FindStringSubmatch(n); m != nil {
				files[m[1]] = filepath.Join(base, sub, n)
			}
		}
	}
	var problems []string
	for id, path := range files {
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: não foi possível ler", rel))
			continue
		}
		content := string(data)
		for _, dep := range queue.ParseDependsOn(content) {
			if !strings.HasPrefix(dep, "feat-") && !strings.HasPrefix(dep, "fix-") {
				dep = "feat-" + dep
			}
			if _, ok := files[dep]; !ok {
				problems = append(problems, fmt.Sprintf("%s: %s depende de %s, que não existe na iniciativa", rel, id, dep))
			}
		}
		if !hasExecutableCriterion(content) {
			problems = append(problems, fmt.Sprintf("%s: os critérios de aceitação não trazem um comando executável (bloco de código com o comando de teste)", rel))
		}
	}
	sort.Strings(problems)
	return problems
}

// hasExecutableCriterion exige, na seção de critérios, um bloco de código com ao menos uma linha de comando.
func hasExecutableCriterion(content string) bool {
	body := sectionBody(content, "critérios de aceitação executáveis")
	inFence := false
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence && line != "" && !strings.HasPrefix(line, "#") {
			return true
		}
	}
	return false
}
