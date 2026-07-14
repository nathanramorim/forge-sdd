# Feature 44-03 — Chave `active_release_branch` e Direcionamento de PR

**Branch:** `feat/44-03-active-release-branch`
**Depende de:** feat-44-01 (usa o mesmo ponto de sabatina para confirmar a branch de destino)
**Paralelizável (worktree):** Não — mesma dependência de sequência de 44-01.

## Descrição
Quando várias features/fixes convergem para uma única entrega, os PRs devem apontar para uma branch de release compartilhada em vez de `main`. Esta feature adiciona a configuração e a lógica de leitura correspondente.

## Critérios de Aceitação Executáveis

1. Nova chave opcional `active_release_branch` (string) em `sdd/.sddrc`.
2. `/novo-fix`, `/nova-feature` e o passo de PR de `/proxima-feature` (Claude e Gemini) usam essa chave como `--base` do `gh pr create` quando presente e válida (branch existe local ou remotamente).
3. Se a chave estiver ausente ou apontar para uma branch inválida, o PR usa `main` como base, com um aviso reportado ao usuário nesse segundo caso.
4. `/nova-feature` e `/novo-fix` detectam quando a branch atual segue o padrão `feat/release-*` ou `release/*` e, via sabatina (feat-44-01), sugerem gravar essa branch como `active_release_branch` em `.sddrc`.
