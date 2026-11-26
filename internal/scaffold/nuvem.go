package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// Forge em sessões do Claude Code na nuvem: um script versionado em .claude/hooks e um hook
// SessionStart que o chama. O script só age com CLAUDE_CODE_REMOTE=true; no computador da pessoa
// sai na hora, sem custo.

const (
	// NuvemScriptRel é o script do hook, relativo ao projeto.
	NuvemScriptRel = ".claude/hooks/forge-nuvem.sh"
	// nuvemHookMarker reconhece a entrada do hook em .claude/settings.json.
	nuvemHookMarker = "forge-nuvem.sh"
	// nuvemHookTimeout cobre a compilação do forge no modo git (dezenas de segundos).
	nuvemHookTimeout = 300
	// NuvemPacoteNPM e NuvemRepoGit são de onde o script instala o forge.
	NuvemPacoteNPM = "@nathanramorim/forge-sdd"
	NuvemRepoGit   = "https://github.com/nathanramorim/forge-sdd.git"
)

const nuvemHookCommand = `bash "$CLAUDE_PROJECT_DIR/` + NuvemScriptRel + `"`

// NuvemRef devolve a versão (npm) ou a branch (git) que o script instala.
func NuvemRef(n config.NuvemConfig, version string) string {
	if n.Ref != "" {
		return n.Ref
	}
	if n.InstalarEfetivo() == config.NuvemInstalarGit {
		return "main"
	}
	return strings.TrimPrefix(version, "v")
}

// RenderNuvemScript monta o script do hook. Os valores vêm de nuvem.* já validados (sem
// metacaracteres de shell) e ainda vão entre aspas.
func RenderNuvemScript(n config.NuvemConfig, version string) string {
	ref := NuvemRef(n, version)
	var instalar string
	if n.InstalarEfetivo() == config.NuvemInstalarGit {
		instalar = fmt.Sprintf(`  tmp="$(mktemp -d)"
  if git clone --quiet --depth 1 --branch %q %q "$tmp/forge" >"$LOG" 2>&1; then
    (cd "$tmp/forge" && go build -o "$BIN_DIR/forge" ./cmd/forge-sdd) >>"$LOG" 2>&1
  fi
  rm -rf "$tmp"`, ref, NuvemRepoGit)
	} else {
		instalar = fmt.Sprintf(`  npm install -g --prefix "$HOME/.local" %q >"$LOG" 2>&1`, NuvemPacoteNPM+"@"+ref)
	}
	return `#!/usr/bin/env bash
# Gerado pelo forge (forge nuvem ativar). Para mudar, use o menu do forge ou forge nuvem ativar.
# Só age em sessão do Claude Code na nuvem (CLAUDE_CODE_REMOTE=true); no computador sai na hora.
set -u
[ "${CLAUDE_CODE_REMOTE:-}" = "true" ] || exit 0
cd "${CLAUDE_PROJECT_DIR:-$PWD}" || exit 0
BIN_DIR="$HOME/.local/bin"
mkdir -p "$BIN_DIR"
export PATH="$BIN_DIR:$PATH"
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export PATH=\"$BIN_DIR:\$PATH\"" >>"$CLAUDE_ENV_FILE"
fi
LOG="${TMPDIR:-/tmp}/forge-nuvem-instalar.log"
if ! command -v forge >/dev/null 2>&1; then
  echo "forge nuvem: instalando o forge ` + ref + ` (` + n.InstalarEfetivo() + `)" >&2
` + instalar + `
fi
if ! command -v forge >/dev/null 2>&1; then
  echo "forge nuvem: não consegui instalar o forge (log em $LOG). Confira a rede do ambiente e rode forge nuvem status."
  exit 0
fi
exec forge nuvem iniciar
`
}

