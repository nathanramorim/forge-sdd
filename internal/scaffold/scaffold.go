package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/forge-sdd/cli/internal/notify"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"text/template"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
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
	refusals.Store(0)

	if !cfg.DryRun {
		if err := migrateLegacyAgentDir(targetDir); err != nil {
			return nil, err
		}
	}

	// Garante ao menos um agente
	agents := cfg.Agents
	if len(agents) == 0 {
		agents = []string{config.AgentCopilot}
	}

	// 1. Globais (sempre) e 2. específicos por agente; a lista é a mesma do plano do update.
	for _, r := range raizesDoRun(cfg) {
		files, err := renderDir(templatesFS, r.root, r.strip, cfg, targetDir)
		if err != nil {
			return nil, err
		}
		created = append(created, files...)
	}

	if !cfg.DryRun {
		cleanObsoleteFiles(targetDir)
		if err := applyStructureUpgrade(cfg, targetDir); err != nil {
			return created, err
		}
		installClaudeInboxHook(agents, targetDir, cfg.SddVersion)
		if _, err := applyAgentModels(targetDir); err != nil {
			return created, err
		}
		if n := refusals.Load(); n > 0 {
			if fi, err := os.Lstat(layout.Sddrc(targetDir)); err != nil || !fi.Mode().IsRegular() {
				return created, fmt.Errorf("instalação abortada: %d destino(s) recusado(s) por segurança (SEG-02) e o projeto ficou sem sdd/.sddrc; revise links simbólicos no projeto", n)
			}
		}
	}
	return created, nil
}

type raizTemplate struct{ root, strip string }

// raizesDoRun lista as raízes de template que Run renderiza, na ordem: globais (sdd/, .agents/, .forge/,
// e .vscode/.github do copilot) e depois as de cada agente. Única fonte (DUP-06) de Run e do plano.
func raizesDoRun(cfg config.Config) []raizTemplate {
	agents := cfg.Agents
	if len(agents) == 0 {
		agents = []string{config.AgentCopilot}
	}
	raizes := []raizTemplate{
		{"templates/sdd", "templates"}, {"templates/.agents", "templates"}, {"templates/.forge", "templates"},
	}
	if slices.Contains(agents, config.AgentCopilot) {
		raizes = append(raizes, raizTemplate{"templates/.vscode", "templates"}, raizTemplate{"templates/.github", "templates"})
	}
	for _, agent := range agents {
		if root := agentTemplateRoot(agent); root != "" {
			raizes = append(raizes, raizTemplate{root, root})
		}
	}
	return raizes
}

// installClaudeInboxHook liga o hook do inbox quando o Claude está entre os agentes. A mesclagem é
// segura (nunca sobrescreve a configuração do usuário); se não for possível, só avisa (SEG-02).
func installClaudeInboxHook(agents []string, targetDir, version string) {
	for _, a := range agents {
		if a != config.AgentClaude {
			continue
		}
		if _, err := MergeClaudeInboxHook(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: hook do inbox não instalado: %s\n", notify.Sanitize(err.Error(), 400, false))
		}
		if _, err := MergeClaudeSessionStartHook(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: hook SessionStart do doctor não instalado: %s\n", notify.Sanitize(err.Error(), 400, false))
		}
		if _, err := installPRGuard(targetDir, version, false); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: hook PreToolUse do pr_guard não instalado: %s\n", notify.Sanitize(err.Error(), 400, false))
		}
		if _, err := MergeClaudeMCPServer(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: servidor MCP não registrado: %s\n", notify.Sanitize(err.Error(), 400, false))
		}
		return
	}
}

// applyStructureUpgrade move o legado de sdd/ para sdd/archive/ (com backup) e remove os
// artefatos de agente obsoletos idênticos ao template anterior. Idempotente: sem legado, não faz nada.
func applyStructureUpgrade(cfg config.Config, targetDir string) error {
	plan, err := PlanLayoutMigration(targetDir)
	if err != nil {
		return err
	}
	if _, err := ApplyLayoutMigration(targetDir, plan); err != nil {
		return err
	}
	if _, err := RemoveObsoleteAgentArtifacts(targetDir, cfg); err != nil {
		return err
	}
	_, _ = config.MigrateLegacyProfiles(targetDir)
	return nil
}

// refusals conta os destinos recusados por segurança na instalação corrente.
var refusals atomic.Int64

