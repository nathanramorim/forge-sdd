---
description: "Alinha a constitution e stack do SDD com a realidade do codebase"
agent: agent
---

Invoque o Specifier para alinhar a estrutura SDD com o codebase atual.

**Ações obrigatórias:**
0. Antes de qualquer outra coisa, pergunte ao usuário (a) em qual idioma o chat deve interagir daqui em diante e (b) em qual idioma escrever commits e PRs (título e descrição). Registre a resposta em `sdd/memory/constitution.md`, seção Regras de Ouro, e siga essa escolha a partir desse ponto, inclusive nesta própria sessão.
0b. Pergunte também se o usuário prefere o nível `padrão` (jargão técnico normal) ou `iniciante` (linguagem simplificada, com exemplos concretos no lugar de termos como "C4 Model" ou "critério de aceitação executável"). Registre a resposta em `sdd/memory/constitution.md`, seção Regras de Ouro, como `Nível de Linguagem: <padrão|iniciante>`, e respeite essa escolha em toda futura invocação de qualquer comando SDD, sem nunca alterar os critérios de aceitação executáveis em si.
0c. **Ferramentas:** Pergunte ao usuário (a) qual VCS/work item system este projeto usa — `github`, `azure-devops` ou `nenhum` — e (b) se os MCPs configurados (`context7`, `git`) estão de fato respondendo. Registre a resposta em `sdd/memory/constitution.md` (seção **Ferramentas e Integrações**) e atualize `sdd/memory/mcps.md` (coluna Status: `ativo`/`indisponível`). Comandos que hoje assumem `gh pr create` ou `context7`/`git` incondicionalmente devem checar esses campos antes de agir.
1. Ler arquivos de projeto (`go.mod`, `package.json`, etc.) e mapear diretórios.
2. Identificar a Missão, Stack e Regras de Ouro (máx. 10).
3. Atualizar:
   - `sdd/memory/constitution.md`
   - `sdd/spec/stack.md`
   - `sdd/spec/overview.md` (com diagramas **C4 Model Mermaid**)

**Guardrail:**
Não apague definições manuais do usuário em `constitution.md` sem perguntar. O objetivo é complementar e organizar.
