package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

// Erros da espera e da continuação por onda (feat-106-24c).
var (
	// ErrDependenteSemPredecessora: a task depende de outra que ainda não está mesclada.
	ErrDependenteSemPredecessora = errors.New("a task não começa antes da predecessora estar mesclada")
	// ErrOndaIncompleta: a onda ainda tem task em andamento; a próxima espera.
	ErrOndaIncompleta = errors.New("a onda só fecha quando todas as tasks dela terminam")
)

// Estados de uma task do ponto de vista da onda. `TaskPendente` já existe em laco.go (mesmo
// valor); os demais são do quadro da onda, que é anterior ao laço do modo chat.
const (
	// TaskMesclada: a entrega entrou na principal; a dependente pode começar.
	TaskMesclada = "mesclada"
	// OndaEmAndamento: o PR existe e está sendo trabalhado.
	OndaEmAndamento = "em-andamento"
	// OndaBloqueada: parou por conflito ou dependência quebrada.
	OndaBloqueada = "bloqueada"
)

// EstadoTaskNaOnda é o estado de uma task para a decisão de continuar.
type EstadoTaskNaOnda struct {
	ID     string
	Estado string
}

// validarEstado recusa estado desconhecido em vez de assumir que está tudo bem.
func validarEstadoDaOnda(e string) error {
	switch e {
	case TaskMesclada, OndaEmAndamento, TaskPendente, OndaBloqueada:
		return nil
	}
	return fmt.Errorf("estado %q desconhecido para a task", e)
}

// PodeComecar diz se a task pode começar, olhando as dependências dela. Uma dependência
// **mesclada** libera; qualquer outra coisa segura, porque o código que ela precisa ainda não
// está na principal.
func PodeComecar(task Task, estados map[string]string) (bool, error) {
	for _, dep := range task.DependsOn {
		estado, ok := estados[dep]
		if !ok {
			// Dependência fora desta onda: assume pendente, não libera.
			return false, fmt.Errorf("%w: %s depende de %s, que não está no quadro da onda",
				ErrDependenteSemPredecessora, task.ID, dep)
		}
		if err := validarEstadoDaOnda(estado); err != nil {
			return false, fmt.Errorf("%s: %w", dep, err)
		}
		if estado != TaskMesclada {
			return false, fmt.Errorf("%w: %s espera %s (estado %q)",
				ErrDependenteSemPredecessora, task.ID, dep, estado)
		}
	}
	return true, nil
}

// FimDaOnda diz se a onda acabou: todas as tasks estão mescladas ou bloqueadas (bloqueada não
// impede o fechamento, mas precisa de decisão humana — nunca é tratada como sucesso).
func FimDaOnda(tasks []EstadoTaskNaOnda) (bool, error) {
	if len(tasks) == 0 {
		return false, errors.New("onda vazia")
	}
	pendentes := 0
	for _, t := range tasks {
		if err := validarEstadoDaOnda(t.Estado); err != nil {
			return false, fmt.Errorf("%s: %w", t.ID, err)
		}
		if t.Estado == OndaEmAndamento || t.Estado == TaskPendente {
			pendentes++
		}
	}
	return pendentes == 0, nil
}

// ResumoDaOnda é o que o `[Continuar]` apresenta: quantas mesclaram, quantas bloqueiam e o que
// ainda pode começar.
type ResumoDaOnda struct {
	Mescladas   []string
	Bloqueadas  []string
	PodeComecar []string
	Aguardando  []string // dependentes que esperam predecessora
}

// ContinuarPorOnda agrupa a continuação: a onda seguinte só começa com a anterior fechada, e cada
// task só começa com as dependências mescladas. O agrupamento é o que o humano aprova com
// `[Continuar]` de uma vez (RN-3, Decisão xix) — não task por task.
func ContinuarPorOnda(plano [][]Task, estados map[string]string) (ResumoDaOnda, bool, error) {
	var res ResumoDaOnda
	todasFechadas := true
	ondaAtualFechada := true // a primeira onda é a "atual" até ela mesma fechar
	liberada := true         // só as ondas cujas anteriores fecharam podem começar

	for _, onda := range plano {
		estadosDaOnda := make([]EstadoTaskNaOnda, 0, len(onda))
		for _, t := range onda {
			estado := estados[t.ID]
			if estado == "" {
				estado = TaskPendente
			}
			estadosDaOnda = append(estadosDaOnda, EstadoTaskNaOnda{ID: t.ID, Estado: estado})
		}

		fim, err := FimDaOnda(estadosDaOnda)
		if err != nil {
			return res, false, err
		}
		if !fim {
			todasFechadas = false
		}

		for _, t := range onda {
			estado := estados[t.ID]
			switch estado {
			case TaskMesclada:
				res.Mescladas = append(res.Mescladas, t.ID)
				continue
			case OndaBloqueada:
				res.Bloqueadas = append(res.Bloqueadas, t.ID)
				continue
			}
			// Pendente ou em andamento: decide se pode começar.
			if estado == OndaEmAndamento {
				// Já começou; não entra em "pode começar" nem em "aguardando".
				continue
			}
			if !liberada || !ondaAtualFechada {
				// Onda anterior aberta: esta espera, sem avaliar dependência.
				res.Aguardando = append(res.Aguardando, t.ID)
				continue
			}
			ok, err := PodeComecar(t, estados)
			if err != nil || !ok {
				res.Aguardando = append(res.Aguardando, t.ID)
				continue
			}
			res.PodeComecar = append(res.PodeComecar, t.ID)
		}

		// Depois de avaliar esta onda, a próxima só é liberada se esta fechou.
		liberada = liberada && fim
		ondaAtualFechada = fim
	}

	return res, todasFechadas, nil
}

// TextoDoContinuar é a pergunta do `[Continuar]`, agrupada por onda: o humano decide pela onda
// inteira, não por task. Bloqueadas aparecem nomeadas, para não passarem por aprovadas.
func TextoDoContinuar(res ResumoDaOnda, onda int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Onda %d pronta para continuar.\n", onda)
	if len(res.Mescladas) > 0 {
		fmt.Fprintf(&b, "  Mescladas (%d): %s\n", len(res.Mescladas), strings.Join(res.Mescladas, ", "))
	}
	if len(res.PodeComecar) > 0 {
		fmt.Fprintf(&b, "  Podem começar (%d): %s\n", len(res.PodeComecar), strings.Join(res.PodeComecar, ", "))
	}
	if len(res.Aguardando) > 0 {
		fmt.Fprintf(&b, "  Aguardando dependência (%d): %s\n", len(res.Aguardando), strings.Join(res.Aguardando, ", "))
	}
	if len(res.Bloqueadas) > 0 {
		fmt.Fprintf(&b, "  BLOQUEADAS (%d): %s — precisam de decisão sua\n",
			len(res.Bloqueadas), strings.Join(res.Bloqueadas, ", "))
	}
	b.WriteString("  [Continuar] segue com a onda; [Ajustar] volta ao plano.")
	return b.String()
}
