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
// no binário não é comparável. Reusa config.CompareVersions (ignora sufixo de pré-lançamento).
func CompararVersao(projeto, binario string) (VersaoStatus, string) {
	p := strings.TrimPrefix(notify.Sanitize(strings.TrimSpace(projeto), 40, false), "v")
	b := strings.TrimPrefix(notify.Sanitize(strings.TrimSpace(binario), 40, false), "v")
	if p == "" || b == "" || b == "dev" {
		return VersaoNaoComparavel, "Versão não comparável com o binário (projeto: '" + p + "', binário: '" + b + "')."
	}
	switch config.CompareVersions(p, b) {
	case 0:
		return VersaoOK, "Projeto em sincronia com o binário: v" + p
	case -1:
		return VersaoProjetoAntigo, "Projeto na v" + p + ", binário na v" + b + ". Rode 'forge update' para atualizar o projeto."
	default:
		return VersaoBinarioAntigo, "Binário na v" + b + " é mais antigo que o projeto (v" + p + "). Atualize o forge (npm i -g @forge-sdd/cli ou baixe a release mais nova)."
	}
}
