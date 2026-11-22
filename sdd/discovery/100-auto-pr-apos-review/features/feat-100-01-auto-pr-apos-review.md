# feat-100-01 — forge queue review: abrir PR automaticamente após veredito conforme

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote novo. Estende `cmd/forge-sdd/queue.go` (`queueReviewCmd`) e `internal/queue/queue.go` para, quando o veredito for `conforme`, executar `git push` + `gh pr create` via `os/exec`. Nenhuma dependência Go nova.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| `queueReviewCmd` | `cmd/forge-sdd/queue.go:161` | Cobra RunE que chama `queue.Review()` e imprime status | Estender: após `conforme`, chamar helper `openPR` |
| `queue.Queue.Branch` | `internal/queue/queue.go:32` | Campo `branch` já persistido no `queue.json` | Reusar: é a branch a empurrar e usar como `--head` do PR |
| `queue.Queue.BaseBranch` | `internal/queue/queue.go:33` | Campo `base_branch` (vazio = `main`) | Reusar como `--base` do PR (fallback `main`) |
| `queue.Item.SpecPath` | `internal/queue/queue.go:44` | Caminho da spec da task | Reusar: extrair cabeçalho H1 para título do PR |
| `queue.Item.GovernanceVerdict.Resumo` | `internal/queue/queue.go:638` | Resumo textual do veredito | Reusar no corpo do PR |
| `gh` CLI | `os/exec` | Disponível na maioria dos ambientes de desenvolvimento | Reusar: `exec.LookPath("gh")` para verificar presença |
| Flag `--no-pr` | `cmd/forge-sdd/queue.go` | Não existe | Criar: flag booleana; quando verdadeira, pula o PR |

## Comportamento esperado

```
# Fluxo normal (conforme):
$ forge queue review feat-98-01 --verdict-file veredito.json
✓ Parecer registrado para feat-98-01: veredito=conforme → status=approved
→ empurrando feat/98-diagram-studio-c4...  ok
→ PR aberto: https://github.com/owner/repo/pull/105

# Com --no-pr:
$ forge queue review feat-98-01 --verdict-file veredito.json --no-pr
✓ Parecer registrado para feat-98-01: veredito=conforme → status=approved
  (PR não aberto: use --no-pr=false ou abra manualmente)

# gh ausente:
→ gh não encontrado; rode 'brew install gh && gh auth login' para abrir PRs automaticamente

# Já tem PR aberto (gh pr view retorna 0):
→ PR já existe para esta branch; nada feito
```

## Decisões do projeto aplicadas

- **ARQ-01**: lógica de PR em `internal/queue/pr.go` (novo arquivo enxuto); `cmd/` só chama.
- **ARQ-04**: sem dependência nova em runtime; `gh` é verificado com `exec.LookPath`.
- **ARQ-10**: título e corpo do PR são montados em `internal/queue/pr.go`, não em `cmd/`.
- **COD-01**: `openPR(dir string, q *Queue, item *Item, out io.Writer)` — responsabilidade única.
- **COD-03**: erros de `git push` e `gh pr create` enriquecidos com contexto.

## Dependências novas

Nenhuma. Usa `os/exec` (stdlib) e `gh` CLI (verificado em runtime; falha graciosamente se ausente).

## Rules e IDs da Constituição aplicáveis

`ARQ-01`, `ARQ-04`, `ARQ-10`, `ARQ-11`, `COD-01`, `COD-03`

## Critérios de Aceitação Executáveis

```bash
# 1. Flag --no-pr registrada no comando
grep -c "no-pr" cmd/forge-sdd/queue.go
# deve retornar >= 1

# 2. Função openPR (ou equivalente) em internal/queue/
grep -rn "func.*PR\|openPR\|createPR" internal/queue/
# deve retornar >= 1

# 3. exec.LookPath("gh") presente
grep -rn "LookPath.*gh\|gh.*LookPath" internal/queue/ cmd/forge-sdd/
# deve retornar >= 1

# 4. Integração: após veredito conforme, mensagem de PR na saída
go test ./cmd/forge-sdd/... -run TestQueueReview -v 2>&1 | grep -c "PR aberto\|pr.*url\|no-pr"
# deve retornar >= 1

# 5. go vet limpo
go vet ./...
# exit 0

# 6. go build limpo
go build ./...
# exit 0
```
