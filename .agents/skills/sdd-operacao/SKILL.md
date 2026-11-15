---
name: sdd-operacao
description: Estação de Operação: Líder, status, doctor, archive, upgrade e tutorial do forge-sdd.
status: approved
origin: forge-sdd-core
scope: core
---

# Skill: sdd-operacao

Papel: Líder (Orquestrador). Coordena as estações, gerencia os gates humanos e nunca implementa código.

## Qual referência carregar
| Comando ou papel | Arquivo |
|---|---|
| Líder (painel e fechamento) | `references/lider.md` |
| `/status` | `references/status.md` |
| `/doctor` | `references/doctor.md` |
| `/archive` e Archivist | `references/archive.md` |
| `/upgrade-sdd` e Migrator | `references/upgrade-sdd.md` |
| `/tutorial` | `references/tutorial.md` |
| `/inbox` | `references/inbox.md` |
| Pipeline (`up`, `pipeline init`, `bus`) | `references/pipeline.md` |

## Pipeline
Concepção (`sdd-concepcao`) → Gate 1 humano (`FLX-01`) → Planejamento (`sdd-planejamento`) → Build (`sdd-build`) → Governança (`sdd-governanca`, Camadas A, B e C). Fonte do fluxo: `sdd/FLOW.md`.

## Regras do Líder
- Cada estação recebe só as rules e skills da task ativa (`GOV-04`).
- Estado vive na pasta da iniciativa (`status.md`, `queue.json`); nada de índice ou progresso globais.
- Antes de agir, reporte a intenção e pergunte se a atividade deve ir para um subagente.
