package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// NaNuvem diz se o processo roda numa sessão do Claude Code na nuvem.
func NaNuvem(getenv func(string) string) bool {
	if getenv == nil {
		getenv = os.Getenv
	}
	return getenv("CLAUDE_CODE_REMOTE") == "true"
}

// ReposFaltando devolve os repositórios do hub (nuvem.repos) que não estão clonados ao lado do
// projeto. Na nuvem, cada repositório do ambiente fica numa pasta irmã com o nome dele.
func ReposFaltando(dir string, repos []string) []string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	pai := filepath.Dir(abs)
	var falta []string
	for _, r := range repos {
		nome := path.Base(r)
		if nome == filepath.Base(abs) {
			continue
		}
		if fi, err := os.Stat(filepath.Join(pai, nome, ".git")); err != nil || fi == nil {
			falta = append(falta, r)
		}
	}
	return falta
}

// ChecklistNuvem é o que só a pessoa faz no claude.ai, uma vez por ambiente da nuvem.
func ChecklistNuvem(n config.NuvemConfig, projeto string) string {
	var b strings.Builder
	b.WriteString("No claude.ai, uma vez por ambiente da nuvem (Claude Code › ambiente › configurações):\n")
	repos := projeto
	if len(n.Repos) > 0 {
		repos += ", " + strings.Join(n.Repos, ", ")
	}
	fmt.Fprintf(&b, "  1. Repositórios: %s.\n", repos)
	if n.InstalarEfetivo() == config.NuvemInstalarGit {
		b.WriteString("  2. Rede: github.com e proxy.golang.org liberados (a política padrão já libera). O ambiente precisa de git e Go.\n")
	} else {
		b.WriteString("  2. Rede: registry.npmjs.org e github.com liberados (a política padrão já libera). O ambiente precisa de Node.\n")
	}
	b.WriteString("  3. Segredos: nenhum. O hook só roda comandos fixos e fica versionado no repositório.\n")
	b.WriteString("Depois, faça commit de .claude/settings.json e .claude/hooks/forge-nuvem.sh: a próxima sessão na nuvem já sobe a esteira.\n")
	return b.String()
}

var duracaoNoResumoRe = regexp.MustCompile(`há [0-9hms ]+`)

// assinaturaDoResumo ignora o que muda a cada segundo (durações, disco e memória) para o resumo só
// reaparecer quando a esteira de fato mudou.
func assinaturaDoResumo(texto string) string {
	var b strings.Builder
	for _, l := range strings.Split(texto, "\n") {
		if strings.HasPrefix(l, "disco:") || strings.HasPrefix(l, "memória:") || strings.HasPrefix(l, "(retrato") {
			continue
		}
		b.WriteString(duracaoNoResumoRe.ReplaceAllString(l, "há _"))
		b.WriteByte('\n')
	}
	h := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(h[:8])
}

// ResumoNuvem devolve o diagnóstico da esteira só quando ele mudou desde a última chamada (a
// assinatura fica em .forge/runtime/painel/resumo.sig); texto vazio quando nada mudou.
func ResumoNuvem(dir string, now time.Time) string {
	texto := FormatDiagnostico(dir, now)
	if texto == "" {
		return ""
	}
	sig := assinaturaDoResumo(texto)
	arq := filepath.Join(layout.ForgeRuntime(dir), "painel", "resumo.sig")
	if ant, err := os.ReadFile(arq); err == nil && strings.TrimSpace(string(ant)) == sig {
		return ""
	}
	if err := os.MkdirAll(filepath.Dir(arq), 0o755); err == nil {
		_ = os.WriteFile(arq, []byte(sig), 0o600)
	}
	return texto
}
