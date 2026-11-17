# feat-90-02 — Telemetria real por job

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** nenhuma

## Problema
`estimateJobTokens` (`cmd/forge-sdd/worker.go`) estima tokens pelo tamanho da spec e do log; o painel mostra 869 mil no Builder enquanto os logs somam 13,6 milhões relidos. Turnos, duração e modelo saem vazios e a versão é fixa em `4.5.0-beta` (`internal/config/config.go`, `internal/acp/dispatcher.go`).

## Comportamento
1. Em `internal/monitor`, ler o log do job (`stream-json`) e somar os eventos `result`: tokens de entrada, cache lido, cache gravado, saída, `num_turns`, `duration_ms`, modelo e `total_cost_usd` (rotulado "preço de tabela").
2. `SessionMetrics` ganha `tokens_cache_read`, `tokens_cache_creation`, `cost_usd_tabela` e `medicao` (`real` ou `estimado`). Sem `result` (agy, gemini, codex, copilot), grava a estimativa com `medicao: estimado`. Campos opcionais: arquivos antigos seguem válidos.
3. O worker chama essa leitura ao fim do job; `estimateJobTokens` sai de `cmd/` (`ARQ-01`, `COD-04`).
4. A versão gravada é a do binário; não sobra versão fixa fora de testes.
5. `forge report --initiative <id>` soma os jobs separando real e estimado; `--jobs` lista cada job (turnos, tokens, modelo). O painel usa o dado real quando existir.
6. Só contadores, modelo, duração e IDs (`SEG-05`).

## Modelo do módulo
**Enxuto** (ARQ-11): estende `internal/monitor` e `cmd/forge-sdd` (flags).

## Decisões do projeto aplicadas
ADR-0009 (D-1 vira o caminho `estimado`), ADR-0014. FLX-04 (telemetria em `sdd/.metrics/`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Registro | `monitor.SessionMetrics`, `RecordSessionMetrics` | Campos de tokens, `turns`, `model` | estender | Esquema compatível |
| Estimativa | `estimateJobTokens`, `EstimateTokensFromBytes` | Por bytes | mover para `monitor` como fallback | `ARQ-01` |
| Relatório | `cmd/forge-sdd/report.go` | Soma por feature | estender; soma vai para `monitor` | `ARQ-01` |
| Painel | `monitor.TokenMetricsPerRole` | Estimativa por papel | reusar com dado real | Mesmo número |
| Versão | `config.go:177`, `acp/dispatcher.go:38` | Fixa | usar a do binário | `DUP-04` |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `DUP-04`, `SEG-05`, `COD-03`, `COD-04`, `TST-04`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-seguranca`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/monitor ./internal/pipeline -run 'TestUsoReal|TestUsoEstimado'
go test ./internal/config ./internal/monitor -run TestVersaoDoBinario
go test ./internal/monitor ./cmd/forge-sdd -run TestReport_UsoRealPorJob
go vet ./... && go test ./...
```
- Com um log de exemplo de dois eventos `result`, os totais batem com a soma; sem `result`, `medicao` é `estimado`.
- O relatório mostra real e estimado em colunas separadas.
