package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

// version é injetada via ldflags: -X main.version=1.9.4
var version = "1.9.4"

var rootCmd = &cobra.Command{
	Use:   "forge-sdd",
	Short: "Scaffolda estruturas Forge-SDD em qualquer projeto",
	Long: `forge-sdd é um CLI que gera a estrutura completa Forge-SDD
(sdd/, .github/, .vscode/) pronta para uso com GitHub Copilot.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if config.DetectProject(".") {
				return runUpdateFlow(cmd, ".")
			} else {
				return runInitFlow(cmd, ".")
			}
		}
		return cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do forge-sdd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init [diretório]",
	Short: "Inicializa a estrutura Forge-SDD no diretório alvo",
	Long: `Cria toda a árvore Forge-SDD no diretório especificado (default: diretório atual).
Preenche os templates com as informações do projeto coletadas interativamente ou via flags.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var dirArg string
		if len(args) == 1 {
			dirArg = args[0]
		}
		return runInitFlow(cmd, dirArg)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [diretório]",
	Short: "Adiciona agente(s) e/ou atualiza a versão de uma estrutura Forge-SDD existente",
	Long: `Lê a configuração atual em sdd/.sddrc e sdd/.sdd-version, então:
  • Informa se a estrutura está desatualizada em relação ao binário instalado
  • Permite atualizar para a versão mais recente ou uma versão específica
  • Permite adicionar novos agentes de IA

A estrutura de domínio em sdd/ (features, spec, memory) nunca é alterada.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}
		return runUpdateFlow(cmd, targetDir)
	},
}

func runInitFlow(cmd *cobra.Command, dirArg string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	tutorial, _ := cmd.Flags().GetBool("tutorial")
	nameFlag, _ := cmd.Flags().GetString("name")

	var targetDir string
	var defaultProjectName string

	// CASO A: Usuário especificou diretório (ex: init . ou init meu-projeto)
	if dirArg != "" {
		targetDir = dirArg
		// Se a pasta física já tiver Forge-SDD, redireciona para update
		if config.DetectProject(targetDir) {
			fmt.Printf("\n→ Estrutura Forge-SDD já detectada em %s. Redirecionando para atualização...\n", targetDir)
			return runUpdateFlow(cmd, targetDir)
		}

		// Extrair nome do projeto padrão a partir do caminho físico
		absPath, err := filepath.Abs(targetDir)
		if err == nil {
			defaultProjectName = filepath.Base(absPath)
		} else {
			defaultProjectName = filepath.Base(targetDir)
		}
	}

	var cfg config.Config
	if yes {
		cfg = config.FromFlags(cmd)
		if f := cmd.Flags().Lookup("agent"); f != nil && f.Changed {
			agents, err := config.ParseAgents(f.Value.String())
			if err != nil {
				return err
			}
			cfg.Agents = agents
		}
		cfg.DryRun = dryRun

		// Se o usuário passou um diretório (Caso A), usa esse diretório e ajusta o nome se nameFlag estiver vazia
		if dirArg != "" {
			if nameFlag == "" {
				cfg.Project = defaultProjectName
			}
		} else {
			// CASO B: Sem argumentos (automático via flags)
			// Usa o nome do projeto para criar a subpasta correspondente
			targetDir = "./" + cfg.Project
		}
	} else {
		// Modo interativo
		if dirArg != "" {
			// Caso A (interativo): passa o nome padrão obtido do diretório
			var err error
			cfg, err = survey.Run(defaultProjectName)
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			cfg.DryRun = dryRun
		} else {
			// Caso B (interativo): sem argumentos, coleta o nome do projeto
			var err error
			cfg, err = survey.Run("")
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			cfg.DryRun = dryRun
			// O diretório alvo será a subpasta com o nome do projeto
			targetDir = "./" + cfg.Project
		}
	}

	if cfg.SddVersion == config.Defaults().SddVersion && version != "dev" {
		cfg.SddVersion = version
	}

	// Criar a pasta se ela não for o diretório atual e se não for dry-run
	if targetDir != "." && !cfg.DryRun {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("falha ao criar diretório %s: %w", targetDir, err)
		}
	}

	created, err := scaffold.Run(cfg, targetDir)
	if err != nil {
		return fmt.Errorf("scaffold falhou: %w", err)
	}

	printSummaryReport(cfg, targetDir, len(created))
	if !cfg.DryRun {
		printCheatSheet(cfg)
		if tutorial {
			printTutorialHint()
		}
	}
	return nil
}

func runUpdateFlow(cmd *cobra.Command, targetDir string) error {
	rc, err := config.ReadSddrc(targetDir)
	if err != nil {
		return err
	}

	namingConventionChanged := false
	if f := cmd.Flags().Lookup("naming-convention"); f != nil && f.Changed {
		rc.NamingConvention = f.Value.String()
		namingConventionChanged = true
	}

	projectVersion, err := config.ReadProjectVersion(targetDir)
	if err != nil {
		projectVersion = rc.Version
	}

	yes, _ := cmd.Flags().GetBool("yes")
	upgradeFlag, _ := cmd.Flags().GetBool("upgrade")
	versionFlag, _ := cmd.Flags().GetString("version")

	var toAdd []string
	var targetVersion string

	if yes {
		agentFlag, _ := cmd.Flags().GetString("agent")

		if agentFlag != "" {
			parsed, err := config.ParseAgents(agentFlag)
			if err != nil {
				return err
			}
			toAdd = config.FilterNewAgents(rc.Agents, parsed)
		}

		switch {
		case versionFlag != "":
			targetVersion = versionFlag
		case upgradeFlag:
			latest, beta, fetchErr := config.FetchNpmVersions()
			if fetchErr != nil {
				return fmt.Errorf("--upgrade: falha ao consultar o NPM Registry: %w", fetchErr)
			}
			targetVersion = config.ResolveUpgradeTarget(latest, beta)
			if targetVersion == "" {
				return fmt.Errorf("--upgrade: NPM Registry não retornou nenhuma versão (latest/beta) publicada")
			}
		}

		if len(toAdd) == 0 && targetVersion == "" && !namingConventionChanged {
			versionOutdated := config.CompareVersions(projectVersion, version) < 0
			if versionOutdated {
				fmt.Printf("Versão desatualizada: projeto v%s → CLI v%s\n", projectVersion, version)
				fmt.Println("Use --upgrade para atualizar ou --version X.Y.Z para uma versão específica.")
			} else {
				fmt.Printf("Nenhuma atualização necessária (versão %s, agentes: %s)\n", projectVersion, strings.Join(rc.Agents, ", "))
			}
			return nil
		}
	} else {
		// Modo interativo: busca versões remoto no NPM Registry
		fmt.Print("Buscando versões disponíveis no NPM Registry...")
		latest, beta, fetchErr := config.FetchNpmVersions()
		if fetchErr != nil {
			fmt.Fprintf(os.Stderr, "\n⚠ Falha ao consultar o NPM Registry (%v).\n  A versão beta mais recente pode não estar disponível nesta execução — usando apenas a versão local do binário (v%s).\n", fetchErr, version)
			latest = version
			beta = ""
		} else {
			fmt.Println(" [ok]")
		}

		if versionFlag != "" {
			targetVersion = versionFlag
		} else {
			chosenVersion, err := survey.RunSmartUpgradePrompt(projectVersion, latest, beta, upgradeFlag)
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			targetVersion = chosenVersion
		}

		choice, err := survey.RunUpdate(rc.Agents, projectVersion, version)
		if err != nil {
			return fmt.Errorf("formulário cancelado: %w", err)
		}
		toAdd = choice.AgentsToAdd
		if targetVersion == "" {
			targetVersion = choice.TargetVersion
		}
	}

	mergedAgents := config.MergeAgents(rc.Agents, toAdd)
	baseCfg := rc.ToConfig()
	baseCfg.Agents = mergedAgents
	if baseCfg.NamingConvention == "" {
		baseCfg.NamingConvention = config.DetectNamingConvention(targetDir)
	}

	var created []string

	if targetVersion != "" {
		baseCfg.SddVersion = targetVersion
		created, err = scaffold.Run(baseCfg, targetDir)
		if err != nil {
			return fmt.Errorf("upgrade falhou: %w", err)
		}
	} else if len(toAdd) > 0 {
		baseCfg.Agents = toAdd
		created, err = scaffold.RunAgents(baseCfg, toAdd, targetDir)
		if err != nil {
			return fmt.Errorf("scaffold falhou: %w", err)
		}
		baseCfg.Agents = mergedAgents
		if err := scaffold.UpdateSddrc(baseCfg, targetDir); err != nil {
			return fmt.Errorf("atualizar sdd/.sddrc: %w", err)
		}
	} else if namingConventionChanged {
		if err := scaffold.UpdateSddrc(baseCfg, targetDir); err != nil {
			return fmt.Errorf("atualizar sdd/.sddrc: %w", err)
		}
		fmt.Printf("✓ Convenção de nomenclatura atualizada para: %s\n", baseCfg.NamingConvention)
	} else {
		fmt.Println("Nenhuma modificação realizada.")
		return nil
	}

	// Recria a config base atualizada com as escolhas finais
	finalCfg := baseCfg
	if targetVersion != "" {
		finalCfg.SddVersion = targetVersion
	}

	printSummaryReport(finalCfg, targetDir, len(created))

	// Próximos passos adicionais específicos para agentes recém-adicionados
	if len(toAdd) > 0 {
		fmt.Println("\nInstruções dos Agentes Adicionados:")
		for _, a := range toAdd {
			switch a {
			case config.AgentClaude:
				fmt.Println("  • Claude: leia CLAUDE.md e abra o projeto no Claude Code")
			case config.AgentGemini:
				fmt.Println("  • Gemini: leia GEMINI.md e abra o projeto no Gemini")
			case config.AgentOpenAI:
				fmt.Println("  • OpenAI: leia OPENAI.md e configure sua chave de API")
			case config.AgentCopilot:
				fmt.Println("  • Copilot: aceite as extensões recomendadas no VS Code")
			}
		}
	}
	return nil
}

func printSummaryReport(cfg config.Config, targetDir string, fileCount int) {
	telemetryStatus := "Ativo (Local)"
	if !cfg.Telemetry {
		telemetryStatus = "Desativado"
	}

	fmt.Printf("\n✓ Estrutura Forge-SDD inicializada/atualizada com sucesso! (%d arquivos)\n\n", fileCount)
	fmt.Println("Resumo do Projeto:")
	fmt.Printf("  • Nome:           %s\n", cfg.Project)
	fmt.Printf("  • Diretório:      %s\n", targetDir)
	fmt.Printf("  • Stack:          %s\n", cfg.Stack)
	fmt.Printf("  • Banco de Dados: %s\n", cfg.DB)
	fmt.Printf("  • Idioma:         %s\n", cfg.Lang)
	fmt.Printf("  • Agentes:        %s\n", strings.Join(cfg.Agents, ", "))
	fmt.Printf("  • Telemetria:     %s\n", telemetryStatus)
	fmt.Println("\nPróximos passos:")
	fmt.Println("  1. Abra o projeto no VS Code ou sua IDE de preferência")
	fmt.Println("  2. Aceite as extensões recomendadas (Copilot, MCP, etc.)")
	fmt.Println("  3. Leia sdd/memory/progress.md para começar a codificar!")
}

func printTutorialHint() {
	fmt.Println("\n🎓 Modo tutorial ativado: abra seu agente de IA e peça \"/tutorial\" para aprender o ciclo SDD com um exemplo guiado e descartável antes de começar sua demanda real.")
}

func printCheatSheet(cfg config.Config) {
	commands := scaffold.CommandCheatSheet(cfg)
	if len(commands) == 0 {
		return
	}
	fmt.Println("\nComandos SDD disponíveis:")
	for _, c := range commands {
		fmt.Printf("  %-18s %s\n", c.Command, c.Description)
	}
}

func init() {
	initCmd.Flags().Bool("yes", false, "Pular prompts e usar flags/defaults")
	initCmd.Flags().Bool("dry-run", false, "Listar arquivos que seriam criados sem gravar no disco")
	initCmd.Flags().Bool("tutorial", false, "Ao final, sugerir rodar /tutorial para aprender o ciclo SDD com um exemplo guiado")
	initCmd.Flags().String("name", "", "Nome do projeto")
	initCmd.Flags().String("stack", "", "Stack principal: go, node, python, rust, other")
	initCmd.Flags().String("db", "", "Banco de dados: postgres, sqlite, mongo, none")
	initCmd.Flags().String("lang", "", "Idioma: pt-BR, en")
	initCmd.Flags().String("version", "", "Versão Forge-SDD (default: 1.9.4)")
	initCmd.Flags().Bool("no-telemetry", false, "Desabilitar telemetria local")
	initCmd.Flags().String("agent", "", "Agente(s) de IA: copilot, claude, gemini (csv, default: copilot)")
	initCmd.Flags().String("naming-convention", "", "Convenção de nomenclatura: sequencial, hash, workitem")
	rootCmd.AddCommand(initCmd)

	updateCmd.Flags().Bool("yes", false, "Pular prompts e usar flags")
	updateCmd.Flags().String("agent", "", "Agente(s) a adicionar: copilot, claude, gemini, openai (csv)")
	updateCmd.Flags().Bool("upgrade", false, "Atualizar estrutura para a versão mais recente do binário")
	updateCmd.Flags().String("version", "", "Atualizar estrutura para uma versão específica (ex: 1.9.4)")
	updateCmd.Flags().String("naming-convention", "", "Convenção de nomenclatura: sequencial, hash, workitem")
	rootCmd.AddCommand(updateCmd)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(lessonsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
