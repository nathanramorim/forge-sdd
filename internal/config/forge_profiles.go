package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// Agentes de CLI oferecidos pelo pipeline por sessão (feat 71). O agy continua aceito se já estiver
// em forge.json (feat 70), mas não faz parte dos agentes oferecidos.
const (
	CLIGemini  = "gemini"
	CLICopilot = "copilot"
	CLICodex   = "codex"
)

// Papéis de sessão do pipeline.
const (
	SessionRoleDiscovery = "discovery"
	SessionRoleBuilder   = "builder"
	SessionRoleReviewer  = "reviewer"
)

// Níveis de permissão do worker (feat 71, D-4).
const (
	PermissionProject = "project" // edição e comandos do fluxo (allowlist)
	PermissionFull    = "full"    // bypass documentado da CLI; só em ambiente isolado
)

// MarkerPermissions é expandido pelo runner em zero ou mais argumentos (PermissionArgs).
const MarkerPermissions = "{permissions}"

// OfferedCLIs são os agentes que o seletor oferece, na ordem do menu.
func OfferedCLIs() []string { return []string{CLIClaude, CLIGemini, CLICopilot, CLICodex} }

// SessionRoles são os papéis na ordem em que o menu pergunta.
func SessionRoles() []string {
	return []string{SessionRoleDiscovery, SessionRoleBuilder, SessionRoleReviewer}
}

func knownCLI(cli string) bool { return cli == CLIAgy || contains(OfferedCLIs(), cli) }

var allowCommandRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,40}$`)

// defaultProfiles reúne os perfis pesquisados na documentação oficial das CLIs (discovery 71).
// Cada perfil é escrito por inteiro: -p é o prompt em claude, gemini e copilot, mas é perfil no
// codex (o modo não interativo dele é `codex exec`).
// A CONFIRMAR (não executados de ponta a ponta): gemini, copilot, codex e agy. O copilot não usa
// --agent por padrão porque as pastas de agentes não foram confirmadas; o papel vai no prompt.
func defaultProfiles() map[string]CLIProfile {
	return map[string]CLIProfile{
		CLIClaude: {
			Command:         "claude",
			Args:            []string{"-p", "{prompt}", "--agent", "{agent}", "-n", "{name}", MarkerPermissions},
			InteractiveArgs: []string{"-n", "{name}", "--agent", "{agent}"},
		},
		CLIGemini: {
			Command:         "gemini",
			Args:            []string{"-p", "{prompt}", MarkerPermissions},
			InteractiveArgs: []string{"-i", "{prompt}"},
		},
		CLICopilot: {
			Command:         "copilot",
			Args:            []string{"-p", "{prompt}", "-n", "{name}", "--no-ask-user", MarkerPermissions},
			InteractiveArgs: []string{"-i", "{prompt}", "-n", "{name}"},
		},
		CLICodex: {
			Command:         "codex",
			Args:            []string{"exec", MarkerPermissions, "{prompt}"},
			InteractiveArgs: []string{"{prompt}"},
		},
		CLIAgy: {
			Command:         "agy",
			Args:            []string{"--agent", "{agent}", "-i", "{prompt}"},
			InteractiveArgs: []string{"--agent", "{agent}", "-i"},
		},
	}
}

// PermissionArgs traduz o nível de permissão em argumentos da CLI (tabela única, DUP-01). allow
// são os comandos de shell liberados no nível project onde a CLI aceita allowlist.
// A CONFIRMAR: gemini no nível project libera só edição (comandos de shell pedem confirmação);
// codex libera a escrita no workspace pela sandbox.
func PermissionArgs(cli, level string, allow []string) []string {
	switch level {
	case PermissionFull:
		switch cli {
		case CLIClaude:
			return []string{"--dangerously-skip-permissions"}
		case CLIGemini:
			return []string{"--approval-mode", "yolo"}
		case CLICopilot:
			return []string{"--allow-all-tools"}
		case CLICodex:
			return []string{"--dangerously-bypass-approvals-and-sandbox"}
		}
	case PermissionProject:
		switch cli {
		case CLIClaude:
			args := []string{"--permission-mode", "acceptEdits", "--allowedTools", "Read", "Edit", "Write", "Glob", "Grep"}
			for _, c := range allow {
				args = append(args, "Bash("+c+" *)")
			}
			return args
		case CLIGemini:
			return []string{"--approval-mode", "auto_edit"}
		case CLICopilot:
			args := []string{"--allow-tool=write"}
			for _, c := range allow {
				args = append(args, "--allow-tool=shell("+c+":*)")
			}
			return args
		case CLICodex:
			return []string{"--sandbox", "workspace-write"}
		}
	}
	return nil
}

// PermissionSummary descreve em uma linha o que o nível libera, para o menu mostrar antes de abrir.
func PermissionSummary(level string, allow []string) string {
	if level == PermissionFull {
		return "TOTAL: sem pedir confirmação (bypass da CLI). Use só em ambiente isolado."
	}
	return "PROJETO: edita arquivos e roda só estes comandos: " + strings.Join(allow, ", ") + "."
}

// CLIFor devolve o agente da sessão: sessions[papel] ou, sem ela, o cli único da 4.1.
func (c ForgeConfig) CLIFor(role string) string {
	if cli, ok := c.Sessions[role]; ok && cli != "" {
		return cli
	}
	return c.CLI
}

// Profile devolve o perfil de uma CLI completando o que um forge.json da 4.1 não trazia
// (interactive_args), para a atualização não quebrar projetos existentes.
func (c ForgeConfig) Profile(cli string) (CLIProfile, bool) {
	p, ok := c.CLIProfiles[cli]
	if !ok {
		return CLIProfile{}, false
	}
	if len(p.InteractiveArgs) == 0 {
		p.InteractiveArgs = defaultProfiles()[cli].InteractiveArgs
	}
	return p, true
}

// WorkerPermissionArgs devolve os argumentos de permissão para uma CLI com a configuração atual.
func (c ForgeConfig) WorkerPermissionArgs(cli string) []string {
	return PermissionArgs(cli, c.Permissions, c.AllowCommands)
}

func (c ForgeConfig) validateSessions() error {
	for role, cli := range c.Sessions {
		if !contains(SessionRoles(), role) {
			return fmt.Errorf("sessions: papel %q desconhecido (use discovery, builder ou reviewer)", role)
		}
		if !knownCLI(cli) {
			return fmt.Errorf("sessions.%s: agente %q inválido (use %s)", role, cli, strings.Join(OfferedCLIs(), ", "))
		}
		if _, ok := c.CLIProfiles[cli]; !ok {
			return fmt.Errorf("sessions.%s: cli_profiles sem perfil para %q", role, cli)
		}
	}
	if c.Permissions != PermissionProject && c.Permissions != PermissionFull {
		return fmt.Errorf("permissions %q inválido (use project ou full)", c.Permissions)
	}
	for _, a := range c.AllowCommands {
		if !allowCommandRe.MatchString(a) {
			return fmt.Errorf("allow_commands: %q inválido (só o nome do comando, sem espaços nem metacaracteres)", a)
		}
	}
	return nil
}

// SaveForgeConfig grava .forge/config/forge.json de forma atômica (sem arquivo parcial em falha).
func SaveForgeConfig(dir string, cfg ForgeConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("forge.json: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar forge.json: %w", err)
	}
	path := layout.ForgeConfigFile(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return queue.WriteFileAtomic(path, append(data, '\n'), 0o644)
}
