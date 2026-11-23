package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/doctor"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/scaffold"
	"github.com/spf13/cobra"
)

// classifyNamingConvention identifica se o nome de um arquivo de feature/discovery/fix
// segue a convenção sequencial ou por hash de 4 dígitos.
func classifyNamingConvention(name string) string {
	return config.ClassifyNamingConvention(name)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor [diretório]",
	Short: "Diagnostica a saúde e integridade da instalação Forge-SDD",
	Long: `Analisa o diretório alvo e valida a presença de arquivos cruciais da metodologia,
inconsistências de configuração dos agentes de IA e progresso das features.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if len(args) == 1 {
			targetDir = args[0]
		}
		if harnessMode, _ := cmd.Flags().GetBool("harness"); harnessMode {
			asJSON, _ := cmd.Flags().GetBool("json")
			return runDoctorHarness(cmd.OutOrStdout(), targetDir, asJSON, doctor.DefaultDeps())
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		fmt.Printf("🩺 Iniciando diagnóstico da metodologia em '%s'...\n\n", targetDir)

		sddrcPath := layout.Sddrc(targetDir)
		versionPath := layout.Version(targetDir)

		hasSddrc := fileExists(sddrcPath)
		hasVersion := fileExists(versionPath)

		if !hasSddrc {
			fmt.Println("✗ Estrutura Forge-SDD não encontrada (sdd/.sddrc ausente).")
			fmt.Println("  Use 'forge init' para inicializar a metodologia neste repositório.")
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
			NamingConvention string `json:"naming_convention"`
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
		if rc.NamingConvention != "" {
			fmt.Println("   - Convenção de Nomenclatura:", rc.NamingConvention)
		}

		// 2. Valida sdd/.sdd-version
		if hasVersion {
			verData, err := os.ReadFile(versionPath)
			if err != nil {
				fmt.Printf("⚠️  Não foi possível ler sdd/.sdd-version: %v\n", err)
			} else {
				vStr := strings.TrimSpace(string(verData))
				if vStr != rc.Version {
					fmt.Printf("⚠️  Descompasso de versão: sdd/.sddrc indica v%s mas sdd/.sdd-version indica v%s\n", notify.Sanitize(rc.Version, 40, false), notify.Sanitize(vStr, 40, false))
				}
				reportDoctorVersao(vStr)
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
				hasOpenAIMD := fileExists(filepath.Join(targetDir, "AGENTS.md")) && fileExists(filepath.Join(targetDir, "OPENAI.md"))
				hasOpenAIDir := fileExists(filepath.Join(targetDir, ".openai"))
				if hasOpenAIMD && hasOpenAIDir {
					fmt.Println("  ✓ OpenAI (GPT-4/Codex): Configurado e integrado (AGENTS.md + OPENAI.md + .openai/)")
				} else {
					fmt.Println("  ⚠️  OpenAI (GPT-4/Codex): Configurado no .sddrc, mas arquivos físicos ausentes ou incompletos.")
				}
			}
		}

		// 3b. Cobertura da barreira de PR (feat-106-15)
		fmt.Println("\n🛡️  Barreira de PR (pr_guard):")
		hookSemScript := reportPRGuardCoverage(targetDir, rc.Agents)

		// 4. Verificação de arquivos essenciais de metodologia
		fmt.Println("\n📁 Arquivos de Metodologia:")
		constitutionPath := layout.Constitution(targetDir)
		progressPath := layout.Progress(targetDir)

		if fileExists(constitutionPath) {
			b, bErr := gate.BuildBrief(targetDir, 0)
			if bErr == nil && b.State == gate.BriefRascunho {
				fmt.Println("  ✓ Constituição (sdd/memory/constitution.md) ativa.")
				fmt.Println("  ⚠️  Constituição em rascunho (rode /constitution para alinhar). Faltam: " + strings.Join(b.Missing, ", "))
			} else {
				fmt.Println("  ✓ Constituição (sdd/memory/constitution.md) ativa.")
			}
		} else {
			fmt.Println("  ✗ Constituição (sdd/memory/constitution.md) não encontrada!")
		}

		if _, hasStatus := layout.StatusPath(targetDir); hasStatus {
			fmt.Println("  ✓ Status da iniciativa (status.md) ativo.")
		} else if fileExists(progressPath) {
			fmt.Println("  ✓ Painel de Progresso (sdd/memory/progress.md) ativo.")
		} else {
			fmt.Println("  ℹ Sem iniciativa ativa na branch atual: o status.md não se aplica (veja `forge status`).")
		}

		// 5. Varredura de features: tasks das iniciativas (sdd/discovery/<id>-<nome>/features/) e,
		// em projetos ainda não migrados, a raiz legada sdd/features/.
		sddRoot := layout.Root(targetDir)
		var activeFeatures []string
		scannedAny := false
		for _, root := range []string{layout.Discovery(targetDir), layout.Features(targetDir)} {
			if _, statErr := os.Stat(root); statErr != nil {
				continue
			}
			scannedAny = true
			_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}
				if !strings.HasPrefix(d.Name(), "feat-") && !strings.HasPrefix(d.Name(), "task-") {
					return nil
				}
				rel, relErr := filepath.Rel(sddRoot, path)
				if relErr != nil {
					rel = d.Name()
				}
				activeFeatures = append(activeFeatures, filepath.ToSlash(rel))
				return nil
			})
		}

		if scannedAny {
			fmt.Printf("\n📋 Features Registradas: %d cadastrada(s)\n", len(activeFeatures))
			if verbose && len(activeFeatures) > 0 {
				fmt.Println("  Caminhos:")
				for _, feat := range activeFeatures {
					fmt.Printf("   - sdd/%s\n", feat)
				}
			}
		} else {
			fmt.Println("\n⚠️  Diretório sdd/discovery não encontrado ou inacessível.")
		}

		// 6. Detecção de deriva de convenção de nomenclatura (sequencial feat-NN vs hash feat-xxxx)
		fmt.Println("\n🔤 Convenção de Nomenclatura:")
		var sequentialFiles, hashFiles []string
		scanConventionDir := func(dir string) {
			_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}
				rel, relErr := filepath.Rel(targetDir, path)
				if relErr != nil {
					rel = path
				}
				name := d.Name()
				if name == "discovery.md" || name == "criteria.md" || name == "plan.md" {
					name = filepath.Base(filepath.Dir(path))
				}
				switch classifyNamingConvention(name) {
				case "sequencial":
					sequentialFiles = append(sequentialFiles, rel)
				case "hash":
					hashFiles = append(hashFiles, rel)
				}
				return nil
			})
		}
		scanConventionDir(layout.Features(targetDir))
		scanConventionDir(layout.Discovery(targetDir))

		if len(sequentialFiles) > 0 && len(hashFiles) > 0 {
			fmt.Println("  ⚠️  Deriva de convenção detectada: nomenclatura sequencial (feat-NN/fix-NN) e por hash (feat-xxxx/fix-xxxx) coexistem no projeto.")
			fmt.Println("     Sequencial:")
			for _, f := range sequentialFiles {
				fmt.Printf("      - %s\n", f)
			}
			fmt.Println("     Hash:")
			for _, f := range hashFiles {
				fmt.Printf("      - %s\n", f)
			}
		} else if len(sequentialFiles)+len(hashFiles) > 0 {
			fmt.Println("  ✓ Convenção de nomenclatura consistente.")
		}

		if rc.NamingConvention == "" {
			detected := config.DetectNamingConvention(targetDir)
			rc.NamingConvention = detected
			cfg := config.Config{
				Project:          rc.Project,
				SddVersion:       rc.Version,
				Lang:             rc.Lang,
				Telemetry:        rc.Telemetry.Enabled,
				Agents:           rc.Agents,
				NamingConvention: detected,
			}
			err = scaffold.UpdateSddrc(cfg, targetDir)
			if err != nil {
				fmt.Printf("  ⚠️  Falha ao atualizar sdd/.sddrc com a convenção detectada: %v\n", err)
			} else {
				fmt.Printf("  ✓ Convenção de nomenclatura ausente corrigida automaticamente para: %s\n", detected)
			}
		}

		// 7. Validação do nome padrão do projeto
		if rc.Project == "meu-projeto" {
			fmt.Println("\n⚠️  O nome do projeto no sdd/.sddrc ainda está configurado como o padrão ('meu-projeto').")
			fmt.Println("   Sugerimos alterá-lo para refletir o nome real do seu projeto em:")
			fmt.Println("   - Chave \"project\" no arquivo sdd/.sddrc")
			for _, agent := range rc.Agents {
				switch agent {
				case config.AgentClaude:
					fmt.Println("   - Cabeçalho do arquivo CLAUDE.md (onde diz: # CLAUDE.md — meu-projeto)")
				case config.AgentGemini:
					fmt.Println("   - Cabeçalho do arquivo GEMINI.md (onde diz: # GEMINI.md — meu-projeto)")
				case config.AgentOpenAI:
					fmt.Println("   - Cabeçalho do arquivo AGENTS.md (onde diz: # AGENTS.md — meu-projeto)")
				}
			}
			fmt.Println("   - Cabeçalho do arquivo sdd/memory/progress.md")
		}

		// 8. Agregador de telemetria (feat-01-02) — fecha o gap "grava mas nunca lê"
		fmt.Println("\n📊 Telemetria de Sessão:")
		summary, err := AggregateSessionMetrics(targetDir)
		if err != nil {
			fmt.Printf("  ⚠️  Falha ao agregar sdd/.metrics: %v\n", err)
		} else if summary.Total == 0 {
			fmt.Println("  — Nenhuma sessão registrada ainda em sdd/.metrics/.")
		} else {
			fmt.Printf("  %d sessão(ões) registrada(s): %d approved, %d rejected, %d blocked\n",
				summary.Total, summary.Approved, summary.Rejected, summary.Blocked)
		}

		if w := config.AdvancedWarning(targetDir); w != "" {
			fmt.Printf("\n⚠️  Modo avançado (experimental): %s\n", w)
		}

		fmt.Println("\n🩺 Diagnóstico finalizado.")
		if hookSemScript {
			return errors.New("erro de configuração: o hook PreToolUse chama scripts/pr_guard.py, que não existe; rode 'forge update'")
		}
		return nil
	},
}

// reportDoctorVersao compara a versão do projeto com a do binário e imprime o veredito.
func reportDoctorVersao(projeto string) {
	st, msg := doctor.CompararVersao(projeto, config.BinaryVersion())
	switch st {
	case doctor.VersaoOK:
		fmt.Println("✓ " + msg)
	case doctor.VersaoBinarioAntigo:
		fmt.Println("✗ " + msg)
	default:
		fmt.Println("⚠️  " + msg)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// runDoctorHarness imprime o diagnóstico do harness (texto ou JSON) e sai com 3 se algo falhou.
func runDoctorHarness(w io.Writer, dir string, asJSON bool, deps doctor.Deps) error {
	rel := doctor.Diagnose(dir, deps)
	if err := doctor.ValidateRelatorio(rel); err != nil {
		return err
	}
	if asJSON {
		data, err := json.Marshal(rel)
		if err != nil {
			return fmt.Errorf("serializar relatório do doctor: %w", err)
		}
		fmt.Fprintln(w, string(data))
	} else {
		for _, c := range rel.Checagens {
			fmt.Fprintf(w, "[%s] %s: %s\n", c.Status, c.ID, c.Detalhe)
			if c.Sugestao != "" {
				fmt.Fprintf(w, "      sugestão: %s\n", c.Sugestao)
			}
		}
	}
	if code := rel.ExitCode(); code != 0 {
		return exitCodeErr{code: code}
	}
	return nil
}

// reportPRGuardCoverage informa se o hook PreToolUse do pr_guard está instalado e em quais
// agentes a barreira de fato vale. Só o Claude Code executa hooks (feat-106-15); nos demais, o
// job "pronto?" do CI é quem segura o PR em draft.
func reportPRGuardCoverage(targetDir string, agents []string) (hookSemScript bool) {
	settings := filepath.Join(targetDir, ".claude", "settings.json")
	instalado, legado := false, false
	if data, err := os.ReadFile(settings); err == nil {
		txt := string(data)
		novo := strings.Contains(txt, "hook pr-guard")
		antigo := strings.Contains(txt, "pr_guard.py")
		instalado = novo || antigo
		legado = antigo && !novo
	}

	if instalado {
		fmt.Println("  ✓ Hook PreToolUse instalado (.claude/settings.json)")
		// O comando novo (`forge hook pr-guard`) tolera script ausente; só o antigo bloqueia o Bash.
		if _, err := os.Stat(scaffold.PRGuardScriptPath(targetDir)); err != nil && legado {
			hookSemScript = true
			fmt.Println("  ✗ Erro de configuração: o hook existe, mas scripts/pr_guard.py não; o hook bloquearia todo Bash. Rode 'forge update' para gravá-lo")
		}
	} else {
		fmt.Println("  ⚠️  Hook PreToolUse ausente; rode 'forge update --upgrade' para instalar")
	}

	if _, err := exec.LookPath("uv"); err != nil {
		fmt.Println("  ⚠️  'uv' não está no PATH; o hook não roda sem ele (instale em https://docs.astral.sh/uv/)")
	} else {
		fmt.Println("  ✓ 'uv' disponível no PATH")
	}

	var comHook, semHook []string
	for _, a := range agents {
		if a == config.AgentClaude {
			comHook = append(comHook, a)
		} else {
			semHook = append(semHook, a)
		}
	}
	if len(comHook) > 0 && instalado {
		fmt.Printf("  ✓ Cobertura: %s (hook ativo)\n", strings.Join(comHook, ", "))
	} else if len(comHook) > 0 {
		fmt.Printf("  ⚠️  Cobertura: %s configurado, mas o hook não está instalado\n", strings.Join(comHook, ", "))
	}
	if len(semHook) > 0 {
		fmt.Printf("  ℹ️  Sem hook: %s — nesses agentes vale o job \"pronto?\" do CI\n", strings.Join(semHook, ", "))
	}
	return hookSemScript
}

func init() {
	doctorCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	doctorCmd.Flags().Bool("harness", false, "Diagnostica o harness do pipeline (somente leitura; sai com 3 se houver falha)")
	doctorCmd.Flags().Bool("verbose", false, "Lista os caminhos de features/tasks (o padrão imprime só o resumo)")
	doctorCmd.Flags().Bool("json", false, "Com --harness, emite o relatório em JSON")
}
