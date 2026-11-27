package scaffold

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/eol"
	"github.com/forge-sdd/cli/internal/notify"
)

// Tabela dos templates de instrução da 5.0 (gerados sem marcador). A adoção de um arquivo sem
// marcador só acontece por igualdade byte a byte com o template da 5.0 renderizado com o .sddrc
// do projeto; a feat-118-04 estende esta tabela para rules e schema.json (DUP-06).
//
//go:embed all:v50
var v50FS embed.FS

const generatedMarker = "<!-- forge-sdd:generated -->"

// instrucaoHeader mapeia cada arquivo de instrução ao cabeçalho do seu template.
var instrucaoHeader = map[string]string{
	"CLAUDE.md":                       "# CLAUDE.md — ",
	"GEMINI.md":                       "# GEMINI.md — ",
	"AGENTS.md":                       "# AGENTS.md — ",
	"OPENAI.md":                       "# OPENAI.md — ",
	".github/copilot-instructions.md": "# Copilot Instructions — ",
	".gemini/system_instructions.md":  "# System Instructions — ",
	".openai/system_instructions.md":  "# System Instructions — ",
}

// v50Table: arquivo de instrução -> (nome na pasta v50, sha256 do template da tag v5.0.0).
var v50Table = map[string]struct{ file, sha string }{
	"CLAUDE.md":                       {"CLAUDE.md.tmpl", "acd240030c1807ea1efe97462ba7c7200932b76c58e19ebbc278b3fedd642068"},
	"GEMINI.md":                       {"GEMINI.md.tmpl", "57f7cc02c265f67d41952e925b9d7b3b3b349dee9b17a0d4a63cf71477be67d6"},
	"OPENAI.md":                       {"OPENAI.md.tmpl", "4a0e24267c165266911718b56d2739259b076e0393549b56814aa5d5e9449eff"},
	".github/copilot-instructions.md": {"copilot-instructions.md.tmpl", "db004c07a7f1cc12569ba214230abaed7423f24a570f823ce03bcd19598d66d4"},
	".gemini/system_instructions.md":  {".gemini__system_instructions.md.tmpl", "885090124f113f5c842303a6cc4333909dae97f11111351448931e8590f6f834"},
	".openai/system_instructions.md":  {".openai__system_instructions.md.tmpl", "e6cd16cd7f87c1b5947cf6b35bbff8da653f7edf039b2b68eb06ff9c8b3b5330"},
	".agents/rules/sdd-arquitetura.md": {".agents__rules__sdd-arquitetura.md.tmpl", "c9f87c461025b2ac602333920d9ecf56005e31ef2a3afb4de6e9a61539038971"},
	".agents/rules/sdd-codigo-e-testes.md": {".agents__rules__sdd-codigo-e-testes.md.tmpl", "891321d57f28cfb910f7d82e9be105f049a6e52387bce6a7ef9f7ba7d4a3c154"},
	".agents/rules/sdd-fluxo-e-portao.md": {".agents__rules__sdd-fluxo-e-portao.md.tmpl", "cfee9d39f9a0109bfaaae1a2d443606d360a1022fab3b506613c66086073eb29"},
	".agents/rules/sdd-reuso-e-duplicacao.md": {".agents__rules__sdd-reuso-e-duplicacao.md.tmpl", "3e790c0ad14f20690b5c36b75cc90b7388f3f036085717bca74ea0fa552a84d0"},
	".agents/rules/sdd-seguranca.md": {".agents__rules__sdd-seguranca.md.tmpl", "34aae185420926822bd34b3e369b24a16a072f7a8dcc0db17a3bf11f1836013e"},
	"sdd/.metrics/schema.json": {"sdd__.metrics__schema.json.tmpl", "31fc36e246325dd30451fdca302efc4f069994d423fbdcc34e6a93d794d5bd15"},
}

// instrucaoAgentFiles: arquivos de instrução por agente (copilot incluso).
var instrucaoAgentFiles = map[string][]string{
	config.AgentClaude:  {"CLAUDE.md"},
	config.AgentGemini:  {"GEMINI.md", ".gemini/system_instructions.md"},
	config.AgentOpenAI:  {"AGENTS.md", "OPENAI.md", ".openai/system_instructions.md"},
	config.AgentCopilot: {".github/copilot-instructions.md"},
}

