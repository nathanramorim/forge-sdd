package bus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// Entry é um job encontrado em um estado da fila.
type Entry struct {
	State string
	Path  string
	Job   Job
}

var workerStatesInOrder = []string{layout.StatePending, layout.StateProcessing, layout.StateDone, layout.StateFailed}

// fileName gera <prio>-<timestamp>-<id>.json: "0" para high e "1" para normal, depois o instante
// em nanossegundos com largura fixa. A ordem lexicográfica é prioridade e depois FIFO (D-4).
func fileName(j *Job) string {
	prio := "1"
	if j.Priority == PriorityHigh {
		prio = "0"
	}
	return fmt.Sprintf("%s-%019d-%s.json", prio, j.CreatedAt.UTC().UnixNano(), j.ID)
}

func idFromName(name string) string {
	parts := strings.SplitN(strings.TrimSuffix(name, ".json"), "-", 3)
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}

func stateDir(dir, role, state string) (string, error) { return layout.QueueState(dir, role, state) }

func checkWorkerRole(role string) error {
	if role != layout.RoleBuilder && role != layout.RoleReviewer {
		return fmt.Errorf("%w: papel %q não tem fila de jobs (use builder ou reviewer)", layout.ErrInvalidQueue, role)
	}
	return nil
}

// listNames devolve os arquivos .json de um diretório em ordem lexicográfica; ignora temporários.
func listNames(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, ".") || !strings.HasSuffix(n, ".json") {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

// readJob lê e valida um job, recusando arquivos acima do limite.
func readJob(path string) (*Job, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s não é arquivo regular", ErrInvalidJob, filepath.Base(path))
	}
	if info.Size() > MaxJobBytes {
		return nil, fmt.Errorf("%w: %s tem %d bytes (limite %d)", ErrTooLarge, filepath.Base(path), info.Size(), MaxJobBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var j Job
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&j); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJob, err)
	}
	if err := j.Validate(""); err != nil {
		return nil, err
	}
	return &j, nil
}

func writeJob(path string, j *Job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	if len(data) > MaxJobBytes {
		return fmt.Errorf("%w: %d bytes (limite %d)", ErrTooLarge, len(data), MaxJobBytes)
	}
	return queue.WriteJSONAtomic(path, j, 0o644)
}

func (j *Job) addEvent(now time.Time, state, by, note string) {
	j.Events = append(j.Events, Event{At: now.UTC(), State: state, By: by, Note: truncate(note, 300)})
	if len(j.Events) > maxEvents {
		j.Events = j.Events[len(j.Events)-maxEvents:]
	}
	j.UpdatedAt = now.UTC()
}

func truncate(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Find localiza o job por id em qualquer estado do papel.
func Find(dir, role, id string) (*Entry, error) {
	if err := checkWorkerRole(role); err != nil {
		return nil, err
	}
	if !idRe.MatchString(id) {
		return nil, fmt.Errorf("%w: id %q inválido", ErrInvalidJob, id)
	}
	for _, st := range workerStatesInOrder {
		sd, err := stateDir(dir, role, st)
		if err != nil {
			return nil, err
		}
		names, err := listNames(sd)
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			if idFromName(n) != id {
				continue
			}
			p := filepath.Join(sd, n)
			j, err := readJob(p)
			if err != nil {
				return nil, err
			}
			return &Entry{State: st, Path: p, Job: *j}, nil
		}
	}
	return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, role, id)
}

