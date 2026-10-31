package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposeValidation(t *testing.T) {
	dir := t.TempDir()

	_, err := Propose(dir, "", "Descrição válida", "1. Passo")
	assert.Error(t, err)

	_, err = Propose(dir, "Slug Invalido", "Descrição válida", "1. Passo")
	assert.Error(t, err)

	_, err = Propose(dir, "slug_invalido", "Descrição válida", "1. Passo")
	assert.Error(t, err)

	_, err = Propose(dir, "slug-valido", "   ", "1. Passo")
	assert.Error(t, err)
}

func TestProposeCreatesCanonicalSkillAndUpdatesIndex(t *testing.T) {
	dir := t.TempDir()

	path, err := Propose(dir, "validacao-entidade-go", "Padroniza validação de structs em Go", "1. Definir método Validate()\n2. Testar casos de borda")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, ".agents", "skills", "validacao-entidade-go", "SKILL.md"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "name: validacao-entidade-go")
	assert.Contains(t, content, "description: Padroniza validação de structs em Go")
	assert.Contains(t, content, "status: proposed")
	assert.Contains(t, content, "origin: auto-governance")
	assert.Contains(t, content, "# Skill: validacao-entidade-go")
	assert.Contains(t, content, "1. Definir método Validate()")

	indexPath := filepath.Join(dir, "sdd", "skills", "index.md")
	indexData, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Contains(t, string(indexData), "| validacao-entidade-go | Padroniza validação de structs em Go | proposed | .agents/skills/validacao-entidade-go/SKILL.md |")
}

func TestApprovePromotesSkillAndSyncsAdapters(t *testing.T) {
	dir := t.TempDir()

	_, err := Propose(dir, "migracao-sql-segura", "Checklist para migrações SQL idempotentes", "1. Usar IF NOT EXISTS\n2. Validar rollback")
	require.NoError(t, err)

	synced, err := Approve(dir, "migracao-sql-segura", []string{"claude", "gemini"})
	require.NoError(t, err)
	require.Len(t, synced, 2)

	canonicalPath := filepath.Join(dir, ".agents", "skills", "migracao-sql-segura", "SKILL.md")
	canonicalData, err := os.ReadFile(canonicalPath)
	require.NoError(t, err)
	assert.Contains(t, string(canonicalData), "status: approved")

	indexData, err := os.ReadFile(filepath.Join(dir, "sdd", "skills", "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(indexData), "| migracao-sql-segura | Checklist para migrações SQL idempotentes | approved | .agents/skills/migracao-sql-segura/SKILL.md |")
	assert.NotContains(t, string(indexData), "| migracao-sql-segura | Checklist para migrações SQL idempotentes | proposed |")

	claudeSkill := filepath.Join(dir, ".claude", "skills", "migracao-sql-segura", "SKILL.md")
	geminiSkill := filepath.Join(dir, ".gemini", "skills", "migracao-sql-segura", "SKILL.md")
	assert.FileExists(t, claudeSkill)
	assert.FileExists(t, geminiSkill)

	claudeData, err := os.ReadFile(claudeSkill)
	require.NoError(t, err)
	assert.Contains(t, string(claudeData), ManagedSkillMarker)
	assert.Contains(t, string(claudeData), "name: migracao-sql-segura")
	assert.Contains(t, string(claudeData), "1. Usar IF NOT EXISTS")
}

func TestSyncReadsSddrcAndPreservesManualSkills(t *testing.T) {
	dir := t.TempDir()

	// Configura sdd/.sddrc apenas com claude e gemini
	sddDir := filepath.Join(dir, "sdd")
	require.NoError(t, os.MkdirAll(sddDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sddDir, ".sddrc"), []byte(`{"project":"demo","version":"3.0.0-beta","agents":["claude","gemini"]}`), 0o644))

	// Cria uma skill aprovada e uma apenas proposta
	_, err := Propose(dir, "skill-aprovada", "Skill que deve sincronizar", "Passo 1")
	require.NoError(t, err)
	_, err = Propose(dir, "skill-proposta", "Skill ainda não aprovada", "Passo X")
	require.NoError(t, err)

	// Cria uma skill manual prévia em .claude/skills/skill-aprovada/SKILL.md (sem ManagedSkillMarker)
	manualClaudeDir := filepath.Join(dir, ".claude", "skills", "skill-aprovada")
	require.NoError(t, os.MkdirAll(manualClaudeDir, 0o755))
	manualContent := "# Skill manual do usuário — não sobrescrever\n"
	require.NoError(t, os.WriteFile(filepath.Join(manualClaudeDir, "SKILL.md"), []byte(manualContent), 0o644))

	// Aprova a primeira skill sem passar agents explícitos (usa .sddrc)
	synced, err := Approve(dir, "skill-aprovada", nil)
	require.NoError(t, err)

	// Apenas o adaptador do gemini deve ter sido gravado, pois o do claude era manual
	assert.Equal(t, []string{filepath.Join(dir, ".gemini", "skills", "skill-aprovada", "SKILL.md")}, synced)

	claudeAfter, err := os.ReadFile(filepath.Join(manualClaudeDir, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, manualContent, string(claudeAfter), "skill manual do usuário não deve ser sobrescrita")

	// Skill proposta não deve ter sido sincronizada
	assert.NoFileExists(t, filepath.Join(dir, ".gemini", "skills", "skill-proposta", "SKILL.md"))
}

func TestListReturnsProposedAndApprovedSkillsSorted(t *testing.T) {
	dir := t.TempDir()

	emptyList, err := List(dir)
	require.NoError(t, err)
	assert.Empty(t, emptyList)

	_, err = Propose(dir, "zeta-skill", "Descrição Zeta", "Corpo Zeta")
	require.NoError(t, err)
	_, err = Propose(dir, "alfa-skill", "Descrição Alfa", "Corpo Alfa")
	require.NoError(t, err)
	_, err = Approve(dir, "alfa-skill", []string{"claude"})
	require.NoError(t, err)

	items, err := List(dir)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "alfa-skill", items[0].Name)
	assert.Equal(t, StatusApproved, items[0].Status)
	assert.Equal(t, "Descrição Alfa", items[0].Description)

	assert.Equal(t, "zeta-skill", items[1].Name)
	assert.Equal(t, StatusProposed, items[1].Status)
	assert.Equal(t, "Descrição Zeta", items[1].Description)
}
