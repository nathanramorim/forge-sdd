package gate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Estados do contexto do projeto (D-2): heurística sobre os artefatos, sem campo novo.
const (
	BriefRascunho = "rascunho"
	BriefPronta   = "pronta"
)

const (
	// DefaultBriefBytes é o orçamento padrão do brief (D-1).
	DefaultBriefBytes = 6 * 1024
	minBriefBytes     = 256
	maxBriefBytes     = 1 << 20
	briefReadLimit    = 256 << 10
	briefItemSep      = " · "
)

// BriefSection é uma linha do brief: um rótulo, seus itens (cada um inteiro) e quantos foram cortados
// pelo orçamento. Source é o arquivo a ler para ver o restante.
type BriefSection struct {
	Label   string
	Items   []string
	Omitted int
	Source  string
}

// Brief é a visão curta e determinística das decisões do projeto, montada só por leitura (D-1).
type Brief struct {
	Project  string
	State    string
	Missing  []string
	Sections []BriefSection
	MaxBytes int
	Size     int
}

// briefShare é a fatia do orçamento que cada seção pode ocupar; o restante cobre cabeçalho, estado, missão e stack.
var briefShare = map[string]float64{
	"Regras": 0.10, "IDs": 0.22, "ADRs vigentes": 0.12, "Módulos": 0.10,
	"Fluxos": 0.08, "Decisões": 0.10, "Segurança": 0.07, "Rules": 0.11,
}

var (
	briefTitleRe  = regexp.MustCompile(`^#\s+Constitui[çc][ãa]o\s+[—–-]\s+(.+?)\s*$`)
	briefRuleRe   = regexp.MustCompile(`^\s*(\d+)\.\s+(.*\S)\s*$`)
	briefIDTailRe = regexp.MustCompile("\\s*\\(`[A-Z]{3}-\\d{2}`\\)\\.?\\s*$")
	briefADRRe    = regexp.MustCompile(`^adr-(\d{4})-.*\.md$`)
	briefADRHead  = regexp.MustCompile(`^#\s+(ADR-\d{4}):?\s*(.*\S)\s*$`)
	briefStatusRe = regexp.MustCompile(`^\s*[-*]\s+\*\*Status:?\*\*:?\s*(.*)$`)
	briefHeadRe   = regexp.MustCompile(`^(#{2,3})\s+(.*\S)\s*$`)
	briefNumRe    = regexp.MustCompile(`^\d+[.)]\s+`)
	briefIDRe     = regexp.MustCompile(`^([A-Z]{3})-(\d{2})$`)
	briefDecIDRe  = regexp.MustCompile(`^D\d+$`)
	briefSepRe    = regexp.MustCompile(`^:?-{2,}:?$`)
)

// briefSkipHeads são títulos estruturais (não nomes de módulo ou fluxo) ignorados ao listar cabeçalhos.
var briefSkipHeads = []string{"diagrama", "detalhamento", "índice", "indice", "visão geral", "visao geral", "overview"}

