package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/queue"
)

// ArchiveResult resume o resultado do auto-archive de uma iniciativa (CA-04).
type ArchiveResult struct {
	Archived      bool     `json:"archived"`
	Initiative    string   `json:"initiative"`
	Reason        string   `json:"reason,omitempty"`
	SourceDir     string   `json:"source_dir,omitempty"`
	ArchiveDir    string   `json:"archive_dir,omitempty"`
	MovedFiles    []string `json:"moved_files,omitempty"`
	TelemetryPath string   `json:"telemetry_path,omitempty"`
}

// CheckAndAutoArchive verifica se uma iniciativa não possui tarefas pendentes e aciona o auto-archive.
func CheckAndAutoArchive(dir string, job bus.Job, now time.Time) (*ArchiveResult, error) {
	initID := job.Feature
	if initID == "" {
		initID = job.Initiative
	}
	if initID == "" {
		return &ArchiveResult{Archived: false, Reason: "sem identificação de iniciativa no job"}, nil
	}

	return AutoArchiveInitiative(dir, initID, now)
}

// AutoArchiveInitiative executa a consolidação e arquivamento de artefatos da iniciativa em sdd/archive/
// e grava a telemetria em sdd/.metrics/ (CA-04).
func AutoArchiveInitiative(dir, initiativeID string, now time.Time) (*ArchiveResult, error) {
	initDir, found := resolveInitiativeDir(dir, initiativeID)
	if !found {
		return &ArchiveResult{Archived: false, Reason: "diretório da iniciativa não encontrado"}, nil
	}

	// 1. Verifica se todas as tarefas da fila foram aprovadas / concluídas
	queueFile := filepath.Join(initDir, layout.QueueFileName)
	if data, err := os.ReadFile(queueFile); err == nil {
		var q queue.Queue
		if err := json.Unmarshal(data, &q); err == nil && len(q.Items) > 0 {
			for _, it := range q.Items {
				if it.Status != queue.StatusApproved {
					return &ArchiveResult{
						Archived:   false,
						Initiative: initiativeID,
						Reason:     fmt.Sprintf("tarefa %s ainda pendente (status: %s)", it.TaskID, it.Status),
					}, nil
				}
			}
		}
	}

	// 2. Destino do arquivo em sdd/archive/discovery/<iniciativa>/
	archiveBase := layout.Archive(dir)
	archiveTarget := filepath.Join(archiveBase, layout.DiscoveryDirName, filepath.Base(initDir))
	if err := os.MkdirAll(archiveTarget, 0o755); err != nil {
		return nil, fmt.Errorf("falha ao criar pasta de arquivo %s: %w", archiveTarget, err)
	}

	// Consolida artefatos para a pasta de histórico
	var movedFiles []string
	entries, err := os.ReadDir(initDir)
	if err == nil {
		for _, e := range entries {
			src := filepath.Join(initDir, e.Name())
			dst := filepath.Join(archiveTarget, e.Name())
			if err := copyRecursive(src, dst); err == nil {
				movedFiles = append(movedFiles, e.Name())
			}
		}
	}

	// 3. Atualiza sdd/memory/progress-log.md
	memDir := layout.Memory(dir)
	_ = os.MkdirAll(memDir, 0o755)
	progressLog := filepath.Join(memDir, "progress-log.md")
	entryLine := fmt.Sprintf("- %s: Iniciativa %s concluída e arquivada com sucesso.\n",
		now.Format("2006-01-02"), filepath.Base(initDir))
	_ = appendToFile(progressLog, entryLine)

	// 4. Registra telemetria de sessão determinística
	metric := monitor.SessionMetrics{
		Schema:           "forge-sdd/metrics/1.0",
		Feature:          filepath.ToSlash(filepath.Join("sdd", "discovery", filepath.Base(initDir), "status.md")),
		Phase:            filepath.Base(initDir),
		AgentPath:        []string{"builder", "reviewer"},
		Outcome:          "approved",
		CriterioAtendido: true,
	}
	tPath, err := monitor.RecordSessionMetrics(dir, metric, now)
	if err != nil {
		return nil, fmt.Errorf("falha ao gravar telemetria do auto-archive: %w", err)
	}

	return &ArchiveResult{
		Archived:      true,
		Initiative:    filepath.Base(initDir),
		SourceDir:     initDir,
		ArchiveDir:    archiveTarget,
		MovedFiles:    movedFiles,
		TelemetryPath: tPath,
	}, nil
}

func resolveInitiativeDir(dir, id string) (string, bool) {
	id = strings.TrimSpace(id)
	discDir := layout.Discovery(dir)

	// Tentativa 1: caminho direto em sdd/discovery/<id>
	direct := filepath.Join(discDir, id)
	if info, err := os.Stat(direct); err == nil && info.IsDir() {
		return direct, true
	}

	// Tentativa 2: correspondência por prefixo ou substring na pasta de discovery
	entries, err := os.ReadDir(discDir)
	if err != nil {
		return "", false
	}

	cleanID := strings.TrimPrefix(id, "feat-")
	cleanID = strings.TrimPrefix(cleanID, "fix-")

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == id || strings.HasPrefix(name, id) || strings.HasPrefix(name, cleanID) {
			return filepath.Join(discDir, name), true
		}
	}

	return "", false
}

func copyRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyRecursive(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func appendToFile(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}
