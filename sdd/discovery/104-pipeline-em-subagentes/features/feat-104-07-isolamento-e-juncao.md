# feat-104-07 — Isolamento: worktree por subagente e junção

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementado (aguarda revisão)
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
set -o pipefail
go test -list 'TestWorktreePorSubagente|TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado' ./internal/pipeline | grep -F 'TestJuncaoConflitoEscala' >/dev/null
go test ./internal/pipeline -run 'TestWorktreePorSubagente|TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado'
# negativo: duas tasks que editam o mesmo arquivo geram conflito marcado; worktree ocupado é recusado
go test ./internal/pipeline -run 'TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado'
go vet ./... && go test ./...
```
- Teste negativo: duas tasks que editam o mesmo arquivo geram conflito marcado, sem merge silencioso.

## A confirmar no Gate 1
Estratégia de junção padrão (merge por lote x rebase).

## Decisões da implementação
- Código em `internal/pipeline/leque.go` (arquivo novo ao lado de `worktree.go`, que continua sendo reusado: `CreateWorktree`, `RemoveWorktree`, `WorktreePath`).
- Lock: `harness.AdquirirLock` com a chave `worktree-principal` (não a da iniciativa, para não colidir com o `RunChat`). Dois leques no mesmo repositório se excluem; o lock é liberado em qualquer recusa.
- Recusa: worktree principal fora da branch base, com alteração em arquivo rastreado (`--untracked-files=no`, para não travar por lixo local) ou com lock de outro dono.
- Junção: `git merge --no-ff` por task, na ordem do plano. Conflito: lista os arquivos, `merge --abort`, task `bloqueada`, branch e worktree preservados, menu `resolver`/`pular`/`abortar` (RN-9). Task com worktree sujo é bloqueada sem merge. Sem `reset --hard`, sem `branch -D` em task com trabalho.
- Órfão de execução interrompida: reaproveitado só se vazio (branch sem commits e worktree limpo; remove com `branch -d`); com trabalho, a preparação para e escala.
- Estratégia padrão (A confirmar no Gate 1): merge por lote, não rebase.
- Tasks seguintes a um conflito continuam sendo integradas (mesmo lote é independente por construção).
- Todas as entradas validadas na borda (regex de id, ref sem `..`/`-`, raiz do repositório); `git` por argv; texto de git sanitizado com `notify.Sanitize`.
