# Progress — forge-sdd

## Status
```
Release v3.2.1-beta — Three.js 3D Mascotes [x] done
Feat 64/65/66 — ACP, Pool, Telemetria [x] done (PR #61-#63)
Feat 67 — UX Studio, Handoff 3D [x] done (PR #67 aberto)
Studio UX — esteira 4 passos + Visao Simples/Tecnica [x] e99793e
```

## Features ativas
- `feat/67-ux-studio-animacao-handoff-linguagem-clara`: PR #67 aberto, base `release/v3.0.1` (v3.2.2-beta). NAO mergear em main.

## Pendente (retomar aqui)
- Chip "5 SESSOES PRONTAS" engana: `acp.SessionPool.Warmup` (internal/acp/pool.go) so cria registros em memoria (estado standby), sem processo/agente fisico. Decidir: (a) rotular honestamente ou (b) subir processo ACP real por estacao.
- Sem emojis na UI.

**Bloqueios:** decisao (a)/(b) acima.

> Histórico completo em `progress-log.md`
