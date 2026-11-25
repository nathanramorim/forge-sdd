package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// Fluxo da 6.0: a pessoa conversa na sessão principal; a Discovery é um agente que ela chama. Ao fim
// da descoberta a spec vai sozinha ao Reviewer (spec_review "de discovery"); o veredito volta pelo
// inbox e quem decide é a pessoa. Aprovada, o /split-features quebra em tasks e despacha direto, sem
// nova revisão de spec antes da fila.

// prefixoRevisaoDiscovery marca o spec_review pedido ao fim da Discovery (fora do modo avançado).
const prefixoRevisaoDiscovery = "disc-"

func isRevisaoDiscovery(j bus.Job) bool { return strings.HasPrefix(j.ID, prefixoRevisaoDiscovery) }

// RevisarDiscovery enfileira a revisão da spec da iniciativa (discovery.md e criteria.md) no Reviewer.
// É idempotente por instante e vale em qualquer autonomia.
func RevisarDiscovery(dir, initiative string, now time.Time) (*bus.Job, string, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, "", err
	}
	if !found {
		return nil, "", ErrNotPipelineMode
	}
	if !layout.ValidInitiative(initiative) {
		return nil, "", fmt.Errorf("iniciativa %q inválida", initiative)
	}
	base := filepath.Join(layout.Discovery(dir), initiative)
	for _, f := range []string{"discovery.md", "criteria.md"} {
		if _, err := os.Stat(filepath.Join(base, f)); err != nil {
			return nil, "", fmt.Errorf("falta %s em sdd/discovery/%s: termine a descoberta antes de pedir a revisão", f, initiative)
		}
	}
	feature := initiative
	if i := strings.Index(initiative, "-"); i >= 0 && featureValida(initiative[i+1:]) {
		feature = initiative[i+1:]
	}
	attempt := nextSpecAttempt(dir, initiative, cfg.Reviewer.MaxReworkAttempts)
	id := prefixoRevisaoDiscovery + strings.TrimPrefix(jobID(initiative, now), "job-")
	if len(id) > 80 {
		id = strings.TrimRight(id[:80], "-.")
	}
	job := bus.Job{
		SchemaVersion:     bus.SchemaVersion,
		ID:                id,
		Type:              bus.TypeSpecReview,
		Priority:          bus.PriorityNormal,
		Feature:           feature,
		Title:             initiativeTitle(base, feature),
		Initiative:        initiative,
		TaskIDs:           []string{},
		SpecPath:          filepath.ToSlash(filepath.Join("sdd", "discovery", initiative, "criteria.md")),
		ADRRefs:           []string{},
		RulesRefs:         []string{},
		PR:                bus.PR{Branch: layout.InitiativeBranch(initiative), BaseBranch: cfg.BaseBranch},
		Attempt:           attempt,
		MaxReworkAttempts: cfg.Reviewer.MaxReworkAttempts,
		ReworkFeedback:    []bus.Feedback{},
	}
	path, _, err := bus.Enqueue(dir, layout.RoleReviewer, job, now)
	if err != nil {
		return nil, "", err
	}
	return &job, path, nil
}

func featureValida(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}
