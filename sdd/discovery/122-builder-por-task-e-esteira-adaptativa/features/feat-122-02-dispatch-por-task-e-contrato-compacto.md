# feat-122-02 — Dispatch por task e contrato compacto entre agentes

Tasks 1.3 e 1.6. Trilha A. Depende de feat-122-01.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline (dispatch, prompt, handoff), cmd/forge-sdd.

## Decisões do projeto aplicadas
ADR-0024 (tier por task, fallback do papel), RN-1, RN-2, RN-14, RN-15 da spec.

## O que já existe
`dispatch.go` enfileira um `feature_build` por feature: estender para um `task_build` por task quando `granularidade=task`, com o pacote de `forge context <task>`. `prompt.go` `BuildPrompt` e `handoff.go` `BuildHandoffCard`: versão compacta para agente (seções em tags XML, chave e valor, listas em tabela de cabeçalho único, sem moldura nem emoji, teto de bytes). Retorno do Builder em JSON validado. Mensagem humana em tópicos gerada pelo Go. `BuilderHints` de paralelo removido no modo task.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `DUP-03`, `DUP-04`, `COD-04`, `SEG-08`, `COD-01`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run "PorTask|Compacto|Humano"
go build ./... && go vet ./...
```
