package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/forge-sdd/cli/internal/config"
)

//go:embed all:templates
var templatesFS embed.FS

// Walk retorna a lista de paths dos templates embutidos (sem o prefixo "templates/").
func Walk() []string {
	var paths []string
	_ = fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

// agentTemplateRoot retorna o prefixo de template para cada agente.
// Retorna "" para agentes sem subdiretório dedicado (copilot usa as pastas globais).
func agentTemplateRoot(agent string) string {
	switch agent {
	case config.AgentClaude:
		return "templates/agents/claude"
	case config.AgentGemini:
		return "templates/agents/gemini"
	case config.AgentOpenAI:
		return "templates/agents/openai"
	default:
		return ""
	}
}

// Run renderiza os templates com cfg e escreve os arquivos em targetDir.
// Retorna a lista de arquivos criados.
func Run(cfg config.Config, targetDir string) ([]string, error) {
	var created []string
	backupTimestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(targetDir, "sdd", "backups", "upgrade-"+backupTimestamp)

	// Garante ao menos um agente
	agents := cfg.Agents
	if len(agents) == 0 {
		agents = []string{config.AgentCopilot}
	}

	// 1. Templates globais (sdd/, .vscode/, e .github/ do copilot)
	//    Sempre renderizados — são agnósticos ao agente de IA.
	globalRoots := []string{"templates/sdd"}

	// .vscode/ e .github/ só são incluídos quando copilot está selecionado
	copilotSelected := false
	for _, a := range agents {
		if a == config.AgentCopilot {
			copilotSelected = true
			break
		}
	}
	if copilotSelected {
		globalRoots = append(globalRoots, "templates/.vscode", "templates/.github")
	}

	for _, root := range globalRoots {
		// stripPrefix = "templates" para preservar o nome do subdiretório no destino
		// ex: templates/.github/foo.tmpl → .github/foo
		files, err := renderDir(templatesFS, root, "templates", cfg, targetDir, backupDir)
		if err != nil {
			return nil, err
		}
		created = append(created, files...)
	}

	// 2. Templates específicos por agente
	for _, agent := range agents {
		root := agentTemplateRoot(agent)
		if root == "" {
			continue // copilot já foi tratado acima
		}
		// stripPrefix = root do agente para que CLAUDE.md fique na raiz do projeto
		// ex: templates/agents/claude/CLAUDE.md.tmpl → CLAUDE.md
		files, err := renderDir(templatesFS, root, root, cfg, targetDir, backupDir)
		if err != nil {
			return nil, err
		}
		created = append(created, files...)
	}

	cleanObsoleteFiles(targetDir)
	return created, nil
}

// renderDir percorre fsys a partir de root, renderiza cada template e escreve em targetDir.
// stripPrefix é o prefixo a remover do path para calcular o destino.
func renderDir(fsys embed.FS, root, stripPrefix string, cfg config.Config, targetDir string, backupDir string) ([]string, error) {
	var created []string

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// destPath: strip prefix e ".tmpl" suffix
		rel := strings.TrimPrefix(path, stripPrefix+"/")
		rel = strings.TrimSuffix(rel, ".tmpl")
		dest := filepath.Join(targetDir, rel)

		if cfg.DryRun {
			fmt.Printf("[DRY] %s\n", dest)
			created = append(created, dest)
			return nil
		}

		// Se o arquivo destino já existe e deve ser preservado, pulamos a escrita física
		if shouldPreserve(dest, targetDir) {
			created = append(created, dest)
			return nil
		}

		// ANTES de escrever fisicamente, se o arquivo já existir no destino,
		// fazemos um backup dele para sdd/backups/upgrade-<timestamp>/
		if _, err := os.Stat(dest); err == nil {
			if err := backupFile(dest, targetDir, backupDir); err != nil {
				fmt.Printf("erro ao fazer backup de %s: %v\n", dest, err)
			} else {
				fmt.Printf("[BACKUP] %s -> %s\n", dest, filepath.Join(backupDir, rel))
			}
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}

		data, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}

		tmpl, err := template.New(path).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", path, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("render template %s: %w", path, err)
		}

		if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}

		created = append(created, dest)
		return nil
	})

	return created, err
}

