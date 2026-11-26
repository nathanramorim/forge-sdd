# feat-122-09 — Retomada após queda de sessão ou máquina

Task 3.4. Integração. Depende de feat-122-03, feat-122-05 e feat-122-07.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline (worker, leque, handoff), internal/bus.

## Decisões do projeto aplicadas
RN-16 a RN-19 da spec. Reusa `bus` (lease e recover), `OutcomeDeferred`/`usagelimit.go`, `ErroOrfao`, `ReabrirTask` e o handoff compacto da feat-122-02.

## O que já existe
`RunWorker` recupera leases vencidos ao subir; limite de uso devolve o job sem gastar tentativa; órfão com trabalho vira menu. Falta: worktree sujo virar retomada automática do mesmo job, handoff de retomada compacto e commit de checkpoint `wip(<task>)`.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-02`, `SEG-03`, `TST-03`, `TST-04`, `DUP-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run Retomada
go build ./... && go vet ./...
```
