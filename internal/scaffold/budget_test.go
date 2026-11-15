package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Os orçamentos têm definição única em gate (DUP-04): o portão e estes testes medem o mesmo limite.
const (
	maxSkillBytes       = gate.MaxSkillBytes
	maxAdapterBytes     = gate.MaxAdapterBytes
	maxRoleAdapterBytes = gate.MaxRoleAdapterBytes
	maxInstructionBytes = gate.MaxInstructionBytes
)

// Orçamento do conjunto de leitura de cada comando, em bytes (feat 69-07). Mede o que um agente lê
// no projeto gerado: instrução raiz, adaptador, skill, referência e memória. A Constitution entra
// inteira no /discovery (não há task para filtrar IDs); o Builder a recebe filtrada por
// `forge-sdd context`, por isso fica fora do conjunto dele.
const (
	maxDiscoveryReadSet = 22 * 1024
	maxBuilderReadSet   = 15 * 1024
)

func renderAllAgents(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Config{
		Project:    "demo",
		Stack:      "go",
		DB:         "none",
		Lang:       "pt-BR",
		SddVersion: "4.0.0-beta",
		Agents:     []string{config.AgentClaude, config.AgentGemini, config.AgentCopilot, config.AgentOpenAI},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	return dir
}

func sizeOf(t *testing.T, path string) int {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, path)
	return int(info.Size())
}

func TestSkillsDeEstacaoRespeitamOrcamento(t *testing.T) {
	dir := renderAllAgents(t)
	for _, skill := range []string{"sdd-concepcao", "sdd-planejamento", "sdd-build", "sdd-governanca", "sdd-operacao"} {
		p := filepath.Join(dir, ".agents", "skills", skill, "SKILL.md")
		assert.LessOrEqual(t, sizeOf(t, p), maxSkillBytes, "%s excede o orçamento", skill)
		data, err := os.ReadFile(p)
		require.NoError(t, err)
		assert.Contains(t, string(data), "scope: core", "%s deve ser skill de núcleo", skill)
		assert.Contains(t, string(data), "status: approved")
	}
}

func TestAdaptadoresEInstrucoesRaizRespeitamOrcamento(t *testing.T) {
	dir := renderAllAgents(t)

	commandDirs := []string{
		filepath.Join(".claude", "commands"),
		filepath.Join(".gemini", "prompts"),
		filepath.Join(".github", "prompts"),
	}
	for _, d := range commandDirs {
		entries, err := os.ReadDir(filepath.Join(dir, d))
		require.NoError(t, err, d)
		require.NotEmpty(t, entries, d)
		for _, e := range entries {
			assert.LessOrEqual(t, sizeOf(t, filepath.Join(dir, d, e.Name())), maxAdapterBytes, "%s/%s", d, e.Name())
		}
	}

	roleDirs := []string{
		filepath.Join(".claude", "agents"),
		filepath.Join(".gemini", "skills"),
		filepath.Join(".github", "chatmodes"),
	}
	for _, d := range roleDirs {
		entries, err := os.ReadDir(filepath.Join(dir, d))
		require.NoError(t, err, d)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			assert.LessOrEqual(t, sizeOf(t, filepath.Join(dir, d, e.Name())), maxRoleAdapterBytes, "%s/%s", d, e.Name())
		}
	}

	for _, f := range []string{"CLAUDE.md", "GEMINI.md", "OPENAI.md", "AGENTS.md", filepath.Join(".github", "copilot-instructions.md")} {
		assert.LessOrEqual(t, sizeOf(t, filepath.Join(dir, f)), maxInstructionBytes, f)
	}
}

func TestScaffoldNaoGeraComandosNemPapeisLegados(t *testing.T) {
	dir := renderAllAgents(t)
	assert.NoDirExists(t, filepath.Join(dir, ".agents", "commands"))
	assert.NoDirExists(t, filepath.Join(dir, ".agents", "roles"))

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(path, ".json") {
			return err
		}
		data, rerr := os.ReadFile(path)
		require.NoError(t, rerr)
		rel, _ := filepath.Rel(dir, path)
		assert.NotContains(t, string(data), ".agents/commands/", rel)
		assert.NotContains(t, string(data), ".agents/roles/", rel)
		return nil
	})
	require.NoError(t, err)
}

