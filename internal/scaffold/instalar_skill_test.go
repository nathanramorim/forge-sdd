package scaffold

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func novoForge(t *testing.T, dir string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sdd"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sdd", ".sddrc"), []byte("{}"), 0o644))
	return dir
}

func skillFiles() map[string][]byte {
	return map[string][]byte{"SKILL.md": []byte("# x"), "scripts/a.py": []byte("print(1)")}
}

func pluginFiles() map[string][]byte {
	return map[string][]byte{
		".claude-plugin/plugin.json": []byte("{}"),
		"scripts/diagrama.py":        []byte("print(1)"),
	}
}

func isolaHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func semArquivosTmp(t *testing.T, raiz string) {
	t.Helper()
	_ = filepath.WalkDir(raiz, func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(d.Name(), ".tmp") {
			t.Errorf("sobrou temporário %s", p)
		}
		return nil
	})
}

func exigeSemGravar(t *testing.T, home, projeto string) {
	t.Helper()
	assert.NoDirExists(t, filepath.Join(home, ".claude"))
	assert.NoDirExists(t, filepath.Join(projeto, ".agents"))
}

// casosInvalidos roda os erros de segurança comuns a skill e plugin.
func casosInvalidos(t *testing.T, inst func(proj, dest, slug string, a map[string][]byte) (Resultado, error), tipoDir string, base map[string][]byte) {
	home := isolaHome(t)
	novo := func() string { return novoForge(t, t.TempDir()) }
	copia := func(extra map[string][]byte) map[string][]byte {
		m := map[string][]byte{}
		for k, v := range base {
			m[k] = v
		}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}

	p := novo()
	_, err := inst(p, t.TempDir(), "ok-slug", base)
	require.Error(t, err, "destino fora da lista")
	exigeSemGravar(t, home, p)

	_, err = inst(p, p, "../x", base)
	require.Error(t, err, "slug com ..")
	exigeSemGravar(t, home, p)

	for nome, chave := range map[string]string{"dotdot": "../fora.txt", "meio": "a/../../fora.txt", "absoluta": filepath.Join(t.TempDir(), "abs.txt"), "barra": "/etc/x"} {
		_, err = inst(p, p, "ok-slug", copia(map[string][]byte{chave: []byte("x")}))
		require.Error(t, err, nome)
		exigeSemGravar(t, home, p)
	}
	semArquivosTmp(t, p)

	if runtime.GOOS == "windows" {
		return
	}
	fora := t.TempDir()
	p2 := novo()
	require.NoError(t, os.MkdirAll(filepath.Join(p2, ".agents", tipoDir), 0o755))
	require.NoError(t, os.Symlink(fora, filepath.Join(p2, ".agents", tipoDir, "ok-slug")))
	_, err = inst(p2, p2, "ok-slug", base)
	require.Error(t, err, "diretório da skill é symlink para fora")
	entradas, _ := os.ReadDir(fora)
	assert.Empty(t, entradas)

	p3 := novo()
	dir := filepath.Join(p3, ".agents", tipoDir, "ok-slug")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.Symlink(filepath.Join(fora, "alvo.txt"), filepath.Join(dir, "zzz.txt")))
	_, err = inst(p3, p3, "ok-slug", copia(map[string][]byte{"zzz.txt": []byte("x")}))
	require.Error(t, err, "arquivo alvo é symlink")
	assert.NoFileExists(t, filepath.Join(fora, "alvo.txt"))
	assert.NoFileExists(t, filepath.Join(dir, "SKILL.md"), "nada parcial")
	semArquivosTmp(t, p3)
}

func TestInstalaSkillSoNoProjetoAberto(t *testing.T) {
	casosInvalidos(t, InstalarSkill, "skills", skillFiles())

	home := isolaHome(t)
	p := novoForge(t, t.TempDir())
	res, err := InstalarSkill(p, p, "minha-skill", skillFiles())
	require.NoError(t, err)
	assert.Len(t, res.Gravados, 2)
	assert.Empty(t, res.Preservados)
	assert.FileExists(t, filepath.Join(p, ".agents", "skills", "minha-skill", "scripts", "a.py"))
	assert.NoDirExists(t, filepath.Join(home, ".claude"))
}

