# feat-122-04 — Telemetria por task e papel com uso real

Task 1.5. Trilha B. Depende de feat-122-01.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/monitor, cmd/forge-sdd/worker.go.

## Decisões do projeto aplicadas
FLX-04, RN-3 da 123 (schema 1.0 sem campo novo).

## O que já existe
`recordJobTelemetry` grava um registro por job; `usage.go` lê o evento `result`. Estender: registro inclui a task (sem campo novo no schema: usa `feature` com o caminho da spec da task e `phase` = papel) e `medicao: real` quando houver evento.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`FLX-04`, `SEG-05`, `DUP-04`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/monitor/... ./cmd/forge-sdd/ -run "Telemetry|Telemetria|Session"
go build ./... && go vet ./...
```