// Enqueue grava o job em <papel>/pending. Duplicata por id (em qualquer estado) é idempotente:
// devolve o caminho existente com created=false e não altera nada.
func Enqueue(dir, role string, job Job, now time.Time) (path string, created bool, err error) {
	if err := checkWorkerRole(role); err != nil {
		return "", false, err
	}
	if job.SchemaVersion == 0 {
		job.SchemaVersion = SchemaVersion
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now.UTC()
	}
	job.UpdatedAt = now.UTC()
	if job.Attempt == 0 {
		job.Attempt = 1
	}
	if job.ADRRefs == nil {
		job.ADRRefs = []string{}
	}
	if job.RulesRefs == nil {
		job.RulesRefs = []string{}
	}
	if job.ReworkFeedback == nil {
		job.ReworkFeedback = []Feedback{}
	}
	job.Lease = nil
	if err := job.Validate(role); err != nil {
		return "", false, err
	}
	if existing, ferr := Find(dir, role, job.ID); ferr == nil {
		return existing.Path, false, nil
	} else if !errors.Is(ferr, ErrNotFound) {
		return "", false, ferr
	}
	job.addEvent(now, layout.StatePending, "", "enqueued")
	pd, err := stateDir(dir, role, layout.StatePending)
	if err != nil {
		return "", false, err
	}
	path = filepath.Join(pd, fileName(&job))
	if err := writeJob(path, &job); err != nil {
		return "", false, err
	}
	return path, true, nil
}

// Claim reivindica o próximo job (prioridade e depois FIFO) com um rename atômico de pending para
// processing e grava o lease. Só um worker vence a disputa; a fila vazia devolve ErrEmpty.
func Claim(dir, role, worker string, ttl time.Duration, now time.Time) (*Entry, error) {
	if err := checkWorkerRole(role); err != nil {
		return nil, err
	}
	if !ValidWorker(worker) {
		return nil, fmt.Errorf("%w: worker %q inválido", ErrInvalidJob, worker)
	}
	if ttl <= 0 {
		ttl = DefaultLeaseTTL
	}
	pd, _ := stateDir(dir, role, layout.StatePending)
	prd, _ := stateDir(dir, role, layout.StateProcessing)
	fd, _ := stateDir(dir, role, layout.StateFailed)

	names, err := listNames(pd)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if err := os.MkdirAll(prd, 0o755); err != nil {
			return nil, err
		}
		src := filepath.Join(pd, name)
		dst := filepath.Join(prd, name)
		if err := os.Rename(src, dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // outro worker venceu
			}
			return nil, fmt.Errorf("reivindicar %s: %w", name, err)
		}
		_ = os.Chtimes(dst, now, now) // o rename preserva o mtime antigo; a janela até o lease começa agora

		job, rerr := readJob(dst)
		if rerr != nil {
			if mkerr := os.MkdirAll(fd, 0o755); mkerr == nil {
				_ = os.Rename(dst, filepath.Join(fd, name))
				_ = os.WriteFile(filepath.Join(fd, name+".reason.txt"), []byte(truncate(rerr.Error(), 500)+"\n"), 0o644)
			}
			continue
		}
		job.Lease = &Lease{Worker: worker, ExpiresAt: now.UTC().Add(ttl)}
		job.addEvent(now, layout.StateProcessing, worker, "claimed")
		if err := writeJob(dst, job); err != nil {
			return nil, err
		}
		return &Entry{State: layout.StateProcessing, Path: dst, Job: *job}, nil
	}
	return nil, ErrEmpty
}

func ownEntry(dir, role, id, worker string) (*Entry, error) {
	e, err := Find(dir, role, id)
	if err != nil {
		return nil, err
	}
	if e.State != layout.StateProcessing {
		return nil, fmt.Errorf("%w: %s está em %s, não em processing", ErrNotFound, id, e.State)
	}
	if worker != "" && e.Job.Lease != nil && e.Job.Lease.Worker != worker {
		return nil, fmt.Errorf("%w: %s", ErrNotOwner, id)
	}
	return e, nil
}

func move(dir, role string, e *Entry, to, by, note string, now time.Time) (*Entry, error) {
	e.Job.Lease = nil
	e.Job.addEvent(now, to, by, note)
	if err := writeJob(e.Path, &e.Job); err != nil {
		return nil, err
	}
	td, err := stateDir(dir, role, to)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(td, 0o755); err != nil {
		return nil, err
	}
	dst := filepath.Join(td, filepath.Base(e.Path))
	if err := os.Rename(e.Path, dst); err != nil {
		return nil, fmt.Errorf("mover %s para %s: %w", e.Job.ID, to, err)
	}
	return &Entry{State: to, Path: dst, Job: e.Job}, nil
}

