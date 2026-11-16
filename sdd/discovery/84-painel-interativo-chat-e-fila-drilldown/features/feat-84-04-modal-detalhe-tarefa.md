# feat-84-04 — Modal de Inspeção e Detalhes da Tarefa (Drill-Down)

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Componente de renderização de detalhes da tarefa em `internal/survey/task_detail.go` (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0004`: Visibilidade completa do veredito da governança e histórico de execução da task.
- `ARQ-01`, `ARQ-04`: Visualização desacoplada de CLI, sem escrita em disco nem concorrência com o worker.
- `DUP-01`: Função única para formatação do card de inspeção detalhada.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Consulta de telemetria | `internal/monitor/monitor.go` | Métricas e tokens de sessão | reusar | Extrair contagem de tokens e tempo decorrido da tarefa. |
| Leitura de spec e veredito | `internal/queue/queue.go` | Caminho da spec e status | reusar | Acessar arquivos modificados e laudo do reviewer registrado na task. |
| Componentes TUI | `internal/survey/queue_drilldown.go` | Navegador da fila | estender | Chamar a exibição de detalhes quando `Enter` for pressionado na lista de tarefas. |

## Dependências novas
- Nenhuma. Stdlib Go pura (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-03`
- `DUP-01`, `DUP-02`
- `TST-03`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. Criar `internal/survey/task_detail.go`:
   - Estrutura `TaskDetailView` exibindo:
     - Título e ID da task.
     - Status atual e duração da execução.
     - Lista de arquivos modificados (adicionados, alterados).
     - Resumo do commit gerado pelo Builder.
     - Laudo completo do Reviewer (regras verificadas, testes, veredito).
   - Tecla `Esc` ou `Enter` para retornar à lista principal de tarefas.
2. Criar `internal/survey/task_detail_test.go`:
   - Teste unitário validando a montagem e exibição dos dados de inspeção para tarefas concluídas e com erro.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/survey -run TestQueueDrilldown_Inspect
```
- `TestQueueDrilldown_InspectDetails`: renderiza o card detalhado contendo arquivos, laudo e tempo, e retorna à lista ao receber sinal de retorno.
- `go vet ./internal/survey/...`: 100% limpo.
