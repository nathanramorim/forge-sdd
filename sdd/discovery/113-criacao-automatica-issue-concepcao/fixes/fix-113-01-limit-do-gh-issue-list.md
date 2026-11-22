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

**Enxuto.** Uma flag e testes em `internal/vcs`; sem comando novo (`ARQ-06`).

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-01` (a saída do `gh` continua validada contra o esquema), `TST-03` (caso negativo: a issue está além da posição 30), `DUP-01` (o limite é uma constante nomeada, um lugar só).

## Decisões do projeto aplicadas

- A busca segue server-side por título e filtrada no cliente (a busca do `gh` é aproximada).
- Limite maior em vez de paginação: simples, determinístico e sem rede extra.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Limite de listagem | `internal/vcs/github.go:35` | `--limit 1` no PR | **não mudar** |
| Runner falso nos testes | `internal/vcs/*_test.go` | `Runner` injetável | **reusar** para registrar os argumentos |

## Comportamento esperado

1. `ProcurarIssueAberta` passa `--limit 200` (constante `limiteBuscaIssue`), documentada: acima disso a busca devolve "não achou" e a criação segue.
2. Se a lista voltar com exatamente `limiteBuscaIssue` itens e nenhum casar, o resultado inclui um **aviso** ("busca no limite; pode existir issue além da janela") que o `forge queue card` imprime higienizado (SEG-08). Não bloqueia.

## Critérios de Aceitação Executáveis

Os testes novos ficam **vermelhos até o Builder** criá-los.

```bash
set -o pipefail
go test ./internal/vcs -run 'TestProcurarIssueAbertaUsaLimite|TestProcurarIssueAbertaAvisaNoLimite' -count=1
[ "$(go test ./internal/vcs -list 'TestProcurarIssueAbertaUsaLimite|TestProcurarIssueAbertaAvisaNoLimite' | grep -c '^Test')" -eq 2 ]
go test ./internal/vcs ./internal/pipeline ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Paginação completa; busca por label ou por corpo.