// Complete move o job de processing para done.
func Complete(dir, role, id, worker, note string, now time.Time) (*Entry, error) {
	e, err := ownEntry(dir, role, id, worker)
	if err != nil {
		return nil, err
	}
	return move(dir, role, e, layout.StateDone, worker, note, now)
}

// Fail move o job de processing para failed, registrando o motivo.
func Fail(dir, role, id, worker, reason string, now time.Time) (*Entry, error) {
	e, err := ownEntry(dir, role, id, worker)
	if err != nil {
		return nil, err
	}
	return move(dir, role, e, layout.StateFailed, worker, reason, now)
}

// Renew estende o lease do job reivindicado pelo worker.
func Renew(dir, role, id, worker string, ttl time.Duration, now time.Time) error {
	e, err := ownEntry(dir, role, id, worker)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = DefaultLeaseTTL
	}
	e.Job.Lease = &Lease{Worker: worker, ExpiresAt: now.UTC().Add(ttl)}
	e.Job.UpdatedAt = now.UTC()
	return writeJob(e.Path, &e.Job)
}

// Recover devolve a pending os jobs de processing com lease vencido (ou sem lease legível e parados
// além da carência) e devolve quantos foram recuperados.
func Recover(dir, role string, now time.Time) (int, error) {
	if err := checkWorkerRole(role); err != nil {
		return 0, err
	}
	prd, _ := stateDir(dir, role, layout.StateProcessing)
	pd, _ := stateDir(dir, role, layout.StatePending)
	names, err := listNames(prd)
	if err != nil {
		return 0, err
	}
	recovered := 0
	for _, name := range names {
		p := filepath.Join(prd, name)
		job, rerr := readJob(p)
		expired := false
		switch {
		case rerr != nil:
			if info, serr := os.Stat(p); serr == nil && now.Sub(info.ModTime()) > orphanLeaseGrace {
				fd, _ := stateDir(dir, role, layout.StateFailed)
				if os.MkdirAll(fd, 0o755) == nil && os.Rename(p, filepath.Join(fd, name)) == nil {
					_ = os.WriteFile(filepath.Join(fd, name+".reason.txt"), []byte(truncate(rerr.Error(), 500)+"\n"), 0o644)
				}
			}
			continue
		case job.Lease != nil:
			expired = now.After(job.Lease.ExpiresAt)
		default:
			if info, serr := os.Stat(p); serr == nil {
				expired = now.Sub(info.ModTime()) > orphanLeaseGrace
			}
		}
		if !expired {
			continue
		}
		by := ""
		if job.Lease != nil {
			by = job.Lease.Worker
		}
		job.Lease = nil
		job.addEvent(now, layout.StatePending, by, "lease expirado: devolvido à fila")
		if err := writeJob(p, job); err != nil {
			return recovered, err
		}
		if err := os.MkdirAll(pd, 0o755); err != nil {
			return recovered, err
		}
		if err := os.Rename(p, filepath.Join(pd, name)); err != nil {
			return recovered, fmt.Errorf("recuperar %s: %w", name, err)
		}
		recovered++
	}
	return recovered, nil
}

// List devolve os jobs de um estado, na ordem em que seriam consumidos.
func List(dir, role, state string) ([]Entry, error) {
	if err := checkWorkerRole(role); err != nil {
		return nil, err
	}
	sd, err := stateDir(dir, role, state)
	if err != nil {
		return nil, err
	}
	names, err := listNames(sd)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, n := range names {
		p := filepath.Join(sd, n)
		j, err := readJob(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", n, err)
		}
		out = append(out, Entry{State: state, Path: p, Job: *j})
	}
	return out, nil
}
