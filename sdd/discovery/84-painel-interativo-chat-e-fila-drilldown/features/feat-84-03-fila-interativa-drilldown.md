# feat-84-03 — Componente TUI de Fila Rolável e Navegável (QueueDrilldownTUI)

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Componente de interface de terminal em `internal/survey/queue_drilldown.go` baseado em Charm `huh` já homologado (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0004`: Consome a fila oficial de `sdd/.queue/queue.json` de forma somente-leitura e não intrusiva.
- `ARQ-01`, `ARQ-04`: Interface isolada em `internal/survey`, sem dependências novas de terceiros além de `huh`.
- `DUP-01`: Componente único para navegação de tarefas com suporte a scroll e paginação completa (sem truncamentos).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Leitura da fila de tasks | `internal/queue/queue.go` | `ListTasks`, estruturas de dados da task | reusar | A TUI consome as structs nativas de tarefa sem duplicação de schema. |
| Componentes de survey | `internal/survey/survey.go` | Menus e selects interativos | estender | Construir `QueueDrilldownTUI` permitindo scroll completo em filas com dezenas de tarefas. |

## Dependências novas
- Nenhuma. Stdlib Go e `github.com/charmbracelet/huh` já presente (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-03`
- `DUP-01`, `DUP-02`
- `TST-03`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. Criar `internal/survey/queue_drilldown.go`:
   - Carregamento de 100% das tarefas da fila da iniciativa atual (sem truncar em "... e mais N tarefas").
   - Badges visuais de status: `[✓] Concluído`, `[▶] Em Andamento`, `[⌛] Na Fila`, `[✗] Falhou`.
   - Navegação por teclado com setas `↑` / `↓` e suporte a teclas de atalho de saída (`q` / `Esc`).
   - Retorno do identificador da tarefa selecionada ao pressionar `Enter`.
2. Criar `internal/survey/queue_drilldown_test.go`:
   - Teste unitário comprovando que listas com 25+ tarefas são carregadas integralmente na estrutura navegável.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/survey -run TestQueueDrilldown_List
```
- `TestQueueDrilldown_ListAllTasks`: todas as tarefas da fila são mapeadas para os itens de seleção com seus respectivos badges.
- `go vet ./internal/survey/...`: 100% limpo.
