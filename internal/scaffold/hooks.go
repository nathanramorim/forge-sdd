package scaffold

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
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
	if err := checkSafeTarget(targetDir, path); err != nil {
		return false, err
	}
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
	if err := writeFileSafe(targetDir, path, append(out, '\n')); err != nil {
		return false, err
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
	if err := checkSafeTarget(targetDir, path); err != nil {
		return false, err
	}
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
	if err := writeFileSafe(targetDir, path, append(out, '\n')); err != nil {
		return false, err
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

// checkSafeTarget recusa (SEG-02) caminho que seja link simbólico ou não seja arquivo regular, e
// diretório pai que seja link simbólico ou resolva para fora do projeto. Arquivo ou pai ausentes
// são aceitos (serão criados).
func checkSafeTarget(targetDir, path string) error {
	// Base única e absoluta: `init .`, `./x` e `a/../b` precisam comparar com o EvalSymlinks do projeto.
	if abs, err := filepath.Abs(targetDir); err == nil {
		targetDir = abs
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	root, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		return fmt.Errorf("resolver %s: %w", targetDir, err)
	}
	dir := filepath.Dir(path)
	if rel, err := filepath.Rel(targetDir, dir); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s está fora do projeto; recusado (SEG-02)", dir)
	} else if rel != "." {
		// Cada componente existente entre o projeto e o destino precisa ser diretório real: um
		// ancestral simbólico faria o MkdirAll criar pastas fora do projeto.
		cur := targetDir
		for _, part := range strings.Split(rel, string(filepath.Separator)) {
			cur = filepath.Join(cur, part)
			fi, err := os.Lstat(cur)
			if errors.Is(err, os.ErrNotExist) {
				break
			}
			if err != nil {
				return fmt.Errorf("verificar %s: %w", cur, err)
			}
			if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
				return fmt.Errorf("%s é link simbólico ou não é diretório; recusado (SEG-02)", cur)
			}
		}
		if real, err := filepath.EvalSymlinks(nearestExisting(dir)); err != nil || (real != root && !strings.HasPrefix(real, root+string(filepath.Separator))) {
			return fmt.Errorf("%s resolve para fora do projeto; recusado (SEG-02)", dir)
		}
	}
	fi, err := os.Lstat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("verificar %s: %w", path, err)
	case !fi.Mode().IsRegular():
		return fmt.Errorf("%s é link simbólico ou não é arquivo regular; recusado (SEG-02)", path)
	}
	return nil
}

// writeFileSafe grava via arquivo temporário exclusivo no mesmo diretório e rename, preservando o
// modo do arquivo existente (novo: 0644). Revalida o alvo logo antes de gravar.
func writeFileSafe(targetDir, path string, data []byte) error {
	return writeFileSafeMode(targetDir, path, data, 0o644)
}

// WriteFileSafe expõe writeFileSafe (SEG-02) aos comandos que gravam no projeto do usuário.
func WriteFileSafe(targetDir, path string, data []byte) error {
	return writeFileSafe(targetDir, path, data)
}

// nearestExisting devolve o ancestral existente mais próximo de dir.
func nearestExisting(dir string) string {
	for {
		if _, err := os.Lstat(dir); err == nil || filepath.Dir(dir) == dir {
			return dir
		}
		dir = filepath.Dir(dir)
	}
}

// writeFileSafeMode é writeFileSafe com o modo de arquivo novo explícito (ex.: 0755 para scripts).
func writeFileSafeMode(targetDir, path string, data []byte, newMode os.FileMode) error {
	if abs, err := filepath.Abs(targetDir); err == nil {
		targetDir = abs
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if err := checkSafeTarget(targetDir, path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := checkSafeTarget(targetDir, path); err != nil {
		return err
	}
	mode := newMode
	if fi, err := os.Lstat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	var buf [6]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp-"+hex.EncodeToString(buf[:]))
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("gravar %s: %w", path, err)
	}
	fail := func(err error) error {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("gravar %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		return fail(err)
	}
	if err := f.Chmod(mode); err != nil {
		return fail(err)
	}
	if err := f.Sync(); err != nil {
		return fail(err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("gravar %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("gravar %s: %w", path, err)
	}
	return nil
}
