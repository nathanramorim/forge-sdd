package queue

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/vcs"
)

// Erros da etapa de PR. Todos são recusas: nada aqui aprova em silêncio.
var (
	// ErrSemVereditoConforme indica que a task não tem parecer `conforme` gravado.
	ErrSemVereditoConforme = errors.New("task sem parecer conforme do Reviewer")
	// ErrSHAVencido indica que o parecer foi dado para outro commit: o código mudou depois.
	ErrSHAVencido = errors.New("o parecer é de outro commit; o HEAD mudou depois da revisão")
)

// autorizaPronto confere se o item pode ter o PR marcado como pronto. Exige um parecer
// `conforme` com `head_sha` igual ao HEAD atual da branch: um veredito velho não vale (FLX-06).
// `headAtual` é injetado para o teste não depender do git.
func autorizaPronto(item *QueueItem, headAtual string) error {
	if item == nil {
		return fmt.Errorf("task não encontrada")
	}
	v := item.GovernanceVerdict
	if v == nil {
		return fmt.Errorf("%s: %w", item.TaskID, ErrSemVereditoConforme)
	}
	if v.Veredito != harness.VereditoConforme {
		return fmt.Errorf("%s: veredito %q não autoriza marcar pronto; é preciso %q",
			item.TaskID, v.Veredito, harness.VereditoConforme)
	}
	if err := harness.ValidarProveniencia(*v); err != nil {
		return fmt.Errorf("%s: %w", item.TaskID, err)
	}
	atual := normalizarSHA(headAtual)
	if atual == "" {
		return fmt.Errorf("%s: não consegui ler o HEAD da branch para comparar com o parecer", item.TaskID)
	}
	if normalizarSHA(v.HeadSHA) != atual {
		return fmt.Errorf("%s: %w (parecer em %s, HEAD em %s)",
			item.TaskID, ErrSHAVencido, curto(v.HeadSHA), curto(atual))
	}
	return nil
}

// normalizarSHA deixa o SHA em minúsculas e sem espaços; a comparação é literal.
func normalizarSHA(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// curto mostra os 7 primeiros caracteres do SHA nas mensagens.
func curto(s string) string {
	s = normalizarSHA(s)
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

// MarcarPronto tira o PR do draft pelo adaptador, só depois do parecer conforme do HEAD atual.
// No VCS `nenhum` devolve ErrSemPR (a branch já está pronta).
func MarcarPronto(dir string, item *QueueItem, adapt vcs.Adaptador, numero int, headAtual string) error {
	if err := autorizaPronto(item, headAtual); err != nil {
		return err
	}
	if adapt == nil {
		return fmt.Errorf("adaptador de VCS ausente")
	}
	if adapt.Nome() == "nenhum" {
		return vcs.ErrSemPR
	}
	if numero <= 0 {
		return fmt.Errorf("%s: número do PR inválido (%d)", item.TaskID, numero)
	}
	return adapt.MarcarPronto(dir, numero)
}

// MesclarPR mescla o PR pelo adaptador, com as mesmas exigências do MarcarPronto: parecer
// conforme e SHA do HEAD atual. Quem decide é o humano; esta função só confere a prova antes.
func MesclarPR(dir string, item *QueueItem, adapt vcs.Adaptador, m vcs.Mescla, headAtual string) error {
	if err := autorizaPronto(item, headAtual); err != nil {
		return err
	}
	if adapt == nil {
		return fmt.Errorf("adaptador de VCS ausente")
	}
	return adapt.Mesclar(dir, m)
}

// ResumoParaPR devolve a linha que o PR leva: o resumo do parecer com o commit revisado, para
// quem lê o PR saber de que revisão se trata.
func ResumoParaPR(item *QueueItem) string {
	if item == nil || item.GovernanceVerdict == nil {
		return ""
	}
	v := item.GovernanceVerdict
	resumo := v.Resumo
	if v.HeadSHA != "" {
		if resumo != "" {
			return fmt.Sprintf("%s (revisado em %s)", resumo, curto(v.HeadSHA))
		}
		return fmt.Sprintf("Revisado em %s", curto(v.HeadSHA))
	}
	return resumo
}

// AvisoSemPR explica, em texto, o que fazer quando o VCS não tem PR.
func AvisoSemPR(out io.Writer, branch string) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "aviso: VCS nenhum; sem PR, a branch %q fica pronta e o merge é local\n", branch)
}
