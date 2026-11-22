package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/gate"
	"github.com/forge-sdd/cli/internal/queue"
)

// SyncAgentModels aplica agent_models aos agentes do projeto (forge update, mesmo sem atualização de
// versão). Com dryRun devolve o que gravaria sem gravar (ARQ-05).
func SyncAgentModels(targetDir string, dryRun bool) ([]string, error) {
	return syncAgentModels(targetDir, dryRun)
}

// applyAgentModels grava o `model:` de agent_models no cabeçalho de .claude/agents/*.md. Sem forge.json
// ou sem a chave, nada muda (inherit). Só edita agentes que já existem no projeto. Retorna os arquivos
// alterados. Todo valor é validado e todo arquivo conferido antes de gravar, para não deixar estado parcial.
func applyAgentModels(targetDir string) ([]string, error) {
	return syncAgentModels(targetDir, false)
}

func syncAgentModels(targetDir string, dryRun bool) ([]string, error) {
	cfg, found, err := config.LoadForgeConfig(targetDir)
	if err != nil {
		return nil, fmt.Errorf("agent_models: %w", err)
	}
	if !found || len(cfg.AgentModels) == 0 {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("agent_models: %w", err)
	}
	type edit struct {
		path string
		data []byte
	}
	var edits []edit
	for _, role := range config.SessionRoles() {
		model, ok := cfg.AgentModels[role]
		if !ok {
			continue
		}
		for _, name := range config.AgentModelFiles(role) {
			path := filepath.Join(targetDir, ".claude", "agents", name+".md")
			raw, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("agent_models: ler %s: %w", path, err)
			}
			out, err := setHeaderModel(string(raw), model)
			if err != nil {
				return nil, fmt.Errorf("agent_models: %s: %w", path, err)
			}
			if out != string(raw) {
				edits = append(edits, edit{path, []byte(out)})
			}
		}
	}
	var changed []string
	if dryRun {
		for _, e := range edits {
			changed = append(changed, e.path)
		}
		return changed, nil
	}
	for _, e := range edits {
		if err := queue.WriteFileAtomic(e.path, e.data, 0o644); err != nil {
			return changed, fmt.Errorf("agent_models: gravar %s: %w", e.path, err)
		}
		changed = append(changed, e.path)
	}
	return changed, nil
}

// setHeaderModel troca a linha `model:` do cabeçalho YAML (entre os dois `---`) e confere o teto de bytes.
func setHeaderModel(content, model string) (string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("sem cabeçalho `---`")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return "", fmt.Errorf("cabeçalho sem linha `model:`")
		}
		if strings.HasPrefix(lines[i], "model:") {
			lines[i] = "model: " + model
			out := strings.Join(lines, "\n")
			if len(out) > gate.MaxRoleAdapterBytes {
				return "", fmt.Errorf("definição com %d bytes passa do teto de %d", len(out), gate.MaxRoleAdapterBytes)
			}
			return out, nil
		}
	}
	return "", fmt.Errorf("cabeçalho não fechado")
}
