package harness

import (
	"fmt"
	"regexp"
	"strings"
)

// Valores aceitos em Parecer.Veredito e Retorno.Classificacao.
const (
	VereditoConforme = "conforme"
	VereditoViolacao = "violacao"
	VereditoDuvida   = "duvida"
)

// Parecer é a passagem Reviewer -> Go: o veredito semântico da Governança (Camada B).
type Parecer struct {
	Veredito       string          `json:"veredito"`
	Achados        []Retorno       `json:"achados"`
	CandidateSkill *CandidateSkill `json:"candidate_skill,omitempty"`
	Resumo         string          `json:"resumo"`
	// Proveniência (106, task 1.1): quem grava é o Go, a partir do HEAD da branch, nunca o agente.
	HeadSHA string `json:"head_sha,omitempty"`
	Ponto   string `json:"ponto,omitempty"`
	Modo    string `json:"modo,omitempty"`
	Feature string `json:"feature,omitempty"`
}

// Pontos de revisão aceitos em Parecer.Ponto.
const (
	PontoSpec    = "spec"
	PontoEntrega = "entrega"
)

var reSHA = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

// ComProveniencia devolve o parecer carimbado com o commit revisado, o ponto, o modo e a feature.
// Sem SHA válido ou com ponto fora do enum, recusa: veredito sem SHA não vale.
func (v Parecer) ComProveniencia(sha, ponto, modo, feature string) (Parecer, error) {
	v.HeadSHA, v.Ponto, v.Feature = strings.ToLower(strings.TrimSpace(sha)), ponto, feature
	if strings.TrimSpace(v.Modo) == "" {
		v.Modo = modo
	}
	if err := ValidarProveniencia(v); err != nil {
		return Parecer{}, err
	}
	return v, nil
}

// ValidarProveniencia exige head_sha hexadecimal e ponto spec ou entrega.
func ValidarProveniencia(v Parecer) error {
	if !reSHA.MatchString(v.HeadSHA) {
		return fmt.Errorf("contrato %s: veredito sem head_sha válido (recebido %q): recusado", ContratoParecer, v.HeadSHA)
	}
	if v.Ponto != PontoSpec && v.Ponto != PontoEntrega {
		return fmt.Errorf("contrato %s: ponto inválido %q: deve ser %q ou %q", ContratoParecer, v.Ponto, PontoSpec, PontoEntrega)
	}
	return nil
}

// Retorno é um achado individual do Parecer, devolvido ao Builder.
type Retorno struct {
	Verificacao    string `json:"verificacao"`
	Regra          string `json:"regra"`
	RegraDescricao string `json:"regra_descricao,omitempty"`
	Classificacao  string `json:"classificacao"`
	Severidade     string `json:"severidade"`
	Arquivo        string `json:"arquivo"`
	Linha          int    `json:"linha"`
	Evidencia      string `json:"evidencia"`
	Explicacao     string `json:"explicacao"`
	Correcao       string `json:"correcao"`
}

// CandidateSkill é uma proposta de auto-skill identificada durante a revisão.
type CandidateSkill struct {
	Name         string `json:"name"`
	Reason       string `json:"reason"`
	ProposedPath string `json:"proposed_path"`
}

// ValidateParecer valida as regras semânticas do Parecer.
func ValidateParecer(v Parecer) error {
	switch v.Veredito {
	case VereditoConforme, VereditoViolacao, VereditoDuvida:
	default:
		return fmt.Errorf("contrato %s: veredito inválido %q: deve ser %q, %q ou %q", ContratoParecer, v.Veredito, VereditoConforme, VereditoViolacao, VereditoDuvida)
	}
	if strings.TrimSpace(v.Resumo) == "" {
		return campoVazio(ContratoParecer, "resumo")
	}
	for idx, f := range v.Achados {
		if err := validateRetorno(idx, f); err != nil {
			return err
		}
	}
	if v.CandidateSkill != nil {
		c := v.CandidateSkill
		if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Reason) == "" || strings.TrimSpace(c.ProposedPath) == "" {
			return fmt.Errorf("contrato %s: candidate_skill incompleta: 'name', 'reason' e 'proposed_path' são obrigatórios quando presente", ContratoParecer)
		}
	}
	return nil
}

func validateRetorno(idx int, f Retorno) error {
	switch f.Classificacao {
	case VereditoViolacao, VereditoDuvida:
	default:
		return fmt.Errorf("contrato %s: achado[%d]: classificacao inválida %q (esperado %q ou %q)", ContratoParecer, idx, f.Classificacao, VereditoViolacao, VereditoDuvida)
	}
	if f.Classificacao != VereditoViolacao {
		return nil
	}
	if strings.TrimSpace(f.Regra) == "" {
		return fmt.Errorf("contrato %s: achado[%d]: violação sem 'regra' preenchida", ContratoParecer, idx)
	}
	if strings.TrimSpace(f.Arquivo) == "" {
		return fmt.Errorf("contrato %s: achado[%d]: violação sem 'arquivo' preenchido", ContratoParecer, idx)
	}
	if f.Linha <= 0 {
		return fmt.Errorf("contrato %s: achado[%d]: violação exige 'linha' > 0 (recebido %d)", ContratoParecer, idx, f.Linha)
	}
	if strings.TrimSpace(f.Evidencia) == "" {
		return fmt.Errorf("contrato %s: achado[%d]: violação sem 'evidencia' explícita (violação sem evidência é inválida)", ContratoParecer, idx)
	}
	return nil
}

// DecodeParecerSchema lê o Parecer estritamente, sem as regras semânticas.
func DecodeParecerSchema(data []byte) (Parecer, error) {
	var v Parecer
	if err := decodeStrict(ContratoParecer, data, &v); err != nil {
		return Parecer{}, err
	}
	if v.Achados == nil {
		v.Achados = []Retorno{}
	}
	return v, nil
}

// DecodeParecer lê estritamente e valida o Parecer.
func DecodeParecer(data []byte) (Parecer, error) {
	v, err := DecodeParecerSchema(data)
	if err != nil {
		return Parecer{}, err
	}
	if err := ValidateParecer(v); err != nil {
		return Parecer{}, err
	}
	return v, nil
}
