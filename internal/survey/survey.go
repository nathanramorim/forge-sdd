package survey

import (
	"fmt"

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
