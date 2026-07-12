---
description: "Realiza discovery de demandas complexas (Produto + Engenharia)"
agent: agent
---

Você deve atuar como um **Analista de Produto Sênior** e um **Engenheiro de Software Sênior**.

**Ação:**
Invoque o processo de discovery para a demanda: ${input:descrição da demanda}

1. **Analise:** Problema, valor, personas e fluxo (visão Produto).
2. **Defina:** Requisitos, impactos técnicos e critérios de aceite (visão Engenharia). Utilize o **C4 Model (Mermaid)** para o desenho da solução.

**Saída:**
Crie os arquivos na pasta `sdd/discovery/`:
- `discovery-XX-<nome>.md` (Estrutura do Produto)
- `criteria-XX-<nome>.md` (Critérios de Aceite Técnicos)
- `plan-XX-<nome>.md` (Roadmap preliminar e estimativa de quebra de tarefas/features)

Se precisar de mais detalhes, peça ao usuário antes de criar os arquivos.

Verifique `Nível de Linguagem` em `sdd/memory/constitution.md`: se `iniciante`, explique conceitos como "C4 Model" ou "critério de aceite" com exemplos concretos e linguagem simples, sem alterar os critérios de aceite em si.

**Handoff:**
Ao finalizar, gere um resumo estruturado para o próximo passo (`/split-features`), listando os arquivos criados e instruindo a quebrar as features geradas organizando-as dentro de uma subpasta de feature com o nome deste discovery (`sdd/features/feat-XX-<nome-do-discovery>/`).
