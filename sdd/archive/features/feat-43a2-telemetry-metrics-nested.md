# Feature 43a2 — Correção de Mapeamento de Métricas de Telemetria

Esta especificação define o fix para gravação de métricas de telemetria local, refletindo de forma correta o agrupamento físico de features e discoveries em subpastas e assegurando a gravação incondicional de métricas de sessões inativas ou parciais.

## Critérios de Aceitação Executáveis

1. **Granularidade do Caminho da Feature:**
   * O campo `"feature"` no JSON de métricas (`sdd/.metrics/session-*.json`) deve referenciar o caminho relativo exato da especificação trabalhada (ex: `sdd/features/feat-<workitem/hash>-<nome>.md` ou a subtask correspondente `sdd/features/feat-<workitem/hash>-<nome>/task-<id>.md`).
2. **Gravação de Sessões Inativas/timeout/timeouts:**
   * Se o usuário encerrar a sessão por inatividade ou timeout, ou caso a sessão termine sem atingir a finalização da feature (`criterio_atendido: false`), o Orquestrador deve registrar obrigatoriamente a métrica com `outcome: blocked` ou `outcome: rejected`, registrando a telemetria do esforço parcial realizado.
