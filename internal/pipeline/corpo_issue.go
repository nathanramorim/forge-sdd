package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/vcs"
)

const (
	secaoDescricaoNegocio = "descrição de negócio"
	maxTituloIssue        = 120
	maxDescricaoIssue     = 8 * 1024
	statusInicialCard     = "Spec"
)

// ErrSemDescricaoNegocio indica spec sem a seção "## Descrição de negócio" preenchida.
var ErrSemDescricaoNegocio = errors.New("a spec não tem a seção \"## Descrição de negócio\" preenchida: a issue não é criada")

// CorpoIssue é o título e o corpo comercial higienizados de uma issue (feat 113-03).
type CorpoIssue struct {
	Titulo string
	Corpo  string
}

func higieniza(s string, max int, keepNewlines bool) string {
	return strings.TrimSpace(notify.Sanitize(gate.MaskSecrets(s), max, keepNewlines))
}

// MontaCorpoIssue lê a spec (relativa a dir, confinada ao projeto) e monta título e corpo só com a
// Descrição de negócio e o link relativo da spec (SEG-02, SEG-05, SEG-08). Nunca cita caminho absoluto.
func MontaCorpoIssue(dir, specRel string) (CorpoIssue, error) {
	var out CorpoIssue
	resolved, err := gate.ConfineToDir(dir, specRel)
	if err != nil {
		return out, err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return out, fmt.Errorf("não foi possível ler a spec %s", filepath.ToSlash(specRel))
	}
	var titulo string
	var desc []string
	inDesc := false
	for _, ln := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(ln, "# ") && titulo == "":
			titulo = strings.TrimSpace(ln[2:])
		case strings.HasPrefix(ln, "## "):
			inDesc = strings.EqualFold(strings.TrimSpace(ln[3:]), secaoDescricaoNegocio)
		case inDesc:
			desc = append(desc, ln)
		}
	}
	out.Titulo = higieniza(titulo, maxTituloIssue, false)
	d := higieniza(strings.Join(desc, "\n"), maxDescricaoIssue, true)
	if out.Titulo == "" {
		return CorpoIssue{}, errors.New("a spec não tem título (linha \"# ...\")")
	}
	if d == "" {
		return CorpoIssue{}, ErrSemDescricaoNegocio
	}
	link := filepath.ToSlash(filepath.Clean(specRel))
	out.Corpo = fmt.Sprintf("%s\n\nEspecificação: [%s](%s)\n", d, link, link)
	return out, nil
}

// ResultadoIssue descreve o que CriarIssueConcepcao fez. Feito=false e Motivo preenchido: nada foi
// criado nem gravado (VCS sem issue), sem erro.
type ResultadoIssue struct {
	Feito     bool
	Reaprovei bool // a evidência já existia e foi mantida
	Motivo    string
	Evidencia CardEvidence
}

// GravaCardEvidence grava <iniciativaRel>/card-evidence.json de forma atômica, confinada ao projeto e
// sem seguir symlink (SEG-02). Evidência válida já existente para a mesma iniciativa é mantida sem
// reescrita (RN-4); devolve a evidência vigente e se reaproveitou.
func GravaCardEvidence(dir, iniciativaRel string, ev CardEvidence) (CardEvidence, bool, error) {
	pasta, err := gate.ConfineToDir(dir, iniciativaRel)
	if err != nil {
		return ev, false, err
	}
	if fi, err := os.Stat(pasta); err != nil || !fi.IsDir() {
		return ev, false, fmt.Errorf("a pasta da iniciativa %s não existe", filepath.ToSlash(iniciativaRel))
	}
	alvo := filepath.Join(pasta, CardEvidenceFile)
	switch fi, err := os.Lstat(alvo); {
	case err == nil && fi.Mode()&os.ModeSymlink != 0:
		return ev, false, fmt.Errorf("%s é link simbólico: recusado (SEG-02)", CardEvidenceFile)
	case err == nil:
		if atual, ok := leEvidenciaExistente(alvo, ev.Feature); ok {
			return atual, true, nil
		}
	case !errors.Is(err, os.ErrNotExist):
		return ev, false, fmt.Errorf("não foi possível inspecionar %s: %w", CardEvidenceFile, err)
	}
	if msg := validaEvidenciaParaGravar(ev); msg != "" {
		return ev, false, errors.New(msg)
	}
	if err := queue.WriteJSONAtomic(alvo, ev, 0o644); err != nil {
		return ev, false, err
	}
	return ev, false, nil
}

