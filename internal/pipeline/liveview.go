package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/gate"
)

const (
	maxLiveLine = 200
	tailLines   = 20
	maxPartial  = 64 * 1024 // linha sem quebra maior que isso é descartada até a próxima
)

var (
	ansiRe      = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
	hookNoiseRe     = regexp.MustCompile(`(?i)\bhook \[[^\]]*\] failed:`)
	skillConflictRe = regexp.MustCompile(`(?i)Skill conflict detected`)
	yoloNoiseRe     = regexp.MustCompile(`(?i)YOLO mode is enabled`)
	ripgrepNoiseRe  = regexp.MustCompile(`(?i)Ripgrep is not available`)
	editTools       = map[string]bool{"edit": true, "write": true, "multiedit": true, "notebookedit": true, "write_file": true, "replace": true, "create_file": true, "edit_file": true, "apply_patch": true, "file_change": true}
	shellTools  = map[string]bool{"bash": true, "run_shell_command": true, "shell": true, "run_command": true, "command_execution": true, "exec_command": true}
)

// LiveOptions configura a visão limpa do que o agente faz.
type LiveOptions struct {
	CLI string
	Raw bool // saída bruta, como antes
	Now func() time.Time
}

// LiveView é o io.Writer que o runner recebe: grava tudo, bruto, no log e mostra no terminal só
// linhas curtas do que o agente faz. Eventos estruturados de cada CLI viram resumos; JSON que não
// entende é ignorado, nunca impresso; segredos são mascarados; o ruído de hooks globais vira uma
// linha só.
type LiveView struct {
	mu        sync.Mutex
	term      io.Writer
	log       io.Writer
	opts      LiveOptions
	partial   bytes.Buffer
	lastEvent time.Time
	hookShown           bool
	skillConflicts      int
	skillConflictsShown bool
	inKey               bool // dentro de um bloco de chave privada: nada dele vai ao terminal
	dropNL    bool // descartando uma linha gigante até a próxima quebra
	tail      []string
}

// NewLiveView liga o terminal e o log (que pode ser nil).
func NewLiveView(term, log io.Writer, opts LiveOptions) *LiveView {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &LiveView{term: term, log: log, opts: opts, lastEvent: opts.Now()}
}

// Write recebe a saída do agente. O log recebe sempre tudo, bruto (arquivo 0600 fora do git: pode
// conter segredos que o agente imprimiu); o terminal recebe a versão limpa e mascarada.
func (v *LiveView) Write(p []byte) (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.log != nil {
		_, _ = v.log.Write(p)
	}
	if v.opts.Raw {
		_, err := v.term.Write(p)
		return len(p), err
	}
	if v.dropNL {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			return len(p), nil
		}
		p, v.dropNL = p[i+1:], false
	}
	v.partial.Write(p)
	if v.partial.Len() > maxPartial && bytes.IndexByte(v.partial.Bytes(), '\n') < 0 {
		v.partial.Reset()
		v.dropNL = true
		return len(p), nil
	}
	for {
		data := v.partial.Bytes()
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		line := string(data[:i])
		v.partial.Next(i + 1)
		v.handleLine(strings.TrimRight(line, "\r"))
	}
	return len(p), nil
}

// Flush processa a última linha sem quebra e emite resumos pendentes.
func (v *LiveView) Flush() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.partial.Len() > 0 {
		line := v.partial.String()
		v.partial.Reset()
		if !v.opts.Raw {
			v.handleLine(line)
		}
	}
	if !v.opts.Raw {
		v.flushSkillConflicts()
	}
}

// Line escreve uma linha já formatada (cabeçalho, rodapé, batimento), sanitizada e mascarada.
func (v *LiveView) Line(s string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.emit(s, false)
}

// Activity devolve o instante do último evento do agente (o batimento só aparece sem atividade).
func (v *LiveView) Activity() time.Time {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.lastEvent
}

