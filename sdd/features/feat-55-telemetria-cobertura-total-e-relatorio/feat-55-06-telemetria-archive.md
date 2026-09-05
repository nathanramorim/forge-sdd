# feat-55-06 — Gravação de métricas em `/archive`

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-06
**Depende de:** —
**Status:** `todo`

## Objetivo

Fecha a cobertura dos comandos que mudam estado do ciclo SDD: `/archive`
compacta `progress.md` em `progress-log.md` e hoje também não grava
telemetria.

## Critérios de Aceitação Executáveis

1. `.agents/commands/archive.md` ganha o passo "Gravação de Métricas
   (determinística)": `forge-sdd session record --feature
   "sdd/memory/progress.md" --outcome approved --criterio-atendido=true`
   quando `telemetry.enabled` (não há feature/fix específico associado a
   um archive — o valor fixo do path do próprio `progress.md` documenta
   que a ação foi uma compactação de histórico, não uma entrega).
2. `.md.tmpl` correspondente atualizado (fonte única).
3. Golden testdata atualizado; `go test ./internal/scaffold/...` passa.

## Handoff

Última task de cobertura de comandos — fecha o escopo definido na
discovery-55 (comandos que mudam estado: discovery, split-features,
nova-feature, novo-fix, proxima-feature, revisar, archive).
