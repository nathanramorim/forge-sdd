package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/eol"
)

var blocoJSON = regexp.MustCompile("(?s)```json\\n(.*?)```")

// ExemploParecerDaSkill extrai o primeiro bloco ```json do contrato do parecer na skill de
// Governança do projeto em dir. Aceita o arquivo com CRLF (checkout no Windows).
func ExemploParecerDaSkill(dir string) ([]byte, error) {
	path := filepath.Join(dir, ".agents", "skills", "sdd-governanca", "references", "veredito.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo exemplo da skill %s: %w", path, err)
	}
	m := blocoJSON.FindSubmatch(eol.LF(data))
	if m == nil {
		return nil, fmt.Errorf("bloco ```json não encontrado em %s", path)
	}
	return m[1], nil
}

// jsonTags retorna as chaves JSON declaradas pela struct do tipo informado.
func jsonTags(tp reflect.Type) []string {
	var tags []string
	for i := 0; i < tp.NumField(); i++ {
		tags = append(tags, strings.Split(tp.Field(i).Tag.Get("json"), ",")[0])
	}
	return tags
}

// VerificarContratoParecer confere se o exemplo da skill de Governança em dir é aceito pelo
// leitor estrito do Parecer e traz todos os campos do contrato.
func VerificarContratoParecer(dir string) error {
	exemplo, err := ExemploParecerDaSkill(dir)
	if err != nil {
		return err
	}
	v, err := DecodeParecerSchema(exemplo)
	if err != nil {
		return fmt.Errorf("o exemplo da skill não passa no contrato %s: %w", ContratoParecer, err)
	}
	if len(v.Achados) == 0 {
		return fmt.Errorf("o exemplo da skill deve trazer ao menos um achado")
	}
	campos := jsonTags(reflect.TypeOf(Parecer{}))
	achado := jsonTags(reflect.TypeOf(Retorno{}))
	raiz, achados := chavesDoExemplo(exemplo)
	var faltam []string
	for _, c := range campos {
		if !raiz[c] {
			faltam = append(faltam, c)
		}
	}
	for _, c := range achado {
		if !achados[c] {
			faltam = append(faltam, "achados[0]."+c)
		}
	}
	if len(faltam) > 0 {
		return fmt.Errorf("campos do contrato %s ausentes no exemplo da skill: %s", ContratoParecer, strings.Join(faltam, ", "))
	}
	return nil
}

// chavesDoExemplo devolve as chaves da raiz e do primeiro achado do exemplo.
func chavesDoExemplo(exemplo []byte) (raiz, achado map[string]bool) {
	var obj struct {
		Achados []map[string]any `json:"achados"`
	}
	var bruto map[string]any
	raiz, achado = map[string]bool{}, map[string]bool{}
	if json.Unmarshal(exemplo, &bruto) == nil {
		for k := range bruto {
			raiz[k] = true
		}
	}
	if json.Unmarshal(exemplo, &obj) == nil && len(obj.Achados) > 0 {
		for k := range obj.Achados[0] {
			achado[k] = true
		}
	}
	return raiz, achado
}
