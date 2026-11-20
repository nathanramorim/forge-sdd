package gate

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// CriteriaHeading é a única seção da Spec de onde saem os critérios executáveis.
const CriteriaHeading = "## Critérios de Aceitação Executáveis"

// NegativeMarker marca, no bloco de critérios, a linha de teste negativo (caso que deve falhar e falha).
const NegativeMarker = "# negativo:"

// criteriaExecTimeout limita a execução do bloco de critérios com bash -e.
const criteriaExecTimeout = 10 * time.Minute

var (
	// grepCountRe acha grep com a opção -c (isolada ou combinada, ex.: -qc, -cE) ou --count.
	grepCountRe = regexp.MustCompile(`\bgrep\b[^|;&]*\s(?:-[A-Za-z]*c[A-Za-z]*|--count)\b`)
	// pipeRe acha um pipe simples, sem confundir com o operador ||.
	pipeRe = regexp.MustCompile(`(^|[^|])\|([^|]|$)`)
	// bareNegationRe acha `! comando` que não termina em `|| exit`: o bash -e ignora o status de `!`.
	bareNegationRe = regexp.MustCompile(`^!\s`)
	// trailingGrepVRe acha pipeline que termina em grep -v: sai com 1 quando não imprime nada.
	trailingGrepVRe = regexp.MustCompile(`\|\s*grep\s+(?:-[A-Za-z]*v[A-Za-z]*)(?:\s[^|]*)?$`)
)

// ExtractCriteriaBlock devolve as linhas dos blocos de código da seção "## Critérios de Aceitação
// Executáveis", e só dela: blocos de outras seções (ex.: "Como regenerar") ficam de fora.
func ExtractCriteriaBlock(specContent string) (string, bool) {
	var out []string
	inSection, inFence, found := false, false, false
	for _, line := range strings.Split(specContent, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inFence && strings.HasPrefix(trimmed, "## ") {
			inSection = trimmed == CriteriaHeading
			continue
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			if inFence {
				found = true
			}
			continue
		}
		if inFence {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n"), found
}

// CriteriaProblems aplica as checagens estáticas ao bloco de critérios: sintaxe (bash -n),
// presença do marcador de teste negativo e padrões ruins conhecidos.
func CriteriaProblems(block string) []string {
	var problems []string
	if err := runBash(context.Background(), "", "-n", block); err != nil {
		problems = append(problems, fmt.Sprintf("bash -n reprovou o bloco: %v", err))
	}

	hasPipefail, hasPipe, negativeIdx := false, false, -1
	lines := strings.Split(block, "\n")
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, NegativeMarker) || strings.Contains(line, " "+NegativeMarker) {
			negativeIdx = i
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "set -o pipefail") || strings.Contains(line, "set -eo pipefail") {
			hasPipefail = true
		}
		if pipeRe.MatchString(line) {
			hasPipe = true
		}
		if grepCountRe.MatchString(line) {
			problems = append(problems, fmt.Sprintf("linha %d usa grep -c, que conta linhas e não ocorrências (num arquivo de uma linha dá 1); use grep -o ... | wc -l", i+1))
		}
		if bareNegationRe.MatchString(line) && !strings.Contains(line, "|| exit") {
			problems = append(problems, fmt.Sprintf("linha %d nega com '!' sem '|| exit 1': o bash -e ignora o status de '!' e a linha nunca falha", i+1))
		}
		if trailingGrepVRe.MatchString(line) && !strings.Contains(line, "|| true") {
			problems = append(problems, fmt.Sprintf("linha %d termina em grep -v: só passa quando imprime algo e quebra sob bash -e quando a saída é vazia", i+1))
		}
	}
	if hasPipe && !hasPipefail {
		problems = append(problems, "o bloco usa pipeline sem 'set -o pipefail': a falha do primeiro comando fica escondida")
	}
	if negativeIdx < 0 {
		problems = append(problems, fmt.Sprintf("falta a linha de teste negativo (marcador %q seguido do comando que prova a falha)", NegativeMarker))
	} else if !hasCommandFrom(lines, negativeIdx) {
		problems = append(problems, fmt.Sprintf("o marcador %q não é seguido de comando", NegativeMarker))
	}
	return problems
}

// hasCommandFrom indica se a linha idx (marcador inline) ou alguma seguinte traz um comando.
func hasCommandFrom(lines []string, idx int) bool {
	if !strings.HasPrefix(strings.TrimSpace(lines[idx]), "#") {
		return true
	}
	for _, l := range lines[idx+1:] {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			return true
		}
	}
	return false
}

// RunCriteriaBlock executa o bloco de critérios com bash -e -o pipefail no diretório dir.
func RunCriteriaBlock(dir, block string) error {
	ctx, cancel := context.WithTimeout(context.Background(), criteriaExecTimeout)
	defer cancel()
	return runBash(ctx, dir, "-e -o pipefail", block)
}

func runBash(ctx context.Context, dir, flags, script string) error {
	args := append(strings.Fields(flags), "-c", script)
	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		tail := strings.TrimSpace(string(out))
		if len(tail) > 300 {
			tail = tail[len(tail)-300:]
		}
		return fmt.Errorf("bash %s: %w: %s", flags, err, tail)
	}
	return nil
}

// checkP25Criteria valida a qualidade dos critérios executáveis da Spec (TST-01): extração só da
// seção de critérios, bash -n, teste negativo e padrões ruins; com execute, roda o bloco sob bash -e.
func checkP25Criteria(dir, specPath, specContent string, specFound, execute bool) CheckResult {
	res := CheckResult{ID: "P-25", Rule: "TST-01"}
	if !specFound {
		res.Evidence = fmt.Sprintf("spec %q ausente; impossível verificar os critérios", specPath)
		res.Remediation = "Criar a Spec com a seção '" + CriteriaHeading + "'."
		return res
	}
	block, ok := ExtractCriteriaBlock(specContent)
	if !ok {
		res.Evidence = fmt.Sprintf("spec %q sem bloco de código na seção '%s'", specPath, CriteriaHeading)
		res.Remediation = "Adicionar um bloco ```bash com os comandos dos critérios nessa seção."
		return res
	}
	problems := CriteriaProblems(block)
	if len(problems) == 0 && execute {
		if err := RunCriteriaBlock(dir, block); err != nil {
			problems = append(problems, fmt.Sprintf("o bloco falhou sob bash -e: %v", err))
		}
	}
	if len(problems) > 0 {
		res.Evidence = strings.Join(problems, "; ")
		res.Remediation = "Corrigir o bloco de critérios: 'set -o pipefail' com pipes, grep -o ... | wc -l em vez de grep -c e uma linha '" + NegativeMarker + "' com o comando do caso que deve falhar."
		return res
	}
	res.Passed = true
	res.Evidence = fmt.Sprintf("critérios de %q verificados (bash -n, teste negativo, sem padrões ruins)", specPath)
	if execute {
		res.Evidence += " e executados com bash -e"
	}
	return res
}
