# feat-72-03 — Falha Legível e Pausa por Limite de Uso

**Branch:** `feat/72-visibilidade-tmux-e-entrega-ao-builder`
**Fase:** 72-03
**Tamanho:** `M`
**Depende de:** 72-02
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Os jobs falharam três vezes por limite de uso do Claude com a razão "o agente saiu com código 1"; a mensagem real só estava no log. Esta task mostra o motivo, devolve o job à fila sem gastar tentativa e pausa o worker até o horário informado.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão do worker e do barramento com uma operação nova (`Requeue`) e um relógio injetável; sem pacote novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Falha do job | `internal/pipeline/worker.go`, `internal/bus/store.go` (`Fail`, `Recover`) | Mover para `failed` ou devolver por lease vencido | estender | motivo com a última linha útil; nova `Requeue` |
| Aviso à pessoa | `internal/notify` (`TypeInfo`, inbox) | Notificação `info` | reusar | aviso de pausa gravado no inbox |
| Saída do agente | `LiveView` (72-02) | Últimas linhas da saída disponíveis | reusar | a classificação lê as últimas linhas capturadas |
| Contrato do job | `internal/bus/job.go` | Eventos com nota por transição | reusar | evento `adiado` com o motivo; `attempt` intacto |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-05`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. Motivo legível: `falhou: o agente saiu com código 1: <última linha útil>` (saneada, sem ANSI, até 200 caracteres, sem segredo), também no evento do job.
2. `bus.Requeue(dir, role, id, worker, reason, now)`: exige dono do lease; `processing` → `pending` sem alterar `attempt`, preservando o nome do arquivo (prioridade e data); evento `adiado` com o motivo. `forge-sdd bus` ganha o subcomando `requeue` para uso manual.
3. Classificação de limite de uso: só com saída diferente de zero e só nas últimas linhas; padrões por CLI (Claude: "hit your ... limit" com "resets <hora> (<fuso>)"; genéricos: "rate limit", "usage limit", "quota", "429"); extrai o horário (com fuso quando houver).
4. Pausa: o job volta à fila, o worker mostra `⏸ limite de uso do <agente>: retomo às <hora> (faltam <tempo>)` com contagem a cada minuto, responde a `Ctrl+C`, e retoma no horário. Teto de 6 horas; horário não entendido vira 30 minutos; passou do teto o job falha com o motivo.
5. Um aviso `info` é gravado uma vez por pausa no inbox ("Builder pausado: limite de uso do claude até 22:40").

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/bus/... ./internal/pipeline/... ./internal/notify/...
```
- `TestRequeue`: dono exigido, `attempt` intacto, ordem da fila preservada, evento gravado, recusa de worker alheio ou vazio.
- `TestWorker_MotivoLegivel`: a razão da falha inclui a última linha útil, saneada e truncada.
- `TestWorker_LimiteDeUso`: mensagens reais do Claude e genéricas devolvem o job, gravam o `info`, pausam até o horário (relógio injetado) e retomam; sem match o job falha como hoje; mensagem com saída zero não conta; teto de 6 h; fuso lido da mensagem.
- Falso positivo: texto "rate limit" no meio de uma saída longa com código 0 não pausa.
