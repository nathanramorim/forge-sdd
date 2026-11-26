package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/bus"
)

// copiarJobParaWorktree põe o arquivo do job no mesmo caminho relativo dentro do worktree (.forge é
// ignorado pelo git), para o Builder ler o job sem sair da própria pasta.
func copiarJobParaWorktree(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

func mustRel(base, p string) string {
	if r, err := filepath.Rel(base, p); err == nil {
		return r
	}
	return p
}

// SecaoRetomada é o handoff compacto de retomada (RN-17): a task continua de onde parou, sem
// reler a conversa anterior. Vazia quando o worktree é novo.
func SecaoRetomada(job bus.Job, wt *WorktreeDaTask) string {
	if wt == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(chaveValor("checkpoint", fmt.Sprintf("faça commits wip(%s): <passo> a cada passo que passa nos testes; o último commit fica sem prefixo wip", job.TaskID)))
	if len(wt.Retomada) == 0 && !wt.Reaberto {
		return "\n" + secaoCompacta("checkpoint", b.String())
	}
	b.WriteString(chaveValor("modo", "retomada: a execução anterior desta task parou no meio"))
	b.WriteString(chaveValor("branch", job.PR.Branch))
	if ult, err := gitLimpo(wt.Path, "log", "-1", "--format=%h %s"); err == nil && ult != "" {
		b.WriteString(chaveValor("ultimo_commit", higieniza(ult, 160, false)))
	}
	if len(wt.Retomada) > 0 {
		arqs := wt.Retomada
		if len(arqs) > 30 {
			arqs = append(arqs[:30:30], fmt.Sprintf("(+%d)", len(wt.Retomada)-30))
		}
		b.WriteString(chaveValor("arquivos_sujos", strings.Join(arqs, ", ")))
		b.WriteString("acao: confira os arquivos sujos, rode os testes e continue; não descarte o que já foi feito\n")
	} else {
		b.WriteString("acao: o worktree foi recriado da branch da task; continue a partir do último commit\n")
	}
	return "\n" + secaoCompacta("retomada", b.String())
}

// encerrarTask fecha o ciclo efêmero depois do Builder de uma task: worktree limpo é apagado (a
// branch fica), sujo fica para a retomada; falha marca a task bloqueada; sucesso roda o despachante.
func encerrarTask(opts WorkerOptions, job bus.Job, reason string, live *LiveView) {
	apagado, err := EncerrarWorktreeDaTask(opts.Dir, job, opts.Name)
	switch {
	case err != nil:
		live.Line("⚠ worktree da task não foi encerrado: " + higieniza(err.Error(), 200, false))
	case !apagado:
		live.Line("⚠ worktree da task ficou com mudanças não commitadas; a próxima execução retoma dele")
	}
	if reason != "" {
		_ = GravarEstadoTask(opts.Dir, job.Initiative, PlanTaskID(job.TaskID), OndaBloqueada, reason, "", opts.Now())
		return
	}
	if r, derr := DespacharIniciativa(opts.Dir, job.Initiative, opts.Now()); derr == nil && r != nil && len(r.Jobs) > 0 {
		live.Line(fmt.Sprintf("▶ despachante: %d task(s) nova(s) na fila (teto %d)", len(r.Jobs), r.Teto.Pares))
	}
}
