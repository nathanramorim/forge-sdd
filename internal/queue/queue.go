package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/forge-sdd/cli/internal/harness"
	"github.com/forge-sdd/cli/internal/layout"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	StatusQueued           = "queued"
	StatusBuilding         = "building"
	StatusReviewPending    = "review_pending"
	StatusApproved         = "approved"
	StatusChangesRequested = "changes_requested"
	StatusBlocked          = "blocked"

	VerdictConforme = harness.VereditoConforme
	VerdictViolacao = harness.VereditoViolacao
	VerdictDuvida   = harness.VereditoDuvida
)

// Queue representa a fila persistida em sdd/discovery/<id>/queue.json (ou na fila legada sdd/.queue/queue.json).
type Queue struct {
	FeatureID  string      `json:"feature_id"`
	Branch     string      `json:"branch"`
	BaseBranch string      `json:"base_branch,omitempty"`
	Origem     string      `json:"origem,omitempty"`
	SpecStatus string      `json:"spec_status"`
	UpdatedAt  string      `json:"updated_at"`
	Items      []QueueItem `json:"items"`
}

// QueueItem representa uma tarefa individual na fila do Builder.
type QueueItem struct {
	TaskID            string   `json:"task_id"`
	Title             string   `json:"title"`
	SpecPath          string   `json:"spec_path"`
	Status            string   `json:"status"`
	DependsOn         []string `json:"depends_on,omitempty"`
	Rules             []string `json:"rules"`
	SkillsUsed        []string `json:"skills_used"`
	Attempts          int      `json:"attempts"`
	BlockedReason     string   `json:"blocked_reason,omitempty"`
	GovernanceVerdict *Verdict `json:"governance_verdict,omitempty"`
}

// Verdict é o Parecer da Camada B; o dono do contrato é internal/harness (DUP-01).
type Verdict = harness.Parecer

// Finding é um achado do parecer (harness.Retorno).
type Finding = harness.Retorno

// CandidateSkill é a proposta de auto-skill do parecer (harness.CandidateSkill).
type CandidateSkill = harness.CandidateSkill

// QueueFilePath retorna o caminho de persistência da fila no diretório alvo.
func QueueFilePath(dir string) string {
	return layout.QueuePath(dir)
}

// Load lê a fila da iniciativa ativa (sdd/discovery/<id>/queue.json) ou, sem iniciativa, de sdd/.queue/queue.json.
// Se o arquivo não existir, retorna uma fila vazia inicializada.
func Load(dir string) (*Queue, error) {
	path := QueueFilePath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Queue{
				SpecStatus: "reviewed",
				Items:      []QueueItem{},
			}, nil
		}
		return nil, fmt.Errorf("falha ao ler %s: %w", path, err)
	}

	var q Queue
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("falha ao decodificar %s: %w", path, err)
	}
	if q.SpecStatus == "" {
		q.SpecStatus = "reviewed"
	}
	if q.Items == nil {
		q.Items = []QueueItem{}
	}
	return &q, nil
}

// Save persiste a fila (no caminho de QueueFilePath) atualizando UpdatedAt de forma atômica
// e sincronizando as filas das estações builder.json e governance.json.
func Save(dir string, q *Queue) error {
	path := QueueFilePath(dir)
	if q.SpecStatus == "" {
		q.SpecStatus = "reviewed"
	}
	q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := WriteJSONAtomic(path, q, 0o644); err != nil {
		return fmt.Errorf("falha ao gravar fila %s: %w", path, err)
	}

	// Sincroniza com as filas canônicas de estações
	_ = SyncLegacyQueue(dir, q)
	return nil
}

