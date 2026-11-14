package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildArgv_PerfisClaudeEAgy(t *testing.T) {
	def := config.DefaultForgeConfig()
	spec := RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "Execute o job x"}

	cmd, args, err := BuildArgv(def.CLIProfiles[config.CLIClaude], spec)
	require.NoError(t, err)
	assert.Equal(t, "claude", cmd)
	assert.Equal(t, []string{"-p", "Execute o job x", "--agent", "builder", "-n", "forge-builder", "--output-format", "stream-json", "--verbose"}, args)

	cmd, args, err = BuildArgv(def.CLIProfiles[config.CLIAgy], spec)
	require.NoError(t, err)
	assert.Equal(t, "agy", cmd)
	assert.Equal(t, []string{"-p", "Execute o job x"}, args, "agy: -p confirmado no uso real; sem permissões no spec o marcador expande para nada")
}

func TestBuildArgv_PromptComShellMetacaracteresEhUmUnicoArgumento(t *testing.T) {
	p := config.CLIProfile{Command: "claude", Args: []string{"-p", "{prompt}"}}
	evil := "Execute; rm -rf ~ && echo $(whoami) `id` | tee /tmp/x\nlinha 2"
	_, args, err := BuildArgv(p, RunSpec{Agent: "builder", Name: "forge-builder", Prompt: evil})
	require.NoError(t, err)
	require.Len(t, args, 2, "o prompt inteiro é um argumento, sem divisão por espaço ou ;")
	assert.Equal(t, evil, args[1])
}

func TestBuildArgv_Violacoes(t *testing.T) {
	p := config.DefaultForgeConfig().CLIProfiles[config.CLIClaude]
	for name, spec := range map[string]RunSpec{
		"agente com espaço":     {Agent: "build er", Name: "forge-builder", Prompt: "x"},
		"agente com ;":          {Agent: "builder;rm", Name: "forge-builder", Prompt: "x"},
		"agente vazio":          {Agent: "", Name: "forge-builder", Prompt: "x"},
		"sessão com barra":      {Agent: "builder", Name: "a/b", Prompt: "x"},
		"prompt vira opção":     {Agent: "builder", Name: "forge-builder", Prompt: "--dangerously-skip-permissions"},
		"prompt com espaço-opt": {Agent: "builder", Name: "forge-builder", Prompt: "  -p x"},
		"prompt com NUL":        {Agent: "builder", Name: "forge-builder", Prompt: "a\x00b"},
	} {
		_, _, err := BuildArgv(p, spec)
		assert.Error(t, err, name)
	}
}

func fakeCLI(t *testing.T, script string) config.CLIProfile {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("script de shell de teste")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fakeagent")
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755))
	return config.CLIProfile{Command: path, Args: []string{"{prompt}"}}
}

func TestExecRunner_ExecutaSaidaECodigo(t *testing.T) {
	profile := fakeCLI(t, `echo "olá $1"; exit 3`)
	var out strings.Builder
	res, err := ExecRunner{Profile: profile}.Run(context.Background(), RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "mundo", Dir: t.TempDir()}, &out)
	require.NoError(t, err)
	assert.Equal(t, 3, res.ExitCode)
	assert.False(t, res.TimedOut)
	assert.Contains(t, out.String(), "olá mundo")
}

func TestExecRunner_TimeoutMataOProcesso(t *testing.T) {
	profile := fakeCLI(t, `sleep 30`)
	start := time.Now()
	res, err := ExecRunner{Profile: profile}.Run(context.Background(), RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "x", Dir: t.TempDir(), Timeout: 200 * time.Millisecond}, &strings.Builder{})
	require.NoError(t, err)
	assert.True(t, res.TimedOut)
	assert.Less(t, time.Since(start), 15*time.Second)
}

func TestExecRunner_CLIInexistenteTemMensagemClara(t *testing.T) {
	_, err := ExecRunner{Profile: config.CLIProfile{Command: "forge-cli-que-nao-existe", Args: []string{}}}.Run(context.Background(), RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "x"}, &strings.Builder{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "não encontrada no PATH")
}
