# feat-84-01 — Formatador de Stream Narrativo para Workers (ChatFeedFormatter)

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Criação de estrutura de dados e formatador puramente in-memory em `internal/monitor/chat_feed.go`, sem dependências externas nem processos de rede (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0004`, `ADR-0008`: Apresentação visual limpa das sessões multi-agente sem acoplar à execução de comandos raw.
- `ARQ-01`, `ARQ-04`: Módulo desacoplado de CLI, residente em `internal/monitor`, utilizando apenas a stdlib Go (`fmt`, `strings`, `time`).
- `DUP-01`: Único ponto de geração e formatação de balões conversacionais de etapas para os workers.
- `COD-01`, `COD-03`: Funções puras e tratamento explícito de estados de eventos.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Monitoramento de processos e tokens | `internal/monitor/monitor.go` | Formatação de tokens e métricas | estender | Adicionar `ChatFeedFormatter` para formatar eventos de ciclo de vida com status humano. |
| Formatação de layout | `internal/layout/layout.go` | Constantes de papéis e nomes | reusar | Reutilizar `RoleBuilder` e `RoleReviewer` na identificação do remetente do balão. |

## Dependências novas
- Nenhuma. Stdlib Go pura (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-03`, `COD-05`
- `DUP-01`, `DUP-02`
- `TST-03`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`.

## Escopo
1. Criar `internal/monitor/chat_feed.go`:
   - Tipos de eventos: `StartTask`, `ReadSpec`, `EditFiles`, `RunValidation`, `TaskSuccess`, `TaskError`.
   - Gerador de balões estruturados com badges visuais (`💭 Analisando...`, `📋 Lendo spec...`, `🔨 Modificando...`, `🧪 Testando...`, `✅ Concluído!`, `❌ Falha`).
   - Suporte a escape ANSI para bordas e cores suaves (com fallback para ASCII seguro quando solicitado).
2. Criar `internal/monitor/chat_feed_test.go`:
   - Testes unitários validando a geração dos balões para cada tipo de evento e formatação de erros.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/monitor -run TestChatFeed
```
- `TestChatFeed_FormatEvents`: gera balões humanizados com quebras de linha e ícones corretos.
- `TestChatFeed_ErrorEvent`: destaca mensagens de erro sem quebrar a estrutura visual.
- `go vet ./internal/monitor/...`: 100% limpo.
