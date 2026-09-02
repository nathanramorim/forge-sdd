# Progress — forge-sdd

## Status
```
Fix 52 — Renomear .agent/ -> .agents/ (com migração em update) [x] done
Feature 03 — Ergonomia de comandos e sincronização (6 subtarefas) [x] done
Release v2.3.0-beta — acumula feat-02 + fix-52 + feat-03 [x] tagueada
```

## Features ativas
Nenhuma feature `todo`/`doing` pendente. Lista completa (status + branch) em `sdd/features/index.md`.

## Próximo passo
**Iniciar:** Nenhuma feature `todo`. `feat-02`+`fix-52`+`feat-03` (branch `feat/03-ergonomia-de-comandos-e-sincronizacao`, tag `v2.3.0-beta`) e `v2.0.0-beta.0` (PR #47) aguardam decisão de promoção a estável.
**Bloqueios:** —

## Handoff da última sessão
- `/status` detectou branch remota órfã `discovery/03-multi-sessao-esteira-feature` (sem entrada em `index.md`, colisão de numeração "03" com `feat-03-ergonomia...`). Investigada e incorporada: `discovery-53`/`criteria-53`/`plan-53-multi-sessao-esteira-feature.md` (esteira Spec→Act→Revisor em 3 sessões, subagentes ou Claude Code Remote). `/split-features` fica pendente de decisão do usuário sobre paridade Gemini/Copilot (sem isolamento automatizável).
- Executada de ponta a ponta, em subagente, a feature `feat-03-ergonomia-de-comandos-e-sincronizacao` (6 subtarefas, todas `done`): comando Claude corrigido (`.claude/commands/*.md`, sem `.prompt`), `forge-sdd update` limpa o nome antigo, `/status` ganhou sincronização remota (`git fetch` + `gh pr list` + seção "Divergência Remota"), clarify condicional em `/nova-feature`/`/novo-fix`/`/discovery` (`sdd/memory/clarify.md`), e confirmação de delegação a subagente no passo `PLAN` do lifecycle.
- Bug pré-existente encontrado e corrigido durante os testes de 03-02: `cleanObsoleteFiles()` rodava mesmo em `--dry-run`, apagando arquivos reais — corrigido com guarda `!cfg.DryRun`.
- Entregas beta acumuladas (feat-02 + fix-52 + feat-03) publicadas como `v2.3.0-beta` (Regra 11 da Constituição — beta não mescla em `main`, tag na própria branch de feature). Ver `sdd/releases/history.md`.

## Última sessão
- 2026-08-26 — feat-03 concluída (6/6 subtarefas `done`); release `v2.3.0-beta` tagueada acumulando feat-02+fix-52+feat-03.

> Histórico completo em `progress-log.md`
