package scaffold

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/template"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func migrateCfg() config.Config {
	return config.Config{
		Project:    "demo",
		Stack:      "go",
		Lang:       "pt-BR",
		SddVersion: "4.0.0-beta",
		Agents:     []string{config.AgentClaude},
	}
}

// legacyRendered devolve o conteúdo que o template anterior geraria para cfg.
func legacyRendered(t *testing.T, kind, name string, cfg config.Config) string {
	t.Helper()
	raw, err := legacyFS.ReadFile("legacy/" + kind + "/" + name + ".md.tmpl")
	require.NoError(t, err)
	tmpl, err := template.New(name).Parse(string(raw))
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, cfg))
	return buf.String()
}

// fixtureV3 monta um projeto no layout v3: discoveries soltos, sdd/features e comandos/papéis antigos.
func fixtureV3(t *testing.T, cfg config.Config) string {
	t.Helper()
	dir := t.TempDir()
	w := func(rel, content string) { testutil.WriteFile(t, dir, rel, content) }

	w("sdd/discovery/discovery-01-alfa.md", "d1")
	w("sdd/discovery/criteria-01-alfa.md", "c1")
	w("sdd/discovery/plan-01-alfa.md", "p1")
	w("sdd/discovery/66-iniciativa-nova/discovery.md", "iniciativa no layout novo")
	w("sdd/discovery/_tutorial/discovery.md", "tutorial")
	w("sdd/discovery/notas-do-usuario.md", "fora do padrão legado")
	w("sdd/discovery/.gitkeep", "")
	w("sdd/features/index.md", "indice")
	w("sdd/features/feat-01-alfa.md", "f1")
	w("sdd/features/feat-65-x/feat-65-01-y.md", "f65")
	w("sdd/memory/progress.md", "# Progress")

	w(".agents/commands/discovery.md", legacyRendered(t, "commands", "discovery", cfg))
	w(".agents/commands/status.md", legacyRendered(t, "commands", "status", cfg))
	w(".agents/roles/builder.md", legacyRendered(t, "roles", "builder", cfg)+"\nEDITADO PELO USUARIO\n")
	return dir
}

func treeHash(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	require.NoError(t, filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			lines = append(lines, "D "+rel)
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		lines = append(lines, "F "+rel+" "+hex.EncodeToString(sum[:]))
		return nil
	}))
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func TestMigrateLayout_PlanoNaoEscreveNada(t *testing.T) {
	cfg := migrateCfg()
	dir := fixtureV3(t, cfg)
	before := treeHash(t, dir)

	plan, err := PlanLayoutMigration(dir)
	require.NoError(t, err)
	obsolete, err := PlanObsoleteAgentArtifacts(dir, cfg)
	require.NoError(t, err)

	assert.Equal(t, before, treeHash(t, dir), "planejar não pode alterar a árvore (ARQ-05)")
	assert.False(t, plan.Empty())

	var froms []string
	for _, m := range plan.Moves {
		froms = append(froms, m.From)
	}
	assert.Contains(t, froms, "sdd/discovery/discovery-01-alfa.md")
	assert.Contains(t, froms, "sdd/features/index.md")
	assert.Contains(t, froms, "sdd/features/feat-65-x")
	assert.NotContains(t, froms, "sdd/discovery/66-iniciativa-nova", "iniciativa do layout novo fica onde está")
	assert.NotContains(t, froms, "sdd/discovery/_tutorial")

	assert.ElementsMatch(t, []string{".agents/commands/discovery.md", ".agents/commands/status.md"}, obsolete.Remove)
	assert.Equal(t, []string{".agents/roles/builder.md"}, obsolete.Preserve)
}

