package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
)

// Perfis de CLI aceitos pelo pipeline (feat 70).
const (
	CLIClaude = "claude"
	CLIAgy    = "agy"
)

// Modos de sessão do pipeline (feat 77).
const (
	SessionModeHeadless = "headless"
	SessionModeLive     = "live"
)

// Modos de contexto entre jobs (feat 77).
const (
	ContextClear      = "clear"
	ContextContinuous = "continuous"
)

// ValidSessionMode diz se m é um modo de sessão aceito.
func ValidSessionMode(m string) bool {
	return m == SessionModeHeadless || m == SessionModeLive
}

// ValidContextMode diz se c é um modo de contexto aceito.
func ValidContextMode(c string) bool {
	return c == ContextClear || c == ContextContinuous
}

// LiveConfig define parâmetros de prontidão para o modo live (feat 77).
type LiveConfig struct {
	StartupWaitSeconds int `json:"startup_wait_seconds,omitempty"`
	ReadyQuietSeconds  int `json:"ready_quiet_seconds,omitempty"`
}

// CLIProfile descreve como invocar um agente de forma não interativa. Os argumentos aceitam os
// marcadores {agent}, {name} e {prompt}; a execução usa argv explícito, nunca shell (SEG-03).
type CLIProfile struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	// InteractiveArgs abre a sessão interativa (Discovery). Vazio em um forge.json da 4.1: o padrão
	// da CLI é aplicado por ForgeConfig.Profile.
	InteractiveArgs []string `json:"interactive_args,omitempty"`
	// ResetCommand é o comando enviado à CLI interativa para limpar o contexto entre jobs (ex: /clear, /new).
	ResetCommand string `json:"reset_command,omitempty"`
}

// PMConfig configura o gerenciador de projeto acessado via MCP pelo agente.
type PMConfig struct {
	Provider              string   `json:"provider"`
	Supported             []string `json:"supported"`
	RequireExistingCard   bool     `json:"require_existing_card"`
	StatusOnImplement     string   `json:"status_on_implement"`
	StatusOnPRApproved    string   `json:"status_on_pr_approved"`
	EvidenceMaxAgeMinutes int      `json:"evidence_max_age_minutes"`
}

// BuilderConfig configura a sessão Builder.
type BuilderConfig struct {
	MaxParallelTasks  int  `json:"max_parallel_tasks"`
	UseGitWorktrees   bool `json:"use_git_worktrees"`
	RunTestsBeforePR  bool `json:"run_tests_before_pr"`
	JobTimeoutMinutes int  `json:"job_timeout_minutes"`
}

// ReviewerConfig configura a sessão Reviewer.
type ReviewerConfig struct {
	MaxReworkAttempts int  `json:"max_rework_attempts"`
	CommentOnPR       bool `json:"comment_on_pr"`
}

// ForgeConfig é o conteúdo de .forge/config/forge.json. A existência do arquivo ativa o modo pipeline.
type ForgeConfig struct {
	CLI string `json:"cli"`
	// Sessions escolhe o agente de cada sessão (discovery, builder, reviewer). Sem a chave, vale cli.
	Sessions    map[string]string `json:"sessions,omitempty"`
	Permissions string            `json:"permissions"`
	// Terminal escolhe onde as sessões abrem: auto (tmux se existir), tmux ou native (janelas do sistema).
	Terminal string `json:"terminal"`
	// SessionMode escolhe o modo do worker: headless (padrão antigo) ou live (CLI aberta no painel).
	SessionMode string `json:"session_mode,omitempty"`
	// Context escolhe a gestão de contexto entre jobs no modo live: clear (padrão) ou continuous.
	Context string `json:"context,omitempty"`
	// Live define parâmetros de tempo do modo live.
	Live LiveConfig `json:"live,omitempty"`
	// AllowCommands são os comandos de shell liberados no nível project (nome do comando apenas).
	AllowCommands []string              `json:"allow_commands"`
	CLIProfiles   map[string]CLIProfile `json:"cli_profiles"`
	PMMCP         PMConfig              `json:"pm_mcp"`
	Builder       BuilderConfig         `json:"builder"`
	Reviewer      ReviewerConfig        `json:"reviewer"`
	ADRDirs       []string              `json:"adr_dirs"`
	BaseBranch    string                `json:"base_branch"`
}

