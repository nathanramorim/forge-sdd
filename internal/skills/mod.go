package skills

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/forge-sdd/cli/internal/config"
)

// Contrato de mod do Claude Code (ADR-0030). O validador só lê arquivos que já estão em memória
// (a origem da instalação): não executa nada, não fala com a rede e não grava.
const (
	// ModSchema é a versão do contrato em .claude-plugin/forge.json.
	ModSchema = "forge-sdd/mod/1.0"

	modPluginJSON = ".claude-plugin/plugin.json"
	modForgeJSON  = ".claude-plugin/forge.json"
	modHooksJSON  = "hooks/hooks.json"

	// maxModManifestBytes limita cada manifesto lido (SEG-01).
	maxModManifestBytes = 64 * 1024
)

var (
	semverModRe = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?$`)
	hookModRe   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9.]{0,63}$`)
	envModRe    = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)
	// semverBinarioRe reconhece uma versão de binário comparável; "dev" e afins ficam de fora.
	semverBinarioRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)

	// Pastas do projeto que um mod nunca pode declarar como escrita ou leitura.
	pastasVedadasMod = map[string]bool{".git": true, ".claude": true, ".agents": true}
)

// PermissoesMod é o bloco `permissions` do contrato. Network e Hooks são obrigatórios; Network é
// ponteiro para distinguir "false declarado" de "ausente".
type PermissoesMod struct {
	Network *bool    `json:"network"`
	Model   *bool    `json:"model"`
	Read    []string `json:"read"`
	Write   []string `json:"write"`
	Env     []string `json:"env"`
	Hooks   []string `json:"hooks"`
}

// ContratoMod é o conteúdo de .claude-plugin/forge.json.
type ContratoMod struct {
	Schema        string        `json:"schema"`
	MinClaudeCode string        `json:"minClaudeCode"`
	MinForge      string        `json:"minForge"`
	Permissions   PermissoesMod `json:"permissions"`
}

type manifestoPluginMod struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type hooksJSONMod struct {
	Modules []string `json:"modules"`
}

