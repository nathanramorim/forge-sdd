---
description: "Alinha a constitution e stack do SDD com a realidade do codebase"
agent: agent
---

Invoque o Specifier para alinhar a estrutura SDD com o codebase atual.

**Ações obrigatórias:**
1. Ler arquivos de projeto (`go.mod`, `package.json`, etc.) e mapear diretórios.
2. Identificar a Missão, Stack e Regras de Ouro (máx. 10).
0c. **Ferramentas:** Perguntar ao usuário (a) qual VCS/work item system este projeto usa — `github`, `azure-devops` ou `nenhum` — e (b) se os MCPs configurados (`context7`, `git`) estão de fato respondendo. Registrar em `sdd/memory/constitution.md` (seção **Ferramentas e Integrações**) e atualizar `sdd/memory/mcps.md` (coluna Status: `ativo`/`indisponível`). Comandos que hoje assumem `gh pr create` ou `context7`/`git` incondicionalmente devem checar esses campos antes de agir.
3. Atualizar:
   - `sdd/memory/constitution.md`
   - `sdd/spec/stack.md`
   - `sdd/spec/overview.md` (com diagramas **C4 Model Mermaid**)

**Guardrail:**
Não apague definições manuais do usuário em `constitution.md` sem perguntar. O objetivo é complementar e organizar.
