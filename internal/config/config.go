package config

import "github.com/spf13/cobra"

// Config contém todas as configurações coletadas durante o init.
type Config struct {
	Project    string // nome do projeto
	Stack      string // runtime principal: go, node, python, rust, other
	DB         string // banco de dados: postgres, sqlite, mongo, none
	Telemetry  bool   // habilitar telemetria local
	Lang       string // idioma dos templates: pt-BR, en
	SddVersion string // versão Forge-SDD a usar
	DryRun     bool   // exibir árvore sem criar arquivos
}

// Defaults retorna uma Config com valores padrão.
func Defaults() Config {
	return Config{
		Project:    "meu-projeto",
		Stack:      "go",
		DB:         "none",
		Telemetry:  true,
		Lang:       "pt-BR",
		SddVersion: "1.1.0",
		DryRun:     false,
	}
}

// FromFlags preenche a Config a partir das flags registradas no cobra.Command.
// Começa com Defaults() e sobrescreve apenas as flags que foram explicitamente passadas.
func FromFlags(cmd *cobra.Command) Config {
	cfg := Defaults()
	if f := cmd.Flags().Lookup("name"); f != nil && f.Changed {
		cfg.Project = f.Value.String()
	}
	if f := cmd.Flags().Lookup("stack"); f != nil && f.Changed {
		cfg.Stack = f.Value.String()
	}
	if f := cmd.Flags().Lookup("db"); f != nil && f.Changed {
		cfg.DB = f.Value.String()
	}
	if f := cmd.Flags().Lookup("lang"); f != nil && f.Changed {
		cfg.Lang = f.Value.String()
	}
	if f := cmd.Flags().Lookup("version"); f != nil && f.Changed {
		cfg.SddVersion = f.Value.String()
	}
	if f := cmd.Flags().Lookup("no-telemetry"); f != nil && f.Changed {
		v, _ := cmd.Flags().GetBool("no-telemetry")
		cfg.Telemetry = !v
	}
	return cfg
}
