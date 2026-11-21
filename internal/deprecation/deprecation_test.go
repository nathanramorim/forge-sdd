package deprecation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projeto(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	return dir
}

func reg() Registro {
	return Registro{
		Quando: time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC), Comando: "queue stations", Contexto: "stations --dir",
		Agente: "claude-code", Substituto: "`forge queue`", Desde: "ADR-0021", VersaoForge: "5.0.0",
	}
}

func TestDepreciadoGravaLinhaCSV(t *testing.T) {
	dir := projeto(t)
	require.NoError(t, Registrar(dir, reg()))
	require.NoError(t, Registrar(dir, reg()))
	b, err := os.ReadFile(filepath.Join(dir, RelPath))
	require.NoError(t, err)
	linhas := strings.Split(strings.TrimSpace(string(b)), "\n")
	require.Len(t, linhas, 3, "cabeçalho único + 2 linhas")
	assert.Equal(t, "data_utc,comando,contexto,agente,substituto,desde,versao_forge", linhas[0])
	assert.Equal(t, "2026-11-20T10:00:00Z,queue stations,stations --dir,claude-code,`forge queue`,ADR-0021,5.0.0", linhas[1])
}

func TestRegistroSoGravaComSdd(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Registrar(dir, reg()))
	_, err := os.Stat(filepath.Join(dir, "sdd"))
	assert.True(t, os.IsNotExist(err))
}

func TestAgenteDetectadoPorVariavel(t *testing.T) {
	env := map[string]string{"CLAUDECODE": "1"}
	assert.Equal(t, "claude-code", Agente(func(k string) string { return env[k] }))
	assert.Equal(t, "desconhecido", Agente(func(string) string { return "" }))
	// o valor nunca vira o rótulo
	env["CLAUDECODE"] = "segredo-xyz"
	assert.Equal(t, "claude-code", Agente(func(k string) string { return env[k] }))
}

func TestContextoSemValoresDeArgumento(t *testing.T) {
	c := Contexto("handoff", []string{"from", "to", "/home/maria/segredo", "x=senha", "A B", "id"})
	assert.Equal(t, "handoff --from --to --id", c)
	assert.NotContains(t, c, "maria")
	assert.NotContains(t, c, "senha")
	assert.Equal(t, "", Contexto("../etc", nil))
}

func TestCelulaNeutralizaFormulaESaneia(t *testing.T) {
	for _, in := range []string{"=1+1", "+cmd", "-2", "@SUM(A1)"} {
		assert.Equal(t, "'"+in, Celula(in), in)
	}
	got := Celula("a\nb\x1b[31mred\x1b[0m\r")
	assert.NotContains(t, got, "\n")
	assert.NotContains(t, got, "\r")
	assert.NotContains(t, got, "\x1b")
}

func TestRegistroRecusaLinkSimbolico(t *testing.T) {
	dir := projeto(t)
	alvo := filepath.Join(t.TempDir(), "fora.csv")
	require.NoError(t, os.WriteFile(alvo, []byte("intacto"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd", "memory", "logs"), 0o755))
	require.NoError(t, os.Symlink(alvo, filepath.Join(dir, RelPath)))
	require.Error(t, Registrar(dir, reg()))
	b, _ := os.ReadFile(alvo)
	assert.Equal(t, "intacto", string(b))

	// pasta de logs como link simbólico para fora
	dir2 := projeto(t)
	fora := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir2, "sdd", "memory"), 0o755))
	require.NoError(t, os.Symlink(fora, filepath.Join(dir2, "sdd", "memory", "logs")))
	require.Error(t, Registrar(dir2, reg()))
	ents, _ := os.ReadDir(fora)
	assert.Empty(t, ents)
}

func TestRegistroRespeitaTamanhoMaximo(t *testing.T) {
	dir := projeto(t)
	p := filepath.Join(dir, RelPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	cheio := []byte(strings.Repeat("x", MaxBytes))
	require.NoError(t, os.WriteFile(p, cheio, 0o644))
	assert.ErrorIs(t, Registrar(dir, reg()), ErrLimite)
	b, _ := os.ReadFile(p)
	assert.Len(t, b, MaxBytes)

	// quase cheio: a linha que estouraria também não grava
	require.NoError(t, os.WriteFile(p, cheio[:MaxBytes-5], 0o644))
	assert.ErrorIs(t, Registrar(dir, reg()), ErrLimite)
	b, _ = os.ReadFile(p)
	assert.Len(t, b, MaxBytes-5)
}

func TestCatalogoCobreDepreciadosEAliases(t *testing.T) {
	for _, c := range []string{"queue stations", "queue handoff", "attach"} {
		e, ok := Buscar(c)
		require.True(t, ok, c)
		assert.NotEmpty(t, e.Substituto)
		assert.Equal(t, "ADR-0021", e.Desde)
	}
	_, ok := Buscar("queue")
	assert.False(t, ok)
	assert.Len(t, AliasesDeTexto, 2)
}
