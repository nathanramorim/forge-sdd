package pipeline

import (
	"os"
	"testing"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/harness"
)

const shaDeTeste = "0123456789abcdef0123456789abcdef01234567"

// Os testes de fluxo não têm repositório: o HEAD é injetado; os testes de proveniência abaixo
// trocam o leitor para provar a gravação e a recusa.
func TestMain(m *testing.M) {
	lerHEAD = func(string) (string, error) { return shaDeTeste, nil }
	os.Exit(m.Run())
}

func TestPersistVerdictGravaProveniencia(t *testing.T) {
	v := harness.Parecer{Veredito: harness.VereditoConforme, Resumo: "ok", Achados: []harness.Retorno{}}
	got, err := persistVerdict(t.TempDir(), bus.Job{Feature: "feat-x"}, harness.PontoSpec, v)
	if err != nil {
		t.Fatal(err)
	}
	if got.HeadSHA != shaDeTeste || got.Ponto != harness.PontoSpec || got.Feature != "feat-x" || got.Modo == "" {
		t.Errorf("proveniência incompleta: %+v", got)
	}
}

func TestPersistVerdictSemHEADRecusado(t *testing.T) {
	orig := lerHEAD
	t.Cleanup(func() { lerHEAD = orig })
	lerHEAD = func(string) (string, error) { return "", os.ErrNotExist }
	v := harness.Parecer{Veredito: harness.VereditoConforme, Resumo: "ok"}
	if _, err := persistVerdict(t.TempDir(), bus.Job{}, harness.PontoEntrega, v); err == nil {
		t.Fatal("esperava recusa do veredito sem SHA")
	}
}
