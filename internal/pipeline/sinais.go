package pipeline

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sinais determinísticos lidos do status.md da iniciativa (RN-4). Eles dizem em que ponto do
// ciclo a spec está, para a conferência do Reviewer só rodar quando o sinal do ponto se repetir.
const (
	// SinalSpecParaRevisao é gravado pela concepção quando a spec está pronta na branch.
	SinalSpecParaRevisao = "spec-para-revisao"
	// SinalProntoParaRevisao é gravado por feature quando a fila dela está vazia e a checagem
	// local passou.
	SinalProntoParaRevisao = "pronto-para-revisao"
	// SinalNenhum é o estado inicial: nada a conferir.
	SinalNenhum = ""
)

// MaxRodadasPorFeature é o teto padrão de rodadas de conferência por feature (RN-5).
const MaxRodadasPorFeature = 3

// sinalValidos são os sinais aceitos no arquivo; qualquer outro é recusado (SEG-01).
var sinalValidos = map[string]bool{
	SinalNenhum:            true,
	SinalSpecParaRevisao:   true,
	SinalProntoParaRevisao: true,
}

// Sinal é o marco gravado no status.md, com o carimbo que permite saber se ele se repetiu.
type Sinal struct {
	Ponto string // spec-para-revisao ou pronto-para-revisao
	Em    time.Time
}

// SinalDeSpec monta o sinal da Conferência 1.
func SinalDeSpec(em time.Time) Sinal { return Sinal{Ponto: SinalSpecParaRevisao, Em: em} }

// SinalDeFeature monta o sinal da Conferência 2, por feature.
func SinalDeFeature(em time.Time) Sinal { return Sinal{Ponto: SinalProntoParaRevisao, Em: em} }

// LinhaStatus devolve a linha do status.md para o sinal, no formato `Sinal: <ponto> <RFC3339>`.
// Sem sinal, devolve vazio: a linha some em vez de virar ruído.
func (s Sinal) LinhaStatus() string {
	if !sinalValidos[s.Ponto] || s.Ponto == SinalNenhum {
		return ""
	}
	return fmt.Sprintf("Sinal: %s %s", s.Ponto, s.Em.UTC().Format(time.RFC3339))
}

// ParseLinhaSinal lê a linha `Sinal: <ponto> <RFC3339>` do status.md. A entrada é dado não
// confiável (SEG-01): ponto desconhecido ou data malformada devolve erro, nunca um sinal parcial.
func ParseLinhaSinal(linha string) (Sinal, error) {
	resto, ok := strings.CutPrefix(strings.TrimSpace(linha), "Sinal:")
	if !ok {
		return Sinal{}, fmt.Errorf("linha de sinal ausente")
	}
	campos := strings.Fields(resto)
	if len(campos) != 2 {
		return Sinal{}, fmt.Errorf("linha de sinal malformada: esperado \"Sinal: <ponto> <RFC3339>\"")
	}
	if !sinalValidos[campos[0]] || campos[0] == SinalNenhum {
		return Sinal{}, fmt.Errorf("sinal %q desconhecido", campos[0])
	}
	em, err := time.Parse(time.RFC3339, campos[1])
	if err != nil {
		return Sinal{}, fmt.Errorf("data do sinal inválida: %w", err)
	}
	return Sinal{Ponto: campos[0], Em: em}, nil
}

// NomeiaCiclo diz qual conferência o sinal dispara. Sem sinal, não dispara nada.
func NomeiaCiclo(ponto string) string {
	switch ponto {
	case SinalSpecParaRevisao:
		return "Conferência 1 (spec)"
	case SinalProntoParaRevisao:
		return "Conferência 2 (feature)"
	}
	return ""
}

// LinhasModoStatus devolve as duas linhas do status.md que registram o modo efetivo da sessão
// (modo do atestado e modo de revisão humana), lidas da Constituição (`FLX-01`, `FLX-06`).
// Valores vazios viram "indefinido" — a sessão nunca finge um modo que não leu.
func LinhasModoStatus(modoAtestado, modoRevisao string) []string {
	atestado := strings.TrimSpace(modoAtestado)
	if atestado == "" {
		atestado = "indefinido"
	}
	revisao := strings.TrimSpace(modoRevisao)
	if revisao == "" {
		revisao = "indefinido"
	}
	return []string{
		"Atestado: " + atestado,
		"Modo: " + revisao,
	}
}

// TetoRodadas é o limite de conferências de uma feature (RN-5). Configurado com valor inválido
// (zero ou negativo), cai no padrão em vez de deixar o laço solto.
type TetoRodadas struct {
	Max int
}

// NovoTetoRodadas normaliza o valor vindo da configuração.
func NovoTetoRodadas(max int) TetoRodadas {
	if max <= 0 {
		max = MaxRodadasPorFeature
	}
	return TetoRodadas{Max: max}
}

// Estourou diz se a rodada pedida passou do teto. A terceira rodada ainda roda; a quarta pergunta
// ao humano, conforme o desenho (padrão 3 passadas).
func (t TetoRodadas) Estourou(rodada int) bool {
	limite := t.Max
	if limite <= 0 {
		limite = MaxRodadasPorFeature
	}
	return rodada > limite
}

// PerguntaAoTeto devolve o texto da pergunta feita ao humano quando o teto estoura. Ele existe
// para que a decisão seja explícita, nunca um silêncio que parece aprovação.
func PerguntaAoTeto(feature string, rodada, teto int) string {
	if feature == "" {
		feature = "a feature"
	}
	return fmt.Sprintf("A %s chegou à rodada %d de conferência (teto %d). Seguir assim mesmo, ajustar ou parar?",
		feature, rodada, teto)
}

// RodadasUsadas conta quantas linhas de histórico de conferência já existem para a feature.
// Ela lê só o formato gravado por RegistroRodada, para não depender de prosa no status.md.
func RodadasUsadas(linhas []string) int {
	n := 0
	for _, l := range linhas {
		resto, ok := strings.CutPrefix(strings.TrimSpace(l), "Rodada:")
		if !ok {
			continue
		}
		rodada, _, _ := strings.Cut(strings.TrimSpace(resto), "—")
		if _, err := strconv.Atoi(strings.TrimSpace(rodada)); err == nil {
			n++
		}
	}
	return n
}

// LinhaRodada devolve a linha que registra uma rodada concluída.
func LinhaRodada(rodada int, resultado string) string {
	if rodada <= 0 {
		return ""
	}
	resultado = strings.TrimSpace(resultado)
	if resultado == "" {
		resultado = "sem resultado"
	}
	return fmt.Sprintf("Rodada: %d — %s", rodada, resultado)
}