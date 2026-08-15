# Prompt: constitution

**Uso:** Peça "/constitution" ou "alinhar o projeto"

**Ação:**
Assuma o papel de **Specifier**. Seu objetivo é garantir que os arquivos `constitution.md` e `stack.md` reflitam a realidade técnica do projeto.

0. **Idioma:** Antes de qualquer outra coisa, pergunte ao usuário (a) em qual idioma o chat deve interagir daqui em diante e (b) em qual idioma escrever commits e PRs (título e descrição). Registre a resposta em `constitution.md`, seção **Regras**, e siga essa escolha a partir desse ponto, inclusive nesta própria sessão.
0c. **Ferramentas:** Pergunte ao usuário (a) qual VCS/work item system este projeto usa — `github`, `azure-devops` ou `nenhum` — e (b) se os MCPs configurados (`context7`, `git`) estão de fato respondendo. Registre a resposta em `constitution.md` (seção **Ferramentas e Integrações**) e atualize `sdd/memory/mcps.md` (coluna Status: `ativo`/`indisponível`). Comandos que hoje assumem `gh pr create` ou `context7`/`git` incondicionalmente devem checar esses campos antes de agir.
1. **Scan:** Leia o diretório raiz e subpastas importantes.
2. **Contexto:** Identifique stack principal, gerenciadores de dependência e padrões arquiteturais.
3. **Povoamento:** Preencha os artefatos do SDD com essas informações, incluindo diagramas **C4 Model (Mermaid)** na `overview.md`, garantindo clareza e objetividade.

Se o projeto estiver vazio, proponha uma arquitetura base para a stack configurada no SDD.
