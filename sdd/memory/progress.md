# Progress — forge-sdd

## Status
```
Fix 52 — Renomear .agent/ -> .agents/ (com migração em update) [x] done
Release v2.3.0 (estável) — PRs #47/#48/#50 promovidos a main [x] publicada
Fix 54 — npm publish 404 na tag latest (token NPM expirado) [x] done
```

## Features ativas
Nenhuma feature `todo`/`doing` pendente. Lista completa em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** Nenhuma feature `todo`.
**Bloqueios:** —

## Handoff da última sessão
- Fix 54: `NPM_TOKEN` expirado (90 dias) causava 404 silencioso (`continue-on-error`) só em `--tag latest`. Usuário renovou token via terminal local; `npm-publish.yml` ganhou verificação pós-publish que falha explicitamente em vez de mascarar erro.
- v2.3.0 promovida a estável via PRs #47/#48/#50 (rebase-merge com `--admin`, autorizado pelo usuário); `/code-review` pré-merge corrigiu 3 achados.

> Histórico completo em `progress-log.md`
