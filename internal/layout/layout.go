// Package layout é a casa única dos caminhos do diretório sdd/ (DUP-01, DUP-04).
//
// Todo código que precisa localizar um arquivo do Forge-SDD resolve o caminho aqui,
// em vez de montar filepath.Join(dir, "sdd", ...) à mão. O pacote não importa
// nenhum outro pacote interno (ARQ-02) e só executa o git com argv explícito (SEG-03).
package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Nomes canônicos de pastas e arquivos do layout.
const (
	SddDirName       = "sdd"
	DiscoveryDirName = "discovery"
	FeaturesDirName  = "features"
	FixesDirName     = "fixes"
	MemoryDirName    = "memory"
	MetricsDirName   = ".metrics"
	QueueDirName     = ".queue"
	StationsDirName  = "stations"
	SkillsDirName    = "skills"
	ArchiveDirName   = "archive"
	RuntimeDirName   = ".runtime"

	SddrcFileName        = ".sddrc"
	VersionFileName      = ".sdd-version"
	AutoPilotFileName    = ".sdd-auto-pilot"
	QueueFileName        = "queue.json"
	ActivityFileName     = "activity.json"
	StatusFileName       = "status.md"
	ConstitutionFileName = "constitution.md"
	ProgressFileName     = "progress.md"
)

// Pipeline multi-sessão (feat 70): configuração versionada e estado de execução em .forge/.
const (
	ForgeDirName          = ".forge"
	ForgeConfigDirName    = "config"
	ForgeRuntimeDirName   = "runtime"
	QueuesDirName         = "queues"
	WorktreesDirName      = "worktrees"
	LogsDirName           = "logs"
	ForgeConfigFileName   = "forge.json"
	NotificationsFileName = "notifications.json"

	RoleBuilder   = "builder"
	RoleReviewer  = "reviewer"
	RoleDiscovery = "discovery"

	StatePending    = "pending"
	StateProcessing = "processing"
	StateDone       = "done"
	StateFailed     = "failed"
	StateInbox      = "inbox"
	StateArchive    = "archive"
)

// ErrInvalidQueue indica papel ou estado de fila fora da lista fechada.
var ErrInvalidQueue = errors.New("fila inválida")

// workerStates e discoveryStates são os estados válidos por papel (SEG-01).
var (
	workerStates    = map[string]bool{StatePending: true, StateProcessing: true, StateDone: true, StateFailed: true}
	discoveryStates = map[string]bool{StateInbox: true, StateArchive: true}
)

// Forge devolve <dir>/.forge.
func Forge(dir string) string { return filepath.Join(dir, ForgeDirName) }

// ForgeConfigFile devolve .forge/config/forge.json.
func ForgeConfigFile(dir string) string {
	return filepath.Join(Forge(dir), ForgeConfigDirName, ForgeConfigFileName)
}

// NotificationsFile devolve .forge/config/notifications.json.
func NotificationsFile(dir string) string {
	return filepath.Join(Forge(dir), ForgeConfigDirName, NotificationsFileName)
}

// ForgeRuntime devolve .forge/runtime (estado de execução, fora do git).
func ForgeRuntime(dir string) string { return filepath.Join(Forge(dir), ForgeRuntimeDirName) }

// Worktrees devolve .forge/runtime/worktrees.
func Worktrees(dir string) string { return filepath.Join(ForgeRuntime(dir), WorktreesDirName) }

// Logs devolve .forge/runtime/logs.
func Logs(dir string) string { return filepath.Join(ForgeRuntime(dir), LogsDirName) }

// QueueState devolve .forge/runtime/queues/<papel>/<estado>. Papel e estado são validados
// contra a lista fechada: builder e reviewer têm pending, processing, done e failed;
// discovery tem inbox e archive.
func QueueState(dir, role, state string) (string, error) {
	switch role {
	case RoleBuilder, RoleReviewer:
		if !workerStates[state] {
			return "", fmt.Errorf("%w: estado %q inválido para o papel %q", ErrInvalidQueue, state, role)
		}
	case RoleDiscovery:
		if !discoveryStates[state] {
			return "", fmt.Errorf("%w: estado %q inválido para o papel %q", ErrInvalidQueue, state, role)
		}
	default:
		return "", fmt.Errorf("%w: papel %q (use builder, reviewer ou discovery)", ErrInvalidQueue, role)
	}
	return filepath.Join(ForgeRuntime(dir), QueuesDirName, role, state), nil
}

