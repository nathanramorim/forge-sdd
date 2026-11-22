package harness

import (
	"strings"
	"testing"
)

func TestDecodeSpecParaBuild(t *testing.T) {
	s, err := DecodeSpecParaBuild([]byte(`{"task_id":"feat-1-01","spec_path":"s.md","branch":"feat/1","rules":["ARQ-01"]}`))
	if err != nil || s.TaskID != "feat-1-01" {
		t.Fatalf("spec válida recusada: %v %+v", err, s)
	}
	casos := map[string]struct{ entrada, cita string }{
		"campo desconhecido": {`{"task_id":"t","spec_path":"s","branch":"b","extra":1}`, "extra"},
		"sem task_id":        {`{"spec_path":"s","branch":"b"}`, "task_id"},
		"sem spec_path":      {`{"task_id":"t","branch":"b"}`, "spec_path"},
		"sem branch":         {`{"task_id":"t","spec_path":"s"}`, "branch"},
	}
	for nome, c := range casos {
		t.Run(nome, func(t *testing.T) {
			_, err := DecodeSpecParaBuild([]byte(c.entrada))
			if err == nil || !strings.Contains(err.Error(), c.cita) || !strings.Contains(err.Error(), ContratoSpecParaBuild) {
				t.Fatalf("erro esperado citando %q e o contrato, obtido: %v", c.cita, err)
			}
		})
	}
}

func TestDecodeResultadoBuild(t *testing.T) {
	r, err := DecodeResultadoBuild([]byte(`{"task_id":"t","branch":"b","commit":"abc","arquivos_alterados":["a.go"],"criterio_atendido":true}`))
	if err != nil || !r.CriterioAtendido {
		t.Fatalf("resultado válido recusado: %v %+v", err, r)
	}
	casos := map[string]struct{ entrada, cita string }{
		"campo desconhecido": {`{"task_id":"t","branch":"b","commit":"c","tokens":3}`, "tokens"},
		"sem task_id":        {`{"branch":"b","commit":"c"}`, "task_id"},
		"sem branch":         {`{"task_id":"t","commit":"c"}`, "branch"},
		"sem commit":         {`{"task_id":"t","branch":"b"}`, "commit"},
	}
	for nome, c := range casos {
		t.Run(nome, func(t *testing.T) {
			_, err := DecodeResultadoBuild([]byte(c.entrada))
			if err == nil || !strings.Contains(err.Error(), c.cita) || !strings.Contains(err.Error(), ContratoResultadoBuild) {
				t.Fatalf("erro esperado citando %q e o contrato, obtido: %v", c.cita, err)
			}
		})
	}
}
