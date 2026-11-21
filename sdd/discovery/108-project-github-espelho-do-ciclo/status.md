# Status — 108-project-github-espelho-do-ciclo

**Status:** Gate 1 aprovado (20/11/2026). Nada implementado. Spec-validation = aprovada e não mesclada; Todo = spec na principal.
**Dono (20/11):** Builder começa sozinho após o merge da spec; via rápida (sem código) vai de Todo a Validation; automações do Project fecham issue e movem cartão.
**Decididas (dono):** disparo = `forge project observar` local; palavra de fechamento só em PR que mira a main, fallback do forge nos encadeados.
**Aberta:** ligar as automações do Project na interface.
**Decisões:** arquivo é a fonte; arrastar é pedido; campo `Status`; etiquetas e prefixo `[tipo]`.
**Próximo passo:** mesclar o PR da spec; `/split-features` da onda 1.
**Sequenciar com:** 105/2.6 e 2.7, 106/0.2, 2.4 e 2.6, 107.
**BPMN:** 3 diagramas em `diagramas/` (ciclo, sincronia, fechamento).
**A verificar:** ator do evento, webhooks, Azure Boards, PNGs do Mermaid.
