# feat-90-05 — Base da iniciativa e PR aberto pelo Go (`forge pipeline pr`)

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** feat-90-01

## Problema
O Builder faz commit, push e PR à mão: custa turnos e erra a base. `dispatch.go:220` e `goal.go:387` usam o `base_branch` global (`main`); o PR #87 abriu contra `main` (+3.908 linhas) e o #86 contra a branch de outro fix.

## Comportamento
1. **Base da iniciativa:** `queue.json` ganha `base_branch`; `forge queue enqueue --base <branch>` grava. Resolvedor único (flag, depois fila, depois origem da branch no reflog "Created from", por último `forge.json`); `dispatch` e `/toca-ai` usam-no, e a saída diz de qual fonte veio.
2. **`forge pipeline pr --id <job> --body-file <arq>`:** roda a Camada A (`forge gate`); recusa branch `main`, branch igual à base, corpo sem as seções de `.github/pull_request_template.md` e ausência de commit à frente da base; `git push origin <branch>:<branch>`; `gh pr create --base <base>` (ou corrige a base de PR já aberto); registra o PR (`bus.UpdatePR`) e marca as tasks `review_pending`. Runner de `git` e `gh` injetável, argv explícito (`SEG-03`). Nunca mescla nem aprova.
3. `worker-job.md` e `BuildPrompt` trocam o passo do PR por esse comando, dentro do orçamento P-24; goldens atualizados. `forge pipeline pr-opened` continua para quem abre à mão.

## Modelo do módulo
**Enxuto** (ARQ-11): `internal/pipeline/pr.go`, `internal/queue`, adaptador em `cmd/forge-sdd`.

## Decisões do projeto aplicadas
ADR-0004 (Builder não mescla), ADR-0012 (D-2), ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Registro do PR | `pr-opened`, `bus.UpdatePR`, `queue.SetTaskStatuses` | Registro após o `gh` | reusar | Mesmo efeito na fila |
| Base do job | `dispatch.go:220`, `goal.go:387` | Duas cópias | extrair resolvedor | `DUP-01`, `DUP-04` |
| Chamada a `gh` no Go | `internal/` | Nenhuma | criar `pipeline/pr.go` | Nome distinto de `deliver.go` (tmux) |
| Corpo do PR | `.github/pull_request_template.md` | Existe | validar seções | `FLX-03` |

## Dependências novas
Nenhuma (`git` e `gh` já em `allow_commands`).

## Rules e IDs da Constituição Aplicáveis
`FLX-02`, `FLX-03`, `FLX-07`, `SEG-01`, `SEG-03`, `DUP-01`, `ARQ-06`, `TST-04`. Rules: `sdd-fluxo-e-portao`, `sdd-seguranca`, `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/queue ./internal/pipeline -run TestBaseDaIniciativa
go test ./internal/pipeline -run TestPipelinePR_
go test ./internal/pipeline -run TestBuildPrompt_PassoDoPR && go test ./internal/scaffold/...
go vet ./... && go test ./... && forge gate
```
- Com `git` e `gh` falsos: push com refspec explícito, PR com a base da iniciativa, correção de base de PR existente, recusas listadas acima.
