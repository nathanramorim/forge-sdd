package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

const (
	// StatusProposed indica que a skill foi proposta (DUP-05 / GOV-03) e aguarda aprovação humana.
	StatusProposed = "proposed"
	// StatusApproved indica que a skill foi aprovada e está apta a ser sincronizada para os agentes.
	StatusApproved = "approved"
	// OriginAutoGovernance identifica skills extraídas pelo motor de governança.
	OriginAutoGovernance = "auto-governance"
	// ManagedSkillMarker identifica adaptadores gerados automaticamente por forge-sdd skills sync,
	// permitindo preservar skills manuais criadas diretamente pelo usuário em .claude/skills ou .gemini/skills.
	ManagedSkillMarker = "<!-- forge-sdd:managed-skill -->"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// SkillInfo descreve uma skill registrada na fonte canônica (.agents/skills/<slug>/SKILL.md).
type SkillInfo struct {
	Name        string
	Description string
	Status      string
	Origin      string
	Body        string
	Path        string
}

// ValidateSlug garante que o identificador da skill seja kebab-case não vazio.
func ValidateSlug(slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return fmt.Errorf("slug da skill é obrigatório")
	}
	if !slugRe.MatchString(slug) {
		return fmt.Errorf("slug inválido %q: use kebab-case (apenas letras minúsculas, números e hífens)", slug)
	}
	return nil
}

// Propose cria uma skill proposta em <dir>/.agents/skills/<slug>/SKILL.md (DUP-10)
// e registra/atualiza a entrada correspondente em <dir>/sdd/skills/index.md.
func Propose(dir, slug, description, body string) (string, error) {
	slug = strings.TrimSpace(slug)
	description = strings.TrimSpace(description)
	body = strings.TrimSpace(body)

	if err := ValidateSlug(slug); err != nil {
		return "", err
	}
	if description == "" {
		return "", fmt.Errorf("descrição da skill é obrigatória")
	}

	skillDir := filepath.Join(dir, ".agents", "skills", slug)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar diretório da skill %s: %w", skillDir, err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	content := renderCanonicalSkill(slug, description, StatusProposed, OriginAutoGovernance, body)
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("falha ao gravar %s: %w", skillPath, err)
	}

	relCanonical := filepath.ToSlash(filepath.Join(".agents", "skills", slug, "SKILL.md"))
	if err := upsertIndex(dir, slug, description, StatusProposed, relCanonical); err != nil {
		return "", err
	}

	return skillPath, nil
}

// Approve promove uma skill de proposed para approved em <dir>/.agents/skills/<slug>/SKILL.md,
// atualiza <dir>/sdd/skills/index.md e executa Sync(dir, agents) automaticamente.
func Approve(dir, slug string, agents []string) ([]string, error) {
	slug = strings.TrimSpace(slug)
	if err := ValidateSlug(slug); err != nil {
		return nil, err
	}

	skillPath := filepath.Join(dir, ".agents", "skills", slug, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill %q não encontrada em %s", slug, skillPath)
		}
		return nil, fmt.Errorf("falha ao ler %s: %w", skillPath, err)
	}

	info := parseSkillFile(slug, skillPath, string(data))
	info.Status = StatusApproved
	if info.Origin == "" {
		info.Origin = OriginAutoGovernance
	}

	updated := renderCanonicalSkill(info.Name, info.Description, info.Status, info.Origin, info.Body)
	if err := os.WriteFile(skillPath, []byte(updated), 0o644); err != nil {
		return nil, fmt.Errorf("falha ao atualizar %s: %w", skillPath, err)
	}

	relCanonical := filepath.ToSlash(filepath.Join(".agents", "skills", slug, "SKILL.md"))
	if err := upsertIndex(dir, info.Name, info.Description, StatusApproved, relCanonical); err != nil {
		return nil, err
	}

	return Sync(dir, agents)
}

