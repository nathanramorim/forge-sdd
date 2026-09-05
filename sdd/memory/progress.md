# Progress — forge-sdd

## Status
```
Release v2.3.0 (estável) — PRs #47/#48/#50 promovidos a main [x] publicada
Fix 54 — npm publish 404 na tag latest (token NPM expirado) [x] done
Feature 55 — Telemetria cobertura total + /report (7 subtarefas) [ ] todo
```

## Features ativas
Feature 55 (55-01..07), branch `feat/55-telemetria-cobertura-total-e-relatorio`, todas `todo`. Lista completa em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** `/proxima-feature` — começar por 55-01 (classificação de tipo de sessão), pré-requisito de 55-02.
**Bloqueios:** —

## Handoff da última sessão
- Discovery-55+split: cobre lacuna de telemetria (só 3/13 comandos gravavam) com `session record` em discovery/split-features/nova-feature/archive + novo `forge-sdd report` (tokens/modelo/duração + idade do projeto).
- Fix 54: `NPM_TOKEN` expirado causava 404 silencioso em `--tag latest`; corrigido e `npm-publish.yml` agora falha explícito.

> Histórico completo em `progress-log.md`
