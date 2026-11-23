package main

import (
	"bytes"
	"encoding/json"
	"github.com/forge-sdd/cli/internal/doctor"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctor_Constituicao(t *testing.T) {
	// 1. Projeto com Constituição em rascunho: avisa sem bloquear
	dirRascunho := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dirRascunho, "sdd", "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dirRascunho, "sdd", ".sddrc"), []byte(`{"project":"p","version":"4.1.0","agents":["claude"],"stack":"go","db":"none","lang":"pt-BR"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirRascunho, "sdd", "memory", "constitution.md"), []byte("# Constituição — P\n## Missão\n[Uma frase]\n"), 0o644))

	out := captureStdout(t, func() {
		_, err := runRoot(t, "doctor", dirRascunho)
		require.NoError(t, err)
	})
	assert.Contains(t, out, "Constituição em rascunho")
	assert.Contains(t, out, "/constitution")

	// 2. Projeto pronta: não avisa
	dirPronta := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dirPronta, "sdd", "memory"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dirPronta, "sdd", "spec"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", ".sddrc"), []byte(`{"project":"p","version":"4.1.0","agents":["claude"],"stack":"go","db":"none","lang":"pt-BR"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "memory", "constitution.md"), []byte("# Constituição — P\n## Missão\nMissão real.\n## Regras\n1. Regra 1.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "modules.md"), []byte("# Módulos\n## Core\nModulo core.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "flows.md"), []byte("# Fluxos\n## Fluxo 1\nFluxo real.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dirPronta, "sdd", "spec", "decisions.md"), []byte("# Decisões\n\n## Decisões Resolvidas\n| ID | Decisão | Escolha |\n|---|---|---|\n| D01 | Banco | SQLite |\n"), 0o644))

	outPronta := captureStdout(t, func() {
		_, err := runRoot(t, "doctor", dirPronta)
		require.NoError(t, err)
	})
	assert.Contains(t, outPronta, "Constituição (sdd/memory/constitution.md) ativa.")
	assert.NotContains(t, outPronta, "Constituição em rascunho")
}

func TestInit_PrimeiroPasso(t *testing.T) {
	// A saída do init lista /constitution antes do pipeline
	dir := t.TempDir()
	out := captureStdout(t, func() {
		_, err := runRoot(t, "init", dir, "--yes", "--name", "test-app", "--stack", "go", "--agent", "claude")
		require.NoError(t, err)
	})
	constIdx := strings.Index(out, "/constitution")
	pipeIdx := strings.Index(out, "pipeline")
	require.NotEqual(t, -1, constIdx, "deve citar /constitution nos próximos passos: %s", out)
	require.NotEqual(t, -1, pipeIdx, "deve citar pipeline")
	assert.Less(t, constIdx, pipeIdx, "/constitution deve aparecer antes do pipeline")

	// O menu lista /constitution em rascunho como passo 1
	p := &scripted{
		selects: map[string][]string{
			"Forge SDD": {"constitution", valExit},
		},
	}
	var menuOut bytes.Buffer
	menuEnv := &menuEnv{
		dir: dir,
		out: &menuOut,
		p:   p,
	}
	err := runMenu(menuEnv)
	require.NoError(t, err)
	assert.Contains(t, menuOut.String(), "Passo 1: Alinhar o projeto")
	assert.Contains(t, menuOut.String(), "/constitution")
}

func TestDoctorHarness_JSONECodigoDeSaida(t *testing.T) {
	deps := doctor.Deps{
		Versao: "5.0.0",
		Runner: func(string, ...string) ([]byte, error) { return nil, nil },
	}
	// Projeto sem sdd/.sddrc: a checagem de versão falha, logo o código é 3 e o JSON sai igual.
	var buf bytes.Buffer
	err := runDoctorHarness(&buf, t.TempDir(), true, deps)
	var ec exitCodeErr
	require.ErrorAs(t, err, &ec)
	assert.Equal(t, 3, ec.ExitCode())
	var rel doctor.Relatorio
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rel))
	require.NoError(t, doctor.ValidateRelatorio(rel))
	assert.Equal(t, 1, rel.Versao)

	// Modo texto lista cada checagem.
	buf.Reset()
	_ = runDoctorHarness(&buf, t.TempDir(), false, deps)
	assert.Contains(t, buf.String(), "[falha] versao:")
}

func TestDoctorAvisaRuleComVersaoNova(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte(`{"project":"p","version":"5.0.0","agents":["claude"],"stack":"go","db":"none","lang":"pt-BR"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".agents", "rules", "sdd-seguranca.md"), []byte("editada por mim\n"), 0o644))
	out := captureStdout(t, func() {
		_, err := runRoot(t, "doctor", dir)
		require.NoError(t, err)
	})
	assert.Contains(t, out, ".agents/rules/sdd-seguranca.md tem versão nova do forge")
}
