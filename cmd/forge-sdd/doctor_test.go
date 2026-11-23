package main

import (
	"bytes"
	"encoding/json"
	"io"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/doctor"
	"github.com/forge-sdd/cli/internal/scaffold"
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
	err := runDoctorHarness(&buf, io.Discard, t.TempDir(), true, deps)
	var ec exitCodeErr
	require.ErrorAs(t, err, &ec)
	assert.Equal(t, 3, ec.ExitCode())
	var rel doctor.Relatorio
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rel))
	require.NoError(t, doctor.ValidateRelatorio(rel))
	assert.Equal(t, 1, rel.Versao)

	// Modo texto lista cada checagem.
	buf.Reset()
	_ = runDoctorHarness(&buf, io.Discard, t.TempDir(), false, deps)
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

func depsHarnessTeste() doctor.Deps {
	return doctor.Deps{
		Versao: "5.0.0",
		Runner: func(string, ...string) ([]byte, error) { return nil, nil },
	}
}

func TestDoctorHarnessFalhaEscreveEmStderr(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		var out, errB bytes.Buffer
		err := runDoctorHarness(&out, &errB, t.TempDir(), asJSON, depsHarnessTeste())
		var ec exitCodeErr
		require.ErrorAs(t, err, &ec)
		assert.Equal(t, 3, ec.ExitCode())
		assert.Contains(t, errB.String(), "[falha] versao:")
		assert.Contains(t, errB.String(), "sugestão:")
		assert.Contains(t, errB.String(), "forge update")
		assert.NotContains(t, errB.String(), "{")
		if asJSON {
			var rel doctor.Relatorio
			require.NoError(t, json.Unmarshal(out.Bytes(), &rel))
		} else {
			assert.Contains(t, out.String(), "[falha] versao:")
		}
	}
}

func TestDoctorHarnessSaudavelNaoEscreveEmStderr(t *testing.T) {
	dir := t.TempDir()
	const ver = ".agents/skills/sdd-governanca/references/veredito.md"
	real, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(ver)))
	require.NoError(t, err)
	grava := func(rel, c string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(c), 0o644))
	}
	grava("sdd/.sddrc", `{"project":"p","version":"5.0.0"}`)
	grava("sdd/.sdd-version", "5.0.0\n")
	grava(ver, string(real))
	for _, e := range []string{"sdd-concepcao", "sdd-planejamento", "sdd-build", "sdd-governanca", "sdd-operacao"} {
		grava(".agents/skills/"+e+"/SKILL.md", "# "+e)
	}
	grava(".agents/rules/sdd-seguranca.md", "# regra")
	var out, errB bytes.Buffer
	d := depsHarnessTeste()
	d.Usuario = "eu"
	require.NoError(t, runDoctorHarness(&out, &errB, dir, false, d))
	assert.Empty(t, errB.String())
	assert.Contains(t, out.String(), "[ok] versao:")
}

func TestDoctorHarnessSemErroVazioNemUsoViaRootCmd(t *testing.T) {
	var out, errB bytes.Buffer
	_ = doctorCmd.Flags().Set("harness", "false")
	t.Cleanup(func() {
		_ = doctorCmd.Flags().Set("harness", "false")
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errB)
	rootCmd.SetArgs([]string{"doctor", "--harness", t.TempDir()})
	err := rootCmd.Execute()
	var ec exitCodeErr
	require.ErrorAs(t, err, &ec)
	assert.Equal(t, 3, ec.ExitCode())
	assert.Contains(t, errB.String(), "[falha] versao:")
	assert.NotContains(t, errB.String(), "Error:")
	assert.NotContains(t, errB.String(), "Usage:")
}

func TestBriefEHookMantemSaidaDeExitCode(t *testing.T) {
	assert.False(t, briefCmd.SilenceErrors)
	assert.False(t, briefCmd.SilenceUsage)
	dir := t.TempDir()
	out, err := runRoot(t, "brief", "--check", "--dir", dir)
	var ec exitCodeErr
	require.ErrorAs(t, err, &ec)
	assert.Equal(t, 3, ec.ExitCode())
	assert.Contains(t, out, "Usage:")
	assert.Contains(t, out, "Error:")
	// forge hook segue silencioso por conta própria.
	assert.Contains(t, hookPRGuardCmd.Short, "avisa")
}

// O init novo tem de sair saudável para o doctor: o teste de "doctor saudável" acima lê o
// veredito.md do repositório; este lê o que o próprio scaffold grava (fix-118-03).
func TestInitNovoPassaNoDoctorHarness(t *testing.T) {
	dir := t.TempDir()
	_, err := scaffold.Run(config.Config{Project: "demo", Stack: "go", Lang: "pt-BR", SddVersion: version, Agents: []string{config.AgentClaude}}, dir)
	require.NoError(t, err)
	var out, errB bytes.Buffer
	d := depsHarnessTeste()
	d.Versao = version
	d.Usuario = "eu"
	assert.NoError(t, runDoctorHarness(&out, &errB, dir, false, d), "saída: %s", out.String())
	assert.NotContains(t, out.String(), "[falha]")
}
