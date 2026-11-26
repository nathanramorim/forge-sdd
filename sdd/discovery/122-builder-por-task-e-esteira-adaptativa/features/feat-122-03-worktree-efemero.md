# feat-122-03 — Worktree efêmero apagado no commit

Task 1.4. Trilha B. Depende de feat-122-01.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline (worktree, leque).

## Decisões do projeto aplicadas
RN-4 da spec; ADR-0022 (branch por feature); SEG-02 (só dentro de layout.Worktrees).

## O que já existe
`worktree.go` `CreateWorktree`/`RemoveWorktree` e `leque.go` `prepararTask`: reusar branch `forge/<ini>-<task>` e proteções; mudar o ciclo para apagar o worktree logo após o commit da task, mantendo a branch. Órfão continua virando menu (`ErroOrfao`). Retrabalho recria o worktree a partir da branch da task.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-02`, `SEG-03`, `TST-04`, `DUP-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run "Efemero|Leque|Worktree"
go build ./... && go vet ./...
```
