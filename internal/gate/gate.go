package gate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	ModeObservacao = "observacao"
	ModeBloqueio   = "bloqueio"

	MaxProgressBytes = 1024
)

// Options configura a execução determinística da Camada A do Portão.
type Options struct {
	Branch       string
	SpecPath     string
	GoModChanged bool
	PRBodyPath   string
	PRBodyText   string
	Now          time.Time
}

// CheckResult representa o resultado de uma verificação individual P-xx da Camada A.
type CheckResult struct {
	ID          string `json:"id"`
	Rule        string `json:"rule"`
	Passed      bool   `json:"passed"`
	Evidence    string `json:"evidence"`
	Remediation string `json:"remediation"`
}

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`PRIVATE\s+KEY-{5}`),
		regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`),
		regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	}
	ruleIDRegex = regexp.MustCompile(`\b(?:ARQ|SEG|DUP|COD|TST|FLX|GOV)-\d{2}\b`)
)

// HasFailures indica se há ao menos uma verificação reprovada no relatório.
func HasFailures(results []CheckResult) bool {
	for _, r := range results {
		if !r.Passed {
			return true
		}
	}
	return false
}

// RunChecks executa as verificações determinísticas da Camada A do Portão.
func RunChecks(dir string, opts Options) ([]CheckResult, error) {
	if dir == "" {
		dir = "."
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var results []CheckResult

	// P-01 (FLX-02): branch válida (não main/master)
	if opts.Branch != "" {
		results = append(results, checkP01Branch(opts.Branch))
	}

	// P-02 (FLX-01) e P-03 (DUP-02): existência da Spec e tabela "O que já existe"
	if opts.SpecPath != "" {
		p02, specContent := checkP02SpecExists(dir, opts.SpecPath)
		results = append(results, p02)
		if p02.Passed {
			results = append(results, checkP03ReuseTable(opts.SpecPath, specContent))
		} else {
			results = append(results, CheckResult{
				ID:          "P-03",
				Rule:        "DUP-02",
				Passed:      false,
				Evidence:    fmt.Sprintf("spec %q ausente; impossível validar seção '## O que já existe'", opts.SpecPath),
				Remediation: "Criar o arquivo de Spec e preencher a seção '## O que já existe' com tabela de decisões (reusar, estender ou criar).",
			})
		}

		// P-14 (SEG-07): se GoModChanged, exige seção "## Dependências novas" não-vazia na Spec
		if opts.GoModChanged {
			results = append(results, checkP14NewDependencies(opts.SpecPath, specContent, p02.Passed))
		}
	}

	// P-13 (SEG-04): varredura de segredos explícitos
	p13, err := checkP13Secrets(dir)
	if err != nil {
		return nil, err
	}
	results = append(results, p13)

	// P-20 (GOV-02): validação de blocos constitution-exception
	if opts.PRBodyPath != "" || opts.PRBodyText != "" {
		results = append(results, checkP20ConstitutionException(dir, opts.PRBodyPath, opts.PRBodyText, now))
	}

	// P-21 (Regra 14 / Orçamento): sdd/memory/progress.md existe e <= 1024 bytes
	results = append(results, checkP21ProgressBudget(dir))

	// P-23 (GOV-04): .agents/rules/*.md referenciam IDs existentes na constitution e sem cópia literal
	results = append(results, checkP23RulesGovernance(dir))

	return results, nil
}

// ValidateFLX01 executa as verificações determinísticas P-02 (existência da Spec) e P-03 (tabela de reuso DUP-02)
// para homologação determinística do Portão FLX-01 (Spec antes do código).
func ValidateFLX01(dir, specPath string) ([]CheckResult, bool) {
	p02, specContent := checkP02SpecExists(dir, specPath)
	if !p02.Passed {
		p03 := CheckResult{
			ID:          "P-03",
			Rule:        "DUP-02",
			Passed:      false,
			Evidence:    fmt.Sprintf("spec %q ausente; impossível validar seção '## O que já existe'", specPath),
			Remediation: "Criar o arquivo de Spec e preencher a seção '## O que já existe' com tabela de decisões (reusar, estender ou criar).",
		}
		return []CheckResult{p02, p03}, false
	}
	p03 := checkP03ReuseTable(specPath, specContent)
	results := []CheckResult{p02, p03}
	return results, !HasFailures(results)
}

// ValidateFLX01WithBranch valida P-01 (se branch não for vazia), P-02 e P-03 para o Portão FLX-01.
func ValidateFLX01WithBranch(dir, specPath, branch string) ([]CheckResult, bool) {
	var results []CheckResult
	if strings.TrimSpace(branch) != "" {
		p01 := checkP01Branch(branch)
		results = append(results, p01)
	}
	flxChecks, _ := ValidateFLX01(dir, specPath)
	results = append(results, flxChecks...)
	return results, !HasFailures(results)
}

func checkP01Branch(branch string) CheckResult {
	b := strings.TrimSpace(branch)
	if b == "main" || b == "master" {
		return CheckResult{
			ID:          "P-01",
			Rule:        "FLX-02",
			Passed:      false,
			Evidence:    fmt.Sprintf("branch atual é %q (commits diretos em main/master são proibidos)", b),
			Remediation: "Criar ou alternar para uma branch de feature/fix (ex: git checkout -b feat/XX-nome).",
		}
	}
	return CheckResult{
		ID:          "P-01",
		Rule:        "FLX-02",
		Passed:      true,
		Evidence:    fmt.Sprintf("branch %q válida", b),
		Remediation: "",
	}
}

func checkP02SpecExists(dir, specPath string) (CheckResult, string) {
	resolved := specPath
	if !filepath.IsAbs(specPath) {
		resolved = filepath.Join(dir, specPath)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return CheckResult{
			ID:          "P-02",
			Rule:        "FLX-01",
			Passed:      false,
			Evidence:    fmt.Sprintf("arquivo de Spec não encontrado em %q", specPath),
			Remediation: "Criar o arquivo de Spec da feature/task antes da implementação.",
		}, ""
	}
	return CheckResult{
		ID:          "P-02",
		Rule:        "FLX-01",
		Passed:      true,
		Evidence:    fmt.Sprintf("Spec encontrada em %q (%d bytes)", specPath, len(data)),
		Remediation: "",
	}, string(data)
}

func checkP03ReuseTable(specPath, content string) CheckResult {
	lines := strings.Split(content, "\n")
	inSection := false
	validRows := 0

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "## ") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "o que já existe") || strings.Contains(lower, "o que ja existe") {
				inSection = true
				continue
			}
			if inSection {
				break
			}
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			cells := splitTableCells(line)
			if len(cells) < 2 || isTableSeparator(cells) || isTableHeader(cells) {
				continue
			}
			rowLower := strings.ToLower(line)
			hasDecision := strings.Contains(rowLower, "reusar") ||
				strings.Contains(rowLower, "estender") ||
				strings.Contains(rowLower, "criar")
			// Justificativa é a última coluna não-vazia
			lastCell := strings.TrimSpace(cells[len(cells)-1])
			hasJustification := lastCell != "" && lastCell != "-" && lastCell != "—"
			if hasDecision && hasJustification {
				validRows++
			}
		}
	}

	if !inSection {
		return CheckResult{
			ID:          "P-03",
			Rule:        "DUP-02",
			Passed:      false,
			Evidence:    fmt.Sprintf("seção '## O que já existe' ausente em %q", specPath),
			Remediation: "Adicionar a seção '## O que já existe' com tabela contendo decisão (reusar, estender ou criar) e justificativa.",
		}
	}
	if validRows == 0 {
		return CheckResult{
			ID:          "P-03",
			Rule:        "DUP-02",
			Passed:      false,
			Evidence:    fmt.Sprintf("seção '## O que já existe' em %q não possui linha de tabela com decisão (reusar/estender/criar) e justificativa", specPath),
			Remediation: "Preencher ao menos uma linha na tabela de '## O que já existe' com decisão (reusar, estender ou criar) e justificativa.",
		}
	}
	return CheckResult{
		ID:          "P-03",
		Rule:        "DUP-02",
		Passed:      true,
		Evidence:    fmt.Sprintf("seção '## O que já existe' válida com %d linha(s) de decisão", validRows),
		Remediation: "",
	}
}

func checkP14NewDependencies(specPath, content string, specExists bool) CheckResult {
	if !specExists {
		return CheckResult{
			ID:          "P-14",
			Rule:        "SEG-07",
			Passed:      false,
			Evidence:    "go.mod alterado, mas arquivo de Spec não existe",
			Remediation: "Criar Spec e documentar a seção '## Dependências novas'.",
		}
	}

	lines := strings.Split(content, "\n")
	inSection := false
	var bodyLines []string

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "## ") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "dependências novas") || strings.Contains(lower, "dependencias novas") {
				inSection = true
				continue
			}
			if inSection {
				break
			}
		}
		if inSection && line != "" {
			bodyLines = append(bodyLines, line)
		}
	}

	joined := strings.TrimSpace(strings.Join(bodyLines, " "))
	lowerJoined := strings.ToLower(joined)
	if !inSection || joined == "" || lowerJoined == "nenhuma" || lowerJoined == "nenhuma." || lowerJoined == "-" || lowerJoined == "—" {
		return CheckResult{
			ID:          "P-14",
			Rule:        "SEG-07",
			Passed:      false,
			Evidence:    fmt.Sprintf("go.mod foi alterado, mas %q não possui seção '## Dependências novas' justificada", specPath),
			Remediation: "Preencher a seção '## Dependências novas' na Spec justificando o pacote adicionado.",
		}
	}

	return CheckResult{
		ID:          "P-14",
		Rule:        "SEG-07",
		Passed:      true,
		Evidence:    fmt.Sprintf("seção '## Dependências novas' preenchida em %q", specPath),
		Remediation: "",
	}
}

func checkP13Secrets(dir string) (CheckResult, error) {
	var findings []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isScannableTextFile(d.Name()) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}

		for lineIdx, line := range strings.Split(string(data), "\n") {
			for _, re := range secretPatterns {
				if match := re.FindString(line); match != "" {
					findings = append(findings, fmt.Sprintf("%s:%d (%s)", filepath.ToSlash(rel), lineIdx+1, maskSecret(match)))
				}
			}
		}
		return nil
	})
	if err != nil {
		return CheckResult{}, fmt.Errorf("erro ao varrer diretório para P-13: %w", err)
	}

	if len(findings) > 0 {
		return CheckResult{
			ID:          "P-13",
			Rule:        "SEG-04",
			Passed:      false,
			Evidence:    fmt.Sprintf("padrão de segredo detectado em: %s", strings.Join(findings, ", ")),
			Remediation: "Remover segredos/chaves privadas do código e rotacionar a credencial exposta.",
		}, nil
	}

	return CheckResult{
		ID:          "P-13",
		Rule:        "SEG-04",
		Passed:      true,
		Evidence:    "nenhum padrão de segredo explícito encontrado",
		Remediation: "",
	}, nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return s
	}
	return s[:6] + "..."
}

func isScannableTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go", ".md", ".json", ".yaml", ".yml", ".toml", ".env", ".sh", ".txt", ".js", ".ts", ".py", ".tmpl":
		return true
	}
	return strings.HasPrefix(name, ".env") || name == ".sddrc"
}

func checkP20ConstitutionException(dir, prBodyPath, prBodyText string, now time.Time) CheckResult {
	text := prBodyText
	if prBodyPath != "" {
		resolved := prBodyPath
		if !filepath.IsAbs(prBodyPath) {
			resolved = filepath.Join(dir, prBodyPath)
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return CheckResult{
				ID:          "P-20",
				Rule:        "GOV-02",
				Passed:      false,
				Evidence:    fmt.Sprintf("falha ao ler arquivo de PR body %q: %v", prBodyPath, err),
				Remediation: "Informar um caminho válido em --pr-body.",
			}
		}
		text = string(data)
	}

	if !strings.Contains(text, "constitution-exception:") {
		return CheckResult{
			ID:          "P-20",
			Rule:        "GOV-02",
			Passed:      true,
			Evidence:    "nenhum bloco constitution-exception declarado",
			Remediation: "",
		}
	}

	lines := strings.Split(text, "\n")
	blockCount := 0
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "constitution-exception:") {
			blockCount++
			fields := map[string]string{}
			for j := i + 1; j < len(lines); j++ {
				sub := strings.TrimSpace(lines[j])
				if sub == "" || strings.HasPrefix(sub, "```") || strings.HasPrefix(sub, "#") || strings.HasPrefix(sub, "constitution-exception:") {
					break
				}
				sub = strings.TrimPrefix(sub, "- ")
				if colon := strings.Index(sub, ":"); colon > 0 {
					k := strings.ToLower(strings.TrimSpace(sub[:colon]))
					v := strings.Trim(strings.TrimSpace(sub[colon+1:]), `"'`)
					fields[k] = v
				}
			}

			for _, reqKey := range []string{"regra", "motivo", "prazo", "responsavel"} {
				if strings.TrimSpace(fields[reqKey]) == "" {
					return CheckResult{
						ID:          "P-20",
						Rule:        "GOV-02",
						Passed:      false,
						Evidence:    fmt.Sprintf("bloco constitution-exception #%d sem campo obrigatório %q", blockCount, reqKey),
						Remediation: "Todo bloco constitution-exception deve conter regra, motivo, prazo (AAAA-MM-DD) e responsavel.",
					}
				}
			}

			prazoStr := fields["prazo"]
			prazoTime, err := time.Parse("2006-01-02", prazoStr)
			if err != nil {
				return CheckResult{
					ID:          "P-20",
					Rule:        "GOV-02",
					Passed:      false,
					Evidence:    fmt.Sprintf("bloco constitution-exception #%d com prazo inválido %q (esperado AAAA-MM-DD)", blockCount, prazoStr),
					Remediation: "Definir 'prazo' no formato AAAA-MM-DD com data igual ou posterior a hoje.",
				}
			}

			nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			if prazoTime.Before(nowDate) {
				return CheckResult{
					ID:          "P-20",
					Rule:        "GOV-02",
					Passed:      false,
					Evidence:    fmt.Sprintf("bloco constitution-exception #%d com prazo expirado (%s < %s)", blockCount, prazoStr, nowDate.Format("2006-01-02")),
					Remediation: "Renovar ou remover a exceção vencida e corrigir a violação subjacente.",
				}
			}
		}
	}

	return CheckResult{
		ID:          "P-20",
		Rule:        "GOV-02",
		Passed:      true,
		Evidence:    fmt.Sprintf("%d bloco(s) constitution-exception válido(s)", blockCount),
		Remediation: "",
	}
}

