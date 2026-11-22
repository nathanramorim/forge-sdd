package pipeline

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/vcs"
)

// Marcador do Gate 1 no status.md da iniciativa (feat-106-22b).
//
// No fluxo novo, o Gate 1 deixa de ser a existência de queue.json e passa a ser uma linha
// explícita no status.md da principal. O Builder só começa com o marcador presente.
const (
	// MarcadorGate1Aprovado é a linha gravada quando o humano responde [Seguir].
	MarcadorGate1Aprovado = "Gate 1: aprovado"
	// MarcadorGate1Ajustar é a linha gravada quando o humano responde [Ajustar].
	MarcadorGate1Ajustar = "Gate 1: ajustar"
	// MarcadorGate1Pendente é o estado antes da resposta humana.
	MarcadorGate1Pendente = "Gate 1: pendente"
)

// gate1LinhaRe casa a linha do marcador, com ou sem negrito em volta do rótulo.
// Depois do valor só entram separadores de frase: espaço, dois-pontos, parêntese, travessão,
// vírgula, ponto. Sem isso, "aprovado-sim-por-favor" passaria pela fronteira \b e viraria
// aprovação — exatamente o que o portão precisa impedir. Texto descritivo depois do
// separador é aceito, porque o marcador real carrega a data ("aprovado em 21/11/2026").
var gate1LinhaRe = regexp.MustCompile(`(?im)^\s*(?:\*\*)?Gate 1:(?:\*\*)?\s*(aprovado|ajustar|pendente)(?:\s*(?:\(|—|:|,|\.|em\b)+.*)?$`)

// Erros do Gate 1 como portão do Builder.
var (
	// ErrBuilderSemSpecNaPrincipal: o Builder não começa sem a spec aprovada na principal.
	ErrBuilderSemSpecNaPrincipal = errors.New("o Builder não começa sem a spec aprovada na principal")
)

// EstadoDoGate1 diz em que ponto o Gate 1 está, lido do status.md.
type EstadoDoGate1 struct {
	// Marcador é o valor encontrado (`aprovado`, `ajustar`, `pendente`) ou vazio.
	Marcador string
	// Encontrado diz se a linha existe.
	Encontrado bool
	// ModoSemHumano diz se a resposta veio automática (modo sem humano).
	ModoSemHumano bool
}

// Aprovado diz se o Gate 1 libera o Builder.
func (e EstadoDoGate1) Aprovado() bool { return e.Encontrado && e.Marcador == "aprovado" }

// LeGate1DoStatus lê o marcador do texto do status.md. Entrada externa é dado não confiável;
// o regex só reconhece os três valores, então lixo não vira aprovação.
func LeGate1DoStatus(texto string) EstadoDoGate1 {
	m := gate1LinhaRe.FindStringSubmatch(texto)
	if m == nil {
		return EstadoDoGate1{}
	}
	modoSemHumano := regexp.MustCompile(`(?im)^\s*(?:\*\*)?Modo:(?:\*\*)?\s*sem-humano\b`).MatchString(texto)
	return EstadoDoGate1{Marcador: strings.ToLower(m[1]), Encontrado: true, ModoSemHumano: modoSemHumano}
}

// LinhaDoMarcador devolve a linha a gravar no status.md para a resposta do Gate 1.
func LinhaDoMarcador(resposta string, semHumano bool, em time.Time) string {
	valor := strings.ToLower(strings.TrimSpace(resposta))
	switch valor {
	case "aprovado", "seguir":
		valor = "aprovado"
	case "ajustar":
		valor = "ajustar"
	default:
		valor = "pendente"
	}
	linha := fmt.Sprintf("Gate 1: %s", valor)
	if semHumano {
		// No modo sem humano a resposta é automática e precisa ficar registrada como tal.
		if em.IsZero() {
			em = time.Now().UTC()
		}
		linha += fmt.Sprintf(" (automático, modo sem humano) — %s", em.UTC().Format(time.RFC3339))
	} else if !em.IsZero() {
		linha += fmt.Sprintf(" — %s", em.UTC().Format(time.RFC3339))
	}
	return linha
}

// GravaMarcadorDoGate1 escreve a linha no status.md da iniciativa, substituindo a anterior se
// houver. O arquivo tem teto de 1 KB (P-21): marcador novo não pode estourar o orçamento.
func GravaMarcadorDoGate1(dir, iniciativa, resposta string, semHumano bool, em time.Time) (string, error) {
	caminho, err := layout.InitiativeStatus(dir, iniciativa)
	if err != nil {
		return "", err
	}
	atual, err := os.ReadFile(caminho)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("ler status.md: %w", err)
	}
	linha := LinhaDoMarcador(resposta, semHumano, em)
	novo := substituirLinhaGate1(string(atual), linha)
	if len(novo) > statusMaxBytes {
		return "", fmt.Errorf("status.md passaria de %d bytes com o marcador; enxugue o status antes", statusMaxBytes)
	}
	temporario := caminho + ".tmp"
	if err := os.WriteFile(temporario, []byte(novo), 0o644); err != nil {
		return "", fmt.Errorf("gravar marcador: %w", err)
	}
	if err := os.Rename(temporario, caminho); err != nil {
		_ = os.Remove(temporario)
		return "", fmt.Errorf("gravar marcador: %w", err)
	}
	return linha, nil
}