// AplicarNuvem grava (ativo) ou remove (inativo) o script e o hook conforme cfg.Nuvem. Devolve o
// que mudou. JSON inválido em settings.json nunca é tocado (vira erro).
func AplicarNuvem(targetDir string, cfg config.ForgeConfig, version string, dryRun bool) ([]string, error) {
	var mudou []string
	script := filepath.Join(targetDir, NuvemScriptRel)
	if !cfg.NuvemAtiva() {
		if ok, err := ajustarHookNuvem(targetDir, false, dryRun); err != nil {
			return mudou, err
		} else if ok {
			mudou = append(mudou, "hook SessionStart da nuvem removido")
		}
		if _, err := os.Lstat(script); err == nil {
			if !dryRun {
				if err := checkSafeTarget(targetDir, script); err != nil {
					return mudou, err
				}
				if err := os.Remove(script); err != nil {
					return mudou, err
				}
			}
			mudou = append(mudou, NuvemScriptRel+" removido")
		}
		return mudou, nil
	}
	novo := []byte(RenderNuvemScript(cfg.NuvemOuPadrao(), version))
	if atual, err := os.ReadFile(script); err != nil || !bytes.Equal(atual, novo) {
		if !dryRun {
			if err := writeFileSafeMode(targetDir, script, novo, 0o755); err != nil {
				return mudou, err
			}
		}
		mudou = append(mudou, NuvemScriptRel)
	}
	if ok, err := ajustarHookNuvem(targetDir, true, dryRun); err != nil {
		return mudou, err
	} else if ok {
		mudou = append(mudou, "hook SessionStart da nuvem")
	}
	return mudou, nil
}

// NuvemHookInstalado diz se .claude/settings.json tem o hook da nuvem.
func NuvemHookInstalado(targetDir string) bool {
	data, err := os.ReadFile(filepath.Join(targetDir, ".claude", "settings.json"))
	return err == nil && bytes.Contains(data, []byte(nuvemHookMarker))
}

// ajustarHookNuvem põe a entrada do hook no início de SessionStart (antes do doctor, que precisa
// do forge instalado) ou a tira. Devolve true se o arquivo mudou.
func ajustarHookNuvem(targetDir string, instalar, dryRun bool) (bool, error) {
	path := filepath.Join(targetDir, ".claude", "settings.json")
	if err := checkSafeTarget(targetDir, path); err != nil {
		return false, err
	}
	root := map[string]any{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		dec := json.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&root); err != nil || dec.More() || root == nil {
			if err == nil {
				err = errors.New("não é um único objeto JSON")
			}
			return false, &InvalidJSONError{Path: path, Fix: "adicione o hook SessionStart manualmente (" + nuvemHookCommand + ")", Err: err}
		}
	case errors.Is(err, os.ErrNotExist):
		if !instalar {
			return false, nil
		}
	default:
		return false, fmt.Errorf("ler %s: %w", path, err)
	}
	hooks, ok := root["hooks"].(map[string]any)
	if root["hooks"] != nil && !ok {
		return false, fmt.Errorf("%s: a chave \"hooks\" não é um objeto; ajuste o hook manualmente", path)
	}
	if hooks == nil {
		hooks = map[string]any{}
	}
	groups, ok := hooks["SessionStart"].([]any)
	if hooks["SessionStart"] != nil && !ok {
		return false, fmt.Errorf("%s: \"hooks.SessionStart\" não é uma lista; ajuste o hook manualmente", path)
	}
	var resto []any
	tinha := false
	for _, g := range groups {
		if grupoDaNuvem(g) {
			tinha = true
			continue
		}
		resto = append(resto, g)
	}
	if instalar == tinha {
		return false, nil
	}
	if instalar {
		entrada := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": nuvemHookCommand, "timeout": nuvemHookTimeout}}}
		resto = append([]any{entrada}, resto...)
	}
	if len(resto) == 0 {
		delete(hooks, "SessionStart")
		root["hooks"] = hooks
		if dryRun {
			return true, nil
		}
		out, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return false, err
		}
		return true, writeFileSafe(targetDir, path, append(out, '\n'))
	}
	return writeHooksRoot(targetDir, path, root, hooks, "SessionStart", resto, dryRun)
}

func grupoDaNuvem(g any) bool {
	gm, ok := g.(map[string]any)
	if !ok {
		return false
	}
	inner, _ := gm["hooks"].([]any)
	for _, h := range inner {
		if hm, ok := h.(map[string]any); ok {
			if cmd, _ := hm["command"].(string); strings.Contains(cmd, nuvemHookMarker) {
				return true
			}
		}
	}
	return false
}
