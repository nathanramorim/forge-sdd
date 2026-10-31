package queue

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Constantes com as 5 estações canônicas de trabalho.
const (
	StationLider        = "lider"
	StationPMArchitect  = "pm-architect"
	StationPlanner      = "planner"
	StationBuilder      = "builder"
	StationGovernance   = "governance"
)

// CanonicalStations retorna a lista ordenada das 5 estações.
func CanonicalStations() []string {
	return []string{
		StationLider,
		StationPMArchitect,
		StationPlanner,
		StationBuilder,
		StationGovernance,
	}
}

// IsValidStation verifica se a estação pertence às 5 canônicas.
func IsValidStation(station string) bool {
	for _, s := range CanonicalStations() {
		if s == station {
			return true
		}
	}
	return false
}

// StationItem representa um item de trabalho dentro da fila de uma estação.
type StationItem struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	Station           string            `json:"station"`
	Operation         string            `json:"operation"`
	Status            string            `json:"status"` // queued, building, review_pending, approved, changes_requested
	ArtifactPath      string            `json:"artifact_path,omitempty"`
	IdempotencyKey    string            `json:"idempotency_key"`
	Attempt           int               `json:"attempt"`
	InputContract     string            `json:"input_contract,omitempty"`
	OutputContract    string            `json:"output_contract,omitempty"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	GovernanceVerdict *Verdict          `json:"governance_verdict,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// StationQueue representa a fila persistida em sdd/.queue/stations/<station>.json.
type StationQueue struct {
	Station       string        `json:"station"`
	Version       int64         `json:"version"`
	UpdatedAt     string        `json:"updated_at"`
	ProcessedKeys []string      `json:"processed_keys"`
	Items         []StationItem `json:"items"`
}

var stationMutex sync.RWMutex

// StationsDir retorna o diretório das 5 filas de estações.
func StationsDir(dir string) string {
	return filepath.Join(dir, "sdd", ".queue", "stations")
}

// StationQueueFilePath retorna o caminho completo do arquivo da estação.
func StationQueueFilePath(dir, station string) string {
	return filepath.Join(StationsDir(dir), station+".json")
}

// CalculateIdempotencyKey calcula uma chave SHA-256 estável para uma operação de estação.
func CalculateIdempotencyKey(station, itemID, operation, artifactPath string, attempt int) string {
	raw := fmt.Sprintf("%s:%s:%s:%s:%d", strings.TrimSpace(station), strings.TrimSpace(itemID), strings.TrimSpace(operation), strings.TrimSpace(artifactPath), attempt)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:16]) // 32 caracteres hex
}

// LoadStationQueue carrega a fila da estação específica. Se não existir, retorna estrutura vazia.
func LoadStationQueue(dir, station string) (*StationQueue, error) {
	stationMutex.RLock()
	defer stationMutex.RUnlock()

	return loadStationQueueLocked(dir, station)
}

func loadStationQueueLocked(dir, station string) (*StationQueue, error) {
	if !IsValidStation(station) {
		return nil, fmt.Errorf("estação inválida: %q (esperado uma de: %s)", station, strings.Join(CanonicalStations(), ", "))
	}

	path := StationQueueFilePath(dir, station)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &StationQueue{
				Station:       station,
				Version:       0,
				UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
				ProcessedKeys: []string{},
				Items:         []StationItem{},
			}, nil
		}
		return nil, fmt.Errorf("falha ao ler %s: %w", path, err)
	}

	var sq StationQueue
	if err := json.Unmarshal(data, &sq); err != nil {
		return nil, fmt.Errorf("falha ao decodificar %s: %w", path, err)
	}
	if sq.Station == "" {
		sq.Station = station
	}
	if sq.ProcessedKeys == nil {
		sq.ProcessedKeys = []string{}
	}
	if sq.Items == nil {
		sq.Items = []StationItem{}
	}
	return &sq, nil
}

// SaveStationQueue persiste a fila de uma estação usando escrita atômica e incrementando a versão.
func SaveStationQueue(dir string, sq *StationQueue) error {
	stationMutex.Lock()
	defer stationMutex.Unlock()

	return saveStationQueueLocked(dir, sq)
}

