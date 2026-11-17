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

// Destinos das sessões do pipeline (feat 72).
const (
	TerminalAuto   = "auto"
	TerminalTmux   = "tmux"
	TerminalNative = "native"
)

// ValidTerminal diz se v é um destino aceito.
func ValidTerminal(v string) bool {
	return v == TerminalAuto || v == TerminalTmux || v == TerminalNative
}

// MarkerModel e MarkerEffort expandem em zero ou mais argumentos, como MarkerPermissions: sem
// escolha de modelo ou esforço o argv fica idêntico ao de antes (ADR-0014 D-5).
const (
	MarkerModel  = "{model}"
	MarkerEffort = "{effort}"
)

// MarkerPermissions é expandido pelo runner em zero ou mais argumentos (PermissionArgs).
const MarkerPermissions = "{permissions}"

// OfferedCLIs são os agentes que o seletor oferece, na ordem do menu.
func OfferedCLIs() []string { return []string{CLIClaude, CLIGemini, CLICopilot, CLICodex} }

// SessionRoles são os papéis na ordem em que o menu pergunta.
func SessionRoles() []string {
	return []string{SessionRoleDiscovery, SessionRoleBuilder, SessionRoleReviewer}
}

// KnownCLI diz se o agente é aceito em forge.json (os oferecidos mais o agy da 4.1).
func KnownCLI(cli string) bool { return cli == CLIAgy || contains(OfferedCLIs(), cli) }

var allowCommandRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,40}$`)

// defaultProfiles reúne os perfis pesquisados na documentação oficial das CLIs (discovery 71).
// Cada perfil é escrito por inteiro: -p é o prompt em claude, gemini e copilot, mas é perfil no
// codex (o modo não interativo dele é `codex exec`).
// A CONFIRMAR (não executados de ponta a ponta): gemini, copilot, codex e agy. As flags de saída
// estruturada (claude --output-format stream-json --verbose, gemini -o stream-json, codex --json)
// alimentam a visão limpa do worker; sem elas a visão cai no acompanhamento por git. O copilot não usa
// --agent por padrão porque as pastas de agentes não foram confirmadas; o papel vai no prompt.
func defaultProfiles() map[string]CLIProfile {
	return map[string]CLIProfile{
		CLIClaude: {
			Command:         "claude",
			Args:            []string{"-p", "{prompt}", "--agent", "{agent}", "-n", "{name}", "--output-format", "stream-json", "--verbose", MarkerPermissions, MarkerModel, MarkerEffort},
			InteractiveArgs: []string{"-n", "{name}", "--agent", "{agent}", MarkerModel, MarkerEffort},
			ResetCommand:    "/clear",
		},
		CLIGemini: {
			Command:         "agy",
			Args:            []string{"-p", "{prompt}", MarkerPermissions, MarkerModel, MarkerEffort},
			InteractiveArgs: []string{"-i", "{prompt}", MarkerModel, MarkerEffort},
			ResetCommand:    "/clear",
		},
		CLICopilot: {
			Command:         "copilot",
			Args:            []string{"-p", "{prompt}", "-n", "{name}", "--no-ask-user", MarkerPermissions},
			InteractiveArgs: []string{"-i", "{prompt}", "-n", "{name}"},
			ResetCommand:    "/clear",
		},
		CLICodex: {
			Command:         "codex",
			Args:            []string{"exec", "--json", MarkerPermissions, "{prompt}"},
			InteractiveArgs: []string{"{prompt}"},
			ResetCommand:    "/new",
		},
		// agy: `-p` e `--dangerously-skip-permissions` confirmados no uso real (barbearia-app, 2026-11-14).
		CLIAgy: {
			Command:         "agy",
			Args:            []string{"-p", "{prompt}", MarkerPermissions, MarkerModel, MarkerEffort},
			InteractiveArgs: []string{"-i", "{prompt}", MarkerModel, MarkerEffort},
			ResetCommand:    "/clear",
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
		case CLIClaude, CLIAgy, CLIGemini:
			return []string{"--dangerously-skip-permissions"}
		case CLICopilot:
			return []string{"--allow-all-tools"}
		case CLICodex:
			return []string{"--dangerously-bypass-approvals-and-sandbox"}
		}
	case PermissionProject:
		switch cli {
		case CLIAgy, CLIGemini:
			// Sem allowlist confirmada: o agy só roda o worker sem confirmação com o bypass (igual a full).
			return []string{"--dangerously-skip-permissions"}
		case CLIClaude:
			args := []string{"--permission-mode", "acceptEdits", "--allowedTools", "Read", "Edit", "Write", "Glob", "Grep"}
			for _, c := range allow {
				args = append(args, "Bash("+c+" *)")
			}
			return args
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
// Normaliza também "gemini" legado (Command: "gemini") para "agy" (fix 78).
func (c ForgeConfig) Profile(cli string) (CLIProfile, bool) {
	p, ok := c.CLIProfiles[cli]
	if !ok {
		if cli == CLIGemini {
			p, ok = c.CLIProfiles[CLIAgy]
		} else if cli == CLIAgy {
			p, ok = c.CLIProfiles[CLIGemini]
		}
		if !ok {
			return CLIProfile{}, false
		}
	}
	if cli == CLIGemini || cli == CLIAgy {
		if p.Command == "gemini" {
			p.Command = "agy"
		}
		for _, a := range p.Args {
			if a == "stream-json" || a == "-o" {
				p.Args = defaultProfiles()[CLIGemini].Args
				break
			}
		}
	}
	if len(p.InteractiveArgs) == 0 {
		p.InteractiveArgs = defaultProfiles()[cli].InteractiveArgs
		if p.ResetCommand == "" {
			p.ResetCommand = defaultProfiles()[cli].ResetCommand
		}
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
		if !KnownCLI(cli) {
			return fmt.Errorf("sessions.%s: agente %q inválido (use %s)", role, cli, strings.Join(OfferedCLIs(), ", "))
		}
		if _, ok := c.Profile(cli); !ok {
			return fmt.Errorf("sessions.%s: cli_profiles sem perfil para %q", role, cli)
		}
	}
	if err := c.validateModels(); err != nil {
		return err
	}
	if !ValidTerminal(c.Terminal) {
		return fmt.Errorf("terminal %q inválido (use auto, tmux ou native)", c.Terminal)
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

// ModelEffortFlags são as flags de modelo e esforço que cada CLI (pelo comando do perfil) aceita.
// Confirmadas no `--help` do build: claude e agy (--model, --effort) e gemini (--model, sem esforço).
// codex e copilot ficam de fora até a flag ser confirmada: só se emite flag que a CLI aceita.
type ModelEffortFlags struct {
	Model   string
	Effort  string
	Efforts []string
}

var modelEffortByCommand = map[string]ModelEffortFlags{
	"claude": {Model: "--model", Effort: "--effort", Efforts: []string{"low", "medium", "high", "xhigh", "max"}},
	"agy":    {Model: "--model", Effort: "--effort", Efforts: []string{"low", "medium", "high", "xhigh", "max"}},
	"gemini": {Model: "--model"},
}

var modelNameRe = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,80}$`)