// renderDir percorre fsys a partir de root, renderiza cada template e escreve em targetDir.
// stripPrefix é o prefixo a remover do path para calcular o destino.
func renderDir(fsys embed.FS, root, stripPrefix string, cfg config.Config, targetDir string) ([]string, error) {
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

		data, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %s: %w", path, err)
		}

		rendered, err := renderTemplate(path, data, cfg)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if d := decideDestino(relSlash, dest, targetDir, rendered, cfg); d.Preservar {
			switch d.Tipo {
			case tipoRegra:
				avisaRegraPreservada(targetDir, dest, relSlash, rendered)
			case tipoInstrucao:
				avisaInstrucaoPreservada(targetDir, dest, rendered)
			}
			created = append(created, dest)
			return nil
		}

		perm := os.FileMode(0644)
		if strings.HasSuffix(dest, ".sh") {
			perm = 0755 // scripts do pipeline (forge-up.sh) precisam ser executáveis
		}
		// SEG-02: destino derivado de dado externo (link simbólico, pai fora do projeto, FIFO) é
		// recusado com aviso; os demais arquivos seguem.
		if err := writeFileSafeMode(targetDir, dest, rendered, perm); err != nil {
			refusals.Add(1)
			fmt.Fprintf(os.Stderr, "aviso: %s não instalado: %s\n", notify.Sanitize(dest, 300, false), notify.Sanitize(err.Error(), 400, false))
			return nil
		}

		created = append(created, dest)
		return nil
	})

	return created, err
}

const (
	tipoInstrucao = "instrucao"
	tipoRegra     = "regra"
	tipoOutro     = "outro"
)

// decisao é o veredito sobre um destino: preservar (com o tipo, que define o aviso) ou escrever.
type decisao struct {
	Preservar bool
	Tipo      string
}

// decideDestino é a única regra de "escrever ou preservar" (DUP-06), usada por renderDir e pelo plano
// do `update --dry-run`. Não escreve nada.
func decideDestino(rel, dest, targetDir string, rendered []byte, cfg config.Config) decisao {
	tipo := tipoOutro
	if isInstrucao(rel) {
		tipo = tipoInstrucao
	}
	// Instrução sem marcador: só a 5.0 idêntica ao template é adotada; senão é do usuário.
	adotada := isInstrucao(rel) && !IsGeneratedInstrucao(dest) && adotavelDe50(rel, dest, cfg)

	// Rules sdd-* e schema.json: do forge (igual à 5.0 ou ao template atual) são atualizados;
	// divergentes são do usuário (preservados, com aviso e .novo ao lado).
	if regraGerenciada(rel) {
		if _, serr := os.Lstat(dest); serr == nil {
			if !arquivoDoForge(rel, dest, rendered, cfg) {
				return decisao{Preservar: true, Tipo: tipoRegra}
			}
			adotada = true
		}
	}
	if !adotada && shouldPreserve(dest, targetDir) {
		return decisao{Preservar: true, Tipo: tipo}
	}
	return decisao{Tipo: tipo}
}

func renderTemplate(path string, data []byte, cfg config.Config) ([]byte, error) {
	tmpl, err := template.New(path).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, fmt.Errorf("render template %s: %w", path, err)
	}
	return buf.Bytes(), nil
}

