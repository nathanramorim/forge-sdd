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
const inboxHookCommand = "forge-sdd inbox --unread --hook"

// MergeClaudeInboxHook garante o hook UserPromptSubmit do inbox em .claude/settings.json sem
// sobrescrever a configuração do usuário (SEG-02): cria o arquivo se faltar, acrescenta só a
// entrada do inbox, não duplica e nunca toca em JSON inválido. Devolve true se o arquivo mudou.
func MergeClaudeInboxHook(targetDir string) (bool, error) {
	path := filepath.Join(targetDir, ".claude", "settings.json")
	root := map[string]any{}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		dec := json.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&root); err != nil {
			return false, fmt.Errorf("%s não é JSON válido; adicione o hook manualmente (%s): %w", path, inboxHookCommand, err)
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
	groups, ok := hooks["UserPromptSubmit"].([]any)
	if hooks["UserPromptSubmit"] != nil && !ok {
		return false, fmt.Errorf("%s: \"hooks.UserPromptSubmit\" não é uma lista; adicione o hook manualmente", path)
	}
	for _, g := range groups {
		if gm, ok := g.(map[string]any); ok {
			if inner, ok := gm["hooks"].([]any); ok {
				for _, h := range inner {
					if hm, ok := h.(map[string]any); ok {
						if cmd, _ := hm["command"].(string); strings.Contains(cmd, "forge-sdd inbox") {
							return false, nil // já instalado
						}
					}
				}
			}
		}
	}
	groups = append(groups, map[string]any{
		"hooks": []any{map[string]any{"type": "command", "command": inboxHookCommand}},
	})
	hooks["UserPromptSubmit"] = groups
	root["hooks"] = hooks

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
