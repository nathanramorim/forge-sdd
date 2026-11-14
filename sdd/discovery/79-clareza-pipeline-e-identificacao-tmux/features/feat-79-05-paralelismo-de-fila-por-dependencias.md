# feat-79-05 — Paralelismo de Fila por Grafo de Dependências (DAG & Ondas)

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-05
**Tamanho:** `M`
**Depende de:** 79-02
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
Hoje o consumo da fila (`internal/queue`) é sequencial (FIFO). Mesmo que duas tarefas sejam independentes (ex.: 79-01 e 79-02 ambas com `Depende de: —`), a CLI não oferece um mecanismo programático para identificar quais tarefas estão desimpedidas para rodar em paralelo. Esta feature introduz o campo `depends_on` na fila, extração automática das dependências da Spec e um resolvedor topológico de DAG, permitindo que a esteira libere lotes de tasks prontas para execução simultânea respeitando `max_parallel_tasks`.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Algoritmo de ordenação topológica de grafos, detecção de ciclos, novas propriedades na estrutura serializada da fila (`internal/queue/queue.go`) e testes exaustivos (`ARQ-11`, `COD-01`).

## Decisões do projeto aplicadas
- `ADR-0004`: Pipeline Multi-Sessão e paralelismo de construção.
- `ARQ-01`, `ARQ-02`: Lógica de domínio restrita a `internal/queue`, sem vazamento para handlers.
- `COD-01`, `COD-03`: Código Go idiomático, enriquecimento de erros (ex.: ciclo de dependência).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Estrutura da fila | `internal/queue/queue.go` | `Queue`, `QueueItem`, `Enqueue`, `Next` | estender | Adicionar `DependsOn []string` e método `EligibleTasks() []QueueItem`. |
| Extração de metadados de spec | `internal/queue/enqueue.go` | Leitura de `task-id`, `rules` e título | estender | Fazer parser regex de `**Depende de:**` do markdown da spec. |
| Testes de fila | `internal/queue/queue_test.go` | Ciclos de transição de status | estender | Testar ordenação topológica, desbloqueio quando dependência aprova e rejeição de ciclos. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `COD-01`, `COD-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Estender `QueueItem` com campo `DependsOn []string json:"depends_on,omitempty"`.
2. No `queue enqueue`, fazer o parse automático do campo `**Depende de:**` da spec markdown (ignorando traços `—` ou tarefas vazias) e salvar no item.
3. Implementar a função `EligibleTasks(q *Queue, max int) ([]QueueItem, error)`:
   - Uma task em estado `queued` é elegível se todas as tasks listadas em `DependsOn` estiverem com status `approved`.
   - Se houver ciclo de dependência (ex.: A depende de B e B depende de A), retorna erro descritivo.
4. Adicionar flag `--batch` no comando `forge-sdd queue next` para retornar a lista de tasks prontas para paralelizar.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/queue -run TestQueueDAGParallelEligibility
```
- Duas tasks com `DependsOn: []` são retornadas simultaneamente como elegíveis.
- Uma task com `DependsOn: ["79-01"]` só se torna elegível após a aprovação da `79-01`.
- Detecção de ciclo retorna erro imediato sem travar em loop infinito.
- `go build ./... && go vet ./...` sem erros.
