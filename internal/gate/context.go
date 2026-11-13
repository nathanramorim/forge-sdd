package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/skills"
)

// ContextID é uma regra da Constitution citada pela spec, com o texto original.
type ContextID struct {
	ID   string
	Text string
}

// ContextRule é um arquivo de .agents/rules/ aplicável à spec e os IDs citados que ele cobre.
type ContextRule struct {
	Path   string
	Covers []string
}

// ContextPack é o contexto mínimo para executar uma task: a spec, só as linhas da Constitution
// que ela cita, e os caminhos das rules e skills aplicáveis (nada é carregado além disso).
type ContextPack struct {
	SpecPath string
	Spec     string
	IDs      []ContextID
	Rules    []ContextRule
	Skills   []string
}

var (
	constitutionIDLine = regexp.MustCompile(`^\s*[-*]\s+\*\*((?:ARQ|SEG|DUP|COD|TST|FLX|GOV)-\d{2}):?\*\*:?\s*(.*)$`)
	ruleFileName       = regexp.MustCompile(`\bsdd-[a-z0-9-]+\.md\b`)
	validStations      = map[string]bool{"concepcao": true, "planejamento": true, "build": true, "governanca": true, "operacao": true}
)

// ConstitutionIDLines devolve o texto de cada regra com ID (linhas "- **ID:** texto") da Constitution.
func ConstitutionIDLines(constitution string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(constitution, "\n") {
		if m := constitutionIDLine.FindStringSubmatch(line); m != nil {
			out[m[1]] = strings.TrimSpace(m[2])
		}
	}
	return out
}

// BuildContext monta o contexto mínimo de uma task. specPath é relativo a dir (ou absoluto);
// extraIDs são IDs adicionais (por exemplo, os da fila). IDs inexistentes na Constitution
// geram erro enriquecido. station é uma de: concepcao, planejamento, build, governanca, operacao.
func BuildContext(dir, specPath string, extraIDs []string, station string) (ContextPack, error) {
	var pack ContextPack
	if !validStations[station] {
		return pack, fmt.Errorf("contexto: estação inválida %q (use concepcao, planejamento, build, governanca ou operacao)", station)
	}

	resolved, err := confineToDir(dir, specPath)
	if err != nil {
		return pack, err
	}
	spec, err := os.ReadFile(resolved)
	if err != nil {
		return pack, fmt.Errorf("contexto: ler spec %q: %w", specPath, err)
	}
	pack.SpecPath = specPath
	pack.Spec = string(spec)

	constitution, err := os.ReadFile(layout.Constitution(dir))
	if err != nil {
		return pack, fmt.Errorf("contexto: ler sdd/memory/constitution.md: %w", err)
	}
	known := ConstitutionIDLines(string(constitution))

	idSet := map[string]bool{}
	for _, id := range ruleIDRegex.FindAllString(pack.Spec, -1) {
		idSet[id] = true
	}
	for _, id := range extraIDs {
		if ruleIDRegex.MatchString(id) {
			idSet[id] = true
		}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var missing []string
	for _, id := range ids {
		text, ok := known[id]
		if !ok {
			missing = append(missing, id)
			continue
		}
		pack.IDs = append(pack.IDs, ContextID{ID: id, Text: text})
	}
	if len(missing) > 0 {
		return pack, fmt.Errorf("contexto: ID(s) citado(s) em %s sem definição em sdd/memory/constitution.md: %s", specPath, strings.Join(missing, ", "))
	}

	pack.Rules = applicableRules(dir, ids, pack.Spec)
	pack.Skills = applicableSkills(dir, station, pack.Spec)
	return pack, nil
}

// confineToDir resolve p relativo a dir e recusa caminho absoluto, ".." e links simbólicos que
// levem para fora de dir (SEG-01): o spec_path vem de queue.json, que é versionado e vem de outras branches.
func confineToDir(dir, p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("contexto: caminho absoluto %q recusado; use um caminho relativo ao projeto (SEG-01)", p)
	}
	joined := filepath.Join(dir, p)
	rel, err := filepath.Rel(dir, joined)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("contexto: caminho %q sai do projeto (SEG-01)", p)
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("contexto: resolver %q: %w", dir, err)
	}
	realFile, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", fmt.Errorf("contexto: ler spec %q: %w", p, err)
	}
	back, err := filepath.Rel(realDir, realFile)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("contexto: %q aponta para fora do projeto por link simbólico (SEG-01)", p)
	}
	return joined, nil
}

func applicableRules(dir string, ids []string, spec string) []ContextRule {
	rulesDir := filepath.Join(dir, ".agents", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil
	}
	named := map[string]bool{}
	for _, n := range ruleFileName.FindAllString(spec, -1) {
		named[n] = true
	}
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}

	var rules []ContextRule
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(rulesDir, e.Name()))
		if err != nil {
			continue
		}
		covered, _, hasRegras := parseRuleFrontmatterRegras(string(data))
		var covers []string
		if hasRegras {
			for _, id := range covered {
				if wanted[id] {
					covers = append(covers, id)
				}
			}
		}
		if len(covers) > 0 || named[e.Name()] {
			rules = append(rules, ContextRule{Path: ".agents/rules/" + e.Name(), Covers: covers})
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Path < rules[j].Path })
	return rules
}

func applicableSkills(dir, station, spec string) []string {
	var out []string
	core := ".agents/skills/sdd-" + station + "/SKILL.md"
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(core))); err == nil {
		out = append(out, core)
	}
	list, err := skills.List(dir)
	if err != nil {
		return out
	}
	for _, sk := range list {
		if sk.Status != skills.StatusApproved || sk.Scope == skills.ScopeCore {
			continue
		}
		if mentionsWord(spec, sk.Name) {
			out = append(out, ".agents/skills/"+sk.Name+"/SKILL.md")
		}
	}
	return out
}

// mentionsWord informa se name aparece em text como palavra inteira (sem casar prefixos de outros nomes).
func mentionsWord(text, name string) bool {
	re, err := regexp.Compile(`(^|[^a-zA-Z0-9-])` + regexp.QuoteMeta(name) + `($|[^a-zA-Z0-9-])`)
	return err == nil && re.MatchString(text)
}

// Render formata o contexto para leitura por agente ou pessoa.
func (p ContextPack) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Contexto: %s\n\n## Spec\n\n%s\n", p.SpecPath, strings.TrimRight(p.Spec, "\n"))

	fmt.Fprintf(&b, "\n## IDs da Constituição (%d)\n\n", len(p.IDs))
	for _, id := range p.IDs {
		fmt.Fprintf(&b, "- **%s:** %s\n", id.ID, id.Text)
	}

	b.WriteString("\n## Rules a carregar\n\n")
	if len(p.Rules) == 0 {
		b.WriteString("- (nenhuma)\n")
	}
	for _, r := range p.Rules {
		if len(r.Covers) > 0 {
			fmt.Fprintf(&b, "- %s (cobre %s)\n", r.Path, strings.Join(r.Covers, ", "))
		} else {
			fmt.Fprintf(&b, "- %s (citada na spec)\n", r.Path)
		}
	}

	b.WriteString("\n## Skills a carregar\n\n")
	if len(p.Skills) == 0 {
		b.WriteString("- (nenhuma)\n")
	}
	for _, s := range p.Skills {
		fmt.Fprintf(&b, "- %s\n", s)
	}
	return b.String()
}