func checkP21ProgressBudget(dir string) CheckResult {
	progressPath := filepath.Join(dir, "sdd", "memory", "progress.md")
	info, err := os.Stat(progressPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CheckResult{
				ID:          "P-21",
				Rule:        "Regra 14 / Orçamento",
				Passed:      false,
				Evidence:    "arquivo sdd/memory/progress.md não encontrado",
				Remediation: "Criar ou restaurar sdd/memory/progress.md com no máximo 1024 bytes.",
			}
		}
		return CheckResult{
			ID:          "P-21",
			Rule:        "Regra 14 / Orçamento",
			Passed:      false,
			Evidence:    fmt.Sprintf("erro ao verificar sdd/memory/progress.md: %v", err),
			Remediation: "Verificar permissões de leitura em sdd/memory/progress.md.",
		}
	}

	size := info.Size()
	if size > MaxProgressBytes {
		return CheckResult{
			ID:          "P-21",
			Rule:        "Regra 14 / Orçamento",
			Passed:      false,
			Evidence:    fmt.Sprintf("sdd/memory/progress.md possui %d bytes (limite máximo: %d bytes)", size, MaxProgressBytes),
			Remediation: "Arquivar entradas concluídas em sdd/memory/progress-log.md para manter progress.md <= 1024 bytes.",
		}
	}

	return CheckResult{
		ID:          "P-21",
		Rule:        "Regra 14 / Orçamento",
		Passed:      true,
		Evidence:    fmt.Sprintf("sdd/memory/progress.md dentro do orçamento (%d/%d bytes)", size, MaxProgressBytes),
		Remediation: "",
	}
}