var (
	envNameRe     = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	commandRe     = regexp.MustCompile(`^[A-Za-z0-9._/\\-]+$`)
	placeholderRe = regexp.MustCompile(`\{[^}]*\}`)
	relDirRe      = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)
	branchRe      = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)
	emailRe       = regexp.MustCompile(`^[^@\s]+@[^@\s]+$`)
	validMarkers  = map[string]bool{"{agent}": true, "{name}": true, "{prompt}": true, MarkerPermissions: true}
)

// DefaultForgeConfig devolve os valores padrão da especificação do pipeline. Os argumentos do
// perfil do agy são os documentados no discovery 70 e estão marcados "a confirmar".
func DefaultForgeConfig() ForgeConfig {
	return ForgeConfig{
		CLI:           CLIClaude,
		Permissions:   PermissionProject,
		Terminal:      TerminalAuto,
		SessionMode:   SessionModeHeadless,
		Context:       ContextClear,
		Live:          LiveConfig{StartupWaitSeconds: 8, ReadyQuietSeconds: 3},
		AllowCommands: []string{"git", "gh", "forge-sdd"},
		CLIProfiles:   defaultProfiles(),
		PMMCP: PMConfig{
			Provider:              "auto",
			Supported:             []string{"clickup", "github-projects", "azure-devops", "jira"},
			RequireExistingCard:   true,
			StatusOnImplement:     "In Progress",
			StatusOnPRApproved:    "Ready for Review",
			EvidenceMaxAgeMinutes: 30,
		},
		Builder:    BuilderConfig{MaxParallelTasks: 3, UseGitWorktrees: true, RunTestsBeforePR: true, JobTimeoutMinutes: 60},
		Reviewer:   ReviewerConfig{MaxReworkAttempts: 3, CommentOnPR: true},
		ADRDirs:    []string{"sdd/spec/adr"},
		BaseBranch: "main",
	}
}

// LoadForgeConfig lê .forge/config/forge.json. found=false (sem erro) significa que o arquivo não
// existe e o projeto segue o modo da 4.0.0-beta. Campo desconhecido é erro: o esquema estrito
// impede segredo literal em campo inexistente (SEG-04).
func LoadForgeConfig(dir string) (cfg ForgeConfig, found bool, err error) {
	path := layout.ForgeConfigFile(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultForgeConfig(), false, nil
		}
		return ForgeConfig{}, false, fmt.Errorf("ler %s: %w", path, err)
	}
	cfg = DefaultForgeConfig()
	if err := strictDecode(data, &cfg); err != nil {
		return ForgeConfig{}, true, fmt.Errorf("%s inválido: %w", path, err)
	}
	if cfg.SessionMode == "" {
		cfg.SessionMode = SessionModeHeadless
	}
	if cfg.Context == "" {
		cfg.Context = ContextClear
	}
	if cfg.Live.StartupWaitSeconds == 0 {
		cfg.Live.StartupWaitSeconds = 8
	}
	if cfg.Live.ReadyQuietSeconds == 0 {
		cfg.Live.ReadyQuietSeconds = 3
	}
	if err := cfg.Validate(); err != nil {
		return ForgeConfig{}, true, fmt.Errorf("%s inválido: %w", path, err)
	}
	return cfg, true, nil
}

func strictDecode(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("conteúdo extra após o objeto JSON")
	}
	return nil
}

