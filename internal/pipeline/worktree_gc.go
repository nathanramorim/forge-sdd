package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// StaleWorktree representa um worktree encontrado em disco com o estado do job associado.
type StaleWorktree struct {
	JobID      string
	TaskID     string
	Path       string
	JobState   string // estado no barramento; "" se orphan
	TaskStatus string // status da task em queue.json; "" se não encontrada
	Safe       bool
	Reason     string
}

// gcNormalizeTaskID remove espaços e o prefixo "feat-" de um ID de task (mesma regra da fila).
func gcNormalizeTaskID(id string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(id), "feat-"))
}

// gcMatchesTask casa o ID de uma task do worktree com o da fila, tolerando o prefixo "feat-" e
// sufixos de nome (ex.: "88-01" casa com "feat-88-01-nome").
func gcMatchesTask(itemID, id string) bool {
	a, b := gcNormalizeTaskID(itemID), gcNormalizeTaskID(id)
	if b == "" {
		return false
	}
	return a == b || strings.HasPrefix(a, b+"-")
}

// ListStaleWorktrees varre .forge/runtime/worktrees/<jobID>/<taskID> e classifica cada worktree.
func ListStaleWorktrees(dir string) ([]StaleWorktree, error) {
	wtRoot := layout.Worktrees(dir)
	jobDirs, err := os.ReadDir(wtRoot)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []StaleWorktree
	for _, jobEntry := range jobDirs {
		if !jobEntry.IsDir() {
			continue
		}
		jobID := jobEntry.Name()
		taskDirs, err := os.ReadDir(filepath.Join(wtRoot, jobID))
		if err != nil {
			continue
		}
		for _, taskEntry := range taskDirs {
			if !taskEntry.IsDir() {
				continue
			}
			taskID := taskEntry.Name()
			path := filepath.Join(wtRoot, jobID, taskID)

			wt := StaleWorktree{
				JobID:  jobID,
				TaskID: taskID,
				Path:   path,
			}

			// Busca job no barramento: tenta builder primeiro, depois reviewer.
			entry, err := bus.Find(dir, layout.RoleBuilder, jobID)
			if errors.Is(err, bus.ErrNotFound) {
				entry, err = bus.Find(dir, layout.RoleReviewer, jobID)
			}

			if err != nil {
				if errors.Is(err, bus.ErrNotFound) {
					// Nenhum dos papéis conhece o job: orphan seguro.
					wt.Safe = true
					wt.Reason = "job não encontrado no barramento (orphan)"
				} else {
					// Erro de leitura: conservador, não remove.
					wt.Safe = false
					wt.Reason = fmt.Sprintf("erro ao consultar o barramento: %v", err)
				}
				result = append(result, wt)
				continue
			}

			wt.JobState = entry.State
			switch entry.State {
			case layout.StatePending:
				wt.Safe = false
				wt.Reason = "job aguardando execução"
			case layout.StateProcessing:
				wt.Safe = false
				wt.Reason = "job em execução"
			case layout.StateFailed:
				wt.Safe = false
				wt.Reason = "job falhou; pode ser retentado"
			case layout.StateDone:
				q, err := queue.LoadForInitiative(dir, entry.Job.Initiative)
				if err != nil || q == nil {
					wt.Safe = false
					wt.Reason = "não foi possível verificar a fila da iniciativa"
					result = append(result, wt)
					continue
				}
				found := false
				for _, item := range q.Items {
					if gcMatchesTask(item.TaskID, taskID) {
						found = true
						wt.TaskStatus = item.Status
						if item.Status == queue.StatusApproved {
							wt.Safe = true
							wt.Reason = "job concluído e aprovado"
						} else {
							wt.Safe = false
							wt.Reason = fmt.Sprintf("task não aprovada (status: %s)", item.Status)
						}
						break
					}
				}
				if !found {
					wt.Safe = false
					wt.Reason = "task não encontrada na fila (tratar como ativa)"
				}
			default:
				wt.Safe = false
				wt.Reason = fmt.Sprintf("estado desconhecido: %s", entry.State)
			}
			result = append(result, wt)
		}
	}
	return result, nil
}

// SweepWorktrees remove os worktrees com Safe=true e executa git worktree prune ao final.
func SweepWorktrees(dir string, targets []StaleWorktree) []error {
	var errs []error
	removed := 0
	for _, t := range targets {
		if !t.Safe {
			continue
		}
		if err := RemoveWorktree(dir, t.JobID, t.TaskID); err != nil {
			errs = append(errs, fmt.Errorf("remover %s/%s: %w", t.JobID, t.TaskID, err))
			continue
		}
		removed++
	}
	if removed > 0 {
		if _, err := layout.Git(dir, "worktree", "prune"); err != nil {
			errs = append(errs, fmt.Errorf("git worktree prune: %w", err))
		}
	}
	return errs
}
