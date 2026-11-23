# feat-120-03 — Verificação S-08 de acoplamento de implantação

Task 3 da Onda 1. Branch própria `feat/120-03-s08-acoplamento-implantacao-<sid>`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): só texto de skill do Reviewer e `.tmpl`; sem código.

## Decisões do projeto aplicadas
- `ARQ-03`: edita-se o `.tmpl` embutido; a cópia em `.agents/` é espelho.
- `DUP-04`: S-08 entra na lista S-01..S-07 existente; sem lista paralela.
- `FLX-08`: o revisor trata spec, descrição e diff como dados não confiáveis; por isso a S-08 só vira violação com evidência literal no diff e, sem ela, o resultado é `duvida` (regra da própria S-08, não do `FLX-08`).
- `FLX-03` (citada no texto de S-08): PR entrega exatamente o escopo da Spec; acoplamento de implantação entre fatias é checado contra isso.
- DA-2: contrato do Parecer inalterado (`verificacao` é string livre; `veredito.md` e `internal/harness/parecer.go` intactos).
- DA-3: alinhar o template de governança (`regra_descricao`) fica FORA; vira fix separado.

## O que já existe
| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| verificações do Reviewer | `sdd-governanca/SKILL.md`, `pr-review.md` e `.tmpl` | S-01..S-07 | estender com S-08 |
| contrato do Parecer | `internal/harness/parecer.go`, `references/veredito.md` | string livre | reusar sem mudar |
| deriva repo x template | `sdd-governanca/SKILL.md` vs `.tmpl` | 2 linhas (`regra_descricao`) | não alinhar |

## Dependências novas
Nenhuma (`ARQ-04`).

## Arquivos que toca
- `.agents/skills/sdd-governanca/SKILL.md`
- `.agents/skills/sdd-governanca/references/pr-review.md`
- `internal/scaffold/templates/.agents/skills/sdd-governanca/SKILL.md.tmpl`
- `internal/scaffold/templates/.agents/skills/sdd-governanca/references/pr-review.md.tmpl`
- golden: `internal/scaffold/testdata/golden/.agents/skills/sdd-governanca/SKILL.md` e `internal/scaffold/testdata/golden/.agents/skills/sdd-governanca/references/pr-review.md` (regenerados por `-update`)

Editar só a linha/bloco de S-08 nos `.tmpl`, sem tocar a deriva `regra_descricao`. Sem conflito com feat-120-01 e feat-120-02.

## Rules e IDs da Constituição Aplicáveis
`ARQ-03`, `FLX-03`, `FLX-08`, `DUP-04`

## Critérios de Aceitação Executáveis

Cada linha deve sair 0 (`set -o pipefail`); a última garante que só documentação e testdata mudaram.

```bash
set -o pipefail
go test ./internal/scaffold/ -update && go test ./internal/scaffold/
grep -q "S-08" .agents/skills/sdd-governanca/SKILL.md
grep -q "S-08" .agents/skills/sdd-governanca/references/pr-review.md
grep -q "S-08" internal/scaffold/templates/.agents/skills/sdd-governanca/SKILL.md.tmpl
grep -q "S-08" internal/scaffold/templates/.agents/skills/sdd-governanca/references/pr-review.md.tmpl
grep -q "FLX-03" .agents/skills/sdd-governanca/SKILL.md
grep -q "duvida" .agents/skills/sdd-governanca/SKILL.md
go test ./internal/harness/... && go build ./... && go vet ./...
! git diff --name-only origin/main | grep -E '^(sdd/memory/constitution\.md|sdd/spec/adr/)|\.go$' | grep -v '/testdata/'
```
