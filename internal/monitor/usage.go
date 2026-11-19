package monitor

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/layout"
)

// Fallbacks de bytes para EstimateJobUsage e EstimateSessionUsage quando o arquivo não está disponível.
const (
	estimateFallbackInputBytes  = 4000
	estimateFallbackOutputBytes = 2000
)

// Valores de SessionMetrics.Medicao.
const (
	MedicaoReal     = "real"
	MedicaoEstimado = "estimado"
	// MedicaoResumo marca o registro do archive: resume a iniciativa e não entra nas somas.
	MedicaoResumo = "resumo"
)

// JobUsage é o uso de um job: medido pelo evento `result` do log stream-json ou estimado por bytes.
type JobUsage struct {
	TokensInput         int
	TokensOutput        int
	TokensCacheRead     int
	TokensCacheCreation int
	Turns               int
	DurationSeconds     int
	Model               string
	CostUSDTabela       float64
	Medicao             string
}

type resultEvent struct {
	Type         string  `json:"type"`
	NumTurns     int     `json:"num_turns"`
	DurationMs   int64   `json:"duration_ms"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        struct {
		Input         int `json:"input_tokens"`
		Output        int `json:"output_tokens"`
		CacheRead     int `json:"cache_read_input_tokens"`
		CacheCreation int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
	ModelUsage map[string]json.RawMessage `json:"modelUsage"`
}

// ParseJobUsage soma os eventos `result` de um log stream-json. O bool é falso quando não há nenhum.
// Linhas que não são JSON são ignoradas; só contadores, modelo e duração são lidos (SEG-05).
func ParseJobUsage(logPath string) (JobUsage, bool) {
	f, err := os.Open(logPath)
	if err != nil {
		return JobUsage{}, false
	}
	defer f.Close()

	u := JobUsage{Medicao: MedicaoReal}
	models := map[string]bool{}
	found := false
	var durationMs int64
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "{") || !strings.Contains(line, `"result"`) {
			continue
		}
		var ev resultEvent
		if json.Unmarshal([]byte(line), &ev) != nil || ev.Type != "result" {
			continue
		}
		found = true
		u.TokensInput += ev.Usage.Input
		u.TokensOutput += ev.Usage.Output
		u.TokensCacheRead += ev.Usage.CacheRead
		u.TokensCacheCreation += ev.Usage.CacheCreation
		u.Turns += ev.NumTurns
		u.CostUSDTabela += ev.TotalCostUSD
		durationMs += ev.DurationMs
		for m := range ev.ModelUsage {
			models[m] = true
		}
	}
	if !found {
		return JobUsage{}, false
	}
	u.DurationSeconds = int(durationMs / 1000)
	names := make([]string, 0, len(models))
	for m := range models {
		names = append(names, m)
	}
	sort.Strings(names)
	u.Model = strings.Join(names, ",")
	return u, true
}

// EstimateJobUsage estima o uso pelo tamanho da spec e do log, para CLIs sem evento `result`.
func EstimateJobUsage(dir, role, jobID, specPath string) JobUsage {
	specBytes := estimateFallbackInputBytes
	if specPath != "" {
		full := specPath
		if dir != "" && !filepath.IsAbs(full) {
			full = filepath.Join(dir, full)
		}
		if info, err := os.Stat(full); err == nil && !info.IsDir() && info.Size() > 0 {
			specBytes = int(info.Size())
		}
	}
	logBytes := estimateFallbackOutputBytes
	if info, err := os.Stat(jobLogPath(dir, role, jobID)); err == nil && !info.IsDir() && info.Size() > 0 {
		logBytes = int(info.Size())
	}
	return JobUsage{
		TokensInput:  EstimateTokensFromBytes(specBytes * 4),
		TokensOutput: EstimateTokensFromBytes(logBytes),
		Medicao:      MedicaoEstimado,
	}
}

// EstimateSessionUsage estima o uso de tokens de uma sessão de skill a partir do tamanho
// do arquivo de feature (proxy para input) e dos arquivos tocados (proxy para output).
// Tenta o feature como caminho relativo direto; se não encontrar, busca via glob de ID.
func EstimateSessionUsage(dir, feature string, filesTouched []string) JobUsage {
	var inputBytes int64
	if feature != "" {
		if info, err := os.Stat(filepath.Join(dir, feature)); err == nil && !info.IsDir() && info.Size() > 0 {
			inputBytes = info.Size()
		} else {
			matches, _ := filepath.Glob(filepath.Join(dir, "sdd", "discovery", "*", "features", feature+"*.md"))
			if len(matches) == 0 {
				matches, _ = filepath.Glob(filepath.Join(dir, "sdd", "discovery", "*", "fixes", feature+"*.md"))
			}
			if len(matches) > 0 {
				if info, err := os.Stat(matches[0]); err == nil {
					inputBytes = info.Size()
				}
			}
		}
	}
	if inputBytes == 0 {
		inputBytes = estimateFallbackInputBytes
	}

	var outputBytes int64
	for _, f := range filesTouched {
		if info, err := os.Stat(filepath.Join(dir, f)); err == nil {
			outputBytes += info.Size()
		}
	}
	if outputBytes == 0 {
		outputBytes = estimateFallbackOutputBytes
	}

	return JobUsage{
		TokensInput:  EstimateTokensFromBytes(int(inputBytes) * 4),
		TokensOutput: EstimateTokensFromBytes(int(outputBytes)),
		Medicao:      MedicaoEstimado,
	}
}

// ReadJobUsage devolve o uso real do log do job e, sem evento `result`, a estimativa.
func ReadJobUsage(dir, role, jobID, specPath string) JobUsage {
	if u, ok := ParseJobUsage(jobLogPath(dir, role, jobID)); ok {
		return u
	}
	return EstimateJobUsage(dir, role, jobID, specPath)
}

func jobLogPath(dir, role, jobID string) string {
	return filepath.Join(layout.Logs(dir), role+"-"+jobID+".log")
}

// UsageBucket soma o uso de um conjunto de sessões com a mesma origem de medição.
type UsageBucket struct {
	Sessions            int
	TokensInput         int
	TokensOutput        int
	TokensCacheRead     int
	TokensCacheCreation int
	Turns               int
	CostUSDTabela       float64
}

// UsageTotals separa o uso real do estimado.
type UsageTotals struct {
	Real     UsageBucket
	Estimado UsageBucket
}

// BelongsToInitiative diz se a sessão pertence à iniciativa (spec sob sdd/discovery/<id>/ ou fase com o id).
func BelongsToInitiative(m SessionMetrics, initiative string) bool {
	if initiative == "" {
		return true
	}
	return strings.Contains(m.Feature, "/"+initiative+"/") || strings.Contains(m.Phase, initiative)
}

// SumUsage soma as sessões; registros antigos, sem `medicao`, contam como estimados.
func SumUsage(ms []SessionMetrics) UsageTotals {
	var t UsageTotals
	for _, m := range ms {
		if m.Medicao == MedicaoResumo {
			continue
		}
		b := &t.Estimado
		if m.Medicao == MedicaoReal {
			b = &t.Real
		}
		b.Sessions++
		b.TokensInput += m.TokensInput
		b.TokensOutput += m.TokensOutput
		b.TokensCacheRead += m.TokensCacheRead
		b.TokensCacheCreation += m.TokensCacheCreation
		b.Turns += m.Turns
		b.CostUSDTabela += m.CostUSDTabela
	}
	return t
}

// ReadSessions lê as sessões de sdd/.metrics que pertencem à iniciativa (arquivos inválidos são ignorados).
func ReadSessions(dir, initiative string) []SessionMetrics {
	entries, err := os.ReadDir(layout.Metrics(dir))
	if err != nil {
		return nil
	}
	var out []SessionMetrics
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, "session-") || !strings.HasSuffix(n, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(layout.Metrics(dir), n))
		if err != nil {
			continue
		}
		var m SessionMetrics
		if json.Unmarshal(data, &m) == nil && BelongsToInitiative(m, initiative) {
			out = append(out, m)
		}
	}
	return out
}
