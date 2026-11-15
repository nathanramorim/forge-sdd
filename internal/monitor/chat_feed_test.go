package monitor

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestChatFeed_RenderStepCards(t *testing.T) {
	fmt := NewChatFeedFormatter(false, false)

	// 1. Início da tarefa
	startOut := fmt.StartTask("feat-84-01", "Formatador de Stream Narrativo")
	if !strings.Contains(startOut, "💭 Analisando requisitos e critérios da task feat-84-01") {
		t.Fatalf("saída inesperada no StartTask: %s", startOut)
	}
	if !strings.Contains(startOut, "(Formatador de Stream Narrativo)") {
		t.Fatalf("esperava título da tarefa na saída: %s", startOut)
	}

	// 2. Leitura da spec
	specOut := fmt.ReadSpec("sdd/discovery/84-painel/criteria.md")
	if !strings.Contains(specOut, "📋 Lendo sdd/discovery/84-painel/criteria.md...") {
		t.Fatalf("saída inesperada no ReadSpec: %s", specOut)
	}

	// 3. Ação em arquivos
	editOut := fmt.EditFiles("internal/monitor/chat_feed.go", "internal/monitor/chat_feed_test.go")
	if !strings.Contains(editOut, "🔨 Editando internal/monitor/chat_feed.go, internal/monitor/chat_feed_test.go...") {
		t.Fatalf("saída inesperada no EditFiles: %s", editOut)
	}

	// 4. Execução de testes
	testOut := fmt.RunValidation("testes automatizados (go test)")
	if !strings.Contains(testOut, "🧪 Executando suite de testes automatizados (go test)...") {
		t.Fatalf("saída inesperada no RunValidation: %s", testOut)
	}

	// 5. Conclusão
	compOut := fmt.CompleteTask("feat-84-01", 75*time.Second, "Handoff despachado para o Reviewer.")
	if !strings.Contains(compOut, "✅ Concluído! Task feat-84-01 finalizada com sucesso em 1m15s.") {
		t.Fatalf("saída inesperada no CompleteTask: %s", compOut)
	}
	if !strings.Contains(compOut, "Handoff despachado para o Reviewer.") {
		t.Fatalf("nota ausente no CompleteTask: %s", compOut)
	}
}

func TestChatFeed_ErrorHighlighting(t *testing.T) {
	fmt := NewChatFeedFormatter(false, false)
	logs := []string{
		"compiling package...",
		"=== RUN   TestSomething",
		"--- FAIL: TestSomething (0.01s)",
		"    something_test.go:42: expected 1, got 2",
		"FAIL",
	}

	errOut := fmt.FormatErrorCard("teste", errors.New("exit status 1"), logs)
	if !strings.Contains(errOut, "❌ Falha na etapa [teste]: exit status 1") {
		t.Fatalf("esperava cabeçalho de erro formatado, obteve: %s", errOut)
	}
	if !strings.Contains(errOut, "Últimos logs") {
		t.Fatalf("esperava bloco de últimos logs, obteve: %s", errOut)
	}
	if !strings.Contains(errOut, "something_test.go:42: expected 1, got 2") {
		t.Fatalf("esperava linha de log específica no bloco de erro: %s", errOut)
	}
}

func TestChatFeed_FormatEvents(t *testing.T) {
	fmt := NewChatFeedFormatter(false, false)
	ev := ChatEvent{
		Type:   EventStartTask,
		Role:   "builder",
		TaskID: "fix-abcc-01",
		Title:  "Telas no frontend-admin",
	}
	out := fmt.FormatEvent(ev)
	if !strings.HasPrefix(out, "[builder] 💭") {
		t.Fatalf("esperava prefixo de papel [builder] e emoji, obteve: %s", out)
	}
}

func TestChatFeed_ErrorEvent(t *testing.T) {
	fmt := NewChatFeedFormatter(false, false)
	ev := ChatEvent{
		Type:  EventTaskError,
		Role:  "reviewer",
		Step:  "validação",
		Error: errors.New("violação da regra ARQ-01"),
		Logs:  []string{"erro em cmd/main.go: regra violada"},
	}
	out := fmt.FormatEvent(ev)
	if !strings.Contains(out, "❌ Falha na etapa [validação]: violação da regra ARQ-01") {
		t.Fatalf("esperava formatação de erro, obteve: %s", out)
	}
}

func TestChatFeed_ASCIIFallback(t *testing.T) {
	fmt := NewChatFeedFormatter(false, true)
	out := fmt.StartTask("84-01", "")
	if !strings.Contains(out, "[?] Analisando requisitos") {
		t.Fatalf("esperava fallback ASCII '[?]', obteve: %s", out)
	}

	errOut := fmt.FormatErrorCard("teste", errors.New("falha"), nil)
	if !strings.Contains(errOut, "[FAIL] Falha na etapa [teste]") {
		t.Fatalf("esperava fallback ASCII '[FAIL]', obteve: %s", errOut)
	}
}
