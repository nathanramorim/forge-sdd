package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeForge(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
}

func TestLoadForgeConfig_SemArquivoEhModoLegado(t *testing.T) {
	cfg, found, err := LoadForgeConfig(t.TempDir())
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, DefaultForgeConfig(), cfg)
}

func TestLoadForgeConfig_PadroesEArquivoParcial(t *testing.T) {
	dir := t.TempDir()
	writeForge(t, dir, ".forge/config/forge.json", `{"cli":"agy","builder":{"max_parallel_tasks":5,"use_git_worktrees":false,"run_tests_before_pr":true,"job_timeout_minutes":10}}`)
	cfg, found, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, CLIAgy, cfg.CLI)
	assert.Equal(t, 5, cfg.Builder.MaxParallelTasks)
	assert.False(t, cfg.Builder.UseGitWorktrees)
	assert.Equal(t, 3, cfg.Reviewer.MaxReworkAttempts, "campos omitidos mantêm o padrão")
	assert.Equal(t, "In Progress", cfg.PMMCP.StatusOnImplement)
}

func TestLoadForgeConfig_Violacoes(t *testing.T) {
	bad := map[string]string{
		"campo desconhecido":      `{"cli":"claude","api_key":"<SEU_TOKEN_AQUI>"}`,
		"cli inválido":            `{"cli":"gpt"}`,
		"paralelismo zero":        `{"builder":{"max_parallel_tasks":0,"job_timeout_minutes":60}}`,
		"paralelismo acima":       `{"builder":{"max_parallel_tasks":9,"job_timeout_minutes":60}}`,
		"tentativas zero":         `{"reviewer":{"max_rework_attempts":0}}`,
		"tentativas acima":        `{"reviewer":{"max_rework_attempts":11}}`,
		"json inválido":           `{"cli":`,
		"conteúdo extra":          `{"cli":"claude"} {"x":1}`,
		"provider fora da lista":  `{"pm_mcp":{"provider":"trello","supported":["clickup"],"status_on_implement":"a","status_on_pr_approved":"b","evidence_max_age_minutes":5}}`,
		"adr com ..":              `{"adr_dirs":["../fora"]}`,
		"adr absoluto":            `{"adr_dirs":["/etc"]}`,
		"base branch com opção":   `{"base_branch":"--upload-pack=x"}`,
		"comando com espaço":      `{"cli_profiles":{"claude":{"command":"claude -p","args":[]}}}`,
		"comando com shell":       `{"cli_profiles":{"claude":{"command":"claude;rm","args":[]}}}`,
		"marcador desconhecido":   `{"cli_profiles":{"claude":{"command":"claude","args":["{secret}"]}}}`,
		"argumento com controle":  `{"cli_profiles":{"claude":{"command":"claude","args":["a\nb"]}}}`,
		"comando vazio":           `{"cli":"agy","cli_profiles":{"agy":{"command":"","args":[]}}}`,
		"comando como opção":      `{"cli_profiles":{"claude":{"command":"-rf","args":[]}}}`,
		"evidência fora da faixa": `{"pm_mcp":{"provider":"auto","supported":["jira"],"status_on_implement":"a","status_on_pr_approved":"b","evidence_max_age_minutes":0}}`,
	}
	for name, content := range bad {
		dir := t.TempDir()
		writeForge(t, dir, ".forge/config/forge.json", content)
		_, found, err := LoadForgeConfig(dir)
		require.Error(t, err, name)
		assert.True(t, found, name)
		assert.NotContains(t, err.Error(), "SEU_TOKEN_AQUI", "%s: o erro não deve repetir o segredo", name)
	}
}

func TestLoadNotificationsConfig(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadNotificationsConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, DefaultNotificationsConfig(), cfg)
	assert.True(t, cfg.Channels.DiscoverySession.Enabled)
	assert.False(t, cfg.Channels.Email.Enabled)

	writeForge(t, dir, ".forge/config/notifications.json", `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"smtp.exemplo.com","port":587,"user_env":"SMTP_USER","password_env":"SMTP_PASS","to":["a@b.com"],"from":"forge@b.com"},"webhook":{"enabled":true,"provider":"slack","url_env":"FORGE_WEBHOOK_URL"}}}`)
	cfg, err = LoadNotificationsConfig(dir)
	require.NoError(t, err)
	assert.True(t, cfg.Channels.Email.Enabled)
	assert.Equal(t, []string{"a@b.com"}, cfg.Channels.Email.To)
	assert.True(t, cfg.Channels.DiscoverySession.Enabled, "padrão preservado quando o canal é omitido")
}

func TestLoadNotificationsConfig_Violacoes(t *testing.T) {
	bad := map[string]string{
		"url literal":              `{"channels":{"webhook":{"enabled":true,"provider":"slack","url_env":"https://hooks.slack.com/services/T000/B000/XXXX"}}}`,
		"campo url":                `{"channels":{"webhook":{"enabled":true,"provider":"slack","url":"https://x"}}}`,
		"senha literal":            `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h","port":25,"password":"hunter2","to":["a@b.com"],"from":"x@y.com"}}}`,
		"user_env com valor":       `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h","port":25,"user_env":"joao@exemplo.com","to":["a@b.com"],"from":"x@y.com"}}}`,
		"canal padrão desligado":   `{"channels":{"discovery_session":{"enabled":false,"desktop_toast":false}}}`,
		"default_channel inválido": `{"default_channel":"email"}`,
		"email sem destinatário":   `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h","port":25,"to":[],"from":"x@y.com"}}}`,
		"email destino inválido":   `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h","port":25,"to":["sem-arroba"],"from":"x@y.com"}}}`,
		"host com barra":           `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h/x","port":25,"to":["a@b.com"],"from":"x@y.com"}}}`,
		"porta inválida":           `{"channels":{"email":{"enabled":true,"provider":"smtp","host":"h","port":70000,"to":["a@b.com"],"from":"x@y.com"}}}`,
		"webhook sem env":          `{"channels":{"webhook":{"enabled":true,"provider":"slack","url_env":""}}}`,
		"provider inválido":        `{"channels":{"webhook":{"enabled":true,"provider":"icq","url_env":"X"}}}`,
	}
	for name, content := range bad {
		dir := t.TempDir()
		writeForge(t, dir, ".forge/config/notifications.json", content)
		_, err := LoadNotificationsConfig(dir)
		require.Error(t, err, name)
		assert.NotContains(t, err.Error(), "hooks.slack.com", name)
		assert.NotContains(t, err.Error(), "hunter2", name)
	}
}

func TestPathsDoForgeSaoOsDoLayout(t *testing.T) {
	dir := t.TempDir()
	writeForge(t, dir, strings.TrimPrefix(layout.ForgeConfigFile(dir), dir+string(filepath.Separator)), `{"cli":"claude"}`)
	_, found, err := LoadForgeConfig(dir)
	require.NoError(t, err)
	assert.True(t, found)
}
