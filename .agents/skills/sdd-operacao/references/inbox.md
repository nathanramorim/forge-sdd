# Comando: inbox

**Uso:** `/inbox` mostra as notificações do pipeline (PR revisado pronto para aprovação, escalonamentos) da sessão Discovery.

1. Rode `forge-sdd inbox` (não lidas) ou `forge-sdd inbox --include-read`.
2. Para cada `pr_approved`, apresente o link do PR: revisado, sem inconformidades e pronto para a aprovação final. Para `escalation`, explique que o ciclo automático terminou e o que a pessoa precisa decidir.
3. Quando a notificação trouxer `Instrução ao agente` (canal `pm_comment`), comente no card pelo MCP.
4. Depois de tratar: `forge-sdd inbox --mark-read <id>`, `--mark-all` ou `--archive`.

O hook da sessão Discovery (`UserPromptSubmit`) já chama `forge-sdd inbox --unread --hook` a cada prompt no Claude; no Gemini e no agy, que não têm esse hook, use `/inbox`.
