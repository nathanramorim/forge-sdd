package survey

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/forge-sdd/cli/internal/config"
)

// Run apresenta o formulário interativo e retorna a Config preenchida.
func Run() (config.Config, error) {
	cfg := config.Defaults()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nome do projeto").
				Description("Ex: meu-servico, auth-api").
				Value(&cfg.Project).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("nome não pode ser vazio")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Stack principal").
				Options(
					huh.NewOption("Go", "go"),
					huh.NewOption("Node.js", "node"),
					huh.NewOption("Python", "python"),
					huh.NewOption("Rust", "rust"),
					huh.NewOption("Outro", "other"),
				).
				Value(&cfg.Stack),

			huh.NewSelect[string]().
				Title("Banco de dados").
				Options(
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("SQLite", "sqlite"),
					huh.NewOption("MongoDB", "mongo"),
					huh.NewOption("Nenhum", "none"),
				).
				Value(&cfg.DB),

			huh.NewMultiSelect[string]().
				Title("Agente(s) de IA").
				Description("Selecione um ou mais agentes que utilizará no projeto").
				Options(
					huh.NewOption("GitHub Copilot", config.AgentCopilot).Selected(true),
					huh.NewOption("Claude (Anthropic)", config.AgentClaude),
					huh.NewOption("Gemini (Google)", config.AgentGemini),
					huh.NewOption("OpenAI (GPT-4/Codex)", config.AgentOpenAI),
				).
				Value(&cfg.Agents).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("selecione ao menos um agente")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Habilitar telemetria local?").
				Value(&cfg.Telemetry),

			huh.NewSelect[string]().
				Title("Idioma dos templates").
				Options(
					huh.NewOption("Português (pt-BR)", "pt-BR"),
					huh.NewOption("English (en)", "en"),
				).
				Value(&cfg.Lang),
		),
	)

	if err := form.Run(); err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// UpdateChoice contém as escolhas do usuário no formulário de update.
type UpdateChoice struct {
	AgentsToAdd   []string
	TargetVersion string // vazio = não atualizar versão
}

// RunUpdate apresenta formulário interativo para atualizar agentes e/ou versão.
// existing: agentes já configurados.
// currentVersion: versão registrada no projeto (sdd/.sdd-version).
// binaryVersion: versão do binário instalado (a mais recente disponível).
func RunUpdate(existing []string, currentVersion, binaryVersion string) (UpdateChoice, error) {
	existingSet := make(map[string]bool, len(existing))
	for _, a := range existing {
		existingSet[a] = true
	}

	type agentOption struct{ id, label string }
	allAgents := []agentOption{
		{config.AgentCopilot, "GitHub Copilot"},
		{config.AgentClaude, "Claude (Anthropic)"},
		{config.AgentGemini, "Gemini (Google)"},
		{config.AgentOpenAI, "OpenAI (GPT-4/Codex)"},
	}

	var options []huh.Option[string]
	var toAdd []string

	for _, a := range allAgents {
		label := a.label
		if existingSet[a.id] {
			label = fmt.Sprintf("%s (Já instalado)", a.label)
		}
		opt := huh.NewOption(label, a.id)
		options = append(options, opt)
	}

	versionOutdated := config.CompareVersions(currentVersion, binaryVersion) < 0
	var groups []*huh.Group

	// Campo de versão (apenas se desatualizado)
	targetVersion := ""
	if versionOutdated {
		targetVersion = binaryVersion
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Atualizar estrutura Forge-SDD (atual: v%s → disponível: v%s)", currentVersion, binaryVersion)).
				Description("Confirme a versão alvo ou deixe vazio para não atualizar").
				Placeholder(binaryVersion).
				Value(&targetVersion),
		))
	}

	// Campo de agentes
	desc := "Selecione novos agentes para instalar ou marque os já existentes para reinstalar/atualizar seus prompts."
	groups = append(groups, huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Configurar/Atualizar agentes de IA").
			Description(desc).
			Options(options...).
			Value(&toAdd),
	))

	if err := huh.NewForm(groups...).Run(); err != nil {
		return UpdateChoice{}, err
	}

	return UpdateChoice{
		AgentsToAdd:   toAdd,
		TargetVersion: strings.TrimSpace(targetVersion),
	}, nil
}

// RunSmartUpgradePrompt pergunta de forma interativa para qual versão o usuário quer atualizar se um projeto existente for detectado.
func RunSmartUpgradePrompt(currentVersion, latestVersion, betaVersion string) (string, error) {
	var targetOption string

	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("Não atualizar (manter na v%s)", currentVersion), "keep"),
	}

	if latestVersion != "" {
		options = append(options, huh.NewOption(fmt.Sprintf("Versão Oficial Estável: v%s (latest)", latestVersion), "latest"))
	}
	if betaVersion != "" {
		options = append(options, huh.NewOption(fmt.Sprintf("Versão de Teste Beta: v%s (beta)", betaVersion), "beta"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Projeto existente detectado!").
				Description(fmt.Sprintf("Deseja atualizar a estrutura do Forge-SDD de v%s para uma versão mais recente?", currentVersion)).
				Options(options...).
				Value(&targetOption),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	switch targetOption {
	case "latest":
		return latestVersion, nil
	case "beta":
		return betaVersion, nil
	default:
		return "", nil // keep/não atualizar
	}
}

// RunConfirm apresenta um prompt de confirmação simples de Sim/Não.
func RunConfirm(title, description string, defaultVal bool) (bool, error) {
	var result bool = defaultVal
	err := huh.NewConfirm().
		Title(title).
		Description(description).
		Value(&result).
		Run()
	return result, err
}


