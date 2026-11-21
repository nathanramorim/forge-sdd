package pipeline

import (
	"fmt"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// ClaimFeature reivindica a feature criando a sua branch a partir de base. O claim é a existência
// da branch (Decisão xvi): o git recusa criar um nome que já existe, então o segundo claim falha.
// Uma referência remota já conhecida (origin/<branch>) também conta como claim feito. Sem rede:
// só olha refs locais. O estado da feature (state/<sid>.json) é gravado depois, na própria branch.
func ClaimFeature(dir, branch, base string) error {
	if err := queue.ValidBranchName(branch); err != nil || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("branch %q inválida para claim", branch)
	}
	if err := queue.ValidBranchName(base); err != nil || strings.HasPrefix(base, "-") {
		return fmt.Errorf("base %q inválida para claim", base)
	}
	for _, ref := range []string{"refs/heads/" + branch, "refs/remotes/origin/" + branch} {
		if _, err := layout.Git(dir, "show-ref", "--verify", "--quiet", ref); err == nil {
			return fmt.Errorf("feature já reivindicada: a branch %q já existe", branch)
		}
	}
	if _, err := layout.Git(dir, "branch", "--no-track", branch, base); err != nil {
		return fmt.Errorf("claim de %q falhou: %w", branch, err)
	}
	return nil
}
