// Package pipeline compõe as peças do pipeline multi-sessão: o gate pré-build e o despacho
// (Discovery), o worker e o Builder paralelo, e a máquina de estados do Reviewer.
package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
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
	"github.com/forge-sdd/cli/internal/queue"
)

// ErrNotPipelineMode indica que não há .forge/config/forge.json: vale o fluxo da 4.0.0-beta.
var ErrNotPipelineMode = errors.New("modo pipeline não ativo: rode forge-sdd pipeline init")

// CardEvidenceFile é o nome padrão da evidência do card, gravada pelo agente na pasta da iniciativa.
const CardEvidenceFile = "card-evidence.json"

var (
	featureSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,80}$`)
	adrTokenRe    = regexp.MustCompile(`\badr-[0-9]{4}[a-z0-9-]*`)
	taskIDRe      = regexp.MustCompile(`^feat-([a-z0-9]+)-([0-9]{2}[a-z]?)`)
	sectionRe     = regexp.MustCompile(`(?m)^##\s+(.+?)\s*$`)
	requiredSecs  = []string{"modelo do módulo", "o que já existe", "dependências novas", "rules e ids da constituição aplicáveis", "critérios de aceitação executáveis"}
)

// CardEvidence é o que o agente registra depois de consultar o card pelo MCP (D-2).
type CardEvidence struct {
	Provider  string `json:"provider"`
	ID        string `json:"id"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	Feature   string `json:"feature"`
	CheckedAt string `json:"checked_at"`
}

// DispatchOptions controla o despacho.
type DispatchOptions struct {
	EvidencePath string // relativo ao projeto; padrão: <iniciativa>/card-evidence.json
	DryRun       bool
	Now          time.Time
}

// DispatchResult descreve o resultado do gate e, se passou, do despacho.
type DispatchResult struct {
	Initiative      string
	Pending         []string // pendências que bloqueiam; vazio quando o gate passou
	Job             *bus.Job
	JobPath         string
	CardInstruction string
	Created         bool
}

// Blocked informa se o gate reprovou.
func (r *DispatchResult) Blocked() bool { return len(r.Pending) > 0 }

// Dispatch roda o gate pré-build e, se passar, enfileira o job em builder/pending. As pendências
// são todas listadas de uma vez; nada é gravado quando há alguma (cenários 1 e 2). Falha de
// sistema vira erro; reprovação do gate é resultado, não erro.
func Dispatch(dir, feature string, opts DispatchOptions) (*DispatchResult, error) {
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
	feature = strings.TrimSpace(feature)
	if !featureSlugRe.MatchString(feature) {
		return nil, fmt.Errorf("feature %q inválida: use letras minúsculas, números e hífens (SEG-01)", feature)
	}

	res := &DispatchResult{}
	initiative, pend := resolveInitiative(dir, feature)
	if len(pend) > 0 {
		res.Pending = pend
		return res, nil
	}
	res.Initiative = initiative
	base := filepath.Join(layout.Discovery(dir), initiative)
	rel := func(p string) string {
		r, _ := filepath.Rel(dir, p)
		return filepath.ToSlash(r)
	}

	// 1. Artefatos da spec.
	for _, f := range []string{"discovery.md", "criteria.md", "plan.md"} {
		if _, err := os.Stat(filepath.Join(base, f)); err != nil {
			res.Pending = append(res.Pending, fmt.Sprintf("falta %s em %s", f, rel(base)))
		}
	}
	criteria := rel(filepath.Join(base, "criteria.md"))
	if _, ok := gate.ValidateFLX01(dir, criteria); !ok {
		res.Pending = append(res.Pending, fmt.Sprintf("%s não passa no gate FLX-01: precisa da seção \"## O que já existe\" com a tabela preenchida (P-02, P-03)", criteria))
	}

	// 2. Quebra em tasks, cada uma com as cinco seções e IDs que existem na Constitution.
	tasks, taskIDs := listTasks(base)
	if len(tasks) == 0 {
		res.Pending = append(res.Pending, fmt.Sprintf("nenhuma task em %s/features (rode /split-features)", rel(base)))
	}
	rulesSet := map[string]bool{}
	adrCited := map[string]bool{}
	adrMissing := map[string]bool{}
	for _, t := range tasks {
		trel := rel(t)
		if err := queue.ValidateSpecReuseCheck(dir, trel); err != nil {
			res.Pending = append(res.Pending, fmt.Sprintf("%s: %v", trel, firstLine(err.Error())))
		}
		data, rerr := os.ReadFile(t)
		if rerr != nil {
			res.Pending = append(res.Pending, fmt.Sprintf("%s: não foi possível ler", trel))
			continue
		}
		if missing := missingSections(string(data)); len(missing) > 0 {
			res.Pending = append(res.Pending, fmt.Sprintf("%s: faltam as seções %s", trel, strings.Join(missing, ", ")))
		}
		pack, cerr := gate.BuildContext(dir, trel, nil, "build")
		if cerr != nil {
			res.Pending = append(res.Pending, firstLine(cerr.Error()))
			continue
		}
		if len(pack.IDs) == 0 {
			res.Pending = append(res.Pending, fmt.Sprintf("%s: nenhum ID da Constituição citado em \"Rules e IDs\"", trel))
		}
		for _, r := range pack.Rules {
			rulesSet[r.Path] = true
		}
		for _, tok := range adrTokensIn(sectionBody(string(data), "rules e ids da constituição aplicáveis")) {
			adrCited[tok] = true
		}
	}
	for _, f := range []string{"criteria.md"} {
		if data, err := os.ReadFile(filepath.Join(base, f)); err == nil {
			for _, tok := range adrTokensIn(sectionBody(string(data), "adrs aplicáveis")) {
				adrCited[tok] = true
			}
		}
	}
	adrFiles := map[string]bool{}
	for tok := range adrCited {
		if p, ok := findADR(dir, cfg.ADRDirs, tok); ok {
			adrFiles[p] = true
		} else {
			adrMissing[tok] = true
		}
	}
	for tok := range adrMissing {
		res.Pending = append(res.Pending, fmt.Sprintf("ADR %q citada nas specs não existe em %s", tok, strings.Join(cfg.ADRDirs, ", ")))
	}
	if len(rulesSet) == 0 && len(tasks) > 0 {
		res.Pending = append(res.Pending, "nenhuma rule de .agents/rules/ cobre os IDs citados nas tasks (mapeie as rules aplicáveis)")
	}
	// ADRs citadas em qualquer ponto de criteria.md e que existem entram no job como referência.
	if data, err := os.ReadFile(filepath.Join(base, "criteria.md")); err == nil {
		for _, tok := range adrTokensIn(string(data)) {
			if p, ok := findADR(dir, cfg.ADRDirs, tok); ok {
				adrFiles[p] = true
			}
		}
	}

	// 3. Evidência do card (D-2).
	evPath := opts.EvidencePath
	if evPath == "" {
		evPath = rel(filepath.Join(base, CardEvidenceFile))
	}
	ev, pend := loadEvidence(dir, evPath, feature, initiative, cfg, now)
	res.Pending = append(res.Pending, pend...)

	sort.Strings(res.Pending)
	if res.Blocked() {
		return res, nil
	}

	// 4. Job.
	job := bus.Job{
		SchemaVersion:     bus.SchemaVersion,
		ID:                jobID(feature, now),
		Type:              bus.TypeFeatureBuild,
		Priority:          bus.PriorityNormal,
		Feature:           feature,
		Title:             initiativeTitle(base, feature),
		Initiative:        initiative,
		TaskIDs:           taskIDs,
		SpecPath:          criteria,
		ADRRefs:           sortedKeys(adrFiles),
		RulesRefs:         sortedKeys(rulesSet),
		Card:              bus.Card{Provider: ev.Provider, ID: ev.ID, URL: ev.URL},
		PR:                bus.PR{Branch: "feat/" + initiative, BaseBranch: cfg.BaseBranch},
		Attempt:           1,
		MaxReworkAttempts: cfg.Reviewer.MaxReworkAttempts,
		ReworkFeedback:    []bus.Feedback{},
	}
	res.Job = &job
	res.CardInstruction = fmt.Sprintf("Atualize o card %s (%s) para %q via MCP.", ev.ID, ev.URL, cfg.PMMCP.StatusOnImplement)
	if opts.DryRun {
		return res, nil
	}
	path, created, err := bus.Enqueue(dir, layout.RoleBuilder, job, now)
	if err != nil {
		return nil, err
	}
	res.JobPath, res.Created = path, created
	return res, nil
}

// resolveInitiative acha a pasta da iniciativa pelo nome da feature (igual ou contido).
func resolveInitiative(dir, feature string) (string, []string) {
	entries, err := os.ReadDir(layout.Discovery(dir))
	if err != nil {
		return "", []string{"sdd/discovery não encontrado"}
	}
	var exact, contains []string
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() || !layout.ValidInitiative(n) {
			continue
		}
		slug := n
		if i := strings.Index(n, "-"); i >= 0 {
			slug = n[i+1:] // sem o <id>-
		}
		switch {
		case slug == feature || n == feature:
			exact = append(exact, n)
		case strings.Contains(n, feature):
			contains = append(contains, n)
		}
	}
	switch {
	case len(exact) == 1:
		return exact[0], nil
	case len(exact) > 1:
		return "", []string{fmt.Sprintf("feature %q é ambígua: %s", feature, strings.Join(exact, ", "))}
	case len(contains) == 1:
		return contains[0], nil
	case len(contains) > 1:
		return "", []string{fmt.Sprintf("feature %q é ambígua: %s (use o nome completo)", feature, strings.Join(contains, ", "))}
	}
	return "", []string{fmt.Sprintf("nenhuma iniciativa em sdd/discovery corresponde a %q (a iniciativa precisa estar na árvore de trabalho atual)", feature)}
}

func listTasks(base string) (paths []string, ids []string) {
	entries, err := os.ReadDir(filepath.Join(base, "features"))
	if err != nil {
		return nil, nil
	}
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, "feat-") || !strings.HasSuffix(n, ".md") {
			continue
		}
		paths = append(paths, filepath.Join(base, "features", n))
		if m := taskIDRe.FindStringSubmatch(n); m != nil {
			ids = append(ids, m[1]+"-"+m[2])
		}
	}
	sort.Strings(paths)
	sort.Strings(ids)
	return paths, ids
}

func missingSections(content string) []string {
	have := map[string]bool{}
	for _, m := range sectionRe.FindAllStringSubmatch(content, -1) {
		have[strings.ToLower(strings.TrimSpace(strings.Trim(m[1], "`")))] = true
		have[strings.ToLower(strings.TrimSpace(m[1]))] = true
	}
	var missing []string
	for _, want := range requiredSecs {
		found := false
		for h := range have {
			if strings.HasPrefix(h, want) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, want)
		}
	}
	return missing
}