func TestMigrateLayout_AplicaMoveFazBackupEEhIdempotente(t *testing.T) {
	cfg := migrateCfg()
	dir := fixtureV3(t, cfg)

	plan, err := PlanLayoutMigration(dir)
	require.NoError(t, err)
	backup, err := ApplyLayoutMigration(dir, plan)
	require.NoError(t, err)
	require.NotEmpty(t, backup)

	// Movido para sdd/archive/ com o mesmo conteúdo.
	data, err := os.ReadFile(filepath.Join(dir, "sdd", "archive", "discovery", "criteria-01-alfa.md"))
	require.NoError(t, err)
	assert.Equal(t, "c1", string(data))
	data, err = os.ReadFile(filepath.Join(dir, "sdd", "archive", "features", "feat-65-x", "feat-65-01-y.md"))
	require.NoError(t, err)
	assert.Equal(t, "f65", string(data))
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "features"))

	// Backup guarda as origens.
	data, err = os.ReadFile(filepath.Join(backup, "sdd", "features", "index.md"))
	require.NoError(t, err)
	assert.Equal(t, "indice", string(data))

	// Preservados.
	assert.FileExists(t, filepath.Join(dir, "sdd", "discovery", "66-iniciativa-nova", "discovery.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "discovery", "_tutorial", "discovery.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "discovery", "notas-do-usuario.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", "memory", "progress.md"))

	// Segunda execução: plano vazio, nada muda, nenhum backup novo.
	after := treeHash(t, dir)
	plan2, err := PlanLayoutMigration(dir)
	require.NoError(t, err)
	assert.True(t, plan2.Empty())
	backup2, err := ApplyLayoutMigration(dir, plan2)
	require.NoError(t, err)
	assert.Empty(t, backup2)
	assert.Equal(t, after, treeHash(t, dir))
}

func TestMigrateLayout_Violacoes(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "sdd/features/feat-01-ok.md", "ok")
	testutil.WriteFile(t, dir, "sdd/features/feat-02-conflito.md", "novo")
	testutil.WriteFile(t, dir, "sdd/archive/features/feat-02-conflito.md", "ja arquivado")
	testutil.WriteFile(t, dir, "sdd/features/pasta-com-link/a.md", "a")

	outside := t.TempDir()
	testutil.WriteFile(t, outside, "segredo.txt", "fora do projeto")
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "sdd", "features", "link-solto")))
	require.NoError(t, os.Symlink(filepath.Join(outside, "segredo.txt"), filepath.Join(dir, "sdd", "features", "pasta-com-link", "link.md")))

	plan, err := PlanLayoutMigration(dir)
	require.NoError(t, err)

	moved := map[string]bool{}
	for _, m := range plan.Moves {
		moved[m.From] = true
	}
	assert.True(t, moved["sdd/features/feat-01-ok.md"])
	assert.False(t, moved["sdd/features/feat-02-conflito.md"], "destino existente não é sobrescrito")
	assert.False(t, moved["sdd/features/link-solto"], "link simbólico é recusado (SEG-02)")
	assert.False(t, moved["sdd/features/pasta-com-link"], "pasta com link simbólico é recusada (SEG-02)")

	reasons := map[string]string{}
	for _, s := range plan.Skipped {
		reasons[s.Path] = s.Reason
	}
	assert.Contains(t, reasons["sdd/features/feat-02-conflito.md"], "destino já existe")
	assert.Contains(t, reasons["sdd/features/link-solto"], "link simbólico")
	assert.Contains(t, reasons["sdd/features/pasta-com-link"], "link simbólico")

	_, err = ApplyLayoutMigration(dir, plan)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, "sdd", "archive", "features", "feat-02-conflito.md"))
	require.NoError(t, err)
	assert.Equal(t, "ja arquivado", string(data), "conteúdo arquivado anteriormente permanece")
	assert.FileExists(t, filepath.Join(dir, "sdd", "features", "feat-02-conflito.md"), "origem preservada quando há conflito")
	assert.FileExists(t, filepath.Join(outside, "segredo.txt"))
}

func TestObsoleteArtifacts_RemoveIdenticosEPreservaEditados(t *testing.T) {
	cfg := migrateCfg()
	dir := fixtureV3(t, cfg)

	plan, err := RemoveObsoleteAgentArtifacts(dir, cfg)
	require.NoError(t, err)
	assert.Len(t, plan.Remove, 2)

	assert.NoFileExists(t, filepath.Join(dir, ".agents", "commands", "discovery.md"))
	assert.NoDirExists(t, filepath.Join(dir, ".agents", "commands"), "pasta vazia é removida")
	assert.FileExists(t, filepath.Join(dir, ".agents", "roles", "builder.md"), "arquivo editado pelo usuário é preservado (SEG-02)")

	after := treeHash(t, dir)
	plan2, err := RemoveObsoleteAgentArtifacts(dir, cfg)
	require.NoError(t, err)
	assert.Empty(t, plan2.Remove)
	assert.Equal(t, after, treeHash(t, dir), "segunda execução não muda nada")
}

