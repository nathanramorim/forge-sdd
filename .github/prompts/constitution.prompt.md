---
description: "Alinha a constitution e stack do SDD com a realidade do codebase"
agent: agent
---

Invoque o Specifier para alinhar a estrutura SDD com o codebase atual.

**Ações obrigatórias:**
1. Ler arquivos de projeto (`go.mod`, `package.json`, etc.) e mapear diretórios.
2. Identificar a Missão, Stack e Regras de Ouro (máx. 10).
3. Atualizar:
   - `sdd/memory/constitution.md`
   - `sdd/spec/stack.md`
   - `sdd/spec/overview.md` (com diagramas **C4 Model Mermaid**)

**Guardrail:**
Não apague definições manuais do usuário em `constitution.md` sem perguntar. O objetivo é complementar e organizar.
