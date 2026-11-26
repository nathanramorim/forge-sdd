# feat-122-08 — Status por task e PR da feature

Tasks 3.1 e 3.2. Integração. Depende de feat-122-05 e feat-122-07.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: cmd/forge-sdd/status.go, internal/pipeline (pr_por_feature).

## Decisões do projeto aplicadas
ADR-0022, ADR-0023, FLX-03.

## O que já existe
`forge status` mostra pares ativos, vagas de disco e tokens por task. PR em draft da feature sai quando todas as tasks estão integradas (reusa `pr_por_feature.go`); squash opcional no merge fica com a pessoa.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`COD-01`, `FLX-03`, `TST-02`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./cmd/forge-sdd/ -run Status ./internal/pipeline/ -run "PRPorFeature|Feature"
go build ./... && go vet ./... && go test ./...
```
