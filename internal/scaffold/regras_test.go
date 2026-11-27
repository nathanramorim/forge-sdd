package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const regraRel = ".agents/rules/sdd-fluxo-e-portao.md"

// projeto50 cria um projeto e devolve o diretório com a rule e o schema trocados pelos da 5.0.
func projeto50(t *testing.T) (string, config.Config) {
	t.Helper()
	dir := t.TempDir()
	cfg := cfgInstrucao()
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for _, rel := range []string{regraRel, schemaMetricasRel} {
		b, ok := renderV50(rel, cfg)
		require.True(t, ok, rel)
		escreverRel(t, dir, rel, string(b))
	}
	return dir, cfg
}

func TestRuleNaoEditadaEAtualizada(t *testing.T) {
	dir, cfg := projeto50(t)
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	assert.Contains(t, lerRel(t, dir, regraRel), "Modo de trabalho")
	_, err = os.Stat(filepath.Join(dir, ".agents/rules/sdd-fluxo-e-portao.novo.md"))
	assert.True(t, os.IsNotExist(err))
	antes := lerRel(t, dir, regraRel)
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	assert.Equal(t, antes, lerRel(t, dir, regraRel))
}

func TestRuleEditadaFicaIntocadaEGeraNovo(t *testing.T) {
	dir, cfg := projeto50(t)
	editado := lerRel(t, dir, regraRel) + "\nMINHA LINHA\n"
	escreverRel(t, dir, regraRel, editado)
	for i := 0; i < 2; i++ {
		_, err := Run(cfg, dir)
		require.NoError(t, err)
		assert.Equal(t, editado, lerRel(t, dir, regraRel))
		assert.Contains(t, lerRel(t, dir, ".agents/rules/sdd-fluxo-e-portao.novo.md"), "Modo de trabalho")
	}
	_, manter := PlanoRegras(dir, cfg)
	assert.Contains(t, manter, regraRel)
}

func TestRuleDe50ReconhecidaPelaTabelaDeHashes(t *testing.T) {
	for rel := range v50Table {
		if !regraGerenciada(rel) {
			continue
		}
		cfg := cfgInstrucao()
		want, ok := renderV50(rel, cfg)
		require.True(t, ok, rel)
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(filepath.FromSlash(rel))), 0o755))
		escreverRel(t, dir, rel, string(want))
		assert.True(t, arquivoDoForge(rel, filepath.Join(dir, filepath.FromSlash(rel)), []byte("outro"), cfg), rel)
		escreverRel(t, dir, rel, string(want)+"x")
		assert.False(t, arquivoDoForge(rel, filepath.Join(dir, filepath.FromSlash(rel)), []byte("outro"), cfg), rel)
	}
}

func TestSchemaNaoEditadoEAtualizado(t *testing.T) {
	dir, cfg := projeto50(t)
	assert.NotContains(t, lerRel(t, dir, schemaMetricasRel), `"spec"`)
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	assert.Contains(t, lerRel(t, dir, schemaMetricasRel), `"spec"`)
}

func TestSchemaEditadoFicaIntocado(t *testing.T) {
	dir, cfg := projeto50(t)
	editado := lerRel(t, dir, schemaMetricasRel) + "\n"
	escreverRel(t, dir, schemaMetricasRel, editado)
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	assert.Equal(t, editado, lerRel(t, dir, schemaMetricasRel))
	assert.FileExists(t, filepath.Join(dir, "sdd/.metrics/schema.novo.json"))
}

func TestRuleDoUsuarioNuncaEMexida(t *testing.T) {
	dir, cfg := projeto50(t)
	escreverRel(t, dir, ".agents/rules/minha-regra.md", "minha\n")
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	assert.Equal(t, "minha\n", lerRel(t, dir, ".agents/rules/minha-regra.md"))
	_, err = os.Stat(filepath.Join(dir, ".agents/rules/minha-regra.novo.md"))
	assert.True(t, os.IsNotExist(err))
}

// fix-118-04: rules e schema do forge com CRLF (checkout no Windows) não contam como editados.
func TestRegrasComCRLFNaoViramEditadas(t *testing.T) {
	dir := t.TempDir()
	cfg := cfgInstrucao()
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	atualizar, manter := PlanoRegras(dir, cfg)
	require.Empty(t, manter)
	geridos := 0
	require.NoError(t, filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if !regraGerenciada(filepath.ToSlash(rel)) {
			return nil
		}
		geridos++
		b, rerr := os.ReadFile(p)
		require.NoError(t, rerr)
		return os.WriteFile(p, []byte(strings.ReplaceAll(string(b), "\n", "\r\n")), 0o644)
	}))
	require.Positive(t, geridos)
	atualizar, manter = PlanoRegras(dir, cfg)
	assert.Empty(t, manter, "CRLF não é edição do usuário")
	assert.Empty(t, atualizar, "conteúdo igual ao template: nada a atualizar")
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, ".agents/rules/sdd-fluxo-e-portao.novo.md"))
	assert.True(t, os.IsNotExist(err), "sem .novo para rule só com CRLF")
}
