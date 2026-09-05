# feat-55-03 — Gravação de métricas em `/discovery`

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-03
**Depende de:** —
**Status:** `done`

## Objetivo

`/discovery` é o único ponto de entrada do ciclo SDD que nunca grava
telemetria — uma sessão que só faz discovery e para (ou é retomada depois
por outro agente) não deixa rastro nenhum.

## Critérios de Aceitação Executáveis

1. `.agents/commands/discovery.md` ganha um passo final "Gravação de
   Métricas (determinística)", no mesmo padrão de `novo-fix.md`: se
   `telemetry.enabled` em `sdd/.sddrc`, executa `forge-sdd session record
   --feature "<discovery-ID-nome.md gerado>" --outcome approved
   --criterio-atendido=true` (discovery entregue = outcome aprovado; não
   há "critério de aceitação" de discovery no sentido de feature, mas o
   outcome reflete que os 3 artefatos foram gerados com sucesso).
2. O `.md.tmpl` correspondente em
   `internal/scaffold/templates/.agents/commands/discovery.md.tmpl` recebe
   o mesmo passo (fonte única — nunca duplicado por agente).
3. `internal/scaffold/testdata/golden/.agents/commands/discovery.md`
   atualizado para refletir o novo conteúdo; `go test ./internal/scaffold/...`
   (golden tests) passa.

## Handoff

Independente de feat-55-01/02 — pode ser feita em paralelo.
