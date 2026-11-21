package vcs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// constitutionComModos monta uma Constituição mínima com os campos de modo.
func constitutionComModos(linhas ...string) string {
	base := "## Ferramentas e Integrações\n| Campo | Valor |\n|---|---|\n"
	for _, l := range linhas {
		base += "| " + l + " |\n"
	}
	return base
}

func TestLeModosDaConstituicaoDaBranchBase(t *testing.T) {
	txt := constitutionComModos(
		"VCS / Work Item System | github",
		"Modo do atestado | local",
		"Modo de revisão humana | com-humano",
		"WIP humano | 3",
		"Módulos permitidos | nenhum",
		"Início automático do Builder | desligado",
	)
	m, err := LeModosDaBranchBase(txt)
	require.NoError(t, err)
	assert.Equal(t, Modos{
		VCS:               "github",
		Atestado:          AtestadoLocal,
		RevisaoHumana:     RevisaoComHumano,
		WIPHumano:         3,
		ModulosPermitidos: ModulosNenhum,
		InicioAutomatico:  InicioDesligado,
	}, m)
	assert.Nil(t, m.ModulosLista(), "nenhum módulo autorizado")
	assert.True(t, m.WIPLigado())

	// A partir de um arquivo real no disco.
	dir := constituicao(t, "github")
	lido, err := LeModos(dir)
	require.NoError(t, err)
	assert.Equal(t, AtestadoLocal, lido.Atestado, "campo ausente vale o padrão")
	assert.Equal(t, RevisaoComHumano, lido.RevisaoHumana)
	assert.Equal(t, WIPHumanoPadrao, lido.WIPHumano)
}

func TestAtestadoPadraoELocal(t *testing.T) {
	// Constituição sem nenhum campo de modo: o padrão é local.
	m, err := LeModosDaBranchBase(constitutionComModos("VCS / Work Item System | github"))
	require.NoError(t, err)
	assert.Equal(t, AtestadoLocal, m.Atestado)

	// Conteúdo vazio também cai no padrão.
	vazio, err := LeModosDaBranchBase("")
	require.NoError(t, err)
	assert.Equal(t, AtestadoLocal, vazio.Atestado)
	assert.Equal(t, RevisaoComHumano, vazio.RevisaoHumana)
	assert.Equal(t, ModulosNenhum, vazio.ModulosPermitidos)
	assert.Equal(t, InicioDesligado, vazio.InicioAutomatico)

	// WIP desligado vira zero e sai do laço.
	desligado, err := LeModosDaBranchBase(constitutionComModos("WIP humano | desligado"))
	require.NoError(t, err)
	assert.False(t, desligado.WIPLigado())
	assert.Equal(t, 0, desligado.WIPHumano)
}

func TestModoDeAtestadoInvalidoRecusado(t *testing.T) {
	for _, valor := range []string{"talvez", "Local", "LOCAL", "vcs-externo", "1"} {
		_, err := LeModosDaBranchBase(constitutionComModos("Modo do atestado | " + valor))
		assert.Error(t, err, "atestado %q deveria ser recusado", valor)
	}

	// VCS desconhecido, revisão inválida e início inválido também são recusados.
	_, err := LeModosDaBranchBase(constitutionComModos("VCS / Work Item System | svn"))
	assert.Error(t, err)
	_, err = LeModosDaBranchBase(constitutionComModos("Modo de revisão humana | sozinho"))
	assert.Error(t, err)
	_, err = LeModosDaBranchBase(constitutionComModos("Início automático do Builder | sempre"))
	assert.Error(t, err)

	// Combinação impossível: sem VCS não há prova externa.
	_, err = LeModosDaBranchBase(constitutionComModos(
		"VCS / Work Item System | nenhum",
		"Modo do atestado | vcs",
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "não combina")

	// A partir do arquivo, o erro cita o caminho.
	dir := t.TempDir()
	escreverConstituicao(t, dir, constitutionComModos("Modo do atestado | ilegal"))
	_, err = LeModos(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "constitution.md")
}

func TestModoNaBranchDaFeatureNaoSobrepoeABase(t *testing.T) {
	// O leitor recebe só o conteúdo da base: qualquer valor vindo da branch da feature não entra.
	base := constitutionComModos(
		"Modo do atestado | local",
		"Modo de revisão humana | com-humano",
	)
	feature := constitutionComModos(
		"Modo do atestado | desligado",
		"Modo de revisão humana | sem-humano",
	)

	daBase, err := LeModosDaBranchBase(base)
	require.NoError(t, err)
	assert.Equal(t, AtestadoLocal, daBase.Atestado)
	assert.Equal(t, RevisaoComHumano, daBase.RevisaoHumana)

	daFeature, err := LeModosDaBranchBase(feature)
	require.NoError(t, err)
	assert.Equal(t, AtestadoDesligado, daFeature.Atestado)
	assert.Equal(t, RevisaoSemHumano, daFeature.RevisaoHumana)

	// A garantia é de quem chama: LeModosDaBranchBase só olha o texto que recebe.
	assert.NotEqual(t, daBase.Atestado, daFeature.Atestado,
		"os dois textos produzem modos diferentes; é o chamador que decide qual ler")
}

func TestModulosListaEValidacao(t *testing.T) {
	m, err := LeModosDaBranchBase(constitutionComModos("Módulos permitidos | quadro, estimativa"))
	require.NoError(t, err)
	assert.Equal(t, []string{"quadro", "estimativa"}, m.ModulosLista())

	// Vazio é recusado: o campo precisa ser explícito.
	_, err = LeModosDaBranchBase(constitutionComModos("Módulos permitidos |"))
	assert.Error(t, err)

	// WIP inválido cai no padrão em vez de quebrar a leitura.
	m, err = LeModosDaBranchBase(constitutionComModos("WIP humano | muitos"))
	require.NoError(t, err)
	assert.Equal(t, WIPHumanoPadrao, m.WIPHumano)
}

// escreverConstituicao grava uma Constituição de teste em dir/sdd/memory.
func escreverConstituicao(t *testing.T, dir, conteudo string) {
	t.Helper()
	p := filepath.Join(dir, "sdd", "memory")
	require.NoError(t, os.MkdirAll(p, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(p, "constitution.md"), []byte(conteudo), 0o644))
}
