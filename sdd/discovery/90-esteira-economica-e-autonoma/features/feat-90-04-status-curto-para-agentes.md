# feat-90-04 — `forge status` curto para agentes

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** nenhuma

## Problema
`forge status` tem 12,8 KB e `--compact` (prometido abaixo de 200 tokens) entrega 6,2 KB porque o "Ao Vivo" despeja o JSON cru do log. Todo comando lê isso no início.

## Comportamento
1. `captureLiveLines` (`internal/pipeline/status_live.go`) formata os eventos pelo `monitor.ChatFeedFormatter`; nunca imprime JSON cru.
2. Sem terminal (`isatty` falso), `forge status` sai curto: até 1,5 KB, só iniciativas ativas, fila e próximo passo. `--full` mostra tudo. Com terminal, nada muda.
3. Testes que esperam a saída completa passam a usar `--full`.

## Modelo do módulo
**Enxuto** (ARQ-11): estende `internal/pipeline` e `cmd/forge-sdd/status.go`.

## Decisões do projeto aplicadas
ADR-0011 (painel), ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Formatador | `monitor.ChatFeedFormatter` | Usado no worker | reusar | `DUP-01` |
| Modo curto | `forge status --compact` | Existe, sem uso | estender como padrão sem terminal | Não depende do agente lembrar |
| Detecção de terminal | `status.go` (`isatty`) | Já importado | reusar | — |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-10`, `COD-01`, `TST-02`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test ./cmd/forge-sdd ./internal/pipeline -run 'TestStatusCurtoSemTerminal|TestAoVivoSemJSONCru'
go vet ./... && go test ./...
```
- Cenário com 17 iniciativas e eventos `stream-json`: saída até 1,5 KB sem terminal e sem `{"type"`.
- `--full` mantém a saída completa, com o "Ao Vivo" formatado.
