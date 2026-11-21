package layout

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// EnvInitiative é a variável de ambiente que fixa a iniciativa, abaixo apenas da flag --iniciativa.
const EnvInitiative = "FORGE_INICIATIVA"

// maxBranchLen limita o nome da branch aceito pelo casamento por prefixo numérico (SEG-01).
const maxBranchLen = 200

// ErrAmbiguousInitiative indica que o número da branch casa com mais de uma pasta de iniciativa.
var ErrAmbiguousInitiative = errors.New("iniciativa ambígua")

// numericBranch ancora a branch por feature: <feat|fix>/<N>-<slug>, com N de 1 a 6 dígitos.
var numericBranch = regexp.MustCompile(`^(?:feat|fix)/([0-9]{1,6})-[a-z0-9-]+$`)

var (
	overrideMu sync.RWMutex
	override   string
)

// SetOverride fixa a iniciativa vinda da flag --iniciativa (vazio limpa). A validação ocorre em
// ResolveInitiative, que é quem conhece a pasta.
func SetOverride(id string) {
	overrideMu.Lock()
	defer overrideMu.Unlock()
	override = id
}

func getOverride() string {
	overrideMu.RLock()
	defer overrideMu.RUnlock()
	return override
}

// RequireInitiative falha quando a iniciativa não pode ser resolvida sem ambiguidade (ou quando a
// flag/variável aponta para uma pasta inexistente). Os comandos que leem a fila a chamam antes.
func RequireInitiative(dir string) error {
	_, _, err := ResolveInitiative(dir)
	return err
}

// ResolveInitiative resolve a iniciativa ativa de dir, nesta ordem: flag (SetOverride), variável
// FORGE_INICIATIVA, pasta de nome igual ao da branch e pasta cujo número casa com o da branch.
// Sem iniciativa devolve id vazio e erro nulo (fila legada); ambiguidade devolve
// ErrAmbiguousInitiative com as pastas que casaram.
func ResolveInitiative(dir string) (id, path string, err error) {
	if v := getOverride(); v != "" {
		return explicitInitiative(dir, v, "--iniciativa")
	}
	if v := os.Getenv(EnvInitiative); v != "" {
		return explicitInitiative(dir, v, EnvInitiative)
	}
	branch, err := CurrentBranch(dir)
	if err != nil {
		return "", "", nil
	}
	if exact, ok := FromBranch(branch); ok {
		if p, ok := existingInitiative(dir, exact); ok {
			return exact, p, nil
		}
	}
	id, err = initiativeOfBranch(dir, branch)
	if err != nil || id == "" {
		return "", "", err
	}
	path, err = Initiative(dir, id)
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}

// InitiativeOfBranch devolve a iniciativa de uma branch informada: a pasta de nome igual, depois a
// pasta de mesmo número; sem pasta, o slug da branch (FromBranch), como antes. ok=false quando a
// branch não aponta para iniciativa.
func InitiativeOfBranch(dir, branch string) (id string, ok bool, err error) {
	exact, exactOK := FromBranch(branch)
	if exactOK {
		if _, found := existingInitiative(dir, exact); found {
			return exact, true, nil
		}
	}
	id, err = initiativeOfBranch(dir, branch)
	if err != nil {
		return "", false, err
	}
	if id != "" {
		return id, true, nil
	}
	return exact, exactOK, nil
}

func explicitInitiative(dir, v, origem string) (id, path string, err error) {
	path, err = Initiative(dir, v)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", origem, err)
	}
	if _, ok := existingInitiative(dir, v); !ok {
		return "", "", fmt.Errorf("%s: a iniciativa %q não existe em %s; corrija ou remova %s", origem, v, Discovery(dir), origem)
	}
	return v, path, nil
}

func existingInitiative(dir, id string) (string, bool) {
	p, err := Initiative(dir, id)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(p); err != nil || !info.IsDir() {
		return "", false
	}
	return p, true
}

// initiativeOfBranch casa o número de uma branch por feature com as pastas reais de
// sdd/discovery. Devolve "" quando a branch não é por feature ou não há pasta; erro de
// ambiguidade quando há mais de uma. O id sai sempre da listagem de pastas (SEG-01).
func initiativeOfBranch(dir, branch string) (string, error) {
	if len(branch) > maxBranchLen {
		return "", nil
	}
	m := numericBranch.FindStringSubmatch(branch)
	if m == nil {
		return "", nil
	}
	entries, err := os.ReadDir(Discovery(dir))
	if err != nil {
		return "", nil
	}
	var found []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || !ValidInitiative(name) {
			continue
		}
		if first, _, _ := strings.Cut(name, "-"); first == m[1] {
			found = append(found, name)
		}
	}
	switch len(found) {
	case 0:
		return "", nil
	case 1:
		return found[0], nil
	}
	return "", fmt.Errorf("%w: a branch %q casa com: %s; use --iniciativa <id> ou %s", ErrAmbiguousInitiative, branch, strings.Join(found, ", "), EnvInitiative)
}
