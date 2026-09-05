# Progress — forge-sdd

## Status
```
Fix 54 — npm publish 404 na tag latest (token NPM expirado) [x] done
Feature 55 — Telemetria cobertura total + /report (7 subtarefas) [x] done
```

## Features ativas
Nenhuma feature `todo`/`doing` pendente. Lista completa em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** Nenhuma feature `todo`. Branch `feat/55-telemetria-cobertura-total-e-relatorio` pronta para `/revisar` + PR.
**Bloqueios:** —

## Handoff da última sessão
- Feature 55 done: `session record` cobre `/discovery`, `/split-features`, `/nova-feature`, `/archive`; novo `forge-sdd report` (tokens/modelo/duração por item + idade do projeto), reaproveitando `AggregateSessionMetrics`.
- Fix 54: `NPM_TOKEN` expirado causava 404 silencioso em `--tag latest`; corrigido e `npm-publish.yml` agora falha explícito.

> Histórico completo em `progress-log.md`
