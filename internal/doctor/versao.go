package doctor

import (
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/notify"
)

// VersaoStatus é o resultado da comparação entre a versão do projeto e a do binário.
type VersaoStatus string

const (
	VersaoOK            VersaoStatus = "ok"
	VersaoProjetoAntigo VersaoStatus = "projeto-antigo"
	VersaoBinarioAntigo VersaoStatus = "binario-antigo"
	VersaoNaoComparavel VersaoStatus = "nao-comparavel"
)

// CompararVersao compara a versão registrada no projeto com a do binário em execução.
// Função pura: valida e sanitiza os dois valores sozinha (SEG-01). Versão vazia ou "dev"
// no binário não é comparável. Usa versaoMenor (mesma regra do doctor --harness), com pré-lançamento anterior à versão final.
func CompararVersao(projeto, binario string) (VersaoStatus, string) {
	p := strings.TrimPrefix(notify.Sanitize(strings.TrimSpace(projeto), 40, false), "v")
	b := strings.TrimPrefix(notify.Sanitize(strings.TrimSpace(binario), 40, false), "v")
	if p == "" || b == "" || b == "dev" {
		return VersaoNaoComparavel, "Versão não comparável com o binário (projeto: '" + p + "', binário: '" + b + "')."
	}
	switch {
	case !versaoMenor(p, b) && !versaoMenor(b, p):
		return VersaoOK, "Projeto em sincronia com o binário: v" + p
	case versaoMenor(p, b):
		return VersaoProjetoAntigo, "Projeto na v" + p + ", binário na v" + b + ". Rode 'forge update' para atualizar o projeto."
	default:
		return VersaoBinarioAntigo, "Binário na v" + b + " é mais antigo que o projeto (v" + p + "). Atualize o forge (npm i -g @forge-sdd/cli ou baixe a release mais nova)."
	}
}

// versaoMenor diz se a < b, tratando um sufixo de pré-lançamento (-beta) como anterior à versão
// final; config.CompareVersions ignora o sufixo e não serve sozinha. Ponto único de comparação
// do doctor e do doctor --harness.
func versaoMenor(a, b string) bool {
	if c := config.CompareVersions(a, b); c != 0 {
		return c < 0
	}
	preA := strings.Contains(a, "-")
	preB := strings.Contains(b, "-")
	if preA != preB {
		return preA
	}
	return preA && a < b
}
