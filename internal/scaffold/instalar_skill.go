package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/skills"
)

// Tipos instaláveis (ADR-0025): só mudam o prefixo e o arquivo-marca de "já instalado".
const (
	TipoSkill  = "skill"
	TipoPlugin = "plugin"
)

// Resultado descreve uma instalação. DisponivelEm preenchido significa no-op: a skill ou o plugin
// já está no projeto aberto ou no hub, e nada foi gravado.
type Resultado struct {
	Gravados     []string
	Preservados  []string
	DisponivelEm string
}

func prefixoInstalacao(tipo string) (prefixo, marca string, err error) {
	switch tipo {
	case TipoSkill:
		return ".agents/skills", "SKILL.md", nil
	case TipoPlugin:
		return ".agents/plugins", ".claude-plugin/plugin.json", nil
	}
	return "", "", fmt.Errorf("tipo %q inválido: use %q ou %q", tipo, TipoSkill, TipoPlugin)
}

func temSddrc(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, layout.SddDirName, layout.SddrcFileName))
	return err == nil && fi.Mode().IsRegular()
}

// ListarDestinosInstalacao devolve o projeto aberto e, se houver, os subprojetos com forge
// (filhos diretos com sdd/.sddrc), ordenados. Link simbólico nunca é destino.
func ListarDestinosInstalacao(projetoAberto string) ([]string, error) {
	raiz, err := filepath.Abs(projetoAberto)
	if err != nil {
		return nil, err
	}
	if fi, err := os.Stat(raiz); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("projeto aberto %q não é um diretório", projetoAberto)
	}
	entradas, err := os.ReadDir(raiz)
	if err != nil {
		return nil, err
	}
	var subs []string
	for _, e := range entradas {
		if e.IsDir() && temSddrc(filepath.Join(raiz, e.Name())) {
			subs = append(subs, filepath.Join(raiz, e.Name()))
		}
	}
	sort.Strings(subs)
	return append([]string{raiz}, subs...), nil
}

func instalado(dir, prefixo, marca, slug string) bool {
	fi, err := os.Stat(filepath.Join(dir, filepath.FromSlash(prefixo), slug, filepath.FromSlash(marca)))
	return err == nil && fi.Mode().IsRegular()
}

// DisponivelViaHub diz onde a skill ou o plugin já está disponível: no projeto aberto ou no
// primeiro ancestral com sdd/.sddrc (o hub) que o tenha instalado. A herança é só de leitura.
func DisponivelViaHub(projetoAberto, tipo, slug string) (string, bool) {
	prefixo, marca, err := prefixoInstalacao(tipo)
	if err != nil || skills.ValidateSlug(slug) != nil {
		return "", false
	}
	dir, err := filepath.Abs(projetoAberto)
	if err != nil {
		return "", false
	}
	if instalado(dir, prefixo, marca, slug) {
		return dir, true
	}
	for pai := filepath.Dir(dir); pai != dir; dir, pai = pai, filepath.Dir(pai) {
		if temSddrc(pai) && instalado(pai, prefixo, marca, slug) {
			return pai, true
		}
	}
	return "", false
}

// InstalarSkill instala em <destino>/.agents/skills/<slug>/.
func InstalarSkill(projetoAberto, destino, slug string, arquivos map[string][]byte) (Resultado, error) {
	return instalar(TipoSkill, projetoAberto, destino, slug, arquivos)
}

// InstalarPlugin instala em <destino>/.agents/plugins/<slug>/.
func InstalarPlugin(projetoAberto, destino, slug string, arquivos map[string][]byte) (Resultado, error) {
	return instalar(TipoPlugin, projetoAberto, destino, slug, arquivos)
}

func instalar(tipo, projetoAberto, destino, slug string, arquivos map[string][]byte) (Resultado, error) {
	var res Resultado
	prefixo, _, err := prefixoInstalacao(tipo)
	if err != nil {
		return res, err
	}
	if err := skills.ValidateSlug(slug); err != nil {
		return res, err
	}
	destinoAbs, err := validarDestino(projetoAberto, destino)
	if err != nil {
		return res, err
	}
	if tipo == TipoPlugin {
		// ADR-0030: plugin que é mod só entra com manifesto e permissões completos; nada é gravado antes.
		if err := skills.ValidarMod(slug, arquivos, config.BinaryVersion()); err != nil {
			return res, err
		}
	}
	if onde, ok := DisponivelViaHub(projetoAberto, tipo, slug); ok {
		res.DisponivelEm = onde
		return res, nil
	}
	if len(arquivos) == 0 {
		return res, fmt.Errorf("nenhum arquivo para instalar em %s/%s", prefixo, slug)
	}
	base := filepath.Join(destinoAbs, filepath.FromSlash(prefixo), slug)
	plano, err := planejar(destinoAbs, base, arquivos)
	if err != nil {
		return res, err
	}
	for _, a := range plano {
		if a.existe {
			res.Preservados = append(res.Preservados, a.caminho)
			continue
		}
		if err := WriteFileSafe(destinoAbs, a.caminho, arquivos[a.chave]); err != nil {
			return res, err
		}
		res.Gravados = append(res.Gravados, a.caminho)
	}
	return res, nil
}

func validarDestino(projetoAberto, destino string) (string, error) {
	lista, err := ListarDestinosInstalacao(projetoAberto)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(destino)
	if err != nil {
		return "", err
	}
	for _, d := range lista {
		if d == abs {
			return d, nil
		}
	}
	return "", fmt.Errorf("destino %q não é o projeto aberto nem um subprojeto com forge (SEG-02)", destino)
}

type itemPlano struct {
	chave, caminho string
	existe         bool
}

// planejar valida TODAS as chaves e alvos antes de gravar o primeiro arquivo (nada parcial).
func planejar(destino, base string, arquivos map[string][]byte) ([]itemPlano, error) {
	chaves := make([]string, 0, len(arquivos))
	for k := range arquivos {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	plano := make([]itemPlano, 0, len(chaves))
	for _, k := range chaves {
		if err := validarChave(k); err != nil {
			return nil, err
		}
		caminho := filepath.Join(base, filepath.FromSlash(strings.ReplaceAll(k, `\`, "/")))
		if err := checkSafeTarget(destino, caminho); err != nil {
			return nil, err
		}
		_, errStat := os.Lstat(caminho)
		plano = append(plano, itemPlano{chave: k, caminho: caminho, existe: errStat == nil})
	}
	return plano, nil
}

func validarChave(k string) error {
	n := strings.ReplaceAll(k, `\`, "/")
	if n == "" || strings.HasPrefix(n, "/") || filepath.IsAbs(k) || filepath.VolumeName(k) != "" || strings.Contains(n, ":") {
		return fmt.Errorf("chave %q deve ser caminho relativo (SEG-01)", k)
	}
	for _, parte := range strings.Split(n, "/") {
		if parte == ".." || parte == "." || parte == "" {
			return fmt.Errorf("chave %q contém segmento inválido (SEG-01)", k)
		}
	}
	return nil
}
