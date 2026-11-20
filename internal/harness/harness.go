// Package harness define os contratos de passagem entre as etapas do pipeline
// (Discovery -> Builder -> Reviewer). Toda leitura é estrita e toda escrita passa
// por uma função Validate*, para que skill e binário não divirjam em silêncio.
package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Nomes dos contratos, citados nas mensagens de erro.
const (
	ContratoSpecParaBuild  = "SpecParaBuild"
	ContratoResultadoBuild = "ResultadoBuild"
	ContratoParecer        = "Parecer"
)

// decodeStrict decodifica data em out recusando campos desconhecidos e conteúdo extra.
func decodeStrict(contrato string, data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("contrato %s: JSON inválido: %w", contrato, err)
	}
	if dec.More() {
		return fmt.Errorf("contrato %s: conteúdo extra após o objeto JSON", contrato)
	}
	return nil
}

func campoVazio(contrato, campo string) error {
	return fmt.Errorf("contrato %s: campo %q é obrigatório e está vazio", contrato, campo)
}
