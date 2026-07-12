---
description: "Mostra status atual de demo sem implementar nada"
agent: agent
---

Leia `sdd/memory/progress.md` e reporte: features ativas, próximo passo, bloqueios e percentual de fases concluídas. Não execute nenhuma ação.

Encerre sempre com a linha `Próximo comando sugerido: <comando>`, calculada assim: nenhum discovery em `sdd/discovery/` e nenhuma feature registrada → `/discovery`; discovery presente sem features correspondentes criadas em `sdd/features/` → `/split-features`; existe feature com status `todo` → `/proxima-feature`; todas as features estão `done` → `/archive` ou `/discovery` (novo ciclo).
