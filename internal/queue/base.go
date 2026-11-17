package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
)

// OrigemTocaAi marca a fila criada pelo motor do /toca-ai (archive automático ao aprovar).
const OrigemTocaAi = "toca-ai"

// Fontes da base do PR, na ordem em que o resolvedor as consulta.
const (
	BaseFromFlag   = "flag"
	BaseFromQueue  = "fila"
	BaseFromGit    = "origem da branch (git)"
	BaseFromConfig = "forge.json"
)

var (
	branchNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,99}$`)
	hexRe        = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
)

// ValidBranchName valida o nome de uma branch vindo de flag ou da fila (SEG-01).
func ValidBranchName(name string) error {
	if !branchNameRe.MatchString(name) || strings.Contains(name, "..") || strings.HasSuffix(name, "/") || strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch %q inválida", name)
	}
	return nil
}

// GitFunc roda `git <args>` em dir e devolve a saída; injetável nos testes (TST-04).
type GitFunc func(dir string, args ...string) (string, error)

// RunGit é o GitFunc real: argv explícito, sem shell (SEG-03).
func RunGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// SetBase grava base_branch na fila ativa (flag --base do enqueue).
func SetBase(dir, base string) error {
	if err := ValidBranchName(base); err != nil {
		return err
	}
	q, err := Load(dir)
	if err != nil {
		return err
	}
	q.BaseBranch = base
	return Save(dir, q)
}

// SetOrigem grava a origem da fila da iniciativa (ex.: "toca-ai"), sem depender da branch atual.
func SetOrigem(dir, initiative, origem string) error {
	q, path, err := loadInitiativeQueue(dir, initiative)
	if err != nil || q == nil {
		return err
	}
	if q.Origem == origem {
		return nil
	}
	q.Origem = origem
	return WriteJSONAtomic(path, q, 0o644)
}

// LoadInitiative lê a fila de uma iniciativa; nil quando ela não existe.
func LoadInitiative(dir, initiative string) (*Queue, error) {
	q, _, err := loadInitiativeQueue(dir, initiative)
	return q, err
}

func loadInitiativeQueue(dir, initiative string) (*Queue, string, error) {
	path, err := layout.InitiativeQueue(dir, initiative)
	if err != nil {
		if errors.Is(err, layout.ErrInvalidInitiative) {
			return nil, "", nil
		}
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, path, nil
		}
		return nil, "", fmt.Errorf("falha ao ler %s: %w", path, err)
	}
	var q Queue
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, "", fmt.Errorf("falha ao decodificar %s: %w", path, err)
	}
	return &q, path, nil
}

// ResolveBase é o único resolvedor da base do PR de uma iniciativa (DUP-04): flag, depois a fila,
// depois a origem da branch no git ("Created from") e, por último, o base_branch do forge.json.
// Devolve a base e de onde ela veio.
func ResolveBase(dir, initiative, flag, configBase string, git GitFunc) (base, source string) {
	if flag != "" && ValidBranchName(flag) == nil {
		return flag, BaseFromFlag
	}
	if q, _ := LoadInitiative(dir, initiative); q != nil && q.BaseBranch != "" && ValidBranchName(q.BaseBranch) == nil {
		return q.BaseBranch, BaseFromQueue
	}
	if git == nil {
		git = RunGit
	}
	if b := createdFrom(dir, layout.InitiativeBranch(initiative), git); b != "" {
		return b, BaseFromGit
	}
	return configBase, BaseFromConfig
}

// createdFrom lê a origem da branch no reflog; vazio quando não há ("HEAD" e hashes não servem).
func createdFrom(dir, branch string, git GitFunc) string {
	out, err := git(dir, "reflog", "show", "--format=%gs", "refs/heads/"+branch)
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		return ""
	}
	lines := strings.Split(out, "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	const marker = "branch: Created from "
	if !strings.HasPrefix(last, marker) {
		return ""
	}
	from := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(last, marker)), "refs/heads/"), "refs/remotes/")
	if from == "" || from == "HEAD" || hexRe.MatchString(from) || from == branch || ValidBranchName(from) != nil {
		return ""
	}
	return from
}
