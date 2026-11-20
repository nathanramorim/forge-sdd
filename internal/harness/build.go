package harness

import "strings"

// SpecParaBuild é a passagem Discovery -> Builder: a task a implementar.
type SpecParaBuild struct {
	TaskID   string   `json:"task_id"`
	SpecPath string   `json:"spec_path"`
	Branch   string   `json:"branch"`
	Rules    []string `json:"rules"`
}

// ResultadoBuild é a passagem Builder -> Reviewer: o que o Builder entregou.
type ResultadoBuild struct {
	TaskID            string   `json:"task_id"`
	Branch            string   `json:"branch"`
	Commit            string   `json:"commit"`
	ArquivosAlterados []string `json:"arquivos_alterados"`
	CriterioAtendido  bool     `json:"criterio_atendido"`
}

// ValidateSpecParaBuild valida os campos obrigatórios da passagem Discovery -> Builder.
func ValidateSpecParaBuild(s SpecParaBuild) error {
	if strings.TrimSpace(s.TaskID) == "" {
		return campoVazio(ContratoSpecParaBuild, "task_id")
	}
	if strings.TrimSpace(s.SpecPath) == "" {
		return campoVazio(ContratoSpecParaBuild, "spec_path")
	}
	if strings.TrimSpace(s.Branch) == "" {
		return campoVazio(ContratoSpecParaBuild, "branch")
	}
	return nil
}

// DecodeSpecParaBuild lê estritamente e valida a passagem Discovery -> Builder.
func DecodeSpecParaBuild(data []byte) (SpecParaBuild, error) {
	var s SpecParaBuild
	if err := decodeStrict(ContratoSpecParaBuild, data, &s); err != nil {
		return SpecParaBuild{}, err
	}
	return s, ValidateSpecParaBuild(s)
}

// ValidateResultadoBuild valida os campos obrigatórios da passagem Builder -> Reviewer.
func ValidateResultadoBuild(r ResultadoBuild) error {
	if strings.TrimSpace(r.TaskID) == "" {
		return campoVazio(ContratoResultadoBuild, "task_id")
	}
	if strings.TrimSpace(r.Branch) == "" {
		return campoVazio(ContratoResultadoBuild, "branch")
	}
	if strings.TrimSpace(r.Commit) == "" {
		return campoVazio(ContratoResultadoBuild, "commit")
	}
	return nil
}

// DecodeResultadoBuild lê estritamente e valida a passagem Builder -> Reviewer.
func DecodeResultadoBuild(data []byte) (ResultadoBuild, error) {
	var r ResultadoBuild
	if err := decodeStrict(ContratoResultadoBuild, data, &r); err != nil {
		return ResultadoBuild{}, err
	}
	return r, ValidateResultadoBuild(r)
}