// BuildBrief lê constitution, sdd/spec, ADRs e rules de dir e monta o brief dentro de maxBytes
// (0 usa o padrão). Somente leitura; arquivo ausente vira seção vazia, não erro.
func BuildBrief(dir string, maxBytes int) (Brief, error) {
	if maxBytes == 0 {
		maxBytes = DefaultBriefBytes
	}
	if maxBytes < minBriefBytes || maxBytes > maxBriefBytes {
		return Brief{}, fmt.Errorf("brief: orçamento %d fora do intervalo %d..%d bytes", maxBytes, minBriefBytes, maxBriefBytes)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return Brief{}, fmt.Errorf("brief: diretório %q inexistente ou inválido", dir)
	}

	const (
		constitutionRel = "sdd/memory/constitution.md"
		modulesRel      = "sdd/spec/modules.md"
		flowsRel        = "sdd/spec/flows.md"
		decisionsRel    = "sdd/spec/decisions.md"
	)
	constitution := briefRead(dir, constitutionRel)
	modules := briefRead(dir, modulesRel)
	flows := briefRead(dir, flowsRel)
	decisionsDoc := briefRead(dir, decisionsRel)

	b := Brief{Project: briefProject(dir, constitution), MaxBytes: maxBytes}

	mission := briefMission(constitution)
	moduleItems := briefHeadings(modules, "<", "[")
	flowItems := briefHeadings(flows, "<", "[")
	decisionItems, decisionSrc := briefDecisions(constitution, decisionsDoc, decisionsRel, constitutionRel)
	securityItems, securitySrc := briefSecurity(dir, constitution, constitutionRel)

	if mission == "" {
		b.Missing = append(b.Missing, "missão (constitution.md)")
	}
	if len(decisionItems) == 0 {
		b.Missing = append(b.Missing, "decisões ("+decisionsRel+")")
	}
	if !briefFilled(modules, moduleItems, "<módulo-principal>", "[descrição]") {
		b.Missing = append(b.Missing, "módulos ("+modulesRel+")")
	}
	if !briefFilled(flows, flowItems, "[entrada]", "[processamento]") {
		b.Missing = append(b.Missing, "fluxos ("+flowsRel+")")
	}
	b.State = BriefPronta
	if len(b.Missing) > 0 {
		b.State = BriefRascunho
	}

	var stack []string
	for _, row := range briefTable(briefSection(constitution, "stack")) {
		if len(row) >= 2 && row[0] != "" && row[1] != "" {
			stack = append(stack, row[0]+" "+row[1])
		}
	}
	if len(stack) == 0 {
		for _, row := range briefTable(briefSection(briefRead(dir, "sdd/spec/stack.md"), "dependências", "dependencias")) {
			if len(row) >= 2 && row[0] != "" && row[1] != "" {
				stack = append(stack, row[0]+" "+row[1])
			}
		}
	}

	estado := b.State
	if len(b.Missing) > 0 {
		estado += ". Faltam: " + strings.Join(b.Missing, ", ")
	}
	var missionItems []string
	if mission != "" {
		missionItems = []string{briefCut(mission, 200)}
	}

	b.Sections = []BriefSection{
		{Label: "Estado", Items: []string{estado}, Source: constitutionRel},
		{Label: "Missão", Items: missionItems, Source: constitutionRel},
		{Label: "Stack", Items: briefCutAll(stack, 40), Source: constitutionRel},
		{Label: "Regras", Items: briefRules(constitution), Source: constitutionRel},
		{Label: "IDs", Items: briefIDs(constitution), Source: constitutionRel},
		{Label: "ADRs vigentes", Items: briefADRs(dir), Source: "sdd/spec/adr/"},
		{Label: "Módulos", Items: briefCutAll(moduleItems, 60), Source: modulesRel},
		{Label: "Fluxos", Items: briefCutAll(flowItems, 60), Source: flowsRel},
		{Label: "Decisões", Items: decisionItems, Source: decisionSrc},
		{Label: "Segurança", Items: securityItems, Source: securitySrc},
		{Label: "Rules", Items: briefRuleFiles(dir), Source: ".agents/rules/"},
	}
	b.fit()
	return b, nil
}

// fit aplica o orçamento: cabeçalho, estado, missão e stack entram sempre; as demais seções têm fatia
// própria e o corte é por item inteiro, com a linha "… N itens omitidos; leia <arquivo>".
func (b *Brief) fit() {
	used := len(b.header(1 << 20)) // reserva o cabeçalho com o maior tamanho possível
	for i := range b.Sections {
		s := &b.Sections[i]
		total := len(s.Items)
		if total == 0 {
			continue
		}
		budget := b.MaxBytes - used
		if share, ok := briefShare[s.Label]; ok {
			if cap := int(share * float64(b.MaxBytes)); cap < budget {
				budget = cap
			}
		}
		keep := total
		if s.Label != "Estado" && s.Label != "Missão" && s.Label != "Stack" {
			keep = 0
			for k := total; k >= 0; k-- {
				if len(s.render(k))+1 <= budget {
					keep = k
					break
				}
			}
			if keep == 0 && len(s.render(0))+1 > budget {
				s.Omitted = total
				s.Items = nil
				continue // nem a linha de omissão cabe; seção de menor prioridade fica de fora
			}
		}
		s.Omitted = total - keep
		s.Items = s.Items[:keep]
		used += len(s.render(keep)) + 1
	}
	b.Size = len(b.Render())
}