func TestAdaptadoresApontamParaArquivosQueExistem(t *testing.T) {
	dir := renderAllAgents(t)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if !strings.HasPrefix(rel, ".claude/") && !strings.HasPrefix(rel, ".gemini/") && !strings.HasPrefix(rel, ".github/") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		require.NoError(t, rerr)
		for _, line := range strings.Split(string(data), "`") {
			if strings.HasPrefix(line, ".agents/skills/") && strings.HasSuffix(line, ".md") {
				assert.FileExists(t, filepath.Join(dir, line), "%s aponta para arquivo inexistente", rel)
			}
		}
		return nil
	})
	require.NoError(t, err)
}

func readSetSize(t *testing.T, dir string, files ...string) int {
	t.Helper()
	total := 0
	for _, f := range files {
		total += sizeOf(t, filepath.Join(dir, filepath.FromSlash(f)))
	}
	return total
}

func TestContextBudget_ConjuntoDeLeituraPorComando(t *testing.T) {
	dir := renderAllAgents(t)

	discovery := readSetSize(t, dir,
		"CLAUDE.md",
		".claude/commands/discovery.md",
		".agents/skills/sdd-concepcao/SKILL.md",
		".agents/skills/sdd-concepcao/references/discovery.md",
		"sdd/memory/clarify.md",
		"sdd/memory/naming-convention.md",
		"sdd/memory/constitution.md",
		"sdd/memory/lessons.md",
		"sdd/spec/concepts.md",
	)
	assert.LessOrEqual(t, discovery, maxDiscoveryReadSet, "conjunto de leitura de /discovery acima do orçamento")

	builder := readSetSize(t, dir,
		"CLAUDE.md",
		".claude/commands/implementar.md",
		".agents/skills/sdd-build/SKILL.md",
		".agents/skills/sdd-build/references/implementar.md",
		"sdd/memory/lessons.md",
	)
	assert.LessOrEqual(t, builder, maxBuilderReadSet, "conjunto de leitura do Builder acima do orçamento")

	t.Logf("conjunto de leitura: /discovery=%d bytes (limite %d), /implementar=%d bytes (limite %d)", discovery, maxDiscoveryReadSet, builder, maxBuilderReadSet)
}

func TestContextBudget_ConstitutionDoScaffoldDentroDoLimite(t *testing.T) {
	dir := renderAllAgents(t)
	assert.LessOrEqual(t, sizeOf(t, filepath.Join(dir, "sdd", "memory", "constitution.md")), gate.MaxConstitutionSize)
}

func TestScaffoldGeradoPassaNoOrcamentoDoPortao(t *testing.T) {
	dir := renderAllAgents(t)
	results, err := gate.RunChecks(dir, gate.Options{})
	require.NoError(t, err)
	for _, r := range results {
		if r.ID == "P-24" {
			assert.True(t, r.Passed, r.Evidence)
			return
		}
	}
	t.Fatal("P-24 ausente do relatório do portão")
}

// Nomes de projeto longos não podem estourar o orçamento das instruções raiz.
func TestInstrucoesRaizComNomeDeProjetoLongoRespeitamOrcamento(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		Project:    strings.Repeat("projeto-com-nome-bem-longo-", 2)[:50],
		Stack:      "typescript",
		DB:         "postgres",
		Lang:       "pt-BR",
		SddVersion: "4.0.0-beta",
		Agents:     []string{config.AgentClaude, config.AgentGemini, config.AgentCopilot, config.AgentOpenAI},
	}
	_, err := Run(cfg, dir)
	require.NoError(t, err)
	for _, f := range []string{"CLAUDE.md", "GEMINI.md", "OPENAI.md", "AGENTS.md", filepath.Join(".github", "copilot-instructions.md")} {
		assert.LessOrEqual(t, sizeOf(t, filepath.Join(dir, f)), maxInstructionBytes, f)
	}
}
