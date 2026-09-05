# feat-55-04 — Gravação de métricas em `/split-features`

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-04
**Depende de:** —
**Status:** `done`

## Objetivo

Mesma lacuna de `/discovery`: quebrar um plano em features não deixa
telemetria, mesmo sendo um passo que consome tokens e cria vários
arquivos.

## Critérios de Aceitação Executáveis

1. `.agents/commands/split-features.md` ganha o passo "Gravação de
   Métricas (determinística)": `forge-sdd session record --feature
   "<pasta sdd/features/feat-XX-nome/ criada>" --outcome approved
   --criterio-atendido=true` quando `telemetry.enabled`.
2. `.md.tmpl` correspondente atualizado (fonte única).
3. Golden testdata atualizado; `go test ./internal/scaffold/...` passa.

## Handoff

Independente de feat-55-01/02/03 — pode ser feita em paralelo.
