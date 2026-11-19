# feat-90-08 — Fix no `dispatch`

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** feat-90-03, feat-90-05

## Problema
`forge dispatch` só lista `features/feat-*.md`. Os fixes 84-03, 88-01 e 88-02 exigiram job montado à mão, sem portão pré-build e com `task_ids` vazio.

## Comportamento
`forge dispatch <iniciativa> --fix fix-88-01` lê `fixes/fix-88-01-*.md`, aplica a checagem da feat-90-03, enfileira o job com `task_ids: ["fix-88-01"]`, branch `fix/88-01-<nome>` e base igual à branch da iniciativa (resolvedor da feat-90-05), e adiciona o item do fix à fila.

## Modelo do módulo
**Enxuto** (ARQ-11): estende `internal/pipeline/dispatch.go` e a flag em `cmd/forge-sdd`.

## Decisões do projeto aplicadas
ADR-0004 (D-2, gate em duas metades), ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Despacho | `dispatch.Dispatch`, `listTasks` | Só `feat-*` | estender | `DUP-03` |
| Checagem | feat-90-03 | Dependências e critérios | reusar | Mesmo portão |
| Base | feat-90-05 | Resolvedor único | reusar | `DUP-01` |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`FLX-01`, `FLX-02`, `DUP-03`, `TST-03`. Rules: `sdd-fluxo-e-portao`, `sdd-reuso-e-duplicacao`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/pipeline -run TestDispatchFix
go vet ./... && go test ./...
```
- Job com `task_ids` preenchido, base da iniciativa e spec incompleta bloqueada com a lista do que falta.
