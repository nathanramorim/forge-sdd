package pipeline

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func live(cli string, raw bool) (*LiveView, *bytes.Buffer, *bytes.Buffer) {
	var term, log bytes.Buffer
	return NewLiveView(&term, &log, LiveOptions{CLI: cli, Raw: raw}), &term, &log
}

func feed(v *LiveView, lines ...string) {
	for _, l := range lines {
		_, _ = v.Write([]byte(l + "\n"))
	}
	v.Flush()
}

func TestLiveView_EventosDoClaude(t *testing.T) {
	v, term, log := live(config.CLIClaude, false)
	events := []string{
		`{"type":"system","subtype":"init","session_id":"x"}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Vou criar a entidade.\nDepois testo."},{"type":"tool_use","name":"Write","input":{"file_path":"/Users/x/app/backend/src/domain/unit.entity.ts","content":"conteudo enorme"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"npm test"}},{"type":"tool_use","name":"Read","input":{"file_path":"/x/a.ts"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","is_error":true,"content":"boom"}]}}`,
		`{"type":"result","subtype":"success","is_error":true,"result":"You've hit your 5-hour limit · resets 10:40pm (America/Sao_Paulo)"}`,
	}
	feed(v, events...)
	assert.Equal(t, strings.Join([]string{
		"  › Vou criar a entidade.",
		"  ✎ src/domain/unit.entity.ts",
		"  $ npm test",
		"  ✗ uma ferramenta falhou",
		"  ✗ You've hit your 5-hour limit · resets 10:40pm (America/Sao_Paulo)",
	}, "\n")+"\n", term.String())
	assert.Equal(t, strings.Join(events, "\n")+"\n", log.String(), "o log guarda tudo, bruto")
	assert.NotContains(t, term.String(), `"type"`, "JSON nunca aparece no terminal")
}

func TestLiveView_EventosDoGemini(t *testing.T) {
	v, term, _ := live(config.CLIGemini, false)
	feed(v,
		`{"type":"init","session_id":"s"}`,
		`{"type":"tool_use","tool_name":"write_file","parameters":{"file_path":"src/a.ts"}}`,
		`{"type":"tool_use","tool_name":"run_shell_command","parameters":{"command":"npm run build"}}`,
		`{"type":"tool_result","status":"error"}`,
		`{"type":"message","role":"assistant","content":"Concluí a tarefa."}`,
		`{"type":"message","role":"user","content":"ignorado"}`,
	)
	assert.Equal(t, "  ✎ src/a.ts\n  $ npm run build\n  ✗ uma ferramenta falhou\n  › Concluí a tarefa.\n", term.String())
}

func TestLiveView_EventosDoCodex(t *testing.T) {
	v, term, _ := live(config.CLICodex, false)
	feed(v,
		`{"type":"thread.started","thread_id":"t"}`,
		`{"type":"item.started","item":{"type":"command_execution","command":"npm test"}}`,
		`{"type":"item.completed","item":{"type":"command_execution","command":"npm test","exit_code":0}}`,
		`{"type":"item.completed","item":{"type":"command_execution","command":"go vet","exit_code":2}}`,
		`{"type":"item.completed","item":{"type":"file_change","changes":[{"path":"src/a.ts","kind":"add"},{"path":"src/b.ts","kind":"update"}]}}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"Feito."}}`,
	)
	assert.Equal(t, "  $ npm test\n  ✓ comando terminou (0)\n  ✗ comando terminou (2)\n  ✎ src/a.ts\n  ✎ src/b.ts\n  › Feito.\n", term.String())
}

func TestLiveView_JSONDesconhecidoOuQuebradoNuncaApareceCru(t *testing.T) {
	v, term, _ := live(config.CLIClaude, false)
	feed(v, `{"type":"coisa_nova","x":1}`, `{"type":"assistant","message":`, `{"incompleto`, `[{"a":1}]`)
	assert.Equal(t, "", term.String())
}

func TestLiveView_TextoComumLimpoETruncado(t *testing.T) {
	v, term, _ := live(config.CLIAgy, false)
	long := strings.Repeat("x", 500)
	feed(v, "\x1b[31mvermelho\x1b[0m", long, "   ", "")
	lines := strings.Split(strings.TrimSpace(term.String()), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "vermelho", strings.TrimSpace(lines[0]), "sem ANSI")
	assert.LessOrEqual(t, len([]rune(lines[1])), maxLiveLine)
	assert.True(t, strings.HasSuffix(lines[1], "…"))
}

