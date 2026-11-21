package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/queue"
)

// Arquivos do /toca-ai.
const (
	// RequirementsFile é o que o Líder grava na iniciativa com tudo que as etapas vão usar.
	RequirementsFile = "requisitos.json"
	// HumanPendingFile é a lista única de pendências humanas da iniciativa.
	HumanPendingFile = "pendencias-humanas.md"
	consentsFileName = "consents.json"
	goalJobPrefix    = "goal-"
	maxReqBytes      = 64 * 1024
	maxConsentText   = 512
)

var goalTaskRe = regexp.MustCompile(`^feat-[a-z0-9]+-[0-9]{2}[a-z]?$`)

var consentIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// Requirements é o esquema de requisitos.json (SEG-01): comandos em argv explícito, nunca shell.
type Requirements struct {
	Comandos       [][]string       `json:"comandos"`
	Ferramentas    []string         `json:"ferramentas"`
	Arquivos       []string         `json:"arquivos"`
	Consentimentos []ConsentRequest `json:"consentimentos"`
}

// ConsentRequest é um consentimento que só a pessoa pode dar.
type ConsentRequest struct {
	ID    string `json:"id"`
	Texto string `json:"texto"`
}

// ConsentRecord é a resposta humana gravada em .forge/runtime/consents.json (sem segredos, SEG-04).
type ConsentRecord struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	At   string `json:"at"`
}

func validText(s string, max int) bool {
	return s != "" && len(s) <= max && !strings.ContainsAny(s, "\x00\r\n\x1b")
}

// LoadRequirements lê e valida requisitos.json. Arquivo ausente devolve nil sem erro.
func LoadRequirements(dir, initiative string) (*Requirements, error) {
	base, err := layout.Initiative(dir, initiative)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(base, RequirementsFile)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxReqBytes {
		return nil, fmt.Errorf("%s: precisa ser arquivo comum de até %d KB", RequirementsFile, maxReqBytes/1024)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Requirements
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("%s inválido: %w", RequirementsFile, err)
	}
	for _, argv := range r.Comandos {
		if len(argv) == 0 || len(argv) > 64 {
			return nil, fmt.Errorf("%s: cada comando é um argv de 1 a 64 itens", RequirementsFile)
		}
		for _, a := range argv {
			if !validText(a, 512) {
				return nil, fmt.Errorf("%s: argumento de comando inválido %q", RequirementsFile, a)
			}
		}
	}
	for _, s := range append(append([]string{}, r.Ferramentas...), r.Arquivos...) {
		if !validText(s, 512) {
			return nil, fmt.Errorf("%s: ferramenta ou arquivo inválido %q", RequirementsFile, s)
		}
	}
	for _, c := range r.Consentimentos {
		if !consentIDRe.MatchString(c.ID) || !validText(c.Texto, maxConsentText) {
			return nil, fmt.Errorf("%s: consentimento %q inválido (id em minúsculas e hífens, texto de uma linha)", RequirementsFile, c.ID)
		}
	}
	return &r, nil
}

func consentsPath(dir string) string {
	return filepath.Join(layout.ForgeRuntime(dir), consentsFileName)
}

