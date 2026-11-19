# Comando: inbox

**Uso:** `/inbox` mostra as notificações do pipeline (PR revisado pronto para aprovação, escalonamentos) da sessão Discovery.

1. Rode `forge inbox` (não lidas) ou `forge inbox --include-read`.
2. Para cada notificação, identifique o tipo:
   - **`pr_approved`** (canal `pr_approved` ou texto contendo "revisado, sem inconformidades e pronto para aprovação final"): use `AskUserQuestion` com título "✅ PR #<n> aprovado — o que fazer?" e opções:
     - "Ver o PR antes — abre o link e aguarda": exiba o link e aguarde; não faça merge.
     - "Mergear agora — gh pr merge <n> --squash --delete-branch": execute `gh pr merge <n> --squash --delete-branch` sem pedir nova confirmação e depois marque como lida (`forge inbox --mark-read <id>`).
   - **`escalation`**: explique que o ciclo automático terminou e o que a pessoa precisa decidir.
3. Quando a notificação trouxer `Instrução ao agente` (canal `pm_comment`), comente no card pelo MCP.
4. Depois de tratar: `forge inbox --mark-read <id>`, `--mark-all` ou `--archive`.

O hook da sessão Discovery (`UserPromptSubmit`) já chama `forge inbox --unread --hook` a cada prompt no Claude; no Gemini e no agy, que não têm esse hook, use `/inbox`.
