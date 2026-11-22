# Etapa 3 — Spec da 98 sem estado de entrega

**Objetivo:** Tirar do PR #214 a fila, os vereditos e o status de implementado, e reconferir.

**Entradas:** PR #214 e a branch feat/98-diagram-studio-c4.

**Saídas:** PR de spec só com spec; veredito JSON gravado.

**Critério de aceitação:** Critérios com go build e go test (TST-01); skill no repositório; instalação opcional e restrita ao projeto aberto.

**O que ler:** sdd/discovery/98-diagram-studio-c4/

## Critérios de Aceitação Executáveis

```sh
gh pr view 214 --json state
go build ./... && go test ./...
```