// sectionBody devolve o texto da seção cujo título começa com name (minúsculas), até a próxima "## ".
func sectionBody(content, name string) string {
	locs := sectionRe.FindAllStringSubmatchIndex(content, -1)
	for i, loc := range locs {
		title := strings.ToLower(strings.TrimSpace(content[loc[2]:loc[3]]))
		if !strings.HasPrefix(title, name) {
			continue
		}
		end := len(content)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		return content[loc[1]:end]
	}
	return ""
}

func adrTokensIn(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range adrTokenRe.FindAllString(text, -1) {
		m = strings.TrimRight(m, "-.")
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

// findADR procura em adr_dirs um arquivo cujo nome começa com o token (adr-0003 casa adr-0003-x.md).
func findADR(dir string, adrDirs []string, token string) (string, bool) {
	for _, d := range adrDirs {
		entries, err := os.ReadDir(filepath.Join(dir, filepath.FromSlash(d)))
		if err != nil {
			continue
		}
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".md") {
				continue
			}
			base := strings.TrimSuffix(n, ".md")
			if base == token || strings.HasPrefix(base, token+"-") || strings.HasPrefix(token, base) {
				return filepath.ToSlash(filepath.Join(d, n)), true
			}
		}
	}
	return "", false
}

// loadEvidence valida a evidência do card (esquema estrito, frescura, provider e feature).
func loadEvidence(dir, evPath, feature, initiative string, cfg config.ForgeConfig, now time.Time) (CardEvidence, []string) {
	var ev CardEvidence
	if !filepath.IsAbs(evPath) {
		if _, err := os.Lstat(filepath.Join(dir, evPath)); errors.Is(err, os.ErrNotExist) {
			return ev, []string{fmt.Sprintf("falta a evidência do card (%s): consulte o card pelo MCP e grave card-evidence.json (card inexistente bloqueia o despacho)", filepath.ToSlash(evPath))}
		}
	}
	resolved, err := gate.ConfineToDir(dir, evPath)
	if err != nil {
		return ev, []string{fmt.Sprintf("evidência do card %q: %s", evPath, firstLine(err.Error()))}
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return ev, []string{fmt.Sprintf("falta a evidência do card (%s): consulte o card pelo MCP e grave card-evidence.json (card inexistente bloqueia o despacho)", evPath)}
	}
	if info.Size() > 16*1024 {
		return ev, []string{fmt.Sprintf("evidência do card %s acima de 16 KB", evPath)}
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return ev, []string{fmt.Sprintf("não foi possível ler a evidência do card %s", evPath)}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ev); err != nil {
		return ev, []string{fmt.Sprintf("evidência do card %s inválida: %v", evPath, err)}
	}
	var pend []string
	add := func(format string, a ...any) { pend = append(pend, fmt.Sprintf(format, a...)) }

	for field, v := range map[string]string{"provider": ev.Provider, "id": ev.ID, "url": ev.URL, "status": ev.Status, "feature": ev.Feature} {
		if len(v) > 512 || strings.ContainsAny(v, "\x00\r\n\x1b") {
			add("evidência do card: campo %s inválido", field)
		}
	}
	supported := false
	for _, s := range cfg.PMMCP.Supported {
		if s == ev.Provider {
			supported = true
		}
	}
	switch {
	case !supported:
		add("evidência do card: provider %q fora de pm_mcp.supported (%s)", ev.Provider, strings.Join(cfg.PMMCP.Supported, ", "))
	case cfg.PMMCP.Provider != "auto" && cfg.PMMCP.Provider != ev.Provider:
		add("evidência do card: provider %q difere do configurado (%s)", ev.Provider, cfg.PMMCP.Provider)
	}
	if cfg.PMMCP.RequireExistingCard && (strings.TrimSpace(ev.ID) == "" || strings.TrimSpace(ev.URL) == "") {
		add("o card precisa existir: id e url são obrigatórios na evidência")
	}
	if u, err := url.Parse(ev.URL); ev.URL != "" && (err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "") {
		add("evidência do card: url inválida")
	}
	if ev.Feature != feature && ev.Feature != initiative {
		add("evidência do card é da feature %q, não de %q", ev.Feature, feature)
	}
	at, perr := time.Parse(time.RFC3339, ev.CheckedAt)
	switch {
	case perr != nil:
		add("evidência do card: checked_at inválido (use RFC3339)")
	case now.Sub(at) > time.Duration(cfg.PMMCP.EvidenceMaxAgeMinutes)*time.Minute:
		add("evidência do card vencida: checada há mais de %d minutos; consulte o card de novo", cfg.PMMCP.EvidenceMaxAgeMinutes)
	case at.Sub(now) > 5*time.Minute:
		add("evidência do card: checked_at no futuro")
	}
	return ev, pend
}

func jobID(feature string, now time.Time) string {
	id := fmt.Sprintf("job-%s-%s", now.UTC().Format("20060102-150405"), feature)
	if len(id) > 80 {
		id = id[:80]
	}
	return strings.TrimRight(id, "-.")
}

func initiativeTitle(base, feature string) string {
	data, err := os.ReadFile(filepath.Join(base, "discovery.md"))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "# ") {
				t := strings.TrimSpace(strings.TrimPrefix(line, "# "))
				t = strings.Map(func(r rune) rune {
					if r < 0x20 || r == 0x7f {
						return ' '
					}
					return r
				}, t)
				if len(t) > 200 {
					t = strings.ToValidUTF8(t[:200], "")
				}
				if t != "" {
					return t
				}
			}
		}
	}
	return feature
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
