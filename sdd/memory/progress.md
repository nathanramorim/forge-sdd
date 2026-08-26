# Progress — forge-sdd

## Status
```
Fix 52 — Renomear .agent/ -> .agents/ (com migração em update) [x] done
Discovery 03 — Ergonomia de comandos e sincronização [x] split em feat-03 (6 subtarefas, todas todo)
```

## Features ativas
`feat-03-ergonomia-de-comandos-e-sincronizacao` (subpasta, branch única `feat/03-ergonomia-de-comandos-e-sincronizacao`) — 6 subtarefas `todo`: 03-01 (nome adaptador Claude), 03-02 (migração update), 03-03 (sync remota /status), 03-04 (clarify), 03-05 (confirmação subagente), 03-06 (docs/release). Lista completa em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** `/proxima-feature` a partir de `feat-03-01-correcao-nome-adaptador-claude.md` (menor risco, maior impacto imediato — comando Claude hoje não executa por padrão). `feat-02`+`fix-52` (PR #48, draft) e `v2.0.0-beta.0` (PR #47) seguem aguardando decisão de promoção a estável.
**Bloqueios:** —

## Handoff da última sessão
- fix-52 concluída: `.agent/` renomeada para `.agents/` (templates, dogfood, golden fixtures, adaptadores, docs). Nova `migrateLegacyAgentDir` em `scaffold.go` migra projetos com `.agent/` legado em `update`, preservando `rules/` do usuário. Ver `sdd/features/fix-52-rename-agent-dir-para-agents.md`.
- Discovery 03 criada e quebrada em 6 features (`sdd/features/feat-03-ergonomia-de-comandos-e-sincronizacao/`): comando Claude quebrado (`.prompt.md` vs. `/nova-feature` sem sufixo), `/status` sem sincronização remota (achado nesta mesma sessão: branch órfã `claude/agent-folder-rename-forge-ww8abl` sem PR), clarify ausente em `/nova-feature`/`/novo-fix`/`/discovery`, e confirmação de delegação a subagente no lifecycle.

## Última sessão
- 2026-08-26 — fix-52 concluída (rename `.agent/` -> `.agents/` + migração em update); discovery-03 criada e quebrada em feat-03 (6 subtarefas).

> Histórico completo em `progress-log.md`
