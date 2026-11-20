# feat-104-11 — Conjunto de avaliação: 10 a 20 features de referência

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementada (2026-11-20)
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
set -o pipefail
go test -list 'TestEvalComparaModos|TestEvalCriteriosTemNegativo' ./internal/monitor | grep -F 'TestEvalComparaModos' >/dev/null
go test -list 'TestEvalComparaModos|TestEvalCriteriosTemNegativo' ./internal/monitor | grep -F 'TestEvalCriteriosTemNegativo' >/dev/null
go test -list 'TestEvalSimulacaoNeutra' ./internal/monitor | grep -F 'TestEvalSimulacaoNeutra' >/dev/null
go test ./internal/monitor -run 'TestEvalComparaModos|TestEvalCriteriosTemNegativo|TestEvalSimulacaoNeutra'
# negativo: compare recusa dados simulados sem --permitir-simulado e o simulador não pode favorecer um modo
go test ./internal/monitor -run 'TestEvalComparaModos|TestEvalSimulacaoNeutra'
# negativo: feature de referência cujo critério não falha no caso ruim reprova o conjunto
go test ./internal/monitor -run 'TestEvalCriteriosTemNegativo'
go vet ./... && go test ./...
```
- Teste negativo: feature de referência cujo critério não falha no caso ruim reprova o conjunto; `eval compare` recusa números simulados sem `--permitir-simulado`; simulação com valores por modo reprova o conjunto.

## Decisões da implementação
- **Quantas features:** 10 (piso da faixa; o carregador aceita de 10 a 20, então crescer não muda código).
- **Repositório de exemplo:** módulo Go isolado em `sdd/eval/exemplo/` (go.mod próprio, fora do `go test ./...` da raiz), sem repositório externo.
- **Validação:** `LoadEvalRefs` reusa `gate.ExtractCriteriaBlock` e `gate.CriteriaProblems` (P-25); uma referência com problema, uma simulação incompleta ou um total fora de 10 a 20 reprovam o conjunto inteiro.
- **Simulador neutro:** `forge eval run` usa só o `simulacao.json` de cada referência, com um único desfecho aplicado igual aos dois modos (o formato por modo é rejeitado). Motivo: a primeira versão tinha valores escritos à mão que favoreciam o modo `chat`, o desenho que a iniciativa propõe, e isso transformava a avaliação numa conclusão pronta. O simulador testa a leitura, a agregação, a gravação e a comparação, não os desenhos.
- **Origem explícita:** `eval-<modo>.json` grava `"origem": "simulado"|"real"` (arquivo sem origem válida é rejeitado); a linha do `eval run` começa com "SIMULADO: números de exemplo, não medidos"; `eval compare` recusa comparar quando algum lado é simulado, a menos que receba `--permitir-simulado`, e então imprime o aviso no topo.
- **Execução real:** não entra nesta feature: tem custo e ficará atrás de um comando manual numa feature posterior. Sem ela, o conjunto não compara desenhos.
- **Métricas:** `sdd/.metrics/eval-<modo>.json` tem estrutura própria; `schema.json` (por sessão) não foi alterado, e os leitores de telemetria já filtram `session-*.json`.

## Status
- [x] Implementada em 2026-11-20 (branch `feat/104-11-avaliacao`).