// render devolve a linha da seção com os k primeiros itens e, se houver corte, a linha de omissão.
func (s BriefSection) render(k int) string {
	if k > len(s.Items) {
		k = len(s.Items)
	}
	omitted := len(s.Items) - k
	var line string
	if k > 0 {
		line = s.Label + ": " + strings.Join(s.Items[:k], briefItemSep)
	}
	if omitted > 0 {
		note := fmt.Sprintf("… %d itens omitidos; leia %s", omitted, s.Source)
		if line == "" {
			return note
		}
		return line + "\n" + note
	}
	return line
}

func (s BriefSection) lines() string {
	if len(s.Items) == 0 && s.Omitted == 0 {
		return ""
	}
	line := ""
	if len(s.Items) > 0 {
		line = s.Label + ": " + strings.Join(s.Items, briefItemSep)
	}
	if s.Omitted > 0 {
		note := fmt.Sprintf("… %d itens omitidos; leia %s", s.Omitted, s.Source)
		if line == "" {
			return note
		}
		return line + "\n" + note
	}
	return line
}

func (b Brief) header(size int) string {
	return fmt.Sprintf("# Contexto do projeto — %s   (%s · %s)\n", b.Project, b.State, briefKB(size))
}

// Render formata o brief para leitura por agente ou pessoa. Seções vazias não aparecem.
func (b Brief) Render() string {
	var body strings.Builder
	for _, s := range b.Sections {
		if text := s.lines(); text != "" {
			body.WriteString(text)
			body.WriteString("\n")
		}
	}
	size := len(b.header(0)) + body.Len()
	size = len(b.header(size)) + body.Len() // o tamanho mostrado inclui o próprio cabeçalho
	return b.header(size) + body.String()
}

func briefKB(size int) string {
	return strings.Replace(fmt.Sprintf("%.1f KB", float64(size)/1024), ".", ",", 1)
}

func briefRead(dir, rel string) string {
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, briefReadLimit))
	if err != nil {
		return ""
	}
	return string(data)
}

func briefProject(dir, constitution string) string {
	for _, line := range strings.Split(constitution, "\n") {
		if m := briefTitleRe.FindStringSubmatch(line); m != nil {
			return strings.TrimSpace(m[1])
		}
		break
	}
	if abs, err := filepath.Abs(dir); err == nil {
		return filepath.Base(abs)
	}
	return filepath.Base(dir)
}

// briefSection devolve o corpo da seção "## título" (nível 2) cujo título começa com um dos prefixos.
func briefSection(content string, prefixes ...string) string {
	var out []string
	inFence, in := false, false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
		} else if !inFence && strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "###") {
			if in {
				break
			}
			if strings.HasPrefix(line, "## ") {
				title := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
				for _, p := range prefixes {
					if strings.HasPrefix(title, strings.ToLower(p)) {
						in = true
					}
				}
				continue
			}
		}
		if in {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func briefMission(constitution string) string {
	for _, line := range strings.Split(briefSection(constitution, "missão", "missao"), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "<!--") {
			continue
		}
		if strings.Contains(t, "[Uma frase") {
			return "" // missão-modelo ainda não preenchida
		}
		return strings.ReplaceAll(t, "`", "")
	}
	return ""
}

