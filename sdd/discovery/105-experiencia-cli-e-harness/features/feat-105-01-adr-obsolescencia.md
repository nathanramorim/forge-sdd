# feat-105-01 — ADR de obsolescência (5 estações, depreciações, `lider`)

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.1)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** nada (decisão do dono já dada no Gate 1)
**Paralelo com:** nenhuma; destrava 1.2, 1.3 e 1.6
**Sequenciar com:** 106 e 107: numerar a ADR como a próxima livre no momento do PR (`ls sdd/spec/adr`), sem reservar número (0021 ou posterior); não editar ADR de outra iniciativa

## Problema
O produto virou 3 sessões (ADR-0004), mas a ADR-0003 ainda fala em "skills por estação" e o CLI expõe `queue stations`, `attach` e `queue handoff`. `ARQ-06` exige ADR para depreciar subcomando; sem ela as tasks 1.2, 1.3 e 1.6 não têm base.

## Comportamento
1. Criar `sdd/spec/adr/adr-NNNN-obsolescencia-das-5-estacoes.md` (NNNN = próxima livre) no modelo `adr-0000-modelo.md`, status `aceito`.
2. Decide: (a) as 5 estações deixam de ser fluxo; é decisão parcialmente supersedendo ADR-0003 (skills `sdd-*` ficam como apoio das 3 sessões); (b) depreciar (só avisar, nada removido) `queue stations`, `attach`, `queue handoff`, `/implement`, `/proxima-feature`; (c) agente `lider` fica, só muda a descrição; (d) remoção futura exige ADR própria.
3. Anotar a supersessão parcial no cabeçalho da ADR-0003 (uma linha apontando a nova ADR), sem reescrevê-la.

## Modelo do módulo
**Sem código.** Só documentação de decisão (`ARQ-11` não se aplica).

## Decisões do projeto aplicadas
ADR-0003 (parcialmente obsoleta, é o alvo), ADR-0004 (3 sessões), ADR-0005 (menu), ADR-0002. `obsoletos.md` O2 a O7 e Gate 1 de 2026-11-20.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| ADR sobre estações | `sdd/spec/adr/adr-0003-skills-por-estacao-e-orcamento.md` | Skills por estação, status aprovado | **Estender** (nota de supersessão parcial) | Não duplicar decisão (`DUP-06`) |
| Próximo número | `sdd/spec/adr/` | Último arquivo: `adr-0020` | **Criar** com a próxima livre | 106/107 podem também numerar ADR |
| Modelo | `sdd/spec/adr/adr-0000-modelo.md` | Modelo de ADR | **Reusar** | Padrão |
| Lista de subcomandos | `.agents/rules/` e `sdd/memory/constitution.md` (`ARQ-06`) | Regra da lista | **Reusar** como citação; não alterar a Constituição | Alterar `ARQ-06` é fora do escopo |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`FLX-05`, `GOV-01`, `ARQ-06`. Rules: `sdd-governanca` (ADR e decisão).

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
N=$(ls sdd/spec/adr | grep -E '^adr-00(2[1-9]|[3-9][0-9])-obsolescencia' | head -1)
test -n "$N"
grep -F 'aceito' "sdd/spec/adr/$N" >/dev/null
grep -F 'ADR-0003' "sdd/spec/adr/$N" >/dev/null
grep -F 'queue stations' "sdd/spec/adr/$N" >/dev/null
grep -F 'lider' "sdd/spec/adr/$N" >/dev/null
grep -F "${N%.md}" sdd/spec/adr/adr-0003-skills-por-estacao-e-orcamento.md >/dev/null || grep -F 'ADR-00' sdd/spec/adr/adr-0003-skills-por-estacao-e-orcamento.md >/dev/null
# negativo: nenhuma ADR duplica número
test -z "$(ls sdd/spec/adr | sed -E 's/^(adr-[0-9]{4}).*/\1/' | sort | uniq -d)"
go build ./... && go test ./...
```
- Nada removido: `git diff --name-status origin/main -- cmd internal .claude` não pode listar `D` (revisão do Reviewer).
- Com 106/107: se outra ADR tomar o número antes do merge, renumerar no rebase, sem editar a da outra iniciativa.
