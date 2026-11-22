# Etapa 1 — Conferir a spec da 115

**Objetivo:** Veredito do Reviewer sobre o PR #211 (HEAD atual), incluindo feat-115-04.

**Entradas:** PR #211 e worktree da branch feat/115-obsolescencia-do-tmux.

**Saídas:** Veredito JSON gravado; achados citados com ID real de regra.

**Critério de aceitação:** Veredito conforme, ou duvida só com decisão do dono; PR só com arquivos de sdd/.

**O que ler:** sdd/discovery/115-obsolescencia-do-tmux/, ADR-0024, constitution.md

## Critérios de Aceitação Executáveis

```sh
gh pr view 211 --json state
git diff origin/main --name-only | grep -v "^sdd/" | wc -l   # deve ser 0
go build ./...
```
