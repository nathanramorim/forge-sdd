# Etapa 4 — Integrar a 5.1 na main

**Objetivo:** Mesclar o #226 com merge commit para a tag v5.1.0-beta ser ancestral da main.

**Entradas:** PR #226; consentimento merge-commit-5-1.

**Saídas:** main com a 5.1 no histórico e a versão 5.2.0-beta preservada.

**Critério de aceitação:** git merge-base --is-ancestor v5.1.0-beta main retorna 0; npm/package.json segue em 5.2.0-beta; go test passa.

**O que ler:** PR #226, sdd/releases/history.md

## Critérios de Aceitação Executáveis

```sh
git merge-base --is-ancestor v5.1.0-beta main
grep "\"version\"" npm/package.json   # deve seguir em 5.2.0-beta
go test ./...
```
