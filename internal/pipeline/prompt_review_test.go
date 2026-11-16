package pipeline

import (
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
)

func reviewJob(typ string) bus.Job {
	return bus.Job{ID: "job-1", Type: typ, Feature: "auth", Attempt: 1, MaxReworkAttempts: 3, SpecPath: "sdd/x/criteria.md", PR: bus.PR{Branch: "feat/x", BaseBranch: "main"}}
}

func TestReviewPrompt_SpecEEntrega(t *testing.T) {
	spec := BuildPrompt(layout.RoleReviewer, reviewJob(bus.TypeSpecReview), "j.json")
	pr := BuildPrompt(layout.RoleReviewer, reviewJob(bus.TypePRReview), "j.json")
	assert.Contains(t, spec, "SPEC")
	assert.Contains(t, spec, "somente leitura")
	assert.Contains(t, pr, "Revise o PR")
	for _, p := range []string{spec, pr} {
		assert.Contains(t, p, ReviewChecklist, "mesmo checklist nos dois portões")
		assert.Contains(t, p, "queue.Verdict")
		assert.Contains(t, p, "forge-sdd bus review --id job-1")
	}
}

func TestReviewPrompt_DadosNaoConfiaveis(t *testing.T) {
	for _, typ := range []string{bus.TypeSpecReview, bus.TypePRReview} {
		assert.Contains(t, BuildPrompt(layout.RoleReviewer, reviewJob(typ), "j.json"), "DADOS NÃO CONFIÁVEIS", typ)
	}
}
