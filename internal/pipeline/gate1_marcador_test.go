package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gate1Em = time.Date(2026, 11, 21, 21, 30, 0, 0, time.UTC)

func TestSpecVoltaAPrincipalAntesDoBuilder(t *testing.T) {
	// Gate 1 aprovado: a spec já voltou à principal e o Builder pode começar.
	status := "# Status — 106\n\n**Gate 1:** aprovado em 21/11/2026\n"
	res, err := LiberaOBuilder(status)
	require.NoError(t, err)
	assert.True(t, res.Liberado)
	assert.True(t, res.SpecVoltou, "a spec volta antes do Builder (RN-1)")

	// A leitura reconhece o marcador com e sem negrito.
	e := LeGate1DoStatus(status)
	assert.True(t, e.Encontrado)
	assert.Equal(t, "aprovado", e.Marcador)
	assert.False(t, e.ModoSemHumano)

	// O marcador é gravado no status.md e substitui a linha anterior, sem duplicar.
	dir := t.TempDir()
	base := filepath.Join(dir, "sdd", "discovery", "106-x")
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "status.md"),
		[]byte("# Status\n\nGate 1: pendente\n"), 0o644))

	linha, err := GravaMarcadorDoGate1(dir, "106-x", "seguir", false, gate1Em)
	require.NoError(t, err)
	assert.Contains(t, linha, "Gate 1: aprovado")

	lido, err := os.ReadFile(filepath.Join(base, "status.md"))
	require.NoError(t, err)
	assert.Equal(t, 1, countOccurrences(string(lido), "Gate 1:"), "não duplica a linha")
	assert.Contains(t, string(lido), "2026-11-21T21:30:00Z")

	novo, err := LiberaOBuilder(string(lido))
	require.NoError(t, err)
	assert.True(t, novo.Liberado)
}

func TestBuilderNaoComecaSemSpecNaPrincipal(t *testing.T) {
	// Sem o marcador: recusa.
	_, err := LiberaOBuilder("# Status\n\nSem marcador nenhum.\n")
	require.ErrorIs(t, err, ErrBuilderSemSpecNaPrincipal)
	assert.Contains(t, err.Error(), "não tem o marcador")

	// Marcador pendente: recusa.
	_, err = LiberaOBuilder("Gate 1: pendente\n")
	require.ErrorIs(t, err, ErrBuilderSemSpecNaPrincipal)
	assert.Contains(t, err.Error(), "pendente")

	// Marcador de ajuste: recusa, citando que o humano disse [Ajustar].
	_, err = LiberaOBuilder("Gate 1: ajustar\n")
	require.ErrorIs(t, err, ErrBuilderSemSpecNaPrincipal)
	assert.Contains(t, err.Error(), "Ajustar")

	// Lixo que imita o marcador não vira aprovação.
	_, err = LiberaOBuilder("Gate 1: aprovado-sim-por-favor\n")
	require.Error(t, err)

	// Modo sem humano sem o marcador também não libera: a resposta automática é registrada.
	res, err := LiberaOBuilder("Modo: sem-humano\n")
	require.ErrorIs(t, err, ErrBuilderSemSpecNaPrincipal)
	assert.False(t, res.Liberado)
}

func TestModoSemHumanoRegistraGate1Automatico(t *testing.T) {
	// A linha do marcador no modo sem humano diz expressamente que foi automático.
	linha := LinhaDoMarcador("seguir", true, gate1Em)
	assert.Contains(t, linha, "Gate 1: aprovado")
	assert.Contains(t, linha, "automático")
	assert.Contains(t, linha, "modo sem humano")
	assert.Contains(t, linha, gate1Em.Format(time.RFC3339))

	// O aviso ao usuário nunca deixa parecer que uma pessoa aprovou.
	aviso := AvisoDoGate1NoModoSemHumano(gate1Em)
	assert.Contains(t, aviso, "nenhuma pessoa revisou")

	// No modo com humano a linha não diz "automático".
	humano := LinhaDoMarcador("seguir", false, gate1Em)
	assert.NotContains(t, humano, "automático")

	// O estado lido reconhece o modo sem humano.
	e := LeGate1DoStatus("Gate 1: aprovado (automático, modo sem humano)\nModo: sem-humano\n")
	assert.True(t, e.Aprovado())
	assert.True(t, e.ModoSemHumano)
}

func TestMesclaAutomaticaExigeModoEAviso(t *testing.T) {
	adapt := &adaptadorDeSpec{nome: "github"}

	// Modo com humano: recusa a mescla automática.
	_, err := MesclaAutomaticaNoModoSemHumano(
		DecisaoDeModo{Modo: ModoComHumano}, "/p", adapt, vcs.Mescla{Numero: 1}, gate1Em)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "só vale no modo sem humano")
	assert.Equal(t, 0, adapt.mesclou)

	// Modo sem humano sem o aviso aceito: recusa.
	_, err = MesclaAutomaticaNoModoSemHumano(
		DecisaoDeModo{Modo: ModoSemHumano}, "/p", adapt, vcs.Mescla{Numero: 1}, gate1Em)
	require.ErrorIs(t, err, ErrSemHumanoSemAviso)
	assert.Equal(t, 0, adapt.mesclou)

	// Modo sem humano com aviso aceito: mescla.
	_, err = MesclaAutomaticaNoModoSemHumano(
		DecisaoDeModo{Modo: ModoSemHumano, ComAvisado: true}, "/p", adapt,
		vcs.Mescla{Numero: 1}, gate1Em)
	require.NoError(t, err)
	assert.Equal(t, 1, adapt.mesclou)

	// VCS nenhum: informa, sem mesclar.
	nenhum := &adaptadorDeSpec{nome: "nenhum"}
	msg, err := MesclaAutomaticaNoModoSemHumano(
		DecisaoDeModo{Modo: ModoSemHumano, ComAvisado: true}, "/p", nenhum,
		vcs.Mescla{Branch: "feat/x"}, gate1Em)
	require.NoError(t, err)
	assert.Contains(t, msg, "merge é local")
	assert.Equal(t, 0, nenhum.mesclou)
}

func TestMarcadorRespeitaOrcamentoDoStatus(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "sdd", "discovery", "106-x")
	require.NoError(t, os.MkdirAll(base, 0o755))

	// status.md já cheio: gravar o marcador estouraria o teto e é recusado.
	cheio := "# Status\n\n" + repeat("x", 1010)
	require.NoError(t, os.WriteFile(filepath.Join(base, "status.md"), []byte(cheio), 0o644))
	_, err := GravaMarcadorDoGate1(dir, "106-x", "seguir", false, gate1Em)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enxugue o status")

	// Caminho do status é devolvido para quem precisa dele.
	caminho, err := CaminhoDoStatus(dir, "106-x")
	require.NoError(t, err)
	assert.Contains(t, caminho, "status.md")
}

func countOccurrences(s, sub string) int {
	n, i := 0, 0
	for {
		j := indexFrom(s, sub, i)
		if j < 0 {
			return n
		}
		n++
		i = j + len(sub)
	}
}

func indexFrom(s, sub string, from int) int {
	if from >= len(s) {
		return -1
	}
	for i := from; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
