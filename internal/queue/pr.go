package queue

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/forge-sdd/cli/internal/vcs"
)

// prLookPath é injetável nos testes (TST-04).
var prLookPath = exec.LookPath

// prRunCmd executa name com args em dir e devolve stdout; injetável nos testes (TST-04).
// O executor real e a lista git/gh (SEG-03) moram em internal/vcs.
var prRunCmd = vcs.RunGitOrGh

// OpenPR faz git push origin <branch> + gh pr create após veredito conforme.
// Falha graciosamente: se gh não está instalado, branch está vazia ou origin não existe,
// imprime um aviso em out e retorna nil.
func OpenPR(dir string, q *Queue, item *QueueItem, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	adapt, err := vcs.DaConstituicao(dir, func(d, n string, a ...string) (string, error) { return prRunCmd(d, n, a...) })
	if err != nil {
		return err
	}
	if adapt.Nome() == "nenhum" {
		fmt.Fprintf(out, "aviso: VCS nenhum; sem PR, a branch %q fica pronta e o merge é local\n", q.Branch)
		return nil
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
	existing, err := adapt.ProcurarPRAberto(dir, branch)
	if err != nil {
		fmt.Fprintf(out, "aviso: falha ao verificar PR existente: %v; PR não criado automaticamente\n", err)
		return nil
	}
	if existing != nil {
		fmt.Fprintf(out, "PR já existe: %s\n", existing.URL)
		return nil
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

	// 9. PR em draft pelo adaptador (RN-7)
	pr, err := adapt.CriarPRDraft(dir, vcs.NovoPR{Base: baseBranch, Branch: branch, Titulo: title, Corpo: resumo})
	if err != nil {
		return fmt.Errorf("gh pr create falhou: %w", err)
	}
	fmt.Fprintf(out, "→ PR aberto: %s\n", pr.URL)
	return nil
}
