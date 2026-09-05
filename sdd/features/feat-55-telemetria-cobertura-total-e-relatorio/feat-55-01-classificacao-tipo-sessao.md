# feat-55-01 — Classificação de tipo de sessão (discovery/feature/fix)

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-01
**Depende de:** —
**Status:** `todo`

## Objetivo

Base para o comando `forge-sdd report` (feat-55-02): uma função que
classifica um `sessionMetrics.Feature` (caminho relativo) como
`discovery`, `feature` ou `fix`, sem exigir migração das métricas já
gravadas (que não têm esse campo hoje).

## Critérios de Aceitação Executáveis

1. Nova função `ClassifySessionType(feature string) string` em
   `cmd/forge-sdd/session.go`, retornando `"discovery"` para paths sob
   `sdd/discovery/`, `"fix"` para paths sob `sdd/features/` que contenham
   `fix-` no nome do arquivo/pasta, `"feature"` para os demais paths sob
   `sdd/features/`, e `"outro"` para qualquer coisa fora desses prefixos
   (não deve gerar erro nem pânico).
2. Teste unitário cobre os 4 casos acima, incluindo um path com
   `feat-XX-nome/fix-YY-nome.md` (fix aninhado dentro de pasta de feature)
   classificado como `fix`.
3. `go build ./...` e `go test ./...` continuam passando.

## Handoff

Task de base — feat-55-02 depende desta função para agrupar por tipo no
relatório.
