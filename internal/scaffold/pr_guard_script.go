package scaffold

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

// prGuardPublicados mapeia o sha256 de cada scripts/pr_guard.py publicado pelo forge (sem marcador,
// como o agente o copiava) para a versão. Cada release nova acrescenta o seu hash; builds entre tags
// ficam de fora de propósito. Só a igualdade byte a byte adota o arquivo (SEG-02).
var prGuardPublicados = map[string]string{
	"ae69dab97db8d920920011de249a3678f4bd1f42416b1fa0e1ff2a085efe44f4": "5.2.0-beta",
	"80ada8e16d1403058f28e6a13b5b62aaf36b6637b946f0025900f54db0098946": "5.2.1-beta",
}

// PRGuardAdotavel diz se scripts/pr_guard.py é arquivo regular, sem marcador, byte a byte igual a um
// guard publicado pelo forge, e devolve a versão. Link simbólico nunca é seguido (SEG-01).
func PRGuardAdotavel(targetDir string) (string, bool) {
	path := PRGuardScriptPath(targetDir)
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return "", false
	}
	if _, gerado := prGuardMarkerVersion(path); gerado {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	sum := sha256.Sum256(data)
	v, ok := prGuardPublicados[hex.EncodeToString(sum[:])]
	return v, ok
}

// PRGuardUserOwned diz se scripts/pr_guard.py existe, sem o marcador e sem ser cópia de um guard
// publicado (do usuário, intocado).
func PRGuardUserOwned(targetDir string) bool {
	path := PRGuardScriptPath(targetDir)
	if _, err := os.Lstat(path); err != nil {
		return false
	}
	if IsGeneratedPRGuard(targetDir) {
		return false
	}
	_, adotavel := PRGuardAdotavel(targetDir)
	return !adotavel
}

// prGuardNovoRel é onde o guard novo vai quando o do usuário não pode ser trocado.
const prGuardNovoRel = "scripts/pr_guard.novo.py"

// PRGuardAvisoSemCorrecoes é o aviso de que o guard do usuário não recebe atualização.
const PRGuardAvisoSemCorrecoes = prGuardScriptRel + " é do usuário: a barreira de PR não receberá correções enquanto ele não for revisado; compare com " + prGuardNovoRel

// installPRGuardNovo grava scripts/pr_guard.novo.py ao lado do guard do usuário, só se o conteúdo
// mudou (idempotente), de forma atômica e sem seguir link simbólico.
func installPRGuardNovo(targetDir, version string, dryRun bool) (bool, error) {
	path := filepath.Join(targetDir, filepath.FromSlash(prGuardNovoRel))
	data, err := renderPRGuard(version)
	if err != nil {
		return false, err
	}
	if fi, err := os.Lstat(path); err == nil && fi.Mode().IsRegular() {
		if cur, err := os.ReadFile(path); err == nil && bytes.Equal(cur, data) {
			return false, nil
		}
	}
	if err := checkSafeTarget(targetDir, path); err != nil {
		return false, err
	}
	if dryRun {
		return true, nil
	}
	if err := writeFileSafeMode(targetDir, path, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
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
// gerado por uma versão menor. Script sem marcador, mas byte a byte igual a um guard publicado, é
// adotado; o demais é do usuário e fica intocado. Devolve true se gravou (ou, em dryRun, se gravaria).
func installPRGuardScript(targetDir, version string, dryRun bool) (bool, error) {
	path := PRGuardScriptPath(targetDir)
	if _, err := os.Lstat(path); err == nil {
		v, gerado := prGuardMarkerVersion(path)
		if gerado {
			if config.CompareVersions(v, version) >= 0 {
				return false, nil
			}
		} else if _, adotavel := PRGuardAdotavel(targetDir); !adotavel {
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
	if PRGuardUserOwned(targetDir) {
		if ok, err := installPRGuardNovo(targetDir, version, dryRun); err != nil {
			return changed, fmt.Errorf("%s não gravado: %w", prGuardNovoRel, err)
		} else if ok {
			changed = append(changed, prGuardNovoRel)
		}
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
