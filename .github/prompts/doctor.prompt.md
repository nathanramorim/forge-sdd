---
description: "Health check da estrutura SDD do forge-sdd"
mode: ask
---

Verifique a saúde da estrutura Forge-SDD deste projeto:
1. `sdd/memory/progress.md` ≤ 1 KB?
2. Todos os chatmodes existem (6 arquivos)?
3. Todos os prompts existem (7 arquivos)?
4. `.vscode/mcp.json` com context7 + git?
5. Alguma feature em `in-progress` sem branch no git?
6. `.metrics/schema.json` presente?

Reporte verde ✓ ou vermelho ✗ por item. Não execute nenhuma correção.