// RunAgents renderiza apenas os templates de agente(s) especificados, sem tocar em sdd/.
// Útil para o comando update, que adiciona novos agentes a um projeto existente.
func RunAgents(cfg config.Config, agents []string, targetDir string) ([]string, error) {
	var created []string
	refusals.Store(0)

	if !cfg.DryRun {
		if err := migrateLegacyAgentDir(targetDir); err != nil {
			return nil, err
		}
	}

	for _, agent := range agents {
		if agent == config.AgentCopilot {
			for _, root := range []string{"templates/.vscode", "templates/.github"} {
				files, err := renderDir(templatesFS, root, "templates", cfg, targetDir)
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
		files, err := renderDir(templatesFS, root, root, cfg, targetDir)
		if err != nil {
			return nil, err
		}
		created = append(created, files...)
	}
	if !cfg.DryRun {
		cleanObsoleteFiles(targetDir)
		installClaudeInboxHook(agents, targetDir, cfg.SddVersion)
		if _, err := applyAgentModels(targetDir); err != nil {
			return created, err
		}
		if n := refusals.Load(); n > 0 && len(created) == 0 {
			return created, fmt.Errorf("nada instalado: %d destino(s) recusado(s) por segurança (SEG-02); revise links simbólicos no projeto", n)
		}
	}
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
	return writeFileSafe(targetDir, layout.Sddrc(targetDir), buf.Bytes())
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

	// scripts/pr_guard.py: sem o marcador `# forge-sdd:generated` é do usuário e nunca é sobrescrito
	// (a regra vem antes do laço genérico; a atualização por versão está em installPRGuardScript).
	if rel == prGuardScriptRel {
		return !IsGeneratedPRGuard(targetDir)
	}

	// Regras de preservação de domínio sob a pasta sdd/
	if strings.HasPrefix(rel, "sdd/") {
		// Exceções estruturais que DEVEM ser atualizadas/sobrescritas:
		if rel == path.Join(layout.SddDirName, layout.VersionFileName) || rel == path.Join(layout.SddDirName, layout.SddrcFileName) {
			return false
		}
		// Todo o restante do domínio (memory, spec, features, releases, README, HOWTO, plan) é preservado
		return true
	}

	// .agents/rules/ é conteúdo do usuário (regras de domínio livres) — nunca sobrescrito
	// uma vez criado, mesmo espírito de preservação de domínio usado em sdd/.
	// As skills de estação (.agents/skills/sdd-*) são geradas pelo forge-sdd e devem
	// continuar sendo atualizadas a cada init/update, como qualquer outro arquivo de
	// configuração de agente. Já as demais skills (aprovadas ou propostas pelo usuário
	// ou pela governança) e as rules pertencem ao projeto e nunca são sobrescritas.
	if strings.HasPrefix(rel, ".agents/skills/sdd-") {
		return false
	}
	if strings.HasPrefix(rel, ".agents/rules/") || strings.HasPrefix(rel, ".agents/skills/") ||
		strings.HasPrefix(rel, ".agents/plugins/") {
		return true
	}

	// AGENTS.md é um arquivo neutro que muitos projetos já têm: só é atualizado se foi gerado pelo
	// forge-sdd (cabeçalho e marcador `<!-- forge-sdd:generated -->`); um arquivo do usuário nunca é sobrescrito.
	if isInstrucao(rel) {
		return !IsGeneratedInstrucao(dest)
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
	// feat-03-01/02: .claude/commands/*.prompt.md foi renomeado para *.md (o
	// Claude Code descobre slash commands pelo nome do arquivo sem ".md" — o
	// infixo ".prompt" quebrava a descoberta). .claude/commands/ nunca é
	// preservado por shouldPreserve (é sempre regenerado a cada init/update),
	// então o arquivo antigo só é removido explicitamente aqui — mesmo padrão
	// já usado acima para install-skill.prompt.md.
	for _, cmd := range commandOrder {
		obsolete = append(obsolete, filepath.Join(".claude", "commands", cmd+".prompt.md"))
	}
	for _, rel := range obsolete {
		path := filepath.Join(targetDir, rel)
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(path)
		}
	}
}

// migrateLegacyAgentDir migra o antigo diretório ".agent" (fonte única de agente,
// pré-renomeação) para ".agents" em projetos que já tinham sido escaffoldados com
// o nome anterior. É um no-op se ".agent" não existir. Nunca perde conteúdo do
// usuário: quando ".agents" ainda não existe, faz um rename atômico (preserva
// rules/ e commands/ como estavam); quando ".agents" já existe (update rodado
// mais de uma vez), mescla arquivo a arquivo sem nunca sobrescrever um arquivo
// já presente no destino, e só então remove o diretório legado.
func migrateLegacyAgentDir(targetDir string) error {
	legacy := filepath.Join(targetDir, ".agent")
	current := filepath.Join(targetDir, ".agents")

	if _, err := os.Stat(legacy); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat %s: %w", legacy, err)
	}

	if _, err := os.Stat(current); os.IsNotExist(err) {
		if err := os.Rename(legacy, current); err != nil {
			return fmt.Errorf("migrar %s para %s: %w", legacy, current, err)
		}
		return nil
	}

	err := filepath.WalkDir(legacy, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(legacy, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(current, rel)
		if _, err := os.Lstat(dest); err == nil {
			// já existe no destino migrado — nunca sobrescreve.
			return nil
		}
		if !d.Type().IsRegular() {
			fmt.Fprintf(os.Stderr, "aviso: %s não migrado: não é arquivo regular (SEG-02)\n", notify.Sanitize(path, 300, false))
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := writeFileSafe(targetDir, dest, data); err != nil {
			fmt.Fprintf(os.Stderr, "aviso: %s não migrado: %s\n", notify.Sanitize(dest, 300, false), notify.Sanitize(err.Error(), 400, false))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("migrar conteúdo de %s para %s: %w", legacy, current, err)
	}

	if err := os.RemoveAll(legacy); err != nil {
		return fmt.Errorf("remover diretório legado %s: %w", legacy, err)
	}
	return nil
}