// LoadConsents devolve os consentimentos gravados; arquivo ausente é lista vazia.
func LoadConsents(dir string) ([]ConsentRecord, error) {
	data, err := os.ReadFile(consentsPath(dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var recs []ConsentRecord
	if err := json.Unmarshal(data, &recs); err != nil {
		return nil, fmt.Errorf("%s inválido: %w", consentsFileName, err)
	}
	return recs, nil
}

// RecordConsent grava (atomicamente) a resposta humana. Só o comando interativo `forge pipeline
// consent` chama; nenhum job grava consentimento (SEG-06).
func RecordConsent(dir, id, text string, now time.Time) error {
	if !consentIDRe.MatchString(id) || !validText(text, maxConsentText) {
		return fmt.Errorf("consentimento %q inválido", id)
	}
	recs, err := LoadConsents(dir)
	if err != nil {
		return err
	}
	out := recs[:0:0]
	for _, r := range recs {
		if r.ID != id {
			out = append(out, r)
		}
	}
	out = append(out, ConsentRecord{ID: id, Text: text, At: now.UTC().Format(time.RFC3339)})
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return queue.WriteFileAtomic(consentsPath(dir), append(data, '\n'), 0o600)
}

// ConsentGiven informa se há resposta humana gravada para o id.
func ConsentGiven(dir, id string) bool {
	recs, _ := LoadConsents(dir)
	for _, r := range recs {
		if r.ID == id {
			return true
		}
	}
	return false
}

// PreflightGoal confere tudo que as etapas precisam da pessoa e devolve TODAS as pendências de uma
// vez, ordenadas. Não grava nada. Sem requisitos.json não há pendência.
func PreflightGoal(dir, initiative string, cfg config.ForgeConfig, env PreflightEnv) ([]string, error) {
	req, err := LoadRequirements(dir, initiative)
	if err != nil || req == nil {
		return nil, err
	}
	var pend []string
	for _, tool := range req.Ferramentas {
		if _, err := env.LookPath(tool); err != nil {
			pend = append(pend, fmt.Sprintf("instalar a ferramenta %q (não está no PATH)", tool))
		}
	}
	for _, f := range req.Arquivos {
		resolved, err := gate.ConfineToDir(dir, f)
		if err != nil || !env.Exists(resolved) {
			pend = append(pend, fmt.Sprintf("fornecer o arquivo %q (dentro do projeto)", f))
		}
	}
	if cfg.Permissions != config.PermissionFull {
		allowed := map[string]bool{}
		for _, a := range cfg.AllowCommands {
			allowed[a] = true
		}
		missing := map[string]bool{}
		for _, argv := range req.Comandos {
			if name := filepath.Base(argv[0]); !allowed[name] {
				missing[name] = true
			}
		}
		for name := range missing {
			pend = append(pend, fmt.Sprintf("liberar o comando %q em allow_commands (forge.json) ou usar permissions full", name))
		}
	}
	for _, c := range req.Consentimentos {
		if !ConsentGiven(dir, c.ID) {
			pend = append(pend, fmt.Sprintf("consentimento %q: %s (responda com: forge pipeline consent %s --iniciativa %s)", c.ID, c.Texto, c.ID, initiative))
		}
	}
	sort.Strings(pend)
	return pend, nil
}

// GoalOptions controla o enfileiramento das etapas.
type GoalOptions struct {
	DryRun bool
	Now    time.Time
	Env    PreflightEnv
	Git    queue.GitFunc // injetável nos testes; padrão: git real
	// SkipSpecGate é usado pelo próprio Go depois do parecer aprovado da spec.
	SkipSpecGate bool
}

// GoalResult descreve o que o motor fez.
type GoalResult struct {
	Initiative string
	Steps      []string // etapas na ordem do DAG
	Pending    []string // preflight: pendências humanas; nada foi enfileirado
	Awaiting   string   // portão da spec: o que falta antes das etapas
	SpecJobID  string
	Created    []string // jobs novos (ids)
	Existing   []string // jobs que já estavam na fila
	BaseBranch string   // base do PR das etapas
	BaseSource string   // de onde a base veio
}

// goalStepJobID é determinístico (sem horário): reenviar não duplica.
func goalStepJobID(initiative, step string) string {
	id := goalJobPrefix + step + "-" + initiative
	if len(id) > 80 {
		id = id[:80]
	}
	return strings.TrimRight(id, "-.")
}

func goalSpecJobID(initiative string, attempt int) string {
	suffix := fmt.Sprintf("-a%d-spec", attempt)
	id := goalJobPrefix + initiative
	if len(id)+len(suffix) > 80 {
		id = id[:80-len(suffix)]
	}
	return strings.TrimRight(id, "-.") + suffix
}

// isGoalJob diz se o job é do /toca-ai (o parecer aprovado libera as etapas, não um build único).
func isGoalJob(j bus.Job) bool { return strings.HasPrefix(j.ID, goalJobPrefix) }

// goalSteps devolve as etapas na ordem do DAG; ciclo ou dependência inexistente é erro com a lista.
func goalSteps(dir, initiative string) ([]Task, error) {
	feats, err := layout.InitiativeFeatures(dir, initiative)
	if err != nil {
		return nil, err
	}
	base := filepath.Dir(feats)
	if _, err := os.Stat(filepath.Join(base, "plan.md")); err != nil {
		return nil, fmt.Errorf("falta plan.md em %s: o Líder gera o plano antes do motor", filepath.ToSlash(base))
	}
	batches, err := PlanWaves(feats, 1)
	if err != nil {
		return nil, err
	}
	var steps []Task
	for _, b := range batches {
		steps = append(steps, b...)
	}
	return steps, nil
}

// gate1Re casa o marcador do Gate 1 como linha própria do status.md, com ou sem negrito.
var gate1Re = regexp.MustCompile(`(?im)^\s*(?:\*\*)?Gate 1:(?:\*\*)?\s*aprovado\b`)

// maxStatusBytes limita a leitura do status.md (dado não confiável, SEG-02).
const maxStatusBytes = 64 << 10

// gate1Marked diz se o status.md da iniciativa traz a linha `Gate 1: aprovado`. Link simbólico,
// arquivo ausente ou id inválido valem como "não aprovado".
func gate1Marked(dir, initiative string) bool {
	path, err := layout.InitiativeStatus(dir, initiative)
	if err != nil {
		return false
	}
	if fi, err := os.Lstat(path); err != nil || !fi.Mode().IsRegular() {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxStatusBytes))
	return err == nil && gate1Re.Match(data)
}

// goalSpecApproved diz se a spec da iniciativa já passou no portão. No modo avançado, vale o parecer
// aprovado do Reviewer; no normal, a aprovação humana (Gate 1): o marcador `Gate 1: aprovado` no
// status.md (que vem da principal) ou, em iniciativa antiga, a queue.json que o Gate 1 criava.
func goalSpecApproved(dir, initiative string, cfg config.ForgeConfig) bool {
	if !cfg.IsAdvanced() {
		base, err := layout.Initiative(dir, initiative)
		if err != nil {
			return false
		}
		if gate1Marked(dir, initiative) {
			return true
		}
		_, err = os.Stat(filepath.Join(base, layout.QueueFileName)) // iniciativa antiga: criada pelo Gate 1 humano
		return err == nil
	}
	entries, err := bus.List(dir, layout.RoleReviewer, layout.StateDone)
	if err != nil {
		return false
	}
	for _, e := range entries {
		j := e.Job
		if j.Type == bus.TypeSpecReview && isGoalJob(j) && j.Initiative == initiative && len(j.Events) > 0 &&
			strings.HasPrefix(j.Events[len(j.Events)-1].Note, "spec: aprovada") {
			return true
		}
	}
	return false
}

// EnqueueGoalSteps é o motor do /toca-ai: confere o pré-voo e o portão da spec e enfileira um job
// por etapa, em ordem topológica, de forma idempotente (reenviar não duplica).
func EnqueueGoalSteps(dir, initiative string, opts GoalOptions) (*GoalResult, error) {
	cfg, found, err := config.LoadForgeConfig(dir)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotPipelineMode
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	if opts.Env.LookPath == nil {
		opts.Env = DefaultPreflightEnv()
	}
	if base, berr := layout.Initiative(dir, initiative); berr == nil {
		// A spec incompleta sai antes de montar o DAG e do build, com tudo o que falta de uma vez.
		if problems := taskSpecProblems(dir, base, ""); len(problems) > 0 {
			return &GoalResult{Initiative: initiative, Pending: problems}, nil
		}
	}
	steps, err := goalSteps(dir, initiative)
	if err != nil {
		return nil, err
	}
	res := &GoalResult{Initiative: initiative}
	base, baseSource := queue.ResolveBase(dir, initiative, "", cfg.BaseBranch, opts.Git)
	res.BaseBranch, res.BaseSource = base, baseSource
	for _, s := range steps {
		res.Steps = append(res.Steps, s.ID)
	}

	if res.Pending, err = PreflightGoal(dir, initiative, cfg, opts.Env); err != nil {
		return nil, err
	}
	sort.Strings(res.Pending)
	if len(res.Pending) > 0 {
		return res, nil
	}

	if !opts.SkipSpecGate && !goalSpecApproved(dir, initiative, cfg) {
		if !cfg.IsAdvanced() {
			res.Awaiting = "a spec ainda não foi aprovada por uma pessoa (Gate 1): aprove e rode de novo; nenhuma etapa foi enfileirada"
			return res, nil
		}
		job := goalJob(cfg, base, initiative, "sdd/discovery/"+initiative+"/plan.md", res.Steps, goalSpecJobID(initiative, 1), now)
		job.Type = bus.TypeSpecReview
		job.Attempt = nextSpecAttempt(dir, initiative, job.MaxReworkAttempts)
		job.ID = goalSpecJobID(initiative, job.Attempt)
		res.SpecJobID = job.ID
		res.Awaiting = "aguardando o parecer da spec pelo Reviewer; as etapas entram na fila quando for aprovada"
		if !opts.DryRun {
			if _, _, err := bus.Enqueue(dir, layout.RoleReviewer, job, now); err != nil {
				return nil, err
			}
		}
		return res, nil
	}

	if !opts.DryRun {
		// O motor do /toca-ai é quem pede o archive automático ao aprovar a última etapa.
		if err := queue.SetOrigem(dir, initiative, queue.OrigemTocaAi); err != nil {
			return nil, err
		}
	}
	for i, s := range steps {
		// CreatedAt crescente: a fila é FIFO, então a ordem do DAG vira a ordem de consumo.
		at := now.UTC().Add(time.Duration(i) * time.Millisecond)
		spec, _ := filepath.Rel(dir, s.File)
		job := goalJob(cfg, base, initiative, filepath.ToSlash(spec), []string{s.ID}, goalStepJobID(initiative, s.ID), at)
		job.Type = bus.TypeFeatureBuild
		job.PR.Branch = layout.InitiativeBranch(initiative) + "-" + s.ID
		job.Title = s.ID + " — " + initiative
		if opts.DryRun {
			res.Created = append(res.Created, job.ID)
			continue
		}
		job.CreatedAt = at
		_, created, err := bus.Enqueue(dir, layout.RoleBuilder, job, at)
		if err != nil {
			return nil, err
		}
		if created {
			res.Created = append(res.Created, job.ID)
		} else {
			res.Existing = append(res.Existing, job.ID)
		}
	}
	return res, nil
}

func goalJob(cfg config.ForgeConfig, base, initiative, specPath string, taskIDs []string, id string, at time.Time) bus.Job {
	return bus.Job{
		SchemaVersion:     bus.SchemaVersion,
		ID:                id,
		Priority:          bus.PriorityNormal,
		Feature:           initiative,
		Title:             "Plano de " + initiative,
		Initiative:        initiative,
		TaskIDs:           taskIDs,
		SpecPath:          filepath.ToSlash(specPath),
		PR:                bus.PR{Branch: layout.InitiativeBranch(initiative), BaseBranch: base},
		Attempt:           1,
		MaxReworkAttempts: cfg.Reviewer.MaxReworkAttempts,
		ReworkFeedback:    []bus.Feedback{},
		CreatedAt:         at,
	}
}

// HumanPending é um imprevisto que só a pessoa resolve, ligado à etapa que ele bloqueia.
type HumanPending struct {
	TaskID string
	Item   string
}

// ReportHumanPending bloqueia as etapas com o item exato, mantém as independentes na fila,
// grava uma lista única e sem repetição em pendencias-humanas.md e avisa pelo inbox.
func ReportHumanPending(dir, initiative string, items []HumanPending, deps notify.Deps, now time.Time) error {
	base, err := layout.Initiative(dir, initiative)
	if err != nil {
		return err
	}
	ncfg, err := config.LoadNotificationsConfig(dir)
	if err != nil {
		return err
	}
	for _, it := range items {
		if !validText(it.Item, 512) || !goalTaskRe.MatchString(it.TaskID) {
			return fmt.Errorf("pendência inválida (etapa %q)", it.TaskID)
		}
		if _, err := queue.BlockTasks(dir, initiative, []string{it.TaskID}, it.Item); err != nil {
			return err
		}
	}
	path := filepath.Join(base, HumanPendingFile)
	seen := map[string]bool{}
	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(l, "- ") && !seen[l] {
				seen[l] = true
				lines = append(lines, l)
			}
		}
	}
	for _, it := range items {
		l := fmt.Sprintf("- [%s] %s", it.TaskID, it.Item)
		if !seen[l] {
			seen[l] = true
			lines = append(lines, l)
		}
	}
	body := "# Pendências humanas — " + initiative + "\n\n" + strings.Join(lines, "\n") + "\n"
	if err := queue.WriteFileAtomic(path, []byte(body), 0o644); err != nil {
		return err
	}
	n := notify.Notification{
		ID: notificationID(initiative, "pendencias"), Type: notify.TypeInfo,
		Title:   "Preciso de você: " + initiative,
		Message: fmt.Sprintf("%d pendência(s) em %s; as etapas independentes continuam.", len(lines), HumanPendingFile),
		Feature: initiative,
	}
	_, err = notify.Dispatch(dir, ncfg, n, deps)
	return err
}
