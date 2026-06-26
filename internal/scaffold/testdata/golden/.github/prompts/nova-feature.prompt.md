---
description: "Cria nova feature para demo a partir de uma descrição"
mode: agent
---

Invoque o Specifier para criar uma feat-XX a partir desta descrição: ${input:descrição da feature}

**Handoff:**
Ao finalizar, gere um resumo para o comando `/proxima-feature`.

> Após o Specifier criar o `sdd/features/feat-XX.md`, use o prompt `proxima-feature` para iniciar a execução — ele cria a branch `feat/*` antes de delegar ao Builder.