func TestRun_MigraELimpaEEhIdempotente(t *testing.T) {
	cfg := migrateCfg()
	dir := fixtureV3(t, cfg)

	_, err := Run(cfg, dir)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "sdd", "archive", "features", "index.md"))
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "features"))
	assert.NoFileExists(t, filepath.Join(dir, ".agents", "commands", "discovery.md"))
	assert.FileExists(t, filepath.Join(dir, ".agents", "roles", "builder.md"))
	assert.FileExists(t, filepath.Join(dir, ".agents", "skills", "sdd-concepcao", "SKILL.md"))
	assert.FileExists(t, filepath.Join(dir, "sdd", ".runtime", ".gitignore"))
	assert.NoFileExists(t, filepath.Join(dir, "sdd", "features", "index.md"))

	// A segunda execução não pode alterar nada do que o usuário ou a migração deixaram.
	// (backups/ é comparado à parte: a segunda execução não cria backup novo.)
	entriesBefore, err := os.ReadDir(filepath.Join(dir, "sdd", "backups"))
	require.NoError(t, err)
	hashBefore := treeHash(t, dir)
	_, err = Run(cfg, dir)
	require.NoError(t, err)
	assert.Equal(t, hashBefore, treeHash(t, dir))
	entriesAfter, err := os.ReadDir(filepath.Join(dir, "sdd", "backups"))
	require.NoError(t, err)
	assert.Len(t, entriesAfter, len(entriesBefore))
}

func TestRun_NovoProjetoNaoGeraEstadoGlobalNemRaizDeFeatures(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(migrateCfg(), dir)
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(dir, "sdd", "memory", "progress.md"))
	assert.NoDirExists(t, filepath.Join(dir, "sdd", "features"))

	// O estado de execução nasce ignorado pelo git: o scaffold basta, sem editar o .gitignore da raiz.
	testutil.InitRepo(t, dir, "main")
	testutil.WriteFile(t, dir, "sdd/.runtime/activity.json", "{}")
	testutil.WriteFile(t, dir, "sdd/.runtime/stations/builder.json", "{}")
	testutil.Git(t, dir, "add", "-A")
	tracked := testutil.Git(t, dir, "ls-files", "sdd/.runtime")
	assert.Equal(t, "sdd/.runtime/.gitignore\n", tracked, "só o .gitignore do runtime é versionado")
}

// Duas branches com iniciativas distintas, a partir de um projeto já migrado, mesclam sem conflito.
func TestProjetoMigrado_DuasIniciativasParalelasMesclamSemConflito(t *testing.T) {
	cfg := migrateCfg()
	dir := fixtureV3(t, cfg)
	_, err := Run(cfg, dir)
	require.NoError(t, err)

	testutil.InitRepo(t, dir, "main")
	testutil.Git(t, dir, "add", "-A")
	testutil.Git(t, dir, "commit", "-q", "-m", "projeto migrado")

	for _, it := range []string{"70-alfa", "71-beta"} {
		testutil.Git(t, dir, "checkout", "-q", "-B", "feat/"+it, "main")
		testutil.WriteFile(t, dir, "sdd/discovery/"+it+"/status.md", "# Status "+it+"\n")
		testutil.WriteFile(t, dir, "sdd/discovery/"+it+"/queue.json", `{"feature_id":"feat-`+it[:2]+`","items":[]}`)
		testutil.Git(t, dir, "add", "-A")
		testutil.Git(t, dir, "commit", "-q", "-m", it)
	}
	testutil.Git(t, dir, "checkout", "-q", "-B", "integracao", "feat/70-alfa")
	out := testutil.Git(t, dir, "merge", "--no-edit", "feat/71-beta")
	assert.NotContains(t, out, "CONFLICT")
}