// Validate confere os limites e formatos do pipeline (SEG-01).
func (c ForgeConfig) Validate() error {
	if !KnownCLI(c.CLI) {
		return fmt.Errorf("cli %q inválido (use %s ou agy)", c.CLI, strings.Join(OfferedCLIs(), ", "))
	}
	if _, ok := c.CLIProfiles[c.CLI]; !ok {
		return fmt.Errorf("cli_profiles sem perfil para %q", c.CLI)
	}
	for name, p := range c.CLIProfiles {
		if err := validateProfile(name, p); err != nil {
			return err
		}
	}
	if err := c.validateSessions(); err != nil {
		return err
	}
	if c.SessionMode != "" && !ValidSessionMode(c.SessionMode) {
		return fmt.Errorf("session_mode %q inválido (use %s ou %s)", c.SessionMode, SessionModeLive, SessionModeHeadless)
	}
	if c.Context != "" && !ValidContextMode(c.Context) {
		return fmt.Errorf("context %q inválido (use %s ou %s)", c.Context, ContextClear, ContextContinuous)
	}
	if c.Live.StartupWaitSeconds < 0 || c.Live.StartupWaitSeconds > 60 {
		return fmt.Errorf("live.startup_wait_seconds %d fora da faixa 0 a 60", c.Live.StartupWaitSeconds)
	}
	if c.Live.ReadyQuietSeconds < 0 || c.Live.ReadyQuietSeconds > 30 {
		return fmt.Errorf("live.ready_quiet_seconds %d fora da faixa 0 a 30", c.Live.ReadyQuietSeconds)
	}
	if c.Builder.MaxParallelTasks < 1 || c.Builder.MaxParallelTasks > 8 {
		return fmt.Errorf("builder.max_parallel_tasks %d fora da faixa 1 a 8", c.Builder.MaxParallelTasks)
	}
	if c.Builder.JobTimeoutMinutes < 1 || c.Builder.JobTimeoutMinutes > 720 {
		return fmt.Errorf("builder.job_timeout_minutes %d fora da faixa 1 a 720", c.Builder.JobTimeoutMinutes)
	}
	if c.Reviewer.MaxReworkAttempts < 1 || c.Reviewer.MaxReworkAttempts > 10 {
		return fmt.Errorf("reviewer.max_rework_attempts %d fora da faixa 1 a 10", c.Reviewer.MaxReworkAttempts)
	}
	if c.PMMCP.EvidenceMaxAgeMinutes < 1 || c.PMMCP.EvidenceMaxAgeMinutes > 1440 {
		return fmt.Errorf("pm_mcp.evidence_max_age_minutes %d fora da faixa 1 a 1440", c.PMMCP.EvidenceMaxAgeMinutes)
	}
	if len(c.PMMCP.Supported) == 0 {
		return errors.New("pm_mcp.supported não pode ser vazio")
	}
	if c.PMMCP.Provider != "auto" && !contains(c.PMMCP.Supported, c.PMMCP.Provider) {
		return fmt.Errorf("pm_mcp.provider %q fora de pm_mcp.supported", c.PMMCP.Provider)
	}
	if strings.TrimSpace(c.PMMCP.StatusOnImplement) == "" || strings.TrimSpace(c.PMMCP.StatusOnPRApproved) == "" {
		return errors.New("pm_mcp.status_on_implement e pm_mcp.status_on_pr_approved são obrigatórios")
	}
	if len(c.ADRDirs) == 0 {
		return errors.New("adr_dirs não pode ser vazio")
	}
	for _, d := range c.ADRDirs {
		if !relDirRe.MatchString(d) || containsDotDot(d) {
			return fmt.Errorf("adr_dirs: %q deve ser um caminho relativo ao projeto", d)
		}
	}
	if !branchRe.MatchString(c.BaseBranch) || containsDotDot(c.BaseBranch) || strings.HasPrefix(c.BaseBranch, "-") {
		return fmt.Errorf("base_branch %q inválida", c.BaseBranch)
	}
	return nil
}

