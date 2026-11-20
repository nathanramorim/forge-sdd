package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const specSeisSecoes = "# t\n\n## Modelo do módulo\n\n## Decisões do projeto aplicadas\n\n## O que já existe\n\n## Dependências novas\n\n## Rules e IDs da Constituição Aplicáveis\n\n## Critérios de Aceitação Executáveis\n"

// runSessionRecord executa `forge session record` zerando antes o estado das flags (o cobra as
// mantém entre execuções) e devolve stdout e stderr. Não depende do git nem do forge do PATH.
func runSessionRecord(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	sessionRecordCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	var out, errb bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errb)
	t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
	rootCmd.SetArgs(append([]string{"session", "record"}, args...))
	err := rootCmd.Execute()
	return out.String(), errb.String(), err
}

func writeSpecFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
}

func specProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeTestSddrc(t, dir, true)
	base := "sdd/discovery/999-teste/"
	writeSpecFile(t, dir, base+"discovery.md", "# discovery\n")
	writeSpecFile(t, dir, base+"criteria.md", "criteria")
	writeSpecFile(t, dir, base+"plan.md", "plan!")
	writeSpecFile(t, dir, base+"status.md", "st")
	writeSpecFile(t, dir, base+"features/feat-999-01-a.md", specSeisSecoes)
	writeSpecFile(t, dir, base+"features/feat-999-02-b.md", "# so um\n## Modelo do módulo\n")
	writeSpecFile(t, dir, base+"fixes/fix-999-01-c.md", specSeisSecoes)
	return dir
}

func lerSessaoUnica(t *testing.T, dir string) (map[string]any, []byte) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "sdd", ".metrics", "session-*.json"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	data, err := os.ReadFile(files[0])
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	return m, data
}

func gravaSpec(t *testing.T, feature string, extra ...string) (map[string]any, string) {
	t.Helper()
	dir := specProject(t)
	args := append([]string{"--dir", dir, "--feature", feature, "--outcome", "approved", "--criterio-atendido=true"}, extra...)
	_, stderr, err := runSessionRecord(t, args...)
	require.NoError(t, err)
	m, _ := lerSessaoUnica(t, dir)
	return m, stderr
}

func TestSessionRecordDiscoveryGravaSpec(t *testing.T) {
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/discovery.md", "--sessao", "spec")
	spec, ok := m["spec"].(map[string]any)
	require.True(t, ok, "esperava o bloco spec")
	assert.Equal(t, "discovery", spec["tipo"])
	assert.Equal(t, "999-teste", spec["iniciativa"])
	assert.EqualValues(t, len("# discovery\n"), spec["bytes"])
	assert.EqualValues(t, 8, spec["criteria_bytes"])
	assert.EqualValues(t, 5, spec["plan_bytes"])
	assert.EqualValues(t, 2, spec["status_bytes"])
	assert.EqualValues(t, 2, spec["features_count"])
	assert.EqualValues(t, 1, spec["fixes_count"])
}

func TestSessionRecordNovaFeatureGravaSpec(t *testing.T) {
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/features/feat-999-01-a.md", "--sessao", "spec")
	spec, ok := m["spec"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "feature", spec["tipo"])
	assert.EqualValues(t, 6, spec["secoes_obrigatorias"])
	assert.NotContains(t, spec, "criteria_bytes")
}

func TestSessionRecordNovoFixGravaSpec(t *testing.T) {
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/fixes/fix-999-01-c.md", "--sessao", "spec")
	spec, ok := m["spec"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "fix", spec["tipo"])
	assert.EqualValues(t, 6, spec["secoes_obrigatorias"])
}

func TestSpecMetricsContaFeaturesEFixes(t *testing.T) {
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/features/feat-999-02-b.md", "--sessao", "spec")
	assert.EqualValues(t, 1, m["spec"].(map[string]any)["secoes_obrigatorias"])
	m, _ = gravaSpec(t, "sdd/discovery/999-teste/discovery.md", "--sessao", "spec")
	spec := m["spec"].(map[string]any)
	assert.EqualValues(t, 2, spec["features_count"])
	assert.EqualValues(t, 1, spec["fixes_count"])
}

func TestSessionRecordSplitFeaturesNaoGravaSpec(t *testing.T) {
	// /split-features não passa a flag e seu --feature é uma pasta.
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/features/")
	assert.NotContains(t, m, "spec")
}

func TestSessionRecordBuilderNaoGravaSpecMesmoEmFeatures(t *testing.T) {
	for _, f := range []string{"sdd/discovery/999-teste/features/feat-999-01-a.md", "sdd/discovery/999-teste/fixes/fix-999-01-c.md"} {
		m, _ := gravaSpec(t, f, "--phase", "999-01", "--agent-path", "lider,builder,governance")
		assert.NotContains(t, m, "spec", f)
	}
}

func TestSessionRecordSemFlagSessaoNaoGravaSpec(t *testing.T) {
	m, _ := gravaSpec(t, "sdd/discovery/999-teste/discovery.md")
	assert.NotContains(t, m, "spec")
}

