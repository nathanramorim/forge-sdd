# Progress — forge-sdd

## Status
```
Fix 52 — Renomear .agent/ -> .agents/ (com migração em update) [x] done
Feature 03 — Ergonomia de comandos e sincronização (6 subtarefas) [x] done
Release v2.3.0 (estável) — PRs #47/#48/#50 promovidos a main [x] publicada
```

## Features ativas
Fix 54 — npm publish falha (404) na tag `latest` (`todo`). Lista completa em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** Fix 54 requer acesso à conta npmjs.com (fora do alcance de agente) para checar escopo do `NPM_TOKEN`.
**Bloqueios:** `npm view @nathanramorim/forge-sdd dist-tags` ainda mostra `latest: 1.9.4`.

## Handoff da última sessão
- v2.3.0 promovida a estável: `/code-review` pré-merge corrigiu 3 achados (cheatsheet, orçamento lessons.md, colisão telemetria); PRs rebase-merged com `--admin` (bypass autorizado pelo usuário).
- Discovery-53 incorporado de branch órfã; `/split-features` pendente de decisão sobre paridade Gemini/Copilot.

> Histórico completo em `progress-log.md`
