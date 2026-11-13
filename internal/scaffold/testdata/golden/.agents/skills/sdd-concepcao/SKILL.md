---
name: sdd-concepcao
description: Estação de Concepção (PM e Arquiteto): discovery, nova feature, fix, C4 e constitution.
status: approved
origin: forge-sdd-core
scope: core
---

# Skill: sdd-concepcao

Papel: PM + Arquiteto. Concebe a Spec e nunca implementa (`FLX-01`).

## Qual referência carregar
| Comando | Arquivo |
|---|---|
| `/discovery` | `references/discovery.md` |
| `/nova-feature` | `references/nova-feature.md` |
| `/novo-fix` | `references/novo-fix.md` |
| `/c4-architecture` | `references/c4-architecture.md` |
| `/constitution` | `references/constitution.md` |

Carregue só a linha do comando pedido.

## Regras comuns
1. **Leitura mínima:** `forge-sdd status` e `sdd/memory/lessons.md`. Da Constitution, só os IDs que a tarefa cita (`forge-sdd context <task>`).
2. **Clarify:** avalie a demanda contra `sdd/memory/clarify.md`; só pergunte se houver lacuna real.
3. **ID e pasta:** `forge-sdd next-id` define o ID. A iniciativa vive em `sdd/discovery/<id>-<nome>/` com `discovery.md`, `criteria.md`, `plan.md`, `status.md`, `features/` e `fixes/`.
4. **Seções obrigatórias da Spec:** `## Modelo do módulo` (`ARQ-11`), `## O que já existe` (`DUP-02`, consulte `sdd/spec/concepts.md` e `.agents/skills/`), `## Dependências novas` (`SEG-07`), `## Rules e IDs da Constituição Aplicáveis`, `## Critérios de Aceitação Executáveis`.
5. **Não aprove** Spec sem `O que já existe`, com critério não executável ou em conflito com a Constitution sem ADR (`GOV-01`).
6. **Handoff:** entregue ao Líder para o Gate 1 (`FLX-01`). Telemetria: `forge-sdd session record`.
