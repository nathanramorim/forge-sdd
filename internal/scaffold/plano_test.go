package scaffold

import (
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hashesDoProjeto(t *testing.T, dir string) map[string][32]byte {
	t.Helper()
	m := map[string][32]byte{}
	require.NoError(t, filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, rerr := os.ReadFile(p)
		require.NoError(t, rerr)
		rel, _ := filepath.Rel(dir, p)
		m[filepath.ToSlash(rel)] = sha256.Sum256(b)
		return nil
	}))
	return m
}

func TestPlanoDoUpdateIgualAoRealSemGravar(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Project: "demo", SddVersion: "5.0.0", Lang: "pt-BR", Agents: []string{config.AgentClaude, config.AgentGemini}, NamingConvention: "sequencial"}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for rel, c := range map[string]string{
		"GEMINI.md":                         "# Minhas regras\n",
		".agents/rules/sdd-seguranca.md":    "editada\n",
		".agents/skills/sdd-build/SKILL.md": "defasada\n",
		".sddrc":                            "",
	} {
		if rel == ".sddrc" {
			continue
		}
		require.NoError(t, os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(c), 0o644))
	}
	cfg.SddVersion = "5.1.0"

	antes := hashesDoProjeto(t, dir)
	plano, err := PlanoDoUpdate(dir, cfg)
	require.NoError(t, err)
	assert.Equal(t, antes, hashesDoProjeto(t, dir), "o plano não grava nada")

	_, err = Run(cfg, dir)
	require.NoError(t, err)
	depois := hashesDoProjeto(t, dir)

	emPlano := map[string]string{}
	for _, it := range plano {
		emPlano[it.Rel] = it.Acao
		switch it.Acao {
		case AcaoCriar:
			_, existia := antes[it.Rel]
			assert.False(t, existia, it.Rel)
			assert.Contains(t, depois, it.Rel)
		case AcaoReescrever:
			assert.NotEqual(t, antes[it.Rel], depois[it.Rel], "%s era para ser reescrito", it.Rel)
		case AcaoIgual, AcaoPreservar:
			assert.Equal(t, antes[it.Rel], depois[it.Rel], "%s não podia mudar (%s)", it.Rel, it.Acao)
		}
	}
	assert.Equal(t, AcaoPreservar, emPlano["GEMINI.md"])
	assert.Equal(t, AcaoPreservar, emPlano[".agents/rules/sdd-seguranca.md"])
	assert.Equal(t, AcaoReescrever, emPlano[".agents/skills/sdd-build/SKILL.md"])

	// Tudo o que mudou no disco e é dos templates estava no plano como criar/reescrever.
	for rel, h := range depois {
		if antes[rel] == h || strings.Contains(rel, ".novo.") || rel == ".claude/settings.json" || rel == "scripts/pr_guard.py" {
			continue
		}
		a := emPlano[rel]
		assert.Truef(t, a == AcaoCriar || a == AcaoReescrever, "%s mudou no real mas o plano disse %q", rel, a)
	}
}
