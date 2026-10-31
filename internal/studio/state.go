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
	Mark           string              `json:"mark"`        // ring, diamond, bar, wave, tri
	MascotState    string              `json:"mascot_state"` // idle, working, thinking, searching, approval, finished, error
	CurrentFocus   string              `json:"current_focus"`
	ExecutingTask  *ActiveTaskDTO      `json:"executing_task,omitempty"`
	ActiveTools    []string            `json:"active_tools,omitempty"`
	CurrentAction  string              `json:"current_action,omitempty"`
	RecentActivity []queue.ActivityItem `json:"recent_activity,omitempty"`
	InputContract  string              `json:"input_contract"`
	OutputContract string              `json:"output_contract"`
	Version        int64               `json:"version"`
	ActiveItems    int                 `json:"active_items"`
	TotalItems     int                 `json:"total_items"`
	Items          []queue.StationItem `json:"items"`
}

// ActiveTaskDTO descreve com precisão a tarefa em processamento por um agente.
type ActiveTaskDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Operation string `json:"operation,omitempty"`
}

// StudioSnapshot representa o estado consolidado determinístico do projeto.
type StudioSnapshot struct {
	ProjectName       string                `json:"project_name"`
	ProjectDir        string                `json:"project_dir"`
	Version           string                `json:"version"`
	ActiveStation     string                `json:"active_station"`
	ActiveStationIdx  int                   `json:"active_station_idx"`
	LastHandoff       *HandoffDTO           `json:"last_handoff,omitempty"`
	GovernanceVerdict *queue.Verdict        `json:"governance_verdict,omitempty"`
	RecentActivity    []queue.ActivityItem  `json:"recent_activity,omitempty"`
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
			MascotState:    "idle",
			CurrentFocus:   "Orquestrando ciclo, monitorando filas e portões",
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
			OutputContract: "discovery-<ID>-<nome>.md + criteria-<ID>-<nome>.md",
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
			OutputContract: "plan-<ID>-<nome>.md + tarefas enfileiradas",
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
			var pendingApprovalItem *queue.StationItem
			var queuedItem *queue.StationItem
			for idxItem, it := range sq.Items {
				if st.ID == queue.StationPMArchitect && (it.Status == "pending_review" || it.Status == "approval" || it.Status == "review_pending") {
					pendingApprovalItem = &sq.Items[idxItem]
				}
				if it.Status == queue.StatusQueued && queuedItem == nil {
					queuedItem = &sq.Items[idxItem]
				}
				if it.Status == queue.StatusBuilding {
					stations[i].MascotState = "working"
					activeStation = st.ID
					activeIdx = i
					stations[i].CurrentFocus = it.Title
					stations[i].ExecutingTask = &ActiveTaskDTO{
						ID:        it.ID,
						Title:     it.Title,
						Status:    it.Status,
						Operation: it.Operation,
					}
				} else if it.Status == queue.StatusChangesRequested {
					stations[i].MascotState = "error"
					stations[i].CurrentFocus = "Retrabalho Solicitado: " + it.Title
					stations[i].ExecutingTask = &ActiveTaskDTO{
						ID:        it.ID,
						Title:     it.Title,
						Status:    it.Status,
						Operation: it.Operation,
					}
				}
				if it.Status != queue.StatusApproved && it.Status != "done" {
					activeCount++
				}
			}
			if st.ID == queue.StationPMArchitect && pendingApprovalItem != nil && stations[i].ExecutingTask == nil {
				stations[i].MascotState = "approval"
				activeStation = st.ID
				activeIdx = i
				stations[i].CurrentFocus = "Aguardando Revisão Humana (FLX-01): " + pendingApprovalItem.Title
				stations[i].ExecutingTask = &ActiveTaskDTO{
					ID:        pendingApprovalItem.ID,
					Title:     pendingApprovalItem.Title,
					Status:    "pending_review",
					Operation: pendingApprovalItem.Operation,
				}
			}
			if queuedItem != nil && stations[i].ExecutingTask == nil {
				stations[i].ExecutingTask = &ActiveTaskDTO{
					ID:        queuedItem.ID,
					Title:     queuedItem.Title,
					Status:    queuedItem.Status,
					Operation: queuedItem.Operation,
				}
			}
			stations[i].ActiveItems = activeCount
		}
	}

	// Se a estação planner tiver tarefa em fila e nenhuma outra estação estiver em building, ativa o Planejador
	hasBuildingStation := false
	plannerIdx := -1
	for idx, s := range stations {
		if s.ID == queue.StationPlanner {
			plannerIdx = idx
		}
		if s.ExecutingTask != nil && s.ExecutingTask.Status == queue.StatusBuilding {
			hasBuildingStation = true
		}
	}
	if !hasBuildingStation && plannerIdx >= 0 && stations[plannerIdx].ExecutingTask != nil && stations[plannerIdx].ExecutingTask.Status == queue.StatusQueued {
		stations[plannerIdx].MascotState = "thinking"
		stations[plannerIdx].CurrentFocus = "Fila: " + stations[plannerIdx].ExecutingTask.Title
		activeStation = queue.StationPlanner
		activeIdx = plannerIdx
	}

	// Carrega legado se existir e vincula à estação ativa correspondente
	legacyQ, _ := queue.Load(a.Dir)
	var lastVerdict *queue.Verdict
	if legacyQ != nil {
		for _, it := range legacyQ.Items {
			if it.GovernanceVerdict != nil {
				lastVerdict = it.GovernanceVerdict
			}
			if it.Status == queue.StatusBuilding {
				for bIdx, s := range stations {
					if s.ID == queue.StationBuilder && stations[bIdx].ExecutingTask == nil {
						stations[bIdx].MascotState = "working"
						stations[bIdx].CurrentFocus = it.Title
						stations[bIdx].ExecutingTask = &ActiveTaskDTO{
							ID:     it.TaskID,
							Title:  it.Title,
							Status: it.Status,
						}
						activeStation = queue.StationBuilder
						activeIdx = bIdx
					}
				}
			} else if it.Status == queue.StatusReviewPending {
				for gIdx, s := range stations {
					if s.ID == queue.StationGovernance && stations[gIdx].ExecutingTask == nil {
						stations[gIdx].MascotState = "approval"
						stations[gIdx].CurrentFocus = "Auditando Conformidade: " + it.Title
						stations[gIdx].ExecutingTask = &ActiveTaskDTO{
							ID:     it.TaskID,
							Title:  it.Title,
							Status: "review_pending",
						}
						activeStation = queue.StationGovernance
						activeIdx = gIdx
					}
				}
			}
		}

		if (legacyQ.SpecStatus == "pending_review" || legacyQ.SpecStatus == "approval" || legacyQ.SpecStatus == "review_pending") {
			for pmIdx, s := range stations {
				if s.ID == queue.StationPMArchitect && stations[pmIdx].ExecutingTask == nil {
					stations[pmIdx].MascotState = "approval"
					stations[pmIdx].CurrentFocus = "Aguardando Revisão Humana (FLX-01)"
					stations[pmIdx].ExecutingTask = &ActiveTaskDTO{
						ID:     legacyQ.FeatureID,
						Title:  "Aguardando Revisão Humana da Spec (FLX-01)",
						Status: "pending_review",
					}
					activeStation = queue.StationPMArchitect
					activeIdx = pmIdx
				}
			}
		}
	}

	// Carrega histórico e ferramentas recentes de atividade
	activities, _ := queue.LoadActivities(a.Dir)
	for i, st := range stations {
		toolMap := make(map[string]bool)
		var stActivities []queue.ActivityItem
		for _, act := range activities {
			if act.Station == st.ID {
				stActivities = append(stActivities, act)
				if len(toolMap) < 4 {
					toolMap[act.Tool] = true
				}
			}
		}
		var tools []string
		for t := range toolMap {
			tools = append(tools, t)
		}
		stations[i].ActiveTools = tools
		stations[i].RecentActivity = stActivities
		if len(stActivities) > 0 {
			stations[i].CurrentAction = fmt.Sprintf("%s: %s", stActivities[0].Tool, stActivities[0].Action)
		}
	}

	// Se nenhuma estação tiver task em building na fila, a atividade viva mais recente define a estação ativa
	if stations[activeIdx].MascotState == "idle" && len(activities) > 0 {
		latest := activities[0]
		for i, st := range stations {
			if st.ID == latest.Station {
				activeStation = st.ID
				activeIdx = i
				if latest.Status == "running" || latest.Status == "working" {
					if st.ID == queue.StationPMArchitect {
						stations[i].MascotState = "thinking"
					} else if st.ID == queue.StationGovernance {
						stations[i].MascotState = "approval"
					} else {
						stations[i].MascotState = "working"
					}
					stations[i].CurrentFocus = fmt.Sprintf("[%s] %s", latest.Tool, latest.Action)
				}
				break
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

	if activeStation == queue.StationPMArchitect && stations[activeIdx].MascotState == "approval" {
		handoff = &HandoffDTO{
			FromStation: queue.StationPMArchitect,
			ToStation:   queue.StationPlanner,
			Payload:     "discovery.md + criteria.md",
			Gate:        "Portão FLX-01 Bloqueado (Aguardando Aprovação Humana)",
			IsLoop:      false,
		}
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
	if len(activities) > 0 {
		fmt.Fprintf(hasher, "%s:%s;", activities[0].Timestamp, activities[0].Action)
	}
	stateHash := hex.EncodeToString(hasher.Sum(nil))

	projectName := filepath.Base(filepath.Clean(a.Dir))
	if projectName == "." || projectName == "/" {
		projectName = "forge-sdd"
	}

	absDir, err := filepath.Abs(a.Dir)
	if err != nil {
		absDir = a.Dir
	}

	return &StudioSnapshot{
		ProjectName:       projectName,
		ProjectDir:        filepath.ToSlash(absDir),
		Version:           readProjectVersion(a.Dir),
		ActiveStation:     activeStation,
		ActiveStationIdx:  activeIdx,
		LastHandoff:       handoff,
		GovernanceVerdict: lastVerdict,
		RecentActivity:    activities,
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
	return "3.2.1-beta"
}
