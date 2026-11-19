package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/monitor"
	"github.com/forge-sdd/cli/internal/notify"
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

// archiveGit é o git do archive automático; nil usa o git real (os testes trocam).
var archiveGit queue.GitFunc

// CheckAndAutoArchive roda quando um job é aprovado. Sem pendência na iniciativa: com `origem: toca-ai`
// na fila (gravada pelo motor do /toca-ai) arquiva sozinho, com as mesmas guardas do `forge archive`;
// nas demais, só avisa "pronta para arquivar" no inbox (ADR-0014 D-4).
func CheckAndAutoArchive(dir string, job bus.Job, now time.Time, ncfg config.NotificationsConfig, deps notify.Deps) (*ArchiveResult, error) {
	initID := job.Initiative
	if initID == "" {
		initID = job.Feature
	}
	if initID == "" {
		return &ArchiveResult{Archived: false, Reason: "sem identificação de iniciativa no job"}, nil
	}
	initDir, found := resolveInitiativeDir(dir, initID)
	if !found {
		return &ArchiveResult{Archived: false, Reason: "diretório da iniciativa não encontrado"}, nil
	}
	name := filepath.Base(initDir)
	if reason := pendingReason(initDir); reason != "" {
		return &ArchiveResult{Initiative: name, Reason: reason}, nil
	}
	if q, _ := queue.LoadInitiative(dir, name); q != nil && q.Origem == queue.OrigemTocaAi {
		return AutoArchiveInitiative(dir, name, now, archiveGit)
	}
	n := notify.Notification{
		Type: notify.TypeInfo, Title: "Iniciativa pronta para arquivar: " + name,
		Message: "Todas as tasks foram aprovadas. Na branch da iniciativa, rode: forge archive --initiative " + name,
		Feature: job.Feature, JobID: job.ID, ID: notificationID(job.ID, "arquivar"),
	}
	if _, err := notify.Dispatch(dir, ncfg, n, deps); err != nil {
		return nil, err
	}
	return &ArchiveResult{Initiative: name, Reason: "pronta para arquivar (aviso enviado)"}, nil
}

var fixStatusRe = regexp.MustCompile("(?m)^\\*\\*Status:\\*\\*\\s*`?([A-Za-z_-]+)`?")

// fixDoneStatuses são os valores de `**Status:**` de um fix que já foi entregue.
var fixDoneStatuses = map[string]bool{"done": true, "approved": true, "merged": true}

// pendingReason devolve por que a iniciativa não pode ser arquivada; vazio quando não há pendência.
// Pendente é task ou fix da fila fora de approved e fix de fixes/ cujo `**Status:**` ainda não é de
// entregue (done, approved ou merged) e que não tem item aprovado na fila.
func pendingReason(initDir string) string {
	approved := map[string]bool{}
	if data, err := os.ReadFile(filepath.Join(initDir, layout.QueueFileName)); err == nil {
		var q queue.Queue
		if json.Unmarshal(data, &q) == nil {
			for _, it := range q.Items {
				if it.Status != queue.StatusApproved {
					return fmt.Sprintf("tarefa %s ainda pendente (status: %s)", it.TaskID, it.Status)
				}
				approved[it.TaskID] = true
			}
		}
	}
	entries, _ := os.ReadDir(filepath.Join(initDir, "fixes"))
	for _, e := range entries {
		n := e.Name()
		m := fixFileRe.FindStringSubmatch(n)
		if e.IsDir() || m == nil || approved[m[1]] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(initDir, "fixes", n))
		if err != nil {
			continue
		}
		if st := fixStatusRe.FindStringSubmatch(string(data)); st != nil && !fixDoneStatuses[strings.ToLower(st[1])] {
			return fmt.Sprintf("fix %s ainda pendente (status: %s)", m[1], st[1])
		}
	}
	return ""
}

