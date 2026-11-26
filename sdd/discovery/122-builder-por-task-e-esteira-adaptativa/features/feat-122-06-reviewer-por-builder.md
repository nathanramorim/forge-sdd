# feat-122-06 — Um Reviewer por Builder ativo

Task 2.3. Trilha A. Depende de feat-122-02.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline (review), internal/bus.

## Decisões do projeto aplicadas
RN-8 da spec; FLX-06; FLX-08 (diff como dado não confiável); SEG-06 (Reviewer só leitura).

## O que já existe
Hoje um Reviewer atende a fila por PR. Estender: cada commit de task gera um job de review com só o diff `base...forge/<ini>-<task>` e a spec da task, tier econômico, sem worktree; até um Reviewer ativo por Builder ativo.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`FLX-06`, `FLX-08`, `SEG-06`, `TST-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run "Par|Review"
go build ./... && go vet ./...
```
