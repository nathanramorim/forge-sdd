# Status — Iniciativa 96

**Iniciativa:** 96-limpeza-worktrees
**Branch:** feat/96-limpeza-worktrees
**Status:** spec_pronta — aguardando Gate 1 (aprovacao humana)
**Proximo passo:** Lider revisa feat-96-01; se aprovada, enfileirar com `forge queue enqueue`.

## Features

| ID | Nome | Status |
|---|---|---|
| feat-96-01 | Deteccao e limpeza de worktrees ociosos na inicializacao da sessao | ready |

## A confirmar

- [ ] Gate 1: aprovacao humana da spec (feat-96-01).
- [ ] `huh` v0.3 oferece `NewMultiSelect`? Se nao, usar `NewSelect` com opcoes listadas uma a uma.
      Confirmar antes do Builder implementar o prompt em `up.go`.
- [ ] A protecao de task `review_pending` e suficiente para cobrir `pr_rework`?
      Hoje o `pr_rework` cria um novo job com novo ID; o worktree do `feature_build` original
      fica com job `done` e, se a task voltou para `changes_requested`, sera protegido.
      Confirmar se ha cenario onde a task fica `approved` prematuramente antes do rework encerrar.
- [ ] `queue.Load` aceita caminho quando `queue.json` nao existe e retorna erro controlado (ou
      vazio)?  Verificar API antes de implementar o fallback "sem queue.json = terminal".