func TestInstalaPluginSoNoProjetoAberto(t *testing.T) {
	casosInvalidos(t, InstalarPlugin, "plugins", pluginFiles())

	home := isolaHome(t)
	p := novoForge(t, t.TempDir())
	res, err := InstalarPlugin(p, p, "diagram-studio", pluginFiles())
	require.NoError(t, err)
	assert.Len(t, res.Gravados, 2)
	assert.FileExists(t, filepath.Join(p, ".agents", "plugins", "diagram-studio", "scripts", "diagrama.py"))
	assert.NoDirExists(t, filepath.Join(home, ".claude"))
	assert.NoDirExists(t, filepath.Join(p, ".agents", "skills"))

	// arquivo do usuário já existente é preservado e reportado
	p2 := novoForge(t, t.TempDir())
	dir := filepath.Join(p2, ".agents", "plugins", "diagram-studio", "scripts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "diagrama.py"), []byte("meu"), 0o644))
	res, err = InstalarPlugin(p2, p2, "diagram-studio", pluginFiles())
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "diagrama.py")}, res.Preservados)
	assert.Len(t, res.Gravados, 1)
	b, _ := os.ReadFile(filepath.Join(dir, "diagrama.py"))
	assert.Equal(t, "meu", string(b))
}

func TestInstalaPerguntaHubOuSubprojeto(t *testing.T) {
	hub := novoForge(t, t.TempDir())
	a := novoForge(t, filepath.Join(hub, "b-api"))
	b := novoForge(t, filepath.Join(hub, "a-web"))
	require.NoError(t, os.MkdirAll(filepath.Join(hub, "sem-forge"), 0o755))
	destinos, err := ListarDestinosInstalacao(hub)
	require.NoError(t, err)
	assert.Equal(t, []string{hub, b, a}, destinos)

	solo := novoForge(t, t.TempDir())
	destinos, err = ListarDestinosInstalacao(solo)
	require.NoError(t, err)
	assert.Equal(t, []string{solo}, destinos)

	_, err = InstalarSkill(hub, a, "minha-skill", skillFiles())
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(a, ".agents", "skills", "minha-skill", "SKILL.md"))
	assert.NoDirExists(t, filepath.Join(b, ".agents"))
	assert.NoDirExists(t, filepath.Join(hub, ".agents"))
}

func TestInstalaNaoRepeteQuandoHubTem(t *testing.T) {
	hub := novoForge(t, t.TempDir())
	sub := novoForge(t, filepath.Join(hub, "api"))
	_, err := InstalarSkill(hub, hub, "minha-skill", skillFiles())
	require.NoError(t, err)

	onde, ok := DisponivelViaHub(sub, TipoSkill, "minha-skill")
	assert.True(t, ok)
	assert.Equal(t, hub, onde)
	_, ok = DisponivelViaHub(sub, TipoPlugin, "minha-skill")
	assert.False(t, ok)

	res, err := InstalarSkill(sub, sub, "minha-skill", skillFiles())
	require.NoError(t, err)
	assert.Equal(t, hub, res.DisponivelEm)
	assert.Empty(t, res.Gravados)
	assert.NoDirExists(t, filepath.Join(sub, ".agents"))

	// plugin pelo hub aberto, destino no subprojeto: também no-op
	_, err = InstalarPlugin(hub, hub, "diagram-studio", pluginFiles())
	require.NoError(t, err)
	res, err = InstalarPlugin(hub, sub, "diagram-studio", pluginFiles())
	require.NoError(t, err)
	assert.Equal(t, hub, res.DisponivelEm)
	assert.NoDirExists(t, filepath.Join(sub, ".agents"))
}

