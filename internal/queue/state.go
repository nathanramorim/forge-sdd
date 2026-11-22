package queue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
)

// Estado de execução por feature (task 0.7a da 106, Decisão xvi): cada feature grava só o seu
// arquivo sdd/discovery/<id>/state/<sid>.json na branch da feature, então dois PRs nunca editam o
// mesmo arquivo. Sem a pasta state/ a fila segue como antes (queue.json guarda o status).
const (
	// StateDirName é a pasta de estados dentro da iniciativa.
	StateDirName = "state"
	// StateSchemaVersion é a versão do esquema do arquivo de estado.
	StateSchemaVersion = 1

	maxStateBytes     = 1 << 20
	maxBlockedReasonB = 2000
)

var (
	sidRe     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	headSHARe = regexp.MustCompile(`^[0-9a-f]{7,64}$`)
	allStatus = map[string]bool{
		StatusQueued: true, StatusBuilding: true, StatusReviewPending: true,
		StatusApproved: true, StatusChangesRequested: true, StatusBlocked: true,
	}
)

// FeatureState é o conteúdo de state/<sid>.json. Attempt é o contador de tentativas da fila
// (QueueItem.Attempts), não um contador novo (DUP-06).
type FeatureState struct {
	Schema            int      `json:"schema"`
	TaskID            string   `json:"task_id"`
	Status            string   `json:"status"`
	Attempt           int      `json:"attempt"`
	UpdatedAt         string   `json:"updated_at"`
	HeadSHA           string   `json:"head_sha,omitempty"`
	BlockedReason     string   `json:"blocked_reason,omitempty"`
	GovernanceVerdict *Verdict `json:"governance_verdict,omitempty"`
}

// StateID devolve o sid (nome do arquivo) de uma task: o ID normalizado, validado.
func StateID(taskID string) (string, error) {
	sid := NormalizeTaskID(taskID)
	if len(sid) > 120 || !sidRe.MatchString(sid) {
		return "", fmt.Errorf("id de task %q inválido para estado por feature", taskID)
	}
	return sid, nil
}

// StateDir devolve sdd/discovery/<id>/state (id validado, SEG-01).
func StateDir(dir, initiative string) (string, error) {
	base, err := layout.Initiative(dir, initiative)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, StateDirName), nil
}

// StatePath devolve sdd/discovery/<id>/state/<sid>.json.
func StatePath(dir, initiative, taskID string) (string, error) {
	sd, err := StateDir(dir, initiative)
	if err != nil {
		return "", err
	}
	sid, err := StateID(taskID)
	if err != nil {
		return "", err
	}
	return filepath.Join(sd, sid+".json"), nil
}

// Validate recusa campos inválidos: o arquivo é dado não confiável (SEG-02).
func (s FeatureState) Validate() error {
	if s.Schema != StateSchemaVersion {
		return fmt.Errorf("esquema %d não suportado (esperado %d)", s.Schema, StateSchemaVersion)
	}
	if _, err := StateID(s.TaskID); err != nil {
		return err
	}
	if !allStatus[s.Status] {
		return fmt.Errorf("status %q inválido", s.Status)
	}
	if s.Attempt < 0 {
		return fmt.Errorf("attempt %d negativo", s.Attempt)
	}
	if s.UpdatedAt != "" {
		if _, err := time.Parse(time.RFC3339, s.UpdatedAt); err != nil {
			return fmt.Errorf("updated_at inválido: %w", err)
		}
	}
	if s.HeadSHA != "" && !headSHARe.MatchString(s.HeadSHA) {
		return fmt.Errorf("head_sha %q inválido", s.HeadSHA)
	}
	if len(s.BlockedReason) > maxBlockedReasonB {
		return errors.New("blocked_reason grande demais")
	}
	return nil
}

// refuseSymlink recusa caminho que seja link simbólico (SEG-02). Ausente não é erro.
func refuseSymlink(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s é link simbólico: recusado", path)
	}
	return nil
}

// stateModeActive informa se a iniciativa usa estado por feature (pasta state/ real existe).
func stateModeActive(dir, initiative string) bool {
	sd, err := StateDir(dir, initiative)
	if err != nil {
		return false
	}
	fi, err := os.Lstat(sd)
	return err == nil && fi.IsDir()
}