func normalizarChaveMod(k string) string {
	return strings.TrimPrefix(strings.ReplaceAll(k, `\`, "/"), "./")
}

func lerArquivoMod(arquivos map[string][]byte, nome string) ([]byte, bool) {
	for k, v := range arquivos {
		if normalizarChaveMod(k) == nome {
			return v, true
		}
	}
	return nil, false
}

// EhMod diz se o plugin é um mod do Claude Code: traz módulo de ganchos (hooks/hooks.json com
// `modules`) ou declara o contrato do forge. Plugin só de skills e comandos não é mod (ADR-0030).
func EhMod(arquivos map[string][]byte) bool {
	if _, ok := lerArquivoMod(arquivos, modForgeJSON); ok {
		return true
	}
	if bruto, ok := lerArquivoMod(arquivos, modHooksJSON); ok {
		var h hooksJSONMod
		if json.Unmarshal(bruto, &h) == nil && len(h.Modules) > 0 {
			return true
		}
	}
	return false
}

// ValidarMod aplica o contrato da ADR-0030 a um plugin ainda não gravado. Devolve nil quando o
// plugin não é mod. versaoForge é a versão do binário (config.BinaryVersion); versão que não é
// X.Y.Z, como "dev", não é comparada com minForge. Todos os problemas voltam juntos.
func ValidarMod(slug string, arquivos map[string][]byte, versaoForge string) error {
	if !EhMod(arquivos) {
		return nil
	}
	var erros []error
	adicionar := func(formato string, a ...any) { erros = append(erros, fmt.Errorf(formato, a...)) }

	// plugin.json: campos do Claude Code que o forge exige.
	bruto, ok := lerArquivoMod(arquivos, modPluginJSON)
	switch {
	case !ok:
		adicionar("%s ausente", modPluginJSON)
	case len(bruto) > maxModManifestBytes:
		adicionar("%s acima de %d bytes", modPluginJSON, maxModManifestBytes)
	default:
		var m manifestoPluginMod
		if err := json.Unmarshal(bruto, &m); err != nil {
			adicionar("%s inválido: %v", modPluginJSON, err)
		} else {
			if m.Name != slug {
				adicionar("%s: name %q deve ser igual ao slug %q", modPluginJSON, m.Name, slug)
			}
			if !semverModRe.MatchString(m.Version) {
				adicionar("%s: version %q deve ser X.Y.Z", modPluginJSON, m.Version)
			}
			if strings.TrimSpace(m.Description) == "" {
				adicionar("%s: description é obrigatória", modPluginJSON)
			}
		}
	}

	// forge.json: o contrato do forge.
	bruto, ok = lerArquivoMod(arquivos, modForgeJSON)
	switch {
	case !ok:
		adicionar("%s ausente: o mod precisa declarar permissões e versões mínimas (ADR-0030)", modForgeJSON)
	case len(bruto) > maxModManifestBytes:
		adicionar("%s acima de %d bytes", modForgeJSON, maxModManifestBytes)
	default:
		erros = append(erros, validarContratoMod(bruto, versaoForge)...)
	}

	// Módulos de ganchos citados precisam existir no pacote.
	if bruto, ok := lerArquivoMod(arquivos, modHooksJSON); ok {
		var h hooksJSONMod
		if err := json.Unmarshal(bruto, &h); err != nil {
			adicionar("%s inválido: %v", modHooksJSON, err)
		}
		for _, mod := range h.Modules {
			alvo := path.Join("hooks", normalizarChaveMod(mod))
			if _, existe := lerArquivoMod(arquivos, alvo); !existe || strings.HasPrefix(alvo, "..") {
				adicionar("%s cita o módulo %q, que o plugin não traz", modHooksJSON, mod)
			}
		}
	}

	if len(erros) == 0 {
		return nil
	}
	return fmt.Errorf("mod %q recusado pelo contrato da ADR-0030: %w", slug, errors.Join(erros...))
}

func validarContratoMod(bruto []byte, versaoForge string) []error {
	var erros []error
	adicionar := func(formato string, a ...any) { erros = append(erros, fmt.Errorf(formato, a...)) }

	var c ContratoMod
	dec := json.NewDecoder(bytes.NewReader(bruto))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return []error{fmt.Errorf("%s inválido: %v", modForgeJSON, err)}
	}
	if c.Schema != ModSchema {
		adicionar("%s: schema %q desconhecido (esperado %q)", modForgeJSON, c.Schema, ModSchema)
	}
	if !semverModRe.MatchString(c.MinClaudeCode) {
		adicionar("%s: minClaudeCode %q ausente ou fora de X.Y.Z", modForgeJSON, c.MinClaudeCode)
	}
	switch {
	case !semverModRe.MatchString(c.MinForge):
		adicionar("%s: minForge %q ausente ou fora de X.Y.Z", modForgeJSON, c.MinForge)
	case semverBinarioRe.MatchString(versaoForge) && config.CompareVersions(c.MinForge, versaoForge) > 0:
		adicionar("%s: o mod exige forge %s ou mais novo (em uso: %s)", modForgeJSON, c.MinForge, versaoForge)
	}
	p := c.Permissions
	if p.Network == nil {
		adicionar("%s: permissions.network é obrigatório (true ou false)", modForgeJSON)
	}
	if p.Write == nil {
		adicionar("%s: permissions.write é obrigatório (lista, vazia se o mod não grava)", modForgeJSON)
	}
	if len(p.Hooks) == 0 {
		adicionar("%s: permissions.hooks é obrigatório e não pode ser vazio", modForgeJSON)
	}
	for _, h := range p.Hooks {
		if !hookModRe.MatchString(h) {
			adicionar("%s: gancho %q inválido em permissions.hooks", modForgeJSON, h)
		}
	}
	for _, par := range []struct {
		campo string
		lista []string
	}{{"write", p.Write}, {"read", p.Read}} {
		for _, caminho := range par.lista {
			if err := validarCaminhoDoMod(caminho); err != nil {
				adicionar("%s: permissions.%s: %v", modForgeJSON, par.campo, err)
			}
		}
	}
	for _, v := range p.Env {
		if !envModRe.MatchString(v) {
			adicionar("%s: variável %q inválida em permissions.env", modForgeJSON, v)
		}
	}
	return erros
}

// validarCaminhoDoMod aceita só caminho relativo ao projeto, sem `..` nem `~`, fora de .git,
// .claude e .agents (SEG-01, SEG-02).
func validarCaminhoDoMod(p string) error {
	n := strings.ReplaceAll(p, `\`, "/")
	if strings.TrimSpace(n) == "" {
		return fmt.Errorf("caminho vazio")
	}
	if strings.HasPrefix(n, "/") || strings.HasPrefix(n, "~") || strings.Contains(n, ":") {
		return fmt.Errorf("%q deve ser relativo ao projeto", p)
	}
	partes := strings.Split(strings.TrimSuffix(n, "/"), "/")
	for _, parte := range partes {
		if parte == ".." || parte == "." || parte == "" {
			return fmt.Errorf("%q tem segmento inválido", p)
		}
	}
	if pastasVedadasMod[partes[0]] {
		return fmt.Errorf("%q aponta para %s, vedado a mods", p, partes[0])
	}
	return nil
}