// RunAgents renderiza apenas os templates de agente(s) especificados, sem tocar em sdd/.
// Útil para o comando update, que adiciona novos agentes a um projeto existente.
func RunAgents(cfg config.Config, agents []string, targetDir string) ([]string, error) {
	var created []string
	backupTimestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(targetDir, "sdd", "backups", "upgrade-"+backupTimestamp)

	for _, agent := range agents {
		if agent == config.AgentCopilot {
			for _, root := range []string{"templates/.vscode", "templates/.github"} {
				files, err := renderDir(templatesFS, root, "templates", cfg, targetDir, backupDir)
				if err != nil {
					return nil, err
				}
				created = append(created, files...)
			}
			continue
		}
		root := agentTemplateRoot(agent)
		if root == "" {
			continue
		}
		files, err := renderDir(templatesFS, root, root, cfg, targetDir, backupDir)
		if err != nil {
			return nil, err
		}
		created = append(created, files...)
	}
	cleanObsoleteFiles(targetDir)
	return created, nil
}

// UpdateSddrc re-renderiza sdd/.sddrc com a config fornecida, sobrescrevendo o existente.
// Ignora shouldPreserve para que o campo agents seja sempre atualizado.
func UpdateSddrc(cfg config.Config, targetDir string) error {
	const tmplPath = "templates/sdd/.sddrc.tmpl"
	data, err := templatesFS.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", tmplPath, err)
	}
	tmpl, err := template.New(tmplPath).Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", tmplPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("render template %s: %w", tmplPath, err)
	}
	dest := filepath.Join(targetDir, "sdd", ".sddrc")
	return os.WriteFile(dest, buf.Bytes(), 0644)
}

// shouldPreserve retorna true se o arquivo de destino já existe e deve ser preservado.
func shouldPreserve(dest string, targetDir string) bool {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return false
	}

	rel, err := filepath.Rel(targetDir, dest)
	if err != nil {
		return false
	}

	// Padroniza separadores para "/"
	rel = filepath.ToSlash(rel)

	// Regras de preservação de domínio sob a pasta sdd/
	if strings.HasPrefix(rel, "sdd/") {
		// Exceções estruturais que DEVEM ser atualizadas/sobrescritas:
		if rel == "sdd/.sdd-version" || rel == "sdd/.sddrc" {
			return false
		}
		// Todo o restante do domínio (memory, spec, features, releases, README, HOWTO, plan) é preservado
		return true
	}

	// Arquivos fora de sdd/ (configurações e instruções de agente) devem ser atualizados
	return false
}

// cleanObsoleteFiles remove arquivos de templates descontinuados ou obsoletos.
func cleanObsoleteFiles(targetDir string) {
	obsolete := []string{
		filepath.Join(".github", "prompts", "install-skill.prompt.md"),
		filepath.Join(".gemini", "prompts", "install-skill.prompt.md"),
		filepath.Join(".claude", "commands", "install-skill.prompt.md"),
	}
	for _, rel := range obsolete {
		path := filepath.Join(targetDir, rel)
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(path)
		}
	}
}

// backupFile salva uma cópia de segurança do arquivo original antes de sobrescrevê-lo.
func backupFile(dest string, targetDir string, backupDir string) error {
	// Se o arquivo de destino não existe, não há o que fazer backup
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return nil
	}

	// Calcula o caminho relativo do arquivo em relação a targetDir
	rel, err := filepath.Rel(targetDir, dest)
	if err != nil {
		return err
	}

	backupDest := filepath.Join(backupDir, rel)
	if err := os.MkdirAll(filepath.Dir(backupDest), 0755); err != nil {
		return err
	}

	// Lê o arquivo original
	content, err := os.ReadFile(dest)
	if err != nil {
		return err
	}

	// Grava o backup
	return os.WriteFile(backupDest, content, 0644)
}

