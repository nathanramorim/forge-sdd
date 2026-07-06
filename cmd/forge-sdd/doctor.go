package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [diretório]",
	Short: "Diagnostica a saúde e integridade da instalação Forge-SDD",
	Long: `Analisa o diretório alvo e valida a presença de arquivos cruciais da metodologia,
inconsistências de configuração dos agentes de IA e progresso das features.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) == 1 {
			targetDir = args[0]
		}

		fmt.Printf("🩺 Iniciando diagnóstico da metodologia em '%s'...\n\n", targetDir)

		sddrcPath := filepath.Join(targetDir, "sdd", ".sddrc")
		versionPath := filepath.Join(targetDir, "sdd", ".sdd-version")

		hasSddrc := fileExists(sddrcPath)
		hasVersion := fileExists(versionPath)

		if !hasSddrc {
			fmt.Println("✗ Estrutura Forge-SDD não encontrada (sdd/.sddrc ausente).")
			fmt.Println("  Use 'forge-sdd init' para inicializar a metodologia neste repositório.")
			return nil
		}

		fmt.Println("✓ Diretório base da metodologia (sdd/) e arquivo de configuração encontrados.")

		// 1. Carrega e valida sdd/.sddrc
		var rc struct {
			Project   string   `json:"project"`
			Version   string   `json:"version"`
			Agents    []string `json:"agents"`
			Stack     string   `json:"stack"`
			DB        string   `json:"db"`
			Lang      string   `json:"lang"`
			Telemetry struct {
				Enabled   bool   `json:"enabled"`
				Anonymous bool   `json:"anonymous"`
				Endpoint  string `json:"endpoint"`
			} `json:"telemetry"`
		}

		rcData, err := os.ReadFile(sddrcPath)
		if err != nil {
			fmt.Printf("✗ Falha ao ler sdd/.sddrc: %v\n", err)
			return nil
		}

		if err := json.Unmarshal(rcData, &rc); err != nil {
			fmt.Printf("✗ Arquivo sdd/.sddrc inválido ou corrompido: %v\n", err)
			return nil
		}

		fmt.Println("   - Projeto:", rc.Project)
		fmt.Println("   - Stack:", rc.Stack)
		fmt.Println("   - Idioma:", rc.Lang)
		fmt.Println("   - Agentes no .sddrc:", strings.Join(rc.Agents, ", "))

		// 2. Valida sdd/.sdd-version
		if hasVersion {
			verData, err := os.ReadFile(versionPath)
			if err != nil {
				fmt.Printf("⚠️  Não foi possível ler sdd/.sdd-version: %v\n", err)
			} else {
				vStr := strings.TrimSpace(string(verData))
				if vStr != rc.Version {
					fmt.Printf("⚠️  Descompasso de versão: sdd/.sddrc indica v%s mas sdd/.sdd-version indica v%s\n", rc.Version, vStr)
				} else {
					fmt.Printf("✓ Versão registrada em sincronia: v%s\n", vStr)
				}
			}
		} else {
			fmt.Println("⚠️  Arquivo sdd/.sdd-version não encontrado.")
		}

		// 3. Validação física dos Agentes
		fmt.Println("\n🤖 Diagnóstico de Agentes:")
		for _, agent := range rc.Agents {
			switch agent {
			case config.AgentClaude:
				hasClaudeMD := fileExists(filepath.Join(targetDir, "CLAUDE.md"))
				hasClaudeDir := fileExists(filepath.Join(targetDir, ".claude"))
				if hasClaudeMD && hasClaudeDir {
					fmt.Println("  ✓ Claude (Anthropic): Configurado e integrado (CLAUDE.md + .claude/)")
				} else {
					fmt.Println("  ⚠️  Claude (Anthropic): Configurado no .sddrc, mas arquivos físicos ausentes ou incompletos.")
				}
			case config.AgentGemini:
				hasGeminiMD := fileExists(filepath.Join(targetDir, "GEMINI.md"))
				hasGeminiDir := fileExists(filepath.Join(targetDir, ".gemini"))
				if hasGeminiMD && hasGeminiDir {
					fmt.Println("  ✓ Gemini (Google): Configurado e integrado (GEMINI.md + .gemini/)")
				} else {
					fmt.Println("  ⚠️  Gemini (Google): Configurado no .sddrc, mas arquivos físicos ausentes ou incompletos.")
				}
			case config.AgentCopilot:
				hasCopilotMD := fileExists(filepath.Join(targetDir, ".github", "copilot-instructions.md"))
				if hasCopilotMD {
					fmt.Println("  ✓ GitHub Copilot: Configurado e integrado (.github/copilot-instructions.md)")
				} else {
					fmt.Println("  ⚠️  GitHub Copilot: Configurado no .sddrc, mas arquivos físicos ausentes (.github/copilot-instructions.md)")
				}
			case config.AgentOpenAI:
				hasOpenAIMD := fileExists(filepath.Join(targetDir, "OPENAI.md"))
				hasOpenAIDir := fileExists(filepath.Join(targetDir, ".openai"))
				if hasOpenAIMD && hasOpenAIDir {
					fmt.Println("  ✓ OpenAI (GPT-4/Codex): Configurado e integrado (OPENAI.md + .openai/)")
				} else {
					fmt.Println("  ⚠️  OpenAI (GPT-4/Codex): Configurado no .sddrc, mas arquivos físicos ausentes ou incompletos.")
				}
			}
		}

		// 4. Verificação de arquivos essenciais de metodologia
		fmt.Println("\n📁 Arquivos de Metodologia:")
		constitutionPath := filepath.Join(targetDir, "sdd", "memory", "constitution.md")
		progressPath := filepath.Join(targetDir, "sdd", "memory", "progress.md")

		if fileExists(constitutionPath) {
			fmt.Println("  ✓ Constituição (sdd/memory/constitution.md) ativa.")
		} else {
			fmt.Println("  ✗ Constituição (sdd/memory/constitution.md) não encontrada!")
		}

		if fileExists(progressPath) {
			fmt.Println("  ✓ Painel de Progresso (sdd/memory/progress.md) ativo.")
		} else {
			fmt.Println("  ✗ Painel de Progresso (sdd/memory/progress.md) não encontrado!")
		}

		// 5. Varredura de features
		featuresDir := filepath.Join(targetDir, "sdd", "features")
		var activeFeatures []string
		err = filepath.WalkDir(featuresDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
				if strings.HasPrefix(d.Name(), "feat-") || strings.HasPrefix(d.Name(), "task-") {
					relPath, err := filepath.Rel(featuresDir, path)
					if err == nil {
						activeFeatures = append(activeFeatures, relPath)
					} else {
						activeFeatures = append(activeFeatures, d.Name())
					}
				}
			}
			return nil
		})

		if err == nil {
			fmt.Printf("\n📋 Features Registradas: %d cadastrada(s)\n", len(activeFeatures))
			if len(activeFeatures) > 0 {
				fmt.Println("  Caminhos:")
				for _, feat := range activeFeatures {
					fmt.Printf("   - sdd/features/%s\n", feat)
				}
			}
		} else {
			fmt.Println("\n⚠️  Diretório sdd/features não encontrado ou inacessível.")
		}

		fmt.Println("\n🩺 Diagnóstico finalizado.")
		return nil
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
