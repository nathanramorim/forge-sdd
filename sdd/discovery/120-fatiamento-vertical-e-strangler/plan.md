# Plano — 120

## Fatias e validação
Esta iniciativa se aplica a si: cada task é uma fatia vertical (arquivo do repo + template + golden) que valida e entra na main sozinha.
- **Piloto (baixo risco):** task 1 (critério do split-features); exercita o ciclo repo + template + golden.
- **Validação:** `go test ./internal/scaffold/...` e os greps C1..C3 por fatia.
- **Reversão:** `git revert` do PR da fatia; sem migração de dados e sem flag.
- Não é migração de legado: a variante Strangler não se aplica aqui.

## Onda 1 (3 tasks, independentes)
| # | Task | Arquivos (repo, `.tmpl`, golden) | Rules/IDs |
|---|---|---|---|
| 1 | Passo 4 do split-features: exigir fatia vertical que valida e entra na main sozinha; "baixo para cima" só dentro da fatia | `split-features.md` | ARQ-03, DUP-01, COD-05 |
| 2 | Seção `Fatias e validação` (piloto, validação, reversão, Strangler por roteador ou flag; aponta o critério, não repete) | `discovery.md` passo 7, `sdd-planejamento/SKILL.md` | ARQ-03, DUP-01, DUP-04 |
| 3 | S-08 acoplamento de implantação (cita `FLX-03`; `duvida` sem evidência literal) | `sdd-governanca/SKILL.md`, `pr-review.md` | ARQ-03, FLX-08, DUP-04 |

Cada task regenera o golden (`go test ./internal/scaffold/ -update`, depois sem a flag) e roda `go build ./... && go vet ./...`. Sem conflito de arquivo entre 1, 2 e 3.
