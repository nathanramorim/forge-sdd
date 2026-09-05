package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/spf13/cobra"
)

// sessionMetrics espelha sdd/.metrics/schema.json.
type sessionMetrics struct {
	Schema           string   `json:"$schema"`
	Feature          string   `json:"feature"`
	Phase            string   `json:"phase,omitempty"`
	AgentPath        []string `json:"agent_path"`
	TokensInput      int      `json:"tokens_input"`
	TokensOutput     int      `json:"tokens_output"`
	Turns            int      `json:"turns"`
	DurationSeconds  int      `json:"duration_seconds"`
	CriterioAtendido bool     `json:"criterio_atendido"`
	Model            string   `json:"model"`
	ReworkLines      int      `json:"rework_lines"`
	Context7Calls    int      `json:"context7_calls"`
	GitCalls         int      `json:"git_calls"`
	FilesTouched     []string `json:"files_touched"`
	Outcome          string   `json:"outcome"`
	SddVersion       string   `json:"sdd_version"`
}

var validOutcomes = map[string]bool{"approved": true, "rejected": true, "blocked": true}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Gerenciamento determinístico de telemetria de sessão",
}

var sessionRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Grava sdd/.metrics/session-<ISO8601>.json de forma determinística (não depende de instrução LLM)",
	Long: `Grava um arquivo de telemetria de sessão respeitando sdd/.metrics/schema.json.

Pensado para ser chamado por QUALQUER comando que encerra uma sessão
(/proxima-feature, /revisar, /novo-fix) — não apenas o último passo de um
prompt longo. Se sdd/.sddrc tiver telemetry.enabled=false, não grava nada
e retorna sucesso (silencioso, por design).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}

		outcome, _ := cmd.Flags().GetString("outcome")
		if !validOutcomes[outcome] {
			return fmt.Errorf("--outcome inválido: %q (use approved, rejected ou blocked)", outcome)
		}

		rc, err := config.ReadSddrc(targetDir)
		if err != nil {
			return err
		}
		if !rc.Telemetry.Enabled {
			fmt.Println("Telemetria desabilitada em sdd/.sddrc (telemetry.enabled=false) — nada gravado.")
			return nil
		}

		feature, _ := cmd.Flags().GetString("feature")
		phase, _ := cmd.Flags().GetString("phase")
		agentPathCsv, _ := cmd.Flags().GetString("agent-path")
		tokensInput, _ := cmd.Flags().GetInt("tokens-input")
		tokensOutput, _ := cmd.Flags().GetInt("tokens-output")
		turns, _ := cmd.Flags().GetInt("turns")
		duration, _ := cmd.Flags().GetInt("duration-seconds")
		criterio, _ := cmd.Flags().GetBool("criterio-atendido")
		model, _ := cmd.Flags().GetString("model")
		reworkLines, _ := cmd.Flags().GetInt("rework-lines")
		context7Calls, _ := cmd.Flags().GetInt("context7-calls")
		gitCalls, _ := cmd.Flags().GetInt("git-calls")
		filesTouchedCsv, _ := cmd.Flags().GetString("files-touched")

		if feature == "" {
			return fmt.Errorf("--feature é obrigatório")
		}

		metrics := sessionMetrics{
			Schema:           "forge-sdd/metrics/1.0",
			Feature:          feature,
			Phase:            phase,
			AgentPath:        splitCsv(agentPathCsv),
			TokensInput:      tokensInput,
			TokensOutput:     tokensOutput,
			Turns:            turns,
			DurationSeconds:  duration,
			CriterioAtendido: criterio,
			Model:            model,
			ReworkLines:      reworkLines,
			Context7Calls:    context7Calls,
			GitCalls:         gitCalls,
			FilesTouched:     splitCsv(filesTouchedCsv),
			Outcome:          outcome,
			SddVersion:       rc.Version,
		}

		path, err := WriteSessionMetrics(targetDir, metrics, time.Now().UTC())
		if err != nil {
			return err
		}
		fmt.Printf("✓ Telemetria gravada em %s\n", path)
		return nil
	},
}

// WriteSessionMetrics grava o arquivo de telemetria em sdd/.metrics/session-<ISO8601>.json.
// Extraída como função exportada (não apenas RunE) para ser testável e reaproveitável.
// O nome do arquivo tem resolução de segundo; se duas gravações caírem no mesmo
// segundo, um sufixo numérico (-2, -3, ...) é anexado para nunca sobrescrever
// silenciosamente uma telemetria já gravada.
func WriteSessionMetrics(targetDir string, metrics sessionMetrics, at time.Time) (string, error) {
	metricsDir := filepath.Join(targetDir, "sdd", ".metrics")
	if err := os.MkdirAll(metricsDir, 0o755); err != nil {
		return "", fmt.Errorf("falha ao criar sdd/.metrics: %w", err)
	}

	base := fmt.Sprintf("session-%s", at.Format("2006-01-02T150405Z"))
	filename := base + ".json"
	path := filepath.Join(metricsDir, filename)
	for i := 2; fileExists(path); i++ {
		filename = fmt.Sprintf("%s-%d.json", base, i)
		path = filepath.Join(metricsDir, filename)
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("falha ao serializar telemetria: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("falha ao gravar %s: %w", path, err)
	}
	return path, nil
}

// SessionMetricsSummary agrega os arquivos sdd/.metrics/session-*.json para
// fechar o gap "grava mas nunca lê" (feat-01-02).
type SessionMetricsSummary struct {
	Total     int
	Approved  int
	Rejected  int
	Blocked   int
	ByFeature map[string]int
}

// sessionFileEntry é um session-*.json já lido e com o timestamp extraído do
// próprio nome do arquivo (sem depender de um campo de data no schema).
type sessionFileEntry struct {
	Metrics   sessionMetrics
	Timestamp time.Time
}

var sessionFilenameRe = regexp.MustCompile(`^session-(\d{4}-\d{2}-\d{2}T\d{6}Z)(?:-\d+)?\.json$`)

// readSessionFiles lê todos os session-*.json em sdd/.metrics (ignorando
// schema.json e arquivos inválidos). Base compartilhada por
// AggregateSessionMetrics e pelo comando `report` (feat-55-02) — evita
// duplicar o parsing de arquivo.
func readSessionFiles(targetDir string) ([]sessionFileEntry, error) {
	metricsDir := filepath.Join(targetDir, "sdd", ".metrics")

	entries, err := os.ReadDir(metricsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("falha ao ler %s: %w", metricsDir, err)
	}

	var result []sessionFileEntry
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "session-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(metricsDir, name))
		if err != nil {
			continue
		}
		var m sessionMetrics
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		var ts time.Time
		if match := sessionFilenameRe.FindStringSubmatch(name); match != nil {
			if parsed, err := time.Parse("2006-01-02T150405Z", match[1]); err == nil {
				ts = parsed
			}
		}
		result = append(result, sessionFileEntry{Metrics: m, Timestamp: ts})
	}
	return result, nil
}

// AggregateSessionMetrics lê todos os session-*.json em sdd/.metrics e
// retorna um resumo agregado.
func AggregateSessionMetrics(targetDir string) (SessionMetricsSummary, error) {
	summary := SessionMetricsSummary{ByFeature: map[string]int{}}
	records, err := readSessionFiles(targetDir)
	if err != nil {
		return summary, err
	}

	for _, r := range records {
		m := r.Metrics
		summary.Total++
		switch m.Outcome {
		case "approved":
			summary.Approved++
		case "rejected":
			summary.Rejected++
		case "blocked":
			summary.Blocked++
		}
		if m.Feature != "" {
			summary.ByFeature[m.Feature]++
		}
	}
	return summary, nil
}

// ClassifySessionType infere o tipo de item (discovery/feature/fix) a partir
// do caminho relativo gravado em sessionMetrics.Feature — sem exigir um campo
// novo no schema nem migração das métricas já gravadas (feat-55-01).
func ClassifySessionType(feature string) string {
	feature = filepath.ToSlash(feature)
	switch {
	case strings.Contains(feature, "sdd/discovery/"):
		return "discovery"
	case strings.Contains(feature, "sdd/features/") && strings.HasPrefix(filepath.Base(feature), "fix-"):
		return "fix"
	case strings.Contains(feature, "sdd/features/"):
		return "feature"
	default:
		return "outro"
	}
}

func splitCsv(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	sessionRecordCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	sessionRecordCmd.Flags().String("feature", "", "Caminho relativo da feature/task (obrigatório)")
	sessionRecordCmd.Flags().String("phase", "", "Identificador de fase/feature (ex: 01-01, 50)")
	sessionRecordCmd.Flags().String("agent-path", "", "Papéis acionados na sessão, separados por vírgula (ex: orquestrador,builder)")
	sessionRecordCmd.Flags().Int("tokens-input", 0, "Estimativa de tokens de entrada")
	sessionRecordCmd.Flags().Int("tokens-output", 0, "Estimativa de tokens de saída")
	sessionRecordCmd.Flags().Int("turns", 0, "Número de turnos da sessão")
	sessionRecordCmd.Flags().Int("duration-seconds", 0, "Duração estimada da sessão em segundos")
	sessionRecordCmd.Flags().Bool("criterio-atendido", false, "Se o critério de aceitação da feature foi atendido")
	sessionRecordCmd.Flags().String("model", "", "Modelo usado na sessão")
	sessionRecordCmd.Flags().Int("rework-lines", 0, "Linhas retrabalhadas na sessão")
	sessionRecordCmd.Flags().Int("context7-calls", 0, "Número de chamadas ao MCP context7")
	sessionRecordCmd.Flags().Int("git-calls", 0, "Número de chamadas ao MCP git")
	sessionRecordCmd.Flags().String("files-touched", "", "Arquivos tocados na sessão, separados por vírgula")
	sessionRecordCmd.Flags().String("outcome", "", "Resultado da sessão: approved, rejected ou blocked (obrigatório)")

	sessionCmd.AddCommand(sessionRecordCmd)
}
