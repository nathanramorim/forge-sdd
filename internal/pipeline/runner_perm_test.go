package pipeline

import (
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildArgv_PorCLIeNivelDePermissao(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	for _, level := range []string{config.PermissionProject, config.PermissionFull} {
		cfg.Permissions = level
		for _, cli := range config.OfferedCLIs() {
			p, _ := cfg.Profile(cli)
			spec := RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "faça o job", Permissions: cfg.WorkerPermissionArgs(cli)}
			cmd, args, err := BuildArgv(p, spec)
			require.NoError(t, err, cli)
			wantCmd := cli
			if cli == config.CLIGemini {
				wantCmd = "agy"
			}
			assert.Equal(t, wantCmd, cmd)
			assert.NotContains(t, args, config.MarkerPermissions, "%s/%s: o marcador sempre é expandido", cli, level)
			for _, perm := range spec.Permissions {
				assert.Contains(t, args, perm, "%s/%s", cli, level)
			}
			assert.Contains(t, args, "faça o job", "o prompt é um único argumento: %s", cli)
		}
	}
	// Montagem exata do codex: exec, sandbox e prompt, sem -p.
	cfg.Permissions = config.PermissionProject
	p, _ := cfg.Profile(config.CLICodex)
	_, args, err := BuildArgv(p, RunSpec{Agent: "reviewer", Name: "forge-reviewer", Prompt: "revise", Permissions: cfg.WorkerPermissionArgs(config.CLICodex)})
	require.NoError(t, err)
	assert.Equal(t, []string{"exec", "--json", "--sandbox", "workspace-write", "revise"}, args)
}

func TestBuildArgv_PermissaoComControleERecusada(t *testing.T) {
	p := config.CLIProfile{Command: "codex", Args: []string{"exec", config.MarkerPermissions, "{prompt}"}}
	_, _, err := BuildArgv(p, RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "x", Permissions: []string{"--a\n--b"}})
	assert.Error(t, err)
	_, _, err = BuildArgv(p, RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "-rf"})
	assert.Error(t, err, "prompt que começa com - continua recusado")
}

func TestBuildArgv_AgyComAsFlagsConfirmadasNoUsoReal(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	p, _ := cfg.Profile(config.CLIAgy)
	for _, level := range []string{config.PermissionProject, config.PermissionFull} {
		cfg.Permissions = level
		cmd, args, err := BuildArgv(p, RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "faça o job", Permissions: cfg.WorkerPermissionArgs(config.CLIAgy)})
		require.NoError(t, err)
		assert.Equal(t, "agy", cmd)
		assert.Equal(t, []string{"-p", "faça o job", "--dangerously-skip-permissions"}, args, level)
	}
}