// ReadFeatureState lê e valida o estado da task. Devolve nil sem erro quando não existe.
func ReadFeatureState(dir, initiative, taskID string) (*FeatureState, error) {
	path, err := StatePath(dir, initiative, taskID)
	if err != nil {
		return nil, err
	}
	if err := refuseSymlink(filepath.Dir(path)); err != nil {
		return nil, err
	}
	if err := refuseSymlink(path); err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxStateBytes+1))
	if err != nil {
		return nil, fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	if len(data) > maxStateBytes {
		return nil, fmt.Errorf("%s excede %d bytes", path, maxStateBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var st FeatureState
	if err := dec.Decode(&st); err != nil {
		return nil, fmt.Errorf("estado inválido em %s: %w", path, err)
	}
	if err := st.Validate(); err != nil {
		return nil, fmt.Errorf("estado inválido em %s: %w", path, err)
	}
	if sid, _ := StateID(taskID); st.TaskID != sid {
		return nil, fmt.Errorf("estado inválido em %s: task_id %q diverge do arquivo", path, st.TaskID)
	}
	return &st, nil
}

// WriteFeatureState grava o estado da task de forma atômica, só no arquivo dela.
func WriteFeatureState(dir, initiative string, st FeatureState) error {
	st.Schema = StateSchemaVersion
	st.TaskID = NormalizeTaskID(st.TaskID)
	st.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := st.Validate(); err != nil {
		return err
	}
	path, err := StatePath(dir, initiative, st.TaskID)
	if err != nil {
		return err
	}
	if err := refuseSymlink(filepath.Dir(path)); err != nil {
		return err
	}
	if err := refuseSymlink(path); err != nil {
		return err
	}
	if err := WriteJSONAtomic(path, &st, 0o644); err != nil {
		return fmt.Errorf("falha ao gravar estado %s: %w", path, err)
	}
	return nil
}

func stateFromItem(it QueueItem) FeatureState {
	return FeatureState{
		TaskID: it.TaskID, Status: it.Status, Attempt: it.Attempts,
		BlockedReason: it.BlockedReason, GovernanceVerdict: it.GovernanceVerdict,
	}
}

func applyState(it *QueueItem, st *FeatureState) {
	it.Status, it.Attempts = st.Status, st.Attempt
	it.BlockedReason, it.GovernanceVerdict = st.BlockedReason, st.GovernanceVerdict
}

// overlayStates sobrepõe aos itens do plano o estado das features que têm arquivo. Fila antiga
// (sem state/) passa intacta. Devolve os índices dos itens sobrepostos.
func overlayStates(dir, initiative string, q *Queue) (map[int]bool, error) {
	overlaid := map[int]bool{}
	if initiative == "" || !stateModeActive(dir, initiative) {
		return overlaid, nil
	}
	for i := range q.Items {
		if _, err := StateID(q.Items[i].TaskID); err != nil {
			continue
		}
		st, err := ReadFeatureState(dir, initiative, q.Items[i].TaskID)
		if err != nil {
			return nil, err
		}
		if st != nil {
			applyState(&q.Items[i], st)
			overlaid[i] = true
		}
	}
	return overlaid, nil
}

// initiativeOfQueuePath devolve o id da iniciativa dona de queuePath (vazio para a fila legada).
func initiativeOfQueuePath(dir, queuePath string) string {
	parent := filepath.Dir(queuePath)
	if filepath.Dir(parent) != layout.Discovery(dir) {
		return ""
	}
	id := filepath.Base(parent)
	if !layout.ValidInitiative(id) {
		return ""
	}
	return id
}

// saveItem persiste só o item i: no modo por feature grava o arquivo da feature e não toca em
// queue.json; sem state/ regrava a fila como sempre.
func saveItem(dir string, q *Queue, i int) error {
	if id := initiativeOfQueuePath(dir, QueueFilePath(dir)); id != "" && stateModeActive(dir, id) {
		return WriteFeatureState(dir, id, stateFromItem(q.Items[i]))
	}
	return Save(dir, q)
}

// restorePlan devolve aos itens sobrepostos os campos de execução gravados em queue.json, para
// que Save (Enqueue, SetBase) não grave estado por feature dentro do plano.
func restorePlan(path string, q *Queue, overlaid map[int]bool) {
	if len(overlaid) == 0 {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var raw Queue
	if json.Unmarshal(data, &raw) != nil {
		return
	}
	byID := map[string]QueueItem{}
	for _, it := range raw.Items {
		byID[it.TaskID] = it
	}
	for i := range overlaid {
		if r, ok := byID[q.Items[i].TaskID]; ok {
			q.Items[i].Status, q.Items[i].Attempts = r.Status, r.Attempts
			q.Items[i].BlockedReason, q.Items[i].GovernanceVerdict = r.BlockedReason, r.GovernanceVerdict
		}
	}
}
