package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
)

// matchesTask segue a mesma regra de Review: id exato ou prefixo "<id>-".
func matchesTask(itemID, id string) bool {
	return itemID == id || strings.HasPrefix(itemID, id+"-")
}

// SetTaskStatuses atualiza o status das tasks informadas na fila de uma iniciativa específica,
// sem depender da branch atual (a iniciativa ativa vem da branch, e o Reviewer ou o Builder podem
// estar em outra). Tasks já aprovadas nunca regridem. Fila ausente ou tasks desconhecidas não são
// erro: devolve quantas tasks mudaram de status.
func SetTaskStatuses(dir, initiative string, taskIDs []string, to string) (int, error) {
	if len(taskIDs) == 0 || strings.TrimSpace(initiative) == "" {
		return 0, nil
	}
	base, err := layout.Initiative(dir, initiative)
	if err != nil {
		if errors.Is(err, layout.ErrInvalidInitiative) {
			return 0, nil // job sem iniciativa válida: não há fila a atualizar
		}
		return 0, err
	}
	path := filepath.Join(base, layout.QueueFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	var q Queue
	if err := json.Unmarshal(data, &q); err != nil {
		return 0, fmt.Errorf("falha ao decodificar %s: %w", path, err)
	}

	changed := 0
	for i := range q.Items {
		for _, id := range taskIDs {
			if !matchesTask(q.Items[i].TaskID, id) {
				continue
			}
			if q.Items[i].Status != to && q.Items[i].Status != StatusApproved {
				q.Items[i].Status = to
				changed++
			}
			break
		}
	}
	if changed == 0 {
		return 0, nil
	}
	q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := WriteJSONAtomic(path, &q, 0o644); err != nil {
		return 0, fmt.Errorf("falha ao gravar fila %s: %w", path, err)
	}
	return changed, nil
}

// BlockTasks marca as tasks como blocked com o motivo (feat 88-06). Tasks já aprovadas não
// regridem. Tasks dependentes ficam paradas porque a dependência não chega a approved; as
// independentes continuam elegíveis. Devolve quantas tasks mudaram.
func BlockTasks(dir, initiative string, taskIDs []string, reason string) (int, error) {
	return mutateTasks(dir, initiative, taskIDs, func(it *QueueItem) bool {
		if it.Status == StatusApproved || (it.Status == StatusBlocked && it.BlockedReason == reason) {
			return false
		}
		it.Status = StatusBlocked
		it.BlockedReason = reason
		return true
	})
}

// mutateTasks aplica fn às tasks citadas e grava a fila de forma atômica quando algo mudou.
func mutateTasks(dir, initiative string, taskIDs []string, fn func(*QueueItem) bool) (int, error) {
	if len(taskIDs) == 0 || strings.TrimSpace(initiative) == "" {
		return 0, nil
	}
	base, err := layout.Initiative(dir, initiative)
	if err != nil {
		if errors.Is(err, layout.ErrInvalidInitiative) {
			return 0, nil
		}
		return 0, err
	}
	path := filepath.Join(base, layout.QueueFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	var q Queue
	if err := json.Unmarshal(data, &q); err != nil {
		return 0, fmt.Errorf("falha ao decodificar %s: %w", path, err)
	}
	changed := 0
	for i := range q.Items {
		for _, id := range taskIDs {
			if matchesTask(q.Items[i].TaskID, id) {
				if fn(&q.Items[i]) {
					changed++
				}
				break
			}
		}
	}
	if changed == 0 {
		return 0, nil
	}
	q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := WriteJSONAtomic(path, &q, 0o644); err != nil {
		return 0, fmt.Errorf("falha ao gravar fila %s: %w", path, err)
	}
	return changed, nil
}