// AutoArchiveInitiative move sdd/discovery/<id> para sdd/archive/discovery/<id> (a origem deixa de
// existir), registra uma linha no progress-log e grava uma telemetria com os totais reais. É
// idempotente. Só vale na branch da iniciativa e sem task ou fix pendente; senão recusa com o motivo.
// git é injetável nos testes; nil usa o git real.
func AutoArchiveInitiative(dir, initiativeID string, now time.Time, git queue.GitFunc) (*ArchiveResult, error) {
	initDir, found := resolveInitiativeDir(dir, initiativeID)
	if !found {
		if archived := archivedDir(dir, initiativeID); archived != "" {
			return &ArchiveResult{Initiative: filepath.Base(archived), ArchiveDir: archived, Reason: "já arquivada"}, nil
		}
		return &ArchiveResult{Archived: false, Reason: "diretório da iniciativa não encontrado"}, nil
	}
	name := filepath.Base(initDir)

	if reason := pendingReason(initDir); reason != "" {
		return &ArchiveResult{Initiative: name, Reason: reason}, nil
	}
	if git == nil {
		git = queue.RunGit
	}
	if reason := branchGuard(dir, name, git); reason != "" {
		return &ArchiveResult{Initiative: name, Reason: reason}, nil
	}

	archiveTarget := filepath.Join(layout.Archive(dir), layout.DiscoveryDirName, name)
	if err := os.MkdirAll(filepath.Dir(archiveTarget), 0o755); err != nil {
		return nil, fmt.Errorf("falha ao criar pasta de arquivo %s: %w", filepath.Dir(archiveTarget), err)
	}
	var movedFiles []string
	if entries, err := os.ReadDir(initDir); err == nil {
		for _, e := range entries {
			movedFiles = append(movedFiles, e.Name())
		}
	}
	// Os totais saem antes do move: a telemetria da iniciativa continua em sdd/.metrics.
	totals := initiativeTotals(dir, name)
	if err := moveDir(initDir, archiveTarget); err != nil {
		return nil, fmt.Errorf("falha ao mover %s para %s: %w", initDir, archiveTarget, err)
	}

	memDir := layout.Memory(dir)
	_ = os.MkdirAll(memDir, 0o755)
	progressLog := filepath.Join(memDir, "progress-log.md")
	entryLine := fmt.Sprintf("- %s: Iniciativa %s concluída e arquivada com sucesso.\n", now.Format("2006-01-02"), name)
	if !logHasArchive(progressLog, name) {
		_ = appendToFile(progressLog, entryLine)
	}

	metric := monitor.SessionMetrics{
		Schema:              "forge-sdd/metrics/1.0",
		Feature:             filepath.ToSlash(filepath.Join("sdd", "archive", "discovery", name, "status.md")),
		Phase:               name,
		AgentPath:           []string{"archive"},
		Outcome:             "approved",
		CriterioAtendido:    true,
		TokensInput:         totals.Real.TokensInput,
		TokensOutput:        totals.Real.TokensOutput,
		TokensCacheRead:     totals.Real.TokensCacheRead,
		TokensCacheCreation: totals.Real.TokensCacheCreation,
		Turns:               totals.Real.Turns,
		CostUSDTabela:       totals.Real.CostUSDTabela,
		Medicao:             monitor.MedicaoResumo,
		SddVersion:          config.BinaryVersion(),
	}
	tPath, err := monitor.RecordSessionMetrics(dir, metric, now)
	if err != nil {
		return nil, fmt.Errorf("falha ao gravar telemetria do auto-archive: %w", err)
	}

	return &ArchiveResult{
		Archived:      true,
		Initiative:    name,
		SourceDir:     initDir,
		ArchiveDir:    archiveTarget,
		MovedFiles:    movedFiles,
		TelemetryPath: tPath,
	}, nil
}

// branchGuard exige estar na branch da iniciativa; vazio quando está.
func branchGuard(dir, name string, git queue.GitFunc) string {
	branch, err := git(dir, "rev-parse", "--abbrev-ref", "HEAD")
	branch = strings.TrimSpace(branch)
	if err != nil || branch == "" {
		return "não consegui saber a branch atual; rode na branch da iniciativa"
	}
	if id, ok := layout.FromBranch(branch); !ok || id != name {
		return fmt.Sprintf("a branch atual (%s) não é a da iniciativa (%s); troque de branch antes de arquivar", branch, layout.InitiativeBranch(name))
	}
	return ""
}

func archivedDir(dir, id string) string {
	p := filepath.Join(layout.Archive(dir), layout.DiscoveryDirName, strings.TrimSpace(id))
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		return p
	}
	return ""
}

// initiativeTotals soma o uso das sessões da iniciativa gravadas em sdd/.metrics.
func initiativeTotals(dir, name string) monitor.UsageTotals {
	return monitor.SumUsage(monitor.ReadSessions(dir, name))
}

func logHasArchive(path, name string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), "Iniciativa "+name+" concluída e arquivada")
}

// moveDir move src para dst. Se dst já existe (cópia antiga), sobrepõe e remove src.
func moveDir(src, dst string) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return os.Rename(src, dst)
	}
	if err := copyRecursive(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// ListArchiveCandidates devolve as iniciativas sem task pendente e com ao menos uma task, em
// sdd/discovery, sem mover nada (forge archive --concluidas --dry-run).
func ListArchiveCandidates(dir string) []string {
	entries, err := os.ReadDir(layout.Discovery(dir))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() || !layout.ValidInitiative(e.Name()) {
			continue
		}
		initDir := filepath.Join(layout.Discovery(dir), e.Name())
		data, err := os.ReadFile(filepath.Join(initDir, layout.QueueFileName))
		if err != nil {
			continue
		}
		var q queue.Queue
		if json.Unmarshal(data, &q) != nil || len(q.Items) == 0 || pendingReason(initDir) != "" {
			continue
		}
		out = append(out, e.Name())
	}
	return out
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