// ValidateSpecReuseCheck verifica P-02 (existência da spec) e P-03 (seção "## O que já existe" com linha de tabela preenchida).
func ValidateSpecReuseCheck(dir, specPath string) error {
	resolved := specPath
	if !filepath.IsAbs(specPath) {
		resolved = filepath.Join(dir, specPath)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("P-02 (FLX-01): arquivo de spec não encontrado em %q", specPath)
		}
		return fmt.Errorf("P-02 (FLX-01): erro ao ler spec %q: %w", specPath, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	inSection := false
	dataRows := 0

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
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
			cells := splitMarkdownRow(line)
			if len(cells) < 2 {
				continue
			}
			if isSeparatorRow(cells) || isHeaderRow(cells) {
				continue
			}
			nonEmpty := 0
			for _, c := range cells {
				if c != "" && c != "-" && c != "—" {
					nonEmpty++
				}
			}
			if nonEmpty >= 2 {
				dataRows++
			}
		}
	}

	if !inSection {
		return fmt.Errorf("P-03 (DUP-02): spec %q não contém a seção obrigatória '## O que já existe'", specPath)
	}
	if dataRows == 0 {
		return fmt.Errorf("P-03 (DUP-02): seção '## O que já existe' em %q não possui nenhuma linha de tabela preenchida", specPath)
	}
	return nil
}

