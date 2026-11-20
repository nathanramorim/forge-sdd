# feat-104-07 — Isolamento: worktree por subagente e junção

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-06

## Problema
Workers construíram tasks no mesmo worktree onde outra sessão trabalhava, sem trava. Com subagentes paralelos o risco cresce.

## Comportamento
1. Cada subagente do Builder recebe worktree próprio em `.forge/runtime/worktrees/<iniciativa>/<task>` e branch temporária.
2. Junção por lote, em ordem do plano: `git merge --no-ff` na branch da iniciativa; conflito marca a task `bloqueada` e escala, nunca resolve sozinho.
3. Recusa iniciar se o worktree principal está sujo ou ocupado por outra sessão (lock da feat-104-06).
4. `git` sempre com argv explícito (`SEG-03`).

## Modelo do módulo
**Enxuto**: estende `internal/pipeline/worktree.go`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0004; `builder.use_git_worktrees`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Worktree por tarefa | `internal/pipeline/worktree.go`, `worktree_gc.go` | Criação e coleta | **Estender** com junção e checagem de sujeira | `DUP-01` |
| Plano de lotes | `forge bus plan`, `PlanWaves` | Ordem por `**Depende de:**` | **Reusar** | — |
| Execução paralela | `.agents/skills/sdd-build/references/worker-job.md` | Subagentes e worktrees | **Reusar** a regra | — |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `COD-03`, `SEG-03`, `TST-01`, `TST-05`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestWorktreePorSubagente|TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado' ./internal/pipeline | grep -q 'TestJuncaoConflitoEscala'
go test ./internal/pipeline -run 'TestWorktreePorSubagente|TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado'
go vet ./... && go test ./...
```
- Teste negativo: duas tasks que editam o mesmo arquivo geram conflito marcado, sem merge silencioso.

## A confirmar no Gate 1
Estratégia de junção padrão (merge por lote x rebase).
