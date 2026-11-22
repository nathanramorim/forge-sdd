package vcs

import "fmt"

// nenhum: sem PR. A branch fica pronta; o merge é local (RN-25).
type nenhum struct{ run Runner }

func (n *nenhum) Nome() string { return "nenhum" }

func (n *nenhum) CriarPRDraft(_ string, p NovoPR) (PR, error) {
	return PR{}, fmt.Errorf("%w: a branch %q fica pronta; o merge é local", ErrSemPR, p.Branch)
}

func (n *nenhum) ProcurarPRAberto(string, string) (*PR, error) { return nil, nil }

func (n *nenhum) CorrigirBase(string, int, string) error { return ErrSemPR }

func (n *nenhum) MarcarPronto(string, int) error { return ErrSemPR }

func (n *nenhum) CriarIssue(string, NovaIssue) (Issue, error) { return Issue{}, ErrSemPR }

func (n *nenhum) ProcurarIssueAberta(string, string) (*Issue, error) { return nil, ErrSemPR }

// Mesclar faz o merge local de Branch na base, só se a base já é a branch atual (não troca de
// branch nem mexe na árvore de trabalho de outra).
func (n *nenhum) Mesclar(dir string, m Mescla) error {
	cur, err := n.run(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	if cur != m.Base {
		return fmt.Errorf("merge local recusado: a branch atual é %q e a base é %q; faça checkout da base", cur, m.Base)
	}
	_, err = n.run(dir, "git", "merge", "--no-ff", "--no-edit", m.Branch)
	return err
}
