package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// prLookPath é injetável nos testes (TST-04).
var prLookPath = exec.LookPath

// prRunCmd executa name com args em dir e devolve stdout; injetável nos testes (TST-04).
// Apenas "git" e "gh" são permitidos (SEG-03).
var prRunCmd = func(dir, name string, args ...string) (string, error) {
	if name != "git" && name != "gh" {
		return "", fmt.Errorf("comando %q não permitido", name)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(stdout.String()), fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// OpenPR faz git push origin <branch> + gh pr create após veredito conforme.
// Falha graciosamente: se gh não está instalado, branch está vazia ou origin não existe,
// imprime um aviso em out e retorna nil.
func OpenPR(dir string, q *Queue, item *QueueItem, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	// 1. Verifica gh
	if _, err := prLookPath("gh"); err != nil {
		fmt.Fprintln(out, "aviso: gh não encontrado; PR não criado automaticamente")
		return nil
	}

	// 2. Verifica branch
	branch := q.Branch
	if branch == "" {
		fmt.Fprintln(out, "aviso: branch não definida na fila; PR não criado automaticamente")
		return nil
	}

	// 3. Verifica se origin existe
	if _, err := prRunCmd(dir, "git", "config", "--get", "remote.origin.url"); err != nil {
		fmt.Fprintln(out, "aviso: remote origin não encontrado; PR não criado automaticamente")
		return nil
	}

	// 4. Verifica se PR já existe
	listOut, err := prRunCmd(dir, "gh", "pr", "list",
		"--head", branch,
		"--state", "open",
		"--json", "url",
		"--limit", "1",
	)
	if err != nil {
		fmt.Fprintf(out, "aviso: falha ao verificar PR existente: %v; PR não criado automaticamente\n", err)
		return nil
	}
	var prList []struct {
		URL string `json:"url"`
	}
	if strings.TrimSpace(listOut) != "" {
		if jsonErr := json.Unmarshal([]byte(listOut), &prList); jsonErr == nil && len(prList) > 0 {
			fmt.Fprintf(out, "PR já existe: %s\n", prList[0].URL)
			return nil
		}
	}

	// 5. Extrai título do PR da spec (DUP-01: usa extractSpecTitle do pacote)
	title := extractSpecTitle(dir, item.SpecPath, item.Title)

	// 6. Monta corpo do PR com o resumo do veredito
	resumo := ""
	if item.GovernanceVerdict != nil {
		resumo = item.GovernanceVerdict.Resumo
	}

	// 7. Resolve base branch
	baseBranch := q.BaseBranch
	if baseBranch == "" {
		baseBranch = "main"
	}

	// 8. git push
	if _, err := prRunCmd(dir, "git", "push", "origin", branch); err != nil {
		return fmt.Errorf("git push falhou: %w", err)
	}

	// 9. gh pr create
	prOut, err := prRunCmd(dir, "gh", "pr", "create",
		"--draft",
		"--title", title,
		"--body", resumo,
		"--base", baseBranch,
		"--head", branch,
	)
	if err != nil {
		return fmt.Errorf("gh pr create falhou: %w", err)
	}

	// 10. Captura URL do PR (última linha não vazia da saída)
	url := lastNonEmptyLine(prOut)
	fmt.Fprintf(out, "→ PR aberto: %s\n", url)
	return nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			return l
		}
	}
	return strings.TrimSpace(s)
}
