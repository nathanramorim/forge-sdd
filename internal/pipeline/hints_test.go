package pipeline

import (
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBuilderHints_SegueAConfiguracao(t *testing.T) {
	job := bus.Job{ID: "job-1"}
	on := BuilderHints(config.BuilderConfig{MaxParallelTasks: 3, UseGitWorktrees: true, RunTestsBeforePR: true}, job)
	assert.Contains(t, on, "pipeline worktree add --id job-1")
	assert.Contains(t, on, "Obrigatório")
	off := BuilderHints(config.BuilderConfig{MaxParallelTasks: 2}, job)
	assert.NotContains(t, off, "worktree add")
	assert.Contains(t, off, "desligado")
	assert.True(t, strings.Contains(off, "subagentes"))
}

func TestAdrTokensIn_AceitaCaixaEEspaco(t *testing.T) {
	assert.Equal(t, []string{"adr-0003", "adr-0004-x"}, adrTokensIn("ADR 0003 e ADR-0004-x"))
}

func TestFindADR_NaoCasaPrefixoDoNomeDoArquivo(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/spec/adr/adr.md", "x")
	_, ok := findADR(dir, []string{"sdd/spec/adr"}, "adr-0003")
	assert.False(t, ok)
}
