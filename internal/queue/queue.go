package queue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	StatusQueued           = "queued"
	StatusBuilding         = "building"
	StatusReviewPending    = "review_pending"
	StatusApproved         = "approved"
	StatusChangesRequested = "changes_requested"

	VerdictConforme = "conforme"
	VerdictViolacao = "violacao"
	VerdictDuvida   = "duvida"
)

// Queue representa a fila persistida em sdd/.queue/queue.json.
type Queue struct {
	FeatureID  string      `json:"feature_id"`
	Branch     string      `json:"branch"`
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
	Rules             []string `json:"rules"`
	SkillsUsed        []string `json:"skills_used"`
	Attempts          int      `json:"attempts"`
	GovernanceVerdict *Verdict `json:"governance_verdict,omitempty"`
}

// Verdict representa o parecer semântico da Camada B (Analista de Governança).
type Verdict struct {
	Veredito       string          `json:"veredito"`
	Achados        []Finding       `json:"achados"`
	CandidateSkill *CandidateSkill `json:"candidate_skill,omitempty"`
	Resumo         string          `json:"resumo"`
}

// Finding representa um achado individual dentro do parecer da Camada B.
type Finding struct {
	Verificacao   string `json:"verificacao"`
	Regra         string `json:"regra"`
	Classificacao string `json:"classificacao"`
	Severidade    string `json:"severidade"`
	Arquivo       string `json:"arquivo"`
	Linha         int    `json:"linha"`
	Evidencia     string `json:"evidencia"`
	Explicacao    string `json:"explicacao"`
	Correcao      string `json:"correcao"`
}

// CandidateSkill representa uma proposta de auto-skill identificada durante a execução/revisão.
type CandidateSkill struct {
	Name         string `json:"name"`
	Reason       string `json:"reason"`
	ProposedPath string `json:"proposed_path"`
}

// QueueFilePath retorna o caminho de persistência da fila no diretório alvo.
func QueueFilePath(dir string) string {
	return filepath.Join(dir, "sdd", ".queue", "queue.json")
}

// Load lê a fila a partir de <dir>/sdd/.queue/queue.json.
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

// Save persiste a fila em <dir>/sdd/.queue/queue.json atualizando UpdatedAt.
func Save(dir string, q *Queue) error {
	path := QueueFilePath(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("falha ao criar diretório da fila: %w", err)
	}
	if q.SpecStatus == "" {
		q.SpecStatus = "reviewed"
	}
	q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return fmt.Errorf("falha ao serializar fila: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
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
		Rules:      rules,
		SkillsUsed: []string{},
		Attempts:   0,
	}
	q.Items = append(q.Items, item)
	return Save(dir, q)
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

// ValidateVerdict valida o schema estrito da Camada B (Analista de Governança).
func ValidateVerdict(v Verdict) error {
	switch v.Veredito {
	case VerdictConforme, VerdictViolacao, VerdictDuvida:
	default:
		return fmt.Errorf("veredito inválido %q: deve ser %q, %q ou %q", v.Veredito, VerdictConforme, VerdictViolacao, VerdictDuvida)
	}

	if strings.TrimSpace(v.Resumo) == "" {
		return fmt.Errorf("campo 'resumo' do veredito não pode ser vazio")
	}

	for idx, f := range v.Achados {
		switch f.Classificacao {
		case VerdictViolacao, VerdictDuvida:
		default:
			return fmt.Errorf("achado[%d]: classificacao inválida %q (esperado %q ou %q)", idx, f.Classificacao, VerdictViolacao, VerdictDuvida)
		}

		if f.Classificacao == VerdictViolacao {
			if strings.TrimSpace(f.Regra) == "" {
				return fmt.Errorf("achado[%d]: violação sem 'regra' preenchida", idx)
			}
			if strings.TrimSpace(f.Arquivo) == "" {
				return fmt.Errorf("achado[%d]: violação sem 'arquivo' preenchido", idx)
			}
			if f.Linha <= 0 {
				return fmt.Errorf("achado[%d]: violação exige 'linha' > 0 (recebido %d)", idx, f.Linha)
			}
			if strings.TrimSpace(f.Evidencia) == "" {
				return fmt.Errorf("achado[%d]: violação sem 'evidencia' explícita (violação sem evidência é inválida)", idx)
			}
		}
	}

	if v.CandidateSkill != nil {
		if strings.TrimSpace(v.CandidateSkill.Name) == "" || strings.TrimSpace(v.CandidateSkill.Reason) == "" || strings.TrimSpace(v.CandidateSkill.ProposedPath) == "" {
			return fmt.Errorf("candidate_skill incompleta: 'name', 'reason' e 'proposed_path' são obrigatórios quando presente")
		}
	}

	return nil
}

// Review decodifica estritamente o JSON de parecer da Camada B, valida suas regras e atualiza o status da task na fila.
func Review(dir, taskID string, verdictJSON []byte) (*QueueItem, error) {
	dec := json.NewDecoder(bytes.NewReader(verdictJSON))
	dec.DisallowUnknownFields()

	var v Verdict
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("parecer JSON inválido: %w", err)
	}
	if v.Achados == nil {
		v.Achados = []Finding{}
	}

	if err := ValidateVerdict(v); err != nil {
		return nil, err
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
		case StatusBuilding, StatusReviewPending, StatusChangesRequested:
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
