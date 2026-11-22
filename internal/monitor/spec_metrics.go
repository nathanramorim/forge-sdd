package monitor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/notify"
)

// SpecMetrics é o bloco opcional "spec" do session-*.json das sessões de concepção
// (/discovery, /nova-feature, /novo-fix): só números, enum e o id da iniciativa saneado
// (SEG-01/SEG-05). Nunca carrega caminho nem conteúdo da spec.
type SpecMetrics struct {
	Tipo               string `json:"tipo"`
	Iniciativa         string `json:"iniciativa"`
	Bytes              int    `json:"bytes"`
	SecoesObrigatorias int    `json:"secoes_obrigatorias,omitempty"`
	CriteriaBytes      int    `json:"criteria_bytes,omitempty"`
	PlanBytes          int    `json:"plan_bytes,omitempty"`
	StatusBytes        int    `json:"status_bytes,omitempty"`
	FeaturesCount      int    `json:"features_count,omitempty"`
	FixesCount         int    `json:"fixes_count,omitempty"`
}

const (
	specIniciativaMax = 80
	specReadLimit     = 1 << 20
)

var specPathRe = regexp.MustCompile(`^sdd/discovery/([0-9a-z]+-[a-z0-9-]+)/(discovery\.md|features/feat-[^/]+\.md|fixes/fix-[^/]+\.md)$`)

// specRequiredHeadings são os seis títulos obrigatórios da skill sdd-concepcao.
var specRequiredHeadings = []string{
	"## Modelo do módulo",
	"## Decisões do projeto aplicadas",
	"## O que já existe",
	"## Dependências novas",
	"## Rules e IDs da Constituição Aplicáveis",
	"## Critérios de Aceitação Executáveis",
}

// CollectSpecMetrics monta o bloco spec a partir de feature (caminho relativo ao projeto).
// Valida na borda: relativo, sem "..", arquivo regular (Lstat, sem link simbólico, também nos
// ancestrais) dentro de sdd/discovery/<id>-<nome>/. Qualquer falha devolve erro e nenhum bloco.
func CollectSpecMetrics(targetDir, feature string) (*SpecMetrics, error) {
	slash := filepath.ToSlash(feature)
	if slash == "" || filepath.IsAbs(feature) || strings.HasPrefix(slash, "/") {
		return nil, errors.New("caminho precisa ser relativo ao projeto")
	}
	for _, part := range strings.Split(slash, "/") {
		if part == ".." {
			return nil, errors.New("caminho com '..' recusado")
		}
	}
	slash = strings.TrimPrefix(slash, "./")
	m := specPathRe.FindStringSubmatch(slash)
	if m == nil {
		return nil, errors.New("caminho fora de sdd/discovery/<id>-<nome>/ (discovery.md, features/feat-*.md ou fixes/fix-*.md)")
	}
	iniciativa, rest := m[1], m[2]

	root, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolver projeto: %w", err)
	}
	full := filepath.Join(root, filepath.FromSlash(slash))
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return nil, fmt.Errorf("arquivo da spec indisponível: %w", err)
	}
	if resolved != full {
		return nil, errors.New("link simbólico no caminho recusado")
	}
	fi, err := os.Lstat(full)
	if err != nil || !fi.Mode().IsRegular() {
		return nil, errors.New("--feature não é arquivo regular")
	}

	sm := &SpecMetrics{Iniciativa: notify.Sanitize(iniciativa, specIniciativaMax+1, false), Bytes: int(fi.Size())}
	switch {
	case rest == "discovery.md":
		sm.Tipo = "discovery"
		dir := filepath.Dir(full)
		sm.CriteriaBytes = regularSize(filepath.Join(dir, "criteria.md"))
		sm.PlanBytes = regularSize(filepath.Join(dir, "plan.md"))
		sm.StatusBytes = regularSize(filepath.Join(dir, "status.md"))
		sm.FeaturesCount = countRegularMd(filepath.Join(dir, "features"))
		sm.FixesCount = countRegularMd(filepath.Join(dir, "fixes"))
	case strings.HasPrefix(rest, "features/"):
		sm.Tipo = "feature"
		sm.SecoesObrigatorias = countRequiredSections(full)
	default:
		sm.Tipo = "fix"
		sm.SecoesObrigatorias = countRequiredSections(full)
	}
	if err := sm.validate(); err != nil {
		return nil, err
	}
	return sm, nil
}

// validate recusa o bloco inteiro quando algum campo é inválido (nunca grava valor ruim).
func (s *SpecMetrics) validate() error {
	switch s.Tipo {
	case "discovery", "feature", "fix":
	default:
		return fmt.Errorf("tipo inválido: %q", s.Tipo)
	}
	if s.Iniciativa == "" || len(s.Iniciativa) > specIniciativaMax {
		return errors.New("iniciativa vazia ou acima de 80 caracteres")
	}
	for _, n := range []int{s.Bytes, s.SecoesObrigatorias, s.CriteriaBytes, s.PlanBytes, s.StatusBytes, s.FeaturesCount, s.FixesCount} {
		if n < 0 {
			return errors.New("valor negativo no bloco spec")
		}
	}
	if s.SecoesObrigatorias > len(specRequiredHeadings) {
		return errors.New("secoes_obrigatorias acima de 6")
	}
	return nil
}

func regularSize(path string) int {
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return 0
	}
	return int(fi.Size())
}

func countRegularMd(dir string) int {
	if fi, err := os.Lstat(dir); err != nil || !fi.IsDir() {
		return 0
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") && e.Type().IsRegular() {
			n++
		}
	}
	return n
}

func countRequiredSections(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, specReadLimit))
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	n := 0
	for _, h := range specRequiredHeadings {
		for _, line := range lines {
			if strings.TrimRight(line, " \r") == h {
				n++
				break
			}
		}
	}
	return n
}
