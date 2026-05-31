package config

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