func validateProfile(name string, p CLIProfile) error {
	if !commandRe.MatchString(p.Command) || strings.HasPrefix(p.Command, "-") {
		return fmt.Errorf("cli_profiles.%s.command %q inválido (sem espaços nem metacaracteres de shell)", name, p.Command)
	}
	for field, args := range map[string][]string{"args": p.Args, "interactive_args": p.InteractiveArgs} {
		for _, a := range args {
			for _, m := range placeholderRe.FindAllString(a, -1) {
				if !validMarkers[m] {
					return fmt.Errorf("cli_profiles.%s.%s: marcador %s desconhecido (use {agent}, {name}, {prompt} ou {permissions})", name, field, m)
				}
			}
			if strings.Contains(a, MarkerPermissions) && a != MarkerPermissions {
				return fmt.Errorf("cli_profiles.%s.%s: %s precisa ser um argumento inteiro", name, field, MarkerPermissions)
			}
			if strings.ContainsAny(a, "\x00\r\n") {
				return fmt.Errorf("cli_profiles.%s.%s: argumento com caractere de controle", name, field)
			}
		}
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func containsDotDot(p string) bool {
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// ---- notificações ----

// DiscoverySessionChannel é o canal padrão: inbox da sessão Discovery.
type DiscoverySessionChannel struct {
	Enabled      bool `json:"enabled"`
	DesktopToast bool `json:"desktop_toast"`
}

// EmailChannel envia por SMTP; credenciais só por variável de ambiente.
type EmailChannel struct {
	Enabled     bool     `json:"enabled"`
	Provider    string   `json:"provider"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	UserEnv     string   `json:"user_env"`
	PasswordEnv string   `json:"password_env"`
	To          []string `json:"to"`
	From        string   `json:"from"`
}

// WebhookChannel posta em uma URL lida de variável de ambiente.
type WebhookChannel struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	URLEnv   string `json:"url_env"`
}

// PMCommentChannel pede ao agente (que alcança o MCP) para comentar no card.
type PMCommentChannel struct {
	Enabled bool `json:"enabled"`
}

// NotificationChannels agrupa os canais.
type NotificationChannels struct {
	DiscoverySession DiscoverySessionChannel `json:"discovery_session"`
	Email            EmailChannel            `json:"email"`
	Webhook          WebhookChannel          `json:"webhook"`
	PMComment        PMCommentChannel        `json:"pm_comment"`
}

// NotificationsConfig é o conteúdo de .forge/config/notifications.json.
type NotificationsConfig struct {
	DefaultChannel string               `json:"default_channel"`
	Channels       NotificationChannels `json:"channels"`
}

// DefaultNotificationsConfig devolve os padrões da especificação.
func DefaultNotificationsConfig() NotificationsConfig {
	return NotificationsConfig{
		DefaultChannel: "discovery_session",
		Channels: NotificationChannels{
			DiscoverySession: DiscoverySessionChannel{Enabled: true, DesktopToast: true},
			Email:            EmailChannel{Provider: "smtp", Port: 587, To: []string{}},
			Webhook:          WebhookChannel{Provider: "slack", URLEnv: "FORGE_WEBHOOK_URL"},
		},
	}
}

// LoadNotificationsConfig lê .forge/config/notifications.json; sem o arquivo devolve os padrões.
func LoadNotificationsConfig(dir string) (NotificationsConfig, error) {
	path := layout.NotificationsFile(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultNotificationsConfig(), nil
		}
		return NotificationsConfig{}, fmt.Errorf("ler %s: %w", path, err)
	}
	cfg := DefaultNotificationsConfig()
	if err := strictDecode(data, &cfg); err != nil {
		return NotificationsConfig{}, fmt.Errorf("%s inválido: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return NotificationsConfig{}, fmt.Errorf("%s inválido: %w", path, err)
	}
	return cfg, nil
}

// Validate confere formatos e a regra de segredos só por ambiente (SEG-04).
func (n NotificationsConfig) Validate() error {
	if n.DefaultChannel != "discovery_session" {
		return fmt.Errorf("default_channel %q inválido (use discovery_session)", n.DefaultChannel)
	}
	if !n.Channels.DiscoverySession.Enabled {
		return errors.New("channels.discovery_session não pode ser desligado: é o canal padrão obrigatório")
	}
	e := n.Channels.Email
	if e.Enabled {
		if e.Provider != "smtp" {
			return fmt.Errorf("channels.email.provider %q inválido (use smtp)", e.Provider)
		}
		if strings.TrimSpace(e.Host) == "" || strings.ContainsAny(e.Host, " /\\@\r\n") {
			return fmt.Errorf("channels.email.host %q inválido", e.Host)
		}
		if e.Port < 1 || e.Port > 65535 {
			return fmt.Errorf("channels.email.port %d fora da faixa 1 a 65535", e.Port)
		}
		if !emailRe.MatchString(e.From) {
			return fmt.Errorf("channels.email.from %q inválido", e.From)
		}
		if len(e.To) == 0 {
			return errors.New("channels.email.to não pode ser vazio quando o canal está ligado")
		}
		for _, to := range e.To {
			if !emailRe.MatchString(to) {
				return fmt.Errorf("channels.email.to: %q inválido", to)
			}
		}
	}
	// SEG-04: vale também com o canal desligado, para nenhum segredo literal ficar versionado.
	for field, v := range map[string]string{"user_env": e.UserEnv, "password_env": e.PasswordEnv} {
		if v != "" && !envNameRe.MatchString(v) {
			return fmt.Errorf("channels.email.%s deve ser o NOME de uma variável de ambiente, não o valor (SEG-04)", field)
		}
	}
	w := n.Channels.Webhook
	if w.Enabled {
		if w.Provider != "slack" && w.Provider != "generic" {
			return fmt.Errorf("channels.webhook.provider %q inválido (use slack ou generic)", w.Provider)
		}
	}
	if w.URLEnv != "" && !envNameRe.MatchString(w.URLEnv) {
		return errors.New("channels.webhook.url_env deve ser o NOME de uma variável de ambiente, não a URL (SEG-04)")
	}
	if w.Enabled && w.URLEnv == "" {
		return errors.New("channels.webhook.url_env é obrigatório quando o canal está ligado")
	}
	return nil
}
