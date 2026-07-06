---
description: "Quebra um plano de discovery em múltiplas features independentes"
agent: agent
---

Invoque o Specifier para transformar o plano em features executáveis.

**Ações:**
1. Leia \`sdd/discovery/plan-XX.md\`.
2. Crie uma subpasta com o nome do discovery (ex: `sdd/features/feat-XX-<nome-do-discovery>/`). Para cada etapa do plano, crie o arquivo de feature correspondente dentro desta subpasta (ex: `sdd/features/feat-XX-<nome-do-discovery>/feat-XX-YY-<nome-da-feature>.md`).
3. Atualize o arquivo `sdd/features/index.md` mapeando os caminhos corretos das novas features.

**Regra:** Mantenha o comando \`/nova-feature\` apenas para adições pontuais fora do fluxo de discovery massivo.