// Sync propaga todas as skills com status approved em <dir>/.agents/skills/*/SKILL.md
// para os diretórios de skills dos agentes configurados (.claude/skills e .gemini/skills).
// Skills manuais preexistentes (sem ManagedSkillMarker) são preservadas sem sobrescrita.
func Sync(dir string, agents []string) ([]string, error) {
	if len(agents) == 0 {
		if rc, err := config.ReadSddrc(dir); err == nil && len(rc.Agents) > 0 {
			agents = rc.Agents
		} else {
			agents = []string{config.AgentClaude, config.AgentGemini}
		}
	}

	skillsList, err := List(dir)
	if err != nil {
		return nil, err
	}

	var synced []string
	for _, sk := range skillsList {
		if sk.Status != StatusApproved {
			continue
		}
		for _, agent := range agents {
			var adapterDir string
			switch agent {
			case config.AgentClaude:
				adapterDir = filepath.Join(dir, ".claude", "skills", sk.Name)
			case config.AgentGemini:
				adapterDir = filepath.Join(dir, ".gemini", "skills", sk.Name)
			default:
				continue
			}

			adapterPath := filepath.Join(adapterDir, "SKILL.md")
			if existing, err := os.ReadFile(adapterPath); err == nil {
				if !strings.Contains(string(existing), ManagedSkillMarker) {
					// Preserva skill manual criada diretamente pelo usuário.
					continue
				}
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("falha ao inspecionar %s: %w", adapterPath, err)
			}

			if err := os.MkdirAll(adapterDir, 0o755); err != nil {
				return nil, fmt.Errorf("falha ao criar %s: %w", adapterDir, err)
			}

			adapterContent := renderAdapterSkill(sk)
			if err := os.WriteFile(adapterPath, []byte(adapterContent), 0o644); err != nil {
				return nil, fmt.Errorf("falha ao gravar adaptador %s: %w", adapterPath, err)
			}
			synced = append(synced, adapterPath)
		}
	}

	return synced, nil
}

