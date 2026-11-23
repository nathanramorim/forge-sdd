# feat-120-02 — Seção `Fatias e validação` no plano

Task 2 da Onda 1. Branch própria `feat/120-02-secao-fatias-e-validacao-<sid>`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): só texto de skills e `.tmpl`; sem código.

## Decisões do projeto aplicadas
- `ARQ-03`: edita-se o `.tmpl` embutido; a cópia em `.agents/` é espelho.
- `DUP-01`: a seção apenas aponta para o critério do passo 4 de `split-features.md`, sem repeti-lo.
- `DUP-04`: sem template novo; sem template de `plan.md` (DA-4).
- `COD-05`: pt-BR.

## O que já existe
| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| descrição do plano | `discovery.md` passo 7; `sdd-planejamento/SKILL.md` e `.tmpl` | "quebra de baixo para cima" | estender |
| critério de fatia | `split-features.md` passo 4 (feat-120-01) | casa única | apontar, não copiar |

## Dependências novas
Nenhuma (`ARQ-04`).

## Arquivos que toca
- `.agents/skills/sdd-concepcao/references/discovery.md` (passo 7)
- `.agents/skills/sdd-planejamento/SKILL.md`
- `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/discovery.md.tmpl`
- `internal/scaffold/templates/.agents/skills/sdd-planejamento/SKILL.md.tmpl`
- golden: `internal/scaffold/testdata/golden/.agents/skills/sdd-planejamento/SKILL.md` e `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/discovery.md` (regenerados por `-update`)

A seção cita piloto, validação, reversão e Strangler (roteador ou flag; só para migração de legado). Sem conflito com feat-120-01 (outro arquivo) e feat-120-03.

## Rules e IDs da Constituição Aplicáveis
`ARQ-03`, `DUP-01`, `DUP-04`

## Critérios de Aceitação Executáveis

Cada linha deve sair 0 (`set -o pipefail`); a última garante que só documentação e testdata mudaram.

```bash
set -o pipefail
go test ./internal/scaffold/ -update && go test ./internal/scaffold/
grep -q "Fatias e validação" .agents/skills/sdd-concepcao/references/discovery.md
grep -q "Fatias e validação" .agents/skills/sdd-planejamento/SKILL.md
grep -q "Fatias e validação" internal/scaffold/templates/.agents/skills/sdd-concepcao/references/discovery.md.tmpl
grep -q "Fatias e validação" internal/scaffold/templates/.agents/skills/sdd-planejamento/SKILL.md.tmpl
grep -qi "strangler" .agents/skills/sdd-planejamento/SKILL.md
grep -qi "roteador" .agents/skills/sdd-planejamento/SKILL.md
go test ./internal/harness/... && go build ./... && go vet ./...
! git diff --name-only origin/main | grep -E '^(sdd/memory/constitution\.md|sdd/spec/adr/)|\.go$' | grep -v '/testdata/'
```
