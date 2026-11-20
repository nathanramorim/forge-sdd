// Package deprecation concentra o catálogo de comandos depreciados do forge e o
// registro CSV de uso (feat-105-07, ADR-0021). Depreciar outro comando = acrescentar
// uma entrada em Catalogo; o aviso e o registro passam todos pelo mesmo ponto.
//
// Dependências: importa internal/scaffold só para WriteFileSafe (SEG-02) e
// internal/notify para Sanitize; nenhum deles importa este pacote, então o sentido
// (cmd -> deprecation -> scaffold/notify) é preservado e não há ciclo.
package deprecation

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/notify"
	"github.com/forge-sdd/cli/internal/scaffold"
)

// MaxBytes é o tamanho máximo do CSV (256 KiB, ~2 mil linhas). Acima disso não grava.
const MaxBytes = 256 * 1024

// RelPath é o caminho do CSV dentro do projeto.
var RelPath = filepath.Join("sdd", "memory", "logs", "depreciados.csv")

// ErrLimite indica que o CSV atingiu MaxBytes e a linha não foi gravada.
var ErrLimite = errors.New("registro de depreciados acima do limite de 256 KiB; nova linha não gravada")

// Entrada descreve um comando depreciado.
type Entrada struct {
	Comando    string // caminho sem o binário raiz, ex.: "queue stations"
	Substituto string
	Desde      string
}

// Catalogo lista os comandos depreciados.
var Catalogo = []Entrada{
	{"queue stations", "`forge queue` e `forge status`", "ADR-0021"},
	{"queue handoff", "`forge queue enqueue` e `forge dispatch`", "ADR-0021"},
	{"attach", "`forge status` (as 3 sessões: Discovery, Builder, Reviewer)", "ADR-0021"},
}

// AliasesDeTexto documenta os aliases de barra depreciados. Não passam pelo Go,
// logo não geram linha de uso; constam só para a revisão do dono.
var AliasesDeTexto = []Entrada{
	{"/implement", "/implementar", "ADR-0021"},
	{"/proxima-feature", "/implementar", "ADR-0021"},
}

// Buscar devolve a entrada do catálogo para o caminho de comando (sem o binário raiz).
func Buscar(caminho string) (Entrada, bool) {
	for _, e := range Catalogo {
		if e.Comando == caminho {
			return e, true
		}
	}
	return Entrada{}, false
}

// agentes mapeia variável de ambiente para rótulo fixo. Só entradas comprovadas;
// o valor da variável nunca é gravado. Gemini CLI, Copilot e Codex ficam em
// "desconhecido" até a variável ser comprovada.
var agentes = []struct{ Var, Rotulo string }{
	{"CLAUDECODE", "claude-code"},
}

// Agente devolve o rótulo fixo do agente detectado por variável de ambiente.
func Agente(getenv func(string) string) string {
	for _, a := range agentes {
		if getenv(a.Var) != "" {
			return a.Rotulo
		}
	}
	return "desconhecido"
}

var nomeFlag = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)

// Contexto monta "subcomando --flag1 --flag2" só com nomes; valores e argumentos
// posicionais nunca entram. Nome fora do padrão é descartado (validação na borda).
func Contexto(subcomando string, flags []string) string {
	partes := []string{}
	if nomeFlag.MatchString(subcomando) {
		partes = append(partes, subcomando)
	}
	for _, f := range flags {
		if nomeFlag.MatchString(f) {
			partes = append(partes, "--"+f)
		}
	}
	return strings.Join(partes, " ")
}

// Celula sanitiza uma célula (sem quebra de linha nem ANSI) e neutraliza injeção
// de fórmula de planilha prefixando `'` quando começa com = + - @.
func Celula(s string) string {
	s = strings.TrimSpace(notify.Sanitize(s, 200, false))
	if s != "" && strings.ContainsRune("=+-@", rune(s[0])) {
		s = "'" + s
	}
	return s
}

// Registro é uma linha do CSV.
type Registro struct {
	Comando, Contexto, Agente, Substituto, Desde, VersaoForge string
	Quando                                                    time.Time
}

var cabecalho = []string{"data_utc", "comando", "contexto", "agente", "substituto", "desde", "versao_forge"}

// Registrar acrescenta uma linha a sdd/memory/logs/depreciados.csv em dir. Só grava
// se dir/sdd existir como diretório. Escrita atômica e recusa link simbólico
// (scaffold.WriteFileSafe). Pode perder linha em concorrência (ler+rename; aceito).
func Registrar(dir string, r Registro) error {
	if fi, err := os.Lstat(filepath.Join(dir, "sdd")); err != nil || !fi.IsDir() {
		return nil
	}
	path := filepath.Join(dir, RelPath)
	var atual []byte
	if fi, err := os.Lstat(path); err == nil {
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("registro de depreciados recusado: %s não é arquivo regular", RelPath)
		}
		if fi.Size() >= MaxBytes {
			return ErrLimite
		}
		atual, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ler %s: %w", RelPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar %s: %w", RelPath, err)
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if len(atual) == 0 {
		_ = w.Write(cabecalho)
	}
	_ = w.Write([]string{
		r.Quando.UTC().Format(time.RFC3339), Celula(r.Comando), Celula(r.Contexto), Celula(r.Agente),
		Celula(r.Substituto), Celula(r.Desde), Celula(r.VersaoForge),
	})
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	novo := append(atual, buf.Bytes()...)
	if len(novo) > MaxBytes {
		return ErrLimite
	}
	if err := scaffold.WriteFileSafe(dir, path, novo); err != nil {
		return fmt.Errorf("gravar %s: %w", RelPath, err)
	}
	return nil
}