func checkP23RulesGovernance(dir string) CheckResult {
	rulesDir := filepath.Join(dir, ".agents", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CheckResult{
				ID:          "P-23",
				Rule:        "GOV-04",
				Passed:      true,
				Evidence:    "diretório .agents/rules ausente (nenhuma rule com frontmatter para validar)",
				Remediation: "",
			}
		}
		return CheckResult{
			ID:          "P-23",
			Rule:        "GOV-04",
			Passed:      false,
			Evidence:    fmt.Sprintf("erro ao ler .agents/rules: %v", err),
			Remediation: "Verificar permissões do diretório .agents/rules.",
		}
	}

	constitutionPath := filepath.Join(dir, "sdd", "memory", "constitution.md")
	constData, err := os.ReadFile(constitutionPath)
	if err != nil {
		return CheckResult{
			ID:          "P-23",
			Rule:        "GOV-04",
			Passed:      false,
			Evidence:    "sdd/memory/constitution.md não encontrado para validar IDs das rules",
			Remediation: "Garantir que sdd/memory/constitution.md exista e contenha os IDs canônicos.",
		}
	}
	constText := string(constData)
	constRuleLines := extractConstitutionRuleLines(constText)

	checkedFiles := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		rulePath := filepath.Join(rulesDir, entry.Name())
		data, readErr := os.ReadFile(rulePath)
		if readErr != nil {
			continue
		}
		content := string(data)
		ids, body, hasRegras := parseRuleFrontmatterRegras(content)
		if !hasRegras {
			continue
		}
		checkedFiles++

		for _, id := range ids {
			if !strings.Contains(constText, id) {
				return CheckResult{
					ID:          "P-23",
					Rule:        "GOV-04",
					Passed:      false,
					Evidence:    fmt.Sprintf("rule %q referencia ID inexistente na constitution: %q", entry.Name(), id),
					Remediation: "Cadastrar o ID em sdd/memory/constitution.md ou corrigir o frontmatter 'regras:' em .agents/rules/.",
				}
			}
		}

		// Verifica se o corpo da rule não contém cópia literal de linhas inteiras de regras da constitution
		for _, rLine := range constRuleLines {
			if strings.Contains(body, rLine) {
				return CheckResult{
					ID:          "P-23",
					Rule:        "GOV-04",
					Passed:      false,
					Evidence:    fmt.Sprintf("rule %q contém cópia literal de linha de regra da constitution: %q", entry.Name(), rLine),
					Remediation: "Remover a cópia literal da regra da constitution e manter na rule apenas o 'como cumprir' operacional (GOV-04).",
				}
			}
		}
	}

	return CheckResult{
		ID:          "P-23",
		Rule:        "GOV-04",
		Passed:      true,
		Evidence:    fmt.Sprintf("%d arquivo(s) de rule com frontmatter validado(s) contra a constitution", checkedFiles),
		Remediation: "",
	}
}