func isInstrucao(rel string) bool {
	_, ok := instrucaoHeader[rel]
	return ok
}

// IsGeneratedInstrucao diz se o arquivo de instrução foi gerado pelo scaffold: cabeçalho conhecido do
// próprio arquivo E o marcador `<!-- forge-sdd:generated -->` nas três primeiras linhas. Só o título
// não basta: um arquivo do usuário pode começar igual. Regra única (DUP-06), usada também pelo AGENTS.md.
func IsGeneratedInstrucao(path string) bool {
	prefix, ok := instrucaoHeaderFor(path)
	if !ok {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	head := strings.SplitN(string(data), "\n", 4)
	if len(head) < 2 || !strings.HasPrefix(head[0], prefix) {
		return false
	}
	for _, l := range head[1:min(len(head), 3)] {
		if strings.TrimSpace(l) == generatedMarker {
			return true
		}
	}
	return false
}

// instrucaoHeaderFor resolve o cabeçalho esperado pelo sufixo do caminho.
func instrucaoHeaderFor(p string) (string, bool) {
	p = filepath.ToSlash(p)
	for rel, h := range instrucaoHeader {
		if p == rel || strings.HasSuffix(p, "/"+rel) {
			return h, true
		}
	}
	return "", false
}

func renderV50(rel string, cfg config.Config) ([]byte, bool) {
	e, ok := v50Table[rel]
	if !ok {
		return nil, false
	}
	src, err := v50FS.ReadFile("v50/" + e.file)
	if err != nil {
		return nil, false
	}
	sum := sha256.Sum256(src)
	if hex.EncodeToString(sum[:]) != e.sha {
		return nil, false
	}
	tmpl, err := template.New(e.file).Parse(string(src))
	if err != nil {
		return nil, false
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

// adotavelDe50 diz se o arquivo sem marcador é idêntico, byte a byte, ao template da 5.0 renderizado com cfg.
func adotavelDe50(rel, dest string, cfg config.Config) bool {
	want, ok := renderV50(rel, cfg)
	if !ok {
		return false
	}
	fi, err := os.Lstat(dest)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	got, err := os.ReadFile(dest)
	return err == nil && eol.Iguais(got, want)
}

// instrucaoDoUsuario: existe, não tem marcador e não é adotável da 5.0 (fica intocado).
func instrucaoDoUsuario(rel, dest string, cfg config.Config) bool {
	if _, err := os.Lstat(dest); err != nil {
		return false
	}
	return !IsGeneratedInstrucao(dest) && !adotavelDe50(rel, dest, cfg)
}

// InstrucoesDoUsuario lista os arquivos de instrução dos agentes de cfg que o update deixa intocados.
func InstrucoesDoUsuario(targetDir string, cfg config.Config) []string {
	var out []string
	for _, a := range cfg.Agents {
		for _, rel := range instrucaoAgentFiles[a] {
			if instrucaoDoUsuario(rel, filepath.Join(targetDir, filepath.FromSlash(rel)), cfg) {
				out = append(out, rel)
			}
		}
	}
	return out
}

// NovoPath devolve `<nome>.novo.<ext>` ao lado do arquivo.
func NovoPath(dest string) string {
	ext := filepath.Ext(dest)
	return strings.TrimSuffix(dest, ext) + ".novo" + ext
}

// avisaInstrucaoPreservada avisa em stderr e grava `<nome>.novo.md` com o que o forge geraria.
func avisaInstrucaoPreservada(targetDir, dest string, rendered []byte) {
	novo := NovoPath(dest)
	msg := "aviso: %s é seu (sem o marcador forge-sdd:generated) e foi mantido; o conteúdo novo do forge está em %s\n"
	if err := writeFileSafeMode(targetDir, novo, rendered, 0644); err != nil {
		refusals.Add(1)
		msg = "aviso: %s é seu (sem o marcador forge-sdd:generated) e foi mantido; %s não gravado\n"
	}
	fmt.Fprintf(os.Stderr, msg, notify.Sanitize(dest, 300, false), notify.Sanitize(novo, 300, false))
}