// briefTable devolve as linhas de dados (sem cabeçalho e separador) das tabelas markdown de body.
func briefTable(body string) [][]string {
	var rows [][]string
	header := true
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "|") {
			header = true
			continue
		}
		cells := strings.Split(strings.Trim(t, "|"), "|")
		for i := range cells {
			cells[i] = strings.ReplaceAll(strings.TrimSpace(cells[i]), "`", "")
		}
		if header {
			header = false
			continue
		}
		sep := true
		for _, c := range cells {
			if !briefSepRe.MatchString(c) {
				sep = false
			}
		}
		empty := true
		for _, c := range cells {
			if c != "" {
				empty = false
			}
		}
		if sep || empty {
			continue
		}
		rows = append(rows, cells)
	}
	return rows
}

// briefHeadings lista os cabeçalhos ## e ### de content (fora de blocos de código), sem os estruturais
// nem os que começam com algum dos marcadores de modelo.
func briefHeadings(content string, templateMarks ...string) []string {
	var out []string
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		m := briefHeadRe.FindStringSubmatch(line)
		if inFence || m == nil {
			continue
		}
		title := strings.ReplaceAll(briefNumRe.ReplaceAllString(m[2], ""), "`", "")
		low := strings.ToLower(title)
		skip := false
		for _, s := range briefSkipHeads {
			if strings.Contains(low, s) {
				skip = true
			}
		}
		for _, mark := range templateMarks {
			if strings.HasPrefix(title, mark) {
				skip = true
			}
		}
		if !skip {
			out = append(out, title)
		}
	}
	return out
}

// briefFilled informa se o arquivo existe, não guarda texto-modelo e tem ao menos um item.
func briefFilled(content string, items []string, templateMarks ...string) bool {
	if strings.TrimSpace(content) == "" || len(items) == 0 {
		return false
	}
	for _, mark := range templateMarks {
		if strings.Contains(content, mark) {
			return false
		}
	}
	return true
}

func briefDecisions(constitution, doc, docRel, constitutionRel string) ([]string, string) {
	var items []string
	for _, row := range briefTable(briefSection(doc, "resolvidas", "decisões resolvidas")) {
		if len(row) >= 3 && briefDecIDRe.MatchString(row[0]) {
			items = append(items, briefCut(row[0]+" "+row[1]+": "+row[2], 80))
		} else if len(row) >= 2 {
			items = append(items, briefCut(row[0]+": "+row[1], 80))
		}
	}
	if len(items) > 0 {
		return items, docRel
	}
	for _, row := range briefTable(briefSection(constitution, "decisões resolvidas", "decisoes resolvidas")) {
		if len(row) >= 2 {
			items = append(items, briefCut(row[0]+": "+row[1], 80))
		}
	}
	return items, constitutionRel
}

// briefSecurity lê a seção "## Segurança" da Constitution ou, na falta, de overview.md e decisions.md.
func briefSecurity(dir, constitution, constitutionRel string) ([]string, string) {
	sources := []struct{ rel, content string }{
		{constitutionRel, constitution},
		{"sdd/spec/overview.md", briefRead(dir, "sdd/spec/overview.md")},
		{"sdd/spec/decisions.md", briefRead(dir, "sdd/spec/decisions.md")},
	}
	for _, src := range sources {
		var items []string
		for _, line := range strings.Split(briefSection(src.content, "segurança", "seguranca"), "\n") {
			t := strings.TrimSpace(line)
			t = strings.TrimSpace(strings.TrimLeft(t, "-*"))
			if t == "" || strings.HasPrefix(t, "|") || strings.HasPrefix(t, "<!--") {
				continue
			}
			items = append(items, briefCut(strings.ReplaceAll(t, "`", ""), 90))
		}
		if len(items) > 0 {
			return items, src.rel
		}
	}
	return nil, constitutionRel
}

