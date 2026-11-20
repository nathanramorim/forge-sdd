# feat-104-11 — Conjunto de avaliação: 10 a 20 features de referência

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-02

## Problema
Sem medição não dá para dizer se subagentes no chat valem mais que 3 sessões no terminal.

## Comportamento
1. `sdd/eval/refs/` com 10 a 20 features de referência pequenas (spec e critério executável com teste negativo, validado pela feat-104-02), em um repositório de exemplo isolado.
2. `forge eval run --mode sessions|chat` roda a suíte e grava taxa de aprovação, voltas do Retorno, tokens, tempo e intervenções humanas em `sdd/.metrics/eval-<modo>.json`.
3. `forge eval compare` imprime a tabela dos dois modos.
4. A suíte roda sem modelo (simulador) em CI; com modelo, só manual.

## Modelo do módulo
**Enxuto**: estende `internal/monitor` (métricas) e `cmd/forge-sdd` (flags).

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0014 (medir antes de mudar).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Telemetria real por job | `internal/monitor`, `forge report --jobs` | Tokens, turnos, custo | **Reusar** | `DUP-01` |
| Fixtures de e2e | `internal/pipeline/e2e_test.go` | Cenários sem modelo | **Reusar** o simulador | — |
| Esquema de métricas | `sdd/.metrics/schema.json` | Por sessão | **Estender** | Compatível |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `TST-01`, `TST-05`, `SEG-05`. Rules: `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestEvalComparaModos|TestEvalCriteriosTemNegativo' ./internal/monitor | grep -q 'TestEvalComparaModos'
go test ./internal/monitor -run 'TestEvalComparaModos|TestEvalCriteriosTemNegativo'
go vet ./... && go test ./...
```
- Teste negativo: feature de referência cujo critério não falha no caso ruim reprova o conjunto.

## A confirmar no Gate 1
Quantas features (10 ou 20) e qual repositório de exemplo.
