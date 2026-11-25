# feat-122-07 — Junção contínua na branch da feature

Task 2.4. Trilha B. Depende de feat-122-03.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline (leque, onda).

## Decisões do projeto aplicadas
RN-9 da spec; contrato do leque (RN-9 da 104).

## O que já existe
`leque.go` `Juntar` junta no fim da onda. Estender: juntar cada task assim que aprovada, em série sob o lock, respeitando a ordem do plano (task posterior aprovada antes espera a anterior); conflito bloqueia, aborta o merge e preserva a branch; integrada apaga a branch da task.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-03`, `TST-03`, `DUP-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run "Juncao|Leque"
go build ./... && go vet ./...
```