func extractConstitutionRuleLines(constText string) []string {
	var out []string
	for _, raw := range strings.Split(constText, "\n") {
		line := strings.TrimSpace(raw)
		if len(line) < 25 {
			continue
		}
		if ruleIDRegex.MatchString(line) {
			out = append(out, line)
		}
	}
	return out
}

func parseRuleFrontmatterRegras(content string) (ids []string, body string, hasRegras bool) {
	trimmed := strings.TrimLeft(content, "\r\n ")
	if !strings.HasPrefix(trimmed, "---") {
		return nil, content, false
	}
	lines := strings.Split(trimmed, "\n")
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return nil, content, false
	}

	frontmatter := strings.Join(lines[1:endIdx], "\n")
	body = strings.Join(lines[endIdx+1:], "\n")

	fmLines := lines[1:endIdx]
	for i := 0; i < len(fmLines); i++ {
		line := strings.TrimSpace(fmLines[i])
		if strings.HasPrefix(line, "regras:") {
			hasRegras = true
			rest := strings.TrimSpace(strings.TrimPrefix(line, "regras:"))
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				inner := strings.TrimSuffix(strings.TrimPrefix(rest, "["), "]")
				for _, part := range strings.Split(inner, ",") {
					id := strings.Trim(strings.TrimSpace(part), `"'`)
					if id != "" {
						ids = append(ids, id)
					}
				}
			} else if rest == "" {
				for j := i + 1; j < len(fmLines); j++ {
					sub := strings.TrimSpace(fmLines[j])
					if strings.HasPrefix(sub, "- ") {
						id := strings.Trim(strings.TrimSpace(strings.TrimPrefix(sub, "- ")), `"'`)
						if id != "" {
							ids = append(ids, id)
						}
					} else if strings.Contains(sub, ":") {
						break
					}
				}
			}
		}
	}
	_ = frontmatter
	return ids, body, hasRegras
}

func splitTableCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func isTableSeparator(cells []string) bool {
	for _, c := range cells {
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

func isTableHeader(cells []string) bool {
	first := strings.ToLower(cells[0])
	return strings.Contains(first, "procurei") || strings.Contains(first, "item") || strings.Contains(first, "componente")
}
