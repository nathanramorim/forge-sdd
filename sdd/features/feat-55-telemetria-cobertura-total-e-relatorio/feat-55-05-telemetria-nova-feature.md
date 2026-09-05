# feat-55-05 — Gravação de métricas em `/nova-feature` e `/novo-fix` (branch+spec)

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-05
**Depende de:** —
**Status:** `todo`

## Objetivo

`/nova-feature` cria branch + especificação e hoje não grava telemetria
(diferente de `/novo-fix`, que já grava — ver `.agents/commands/novo-fix.md`).
Alinhar `/nova-feature` ao mesmo padrão.

## Critérios de Aceitação Executáveis

1. `.agents/commands/nova-feature.md` ganha o passo "Gravação de
   Métricas (determinística)", no mesmo padrão de `novo-fix.md`:
   `forge-sdd session record --feature "<arquivo feat criado>" --outcome
   blocked --criterio-atendido=false` (a feature ainda não foi
   implementada nesse ponto do fluxo — mesmo raciocínio já usado em
   `novo-fix.md`).
2. `.md.tmpl` correspondente atualizado (fonte única).
3. Golden testdata atualizado; `go test ./internal/scaffold/...` passa.

## Handoff

Independente das demais tasks desta feature — pode ser feita em
paralelo.
