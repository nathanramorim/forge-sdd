package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/vcs"
)

const (
	iniTeste  = "sdd/discovery/900-teste"
	specTeste = iniTeste + "/features/feat-900-01-x.md"
)

func projetoComSpec(t *testing.T, conteudo string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, filepath.FromSlash(specTeste))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

const specOk = "# Feature 900-01 — Titulo\n\n## Objetivo\ninterno ARQ-11\n\n## Descrição de negócio\nEntrega valor ao cliente.\n\n## Modelo\nsegredo interno\n"

func TestCorpoIssueExtraiDescricaoDeNegocio(t *testing.T) {
	c, err := MontaCorpoIssue(projetoComSpec(t, specOk), specTeste)
	if err != nil {
		t.Fatal(err)
	}
	if c.Titulo != "Feature 900-01 — Titulo" || !strings.Contains(c.Corpo, "Entrega valor ao cliente.") ||
		strings.Contains(c.Corpo, "interno") || !strings.Contains(c.Corpo, "]("+specTeste+")") {
		t.Fatalf("corpo inesperado: %+v", c)
	}
}

func TestCorpoIssueMascaraSegredosEAnsi(t *testing.T) {
	spec := "# T \x1b[31mx\x1b[0m\n\n## Descrição de negócio\nusa API_KEY=abc123supersecret e \x1b[31mvermelho\x1b[0m\x07 ok\n"
	c, err := MontaCorpoIssue(projetoComSpec(t, spec), specTeste)
	if err != nil {
		t.Fatal(err)
	}
	all := c.Titulo + c.Corpo
	if strings.Contains(all, "abc123supersecret") || strings.ContainsAny(all, "\x1b\x07") || !strings.Contains(c.Corpo, "***") {
		t.Fatalf("segredo/ANSI vazou: %q", all)
	}
}

func TestCorpoIssueSemCaminhoAbsoluto(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	c, err := MontaCorpoIssue(dir, specTeste)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.Corpo, dir) || strings.Contains(c.Corpo, "(/") {
		t.Fatalf("caminho absoluto no corpo: %q", c.Corpo)
	}
	if _, err := MontaCorpoIssue(dir, filepath.Join(dir, filepath.FromSlash(specTeste))); err == nil {
		t.Fatal("spec com caminho absoluto deveria ser recusada")
	}
	if _, err := MontaCorpoIssue(dir, "../fora.md"); err == nil {
		t.Fatal("path traversal deveria ser recusado")
	}
}

func TestCorpoIssueSemDescricaoDeNegocioFalha(t *testing.T) {
	_, err := MontaCorpoIssue(projetoComSpec(t, "# T\n\n## Objetivo\nx\n"), specTeste)
	if !errors.Is(err, ErrSemDescricaoNegocio) {
		t.Fatalf("esperava ErrSemDescricaoNegocio, veio %v", err)
	}
	_, err = MontaCorpoIssue(projetoComSpec(t, "# T\n\n## Descrição de negócio\n\n   \n## X\n"), specTeste)
	if !errors.Is(err, ErrSemDescricaoNegocio) {
		t.Fatalf("seção vazia deveria falhar, veio %v", err)
	}
}

func evTeste() CardEvidence {
	return CardEvidence{Provider: "github-projects", ID: "7", URL: "https://github.com/o/r/issues/7", Status: "Spec", Feature: "900-teste", CheckedAt: time.Now().UTC().Format(time.RFC3339)}
}

func TestGravaCardEvidenceAtomico(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	if _, reap, err := GravaCardEvidence(dir, iniTeste, evTeste()); err != nil || reap {
		t.Fatalf("err=%v reap=%v", err, reap)
	}
	alvo := filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)
	if _, err := os.Stat(alvo); err != nil {
		t.Fatal(err)
	}
	ents, _ := os.ReadDir(filepath.Dir(alvo))
	for _, e := range ents {
		if strings.Contains(e.Name(), ".tmp") {
			t.Fatalf("temporário sobrou: %s", e.Name())
		}
	}
	bad := evTeste()
	bad.URL = "javascript:alert(1)"
	if _, _, err := GravaCardEvidence(dir, iniTeste, bad); err == nil {
		// existente é reaproveitada; remove e tenta de novo
		_ = os.Remove(alvo)
		if _, _, err := GravaCardEvidence(dir, iniTeste, bad); err == nil {
			t.Fatal("url inválida deveria ser recusada")
		}
	}
}

