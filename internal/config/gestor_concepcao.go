package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Modos do gestor ativo na concepção (feat 113-02).
const (
	GestorDesligado = "desligado"
	GestorGitHub    = "github"
	GestorMCP       = "mcp"
)

const maxArquivoGestorBytes = 256 << 10

var provedoresMCP = []string{"clickup", "jira", "azure-devops"}

// GestorConcepcao é a decisão do resolvedor: Modo e, no modo mcp, o Provider.
type GestorConcepcao struct {
	Modo     string
	Provider string
}

// ResolverGestorConcepcao decide o gestor ativo na concepção. Sem create_on_conception
// e sem sdd/modulos/project.json o resultado é desligado, sem efeito externo. vcs é o VCS da
// Constituição (vcs.LeVCS), recebido por parâmetro porque internal/vcs não pode ser importado aqui.
// Conteúdo de mcps.md é entrada não confiável: só o valor exato "ativo" liga (SEG-01).
func ResolverGestorConcepcao(dir, vcs string) (GestorConcepcao, error) {
	off := GestorConcepcao{Modo: GestorDesligado}
	// Sem forge.json vale o padrão (provider auto); project.json ainda pode ativar o gestor GitHub.
	cfg, _, err := LoadForgeConfig(dir)
	if err != nil {
		return off, err
	}
	prov := cfg.PMMCP.Provider
	if prov != "auto" && prov != "github" && prov != "github-projects" && !contains(provedoresMCP, prov) {
		return off, fmt.Errorf("pm_mcp.provider %q não suportado na concepção", prov)
	}
	if !cfg.PMMCP.CreateOnConception && !arquivoRegularConfinado(dir, filepath.Join("sdd", "modulos", "project.json")) {
		return off, nil
	}
	if contains(provedoresMCP, prov) {
		if LerStatusMCPs(dir)[prov] {
			return GestorConcepcao{Modo: GestorMCP, Provider: prov}, nil
		}
		return off, nil
	}
	if strings.EqualFold(strings.TrimSpace(vcs), "github") {
		return GestorConcepcao{Modo: GestorGitHub}, nil
	}
	return off, nil
}

// LerStatusMCPs lê a tabela de sdd/memory/mcps.md e devolve os MCPs com status exatamente `ativo`.
// Arquivo ausente, fora do projeto ou malformado resulta em mapa vazio, nunca em erro.
func LerStatusMCPs(dir string) map[string]bool {
	ativos := map[string]bool{}
	data, err := lerConfinado(dir, filepath.Join("sdd", "memory", "mcps.md"))
	if err != nil {
		return ativos
	}
	for _, linha := range strings.Split(string(data), "\n") {
		linha = strings.TrimSpace(linha)
		if !strings.HasPrefix(linha, "|") {
			continue
		}
		c := strings.Split(strings.Trim(linha, "|"), "|")
		if len(c) < 2 {
			continue
		}
		if strings.TrimSpace(c[1]) == "ativo" {
			ativos[strings.ToLower(strings.TrimSpace(c[0]))] = true
		}
	}
	return ativos
}

// caminhoConfinado resolve rel dentro de dir e recusa saída do projeto por symlink.
func caminhoConfinado(dir, rel string) (string, error) {
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(filepath.Join(root, rel))
	if err != nil {
		return "", err
	}
	back, err := filepath.Rel(root, real)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", errors.New("caminho fora do projeto")
	}
	fi, err := os.Stat(real)
	if err != nil || !fi.Mode().IsRegular() {
		return "", errors.New("não é arquivo regular")
	}
	return real, nil
}

func arquivoRegularConfinado(dir, rel string) bool {
	_, err := caminhoConfinado(dir, rel)
	return err == nil
}

func lerConfinado(dir, rel string) ([]byte, error) {
	p, err := caminhoConfinado(dir, rel)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, maxArquivoGestorBytes))
}
