# feat-68-03 — SessionPool com Warmup Físico, Falha Parcial e Status de Conectadas

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-03-pool-warmup-fisico-e-status-conectadas
**Tamanho:** `M`
**Depende de:** feat-68-02, feat-68-06
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Escopo (após fatiamento): warmup físico e status; restart/backoff/FailureHistory foram para 68-03b. Faz `SessionPool.Warmup` iniciar um `ProcessRunner` por estação conforme o adaptador configurado. `standby` somente após health check ok; falha leva a `error` com motivo, sem abortar as demais. `HeartbeatMonitor` passa a detectar morte do filho (estado `error`) e `pool/status` expõe `pid`, `agent`, `connected` e `connectedCount`/`total`. `Close` encerra todos os filhos.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão do pool existente; reaproveita máquina de estados e listeners (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Warmup | `internal/acp/pool.go` | Registro em memória | estender | Spawn real; mantém contrato `PoolWarmupResult` |
| Estados e transições | `pool.go` `TransitionState` | `standby/error/offline` | reusar | Sem novos estados |
| Heartbeat | `internal/acp/heartbeat.go` | `Tick` por timestamps | estender | Consulta liveness do runner |
| Status/RPC | `pool.go` `Status()`, `dispatcher.go` | `pool/status` | estender | Campos novos com `omitempty` |
| Config de adaptadores | `internal/config` (68-01) | Bloco `acp` | reusar | Pool recebe adaptadores injetados (ARQ-01: `internal/` não importa `cmd/`) |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-02`, `ARQ-11`, `DUP-03`, `COD-01`, `COD-03`, `TST-03` (casos conformes e de violação), `TST-04`.
- Rules: `sdd-acp.md`, `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05` (pt-BR, sem emojis), `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestPoolPhysical" ./internal/acp/...
```
- 5 estações com adaptador falso: 5 PIDs distintos e `connectedCount == 5`.
- 1 adaptador inválido: 4 conectadas, 1 em `error`, `Warmup` não retorna erro fatal.
- Matar um filho: estação vai a `error` em até 1 tick; `Close` deixa 0 filhos vivos.
- Sem bloco `acp` configurado: estações em `error` com motivo claro (sem falso `standby`).

### Ajuste por 68-06
- Depende também de 68-06. O pool só cria processos para agentes em `acp.enabledAgents`; estações sem agente habilitado ficam "nao configurada" (não é erro). `Status` expõe `total` = estações habilitadas/configuradas (N do chip). Teste: `TestPoolEnabledAgentsOnly`.
