package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func specDir(t *testing.T, nome string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "sdd", "discovery", nome, "discovery.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte("# d\nSEGREDO-DO-CONTEUDO\n"), 0o644))
	return dir
}

func TestSpecMetricsIniciativaSaneada(t *testing.T) {
	// Pasta com ESC/ANSI não casa o padrão <id>-<nome>: o bloco não é gravado.
	dir := specDir(t, "999-a\x1b[31mb")
	sm, err := CollectSpecMetrics(dir, "sdd/discovery/999-a\x1b[31mb/discovery.md")
	assert.Nil(t, sm)
	assert.Error(t, err)

	// Pasta com mais de 80 caracteres também não entra.
	longo := "999-" + strings.Repeat("a", 90)
	dir = specDir(t, longo)
	sm, err = CollectSpecMetrics(dir, "sdd/discovery/"+longo+"/discovery.md")
	assert.Nil(t, sm)
	assert.Error(t, err)

	// Valor ruim vindo de fora (função pública) é descartado e nada com ESC chega ao arquivo.
	dir = t.TempDir()
	path, err := RecordSessionMetrics(dir, SessionMetrics{Spec: &SpecMetrics{Tipo: "fix", Iniciativa: "9-\x1b[31mx", Bytes: 1}}, time.Now())
	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "\x1b")
}

func TestSpecMetricsJSONSemCaminhoNemConteudo(t *testing.T) {
	dir := specDir(t, "999-teste")
	sm, err := CollectSpecMetrics(dir, "sdd/discovery/999-teste/discovery.md")
	require.NoError(t, err)
	data, err := json.Marshal(sm)
	require.NoError(t, err)
	s := string(data)
	assert.NotContains(t, s, dir)
	assert.NotContains(t, s, "sdd/discovery")
	assert.NotContains(t, s, "SEGREDO")
	assert.NotContains(t, s, "discovery.md")
}

func TestSpecMetricsCampoInvalidoNaoEntra(t *testing.T) {
	invalidos := []*SpecMetrics{
		{Tipo: "outro", Iniciativa: "1-a", Bytes: 1},
		{Tipo: "fix", Iniciativa: "", Bytes: 1},
		{Tipo: "fix", Iniciativa: strings.Repeat("a", 81), Bytes: 1},
		{Tipo: "fix", Iniciativa: "1-a", Bytes: -1},
		{Tipo: "fix", Iniciativa: "1-a", Bytes: 1, SecoesObrigatorias: 7},
		{Tipo: "discovery", Iniciativa: "1-a", Bytes: 1, PlanBytes: -3},
	}
	for _, sm := range invalidos {
		path, err := RecordSessionMetrics(t.TempDir(), SessionMetrics{Spec: sm}, time.Now())
		require.NoError(t, err)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"spec"`, "%+v", sm)
	}
	path, err := RecordSessionMetrics(t.TempDir(), SessionMetrics{Spec: &SpecMetrics{Tipo: "fix", Iniciativa: "1-a", Bytes: 1}}, time.Now())
	require.NoError(t, err)
	data, _ := os.ReadFile(path)
	assert.Contains(t, string(data), `"spec"`)
}

func TestSpecMetricsAbsolutoRecusado(t *testing.T) {
	dir := specDir(t, "999-teste")
	sm, err := CollectSpecMetrics(dir, filepath.Join(dir, "sdd/discovery/999-teste/discovery.md"))
	assert.Nil(t, sm)
	assert.Error(t, err)
}
