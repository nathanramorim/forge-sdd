package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
)

var (
	wtIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,80}$`)
	refRe  = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)
)

// WorktreePath devolve .forge/runtime/worktrees/<job>/<task>, validando os componentes (SEG-01).
func WorktreePath(dir, jobID, taskID string) (string, error) {
	if !wtIDRe.MatchString(jobID) || !wtIDRe.MatchString(taskID) {
		return "", fmt.Errorf("job %q ou task %q inválido para worktree", jobID, taskID)
	}
	return filepath.Join(layout.Worktrees(dir), jobID, taskID), nil
}

// CreateWorktree cria um worktree isolado para a task a partir de baseRef, em uma branch própria
// forge/<job>-<task>. Todos os argumentos do git são explícitos.
func CreateWorktree(dir, jobID, taskID, baseRef string) (string, error) {
	path, err := WorktreePath(dir, jobID, taskID)
	if err != nil {
		return "", err
	}
	if !refRe.MatchString(baseRef) || strings.HasPrefix(baseRef, "-") || strings.Contains(baseRef, "..") {
		return "", fmt.Errorf("ref base %q inválida", baseRef)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	branch := "forge/" + jobID + "-" + taskID
	if _, err := layout.Git(dir, "worktree", "add", "-b", branch, path, baseRef); err != nil {
		return "", fmt.Errorf("criar worktree: %w", err)
	}
	return path, nil
}

// RemoveWorktree remove o worktree e a branch auxiliar. O caminho precisa estar dentro de
// .forge/runtime/worktrees (links simbólicos são resolvidos antes de conferir).
func RemoveWorktree(dir, jobID, taskID string) error {
	path, err := WorktreePath(dir, jobID, taskID)
	if err != nil {
		return err
	}
	root, err := filepath.EvalSymlinks(layout.Worktrees(dir))
	if err != nil {
		return nil // nada a remover
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil // já removido
	}
	rel, err := filepath.Rel(root, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("worktree %q fora de %s: recusado", path, layout.Worktrees(dir))
	}
	if _, err := layout.Git(dir, "worktree", "remove", "--force", real); err != nil {
		return fmt.Errorf("remover worktree: %w", err)
	}
	_, _ = layout.Git(dir, "branch", "-D", "forge/"+jobID+"-"+taskID)
	_ = os.Remove(filepath.Dir(path)) // remove a pasta do job quando vazia
	return nil
}
