# feat-84-02 — Integração do Stream Narrativo no Ciclo de Vida do Worker

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Atualização do loop de eventos e apresentação no comando `worker` (`cmd/forge-sdd/worker.go`), sem novos processos em background (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0004`, `ADR-0008`: Apresentação humana das sessões Builder e Reviewer mantendo o desacoplamento de fila.
- `ARQ-01`: Adaptação da saída no CLI delegando a formatação ao pacote `internal/monitor`.
- `COD-01`, `COD-04`: Remoção do log monótono de heartbeat legado (`trabalhando há XmYs • Z arquivos alterados`), substituindo por marcos legíveis.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Loop de execução do worker | `cmd/forge-sdd/worker.go` | Polling e execução de tarefas | estender | Integrar `ChatFeedFormatter` emitindo eventos em tempo real para stdout do worker. |
| Formatador de feed | `internal/monitor/chat_feed.go` | Gerador de balões de etapas | reusar | Reuso direto do formatador construído na task 84-01. |

## Dependências novas
- Nenhuma. Stdlib Go pura (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-06`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-03`, `COD-04`
- `DUP-01`, `DUP-02`
- `TST-03`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. Em `cmd/forge-sdd/worker.go`:
   - Instanciar e acionar o `ChatFeedFormatter` nas transições de estado do worker:
     - Quando recebe task: emite balão de início e análise de requisitos.
     - Quando arquivos são tocados: emite balão de modificação com lista limpa de arquivos.
     - Quando validações/testes rodam: emite balão de execução de validação.
     - Ao concluir ou falhar: emite card de encerramento com tempo e status final.
2. Em `cmd/forge-sdd/worker_test.go`:
   - Teste unitário verificando que a saída do worker reflete a sequência narrativa de eventos.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./cmd/forge-sdd -run TestWorker_ChatFeed
```
- `TestWorker_ChatFeed_Sequence`: o worker emite os balões conversacionais sem lançar saídas cruas de bash.
- `go vet ./cmd/forge-sdd/...`: 100% limpo.
