package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func taskFile(t *testing.T, dir, name, depends string) {
	t.Helper()
	testutil.WriteFile(t, dir, name+".md", fmt.Sprintf("# %s\n\n**Branch:** feat/x\n**Depende de:** %s\n**Status:** approved\n", name, depends))
}

func ids(batches [][]Task) [][]string {
	var out [][]string
	for _, b := range batches {
		var row []string
		for _, t := range b {
			row = append(row, t.ID)
		}
		out = append(out, row)
	}
	return out
}

func TestPlanWaves_OrdenaPorDependenciaEParalelizaOIndependente(t *testing.T) {
	dir := t.TempDir()
	taskFile(t, dir, "feat-70-01-base", "—")
	taskFile(t, dir, "feat-70-02-bus", "feat-70-01")
	taskFile(t, dir, "feat-70-06-notify", "feat-70-01")
	taskFile(t, dir, "feat-70-03-gate", "feat-70-01, feat-70-02")
	taskFile(t, dir, "feat-70-04-worker", "feat-70-02, feat-70-03")

	batches, err := PlanWaves(dir, 3)
	require.NoError(t, err)
	assert.Equal(t, [][]string{
		{"feat-70-01"},
		{"feat-70-02", "feat-70-06"},
		{"feat-70-03"},
		{"feat-70-04"},
	}, ids(batches))
	assert.Equal(t, 1, batches[1][0].Level)
	assert.Contains(t, FormatWaves(batches), "em paralelo): feat-70-02, feat-70-06")
}

func TestPlanWaves_GrafoIndependenteDivideEmLotesDeMaxParallel(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 7; i++ {
		taskFile(t, dir, fmt.Sprintf("feat-80-%02d-t", i), "—")
	}
	batches, err := PlanWaves(dir, 3)
	require.NoError(t, err)
	require.Len(t, batches, 3)
	assert.Len(t, batches[0], 3)
	assert.Len(t, batches[1], 3)
	assert.Len(t, batches[2], 1)

	batches, err = PlanWaves(dir, 0)
	require.NoError(t, err)
	assert.Len(t, batches, 7, "limite inválido vira 1")
}

func TestPlanWaves_Violacoes(t *testing.T) {
	cases := map[string]func(dir string){
		"ciclo": func(dir string) {
			taskFile(t, dir, "feat-90-01-a", "feat-90-02")
			taskFile(t, dir, "feat-90-02-b", "feat-90-01")
		},
		"dependência inexistente": func(dir string) { taskFile(t, dir, "feat-90-01-a", "feat-90-09") },
		"depende de si":           func(dir string) { taskFile(t, dir, "feat-90-01-a", "feat-90-01") },
		"task duplicada": func(dir string) {
			taskFile(t, dir, "feat-90-01-a", "—")
			taskFile(t, dir, "feat-90-01-b", "—")
		},
		"sem tasks": func(dir string) { testutil.WriteFile(t, dir, "leiame.md", "x") },
	}
	for name, setup := range cases {
		dir := t.TempDir()
		setup(dir)
		_, err := PlanWaves(dir, 3)
		require.Error(t, err, name)
	}
	_, err := PlanWaves(filepath.Join(t.TempDir(), "nao-existe"), 3)
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), "panic"))
}

func TestPlanWaves_LeDependeDeApenasDoCabecalho(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "feat-70-01-a.md", "# a\n**Depende de:** —\n")
	body := "# b\n**Depende de:** feat-70-01\n" + strings.Repeat("texto\n", 60) + "**Depende de:** feat-99-99\n"
	testutil.WriteFile(t, dir, "feat-70-02-b.md", body)
	batches, err := PlanWaves(dir, 3)
	require.NoError(t, err, "uma menção posterior no corpo não conta como dependência")
	assert.Equal(t, [][]string{{"feat-70-01"}, {"feat-70-02"}}, ids(batches))
}
