Leia `sdd/memory/progress.md`, identifique a próxima feature com status `todo`, reporte as tasks e aguarde confirmação.

Quando confirmado:
1. Execute `git checkout -b <branch>` usando o nome de branch definido no arquivo `sdd/features/feat-XX.md` (campo **Branch**)
2. Implemente as tasks listadas
3. Ao concluir, gere um **Handoff** para a revisão, marque as tasks como concluídas em `progress.md` e `index.md`, e escreva uma release note concisa em linguagem de produto (curta, sem termos técnicos de código, ex: "Agora é possível...") registrando-a no topo de `sdd/releases/history.md` (sob a seção ## Entregas).
4. **PR Automático (gh CLI):** Faça o commit das alterações de progresso, execute `git push origin <branch>` e crie IMEDIATAMENTE o PR com `gh pr create --fill`.
5. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc` (`telemetry.enabled` como `true`), grave a telemetria em `sdd/.metrics/session-<ISO8601>.json` respeitando o schema local:
   - O campo `"feature"` deve refletir o caminho relativo completo da feature/fix ou task correspondente (ex: `sdd/features/feat-<workitem/hash>-<nome>/task-<id>.md` ou `sdd/features/fix-<workitem/hash>-<nome>/task-<id>.md`).
   - Se a sessão for inativa, cancelada, sofrer timeout ou encerrar sem atingir a finalização da feature/fix (`criterio_atendido: false`), grave a métrica marcando `outcome: blocked` ou `outcome: rejected` para registrar o esforço parcial.
   - **Estimativa de Tokens:** Estime os campos `"tokens_input"` e `"tokens_output"` de forma realista baseando-se nos caracteres/palavras processados e gerados na sessão (conversão média: 1 token ≈ 4 caracteres ou 0.75 palavras), nunca os deixando zerados se houve interação.
6. **Fallbacks:** Se o `gh` não estiver disponível, faça o merge local na main: `git checkout main && git merge --no-ff <branch>` e informe o usuário.
