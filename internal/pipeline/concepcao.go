package pipeline

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/vcs"
)

var (
	featureConcepcaoRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)
	idCardRe           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._#:/-]{0,63}$`)
)

// OpcoesConcepcao reúne a entrada do vínculo da concepção (feat 113-04). Provider, ID e URL
// preenchidos juntos registram o retorno de um MCP; sem eles o gestor é resolvido.
type OpcoesConcepcao struct {
	Dir, Feature      string
	Provider, ID, URL string
	DryRun            bool
	Run               vcs.Runner
	Now               time.Time
}

// ResultadoConcepcao: Motivo e Aviso são texto livre e devem ser higienizados por quem imprime
// (Higieniza). Falha externa vira Aviso, nunca erro (RN-5).
type ResultadoConcepcao struct {
	Modo      string
	Feito     bool
	Reaprovei bool
	DryRun    bool
	Motivo    string
	Aviso     string
	Evidencia CardEvidence
}

// Higieniza mascara segredos e remove controle/ANSI de texto não confiável antes de imprimir (SEG-08).
func Higieniza(s string) string { return higieniza(s, 300, false) }

// ValidaCamposCard recusa feature, provider, id e url inválidos (SEG-01). A URL exige https.
func ValidaCamposCard(supported []string, feature, provider, id, rawURL string) error {
	if !featureConcepcaoRe.MatchString(feature) {
		return errors.New("--feature inválida")
	}
	if provider == "" && id == "" && rawURL == "" {
		return nil
	}
	if provider == "" || id == "" || rawURL == "" {
		return errors.New("--provider, --id e --url devem vir juntos")
	}
	ok := false
	for _, s := range supported {
		ok = ok || s == provider
	}
	if !ok {
		return errors.New("--provider fora de pm_mcp.supported")
	}
	if !idCardRe.MatchString(id) {
		return errors.New("--id inválido")
	}
	u, err := url.Parse(rawURL)
	if err != nil || len(rawURL) > 512 || strings.ContainsAny(rawURL, "\x00\r\n\x1b \t") || u.Scheme != "https" || u.Host == "" {
		return errors.New("--url inválida (exige https)")
	}
	return nil
}

func acharSpec(dir, feature string) (specRel, iniRel string, err error) {
	m, _ := filepath.Glob(filepath.Join(dir, "sdd", "discovery", "*", "features", feature+".md"))
	if len(m) != 1 {
		return "", "", fmt.Errorf("spec da feature %s não encontrada", feature)
	}
	rel, err := filepath.Rel(dir, m[0])
	if err != nil {
		return "", "", err
	}
	rel = filepath.ToSlash(rel)
	return rel, filepath.ToSlash(filepath.Dir(filepath.Dir(rel))), nil
}

// Concepcao orquestra o vínculo da spec com o card: desligado = custo zero; github cria ou reaproveita
// a issue; mcp registra o retorno informado. DryRun não chama o gh nem grava (ARQ-05).
func Concepcao(o OpcoesConcepcao) (ResultadoConcepcao, error) {
	res := ResultadoConcepcao{DryRun: o.DryRun}
	cfg, _, err := config.LoadForgeConfig(o.Dir)
	if err != nil {
		return res, err
	}
	if err := ValidaCamposCard(cfg.PMMCP.Supported, o.Feature, o.Provider, o.ID, o.URL); err != nil {
		return res, err
	}
	nome, err := vcs.LeVCS(o.Dir)
	if err != nil {
		return res, err
	}
	g, err := config.ResolverGestorConcepcao(o.Dir, nome)
	if err != nil {
		return res, err
	}
	res.Modo = g.Modo
	if g.Modo == config.GestorDesligado {
		return res, nil
	}
	specRel, iniRel, err := acharSpec(o.Dir, o.Feature)
	if err != nil {
		return res, err
	}
	now := o.Now
	if now.IsZero() {
		now = time.Now()
	}
	pasta := filepath.Join(o.Dir, filepath.FromSlash(iniRel), CardEvidenceFile)
	if atual, ok := leEvidenciaExistente(pasta, o.Feature); ok && (o.DryRun || o.ID == "") {
		res.Reaprovei, res.Evidencia = true, atual
		res.Motivo = fmt.Sprintf("a evidência já existe (id %s, %s)", atual.ID, atual.URL)
		return res, nil
	}
	if o.ID != "" { // retorno do MCP
		if o.DryRun {
			res.Motivo = "dry-run: registraria a evidência do card"
			return res, nil
		}
		ev, reap, err := GravaCardEvidence(o.Dir, iniRel, CardEvidence{Provider: o.Provider, ID: o.ID, URL: o.URL,
			Status: statusInicialCard, Feature: o.Feature, CheckedAt: now.UTC().Format(time.RFC3339)})
		if err != nil {
			return res, err
		}
		res.Feito, res.Reaprovei, res.Evidencia = !reap, reap, ev
		return res, nil
	}
	if g.Modo == config.GestorMCP {
		res.Motivo = fmt.Sprintf("crie o card via MCP %s e registre com --provider, --id e --url", g.Provider)
		return res, nil
	}
	if o.DryRun {
		res.Motivo = "dry-run: criaria a issue e gravaria a evidência"
		return res, nil
	}
	ad, err := vcs.Novo(nome, o.Run)
	if err != nil {
		res.Aviso = err.Error()
		return res, nil
	}
	r, err := CriarIssueConcepcao(o.Dir, specRel, iniRel, ad, now)
	if err != nil {
		res.Aviso = "o card não foi criado: " + err.Error()
		return res, nil
	}
	if !r.Feito && !r.Reaprovei { // ErrSemPR: sem efeito, sem erro
		res.Motivo = r.Motivo
		return res, nil
	}
	res.Feito, res.Reaprovei, res.Motivo, res.Evidencia = r.Feito, r.Reaprovei, r.Motivo, r.Evidencia
	return res, nil
}
