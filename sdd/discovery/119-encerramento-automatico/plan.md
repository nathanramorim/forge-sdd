# 119 — Plano

Quebra de baixo para cima; máx. 7 tasks por onda.

## Onda 1

| Task | O quê | Depende de | Rules e IDs |
|---|---|---|---|
| fix-119-01 | critério único de "concluída"; `ProcurarPRMesclado`; `forge archive --reconciliar [--aplicar]`; aviso do `doctor`; texto do CLOSE nas skills | — | `DUP-01`, `DUP-02`, `ARQ-01`, `ARQ-05`, `ARQ-10`, `COD-01`, `TST-02`, `TST-03`; rules `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao` |

Uma task só (escopo M); o Builder pode dividi-la em PRs menores se passar de ~400 linhas. Detalhe: `fixes/fix-119-01-encerramento-de-entrega-concluida.md`.