func TestGravaCardEvidenceIdempotenteReaproveitaExistente(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	if _, _, err := GravaCardEvidence(dir, iniTeste, evTeste()); err != nil {
		t.Fatal(err)
	}
	alvo := filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)
	antes, _ := os.ReadFile(alvo)
	outro := evTeste()
	outro.ID, outro.URL = "8", "https://github.com/o/r/issues/8"
	got, reap, err := GravaCardEvidence(dir, iniTeste, outro)
	depois, _ := os.ReadFile(alvo)
	if err != nil || !reap || got.ID != "7" || string(antes) != string(depois) {
		t.Fatalf("err=%v reap=%v got=%+v", err, reap, got)
	}
}

func TestGravaCardEvidenceRecusaSymlinkEFora(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink exige privilégio no Windows")
	}
	dir := projetoComSpec(t, specOk)
	fora := filepath.Join(t.TempDir(), "alvo.json")
	if err := os.WriteFile(fora, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)
	if err := os.Symlink(fora, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := GravaCardEvidence(dir, iniTeste, evTeste()); err == nil {
		t.Fatal("symlink deveria ser recusado")
	}
	if b, _ := os.ReadFile(fora); string(b) != "{}" {
		t.Fatal("arquivo fora do projeto foi alterado")
	}
	if _, _, err := GravaCardEvidence(dir, "../fora", evTeste()); err == nil {
		t.Fatal("caminho fora deveria ser recusado")
	}
	pastaFora := t.TempDir()
	if err := os.Symlink(pastaFora, filepath.Join(dir, "sdd", "ln")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := GravaCardEvidence(dir, "sdd/ln", evTeste()); err == nil {
		t.Fatal("pasta symlink para fora deveria ser recusada")
	}
}

func TestGravaCardEvidenceCompativelComLoadEvidence(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	ev := evTeste()
	if _, _, err := GravaCardEvidence(dir, iniTeste, ev); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := config.LoadForgeConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.ToSlash(filepath.Join(iniTeste, CardEvidenceFile))
	got, pend := loadEvidence(dir, rel, "900-teste", "900-teste", cfg, time.Now())
	if len(pend) != 0 || got.ID != "7" {
		t.Fatalf("loadEvidence recusou: %v", pend)
	}
}

type adFake struct {
	vcs.Adaptador
	achada   *vcs.Issue
	semPR    bool
	criadas  int
	buscadas int
}

func (a *adFake) ProcurarIssueAberta(string, string) (*vcs.Issue, error) {
	a.buscadas++
	if a.semPR {
		return nil, vcs.ErrSemPR
	}
	return a.achada, nil
}

func (a *adFake) CriarIssue(string, vcs.NovaIssue) (vcs.Issue, error) {
	if a.semPR {
		return vcs.Issue{}, vcs.ErrSemPR
	}
	a.criadas++
	return vcs.Issue{Numero: 7, URL: "https://github.com/o/r/issues/7"}, nil
}

func TestCriarIssueConcepcaoSemPRSemEfeito(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	res, err := CriarIssueConcepcao(dir, specTeste, iniTeste, &adFake{semPR: true}, time.Now())
	if err != nil || res.Feito || res.Motivo == "" {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)); err == nil {
		t.Fatal("evidência não deveria existir")
	}
}

func TestCriarIssueConcepcaoCriaEReaproveita(t *testing.T) {
	dir := projetoComSpec(t, specOk)
	ad := &adFake{}
	res, err := CriarIssueConcepcao(dir, specTeste, iniTeste, ad, time.Now())
	if err != nil || !res.Feito || ad.criadas != 1 || res.Evidencia.ID != "7" {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	cfg, _, _ := config.LoadForgeConfig(dir)
	if _, pend := loadEvidence(dir, iniTeste+"/"+CardEvidenceFile, "900-teste", "900-teste", cfg, time.Now()); len(pend) != 0 {
		t.Fatalf("evidência gravada não passa em loadEvidence: %v", pend)
	}
	res, err = CriarIssueConcepcao(dir, specTeste, iniTeste, &adFake{achada: &vcs.Issue{Numero: 9, URL: "https://github.com/o/r/issues/9"}}, time.Now())
	if err != nil || !res.Reaprovei || res.Evidencia.ID != "7" {
		t.Fatalf("err=%v res=%+v", err, res)
	}
}