func saveStationQueueLocked(dir string, sq *StationQueue) error {
	if !IsValidStation(sq.Station) {
		return fmt.Errorf("estação inválida ao salvar: %q", sq.Station)
	}
	sq.Version++
	sq.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	path := StationQueueFilePath(dir, sq.Station)
	return WriteJSONAtomic(path, sq, 0644)
}

// EnqueueStationItem enfileira um item na estação de forma estritamente idempotente.
// Retorna (item, isIdempotentHit, error).
func EnqueueStationItem(dir string, item StationItem) (*StationItem, bool, error) {
	stationMutex.Lock()
	defer stationMutex.Unlock()

	if !IsValidStation(item.Station) {
		return nil, false, fmt.Errorf("estação inválida: %q", item.Station)
	}
	if item.ID == "" {
		return nil, false, fmt.Errorf("ID do item da estação não pode ser vazio")
	}
	if item.Operation == "" {
		item.Operation = "process"
	}
	if item.Status == "" {
		item.Status = StatusQueued
	}
	if item.IdempotencyKey == "" {
		item.IdempotencyKey = CalculateIdempotencyKey(item.Station, item.ID, item.Operation, item.ArtifactPath, item.Attempt)
	}

	sq, err := loadStationQueueLocked(dir, item.Station)
	if err != nil {
		return nil, false, err
	}

	// 1. Verifica se a chave de idempotência já foi processada
	for _, key := range sq.ProcessedKeys {
		if key == item.IdempotencyKey {
			// Procura o item existente
			for _, it := range sq.Items {
				if it.IdempotencyKey == item.IdempotencyKey || it.ID == item.ID {
					return &it, true, nil
				}
			}
			return &item, true, nil
		}
	}

	// 2. Verifica se já existe um item com o mesmo ID
	for idx, it := range sq.Items {
		if it.ID == item.ID {
			// Item já existe na fila: se mesma chave ou mesmo status, considera hit idempotente
			if it.IdempotencyKey == item.IdempotencyKey {
				return &it, true, nil
			}
			// Atualização de dados mantendo histórico
			item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			sq.Items[idx] = item
			sq.ProcessedKeys = appendProcessedKey(sq.ProcessedKeys, item.IdempotencyKey)
			if err := saveStationQueueLocked(dir, sq); err != nil {
				return nil, false, err
			}
			return &item, false, nil
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if item.CreatedAt == "" {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	sq.Items = append(sq.Items, item)
	sq.ProcessedKeys = appendProcessedKey(sq.ProcessedKeys, item.IdempotencyKey)

	if err := saveStationQueueLocked(dir, sq); err != nil {
		return nil, false, err
	}

	return &item, false, nil
}

// Handoff realiza a transição segura e idempotente de um artefato entre duas estações.
// Conclui na estação de origem e enfileira na estação de destino.
func Handoff(dir, fromStation, toStation, itemID, title, artifactPath, inputContract, outputContract, idempotencyKey string) (*StationItem, bool, error) {
	stationMutex.Lock()
	defer stationMutex.Unlock()

	if !IsValidStation(fromStation) || !IsValidStation(toStation) {
		return nil, false, fmt.Errorf("estação inválida no handoff (%s -> %s)", fromStation, toStation)
	}

	if idempotencyKey == "" {
		idempotencyKey = CalculateIdempotencyKey(toStation, itemID, "handoff-from-"+fromStation, artifactPath, 1)
	}

	// 1. Atualiza estação de origem (marca qualquer item correspondente como approved/concluído)
	fromSQ, err := loadStationQueueLocked(dir, fromStation)
	if err == nil {
		updatedFrom := false
		for i, it := range fromSQ.Items {
			if it.ID == itemID || (artifactPath != "" && it.ArtifactPath == artifactPath) {
				fromSQ.Items[i].Status = StatusApproved
				fromSQ.Items[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
				updatedFrom = true
			}
		}
		if updatedFrom {
			_ = saveStationQueueLocked(dir, fromSQ)
		}
	}

	// 2. Enfileira na estação de destino de forma idempotente
	toSQ, err := loadStationQueueLocked(dir, toStation)
	if err != nil {
		return nil, false, err
	}

	for _, key := range toSQ.ProcessedKeys {
		if key == idempotencyKey {
			for _, it := range toSQ.Items {
				if it.IdempotencyKey == idempotencyKey || it.ID == itemID {
					return &it, true, nil
				}
			}
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	newItem := StationItem{
		ID:             itemID,
		Title:          title,
		Station:        toStation,
		Operation:      "handoff",
		Status:         StatusQueued,
		ArtifactPath:   artifactPath,
		IdempotencyKey: idempotencyKey,
		Attempt:        1,
		InputContract:  inputContract,
		OutputContract: outputContract,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	toSQ.Items = append(toSQ.Items, newItem)
	toSQ.ProcessedKeys = appendProcessedKey(toSQ.ProcessedKeys, idempotencyKey)

	if err := saveStationQueueLocked(dir, toSQ); err != nil {
		return nil, false, err
	}

	return &newItem, false, nil
}

// SyncLegacyQueue sincroniza uma Queue legado (sdd/.queue/queue.json) com as filas das estações.
// As tasks do Builder são sincronizadas em builder.json e os pareceres de governança em governance.json.
func SyncLegacyQueue(dir string, q *Queue) error {
	if q == nil {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Sincroniza builder
	builderSQ, err := LoadStationQueue(dir, StationBuilder)
	if err != nil {
		builderSQ = &StationQueue{Station: StationBuilder, ProcessedKeys: []string{}}
	}

	for _, it := range q.Items {
		idemKey := CalculateIdempotencyKey(StationBuilder, it.TaskID, "build", it.SpecPath, it.Attempts)
		found := false
		for idx, sit := range builderSQ.Items {
			if sit.ID == it.TaskID {
				builderSQ.Items[idx].Status = it.Status
				builderSQ.Items[idx].Attempt = it.Attempts
				builderSQ.Items[idx].GovernanceVerdict = it.GovernanceVerdict
				builderSQ.Items[idx].UpdatedAt = now
				found = true
				break
			}
		}
		if !found {
			builderSQ.Items = append(builderSQ.Items, StationItem{
				ID:                it.TaskID,
				Title:             it.Title,
				Station:           StationBuilder,
				Operation:         "build",
				Status:            it.Status,
				ArtifactPath:      it.SpecPath,
				IdempotencyKey:    idemKey,
				Attempt:           it.Attempts,
				CreatedAt:         now,
				UpdatedAt:         now,
				GovernanceVerdict: it.GovernanceVerdict,
			})
			builderSQ.ProcessedKeys = appendProcessedKey(builderSQ.ProcessedKeys, idemKey)
		}
	}
	_ = SaveStationQueue(dir, builderSQ)

	// Sincroniza governance
	govSQ, err := LoadStationQueue(dir, StationGovernance)
	if err != nil {
		govSQ = &StationQueue{Station: StationGovernance, ProcessedKeys: []string{}}
	}

	for _, it := range q.Items {
		if it.GovernanceVerdict != nil {
			govKey := CalculateIdempotencyKey(StationGovernance, it.TaskID, "review", it.SpecPath, it.Attempts)
			found := false
			for idx, sit := range govSQ.Items {
				if sit.ID == it.TaskID {
					govSQ.Items[idx].GovernanceVerdict = it.GovernanceVerdict
					govSQ.Items[idx].Status = it.Status
					govSQ.Items[idx].UpdatedAt = now
					found = true
					break
				}
			}
			if !found {
				govSQ.Items = append(govSQ.Items, StationItem{
					ID:                it.TaskID,
					Title:             "Auditoria Camada B: " + it.Title,
					Station:           StationGovernance,
					Operation:         "review",
					Status:            it.Status,
					ArtifactPath:      it.SpecPath,
					IdempotencyKey:    govKey,
					Attempt:           it.Attempts,
					CreatedAt:         now,
					UpdatedAt:         now,
					GovernanceVerdict: it.GovernanceVerdict,
				})
				govSQ.ProcessedKeys = appendProcessedKey(govSQ.ProcessedKeys, govKey)
			}
		}
	}
	_ = SaveStationQueue(dir, govSQ)

	return nil
}

func appendProcessedKey(slice []string, key string) []string {
	for _, k := range slice {
		if k == key {
			return slice
		}
	}
	// Mantém janela deslizante de no máximo 200 chaves para evitar crescimento indefinido
	if len(slice) >= 200 {
		slice = slice[1:]
	}
	return append(slice, key)
}