// Inbox devolve .forge/runtime/queues/discovery/inbox.
func Inbox(dir string) string {
	p, _ := QueueState(dir, RoleDiscovery, StateInbox)
	return p
}

// InboxArchive devolve .forge/runtime/queues/discovery/archive.
func InboxArchive(dir string) string {
	p, _ := QueueState(dir, RoleDiscovery, StateArchive)
	return p
}

// initiativeSlug valida o nome da pasta de uma iniciativa: <id>-<nome>, minúsculo,
// sem separadores de caminho nem pontos (SEG-01).
var initiativeSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)+$`)

// branchPrefixes são os prefixos de branch que apontam para uma iniciativa.
var branchPrefixes = map[string]bool{"feat": true, "fix": true}

// ErrInvalidInitiative indica um identificador de iniciativa fora do padrão <id>-<nome>.
var ErrInvalidInitiative = errors.New("identificador de iniciativa inválido")

// Root devolve <dir>/sdd.
func Root(dir string) string { return filepath.Join(dir, SddDirName) }

// Sddrc devolve o caminho de sdd/.sddrc.
func Sddrc(dir string) string { return filepath.Join(Root(dir), SddrcFileName) }

// Version devolve o caminho de sdd/.sdd-version.
func Version(dir string) string { return filepath.Join(Root(dir), VersionFileName) }

// AutoPilot devolve o caminho do marcador sdd/.sdd-auto-pilot.
func AutoPilot(dir string) string { return filepath.Join(Root(dir), AutoPilotFileName) }

// Memory devolve sdd/memory.
func Memory(dir string) string { return filepath.Join(Root(dir), MemoryDirName) }

// Constitution devolve sdd/memory/constitution.md.
func Constitution(dir string) string { return filepath.Join(Memory(dir), ConstitutionFileName) }

// Progress devolve sdd/memory/progress.md (estado global, legado até a feat 69-03).
func Progress(dir string) string { return filepath.Join(Memory(dir), ProgressFileName) }

// Metrics devolve sdd/.metrics.
func Metrics(dir string) string { return filepath.Join(Root(dir), MetricsDirName) }

// Skills devolve sdd/skills (índice de skills).
func Skills(dir string) string { return filepath.Join(Root(dir), SkillsDirName) }

// Discovery devolve sdd/discovery.
func Discovery(dir string) string { return filepath.Join(Root(dir), DiscoveryDirName) }

// Features devolve sdd/features (raiz legada de features).
func Features(dir string) string { return filepath.Join(Root(dir), FeaturesDirName) }

// Archive devolve sdd/archive.
func Archive(dir string) string { return filepath.Join(Root(dir), ArchiveDirName) }

// Runtime devolve sdd/.runtime (estado de execução fora do git).
func Runtime(dir string) string { return filepath.Join(Root(dir), RuntimeDirName) }

// QueueDir devolve sdd/.queue.
func QueueDir(dir string) string { return filepath.Join(Root(dir), QueueDirName) }

// LegacyQueueFile devolve sdd/.queue/queue.json, a fila global usada antes da feat 69-03.
// Só é usada quando a branch atual não aponta para uma iniciativa existente.
func LegacyQueueFile(dir string) string { return filepath.Join(QueueDir(dir), QueueFileName) }

// StationsDir devolve sdd/.runtime/stations (estado de execução, fora do git).
func StationsDir(dir string) string { return filepath.Join(Runtime(dir), StationsDirName) }

// ActivityFile devolve sdd/.runtime/activity.json (estado de execução, fora do git).
func ActivityFile(dir string) string { return filepath.Join(Runtime(dir), ActivityFileName) }

// ActiveInitiative devolve o id e a pasta da iniciativa da branch atual, desde que a pasta
// exista em sdd/discovery. Sem git, branch fora do padrão ou pasta ausente devolve ok=false.
func ActiveInitiative(dir string) (id, path string, ok bool) {
	id, found, err := Current(dir)
	if err != nil || !found {
		return "", "", false
	}
	path, err = Initiative(dir, id)
	if err != nil {
		return "", "", false
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return "", "", false
	}
	return id, path, true
}

// QueuePath devolve a fila do Builder da iniciativa ativa (sdd/discovery/<id>/queue.json) e,
// sem iniciativa ativa, a fila legada sdd/.queue/queue.json.
func QueuePath(dir string) string {
	if _, path, ok := ActiveInitiative(dir); ok {
		return filepath.Join(path, QueueFileName)
	}
	return LegacyQueueFile(dir)
}

// StatusPath devolve o status.md da iniciativa ativa. ok=false indica que não há iniciativa ativa.
func StatusPath(dir string) (path string, ok bool) {
	_, base, ok := ActiveInitiative(dir)
	if !ok {
		return "", false
	}
	return filepath.Join(base, StatusFileName), true
}

// ValidInitiative informa se id tem o formato <id>-<nome> aceito pelo layout.
func ValidInitiative(id string) bool { return initiativeSlug.MatchString(id) }

// Initiative devolve sdd/discovery/<id>. O id é validado contra path traversal (SEG-01).
func Initiative(dir, id string) (string, error) {
	if !ValidInitiative(id) {
		return "", fmt.Errorf("%w: %q (esperado <id>-<nome> em minúsculas, sem separadores de caminho)", ErrInvalidInitiative, id)
	}
	return filepath.Join(Discovery(dir), id), nil
}

// InitiativeFeatures devolve sdd/discovery/<id>/features.
func InitiativeFeatures(dir, id string) (string, error) {
	base, err := Initiative(dir, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, FeaturesDirName), nil
}

// InitiativeFixes devolve sdd/discovery/<id>/fixes.
func InitiativeFixes(dir, id string) (string, error) {
	base, err := Initiative(dir, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, FixesDirName), nil
}

// InitiativeStatus devolve sdd/discovery/<id>/status.md.
func InitiativeStatus(dir, id string) (string, error) {
	base, err := Initiative(dir, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, StatusFileName), nil
}

// InitiativeQueue devolve sdd/discovery/<id>/queue.json.
func InitiativeQueue(dir, id string) (string, error) {
	base, err := Initiative(dir, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, QueueFileName), nil
}

// FromBranch extrai o identificador da iniciativa de uma branch <prefixo>/<id>-<nome>.
// Branches fora do padrão (main, release/*, chore/*, HEAD destacado) devolvem ok=false.
func FromBranch(branch string) (id string, ok bool) {
	prefix, slug, found := strings.Cut(strings.TrimSpace(branch), "/")
	if !found || !branchPrefixes[prefix] || !ValidInitiative(slug) {
		return "", false
	}
	return slug, true
}

// Current devolve a iniciativa da branch atual de dir. Sem iniciativa ativa devolve
// ok=false e erro nulo; só falha se o git não puder ser consultado.
func Current(dir string) (id string, ok bool, err error) {
	branch, err := CurrentBranch(dir)
	if err != nil {
		return "", false, err
	}
	id, ok = FromBranch(branch)
	return id, ok, nil
}

// CurrentBranch devolve o nome da branch atual (vazio em HEAD destacado).
func CurrentBranch(dir string) (string, error) {
	headPath := filepath.Join(dir, ".git", "HEAD")
	if data, err := os.ReadFile(headPath); err == nil {
		s := strings.TrimSpace(string(data))
		if strings.HasPrefix(s, "ref: refs/heads/") {
			return strings.TrimPrefix(s, "ref: refs/heads/"), nil
		}
	}
	out, err := gitOut(dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