// leEvidenciaExistente devolve a evidência válida da mesma feature; inválida, ilegível, symlink ou
// de outra feature não conta (nada é apagado).
func leEvidenciaExistente(alvo, feature string) (CardEvidence, bool) {
	var atual CardEvidence
	fi, err := os.Lstat(alvo)
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > 16*1024 {
		return atual, false
	}
	raw, err := os.ReadFile(alvo)
	if err != nil {
		return atual, false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&atual) != nil || atual.Feature != feature || atual.ID == "" || atual.URL == "" {
		return CardEvidence{}, false
	}
	return atual, true
}

func validaEvidenciaParaGravar(ev CardEvidence) string {
	for nome, v := range map[string]string{"provider": ev.Provider, "id": ev.ID, "url": ev.URL, "status": ev.Status, "feature": ev.Feature} {
		if strings.TrimSpace(v) == "" || len(v) > 512 || strings.ContainsAny(v, "\x00\r\n\x1b") {
			return fmt.Sprintf("evidência do card: campo %s inválido", nome)
		}
	}
	if u, err := url.Parse(ev.URL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "evidência do card: url inválida"
	}
	if _, err := time.Parse(time.RFC3339, ev.CheckedAt); err != nil {
		return "evidência do card: checked_at inválido (use RFC3339)"
	}
	return ""
}

// CriarIssueConcepcao cria (ou reaproveita) a issue da concepção e grava a evidência. specRel e
// iniciativaRel são relativos ao projeto. O adaptador `nenhum` devolve vcs.ErrSemPR: isso é "sem
// efeito, sem erro" (nada criado, nada gravado). O título-base é o título da spec, estável, para o
// filtro exato de ProcurarIssueAberta; esta busca vê só os 30 itens mais recentes do `gh`, então uma
// issue antiga fora da janela pode ser duplicada (falha segura; follow-up: --limit maior em internal/vcs).
func CriarIssueConcepcao(dir, specRel, iniciativaRel string, ad vcs.Adaptador, now time.Time) (ResultadoIssue, error) {
	var res ResultadoIssue
	feature := filepath.Base(filepath.Clean(iniciativaRel))
	corpo, err := MontaCorpoIssue(dir, specRel)
	if err != nil {
		return res, err
	}
	semEfeito := func() (ResultadoIssue, error) {
		return ResultadoIssue{Motivo: "VCS sem suporte a issue: nada foi criado nem gravado"}, nil
	}
	// A evidência tem precedência sobre a busca por título (RN-4): sem chamada externa.
	if pasta, cerr := gate.ConfineToDir(dir, iniciativaRel); cerr == nil {
		if atual, ok := leEvidenciaExistente(filepath.Join(pasta, CardEvidenceFile), feature); ok {
			return ResultadoIssue{Reaprovei: true, Motivo: fmt.Sprintf("a evidência já existe (id %s, %s): nada foi criado", atual.ID, atual.URL), Evidencia: atual}, nil
		}
	}
	iss, err := ad.ProcurarIssueAberta(dir, corpo.Titulo)
	if errors.Is(err, vcs.ErrSemPR) {
		return semEfeito()
	}
	if err != nil {
		return res, err
	}
	if iss == nil {
		novo, err := ad.CriarIssue(dir, vcs.NovaIssue{Titulo: corpo.Titulo, Corpo: corpo.Corpo})
		if errors.Is(err, vcs.ErrSemPR) {
			return semEfeito()
		}
		if err != nil {
			return res, err
		}
		iss = &novo
	}
	ev, reaproveitou, err := GravaCardEvidence(dir, iniciativaRel, CardEvidence{
		Provider: "github-projects", ID: strconv.Itoa(iss.Numero), URL: iss.URL, Status: statusInicialCard,
		Feature: feature, CheckedAt: now.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return res, err
	}
	return ResultadoIssue{Feito: true, Reaprovei: reaproveitou, Evidencia: ev}, nil
}
