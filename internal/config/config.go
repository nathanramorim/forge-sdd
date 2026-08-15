package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Agentes suportados.
const (
	AgentCopilot = "copilot"
	AgentClaude  = "claude"
	AgentGemini  = "gemini"
	AgentOpenAI  = "openai"
)

var validAgents = map[string]bool{
	AgentCopilot: true,
	AgentClaude:  true,
	AgentGemini:  true,
	AgentOpenAI:  true,
}

// Sddrc representa o conteúdo persistido em sdd/.sddrc.
type Sddrc struct {
	Project          string   `json:"project"`
	Version          string   `json:"version"`
	Lang             string   `json:"lang"`
	Agents           []string `json:"agents"`
	Telemetry        struct {
		Enabled bool `json:"enabled"`
	} `json:"telemetry"`
	NamingConvention string   `json:"naming_convention"`
}

// ToConfig converte Sddrc em Config para reuso nos comandos de scaffold.
func (r Sddrc) ToConfig() Config {
	return Config{
		Project:          r.Project,
		SddVersion:       r.Version,
		Lang:             r.Lang,
		Telemetry:        r.Telemetry.Enabled,
		Agents:           r.Agents,
		NamingConvention: r.NamingConvention,
	}
}

// ReadSddrc lê e parseia sdd/.sddrc no diretório alvo.
func ReadSddrc(targetDir string) (Sddrc, error) {
	path := filepath.Join(targetDir, "sdd", ".sddrc")
	data, err := os.ReadFile(path)
	if err != nil {
		return Sddrc{}, fmt.Errorf("sdd/.sddrc não encontrado em %s (execute forge-sdd init primeiro): %w", targetDir, err)
	}
	var rc Sddrc
	if err := json.Unmarshal(data, &rc); err != nil {
		return Sddrc{}, fmt.Errorf("sdd/.sddrc inválido: %w", err)
	}
	return rc, nil
}

// MergeAgents retorna existing + new sem duplicatas, preservando a ordem.
func MergeAgents(existing, new []string) []string {
	seen := make(map[string]bool, len(existing))
	result := make([]string, len(existing))
	copy(result, existing)
	for _, a := range existing {
		seen[a] = true
	}
	for _, a := range new {
		if !seen[a] {
			seen[a] = true
			result = append(result, a)
		}
	}
	return result
}

// FilterNewAgents retorna apenas os candidatos que ainda não estão em existing.
func FilterNewAgents(existing, candidates []string) []string {
	seen := make(map[string]bool, len(existing))
	for _, a := range existing {
		seen[a] = true
	}
	var result []string
	for _, a := range candidates {
		if !seen[a] {
			result = append(result, a)
		}
	}
	return result
}

