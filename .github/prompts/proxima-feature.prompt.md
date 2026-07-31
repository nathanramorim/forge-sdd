---
description: "Inicia sessão na próxima feature pendente de forge-sdd"
agent: agent
---

Invoque o Orquestrador para conduzir a sessão completa a partir da próxima feature ou fix com status `todo` em `sdd/memory/progress.md`: guardrail de budget de `progress.md`, leitura das tasks e confirmação do plano, criação da branch, delegação ao Builder, revisão, atualização de `progress.md`/`index.md`, release note em `sdd/releases/history.md`, PR automático (`gh pr create --fill`) e gravação de métricas de telemetria (se `telemetry.enabled` em `sdd/.sddrc`).
