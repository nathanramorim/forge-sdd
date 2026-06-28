package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/forge-sdd/cli/internal/survey"
	"github.com/spf13/cobra"
)

// version é injetada via ldflags: -X main.version=1.5.3
var version = "1.5.3"

var rootCmd = &cobra.Command{
	Use:   "forge-sdd",
	Short: "Scaffolda estruturas Forge-SDD em qualquer projeto",
	Long: `forge-sdd é um CLI que gera a estrutura completa Forge-SDD
(sdd/, .github/, .vscode/) pronta para uso com GitHub Copilot.`,
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
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}

		yes, _ := cmd.Flags().GetBool("yes")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		var cfg config.Config
		if yes {
			cfg = config.FromFlags(cmd)
			// validar --agent explicitamente para retornar erro claro
			if f := cmd.Flags().Lookup("agent"); f != nil && f.Changed {
				agents, err := config.ParseAgents(f.Value.String())
				if err != nil {
					return err
				}
				cfg.Agents = agents
			}
			cfg.DryRun = dryRun
		} else {
			var err error
			cfg, err = survey.Run()
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			cfg.DryRun = dryRun
		}

		// usa versão do binário se não sobrescrita pela flag --version
		if cfg.SddVersion == config.Defaults().SddVersion && version != "dev" {
			cfg.SddVersion = version
		}

		created, err := scaffold.Run(cfg, targetDir)
		if err != nil {
			return fmt.Errorf("scaffold falhou: %w", err)
		}

		fmt.Printf("\n✓ Estrutura Forge-SDD criada em %s (%d arquivos)\n\n", targetDir, len(created))
		fmt.Println("Próximos passos:")
		fmt.Println("  1. Abra o projeto no VS Code")
		fmt.Println("  2. Aceite as extensões recomendadas (Copilot, MCP)")
		fmt.Println("  3. Leia sdd/memory/progress.md para começar")
		return nil
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

		rc, err := config.ReadSddrc(targetDir)
		if err != nil {
			return err
		}

		projectVersion, err := config.ReadProjectVersion(targetDir)
		if err != nil {
			// .sdd-version ausente em projetos legados — usa versão do .sddrc
			projectVersion = rc.Version
		}

		yes, _ := cmd.Flags().GetBool("yes")

		var toAdd []string
		var targetVersion string // vazio = sem atualização de versão

		if yes {
			agentFlag, _ := cmd.Flags().GetString("agent")
			upgradeFlag, _ := cmd.Flags().GetBool("upgrade")
			versionFlag, _ := cmd.Flags().GetString("version")

			// Resolve agentes a adicionar
			if agentFlag != "" {
				parsed, err := config.ParseAgents(agentFlag)
				if err != nil {
					return err
				}
				toAdd = config.FilterNewAgents(rc.Agents, parsed)
			}

			// Resolve versão alvo
			switch {
			case versionFlag != "":
				targetVersion = versionFlag
			case upgradeFlag:
				targetVersion = version
			}

			if len(toAdd) == 0 && targetVersion == "" {
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
			choice, err := survey.RunUpdate(rc.Agents, projectVersion, version)
			if err != nil {
				return fmt.Errorf("formulário cancelado: %w", err)
			}
			toAdd = choice.AgentsToAdd
			targetVersion = choice.TargetVersion
		}

		mergedAgents := config.MergeAgents(rc.Agents, toAdd)
		baseCfg := rc.ToConfig()
		baseCfg.Agents = mergedAgents

		var created []string

		if targetVersion != "" {
			// Upgrade de versão: re-renderiza toda a estrutura (exceto sdd/ domain)
			baseCfg.SddVersion = targetVersion
			created, err = scaffold.Run(baseCfg, targetDir)
			if err != nil {
				return fmt.Errorf("upgrade falhou: %w", err)
			}
		} else {
			// Apenas agentes novos
			baseCfg.Agents = toAdd
			created, err = scaffold.RunAgents(baseCfg, toAdd, targetDir)
			if err != nil {
				return fmt.Errorf("scaffold falhou: %w", err)
			}
			baseCfg.Agents = mergedAgents
			if err := scaffold.UpdateSddrc(baseCfg, targetDir); err != nil {
				return fmt.Errorf("atualizar sdd/.sddrc: %w", err)
			}
		}

		fmt.Printf("\n✓ Atualização concluída (%d arquivos)\n\n", len(created))

		if targetVersion != "" {
			fmt.Printf("  Versão: v%s → v%s\n", projectVersion, targetVersion)
		}
		if len(toAdd) > 0 {
			fmt.Printf("  Agentes adicionados: %s\n", strings.Join(toAdd, ", "))
		}

		fmt.Println("\nPróximos passos:")
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
		if targetVersion != "" {
			fmt.Println("  • Estrutura atualizada: revise sdd/memory/progress.md para retomar o trabalho")
		}
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("yes", false, "Pular prompts e usar flags/defaults")
	initCmd.Flags().Bool("dry-run", false, "Listar arquivos que seriam criados sem gravar no disco")
	initCmd.Flags().String("name", "", "Nome do projeto")
	initCmd.Flags().String("stack", "", "Stack principal: go, node, python, rust, other")
	initCmd.Flags().String("db", "", "Banco de dados: postgres, sqlite, mongo, none")
	initCmd.Flags().String("lang", "", "Idioma: pt-BR, en")
	initCmd.Flags().String("version", "", "Versão Forge-SDD (default: 1.5.3)")
	initCmd.Flags().Bool("no-telemetry", false, "Desabilitar telemetria local")
	initCmd.Flags().String("agent", "", "Agente(s) de IA: copilot, claude, gemini (csv, default: copilot)")
	rootCmd.AddCommand(initCmd)

	updateCmd.Flags().Bool("yes", false, "Pular prompts e usar flags")
	updateCmd.Flags().String("agent", "", "Agente(s) a adicionar: copilot, claude, gemini, openai (csv)")
	updateCmd.Flags().Bool("upgrade", false, "Atualizar estrutura para a versão mais recente do binário")
	updateCmd.Flags().String("version", "", "Atualizar estrutura para uma versão específica (ex: 1.6.0)")
	rootCmd.AddCommand(updateCmd)

	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
