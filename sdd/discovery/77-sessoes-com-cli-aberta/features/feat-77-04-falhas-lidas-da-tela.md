# feat-77-04 — Falhas lidas da tela

**Branch:** `feat/77-sessoes-com-cli-aberta`
**Fase:** 77-04
**Tamanho:** `M`
**Depende de:** 77-03
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Detecção de anomalias na tela do painel (`tmux capture-pane`) durante a execução no modo `live`: limite de uso (rate limit / quota exceeded), agente indisponível (login inválido, tier incorreto) ou fechamento abrupto do painel.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Leitura e sanitização de telas de terminal, parsing resiliente de erros e integração com barramento de jobs e notificações (`ARQ-11`, `COD-01`, `SEG-05`).

## Decisões do projeto aplicadas
- `D-6 (ADR-0007)`: Classificação de erros a partir das últimas 20 linhas de `capture-pane` usando `DetectUsageLimit` e `DetectAgentUnavailable`.
- `Requeue`: Devolução atômica do job para `pending` com motivo explícito sem queimar tentativas indevidamente.
- `SEG-05`: Máscara de segredos em logs e avisos.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Detector de limite de uso | `internal/pipeline/usagelimit.go` | `DetectUsageLimit` | reusar | Aplicar sobre as linhas lidas da tela. |
| Detector de agente indisponível | `internal/pipeline/agentstate.go` | `DetectAgentUnavailable` | reusar | Aplicar sobre a tela do painel interativo. |
| Devolução de jobs | `internal/bus/requeue.go` | `bus.Requeue` | reusar | Devolver job preservando ordem da fila. |
| Notificações | `internal/notify` | `Dispatch` | reusar | Enviar alerta ao inbox da Discovery em falhas críticas. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-05`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Ler periodicamente as últimas 20 linhas do painel via `capture-pane` enquanto aguarda término do job.
2. Identificar limite de uso: acionar `bus.Requeue`, pausar despachante até horário indicado ou padrão, e notificar inbox.
3. Identificar agente indisponível: acionar `bus.Requeue`, emitir aviso ao inbox e interromper despachante com instrução de troca de agente.
4. Identificar painel fechado ou sessão destruída: acionar `bus.Requeue` e parar.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/pipeline -run DeliverTela
```
- Simulações de tela com rate limit e IneligibleTierError acionam requeue correto, pausa e mensagens no inbox.