// Tail devolve as últimas linhas úteis da saída (texto comum e erros), já sanitizadas.
func (v *LiveView) Tail(n int) []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if n > len(v.tail) {
		n = len(v.tail)
	}
	return append([]string(nil), v.tail[len(v.tail)-n:]...)
}

func (v *LiveView) emit(s string, useful bool) {
	s = clean(s)
	if strings.TrimSpace(s) == "" {
		return
	}
	if useful {
		v.tail = append(v.tail, strings.TrimSpace(s))
		if len(v.tail) > tailLines {
			v.tail = v.tail[len(v.tail)-tailLines:]
		}
	}
	fmt.Fprintln(v.term, s)
}

// clean tira ANSI e controle, mascara segredos e trunca (SEG-05).
func clean(s string) string {
	s = ansiRe.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' || r == 0x7f {
			return -1
		}
		return r
	}, s)
	s = gate.MaskSecrets(s)
	if r := []rune(s); len(r) > maxLiveLine {
		s = string(r[:maxLiveLine-1]) + "…"
	}
	return strings.TrimRight(s, " \t")
}

var (
	// Montados em partes para o P-13 não casar com o próprio código-fonte.
	keyBeginRe = regexp.MustCompile("-----BEGIN [A-Z ]*PRIVATE" + " KEY-----")
	keyEndRe   = regexp.MustCompile("-----END [A-Z ]*PRIVATE" + " KEY-----")
)

func (v *LiveView) handleLine(line string) {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return
	}
	if v.inKey {
		if keyEndRe.MatchString(trim) {
			v.inKey = false
		}
		return
	}
	if keyBeginRe.MatchString(trim) {
		v.inKey = !keyEndRe.MatchString(trim)
		v.emit("  (chave privada omitida)", false)
		return
	}
	if hookNoiseRe.MatchString(trim) {
		if !v.hookShown {
			v.hookShown = true
			v.emit("  (hooks globais do agente falharam: ignorado)", false)
		}
		return
	}
	if skillConflictRe.MatchString(trim) {
		v.skillConflicts++
		return
	}
	if yoloNoiseRe.MatchString(trim) || ripgrepNoiseRe.MatchString(trim) {
		return
	}
	v.flushSkillConflicts()
	if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[{") {
		var ev map[string]any
		if json.Unmarshal([]byte(trim), &ev) != nil {
			return // JSON quebrado ou truncado: nunca aparece cru
		}
		v.lastEvent = v.opts.Now()
		for _, l := range summarizeEvent(v.opts.CLI, ev) {
			v.emit(l, strings.Contains(l, "✗"))
		}
		return
	}
	v.lastEvent = v.opts.Now()
	v.emit("  "+trim, true)
}

func (v *LiveView) flushSkillConflicts() {
	if v.skillConflicts > 0 && !v.skillConflictsShown {
		v.skillConflictsShown = true
		v.emit(fmt.Sprintf("  (%d skills duplicadas nas suas pastas globais: ignorado; o log bruto guarda tudo)", v.skillConflicts), false)
	}
}

