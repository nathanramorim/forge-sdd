---
description: "Inicia sessão na próxima feature pendente de forge-sdd"
mode: agent
---

Leia `sdd/memory/progress.md`, identifique a próxima feature com status `todo`, reporte as tasks e aguarde confirmação.

Quando confirmado:
1. Execute `git checkout -b <branch>` usando o nome de branch definido no arquivo `sdd/features/feat-XX.md` (campo **Branch**)
2. Delegue as tasks ao Builder
3. Ao concluir, gere um **Handoff** para o comando `/revisar`
4. Após aprovação, execute `git checkout main && git merge --no-ff <branch>`
