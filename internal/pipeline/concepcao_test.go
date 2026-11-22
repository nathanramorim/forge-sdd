package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
)

func projetoConcepcao(t *testing.T, ligado bool) (string, *[]string) {
	t.Helper()
	dir := projetoComSpec(t, specOk)
	if ligado {
		cfg := config.DefaultForgeConfig()
		cfg.PMMCP.CreateOnConception = true
		if err := config.SaveForgeConfig(dir, cfg); err != nil {
			t.Fatal(err)
		}
	}
	return dir, &[]string{}
}

func runnerFake(chamadas *[]string, falha bool) func(dir, name string, args ...string) (string, error) {
	return func(dir, name string, args ...string) (string, error) {
		*chamadas = append(*chamadas, strings.Join(args, " "))
		if falha {
			return "", errors.New("gh caiu")
		}
		if len(args) > 1 && args[1] == "list" {
			return "[]", nil
		}
		return "https://github.com/o/r/issues/5\n", nil
	}
}

func TestConcepcaoCustoZeroQuandoGestorNaoConfigurado(t *testing.T) {
	dir, ch := projetoConcepcao(t, false)
	res, err := Concepcao(OpcoesConcepcao{Dir: dir, Feature: "feat-900-01-x", Run: runnerFake(ch, false)})
	if err != nil || res.Feito || len(*ch) != 0 {
		t.Fatalf("err=%v res=%+v chamadas=%v", err, res, *ch)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)); err == nil {
		t.Fatal("não deveria gravar")
	}
}

func TestConcepcaoCriaIssueEGravaCardEvidence(t *testing.T) {
	dir, ch := projetoConcepcao(t, true)
	o := OpcoesConcepcao{Dir: dir, Feature: "feat-900-01-x", Run: runnerFake(ch, false), Now: time.Now()}
	res, err := Concepcao(o)
	if err != nil || !res.Feito || res.Evidencia.ID != "5" {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	n := len(*ch)
	res, err = Concepcao(o)
	if err != nil || !res.Reaprovei || len(*ch) != n {
		t.Fatalf("segunda execução deveria reaproveitar sem chamar o gh: err=%v res=%+v", err, res)
	}
}

func TestConcepcaoFalhaExternaViraAvisoSemBloquearSpec(t *testing.T) {
	dir, ch := projetoConcepcao(t, true)
	res, err := Concepcao(OpcoesConcepcao{Dir: dir, Feature: "feat-900-01-x", Run: runnerFake(ch, true)})
	if err != nil || res.Aviso == "" || res.Feito {
		t.Fatalf("err=%v res=%+v", err, res)
	}
}

func TestConcepcaoDryRunEMCP(t *testing.T) {
	dir, ch := projetoConcepcao(t, true)
	res, err := Concepcao(OpcoesConcepcao{Dir: dir, Feature: "feat-900-01-x", DryRun: true, Run: runnerFake(ch, false)})
	if err != nil || res.Feito || len(*ch) != 0 {
		t.Fatalf("err=%v res=%+v ch=%v", err, res, *ch)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(iniTeste), CardEvidenceFile)); err == nil {
		t.Fatal("dry-run gravou")
	}
	res, err = Concepcao(OpcoesConcepcao{Dir: dir, Feature: "feat-900-01-x", Provider: "clickup", ID: "abc1", URL: "https://app.clickup.com/t/abc1"})
	if err != nil || !res.Feito || res.Evidencia.Provider != "clickup" {
		t.Fatalf("err=%v res=%+v", err, res)
	}
}

func TestValidaCamposCardRecusa(t *testing.T) {
	sup := []string{"clickup"}
	ruins := [][4]string{
		{"clickup", "1", "http://x.com/a", ""}, {"clickup", "1", "https://x.com/\x1b[0m", ""},
		{"jira", "1", "https://x.com/a", ""}, {"clickup", "a\nb", "https://x.com/a", ""},
		{"clickup", strings.Repeat("a", 90), "https://x.com/a", ""}, {"clickup", "", "https://x.com/a", ""},
	}
	for _, r := range ruins {
		if ValidaCamposCard(sup, "feat-1", r[0], r[1], r[2]) == nil {
			t.Errorf("deveria recusar %q", r)
		}
	}
	if ValidaCamposCard(sup, "../x", "", "", "") == nil {
		t.Error("feature inválida")
	}
}
