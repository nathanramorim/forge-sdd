package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
)

// taskJobRefs são as referências próprias de uma task (RN-2): só as rules e ADRs que ela cita.
type taskJobRefs struct {
	Spec  string // caminho relativo da spec da task
	Rules []string
	ADRs  []string
}

// PlanTaskID converte o id curto do job ("122-01") no id do plano e do leque ("feat-122-01").
func PlanTaskID(taskID string) string { return "feat-" + taskID }

// TaskBranch é a branch interna da task: forge/<iniciativa>-<task do plano>, a mesma do leque.
func TaskBranch(initiative, taskID string) string {
	return "forge/" + initiative + "-" + PlanTaskID(taskID)
}

// taskJobID compõe o id do job da task a partir do pai, dentro do limite de 80 caracteres.
func taskJobID(parent, taskID string) string {
	suf := "-" + strings.ToLower(taskID)
	if len(parent)+len(suf) > 80 {
		parent = parent[:80-len(suf)]
	}
	return strings.TrimRight(parent, "-.") + suf
}

// TaskJobFrom deriva o job task_build de uma task a partir do job pai (feature_build). O pacote é
// o mínimo da task: a spec dela e as referências dela; o PR da task aponta da branch da task para a
// branch da feature (ADR-0029).
func TaskJobFrom(parent bus.Job, taskID, title string, refs taskJobRefs) bus.Job {
	j := parent
	j.ID = taskJobID(parent.ID, taskID)
	j.Type = bus.TypeTaskBuild
	j.TaskID = taskID
	j.FeatureJob = parent.ID
	j.Title = title
	j.SpecPath = refs.Spec
	j.ADRRefs = append([]string{}, refs.ADRs...)
	j.RulesRefs = append([]string{}, refs.Rules...)
	j.PR = bus.PR{Branch: TaskBranch(parent.Initiative, taskID), BaseBranch: layout.InitiativeBranch(parent.Initiative)}
	j.Attempt = 1
	j.ReworkFeedback = []bus.Feedback{}
	j.Lease, j.Events, j.CreatedAt = nil, nil, time.Time{}
	return j
}

// ElegiveisDoPlano devolve, na ordem do plano, as tasks que podem começar: pendentes cujas
// dependências estão todas mescladas (PodeComecar). estados é indexado pelo id do plano
// ("feat-122-01"); task sem estado vale pendente. Tasks já em andamento, mescladas ou bloqueadas
// não voltam.
func ElegiveisDoPlano(batches [][]Task, estados map[string]string) []Task {
	quadro := map[string]string{}
	for _, b := range batches {
		for _, t := range b {
			if e, ok := estados[t.ID]; ok {
				quadro[t.ID] = e
			} else {
				quadro[t.ID] = TaskPendente
			}
		}
	}
	var out []Task
	for _, b := range batches {
		for _, t := range b {
			if quadro[t.ID] != TaskPendente {
				continue
			}
			if ok, err := PodeComecar(t, quadro); err == nil && ok {
				out = append(out, t)
			}
		}
	}
	return out
}

// planoDeTasks lê as ondas do plano da iniciativa (PlanWaves) com a largura máxima das ondas.
func planoDeTasks(dir, initiative string, largura int) ([][]Task, error) {
	return PlanWaves(filepath.Join(layout.Discovery(dir), initiative, "features"), largura)
}

// idCurtoDaTask converte "feat-122-01" em "122-01".
func idCurtoDaTask(planID string) string { return strings.TrimPrefix(planID, "feat-") }

// EnfileiraTasks enfileira um task_build por task elegível (RN-1, RN-3). Os jobs vêm de TaskJobFrom
// já com as referências por task. É idempotente: job existente não duplica.
func EnfileiraTasks(dir string, jobs []bus.Job, now time.Time) (criados []string, paths []string, err error) {
	for _, j := range jobs {
		p, created, eerr := bus.Enqueue(dir, layout.RoleBuilder, j, now)
		if eerr != nil {
			return criados, paths, fmt.Errorf("enfileirar %s: %w", j.TaskID, eerr)
		}
		paths = append(paths, p)
		if created {
			criados = append(criados, j.ID)
		}
	}
	return criados, paths, nil
}

// DescartarJobsSubstituidos tira da fila do Builder os jobs pendentes da mesma iniciativa que um
// despacho novo substitui: feature_build e task_build de despachos anteriores (inclusive os da 5.x,
// que um worker novo pegaria antes das tasks). Fixes (branch fix/) e os jobs do despacho atual
// (manter) ficam. Devolve os ids descartados.
func DescartarJobsSubstituidos(dir, initiative string, manter map[string]bool, novo string, now time.Time) ([]string, error) {
	pend, err := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	if err != nil {
		return nil, err
	}
	var fora []string
	for _, e := range pend {
		j := e.Job
		if j.Initiative != initiative || manter[j.ID] || strings.HasPrefix(j.PR.Branch, "fix/") {
			continue
		}
		if j.Type != bus.TypeFeatureBuild && j.Type != bus.TypeTaskBuild {
			continue
		}
		if _, derr := bus.Descartar(dir, layout.RoleBuilder, j.ID, "dispatch", "substituído pelo despacho "+novo, now); derr != nil {
			return fora, derr
		}
		fora = append(fora, j.ID)
	}
	return fora, nil
}
