# Progress — forge-sdd

## Status
```
Fix 52 — Renomear .agent/ -> .agents/ (com migração em update) [x] done
```

## Features ativas
Nenhuma feature `todo`/`doing` pendente. Lista completa (status + branch) em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** Nenhuma feature `todo`. `feat-02`+`fix-52` (branch `claude/agent-folder-rename-forge-ww8abl`) e `v2.0.0-beta.0` aguardam push/PR/decisão de promoção a estável.
**Bloqueios:** —

## Handoff da última sessão
- fix-52 concluída: `.agent/` renomeada para `.agents/` (templates, dogfood, golden fixtures, adaptadores, docs). Nova `migrateLegacyAgentDir` em `scaffold.go` migra projetos com `.agent/` legado em `update`, preservando `rules/` do usuário. Ver `sdd/features/fix-52-rename-agent-dir-para-agents.md`.

## Última sessão
- 2026-08-26 — fix-52 concluída (rename `.agent/` -> `.agents/` + migração em update).

> Histórico completo em `progress-log.md`
