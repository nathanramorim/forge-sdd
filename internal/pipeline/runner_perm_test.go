package pipeline

import (
	"strings"
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

func TestModeloPorSessao_ArgvIdenticoSemEscolhaEComEscolhaNoClaude(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	p, _ := cfg.Profile(config.CLIClaude)
	base := RunSpec{Agent: "builder", Name: "forge-builder", Prompt: "faça", Permissions: cfg.WorkerPermissionArgs(config.CLIClaude)}

	_, semEscolha, err := BuildArgv(p, base)
	require.NoError(t, err)
	assert.NotContains(t, semEscolha, "--model")
	assert.NotContains(t, semEscolha, "--effort")
	assert.NotContains(t, semEscolha, config.MarkerModel)

	// O perfil antigo, sem marcadores, gera exatamente o mesmo argv.
	old := p
	old.Args = []string{"-p", "{prompt}", "--agent", "{agent}", "-n", "{name}", "--output-format", "stream-json", "--verbose", config.MarkerPermissions}
	_, antigo, err := BuildArgv(old, base)
	require.NoError(t, err)
	assert.Equal(t, antigo, semEscolha)

	cfg.Models = map[string]string{"builder": "sonnet"}
	cfg.Efforts = map[string]string{"builder": "low"}
	base.Model, base.Effort = cfg.SessionModelArgs("builder")
	_, comEscolha, err := BuildArgv(p, base)
	require.NoError(t, err)
	assert.Equal(t, append(append([]string{}, semEscolha...), "--model", "sonnet", "--effort", "low"), comEscolha)

	// Perfil gravado sem o marcador ignora a escolha (o preflight avisa).
	_, ignorado, err := BuildArgv(old, base)
	require.NoError(t, err)
	assert.Equal(t, antigo, ignorado)
}

func TestEsforco_PreflightAvisaPerfilSemMarcador(t *testing.T) {
	cfg := config.DefaultForgeConfig()
	cfg.Efforts = map[string]string{"reviewer": "low"}
	p := cfg.CLIProfiles[config.CLIClaude]
	p.Args = []string{"-p", "{prompt}", "--agent", "{agent}", "-n", "{name}", config.MarkerPermissions}
	cfg.CLIProfiles[config.CLIClaude] = p
	env := PreflightEnv{
		LookPath: func(n string) (string, error) { return "/usr/bin/" + n, nil },
		Exists:   func(string) bool { return true },
	}

	items := Preflight(t.TempDir(), cfg, env)

	var warned bool
	for _, it := range items {
		if it.Level == LevelWarn && it.Session == config.SessionRoleReviewer && strings.Contains(it.Message, config.MarkerEffort) {
			warned = true
		}
	}
	assert.True(t, warned, "o perfil sem {effort} deve gerar aviso")
}