func TestLiveView_SegredoMascarado(t *testing.T) {
	skKey := "sk-" + strings.Repeat("a", 24)
	ghKey := "ghp_" + strings.Repeat("b", 24)
	v, term, log := live(config.CLIClaude, false)
	feed(v,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"curl -H 'Authorization: Bearer abc123' https://x; export API_KEY=`+skKey+`"}}]}}`,
		"token="+ghKey,
	)
	out := term.String()
	assert.NotContains(t, out, skKey)
	assert.NotContains(t, out, ghKey)
	assert.NotContains(t, out, "abc123")
	assert.Contains(t, out, "***")
	assert.Contains(t, log.String(), skKey, "o log bruto guarda o que o agente imprimiu (arquivo 0600, fora do git)")
}

func TestLiveView_RuidoDeHookColapsaEmUmaLinha(t *testing.T) {
	v, term, _ := live(config.CLIClaude, false)
	noise := "SessionEnd hook [http://127.0.0.1:47600/hooks/claude-code/SessionEnd] failed: connect ECONNREFUSED 127.0.0.1:47600"
	feed(v, noise, noise, "PermissionRequest hook [http://127.0.0.1:47600/x] failed: connect ECONNREFUSED")
	assert.Equal(t, "  (hooks globais do agente falharam: ignorado)\n", term.String())
}

func TestLiveView_RawDevolveASaidaBruta(t *testing.T) {
	v, term, log := live(config.CLIClaude, true)
	raw := "{\"type\":\"system\"}\nlinha comum\n"
	_, _ = v.Write([]byte(raw))
	assert.Equal(t, raw, term.String())
	assert.Equal(t, raw, log.String())
}

func TestLiveView_LinhaPartidaEmDoisWrites(t *testing.T) {
	v, term, _ := live(config.CLIClaude, false)
	_, _ = v.Write([]byte(`{"type":"assistant","message":{"content":[{"type":"tool_use","na`))
	_, _ = v.Write([]byte(`me":"Bash","input":{"command":"ls"}}]}}` + "\n"))
	assert.Equal(t, "  $ ls\n", term.String())
}

func TestLiveView_TailGuardaAsUltimasLinhasUteis(t *testing.T) {
	v, _, _ := live(config.CLIAgy, false)
	for i := 0; i < 30; i++ {
		feed(v, "linha "+strings.Repeat("a", i%3))
	}
	feed(v, "You've hit your 5-hour limit")
	tail := v.Tail(3)
	require.Len(t, tail, 3)
	assert.Equal(t, "You've hit your 5-hour limit", tail[2])
	assert.LessOrEqual(t, len(v.Tail(100)), tailLines)
}

// ---- GitWatcher ----

type gitScript struct {
	outs []string
	err  error
	i    int
}

func (g *gitScript) run(ctx context.Context, dir string, args ...string) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	o := g.outs[min(g.i, len(g.outs)-1)]
	g.i++
	return o, nil
}

func TestGitWatch_ArquivoNovoEAlteradoViramLinhas(t *testing.T) {
	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	g := &gitScript{outs: []string{
		"?? velho.txt\n",
		"?? velho.txt\n?? src/a.ts\n",
		"?? velho.txt\n?? src/a.ts\n M src/b.ts\n?? .forge/runtime/x\n",
	}}
	w := NewGitWatcher("/p", func() time.Time { return now }, nil)
	w.Git = g.run
	ctx := context.Background()
	assert.Empty(t, w.Poll(ctx), "o primeiro consulta é só o estado de partida")
	assert.Equal(t, []string{"  ✎ src/a.ts"}, w.Poll(ctx))
	assert.Equal(t, []string{"  ✎ src/b.ts"}, w.Poll(ctx), "arquivos do próprio .forge são ignorados")
	assert.Equal(t, 2, w.Count())
}

func TestGitWatch_MuitosArquivosResumem(t *testing.T) {
	now := time.Now()
	var many strings.Builder
	for i := 0; i < 9; i++ {
		many.WriteString("?? f" + string(rune('a'+i)) + ".ts\n")
	}
	g := &gitScript{outs: []string{"", many.String()}}
	w := NewGitWatcher("/p", func() time.Time { return now }, nil)
	w.Git = g.run
	w.Poll(context.Background())
	lines := w.Poll(context.Background())
	require.Len(t, lines, maxWatchLines+1)
	assert.Equal(t, "  ✎ (+4 arquivos)", lines[len(lines)-1])
}

func TestGitWatch_BatimentoSemMudanca(t *testing.T) {
	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	cur := now
	g := &gitScript{outs: []string{"?? a.ts\n"}}
	var activity time.Time
	w := NewGitWatcher("/p", func() time.Time { return cur }, func() time.Time { return activity })
	w.Git = g.run
	ctx := context.Background()
	w.Poll(ctx)
	cur = now.Add(10 * time.Second)
	assert.Empty(t, w.Poll(ctx), "antes de 30 s não há batimento")
	cur = now.Add(31 * time.Second)
	assert.Equal(t, []string{"  … trabalhando há 31s · 0 arquivos alterados"}, w.Poll(ctx))
	cur = now.Add(45 * time.Second)
	assert.Empty(t, w.Poll(ctx), "o próximo só 30 s depois")
	cur = now.Add(2*time.Minute + 10*time.Second)
	activity = cur.Add(-5 * time.Second) // o agente emitiu evento há pouco: sem batimento
	assert.Empty(t, w.Poll(ctx))
}

func TestGitWatch_SemGitNaoQuebra(t *testing.T) {
	w := NewGitWatcher("/p", nil, nil)
	w.Git = (&gitScript{err: errors.New("not a git repository")}).run
	assert.Empty(t, w.Poll(context.Background()))
	assert.Equal(t, 0, w.Count())
}

func TestFormatElapsed(t *testing.T) {
	assert.Equal(t, "45s", FormatElapsed(45*time.Second))
	assert.Equal(t, "2m10s", FormatElapsed(130*time.Second))
	assert.Equal(t, "1h05m", FormatElapsed(65*time.Minute))
}

// ---- integração com o worker ----

func TestWorker_TerminalLimpoELogBruto(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	runner := &fakeRunner{fn: func(RunSpec) (RunResult, error) { return RunResult{}, errors.New("não usado") }}
	runner.fn = func(spec RunSpec) (RunResult, error) {
		_, err := bus.UpdatePR(dir, layout.RoleBuilder, "job-1", "forge-builder", 12, "https://github.com/acme/app/pull/12", "feat/77-auth-oauth", now)
		return RunResult{}, err
	}
	var term bytes.Buffer
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Out: &term, Runner: runner, WatchEvery: -1, Now: func() time.Time { return now }}))
	out := term.String()
	assert.Contains(t, out, "▶ job auth-oauth · feature_build · tentativa 1 · claude · permissões project")
	assert.Contains(t, out, "  agente falso rodando", "a saída do agente aparece limpa")
	assert.Contains(t, out, "■ concluído em 0s")

	logs, err := filepath.Glob(filepath.Join(layout.Logs(dir), "builder-job-1.log"))
	require.NoError(t, err)
	require.Len(t, logs, 1)
	raw, _ := os.ReadFile(logs[0])
	assert.Contains(t, string(raw), "agente falso rodando")
	assert.NotContains(t, string(raw), "▶ job", "o log guarda só a saída do agente")
}

func TestWorker_RawNaoUsaAVisaoLimpa(t *testing.T) {
	dir := workerProject(t)
	seedJob(t, dir, layout.RoleBuilder, "job-1", bus.TypeFeatureBuild, bus.PriorityNormal, now)
	runner := &fakeRunner{fn: func(RunSpec) (RunResult, error) {
		_, err := bus.UpdatePR(dir, layout.RoleBuilder, "job-1", "forge-builder", 12, "https://github.com/acme/app/pull/12", "feat/77-auth-oauth", now)
		return RunResult{}, err
	}}
	var term bytes.Buffer
	require.NoError(t, RunWorker(context.Background(), WorkerOptions{Dir: dir, Role: layout.RoleBuilder, Once: true, Raw: true, Out: &term, Runner: runner, WatchEvery: -1, Now: func() time.Time { return now }}))
	assert.Contains(t, term.String(), "agente falso rodando\n", "sem o recuo da visão limpa")
}
