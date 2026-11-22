# Feature 113-06 — References de `/discovery`, `/nova-feature` e `/novo-fix`, templates e goldens

## Objetivo
Acrescentar às 3 references de concepção um passo curto de criação/vínculo da issue ou card antes do Gate 1, sem inflar o orçamento P-24.

## Modelo do módulo
**Enxuto** (ARQ-11): só texto de references e seus templates espelhados; sem código novo.

## Decisões do projeto aplicadas
- **ADR-0003** (`adr-0003-skills-por-estacao-e-orcamento.md`, existe): `SKILL.md` de concepção fica intocado (perto do teto de 2.048 bytes); o passo entra só nas references.
- **ADR-0004** (existe): o vínculo gravado na concepção é o que o `Dispatch` carrega depois.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| References de concepção | `.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md` | Existem; sem menção a issue/card/MCP | Estender | Passo enxuto |
| Templates e goldens | `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/*.md.tmpl`; goldens em `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/*.md` | Existem. Os comandos do Claude também têm `internal/scaffold/templates/agents/claude/.claude/commands/{discovery,nova-feature,novo-fix}.md.tmpl`, que o plano não cita | Estender só as references | Não tocar os comandos por ora; se o Builder achar necessário, avisar o dono |
| Orçamento P-24 | `internal/gate/gate.go:833` (`checkP24ContextBudgets`); `internal/scaffold/budget_test.go:155,183` (`TestContextBudget_*`); `internal/gate/p24_budget_test.go` (`TestCheckP24_*`) | Existem | Reusar | O texto novo deve ser curto, para não estourar o conjunto de leitura por comando |
| Golden | `internal/scaffold/golden_test.go:14,16` | `TestGoldenInit` com flag `-update` | Reusar | Regenerar goldens com `-update` e revisar o diff |
| Modo da execução | fix-89-01 (já na `main`) | Chat não presume `forge up` | Aplicar | O passo chama `forge queue card` direto, sem tmux nem `forge up` |

## Dependências novas
Nenhuma (SEG-07, ARQ-04).

## Rules e IDs da Constituição Aplicáveis
`ARQ-03`, `DUP-10`, `COD-05`, `TST-02`, `TST-03`.

## Critérios de Aceitação Executáveis
Os testes de orçamento e goldens (`go test ./internal/scaffold`) devem passar; o texto novo deve ser curto. `SKILL.md` e Constituição não mudam.
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/scaffold
go test ./internal/scaffold -run 'TestContextBudget_ConjuntoDeLeituraPorComando|TestContextBudget_ConstitutionDoScaffoldDentroDoLimite|TestGoldenInit'
go test ./internal/gate -run 'TestCheckP24'
go test ./internal/scaffold -run 'TestReferencesConcepcaoCitamForgeQueueCard|TestReferencesConcepcaoNaoPressupoemForgeUp'
git diff --quiet origin/main -- .agents/skills/sdd-concepcao/SKILL.md
grep -F 'forge queue card' .agents/skills/sdd-concepcao/references/discovery.md .agents/skills/sdd-concepcao/references/nova-feature.md .agents/skills/sdd-concepcao/references/novo-fix.md

# Os testes citados existem de verdade (sem isso o -run passa vazio):
test "$(go test ./internal/scaffold -list 'TestContextBudget_ConjuntoDeLeituraPorComando|TestContextBudget_ConstitutionDoScaffoldDentroDoLimite|TestGoldenInit' | grep -c '^Test')" -ge 3   # pelo menos 3 (go test -run sem correspondência sai com 0)
test "$(go test ./internal/gate -list 'TestCheckP24' | grep -c '^Test')" -ge 1   # pelo menos 1 (go test -run sem correspondência sai com 0)
test "$(go test ./internal/scaffold -list 'TestReferencesConcepcaoCitamForgeQueueCard|TestReferencesConcepcaoNaoPressupoemForgeUp' | grep -c '^Test')" -ge 2   # pelo menos 2 (go test -run sem correspondência sai com 0)
```
Casos: as 3 references citam `forge queue card` e o caso "gestor desligado: nada a fazer"; nenhuma menciona `forge up` como pré-requisito (negativo); `SKILL.md` idêntico à `main`; goldens regenerados e coerentes com os templates.

## Depende de
113-04, 113-05.
