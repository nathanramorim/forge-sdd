package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dispatchTask = "# feat-77-01\n\n## Modelo do módulo\n- enxuto\n\n## O que já existe\n\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| a | b | c | criar | d |\n\n## Dependências novas\n- Nenhuma.\n\n## Rules e IDs da Constituição Aplicáveis\n- `SEG-01`.\n\n## Critérios de Aceitação Executáveis\n```bash\ngo test ./...\n```\n"

func dispatchProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfg, _ := json.Marshal(config.DefaultForgeConfig())
	testutil.WriteFile(t, dir, ".forge/config/forge.json", string(cfg))
	testutil.WriteFile(t, dir, "sdd/memory/constitution.md", "- **SEG-01:** Entrada validada.\n")
	testutil.WriteFile(t, dir, ".agents/rules/sdd-seguranca.md", "---\ndescription: s\nregras: [SEG-01]\n---\n")
	base := "sdd/discovery/77-auth-oauth"
	testutil.WriteFile(t, dir, base+"/discovery.md", "# Discovery 77\n")
	testutil.WriteFile(t, dir, base+"/criteria.md", "# C\n\n## O que já existe\n\n| Procurei | Onde | Achei | Decisão | Justificativa |\n|---|---|---|---|---|\n| a | b | c | criar | d |\n")
	testutil.WriteFile(t, dir, base+"/plan.md", "# P\n")
	testutil.WriteFile(t, dir, base+"/features/feat-77-01-auth.md", dispatchTask)
	return dir
}

func resetDispatchFlags() {
	_ = dispatchCmd.Flags().Set("card-evidence", "")
	_ = dispatchCmd.Flags().Set("dry-run", "false")
}

func TestDispatchCLI_SemForgeJsonOrientaPipelineInit(t *testing.T) {
	resetDispatchFlags()
	_, err := runRoot(t, "dispatch", "auth-oauth", "--dir", t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline init")
	assert.Contains(t, err.Error(), "4.0.0-beta")
}

func TestDispatchCLI_BloqueiaSemCardEDespachaComCard(t *testing.T) {
	dir := dispatchProject(t)
	resetDispatchFlags()
	out, err := runRoot(t, "dispatch", "auth-oauth", "--dir", dir)
	require.Error(t, err)
	assert.Contains(t, out, "Despacho bloqueado")
	assert.Contains(t, out, "falta a evidência do card")
	pend, _ := bus.List(dir, layout.RoleBuilder, layout.StatePending)
	assert.Empty(t, pend)

	ev := fmt.Sprintf(`{"provider":"clickup","id":"CU-1","url":"https://app.clickup.com/t/1","status":"To Do","feature":"auth-oauth","checked_at":%q}`, time.Now().UTC().Format(time.RFC3339))
	testutil.WriteFile(t, dir, "sdd/discovery/77-auth-oauth/card-evidence.json", ev)

	resetDispatchFlags()
	out, err = runRoot(t, "dispatch", "auth-oauth", "--dir", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "[DRY-RUN]")
	pend, _ = bus.List(dir, layout.RoleBuilder, layout.StatePending)
	assert.Empty(t, pend)

	resetDispatchFlags()
	out, err = runRoot(t, "dispatch", "auth-oauth", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "enfileirado em builder/pending")
	assert.Contains(t, out, `para "In Progress"`)
	pend, _ = bus.List(dir, layout.RoleBuilder, layout.StatePending)
	require.Len(t, pend, 1)
	assert.Equal(t, "CU-1", pend[0].Job.Card.ID)
	assert.FileExists(t, filepath.Join(dir, "sdd/discovery/77-auth-oauth/card-evidence.json"))
}