// List retorna todas as skills encontradas em <dir>/.agents/skills/*/SKILL.md ordenadas por nome.
func List(dir string) ([]SkillInfo, error) {
	root := filepath.Join(dir, ".agents", "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []SkillInfo{}, nil
		}
		return nil, fmt.Errorf("falha ao listar %s: %w", root, err)
	}

	var result []SkillInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		skillPath := filepath.Join(root, slug, "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("falha ao ler %s: %w", skillPath, err)
		}
		result = append(result, parseSkillFile(slug, skillPath, string(data)))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func renderCanonicalSkill(slug, description, status, origin, body string) string {
	if origin == "" {
		origin = OriginAutoGovernance
	}
	header := fmt.Sprintf("# Skill: %s", slug)
	trimmedBody := strings.TrimSpace(body)
	if strings.HasPrefix(trimmedBody, "# Skill:") {
		idx := strings.Index(trimmedBody, "\n")
		if idx == -1 {
			trimmedBody = ""
		} else {
			trimmedBody = strings.TrimSpace(trimmedBody[idx+1:])
		}
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", slug))
	b.WriteString(fmt.Sprintf("description: %s\n", description))
	b.WriteString(fmt.Sprintf("status: %s\n", status))
	b.WriteString(fmt.Sprintf("origin: %s\n", origin))
	b.WriteString("---\n\n")
	b.WriteString(header)
	b.WriteString("\n\n")
	if trimmedBody != "" {
		b.WriteString(trimmedBody)
		b.WriteString("\n")
	}
	return b.String()
}

func renderAdapterSkill(sk SkillInfo) string {
	relCanonical := filepath.ToSlash(filepath.Join(".agents", "skills", sk.Name, "SKILL.md"))
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", sk.Name))
	b.WriteString(fmt.Sprintf("description: %s\n", sk.Description))
	b.WriteString("---\n")
	b.WriteString(ManagedSkillMarker + "\n")
	b.WriteString(fmt.Sprintf("<!-- Fonte canônica (DUP-10): %s -->\n\n", relCanonical))
	b.WriteString(fmt.Sprintf("# Skill: %s\n\n", sk.Name))
	if strings.TrimSpace(sk.Body) != "" {
		b.WriteString(strings.TrimSpace(sk.Body))
		b.WriteString("\n")
	}
	return b.String()
}

func parseSkillFile(fallbackSlug, path, content string) SkillInfo {
	info := SkillInfo{
		Name:   fallbackSlug,
		Status: StatusProposed,
		Origin: OriginAutoGovernance,
		Path:   path,
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	body := normalized

	if strings.HasPrefix(normalized, "---\n") {
		rest := normalized[4:]
		if endIdx := strings.Index(rest, "\n---\n"); endIdx != -1 {
			fm := rest[:endIdx]
			body = rest[endIdx+5:]
			for _, line := range strings.Split(fm, "\n") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				switch key {
				case "name":
					if val != "" {
						info.Name = val
					}
				case "description":
					info.Description = val
				case "status":
					if val != "" {
						info.Status = val
					}
				case "origin":
					if val != "" {
						info.Origin = val
					}
				}
			}
		}
	}

	trimmedBody := strings.TrimSpace(body)
	if strings.HasPrefix(trimmedBody, "# Skill:") {
		if idx := strings.Index(trimmedBody, "\n"); idx != -1 {
			trimmedBody = strings.TrimSpace(trimmedBody[idx+1:])
		} else {
			trimmedBody = ""
		}
	}
	info.Body = trimmedBody
	return info
}

const defaultIndexTemplate = `# Skills Index

| Skill | Quando usar | Status | Caminho |
|-------|-------------|--------|---------|

> Builder deve consultar este índice antes de carregar qualquer skill (` + "`DUP-05`" + `, ` + "`DUP-10`" + `).
> Sem entrada ` + "`approved`" + ` aqui → não carregar.
`

func upsertIndex(dir, slug, description, status, relPath string) error {
	skillsDir := filepath.Join(dir, "sdd", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("falha ao criar %s: %w", skillsDir, err)
	}
	indexPath := filepath.Join(skillsDir, "index.md")

	content := defaultIndexTemplate
	if data, err := os.ReadFile(indexPath); err == nil {
		content = strings.ReplaceAll(string(data), "\r\n", "\n")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("falha ao ler %s: %w", indexPath, err)
	}

	newRow := fmt.Sprintf("| %s | %s | %s | %s |", slug, description, status, relPath)
	lines := strings.Split(content, "\n")
	var out []string
	replaced := false
	inserted := false
	lastTableLineIdx := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Remove linha placeholder vazia se existir
		if strings.Contains(trimmed, "(nenhuma skill criada ainda)") {
			continue
		}
		// Atualiza cabeçalhos legados de 3 colunas para 4 colunas se necessário
		if strings.HasPrefix(trimmed, "| Skill") && !strings.Contains(trimmed, "Status") {
			out = append(out, "| Skill | Quando usar | Status | Caminho |")
			lastTableLineIdx = len(out) - 1
			continue
		}
		if strings.HasPrefix(trimmed, "|---") && strings.Count(trimmed, "|") == 4 {
			out = append(out, "|-------|-------------|--------|---------|")
			lastTableLineIdx = len(out) - 1
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			cells := strings.Split(trimmed, "|")
			if len(cells) >= 3 && strings.TrimSpace(cells[1]) == slug {
				out = append(out, newRow)
				replaced = true
				lastTableLineIdx = len(out) - 1
				continue
			}
			lastTableLineIdx = len(out)
		}
		out = append(out, line)
	}

	if !replaced {
		if lastTableLineIdx >= 0 {
			updated := make([]string, 0, len(out)+1)
			updated = append(updated, out[:lastTableLineIdx+1]...)
			updated = append(updated, newRow)
			updated = append(updated, out[lastTableLineIdx+1:]...)
			out = updated
			inserted = true
		}
		if !inserted {
			out = append(out, newRow)
		}
	}

	finalContent := strings.Join(out, "\n")
	if !strings.HasSuffix(finalContent, "\n") {
		finalContent += "\n"
	}

	if err := os.WriteFile(indexPath, []byte(finalContent), 0o644); err != nil {
		return fmt.Errorf("falha ao gravar %s: %w", indexPath, err)
	}
	return nil
}
