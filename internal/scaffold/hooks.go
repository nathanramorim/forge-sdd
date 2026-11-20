package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// inboxHookCommand é o comando que o hook da sessão Discovery executa a cada prompt.
const inboxHookCommand = "forge inbox --unread --hook"

// MergeClaudeInboxHook garante o hook UserPromptSubmit do inbox em .claude/settings.json sem
// sobrescrever a configuração do usuário (SEG-02): cria o arquivo se faltar, acrescenta só a
// entrada do inbox, não duplica e nunca toca em JSON inválido. Devolve true se o arquivo mudou.
func MergeClaudeInboxHook(targetDir string) (bool, error) {
	return mergeClaudeHook(targetDir, "UserPromptSubmit", inboxHookCommand, []string{"forge inbox", "forge-sdd inbox"}, false)
}

// doctorHookCommand roda o diagnóstico do harness ao abrir a sessão; a saída vira contexto.
const doctorHookCommand = "forge doctor --harness"

// MergeClaudeSessionStartHook garante o hook SessionStart do doctor em .claude/settings.json, com
// as mesmas garantias do hook do inbox (não sobrescreve, não duplica, não toca JSON inválido).
func MergeClaudeSessionStartHook(targetDir string) (bool, error) {
	return mergeClaudeHook(targetDir, "SessionStart", doctorHookCommand, []string{"forge doctor", "forge-sdd doctor"}, false)
}

func mergeClaudeHook(targetDir, event, command string, markers []string, dryRun bool) (bool, error) {
	path := filepath.Join(targetDir, ".claude", "settings.json")
	root := map[string]any{}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		dec := json.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&root); err != nil {
			return false, fmt.Errorf("%s não é JSON válido; adicione o hook manualmente (%s): %w", path, command, err)
		}
		if dec.More() {
			return false, fmt.Errorf("%s tem conteúdo extra após o objeto JSON; adicione o hook manualmente", path)
		}
		if root == nil {
			return false, fmt.Errorf("%s não contém um objeto JSON; adicione o hook manualmente", path)
		}
	case errors.Is(err, os.ErrNotExist):
		// cria do zero
	default:
		return false, fmt.Errorf("ler %s: %w", path, err)
	}

	hooks, ok := root["hooks"].(map[string]any)
	if root["hooks"] != nil && !ok {
		return false, fmt.Errorf("%s: a chave \"hooks\" não é um objeto; adicione o hook manualmente", path)
	}
	if hooks == nil {
		hooks = map[string]any{}
	}
	groups, ok := hooks[event].([]any)
	if hooks[event] != nil && !ok {
		return false, fmt.Errorf("%s: \"hooks.%s\" não é uma lista; adicione o hook manualmente", path, event)
	}
	for _, g := range groups {
		if gm, ok := g.(map[string]any); ok {
			if inner, ok := gm["hooks"].([]any); ok {
				for _, h := range inner {
					if hm, ok := h.(map[string]any); ok {
						cmd, _ := hm["command"].(string)
						for _, m := range markers {
							if strings.Contains(cmd, m) {
								return false, nil // já instalado
							}
						}
					}
				}
			}
		}
	}
	groups = append(groups, map[string]any{
		"hooks": []any{map[string]any{"type": "command", "command": command}},
	})
	hooks[event] = groups
	root["hooks"] = hooks
	if dryRun {
		return true, nil
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		return false, fmt.Errorf("gravar %s: %w", path, err)
	}
	return true, nil
}

// mcpServerName é a chave do servidor forge-sdd em .mcp.json.
const mcpServerName = "forge-sdd"

// MergeClaudeMCPServer registra `forge mcp serve` em .mcp.json (Claude Code) sem sobrescrever o
// que o usuário já tem: cria o arquivo se faltar, acrescenta só a entrada forge-sdd, não duplica
// e nunca toca em JSON inválido. Devolve true se o arquivo mudou.
func MergeClaudeMCPServer(targetDir string) (bool, error) {
	return mergeClaudeMCPServer(targetDir, false)
}

func mergeClaudeMCPServer(targetDir string, dryRun bool) (bool, error) {
	path := filepath.Join(targetDir, ".mcp.json")
	root := map[string]any{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		dec := json.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&root); err != nil || root == nil {
			return false, fmt.Errorf("%s não é um objeto JSON válido; registre o servidor manualmente (forge mcp serve): %v", path, err)
		}
		if dec.More() {
			return false, fmt.Errorf("%s tem conteúdo extra após o objeto JSON; registre o servidor manualmente", path)
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return false, fmt.Errorf("ler %s: %w", path, err)
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if root["mcpServers"] != nil && !ok {
		return false, fmt.Errorf("%s: \"mcpServers\" não é um objeto; registre o servidor manualmente", path)
	}
	if servers == nil {
		servers = map[string]any{}
	}
	if _, existe := servers[mcpServerName]; existe {
		return false, nil
	}
	servers[mcpServerName] = map[string]any{"command": "forge", "args": []any{"mcp", "serve"}}
	root["mcpServers"] = servers
	if dryRun {
		return true, nil
	}
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		return false, fmt.Errorf("gravar %s: %w", path, err)
	}
	return true, nil
}

// SyncClaudeIntegration aplica o hook SessionStart do doctor e o servidor MCP forge-sdd mesmo sem
// atualização de versão (forge update). Com dryRun só informa o que gravaria. Erros de um item
// (ex.: JSON inválido) viram aviso e não impedem o outro; o arquivo inválido nunca é tocado.
func SyncClaudeIntegration(targetDir string, dryRun bool) (changed, warnings []string) {
	if ok, err := mergeClaudeHook(targetDir, "SessionStart", doctorHookCommand, []string{"forge doctor", "forge-sdd doctor"}, dryRun); err != nil {
		warnings = append(warnings, "hook SessionStart não instalado: "+err.Error())
	} else if ok {
		changed = append(changed, "hook SessionStart")
	}
	if ok, err := mergeClaudeMCPServer(targetDir, dryRun); err != nil {
		warnings = append(warnings, "MCP forge-sdd não registrado: "+err.Error())
	} else if ok {
		changed = append(changed, "MCP forge-sdd")
	}
	return changed, warnings
}