func TestSessionRecordFlagComPastaNaoGravaSpec(t *testing.T) {
	m, stderr := gravaSpec(t, "sdd/discovery/999-teste/features/", "--sessao", "spec")
	assert.NotContains(t, m, "spec")
	assert.Contains(t, stderr, "aviso")
}

func TestSessionRecordSpecComPontosPontoNaoGrava(t *testing.T) {
	m, stderr := gravaSpec(t, "sdd/discovery/999-teste/../999-teste/discovery.md", "--sessao", "spec")
	assert.NotContains(t, m, "spec")
	assert.Contains(t, stderr, "aviso")
}

func TestSessionRecordSpecSimbolicoNaoSegue(t *testing.T) {
	dir := specProject(t)
	out := filepath.Join(t.TempDir(), "externo.md")
	require.NoError(t, os.WriteFile(out, []byte(specSeisSecoes), 0o644))
	link := filepath.Join(dir, "sdd/discovery/999-teste/features/feat-999-09-link.md")
	if err := os.Symlink(out, link); err != nil {
		t.Skip("link simbólico indisponível:", err)
	}
	_, stderr, err := runSessionRecord(t, "--dir", dir, "--feature", "sdd/discovery/999-teste/features/feat-999-09-link.md",
		"--sessao", "spec", "--outcome", "approved")
	require.NoError(t, err)
	m, _ := lerSessaoUnica(t, dir)
	assert.NotContains(t, m, "spec")
	assert.Contains(t, stderr, "aviso")
}

func TestSessionRecordSpecForaDeDiscoveryNaoGrava(t *testing.T) {
	dir := specProject(t)
	writeSpecFile(t, dir, "sdd/features/feat-x.md", specSeisSecoes)
	for _, f := range []string{"sdd/features/feat-x.md", "/etc/passwd"} {
		_, stderr, err := runSessionRecord(t, "--dir", dir, "--feature", f, "--sessao", "spec", "--outcome", "approved")
		require.NoError(t, err)
		assert.Contains(t, stderr, "aviso")
	}
	files, _ := filepath.Glob(filepath.Join(dir, "sdd", ".metrics", "session-*.json"))
	for _, p := range files {
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"spec"`)
	}
	assert.Len(t, files, 2)
}

func TestSessionRecordSpecPastaAusenteGravaSemSpec(t *testing.T) {
	m, stderr := gravaSpec(t, "sdd/discovery/888-nao-existe/discovery.md", "--sessao", "spec")
	assert.NotContains(t, m, "spec")
	assert.Contains(t, stderr, "aviso")
}

func TestSessionRecordDryRunNaoGrava(t *testing.T) {
	dir := t.TempDir()
	writeTestSddrc(t, dir, true)
	writeSpecFile(t, dir, "sdd/discovery/999-teste/discovery.md", "# d\n")
	stdout, _, err := runSessionRecord(t, "--dir", dir, "--feature", "sdd/discovery/999-teste/discovery.md",
		"--sessao", "spec", "--outcome", "approved", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"spec"`)
	_, statErr := os.Stat(filepath.Join(dir, "sdd", ".metrics"))
	assert.True(t, os.IsNotExist(statErr), "--dry-run não pode criar sdd/.metrics")
}

func TestSessionRecordFalhaDeGravacaoDevolveErro(t *testing.T) {
	dir := specProject(t)
	externo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	if err := os.Symlink(externo, filepath.Join(dir, "sdd", ".metrics")); err != nil {
		t.Skip("link simbólico indisponível:", err)
	}
	_, _, err := runSessionRecord(t, "--dir", dir, "--feature", "sdd/discovery/999-teste/discovery.md",
		"--sessao", "spec", "--outcome", "approved")
	require.Error(t, err)
	entries, _ := os.ReadDir(externo)
	assert.Empty(t, entries, "nada pode ser gravado fora do projeto")
}

func TestSessionRecordArquivoAntigoSemSpecAindaLe(t *testing.T) {
	dir := t.TempDir()
	metrics := filepath.Join(dir, "sdd", ".metrics")
	require.NoError(t, os.MkdirAll(metrics, 0o755))
	old := `{"$schema":"forge-sdd/metrics/1.0","feature":"x.md","agent_path":null,"outcome":"approved","complexidade":null}`
	require.NoError(t, os.WriteFile(filepath.Join(metrics, "session-2026-01-01T000000Z.json"), []byte(old), 0o644))
	s, err := AggregateSessionMetrics(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Approved)
}

func TestLeitoresDeSessaoToleramSpec(t *testing.T) {
	dir := specProject(t)
	_, _, err := runSessionRecord(t, "--dir", dir, "--feature", "sdd/discovery/999-teste/discovery.md", "--sessao", "spec", "--outcome", "approved")
	require.NoError(t, err)
	s, err := AggregateSessionMetrics(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Total)
	assert.Equal(t, 1, s.Approved)
	m, data := lerSessaoUnica(t, dir)
	assert.Contains(t, m, "spec")
	assert.False(t, strings.Contains(string(data), dir), "o JSON não pode carregar o caminho do projeto")
}
