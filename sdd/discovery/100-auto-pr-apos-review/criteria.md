# Critérios de Arquitetura — 100-auto-pr-apos-review

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote Go novo. Estende `cmd/forge-sdd/queue.go` e adiciona `internal/queue/pr.go`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| `queueReviewCmd` | `cmd/forge-sdd/queue.go:161` | Cobra RunE que chama `queue.Review()` | Estender: chamar `openPR` após conforme |
| `queue.Queue.Branch` | `internal/queue/queue.go:32` | Campo `branch` no `queue.json` | Reusar como `--head` do PR |
| `queue.Queue.BaseBranch` | `internal/queue/queue.go:33` | Campo `base_branch` (vazio = `main`) | Reusar como `--base` do PR |
| `queue.Item.SpecPath` | `internal/queue/queue.go:44` | Caminho da spec da task | Reusar para extrair título (H1) |
| `gh` CLI | runtime | Disponível na maioria dos ambientes | Reusar via `exec.LookPath("gh")` |

## Dependências novas

Nenhuma. Usa `os/exec` (stdlib).

## Critérios de aceitação executáveis

Ver `features/feat-100-01-auto-pr-apos-review.md`.
