package vcs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Modos reúne os campos de configuração de modo da tabela `Ferramentas e Integrações` da
// Constituição (rule `.agents/rules/sdd-modos-de-projeto.md`). Todos são lidos da branch base:
// a branch da feature não os sobrepõe (ADR-0022).
type Modos struct {
	// VCS é o sistema de versionamento: github, azure-devops ou nenhum.
	VCS string
	// Atestado é o modo do atestado: desligado, local (padrão) ou vcs (adiado).
	Atestado string
	// RevisaoHumana é com-humano (padrão) ou sem-humano.
	RevisaoHumana string
	// WIPHumano é o limite de itens aguardando o humano; 0 significa desligado.
	WIPHumano int
	// ModulosPermitidos é `nenhum` (padrão) ou a lista de módulos autorizados.
	ModulosPermitidos string
	// InicioAutomatico é desligado (padrão) ou local.
	InicioAutomatico string
}

// Valores aceitos. O leitor recusa o que não está aqui (SEG-01): um typo na Constituição vira
// erro na borda, não um comportamento silencioso.
const (
	AtestadoDesligado = "desligado"
	AtestadoLocal     = "local"
	AtestadoVCS       = "vcs"

	RevisaoComHumano = "com-humano"
	RevisaoSemHumano = "sem-humano"

	InicioDesligado = "desligado"
	InicioLocal     = "local"

	ModulosNenhum = "nenhum"

	// WIPHumanoPadrao é o limite padrão de itens aguardando o humano.
	WIPHumanoPadrao = 3
)

var (
	// campoConstituicaoRe casa uma linha de tabela `| Campo | Valor |`.
	campoConstituicaoRe = regexp.MustCompile(`(?m)^\|\s*([^|]+?)\s*\|\s*([^|]*?)\s*\|\s*$`)
)

// modosPadrao são os valores de um projeto que ainda não preencheu os campos novos.
func modosPadrao() Modos {
	return Modos{
		VCS:               "github",
		Atestado:          AtestadoLocal,
		RevisaoHumana:     RevisaoComHumano,
		WIPHumano:         WIPHumanoPadrao,
		ModulosPermitidos: ModulosNenhum,
		InicioAutomatico:  InicioDesligado,
	}
}

// camposModos mapeia o rótulo da tabela ao ponteiro do campo.
func (m *Modos) aplicar(campo, valor string) error {
	v := strings.TrimSpace(valor)
	switch campo {
	case "VCS / Work Item System":
		m.VCS = v
	case "Modo do atestado":
		m.Atestado = v
	case "Modo de revisão humana":
		m.RevisaoHumana = v
	case "WIP humano":
		m.WIPHumano = parseWIP(v)
	case "Módulos permitidos":
		m.ModulosPermitidos = v
	case "Início automático do Builder":
		m.InicioAutomatico = v
	}
	return nil
}

// parseWIP aceita um inteiro positivo, `desligado` (0) ou vazio (padrão).
func parseWIP(v string) int {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "desligado":
		if strings.EqualFold(strings.TrimSpace(v), "desligado") {
			return 0
		}
		return WIPHumanoPadrao
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 0 {
		return WIPHumanoPadrao
	}
	return n
}

// Validar confere os valores lidos. Devolve todos os problemas de uma vez.
func (m Modos) Validar() error {
	var problemas []string
	switch m.VCS {
	case "github", "nenhum", "azure-devops":
	default:
		problemas = append(problemas, fmt.Sprintf("VCS %q desconhecido (use github, azure-devops ou nenhum)", m.VCS))
	}
	switch m.Atestado {
	case AtestadoDesligado, AtestadoLocal, AtestadoVCS:
	default:
		problemas = append(problemas, fmt.Sprintf("Modo do atestado %q inválido (use desligado, local ou vcs)", m.Atestado))
	}
	// `nenhum` só admite desligado ou local: sem VCS não há prova externa para o modo vcs.
	if m.VCS == "nenhum" && m.Atestado == AtestadoVCS {
		problemas = append(problemas, "Modo do atestado \"vcs\" não combina com VCS \"nenhum\"")
	}
	switch m.RevisaoHumana {
	case RevisaoComHumano, RevisaoSemHumano:
	default:
		problemas = append(problemas, fmt.Sprintf("Modo de revisão humana %q inválido (use com-humano ou sem-humano)", m.RevisaoHumana))
	}
	if m.WIPHumano < 0 {
		problemas = append(problemas, "WIP humano não pode ser negativo")
	}
	if strings.TrimSpace(m.ModulosPermitidos) == "" {
		problemas = append(problemas, "Módulos permitidos vazio (use nenhum ou uma lista)")
	}
	switch m.InicioAutomatico {
	case InicioDesligado, InicioLocal:
	default:
		problemas = append(problemas, fmt.Sprintf("Início automático do Builder %q inválido (use desligado ou local)", m.InicioAutomatico))
	}
	if len(problemas) > 0 {
		return errors.New(strings.Join(problemas, "; "))
	}
	return nil
}

// ModulosLista devolve os módulos autorizados; vazio quando o valor é `nenhum`.
func (m Modos) ModulosLista() []string {
	v := strings.TrimSpace(m.ModulosPermitidos)
	if v == "" || strings.EqualFold(v, ModulosNenhum) {
		return nil
	}
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// WIPLigado diz se o limite de WIP humano está ativo.
func (m Modos) WIPLigado() bool { return m.WIPHumano > 0 }

// LeModos lê todos os campos de modo da Constituição de dir. Arquivo ausente devolve os padrões;
// campo ausente mantém o padrão do campo. Nenhum valor inválido passa silencioso: Validar() o
// recusa (SEG-01).
func LeModos(dir string) (Modos, error) {
	m := modosPadrao()
	data, err := os.ReadFile(filepath.Join(dir, "sdd", "memory", "constitution.md"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return m, nil
		}
		return m, fmt.Errorf("ler a Constituição: %w", err)
	}
	for _, sub := range campoConstituicaoRe.FindAllStringSubmatch(string(data), -1) {
		campo := strings.TrimSpace(sub[1])
		valor := sub[2]
		// Linhas de cabeçalho e separador da tabela não casam com campo conhecido.
		if err := m.aplicar(campo, valor); err != nil {
			return m, err
		}
	}
	if err := m.Validar(); err != nil {
		return m, fmt.Errorf("modos inválidos em sdd/memory/constitution.md: %w", err)
	}
	return m, nil
}

// LeModosDaBranchBase lê os modos a partir do conteúdo da Constituição da branch base. É o
// caminho que a sessão usa: a Constituição da branch da feature não sobrepõe a base (ADR-0022).
// Sem conteúdo (arquivo ausente na base), devolve os padrões.
func LeModosDaBranchBase(conteudo string) (Modos, error) {
	m := modosPadrao()
	if strings.TrimSpace(conteudo) == "" {
		return m, nil
	}
	for _, sub := range campoConstituicaoRe.FindAllStringSubmatch(conteudo, -1) {
		_ = m.aplicar(strings.TrimSpace(sub[1]), sub[2])
	}
	if err := m.Validar(); err != nil {
		return m, fmt.Errorf("modos inválidos na Constitution da branch base: %w", err)
	}
	return m, nil
}