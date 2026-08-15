---
description: "Health check da estrutura SDD de forge-sdd"
agent: agent
---

Verifique a saúde da estrutura Forge-SDD deste projeto:
1. `sdd/memory/progress.md` ≤ 1 KB?
2. Todos os chatmodes existem (6 arquivos)?
3. Todos os prompts existem (7 arquivos)?
4. `.vscode/mcp.json` com context7 + git?
5. Alguma feature `in-progress` sem branch no git?
6. `sdd/.metrics/schema.json` presente?
7. Execute `forge-sdd doctor` no terminal e inclua o resumo agregado de telemetria (sessões approved/rejected/blocked) que ele imprime.

Reporte verde ✓ ou vermelho ✗ por item. Não execute correções.