func briefRules(constitution string) []string {
	var out []string
	for _, line := range strings.Split(briefSection(constitution, "regras"), "\n") {
		m := briefRuleRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := briefIDTailRe.ReplaceAllString(m[2], "")
		out = append(out, m[1]+" "+briefCut(briefFirstSentence(text), 60))
	}
	return out
}

func briefIDs(constitution string) []string {
	lines := ConstitutionIDLines(constitution)
	ids := make([]string, 0, len(lines))
	for id := range lines {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id+" "+briefCut(briefFirstSentence(lines[id]), 70))
	}
	return out
}

// briefADRs lista as ADRs vigentes (fora o modelo 0000 e as substituídas ou rejeitadas), em ordem de arquivo.
func briefADRs(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, "sdd", "spec", "adr"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		m := briefADRRe.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil || m[1] == "0000" {
			continue
		}
		var id, title, status string
		for _, line := range strings.Split(briefRead(dir, "sdd/spec/adr/"+e.Name()), "\n") {
			if h := briefADRHead.FindStringSubmatch(line); h != nil && id == "" {
				id, title = h[1], h[2]
			} else if s := briefStatusRe.FindStringSubmatch(line); s != nil && status == "" {
				status = strings.ReplaceAll(strings.TrimSpace(s[1]), "`", "")
			}
		}
		low := strings.ToLower(status)
		if id == "" || strings.HasPrefix(low, "substitu") || strings.HasPrefix(low, "rejeitad") {
			continue
		}
		if fields := strings.Fields(status); len(fields) > 0 {
			status = fields[0]
		}
		item := id + " " + briefCut(title, 60)
		if status != "" {
			item += " (" + status + ")"
		}
		out = append(out, item)
	}
	return out
}

// briefRuleFiles lista as rules de .agents/rules com `regras:` no frontmatter: caminho, IDs cobertos e descrição.
func briefRuleFiles(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "rules"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		content := briefRead(dir, ".agents/rules/"+e.Name())
		ids, _, ok := parseRuleFrontmatterRegras(content)
		if !ok {
			continue
		}
		item := ".agents/rules/" + e.Name() + " (" + briefCompactIDs(ids) + ")"
		if desc := briefFrontmatterValue(content, "description"); desc != "" {
			item += ": " + briefCut(briefFirstSentence(desc), 50)
		}
		out = append(out, item)
	}
	return out
}

// briefCompactIDs agrupa IDs por prefixo: SEG-01, SEG-02, ARQ-01 vira "ARQ-01 SEG-01,02".
func briefCompactIDs(ids []string) string {
	var order []string
	nums := map[string][]string{}
	for _, id := range ids {
		m := briefIDRe.FindStringSubmatch(id)
		if m == nil {
			order = append(order, id)
			continue
		}
		if _, seen := nums[m[1]]; !seen {
			order = append(order, m[1])
		}
		nums[m[1]] = append(nums[m[1]], m[2])
	}
	sort.Strings(order)
	parts := make([]string, 0, len(order))
	for _, o := range order {
		if n, ok := nums[o]; ok {
			parts = append(parts, o+"-"+strings.Join(n, ","))
		} else {
			parts = append(parts, o)
		}
	}
	return strings.Join(parts, " ")
}

func briefFrontmatterValue(content, key string) string {
	lines := strings.Split(strings.TrimLeft(content, "\r\n "), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), key+":"); ok {
			return strings.Trim(strings.TrimSpace(rest), `"'`)
		}
	}
	return ""
}

// briefFirstSentence devolve o texto até o primeiro ponto final seguido de espaço (ou o texto todo), sem crases.
func briefFirstSentence(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "`", "")
	if i := strings.Index(s, ". "); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(strings.TrimSuffix(s, "."))
}

// briefCut limita s a n runas, terminando com "…" quando corta; nunca devolve quebra de linha.
func briefCut(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n-1])) + "…"
}

func briefCutAll(items []string, n int) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, briefCut(it, n))
	}
	return out
}