func TestInstalaNoSubprojetoQuandoSemHub(t *testing.T) {
	hub := novoForge(t, t.TempDir())
	sub := novoForge(t, filepath.Join(hub, "api"))
	outro := novoForge(t, filepath.Join(hub, "web"))

	_, ok := DisponivelViaHub(sub, TipoSkill, "minha-skill")
	assert.False(t, ok)
	res, err := InstalarSkill(sub, sub, "minha-skill", skillFiles())
	require.NoError(t, err)
	assert.Empty(t, res.DisponivelEm)
	assert.Len(t, res.Gravados, 2)
	assert.NoDirExists(t, filepath.Join(hub, ".agents"), "o subprojeto não grava no hub")
	assert.NoDirExists(t, filepath.Join(outro, ".agents"))

	_, err = InstalarSkill(hub, hub, "minha-skill", skillFiles())
	require.NoError(t, err)
	onde, ok := DisponivelViaHub(outro, TipoSkill, "minha-skill")
	assert.True(t, ok)
	assert.Equal(t, hub, onde)
	assert.NoDirExists(t, filepath.Join(outro, ".agents"))

	solo := t.TempDir() // sem forge acima: instala nele
	_, ok = DisponivelViaHub(solo, TipoSkill, "x-skill")
	assert.False(t, ok)
}

func TestInstalaOpcionalPorSkill(t *testing.T) {
	p := novoForge(t, t.TempDir())
	_, err := InstalarSkill(p, p, "", skillFiles())
	require.Error(t, err)
	assert.NoDirExists(t, filepath.Join(p, ".agents"))

	_, err = InstalarSkill(p, p, "skill-a", skillFiles())
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(p, ".agents", "skills", "skill-b"))

	// arquivo do usuário em skill ainda sem SKILL.md: intacto e reportado
	dir := filepath.Join(p, ".agents", "skills", "skill-c")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notas.md"), []byte("meu"), 0o644))
	res, err := InstalarSkill(p, p, "skill-c", map[string][]byte{"notas.md": []byte("novo"), "SKILL.md": []byte("# c")})
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "notas.md")}, res.Preservados)
	b, _ := os.ReadFile(filepath.Join(dir, "notas.md"))
	assert.Equal(t, "meu", string(b))

	// falha: um arquivo inválido impede qualquer gravação e não deixa .tmp
	_, err = InstalarSkill(p, p, "skill-d", map[string][]byte{"SKILL.md": []byte("# d"), "../x": []byte("x")})
	require.Error(t, err)
	assert.NoDirExists(t, filepath.Join(p, ".agents", "skills", "skill-d"))
	semArquivosTmp(t, p)

	_, err = InstalarSkill(p, p, "skill-e", nil)
	require.Error(t, err)
	_, err = instalar("outro", p, p, "skill-e", skillFiles())
	require.Error(t, err)
}

func TestShouldPreserveAgentsSkillsEPlugins(t *testing.T) {
	p := t.TempDir()
	mk := func(rel string) string {
		f := filepath.Join(p, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(f), 0o755))
		require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
		return f
	}
	assert.True(t, shouldPreserve(mk(".agents/skills/minha/SKILL.md"), p))
	assert.True(t, shouldPreserve(mk(".agents/plugins/diagram-studio/scripts/diagrama.py"), p))
	assert.False(t, shouldPreserve(mk(".agents/skills/sdd-build/SKILL.md"), p))
	assert.False(t, shouldPreserve(filepath.Join(p, ".agents/plugins/novo/x"), p), "inexistente não preserva")
}

func TestLerOrigemInstalacao(t *testing.T) {
	o := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(o, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(o, ".git", "HEAD"), []byte("x"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(o, "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(o, "scripts", "a.py"), []byte("1"), 0o644))
	m, err := LerOrigemInstalacao(o)
	require.NoError(t, err)
	assert.Equal(t, map[string][]byte{"scripts/a.py": []byte("1")}, m)
	_, err = LerOrigemInstalacao(filepath.Join(o, "nao-existe"))
	require.Error(t, err)
}
