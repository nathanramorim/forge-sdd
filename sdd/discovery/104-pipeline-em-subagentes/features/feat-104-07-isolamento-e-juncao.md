# feat-104-07 — Isolamento: worktree por subagente e junção

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementado; "merge por lote com `--no-ff`" é decisão do Builder e aguarda confirmação do dono no PR
**Depende de:** feat-104-06

## Problema
Workers construíram tasks no mesmo worktree onde outra sessão trabalhava, sem trava. Com subagentes paralelos o risco cresce.

## Comportamento
1. Cada subagente do Builder recebe worktree próprio em `.forge/runtime/worktrees/<iniciativa>/<task>` e branch temporária.
2. Junção por lote, em ordem do plano: `git merge --no-ff` na branch da iniciativa; conflito marca a task `bloqueada` e escala, nunca resolve sozinho.
3. Recusa iniciar se o worktree principal está sujo ou ocupado por outra sessão (lock da feat-104-06).
4. `git` sempre com argv explícito (`SEG-03`).

## Regras de negócio
- Os hooks do git do repositório rodam no merge (decisão do coordenador; sem `--no-verify` nem `core.hooksPath`). Risco: em repositório não confiável um hook executa código arbitrário; use só em repositório do próprio usuário.

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
# negativo: duas tasks que editam o mesmo arquivo geram conflito marcado; worktree ocupado é recusado; falha de merge sem conflito não vira conflito
go test -list 'TestJuncaoFalhaSemConflitoNaoViraConflito|TestJuncaoSemIdentidadeUsaFallback' ./internal/pipeline | grep -F 'TestJuncaoFalhaSemConflitoNaoViraConflito' >/dev/null
go test ./internal/pipeline -run 'TestJuncaoFalhaSemConflitoNaoViraConflito|TestJuncaoSemIdentidadeUsaFallback|TestJuncaoNaoSobrescreve|TestJuncaoIdentidadeGlobal'
go test ./internal/pipeline -run 'TestJuncaoConflitoEscala|TestRecusaWorktreeOcupado'
go vet ./... && go test ./...
```
- Teste negativo: duas tasks que editam o mesmo arquivo geram conflito marcado, sem merge silencioso.


## Decisões da implementação
- Código em `internal/pipeline/leque.go` (arquivo novo ao lado de `worktree.go`, que continua sendo reusado: `CreateWorktree`, `RemoveWorktree`, `WorktreePath`).
- Lock: `harness.AdquirirLock` com a chave `worktree-principal` (não a da iniciativa, para não colidir com o `RunChat`). Dois leques no mesmo repositório se excluem; o lock é liberado em qualquer recusa.
- Recusa: worktree principal fora da branch base, com alteração em arquivo rastreado (`--untracked-files=no`, para não travar por lixo local) ou com lock de outro dono.
- Junção: `git merge --no-ff` por task, na ordem do plano. Conflito: lista os arquivos, `merge --abort`, task `bloqueada`, branch e worktree preservados, menu `resolver`/`pular`/`abortar` (RN-9). Task com worktree sujo é bloqueada sem merge. Sem `reset --hard`, sem `branch -D` em task com trabalho.
- Órfão de execução interrompida: reaproveitado só se vazio (branch sem commits e worktree limpo; remove com `branch -d`); com trabalho, a preparação para e escala.
- Estratégia: "merge por lote com `--no-ff`" (não rebase) é decisão do Builder e aguarda confirmação do dono no PR.
- Tasks seguintes a um conflito continuam sendo integradas (mesmo lote é independente por construção).
- Todas as entradas validadas na borda (regex de id, ref sem `..`/`-`, raiz do repositório); `git` por argv; texto de git sanitizado com `notify.Sanitize`.
- Órfão: só é removido se provar vazio (sem arquivos, inclusive não rastreados; HEAD na branch esperada; `rev-list HEAD ^base` = 0; branch sem commits fora da base). Qualquer dúvida ou erro conta como trabalho e devolve `*ErroOrfao` com menu (retomar/preservar/abortar), sem tocar em nada.
- Fora de `leque.go`: `CreateWorktree` (`worktree.go`) passou a conferir, com `EvalSymlinks`, que o pai do caminho está em `.forge/runtime/worktrees` (e este no repositório) e que o caminho não é link; o conserto natural fica lá para proteger todos os chamadores, não só o leque.
- `Juntar` devolve os resultados parciais junto com o erro fatal. `PrepararLeque` libera o lock por `defer` em erro ou panic; `Juntar` renova a lease antes de cada merge. Os descartes de erro restantes (`removerSeguro`, lista de arquivos em conflito) estão comentados no código.
- Hooks: ver Regras de negócio.
- Identidade do git: o merge só recebe `-c user.name=forge-sdd -c user.email=forge-sdd@localhost` quando `git config --get user.name/user.email` não resolve (CI, contêiner); a identidade do usuário nunca é trocada e nada é gravado na config. Falha de merge sem arquivos em conflito não é rotulada de conflito: a task fica `bloqueada` com o primeiro erro do git sanitizado em `Motivo` e sem `Arquivos`; `merge --abort` só roda com merge em andamento.
