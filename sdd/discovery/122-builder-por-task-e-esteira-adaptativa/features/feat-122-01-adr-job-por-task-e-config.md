# feat-122-01 — ADR-0029, tipo task_build e chaves de configuração

Tasks 1.1 e 1.2. Trilha A.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/bus, internal/config, sdd/spec/adr.

## Decisões do projeto aplicadas
ADR-0004, ADR-0022, ADR-0024; FLX-05 exige a ADR-0029 (job por task, worktree efêmero, pares Builder e Reviewer, junção contínua, teto por disco, contrato compacto entre agentes).

## O que já existe
`internal/bus/job.go` (tipos `feature_build`, `pr_rework`): estender com `task_build`, campos `task_id` e `feature_job`. `internal/config/forge.go` `BuilderConfig`: estender com `granularidade` (`feature`|`task`, padrão `feature`) e `limite_disco` (50..95, padrão 80).

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`FLX-05`, `GOV-01`, `ARQ-01`, `SEG-01`, `DUP-04`, `TST-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/bus/... ./internal/config/...
test -f sdd/spec/adr/adr-0029-*.md
go build ./... && go vet ./...
```
