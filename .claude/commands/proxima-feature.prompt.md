Leia `sdd/memory/progress.md`, identifique a próxima feature com status `todo`, reporte as tasks e aguarde confirmação.

Quando confirmado:
1. Execute `git checkout -b <branch>` usando o nome de branch definido no arquivo `sdd/features/feat-XX.md` (campo **Branch**)
2. Implemente as tasks listadas
3. Ao concluir, gere um **Handoff** para a revisão, marque as tasks como concluídas em `progress.md` e `index.md`, e escreva uma release note concisa em linguagem de produto (curta, sem termos técnicos de código, ex: "Agora é possível...") registrando-a no topo de `sdd/releases/history.md` (sob a seção ## Entregas).
4. **PR Automático (gh CLI):** Faça o commit das alterações de progresso, execute `git push origin <branch>` e crie IMEDIATAMENTE o PR com `gh pr create --fill`.
5. **Fallbacks:** Se o `gh` não estiver disponível, faça o merge local na main: `git checkout main && git merge --no-ff <branch>` e informe o usuário.
