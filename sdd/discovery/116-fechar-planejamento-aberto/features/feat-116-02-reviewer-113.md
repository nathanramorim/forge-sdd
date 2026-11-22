# Etapa 2 — Conferir a spec da 113

**Objetivo:** Reconferir o PR #175 depois de forge card virar forge queue card.

**Entradas:** PR #175.

**Saídas:** Veredito JSON gravado.

**Critério de aceitação:** Sem pendência de ADR ou de emenda a ARQ-06; plano coerente.

**O que ler:** sdd/discovery/113-criacao-automatica-issue-concepcao/, ARQ-06 na constitution.md

## Critérios de Aceitação Executáveis

```sh
gh pr view 175 --json mergeable,isDraft
grep -rn "forge card" sdd/discovery/113-criacao-automatica-issue-concepcao | wc -l   # deve ser 0
```
