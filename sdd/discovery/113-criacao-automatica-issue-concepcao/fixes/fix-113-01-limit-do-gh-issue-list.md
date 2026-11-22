# fix-113-01 — A busca de issue aberta olha só os 30 primeiros resultados

**Branch da spec:** `fix/113-07-limit-gh`; **da correção:** `fix/113-07-impl-limit-gh`, a partir da principal
**Origem:** Iniciativa 113, observação do Reviewer da feat-113-03 (22/11/2026): "a janela de 30 itens do `gh` continua valendo para a primeira criação, sem evidência".
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

`ProcurarIssueAberta` (`internal/vcs/issue.go:49`) roda `gh issue list --state open --search <título>` **sem `--limit`**; o padrão do `gh` é 30. Num repositório com muitas issues abertas parecidas, a issue já existente pode ficar **fora** dos 30 primeiros, e a concepção cria uma **duplicada**. A evidência `card-evidence.json` evita o problema nas execuções seguintes, mas não na primeira.

## Feature Correlacionada

- **Arquivo de origem:** `internal/vcs/issue.go` (feat-113-01, `ProcurarIssueAberta`).
- **Causa-raiz:** o limite padrão do `gh` foi aceito sem decisão; o filtro exato roda no cliente, depois do corte.
- **Evidência:** `issue.go:49-50` (sem `--limit`); o `ProcurarPRAberto` usa `--limit 1` de propósito, a issue não.

## Modelo do módulo

**Enxuto.** Uma flag, um comentário e testes em `internal/vcs` e `internal/pipeline`; sem comando novo nem mudança de assinatura.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-01` (a saída do `gh` continua validada contra o esquema), `TST-03` (casos conformes e negativos), `DUP-01` (o limite é uma constante nomeada, um lugar só), `SEG-07` (nenhuma dependência).

## Decisões do projeto aplicadas

- A busca segue server-side por título e filtrada no cliente (a busca do `gh` é aproximada).
- Limite maior em vez de paginação. O `gh` pagina a API por dentro (100 por página): `--limit 200` custa **2 requisições** em vez de 1, aceito. A busca do GitHub tem teto de 1000 resultados.
- **Aviso de "busca no limite" fica fora desta spec** (follow-up): ele exigiria um canal novo (mudar a assinatura de `ProcurarIssueAberta` atinge github, nenhum, `corpo_issue.go` e 5 fakes de teste, ou uma sentinela de erro). Com 200 o risco restante é pequeno.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Limite de listagem | `internal/vcs/github.go:35` | `--limit 1` no PR | **não mudar** |
| Runner falso nos testes | `internal/vcs/*_test.go` | `Runner` injetável | **reusar** para registrar os argumentos |

## Comportamento esperado

1. `ProcurarIssueAberta` passa `--limit 200` (constante `limiteBuscaIssue`), documentada. O comentário em `internal/pipeline/corpo_issue.go` (~linhas 157-159) que cita esse follow-up é atualizado.
2. O restante não muda: a busca segue server-side por título, filtrada no cliente.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**: os 2 testes novos não existem, e o teste existente `TestGitHubProcurarIssueAbertaIdempotente` (`internal/vcs/issue_test.go:41`) afirma o argv antigo e **deve ser atualizado** para `... --search 'Minha spec' --limit 200 --json number,url,title`; só então o último comando fica verde.

```bash
set -o pipefail
# 1. o argv leva --limit 200 (o runner falso registra o argv e afirma o par `--limit`,`200`)
go test ./internal/vcs -run 'TestProcurarIssueAbertaUsaLimite' -count=1
# 2. issue na posição >30 de uma lista de 200 itens, com um casando por título: é achada (runner falso devolve 200 itens)
go test ./internal/vcs -run 'TestProcurarIssueAbertaAchaAlemDos30' -count=1
# 3. existência (só prova que existem; a aprovação é o -run acima)
[ "$(go test ./internal/vcs -list 'TestProcurarIssueAbertaUsaLimite|TestProcurarIssueAbertaAchaAlemDos30' | grep -c '^Test')" -eq 2 ]
# 4. o teste existente atualizado e o conjunto afetado
go test ./internal/vcs ./internal/pipeline ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Aviso de busca no limite (follow-up, pede canal novo) e paginação completa.
- Busca por label ou por corpo.
