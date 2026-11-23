package scaffold

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// prGuardScriptRel é o caminho, dentro do projeto, do script que o hook PreToolUse executa.
const prGuardScriptRel = "scripts/pr_guard.py"

// prGuardTemplate é o script embutido; é cópia de scripts/pr_guard.py do repositório (um teste
// garante a igualdade) e não passa por text/template.
const prGuardTemplate = "templates/scripts/pr_guard.py.tmpl"

// prGuardMarker marca o script gerado pelo forge-sdd; sem ele o arquivo é do usuário.
const prGuardMarker = "# forge-sdd:generated v"

// PRGuardScriptPath devolve o caminho do script do guard no projeto.
func PRGuardScriptPath(targetDir string) string {
	return filepath.Join(targetDir, filepath.FromSlash(prGuardScriptRel))
}

// prGuardMarkerVersion devolve a versão do marcador e se o arquivo foi gerado pelo forge-sdd.
// O marcador só vale nas três primeiras linhas (entre o shebang e o bloco PEP 723).
func prGuardMarkerVersion(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for i := 0; i < 3 && sc.Scan(); i++ {
		if v, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), prGuardMarker); ok {
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

// IsGeneratedPRGuard diz se scripts/pr_guard.py existe e foi gerado pelo forge-sdd.
func IsGeneratedPRGuard(targetDir string) bool {
	_, ok := prGuardMarkerVersion(PRGuardScriptPath(targetDir))
	return ok
}

// PRGuardUserOwned diz se scripts/pr_guard.py existe sem o marcador (do usuário, intocado).
func PRGuardUserOwned(targetDir string) bool {
	path := PRGuardScriptPath(targetDir)
	if _, err := os.Lstat(path); err != nil {
		return false
	}
	return !IsGeneratedPRGuard(targetDir)
}

// renderPRGuard devolve o template com o marcador entre o shebang e o bloco PEP 723.
func renderPRGuard(version string) ([]byte, error) {
	data, err := templatesFS.ReadFile(prGuardTemplate)
	if err != nil {
		return nil, fmt.Errorf("ler template %s: %w", prGuardTemplate, err)
	}
	nl := bytes.IndexByte(data, '\n')
	if nl < 0 || !bytes.HasPrefix(data, []byte("#!")) {
		return nil, fmt.Errorf("template %s sem shebang", prGuardTemplate)
	}
	var buf bytes.Buffer
	buf.Write(data[:nl+1])
	buf.WriteString(prGuardMarker + version + "\n")
	buf.Write(data[nl+1:])
	return buf.Bytes(), nil
}

// installPRGuardScript grava scripts/pr_guard.py (atomicamente, SEG-02) quando falta ou quando foi
// gerado por uma versão menor. Script sem marcador é do usuário e fica intocado. Devolve true se
// gravou (ou, em dryRun, se gravaria).
func installPRGuardScript(targetDir, version string, dryRun bool) (bool, error) {
	path := PRGuardScriptPath(targetDir)
	if _, err := os.Lstat(path); err == nil {
		v, gerado := prGuardMarkerVersion(path)
		if !gerado || config.CompareVersions(v, version) >= 0 {
			return false, nil
		}
	}
	data, err := renderPRGuard(version)
	if err != nil {
		return false, err
	}
	if dryRun {
		return true, checkSafeTarget(targetDir, path)
	}
	if err := writeFileSafeMode(targetDir, path, data, 0o755); err != nil {
		return false, err
	}
	return true, nil
}

// installPRGuard entrega o script e só então mescla o hook PreToolUse: se o script não puder ser
// gravado, o hook não é instalado (o hook sem script bloquearia todo Bash). Devolve o que mudou.
func installPRGuard(targetDir, version string, dryRun bool) (changed []string, err error) {
	if ok, err := installPRGuardScript(targetDir, version, dryRun); err != nil {
		return nil, fmt.Errorf("%s não gravado, hook PreToolUse não instalado: %w", prGuardScriptRel, err)
	} else if ok {
		changed = append(changed, prGuardScriptRel)
	}
	ok, err := mergeClaudeHookWithMatcher(targetDir, "PreToolUse", prGuardMatcher, prGuardCommand, prGuardMarkers, dryRun)
	if err != nil {
		return changed, err
	}
	if ok {
		changed = append(changed, "hook PreToolUse")
	}
	return changed, nil
}