func splitMarkdownRow(line string) []string {
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

func isSeparatorRow(cells []string) bool {
	for _, c := range cells {
		cleaned := strings.Trim(c, "-: ")
		if cleaned != "" {
			return false
		}
	}
	return true
}

func isHeaderRow(cells []string) bool {
	first := strings.ToLower(cells[0])
	return strings.Contains(first, "procurei") || strings.Contains(first, "item") || strings.Contains(first, "componente")
}

// Enqueue valida a spec (P-02 e P-03) e adiciona ou re-enfileira uma task na fila.
func Enqueue(dir, featureID, branch, taskID, title, specPath string, rules []string) error {
	if err := ValidateSpecReuseCheck(dir, specPath); err != nil {
		return err
	}

	q, err := Load(dir)
	if err != nil {
		return err
	}
	if featureID != "" {
		q.FeatureID = featureID
	}
	if branch != "" {
		if id, ok := layout.FromBranch(branch); ok {
			if activeID, _, active := layout.ActiveInitiative(dir); active && activeID != id {
				return fmt.Errorf("a branch %q pertence à iniciativa %q, mas a branch atual é da iniciativa %q; troque de branch antes de enfileirar", branch, id, activeID)
			}
		}
		q.Branch = branch
	}
	if q.SpecStatus == "" {
		q.SpecStatus = "reviewed"
	}

	if taskID == "" {
		base := filepath.Base(specPath)
		taskID = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if title == "" {
		title = extractSpecTitle(dir, specPath, taskID)
	}
	if rules == nil {
		rules = []string{}
	}

	dependsOn := extractSpecDependsOn(dir, specPath)

	for i := range q.Items {
		if q.Items[i].TaskID == taskID {
			if q.Items[i].Status == StatusChangesRequested {
				q.Items[i].Status = StatusQueued
				if title != "" {
					q.Items[i].Title = title
				}
				if specPath != "" {
					q.Items[i].SpecPath = specPath
				}
				if len(rules) > 0 {
					q.Items[i].Rules = rules
				}
				if len(dependsOn) > 0 {
					q.Items[i].DependsOn = dependsOn
				}
				return Save(dir, q)
			}
			return fmt.Errorf("task %q já existe na fila com status %q", taskID, q.Items[i].Status)
		}
	}

	item := QueueItem{
		TaskID:     taskID,
		Title:      title,
		SpecPath:   specPath,
		Status:     StatusQueued,
		DependsOn:  dependsOn,
		Rules:      rules,
		SkillsUsed: []string{},
		Attempts:   0,
	}
	q.Items = append(q.Items, item)
	return Save(dir, q)
}

func extractSpecDependsOn(dir, specPath string) []string {
	resolved := specPath
	if !filepath.IsAbs(specPath) {
		resolved = filepath.Join(dir, specPath)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil
	}
	return ParseDependsOn(string(data))
}

var (
	dependsLineRe = regexp.MustCompile(`(?i)^\*\*Depende de:\*\*\s*(.*)$`)
	parenRe       = regexp.MustCompile(`\([^)]*\)`)
	depRefRe      = regexp.MustCompile(`\b(?:(?:feat|fix)-[a-z0-9]+-[0-9]{2}[a-z]?|[0-9]+-[0-9]{2}[a-z]?)\b`)
	cleanDepRe    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	junkDeps      = map[string]bool{"nenhuma": true, "nenhum": true, "e": true, "usa": true}
)

// ParseDependsOn é o único leitor do cabeçalho `**Depende de:**` (DUP-01). Aceita `feat-90-02`,
// `fix-88-01` e `88-02` (devolvidos como escritos); `nenhuma`, `—` e texto entre parênteses
// não geram dependência. Lê só as primeiras linhas, onde fica o cabeçalho.
func ParseDependsOn(content string) []string {
	for i, line := range strings.Split(content, "\n") {
		if i >= 40 {
			break
		}
		m := dependsLineRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		val := parenRe.ReplaceAllString(m[1], "")
		seen := map[string]bool{}
		var deps []string
		for _, ref := range depRefRe.FindAllString(strings.ToLower(val), -1) {
			if !seen[ref] {
				seen[ref] = true
				deps = append(deps, ref)
			}
		}
		return deps
	}
	return nil
}

// cleanDependsOn descarta lixo de filas antigas (`nenhuma`, `(usa`) na leitura, sem editar o arquivo.
func cleanDependsOn(deps []string) []string {
	var out []string
	for _, d := range deps {
		if id := NormalizeTaskID(d); cleanDepRe.MatchString(id) && !junkDeps[strings.ToLower(id)] {
			out = append(out, d)
		}
	}
	return out
}

// EligibleTasks resolve o grafo de dependências da fila (DAG) e devolve as tarefas com status "queued"
// cujas dependências em DependsOn já estejam todas com status "approved".
// Se houver ciclo de dependências, retorna erro descritivo.
// Se max > 0, limita o retorno a até max tarefas elegíveis.
func EligibleTasks(q *Queue, max int) ([]QueueItem, error) {
	if q == nil || len(q.Items) == 0 {
		return []QueueItem{}, nil
	}

	normalizeID := NormalizeTaskID

	// 1. Mapeamento de itens por ID normalizado
	itemMap := make(map[string]QueueItem)
	for _, it := range q.Items {
		itemMap[normalizeID(it.TaskID)] = it
	}

	// 2. Grafo de adjacências e detecção de ciclos
	adj := make(map[string][]string)
	for _, it := range q.Items {
		u := normalizeID(it.TaskID)
		for _, dep := range cleanDependsOn(it.DependsOn) {
			v := normalizeID(dep)
			adj[u] = append(adj[u], v)
		}
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)
	state := make(map[string]int)

	var hasCycle func(u string, path []string) (bool, []string)
	hasCycle = func(u string, path []string) (bool, []string) {
		state[u] = visiting
		path = append(path, u)
		for _, v := range adj[u] {
			if state[v] == visiting {
				return true, append(path, v)
			}
			if state[v] == unvisited {
				if cycle, cyclePath := hasCycle(v, path); cycle {
					return true, cyclePath
				}
			}
		}
		state[u] = visited
		return false, nil
	}

	for _, it := range q.Items {
		id := normalizeID(it.TaskID)
		if state[id] == unvisited {
			if cycle, cyclePath := hasCycle(id, nil); cycle {
				return nil, fmt.Errorf("ciclo de dependência detectado: %s", strings.Join(cyclePath, " -> "))
			}
		}
	}

	// 3. Seleção de tarefas elegíveis: status == queued e todas DependsOn == approved
	var eligible []QueueItem
	for _, it := range q.Items {
		if it.Status != StatusQueued {
			continue
		}
		allApproved := true
		for _, dep := range cleanDependsOn(it.DependsOn) {
			normDep := normalizeID(dep)
			parent, exists := itemMap[normDep]
			if !exists || parent.Status != StatusApproved {
				allApproved = false
				break
			}
		}
		if allApproved {
			eligible = append(eligible, it)
			if max > 0 && len(eligible) >= max {
				break
			}
		}
	}

	return eligible, nil
}

func extractSpecTitle(dir, specPath, fallback string) string {
	resolved := specPath
	if !filepath.IsAbs(specPath) {
		resolved = filepath.Join(dir, specPath)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fallback
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return fallback
}

// Next busca o primeiro item com status "queued" (ou "changes_requested"),
// incrementa Attempts, muda status para "building", salva queue.json e retorna o item.
func Next(dir string) (*QueueItem, error) {
	q, err := Load(dir)
	if err != nil {
		return nil, err
	}

	for i := range q.Items {
		st := q.Items[i].Status
		if st == StatusQueued || st == StatusChangesRequested {
			q.Items[i].Status = StatusBuilding
			q.Items[i].Attempts++
			if err := Save(dir, q); err != nil {
				return nil, err
			}
			itemCopy := q.Items[i]
			return &itemCopy, nil
		}
	}
	return nil, fmt.Errorf("nenhuma task pendente na fila (status %q ou %q)", StatusQueued, StatusChangesRequested)
}

// Complete muda o status de uma task de "building" para "review_pending".
func Complete(dir, taskID string) error {
	q, err := Load(dir)
	if err != nil {
		return err
	}

	for i := range q.Items {
		if q.Items[i].TaskID == taskID || strings.HasPrefix(q.Items[i].TaskID, taskID+"-") {
			if q.Items[i].Status != StatusBuilding {
				return fmt.Errorf("task %q não está em %q (status atual: %q)", taskID, StatusBuilding, q.Items[i].Status)
			}
			q.Items[i].Status = StatusReviewPending
			return Save(dir, q)
		}
	}
	return fmt.Errorf("task %q não encontrada na fila", taskID)
}

// ValidateVerdict valida o schema estrito da Camada B pelo contrato harness.Parecer.
func ValidateVerdict(v Verdict) error {
	return harness.ValidateParecer(v)
}

// Review decodifica estritamente o JSON de parecer da Camada B, valida suas regras e atualiza o status da task na fila.
func Review(dir, taskID string, verdictJSON []byte) (*QueueItem, error) {
	v, err := harness.DecodeParecer(verdictJSON)
	if err != nil {
		return nil, fmt.Errorf("parecer da task %s: %w", taskID, err)
	}

	q, err := Load(dir)
	if err != nil {
		return nil, err
	}

	for i := range q.Items {
		if q.Items[i].TaskID == taskID || strings.HasPrefix(q.Items[i].TaskID, taskID+"-") {
			q.Items[i].GovernanceVerdict = &v
			switch v.Veredito {
			case VerdictConforme:
				q.Items[i].Status = StatusApproved
			case VerdictViolacao:
				q.Items[i].Status = StatusChangesRequested
			case VerdictDuvida:
				q.Items[i].Status = StatusReviewPending
			}
			if err := Save(dir, q); err != nil {
				return nil, err
			}
			itemCopy := q.Items[i]
			return &itemCopy, nil
		}
	}

	return nil, fmt.Errorf("task %q não encontrada na fila", taskID)
}

// FormatExecutivePanel gera o Painel Executivo do Líder (Camada C) com as 4 seções obrigatórias.
func FormatExecutivePanel(q *Queue) string {
	if q == nil {
		q = &Queue{}
	}

	var approved []QueueItem
	var inProgress []QueueItem
	var queued []QueueItem
	var candidateSkills []struct {
		TaskID string
		Skill  CandidateSkill
	}

	for _, item := range q.Items {
		switch item.Status {
		case StatusApproved:
			approved = append(approved, item)
		case StatusBuilding, StatusReviewPending, StatusChangesRequested, StatusBlocked:
			inProgress = append(inProgress, item)
		case StatusQueued:
			queued = append(queued, item)
		}
		if item.GovernanceVerdict != nil && item.GovernanceVerdict.CandidateSkill != nil {
			candidateSkills = append(candidateSkills, struct {
				TaskID string
				Skill  CandidateSkill
			}{
				TaskID: item.TaskID,
				Skill:  *item.GovernanceVerdict.CandidateSkill,
			})
		}
	}

	var b strings.Builder
	headerFeature := q.FeatureID
	if headerFeature == "" {
		headerFeature = "(sem feature)"
	}
	fmt.Fprintf(&b, "# Painel Executivo do Líder — %s\n\n", headerFeature)

	b.WriteString("## 1. Concluído e Aprovado pela Governança (approved)\n")
	if len(approved) == 0 {
		b.WriteString("- Nenhuma task aprovada até o momento.\n\n")
	} else {
		for _, it := range approved {
			resumo := ""
			if it.GovernanceVerdict != nil && it.GovernanceVerdict.Resumo != "" {
				resumo = " — " + it.GovernanceVerdict.Resumo
			}
			fmt.Fprintf(&b, "- [x] `%s`: %s (`%s`)%s\n", it.TaskID, it.Title, it.Status, resumo)
		}
		b.WriteString("\n")
	}

	b.WriteString("## 2. Em Andamento / Em Revisão (building, review_pending, changes_requested)\n")
	if len(inProgress) == 0 {
		b.WriteString("- Nenhuma task em andamento ou revisão.\n\n")
	} else {
		for _, it := range inProgress {
			detail := ""
			if it.GovernanceVerdict != nil && it.GovernanceVerdict.Resumo != "" {
				detail = " — " + it.GovernanceVerdict.Resumo
			}
			if it.Status == StatusBlocked && it.BlockedReason != "" {
				detail = " — bloqueada: " + it.BlockedReason
			}
			fmt.Fprintf(&b, "- [~] `%s`: %s (`%s`, tentativas: %d)%s\n", it.TaskID, it.Title, it.Status, it.Attempts, detail)
		}
		b.WriteString("\n")
	}

	b.WriteString("## 3. O que ainda não foi desenvolvido (queued)\n")
	if len(queued) == 0 {
		b.WriteString("- Nenhuma task aguardando na fila.\n\n")
	} else {
		for _, it := range queued {
			fmt.Fprintf(&b, "- [ ] `%s`: %s (`%s`)\n", it.TaskID, it.Title, it.Status)
		}
		b.WriteString("\n")
	}

	b.WriteString("## 4. Auto-Skills Propostas (candidate_skill)\n")
	if len(candidateSkills) == 0 {
		b.WriteString("- Nenhuma auto-skill proposta.\n")
	} else {
		for _, cs := range candidateSkills {
			fmt.Fprintf(&b, "- `%s` (de `%s` → `%s`): %s\n", cs.Skill.Name, cs.TaskID, cs.Skill.ProposedPath, cs.Skill.Reason)
		}
	}

	return b.String()
}

// NormalizeTaskID é a regra única de ID de task: sem espaços nas pontas e sem o prefixo "feat-".
func NormalizeTaskID(id string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(id), "feat-"))
}

// LoadForInitiative lê a fila de uma iniciativa específica pelo ID, sem depender da branch ativa.
// Retorna fila vazia (não erro) quando queue.json não existe.
func LoadForInitiative(dir, initiative string) (*Queue, error) {
	q, err := LoadInitiative(dir, initiative)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return &Queue{Items: []QueueItem{}}, nil
	}
	if q.Items == nil {
		q.Items = []QueueItem{}
	}
	return q, nil
}
