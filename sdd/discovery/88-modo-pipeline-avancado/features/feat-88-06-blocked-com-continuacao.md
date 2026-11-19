# feat-88-06 — Task `blocked` não para a fila

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** feat-88-04

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/queue` e `internal/notify` (ARQ-01).

## Decisões do projeto aplicadas
- ADR-0012.
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Status e elegibilidade | `queue.EligibleTasks`, constantes `Status*` | DAG de elegíveis por `approved` | estender | Status `blocked`; dependentes ficam parados, independentes seguem. |
| Inbox | `internal/notify` | Notificações idempotentes | reusar | Motivo do bloqueio no inbox. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `COD-03`, `DUP-01`, `TST-01`
- Rules: `.agents/rules/sdd-arquitetura.md`.

## Escopo
1. Status `blocked` com motivo; `EligibleTasks` ignora `blocked` e seus dependentes, mas despacha as independentes.
2. `forge status` e `queue list` mostram `blocked`.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/queue -run TestBlocked
```
- `TestBlockedNaoBloqueiaIndependentes`, `TestBlockedBloqueiaDependentes`.
- `go vet ./...` limpo.
