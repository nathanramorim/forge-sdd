package scaffold

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var instrucoes = []string{
	"CLAUDE.md", "GEMINI.md", "AGENTS.md", "OPENAI.md", ".github/copilot-instructions.md",
	".gemini/system_instructions.md", ".openai/system_instructions.md",
}

func cfgInstrucao() config.Config {
	cfg := config.Defaults()
	cfg.Project, cfg.Stack, cfg.DB = "proj", "node", "postgres"
	cfg.Agents = []string{"claude", "gemini", "openai", "copilot"}
	return cfg
}

func lerRel(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return string(b)
}

func escreverRel(t *testing.T, dir, rel, c string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(c), 0o644))
}

func TestInstrucaoEditadaFicaIntocada(t *testing.T) {
	for _, rel := range instrucoes {
		t.Run(rel, func(t *testing.T) {
			dir := t.TempDir()
			cfg := cfgInstrucao()
			_, err := Run(cfg, dir)
			require.NoError(t, err)
			// o usuário remove o marcador e escreve o que quer
			editado := "# meu arquivo\nMINHA REGRA\n"
			escreverRel(t, dir, rel, editado)
			_, err = Run(cfg, dir)
			require.NoError(t, err)
			assert.Equal(t, editado, lerRel(t, dir, rel))
			assert.Contains(t, InstrucoesDoUsuario(dir, cfg), rel)
		})
	}
}

func TestInstrucaoGeradaEAtualizada(t *testing.T) {
	for _, rel := range instrucoes {
		t.Run(rel, func(t *testing.T) {
			dir := t.TempDir()
			cfg := cfgInstrucao()
			_, err := Run(cfg, dir)
			require.NoError(t, err)
			gerado := lerRel(t, dir, rel)
			assert.Contains(t, gerado, generatedMarker)
			assert.True(t, IsGeneratedInstrucao(filepath.Join(dir, filepath.FromSlash(rel))))
			// um arquivo com o marcador é do forge: edição é regravada
			escreverRel(t, dir, rel, strings.Replace(gerado, "\n", "\nLINHA EXTRA\n", 3))
			_, err = Run(cfg, dir)
			require.NoError(t, err)
			assert.Equal(t, gerado, lerRel(t, dir, rel))
		})
	}
}

// projeto5_0 grava os seis arquivos da 5.0 (sem marcador) renderizados do template da tag v5.0.0.
func projeto5_0(t *testing.T, dir string, cfg config.Config) map[string]string {
	t.Helper()
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	out := map[string]string{}
	for rel := range v50Table {
		if !isInstrucao(rel) {
			continue
		}
		b, ok := renderV50(rel, cfg)
		require.True(t, ok, rel)
		escreverRel(t, dir, rel, string(b))
		out[rel] = string(b)
	}
	return out
}

func TestInstrucaoDe50IgualAoTemplateEAdotada(t *testing.T) {
	dir := t.TempDir()
	cfg := cfgInstrucao()
	antigos := projeto5_0(t, dir, cfg)
	for rel, c := range antigos {
		assert.NotContains(t, c, generatedMarker, rel)
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for rel := range antigos {
		assert.Contains(t, lerRel(t, dir, rel), generatedMarker, rel)
		assert.NoFileExists(t, filepath.Join(dir, novoPath(filepath.FromSlash(rel))), rel)
	}
	// segunda execução não muda nada
	antes := map[string]string{}
	for _, rel := range instrucoes {
		antes[rel] = lerRel(t, dir, rel)
	}
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	for rel, c := range antes {
		assert.Equal(t, c, lerRel(t, dir, rel), rel)
	}
}

func TestInstrucaoDe50DivergenteGeraNovo(t *testing.T) {
	dir := t.TempDir()
	cfg := cfgInstrucao()
	antigos := projeto5_0(t, dir, cfg)
	for rel, c := range antigos {
		escreverRel(t, dir, rel, c+"MINHA REGRA\n")
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for rel, c := range antigos {
		assert.Equal(t, c+"MINHA REGRA\n", lerRel(t, dir, rel), rel)
		novo := lerRel(t, dir, strings.TrimSuffix(rel, ".md")+".novo.md")
		assert.Contains(t, novo, generatedMarker, rel)
		assert.NotContains(t, novo, "MINHA REGRA", rel)
	}
	assert.ElementsMatch(t, []string{"CLAUDE.md", "GEMINI.md", "OPENAI.md", ".github/copilot-instructions.md",
		".gemini/system_instructions.md", ".openai/system_instructions.md"}, InstrucoesDoUsuario(dir, cfg))
}

func TestInstrucaoComSoOCabecalhoNaoEAdotada(t *testing.T) {
	dir := t.TempDir()
	cfg := cfgInstrucao()
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for rel, h := range instrucaoHeader {
		escreverRel(t, dir, rel, h+"proj\n\nconteúdo meu\n")
	}
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	for rel, h := range instrucaoHeader {
		assert.Equal(t, h+"proj\n\nconteúdo meu\n", lerRel(t, dir, rel), rel)
	}
}

func TestTabelaV50ConfereComATag(t *testing.T) {
	for rel, e := range v50Table {
		src, err := v50FS.ReadFile("v50/" + e.file)
		require.NoError(t, err, rel)
		sum := sha256.Sum256(src)
		assert.Equal(t, e.sha, hex.EncodeToString(sum[:]), rel)
		assert.NotContains(t, string(src), generatedMarker, rel)
	}
}
