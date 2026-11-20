# feat-104-12 — Observabilidade: rastreio por nó (subagente)

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-06

## Problema
Hoje a telemetria é por sessão; com subagentes não se sabe quem rodou, com que contexto, custo e por que parou.

## Comportamento
1. Cada subagente grava um nó em `sdd/.metrics/trace-<job>.jsonl`: `no`, `papel`, `modelo`, `contexto_bytes`, `tokens`, `custo_usd_tabela`, `motivo_parada` (ok|teto|erro|escalada), `pai`.
2. `forge report --nos` agrega por papel e por job; `session record` segue igual.
3. Sem segredos (`SEG-05`); medição `estimado` marcada como tal.

## Modelo do módulo
**Enxuto**: estende `internal/monitor`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0009, ADR-0011, ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Métricas de sessão | `monitor.SessionMetrics`, `RecordSessionMetrics` | Por sessão | **Estender** | `DUP-01` |
| Relatório | `cmd/forge-sdd/report.go` | `--jobs`, `--initiative` | **Estender** com `--nos` | — |
| Stream do worker | `monitor.ChatFeedFormatter` | Linhas narrativas | **Reusar** | — |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `SEG-05`, `COD-01`, `TST-01`. Rules: `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestTraceGravaNo|TestReportNos' ./internal/monitor | grep -q 'TestTraceGravaNo'
go test ./internal/monitor -run 'TestTraceGravaNo|TestReportNos'
go vet ./... && go test ./...
```
- Teste negativo: nó sem `motivo_parada` é recusado.

## A confirmar no Gate 1
Granularidade (um nó por subagente ou por chamada de ferramenta).