// substituirLinhaGate1 troca a linha do Gate 1 pela nova, ou a insere depois do título.
func substituirLinhaGate1(texto, linha string) string {
	linhas := strings.Split(texto, "\n")
	achou := false
	for i, l := range linhas {
		if gate1LinhaRe.MatchString(l) {
			linhas[i] = linha
			achou = true
			break
		}
	}
	if achou {
		return strings.Join(linhas, "\n")
	}
	if len(linhas) > 0 && strings.HasPrefix(linhas[0], "# ") {
		resto := append([]string{linhas[0], "", linha}, linhas[1:]...)
		return strings.Join(resto, "\n")
	}
	return linha + "\n" + strings.Join(linhas, "\n")
}

// ResultadoDoGate1 descreve o efeito no Builder.
type ResultadoDoGate1 struct {
	Liberado   bool
	SpecVoltou bool // no modo sem humano, a spec já voltou à principal
	Motivo     string
}

// LiberaOBuilder decide se o Builder pode começar a feature. Exige o Gate 1 aprovado no
// status.md; no modo sem humano, a resposta automática também precisa estar registrada.
func LiberaOBuilder(status string) (ResultadoDoGate1, error) {
	e := LeGate1DoStatus(status)
	if !e.Encontrado {
		return ResultadoDoGate1{}, fmt.Errorf("%w: o status.md não tem o marcador %q",
			ErrBuilderSemSpecNaPrincipal, MarcadorGate1Aprovado)
	}
	switch e.Marcador {
	case "aprovado":
		return ResultadoDoGate1{
			Liberado:   true,
			SpecVoltou: true, // a spec volta à principal antes do Builder (RN-1)
			Motivo:     "Gate 1 aprovado; a spec está na principal e o Builder pode começar",
		}, nil
	case "ajustar":
		return ResultadoDoGate1{}, fmt.Errorf("%w: o Gate 1 pediu ajuste (o humano respondeu [Ajustar])",
			ErrBuilderSemSpecNaPrincipal)
	default:
		return ResultadoDoGate1{}, fmt.Errorf("%w: o Gate 1 está pendente", ErrBuilderSemSpecNaPrincipal)
	}
}

// MesclaAutomaticaNoModoSemHumano faz a spec voltar à principal sem esperar o [Seguir] humano,
// mas só no modo sem humano — e só com o aviso de riscos aceito (feat-106-17).
func MesclaAutomaticaNoModoSemHumano(dec DecisaoDeModo, dir string, adapt vcs.Adaptador, m vcs.Mescla, em time.Time) (string, error) {
	if dec.Modo != ModoSemHumano {
		return "", fmt.Errorf("a mescla automática só vale no modo sem humano (modo atual %q)", dec.Modo)
	}
	if !dec.ComAvisado {
		return "", fmt.Errorf("%w: sem o aviso de riscos aceito", ErrSemHumanoSemAviso)
	}
	if adapt == nil {
		return "", ErrSpecSemAdaptador
	}
	if adapt.Nome() == "nenhum" {
		return fmt.Sprintf("VCS nenhum: a spec em %q fica pronta; o merge é local", m.Branch), nil
	}
	if err := adapt.Mesclar(dir, m); err != nil {
		return "", fmt.Errorf("mesclar a spec automaticamente: %w", err)
	}
	return "", nil
}

// AvisoDoGate1NoModoSemHumano é o texto que acompanha a resposta automática. Ele existe para a
// decisão nunca parecer humana: diz que foi o modo, e não uma pessoa, que aprovou.
func AvisoDoGate1NoModoSemHumano(em time.Time) string {
	if em.IsZero() {
		em = time.Now().UTC()
	}
	return fmt.Sprintf("Gate 1 respondido automaticamente pelo modo sem humano em %s; "+
		"nenhuma pessoa revisou a spec antes desta aprovação.", em.UTC().Format(time.RFC3339))
}

// EscreveAvisoNoOut para a interface do usuário, sem quebrar o pipeline quando out é nil.
func EscreveAvisoNoOut(out io.Writer, texto string) {
	if out == nil || texto == "" {
		return
	}
	fmt.Fprintln(out, texto)
}

// statusMaxBytes é o teto do status.md por iniciativa (P-21, GOV-03).
const statusMaxBytes = 1024

// CaminhoDoStatus devolve o status.md da iniciativa, para o chamador não repetir o path.
func CaminhoDoStatus(dir, iniciativa string) (string, error) {
	p, err := layout.InitiativeStatus(dir, iniciativa)
	if err != nil {
		return "", err
	}
	return filepath.Clean(p), nil
}
