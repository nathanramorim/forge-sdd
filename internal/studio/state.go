package studio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
)

// StationDTO representa a projeção para o Studio de uma das 5 estações.
type StationDTO struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Role           string              `json:"role"`
	Color          string              `json:"color"`
	Mark           string              `json:"mark"` // ring, diamond, bar, wave, tri
	MascotState    string              `json:"mascot_state"` // idle, working, thinking, searching, approval, finished, error
	CurrentFocus   string              `json:"current_focus"`
	InputContract  string              `json:"input_contract"`
	OutputContract string              `json:"output_contract"`
	Version        int64               `json:"version"`
	ActiveItems    int                 `json:"active_items"`
	TotalItems     int                 `json:"total_items"`
	Items          []queue.StationItem `json:"items"`
}

// StudioSnapshot representa o estado consolidado determinístico do projeto.
type StudioSnapshot struct {
	ProjectName       string                `json:"project_name"`
	Version           string                `json:"version"`
	ActiveStation     string                `json:"active_station"`
	ActiveStationIdx  int                   `json:"active_station_idx"`
	LastHandoff       *HandoffDTO           `json:"last_handoff,omitempty"`
	GovernanceVerdict *queue.Verdict        `json:"governance_verdict,omitempty"`
	StateHash         string                `json:"state_hash"`
	UpdatedAt         string                `json:"updated_at"`
	Stations          []StationDTO          `json:"stations"`
	LegacyQueue       *queue.Queue          `json:"legacy_queue,omitempty"`
}

// HandoffDTO representa o bastão atual ou em trânsito.
type HandoffDTO struct {
	FromStation string `json:"from_station"`
	ToStation   string `json:"to_station"`
	Payload     string `json:"payload"`
	Gate        string `json:"gate"`
	IsLoop      bool   `json:"is_loop"` // true quando changes_requested faz 05 -> 04
}

// Aggregator lê passivamente o filesystem SDD e monta o StudioSnapshot.
type Aggregator struct {
	Dir string
}

// NewAggregator cria uma nova instância de leitura passiva sobre o diretório do projeto.
func NewAggregator(dir string) *Aggregator {
	return &Aggregator{Dir: dir}
}

// GetSnapshot constrói o snapshot completo em modo somente-leitura.
func (a *Aggregator) GetSnapshot() (*StudioSnapshot, error) {
	stations := []StationDTO{
		{
			ID:             queue.StationLider,
			Name:           "01. O Líder",
			Role:           "Orquestrador · Maestro do Ciclo",
			Color:          "#f2a65a",
			Mark:           "ring",
			MascotState:    "working",
			CurrentFocus:   "Coordenando estações e reportando ao humano",
			InputContract:  "Demanda do Humano / Parecer da Governança",
			OutputContract: "Demanda Estruturada → Estação 02",
		},
		{
			ID:             queue.StationPMArchitect,
			Name:           "02. PM & Arquiteto",
			Role:           "Concepção da Spec & Reuse Check",
			Color:          "#8aa4ff",
			Mark:           "diamond",
			MascotState:    "idle",
			CurrentFocus:   "Elaborando Discovery, Criteria e C4",
			InputContract:  "Demanda do Líder + Código Atual",
			OutputContract: "discovery.md + criteria.md",
		},
		{
			ID:             queue.StationPlanner,
			Name:           "03. O Planejador",
			Role:           "Fatiamento & Mapeamento de Regras",
			Color:          "#c9a7ff",
			Mark:           "bar",
			MascotState:    "idle",
			CurrentFocus:   "Quebrando Spec em Tasks P/M/G determinísticas",
			InputContract:  "Spec Aprovada (Portão FLX-01)",
			OutputContract: "plan.md + tarefas enfileiradas",
		},
		{
			ID:             queue.StationBuilder,
			Name:           "04. Fila de Builders",
			Role:           "Implementação em Contexto Isolado",
			Color:          "#4fd1b5",
			Mark:           "wave",
			MascotState:    "idle",
			CurrentFocus:   "Codificando e executando testes locais",
			InputContract:  "Task da Fila + Rules Mapeadas",
			OutputContract: "Código + Testes + Portão A (PASS)",
		},
		{
			ID:             queue.StationGovernance,
			Name:           "05. Analista de Governança",
			Role:           "Auditoria Somente Leitura (Camada B)",
			Color:          "#ff8a5c",
			Mark:           "tri",
			MascotState:    "idle",
			CurrentFocus:   "Inspecionando conformidade arquitetural e segurança",
			InputContract:  "Diff + Spec + IDs da Constituição",
			OutputContract: "Laudo JSON → Líder (ou ↺ Builder)",
		},
	}

	activeStation := queue.StationLider
	activeIdx := 0

	// Carrega filas das 5 estações
	for i, st := range stations {
		sq, err := queue.LoadStationQueue(a.Dir, st.ID)
		if err == nil && sq != nil {
			stations[i].Version = sq.Version
			stations[i].TotalItems = len(sq.Items)
			stations[i].Items = sq.Items

			activeCount := 0
			for _, it := range sq.Items {
				if it.Status == queue.StatusBuilding {
					stations[i].MascotState = "working"
					activeStation = st.ID
					activeIdx = i
					stations[i].CurrentFocus = it.Title
				} else if it.Status == queue.StatusChangesRequested {
					stations[i].MascotState = "error"
				}
				if it.Status != queue.StatusApproved {
					activeCount++
				}
			}
			stations[i].ActiveItems = activeCount
		}
	}

	// Carrega legado se existir
	legacyQ, _ := queue.Load(a.Dir)

	// Inspeciona governança
	var lastVerdict *queue.Verdict
	if legacyQ != nil {
		for _, it := range legacyQ.Items {
			if it.GovernanceVerdict != nil {
				lastVerdict = it.GovernanceVerdict
			}
		}
	}

	handoff := &HandoffDTO{
		FromStation: stations[activeIdx].ID,
		ToStation:   stations[(activeIdx+1)%len(stations)].ID,
		Payload:     stations[activeIdx].OutputContract,
		Gate:        "Portão Determinístico SDD",
		IsLoop:      false,
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Gera StateHash para SSE (versões combinadas)
	hasher := sha256.New()
	for _, st := range stations {
		fmt.Fprintf(hasher, "%s:%d:%d;", st.ID, st.Version, st.ActiveItems)
	}
	if legacyQ != nil {
		fmt.Fprintf(hasher, "%s:%s;", legacyQ.FeatureID, legacyQ.UpdatedAt)
	}
	stateHash := hex.EncodeToString(hasher.Sum(nil))

	projectName := filepath.Base(filepath.Clean(a.Dir))
	if projectName == "." || projectName == "/" {
		projectName = "forge-sdd"
	}

	return &StudioSnapshot{
		ProjectName:       projectName,
		Version:           readProjectVersion(a.Dir),
		ActiveStation:     activeStation,
		ActiveStationIdx:  activeIdx,
		LastHandoff:       handoff,
		GovernanceVerdict: lastVerdict,
		StateHash:         stateHash,
		UpdatedAt:         now,
		Stations:          stations,
		LegacyQueue:       legacyQ,
	}, nil
}

func readProjectVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "sdd", ".sdd-version"))
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}
	return "3.2.0-beta"
}