// ReadProjectVersion lê a versão registrada em sdd/.sdd-version.
func ReadProjectVersion(targetDir string) (string, error) {
	path := filepath.Join(targetDir, "sdd", ".sdd-version")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("sdd/.sdd-version não encontrado: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// CompareVersions compara duas versões no formato "X.Y.Z".
// Retorna -1 se a < b, 0 se a == b, 1 se a > b.
func CompareVersions(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := range pa {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	var major, minor, patch int
	_, _ = fmt.Sscanf(strings.TrimPrefix(v, "v"), "%d.%d.%d", &major, &minor, &patch)
	return [3]int{major, minor, patch}
}

// ParseAgents converte uma string csv ("copilot,claude") em []string validado.
func ParseAgents(csv string) ([]string, error) {
	if csv == "" {
		return []string{AgentCopilot}, nil
	}
	parts := strings.Split(csv, ",")
	seen := make(map[string]bool, len(parts))
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if !validAgents[p] {
			return nil, fmt.Errorf("agente desconhecido: %q (válidos: copilot, claude, gemini, openai)", p)
		}
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result, nil
}

// Config contém todas as configurações coletadas durante o init.
type Config struct {
	Project          string   // nome do projeto
	Stack            string   // runtime principal: go, node, python, rust, other
	DB               string   // banco de dados: postgres, sqlite, mongo, none
	Telemetry        bool     // habilitar telemetria local
	Lang             string   // idioma dos templates: pt-BR, en
	SddVersion       string   // versão Forge-SDD a usar
	DryRun           bool     // exibir árvore sem criar arquivos
	Agents           []string // agentes de IA: copilot, claude, gemini
	NamingConvention string   // Convenção de nomenclatura: sequencial, hash, workitem
}

// Defaults retorna uma Config com valores padrão.
func Defaults() Config {
	return Config{
		Project:          "meu-projeto",
		Stack:            "go",
		DB:               "none",
		Telemetry:        true,
		Lang:             "pt-BR",
		SddVersion:       "2.0.0-beta",
		DryRun:           false,
		Agents:           []string{AgentCopilot},
		NamingConvention: "sequencial",
	}
}

// FromFlags preenche a Config a partir das flags registradas no cobra.Command.
// Começa com Defaults() e sobrescreve apenas as flags que foram explicitamente passadas.
func FromFlags(cmd *cobra.Command) Config {
	cfg := Defaults()
	if f := cmd.Flags().Lookup("name"); f != nil && f.Changed {
		cfg.Project = f.Value.String()
	}
	if f := cmd.Flags().Lookup("stack"); f != nil && f.Changed {
		cfg.Stack = f.Value.String()
	}
	if f := cmd.Flags().Lookup("db"); f != nil && f.Changed {
		cfg.DB = f.Value.String()
	}
	if f := cmd.Flags().Lookup("lang"); f != nil && f.Changed {
		cfg.Lang = f.Value.String()
	}
	if f := cmd.Flags().Lookup("version"); f != nil && f.Changed {
		cfg.SddVersion = f.Value.String()
	}
	if f := cmd.Flags().Lookup("no-telemetry"); f != nil && f.Changed {
		v, _ := cmd.Flags().GetBool("no-telemetry")
		cfg.Telemetry = !v
	}
	if f := cmd.Flags().Lookup("agent"); f != nil && f.Changed {
		agents, err := ParseAgents(f.Value.String())
		if err != nil {
			// erro de validação — mantém default; será relatado em RunE
			cfg.Agents = []string{AgentCopilot}
		} else {
			cfg.Agents = agents
		}
	}
	if f := cmd.Flags().Lookup("naming-convention"); f != nil && f.Changed {
		cfg.NamingConvention = f.Value.String()
	}
	return cfg
}

// DetectProject verifica se o diretório já contém um projeto Forge-SDD.
func DetectProject(dir string) bool {
	path := filepath.Join(dir, "sdd", ".sddrc")
	_, err := os.Stat(path)
	return err == nil
}

var NpmRegistryURL = "https://registry.npmjs.org/@nathanramorim/forge-sdd"

// npmFetchTimeout é o timeout do client HTTP usado por FetchNpmVersions.
// 2s se mostrou curto demais em redes mais lentas, causando fallback silencioso
// para a ausência da versão beta (ver fix-45-update-beta-version-detection).
var npmFetchTimeout = 8 * time.Second

type npmRegistryResponse struct {
	DistTags map[string]string `json:"dist-tags"`
}

// FetchNpmVersions busca as últimas versões publicadas (latest e beta) no NPM Registry.
func FetchNpmVersions() (latest string, beta string, err error) {
	client := &http.Client{
		Timeout: npmFetchTimeout,
	}
	resp, err := client.Get(NpmRegistryURL)
	if err != nil {
		return "", "", fmt.Errorf("erro na requisição do NPM Registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("status HTTP inválido: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("erro ao ler corpo da resposta: %w", err)
	}

	var data npmRegistryResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", fmt.Errorf("erro ao decodificar JSON do NPM: %w", err)
	}

	latest = data.DistTags["latest"]
	beta = data.DistTags["beta"]
	return latest, beta, nil
}

var sequentialIDRe = regexp.MustCompile(`^(?:feat|fix|discovery|criteria|plan)-(\d{2})(?:[-.]|$)`)
var hashIDRe = regexp.MustCompile(`^(?:feat|fix|discovery|criteria|plan)-([0-9a-f]{4})(?:[-.]|$)`)

// ClassifyNamingConvention identifica se o nome de um arquivo de feature/discovery/fix
// segue a convenção sequencial (feat-NN/fix-NN) ou por hash de 4 dígitos (feat-xxxx/fix-xxxx).
// Retorna "" quando o nome não se encaixa em nenhuma das duas (ex: index.md).
func ClassifyNamingConvention(name string) string {
	if sequentialIDRe.MatchString(name) {
		return "sequencial"
	}
	if hashIDRe.MatchString(name) {
		return "hash"
	}
	return ""
}

// DetectNamingConvention varre o diretório e tenta inferir a convenção com base nos arquivos.
func DetectNamingConvention(targetDir string) string {
	featuresDir := filepath.Join(targetDir, "sdd", "features")
	discoveryDir := filepath.Join(targetDir, "sdd", "discovery")

	var seqCount, hashCount int
	scanDir := func(dir string) {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			switch ClassifyNamingConvention(d.Name()) {
			case "sequencial":
				seqCount++
			case "hash":
				hashCount++
			}
			return nil
		})
	}
	scanDir(featuresDir)
	scanDir(discoveryDir)

	if hashCount > 0 && seqCount == 0 {
		return "hash"
	}
	if seqCount > 0 && hashCount == 0 {
		return "sequencial"
	}
	if hashCount > seqCount {
		return "hash"
	}
	// default padrão
	return "sequencial"
}

// ResolveUpgradeTarget escolhe a versão mais recente aplicável entre latest e beta
// (comparando via CompareVersions), usada para resolver --upgrade sem depender
// da versão compilada localmente no binário.
func ResolveUpgradeTarget(latest, beta string) string {
	switch {
	case latest == "":
		return beta
	case beta == "":
		return latest
	case CompareVersions(beta, latest) > 0:
		return beta
	default:
		return latest
	}
}

