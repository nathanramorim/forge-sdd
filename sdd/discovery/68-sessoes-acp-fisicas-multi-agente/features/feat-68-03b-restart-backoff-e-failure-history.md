# feat-68-03b — Restart Automático com Backoff Exponencial e FailureHistory

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-03b
**Tamanho:** `M`
**Depende de:** `feat-68-03`
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Parte fatiada da 68-03. Quando o filho de uma estação cai, o pool reinicia com backoff e registra o histórico de falhas que alimenta o painel da 68-05.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão do pool/runner em `internal/acp` com política de restart pura e relógio/sleeper injetável (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Pool físico | `internal/acp/pool.go` (68-03) | Warmup físico, `Status` | estender | Observa `Exited()` do runner e reinicia |
| Saída do processo | `internal/acp/process.go` (68-02) | `ExitInfo` | reusar | Causa e exit code para o histórico |
| Estados | `internal/acp/types.go` | `SessionState` | estender | Novo valor `restarting` em `ValidSessionStates` |
| Config de restart | `internal/config` (68-01) | `acp.restart.*` | reusar | Valores configuráveis |
| Backoff/histórico | `internal/acp` | Inexistente | criar | `restart.go`, `failures.go` e testes (`COD-01`) |

## Dependências novas
- Nenhuma. Stdlib (`time`, `math/rand`) (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-02`, `ARQ-11`, `SEG-05` (histórico sem args/env), `COD-01`, `COD-03`, `TST-03`, `TST-04`. Rules: `sdd-acp.md`, `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`. Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo e Critérios de Aceitação Executáveis
- Ao detectar saída do filho, o pool aplica restart com backoff exponencial: `min(initialBackoffMs * multiplier^n, maxBackoffMs)` com jitter; estado novo `restarting` ("reiniciando"); sucesso volta a `standby`; após `maxAttempts` falhas consecutivas vai a `error`. O contador zera após `stableResetMs` estável. Tudo lido de `acp.restart.*`.
- `FailureHistory` em memória por estação (limite fixo de entradas, p.ex. 50): `timestamp`, `cause`, `exitCode`, `attempt`, `state`. Não grava args/env (`SEG-05`). Fonte para 68-05.
- `Close` e cancelamento de contexto interrompem backoffs pendentes (sem goroutine vazada).
- Testes: relógio/sleeper injetável para determinismo (`TST-04`).
- Critérios adicionais:
```bash
go test -race -run "TestPoolRestartBackoff|TestPoolFailureHistory" ./internal/acp/...
```
  - Backoff cresce até o máximo e respeita jitter dentro da faixa.
  - Esgotar tentativas leva a `error`; período estável zera o contador.
  - Histórico guarda causa e exit code e respeita o limite.