// summarizeEvent traduz um evento estruturado em linhas curtas. Formatos conhecidos: claude e
// gemini (stream-json) e codex (--json); o que não reconhece vira nada (A CONFIRMAR por CLI).
func summarizeEvent(cli string, ev map[string]any) []string {
	typ, _ := ev["type"].(string)
	var out []string
	switch typ {
	case "assistant":
		if msg, ok := ev["message"].(map[string]any); ok {
			if blocks, ok := msg["content"].([]any); ok {
				for _, b := range blocks {
					out = append(out, blockLines(b)...)
				}
			}
		}
	case "user":
		if msg, ok := ev["message"].(map[string]any); ok {
			if blocks, ok := msg["content"].([]any); ok {
				for _, b := range blocks {
					if m, ok := b.(map[string]any); ok && m["type"] == "tool_result" && m["is_error"] == true {
						out = append(out, "  ✗ uma ferramenta falhou")
					}
				}
			}
		}
	case "tool_use": // gemini
		out = append(out, toolLines(firstString(ev, "tool_name", "name"), firstMap(ev, "parameters", "input", "args"))...)
	case "tool_result": // gemini
		if st, _ := ev["status"].(string); st != "" && st != "success" && st != "ok" {
			out = append(out, "  ✗ uma ferramenta falhou")
		}
	case "message": // gemini
		if role, _ := ev["role"].(string); role == "assistant" || role == "model" {
			if t := firstTrimmed(firstString(ev, "content", "text")); t != "" {
				out = append(out, "  › "+t)
			}
		}
	case "item.started", "item.completed": // codex
		out = append(out, codexItemLines(typ, ev)...)
	case "result":
		if ev["is_error"] == true {
			if t := firstTrimmed(firstString(ev, "result", "error", "message")); t != "" {
				out = append(out, "  ✗ "+t)
			}
		}
	case "error":
		if t := firstTrimmed(firstString(ev, "message", "error")); t != "" {
			out = append(out, "  ✗ "+t)
		}
	}
	return out
}

func blockLines(b any) []string {
	m, ok := b.(map[string]any)
	if !ok {
		return nil
	}
	switch m["type"] {
	case "text":
		if t := firstTrimmed(firstString(m, "text")); t != "" {
			return []string{"  › " + t}
		}
	case "tool_use":
		return toolLines(firstString(m, "name"), firstMap(m, "input"))
	}
	return nil
}

func codexItemLines(typ string, ev map[string]any) []string {
	item, _ := ev["item"].(map[string]any)
	if item == nil {
		return nil
	}
	it, _ := item["type"].(string)
	switch {
	case shellTools[it] && typ == "item.started":
		return []string{"  $ " + firstTrimmed(firstString(item, "command"))}
	case shellTools[it] && typ == "item.completed":
		if code, ok := item["exit_code"].(float64); ok {
			if code == 0 {
				return []string{"  ✓ comando terminou (0)"}
			}
			return []string{fmt.Sprintf("  ✗ comando terminou (%d)", int(code))}
		}
	case editTools[it] && typ == "item.completed":
		var out []string
		if ch, ok := item["changes"].([]any); ok {
			for i, c := range ch {
				if i >= 5 {
					out = append(out, fmt.Sprintf("  ✎ (+%d arquivos)", len(ch)-5))
					break
				}
				if cm, ok := c.(map[string]any); ok {
					out = append(out, "  ✎ "+shortPath(firstString(cm, "path")))
				}
			}
		}
		return out
	case it == "agent_message" && typ == "item.completed":
		if t := firstTrimmed(firstString(item, "text")); t != "" {
			return []string{"  › " + t}
		}
	}
	return nil
}

func toolLines(name string, input map[string]any) []string {
	n := strings.ToLower(name)
	switch {
	case editTools[n]:
		if p := firstString(input, "file_path", "path", "filePath", "absolute_path"); p != "" {
			return []string{"  ✎ " + shortPath(p)}
		}
		return []string{"  ✎ (arquivo)"}
	case shellTools[n]:
		if c := firstTrimmed(firstString(input, "command", "cmd")); c != "" {
			return []string{"  $ " + c}
		}
	}
	return nil
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func firstMap(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		if mm, ok := m[k].(map[string]any); ok {
			return mm
		}
	}
	return nil
}

func firstTrimmed(s string) string { return strings.TrimSpace(firstLine(strings.TrimSpace(s))) }

// shortPath mostra o caminho relativo ao diretório de trabalho quando possível.
func shortPath(p string) string {
	if filepath.IsAbs(p) {
		if i := strings.Index(p, "/src/"); i >= 0 {
			return p[i+1:]
		}
		return filepath.Base(filepath.Dir(p)) + "/" + filepath.Base(p)
	}
	return p
}
