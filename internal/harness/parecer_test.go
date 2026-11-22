package harness

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// exemploDaSkill extrai o primeiro bloco ```json do contrato do parecer na skill de Governança.
func exemploDaSkill(t *testing.T) []byte {
	t.Helper()
	exemplo, err := ExemploParecerDaSkill(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return exemplo
}

func TestContratoVeredito(t *testing.T) {
	exemplo := exemploDaSkill(t)
	v, err := DecodeParecerSchema(exemplo)
	if err != nil {
		t.Fatalf("o exemplo da skill não passa no contrato Parecer: %v", err)
	}
	if len(v.Achados) == 0 {
		t.Fatal("o exemplo da skill deve trazer ao menos um achado")
	}
	if strings.TrimSpace(v.Achados[0].RegraDescricao) == "" {
		t.Error("o exemplo da skill deve preencher regra_descricao")
	}

	var bruto struct {
		Achados []map[string]json.RawMessage `json:"achados"`
	}
	var raiz map[string]json.RawMessage
	if err := json.Unmarshal(exemplo, &raiz); err != nil {
		t.Fatalf("exemplo da skill: %v", err)
	}
	if err := json.Unmarshal(exemplo, &bruto); err != nil {
		t.Fatalf("exemplo da skill: %v", err)
	}
	for _, campo := range jsonTags(reflect.TypeOf(Parecer{})) {
		if _, ok := raiz[campo]; !ok {
			t.Errorf("campo %q do contrato Parecer ausente no exemplo da skill", campo)
		}
	}
	for _, campo := range jsonTags(reflect.TypeOf(Retorno{})) {
		if _, ok := bruto.Achados[0][campo]; !ok {
			t.Errorf("campo %q do contrato Retorno ausente no exemplo da skill", campo)
		}
	}
}

func TestContratoRecusaCampoDesconhecido(t *testing.T) {
	casos := map[string]string{
		"raiz":   `{"veredito":"conforme","achados":[],"resumo":"ok","nota_extra":"x"}`,
		"achado": `{"veredito":"duvida","achados":[{"classificacao":"duvida","campo_novo":1}],"resumo":"ok"}`,
	}
	for nome, entrada := range casos {
		t.Run(nome, func(t *testing.T) {
			_, err := DecodeParecer([]byte(entrada))
			if err == nil {
				t.Fatal("esperava recusa de campo desconhecido")
			}
			if !strings.Contains(err.Error(), ContratoParecer) {
				t.Errorf("erro não cita o contrato: %v", err)
			}
			if !strings.Contains(err.Error(), "nota_extra") && !strings.Contains(err.Error(), "campo_novo") {
				t.Errorf("erro não cita o campo: %v", err)
			}
		})
	}
}

func TestDecodeParecerValido(t *testing.T) {
	v, err := DecodeParecer([]byte(`{"veredito":"violacao","achados":[{"verificacao":"S-01","regra":"ARQ-01","regra_descricao":"texto","classificacao":"violacao","severidade":"alta","arquivo":"a.go","linha":3,"evidencia":"x","explicacao":"e","correcao":"c"}],"resumo":"r"}`))
	if err != nil {
		t.Fatalf("parecer válido recusado: %v", err)
	}
	if v.Achados[0].RegraDescricao != "texto" {
		t.Errorf("regra_descricao perdida: %+v", v.Achados[0])
	}
}

func TestDecodeParecerSemAchadosNormaliza(t *testing.T) {
	v, err := DecodeParecer([]byte(`{"veredito":"conforme","resumo":"ok"}`))
	if err != nil {
		t.Fatalf("parecer sem achados recusado: %v", err)
	}
	if v.Achados == nil {
		t.Error("achados nulo deveria virar lista vazia")
	}
}

func TestDecodeParecerInvalido(t *testing.T) {
	casos := map[string]string{
		"veredito fora do enum":  `{"veredito":"talvez","resumo":"r"}`,
		"resumo vazio":           `{"veredito":"conforme","resumo":"  "}`,
		"violacao sem evidencia": `{"veredito":"violacao","achados":[{"classificacao":"violacao","regra":"ARQ-01","arquivo":"a.go","linha":1}],"resumo":"r"}`,
		"violacao linha zero":    `{"veredito":"violacao","achados":[{"classificacao":"violacao","regra":"ARQ-01","arquivo":"a.go","evidencia":"x"}],"resumo":"r"}`,
		"classificacao invalida": `{"veredito":"duvida","achados":[{"classificacao":"ok"}],"resumo":"r"}`,
		"candidate incompleta":   `{"veredito":"conforme","candidate_skill":{"name":"x"},"resumo":"r"}`,
		"json truncado":          `{"veredito":`,
		"conteudo extra":         `{"veredito":"conforme","resumo":"r"}{}`,
	}
	for nome, entrada := range casos {
		t.Run(nome, func(t *testing.T) {
			if _, err := DecodeParecer([]byte(entrada)); err == nil {
				t.Fatal("esperava erro")
			} else if !strings.Contains(err.Error(), ContratoParecer) {
				t.Errorf("erro não cita o contrato: %v", err)
			}
		})
	}
}

func TestParecerGravaHeadSHA(t *testing.T) {
	sha := "0123456789abcdef0123456789abcdef01234567"
	v, err := DecodeParecer([]byte(`{"veredito":"conforme","resumo":"ok"}`))
	if err != nil {
		t.Fatal(err)
	}
	c, err := v.ComProveniencia(sha, PontoEntrega, "normal", "feat-106-11")
	if err != nil {
		t.Fatalf("proveniência válida recusada: %v", err)
	}
	dados, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	lido, err := DecodeParecer(dados)
	if err != nil {
		t.Fatalf("parecer carimbado não relê: %v", err)
	}
	if lido.HeadSHA != sha || lido.Ponto != PontoEntrega || lido.Modo != "normal" || lido.Feature != "feat-106-11" {
		t.Errorf("proveniência perdida: %+v", lido)
	}
}

func TestParecerSemHeadSHARecusado(t *testing.T) {
	v := Parecer{Veredito: VereditoConforme, Resumo: "ok", Achados: []Retorno{}}
	for nome, c := range map[string]struct{ sha, ponto string }{
		"sha vazio":      {"", PontoSpec},
		"sha inválido":   {"zzz", PontoSpec},
		"ponto inválido": {"0123456789abcdef", "outro"},
	} {
		t.Run(nome, func(t *testing.T) {
			if _, err := v.ComProveniencia(c.sha, c.ponto, "normal", "f"); err == nil {
				t.Fatal("esperava recusa")
			}
		})
	}
}