// ValidModel diz se o nome do modelo é aceito (SEG-01).
func ValidModel(m string) bool { return modelNameRe.MatchString(m) }

// FlagsForCommand devolve as flags de modelo e esforço de uma CLI; ok=false quando não há flag confirmada.
func FlagsForCommand(command string) (ModelEffortFlags, bool) {
	f, ok := modelEffortByCommand[filepath.Base(command)]
	return f, ok
}

// ModelPresets são as sugestões do Settings por comando da CLI; a primeira opção do menu é sempre "padrão da CLI".
func ModelPresets(command string) []string {
	if filepath.Base(command) == "claude" {
		return []string{"haiku", "sonnet", "opus"}
	}
	return nil
}

// SessionModelArgs devolve os argumentos de modelo e de esforço da sessão do papel; vazios sem escolha.
func (c ForgeConfig) SessionModelArgs(role string) (model, effort []string) {
	profile, ok := c.Profile(c.CLIFor(role))
	if !ok {
		return nil, nil
	}
	flags, ok := FlagsForCommand(profile.Command)
	if !ok {
		return nil, nil
	}
	if m := c.Models[role]; m != "" && flags.Model != "" {
		model = []string{flags.Model, m}
	}
	if e := c.Efforts[role]; e != "" && flags.Effort != "" {
		effort = []string{flags.Effort, e}
	}
	return model, effort
}

// MissingModelMarkers lista os marcadores que faltam no perfil da sessão, quando a CLI tem flag
// confirmada para eles (claude, agy e, só para modelo, gemini). Vale com ou sem escolha: um perfil
// gravado sem o marcador (como o de forge.json antigo) não deixaria o Settings aplicar nada.
func (c ForgeConfig) MissingModelMarkers(role string) []string {
	profile, ok := c.Profile(c.CLIFor(role))
	if !ok {
		return nil
	}
	flags, supported := FlagsForCommand(profile.Command)
	if !supported {
		return nil
	}
	var missing []string
	if flags.Model != "" && !contains(profile.Args, MarkerModel) {
		missing = append(missing, MarkerModel)
	}
	if flags.Effort != "" && !contains(profile.Args, MarkerEffort) {
		missing = append(missing, MarkerEffort)
	}
	return missing
}

func (c ForgeConfig) validateModels() error {
	check := func(field string, values map[string]string, effort bool) error {
		for role, v := range values {
			if !contains(SessionRoles(), role) {
				return fmt.Errorf("%s: papel %q desconhecido (use discovery, builder ou reviewer)", field, role)
			}
			if v == "" {
				continue
			}
			cli := c.CLIFor(role)
			profile, ok := c.Profile(cli)
			if !ok {
				return fmt.Errorf("%s.%s: cli_profiles sem perfil para %q", field, role, cli)
			}
			flags, supported := FlagsForCommand(profile.Command)
			switch {
			case !effort && !ValidModel(v):
				return fmt.Errorf("%s.%s: %q inválido (use letras, números, ponto, dois-pontos, hífen e sublinhado, até 80)", field, role, v)
			case !supported || (!effort && flags.Model == "") || (effort && flags.Effort == ""):
				return fmt.Errorf("%s.%s: a CLI %q não tem a flag confirmada (remova a escolha)", field, role, profile.Command)
			case effort && !contains(flags.Efforts, v):
				return fmt.Errorf("%s.%s: esforço %q fora da lista da CLI %s (%s)", field, role, v, profile.Command, strings.Join(flags.Efforts, ", "))
			}
		}
		return nil
	}
	if err := check("models", c.Models, false); err != nil {
		return err
	}
	return check("efforts", c.Efforts, true)
}
